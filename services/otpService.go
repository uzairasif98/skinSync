package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
	"skinSync/utils"

	"gorm.io/gorm"
)

const (
	otpLength     = 6
	otpExpiry     = 5 * time.Minute
	maxOTPAttempt = 5
)

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

// SendOTP generates OTP, saves to DB, and sends email
func SendOTP(req reqdto.SendOTPRequest) (resdto.BaseResponse, error) {
	db := config.DB
	if db == nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "database not initialized"}, errors.New("database not initialized")
	}

	// Invalidate any existing unused OTPs for this email
	db.Model(&models.OTP{}).
		Where("email = ? AND is_used = ?", req.Email, false).
		Update("is_used", true)

	// Generate new OTP
	otpCode, err := GenerateOTP()
	if err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to generate OTP"}, err
	}

	// Save OTP to database
	otp := models.OTP{
		Email:     req.Email,
		OTPCode:   otpCode,
		ExpiresAt: time.Now().Add(otpExpiry),
		IsUsed:    false,
		Attempts:  0,
	}

	if err := db.Create(&otp).Error; err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to save OTP"}, err
	}

	// Send OTP via email
	if err := utils.SendOTPEmail(req.Email, otpCode); err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to send OTP email"}, err
	}

	return resdto.BaseResponse{
		IsSuccess: true,
		Message:   "OTP sent to email",
	}, nil
}

// VerifyOTP verifies OTP and returns login response
func VerifyOTP(req reqdto.VerifyOTPRequest) (*resdto.LoginResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Find OTP record
	var otp models.OTP
	err := db.Where("email = ? AND is_used = ?", req.Email, false).
		Order("created_at DESC").
		First(&otp).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("OTP not found. Please request a new one")
	} else if err != nil {
		return nil, err
	}

	// Check if OTP expired
	if time.Now().After(otp.ExpiresAt) {
		otp.IsUsed = true
		db.Save(&otp)
		return nil, errors.New("OTP expired. Please request a new one")
	}

	// Check max attempts
	if otp.Attempts >= maxOTPAttempt {
		otp.IsUsed = true
		db.Save(&otp)
		return nil, errors.New("too many failed attempts. Please request a new OTP")
	}

	// Verify OTP code
	if otp.OTPCode != req.OTP {
		otp.Attempts++
		db.Save(&otp)
		remaining := maxOTPAttempt - otp.Attempts
		return nil, fmt.Errorf("invalid OTP. %d attempts remaining", remaining)
	}

	// Mark OTP as used
	otp.IsUsed = true
	db.Save(&otp)

	// Find or create user
	var user models.User
	var provider models.AuthProvider

	err = db.Where("provider = ? AND email = ?", "email", req.Email).First(&provider).Error
	if err == nil {
		// Provider exists, get user
		if err = db.First(&user, provider.UserID).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Check if user exists by primary email
		err = db.Where("primary_email = ?", req.Email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user
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
	resp.Data.User = user

	return resp, nil
}
