package models

import "time"

// AuthToken - access + refresh token pairs (refresh hashed)
type AuthToken struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	UserID           uint64    `gorm:"index;not null"`
	AccessToken      string    `gorm:"size:500;uniqueIndex"`
	RefreshTokenHash string    `gorm:"size:255;uniqueIndex"`
	AccessExpiresAt  time.Time `gorm:"index"`
	RefreshExpiresAt time.Time `gorm:"index"`
	DeviceInfo       *string   `gorm:"size:255"`
	IPAddress        *string   `gorm:"size:45"`
	IsRevoked        bool      `gorm:"default:false"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
