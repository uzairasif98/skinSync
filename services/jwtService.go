package services

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

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
