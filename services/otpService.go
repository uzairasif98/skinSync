package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
	"skinSync/utils"

	"gorm.io/gorm"
)

const otpLength = 6

// getOTPExpiry returns OTP expiry duration from env (default 5 minutes)
func getOTPExpiry() time.Duration {
	minutes, err := strconv.Atoi(os.Getenv("OTP_EXPIRY_MINUTES"))
	if err != nil || minutes <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

// getResendCooldown returns resend cooldown from env (default 1 minute)
func getResendCooldown() time.Duration {
	minutes, err := strconv.Atoi(os.Getenv("OTP_RESEND_COOLDOWN_MINUTES"))
	if err != nil || minutes <= 0 {
		return 1 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

// getMaxOTPAttempts returns max attempts from env (default 5)
func getMaxOTPAttempts() int {
	attempts, err := strconv.Atoi(os.Getenv("OTP_MAX_ATTEMPTS"))
	if err != nil || attempts <= 0 {
		return 5
	}
	return attempts
}

// OTPData stores OTP information in memory
type OTPData struct {
	Code       string
	ExpiresAt  time.Time
	LastSentAt time.Time
	Attempts   int
}

// In-memory OTP store with mutex for thread safety
var (
	otpStore = make(map[string]OTPData)
	otpMutex sync.Mutex
)

// StartOTPCleanup runs a background goroutine to clean expired OTPs
func StartOTPCleanup() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			otpMutex.Lock()
			for email, otpData := range otpStore {
				if time.Now().After(otpData.ExpiresAt) {
					delete(otpStore, email)
				}
			}
			otpMutex.Unlock()
		}
	}()
}

// GenerateOTP generates a random 6-digit OTP
func GenerateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, otpLength)
	randomBytes := make([]byte, otpLength)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < otpLength; i++ {
		otp[i] = digits[randomBytes[i]%10]
	}
	return string(otp), nil
}

// SendOTP generates OTP, stores in memory, and sends email
func SendOTP(req reqdto.SendOTPRequest) (resdto.BaseResponse, error) {
	// Check rate limiting
	otpMutex.Lock()
	existingOTP, exists := otpStore[req.Email]
	if exists && time.Since(existingOTP.LastSentAt) < getResendCooldown() {
		otpMutex.Unlock()
		return resdto.BaseResponse{IsSuccess: false, Message: "Please wait before requesting a new OTP"}, errors.New("rate limited")
	}
	otpMutex.Unlock()

	// Generate new OTP
	otpCode, err := GenerateOTP()
	if err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to generate OTP"}, err
	}

	// Store OTP in memory
	otpMutex.Lock()
	otpStore[req.Email] = OTPData{
		Code:       otpCode,
		ExpiresAt:  time.Now().Add(getOTPExpiry()),
		LastSentAt: time.Now(),
		Attempts:   0,
	}
	otpMutex.Unlock()

	// Send OTP via email
	if err := utils.SendOTPEmail(req.Email, otpCode); err != nil {
		// Remove from store if email failed
		otpMutex.Lock()
		delete(otpStore, req.Email)
		otpMutex.Unlock()
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to send OTP email"}, err
	}

	return resdto.BaseResponse{
		IsSuccess: true,
		Message:   "OTP sent to email",
	}, nil
}

// VerifyOTP verifies OTP from memory and returns login response
func VerifyOTP(req reqdto.VerifyOTPRequest) (*resdto.LoginResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Get OTP from memory
	otpMutex.Lock()
	otpData, exists := otpStore[req.Email]
	otpMutex.Unlock()

	if !exists {
		return nil, errors.New("OTP not found. Please request a new one")
	}

	// Check if OTP expired
	if time.Now().After(otpData.ExpiresAt) {
		otpMutex.Lock()
		delete(otpStore, req.Email)
		otpMutex.Unlock()
		return nil, errors.New("OTP expired. Please request a new one")
	}

	// Check max attempts
	if otpData.Attempts >= getMaxOTPAttempts() {
		otpMutex.Lock()
		delete(otpStore, req.Email)
		otpMutex.Unlock()
		return nil, errors.New("too many failed attempts. Please request a new OTP")
	}

	// Verify OTP code
	if otpData.Code != req.OTP {
		otpMutex.Lock()
		otpData.Attempts++
		otpStore[req.Email] = otpData
		otpMutex.Unlock()
		remaining := getMaxOTPAttempts() - otpData.Attempts
		return nil, fmt.Errorf("invalid OTP. %d attempts remaining", remaining)
	}

	// Remove OTP after successful verification
	otpMutex.Lock()
	delete(otpStore, req.Email)
	otpMutex.Unlock()

	// Find or create user
	var user models.User
	var provider models.AuthProvider

	err := db.Where("provider = ? AND email = ?", "email", req.Email).First(&provider).Error
	if err == nil {
		// Provider exists, get user (returning user)
		if err = db.First(&user, provider.UserID).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Check if user exists by primary email
		err = db.Where("primary_email = ?", req.Email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user (first time signup)
			user = models.User{PrimaryEmail: &req.Email, Status: "active"}
			if err = db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		// Create auth provider
		provider = models.AuthProvider{
			UserID:   user.ID,
			Provider: "email",
			Email:    &req.Email,
		}
		if err = db.Create(&provider).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// Check if user has a profile to determine isFirstLogin
	var profile models.UserProfile
	isFirstLogin := true
	if err := db.Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
		// Profile exists - not first login
		isFirstLogin = false
	}

	// Generate tokens
	token, err := GenerateJWT(req.Email, uint(user.ID), "")
	if err != nil {
		return nil, err
	}

	raw, hashed, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	authToken := models.AuthToken{
		UserID:           user.ID,
		AccessToken:      token,
		RefreshTokenHash: hashed,
		AccessExpiresAt:  time.Now().Add(accessTokenDuration),
		RefreshExpiresAt: time.Now().Add(refreshTokenDuration),
		DeviceInfo:       req.DeviceInfo,
		IPAddress:        req.IPAddress,
	}

	if err = db.Create(&authToken).Error; err != nil {
		return nil, err
	}

	resp := &resdto.LoginResponse{}
	resp.IsSuccess = true
	resp.Message = "Logged in"
	resp.Data.AccessToken = token
	resp.Data.RefreshToken = raw
	resp.Data.AccessExpiresAt = authToken.AccessExpiresAt.Unix()
	resp.Data.RefreshExpiresAt = authToken.RefreshExpiresAt.Unix()
	resp.Data.IsFirstLogin = isFirstLogin
	resp.Data.User = user

	return resp, nil
}
