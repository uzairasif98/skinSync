package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// ==================== TOKEN BLACKLIST ====================

// In-memory blacklist for logged-out tokens
var (
	tokenBlacklist = make(map[string]time.Time) // token -> expiry time
	blacklistMutex sync.Mutex
)

// BlacklistToken adds a token to the blacklist until it expires
func BlacklistToken(token string, expiresAt time.Time) {
	blacklistMutex.Lock()
	tokenBlacklist[token] = expiresAt
	blacklistMutex.Unlock()
}

// IsTokenBlacklisted checks if a token has been logged out
func IsTokenBlacklisted(token string) bool {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	_, exists := tokenBlacklist[token]
	return exists
}

// StartTokenBlacklistCleanup runs a background goroutine to clean expired tokens
func StartTokenBlacklistCleanup() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			blacklistMutex.Lock()
			for token, expiresAt := range tokenBlacklist {
				if time.Now().After(expiresAt) {
					delete(tokenBlacklist, token)
				}
			}
			blacklistMutex.Unlock()
		}
	}()
}

// TODO: Refresh token. Reseacrh how it works and then implement.
var jwtSecret = []byte("skinSync")

// Access/refresh expiry durations
var accessTokenDuration = time.Hour * 24
var refreshTokenDuration = time.Hour * 24 * 14 // 2 weeks

// Hash the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Check the password against the hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate JWT token
func GenerateJWT(email string, user_id uint, deviceID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"user_id": user_id,
		"exp":     time.Now().Add(accessTokenDuration).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// GenerateClinicJWT generates JWT token for clinic users with clinic context
func GenerateClinicJWT(email string, clinicUserID uint64, clinicID uint64, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":          email,
		"clinic_user_id": clinicUserID,
		"clinic_id":      clinicID,
		"role":           role,
		"exp":            time.Now().Add(accessTokenDuration).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken returns a cryptographically secure random token and
// the same token hashed using bcrypt.
func GenerateRefreshToken() (rawToken string, hashedToken string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(b)
	hashed, err := HashPassword(raw)
	if err != nil {
		return "", "", err
	}
	return raw, hashed, nil
}
