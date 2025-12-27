package models

import "time"

// OTP - stores one-time passwords for email verification
type OTP struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Email     string    `gorm:"size:255;not null;index"`
	OTPCode   string    `gorm:"size:6;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
	Attempts  int       `gorm:"default:0"`
	CreatedAt time.Time
}
