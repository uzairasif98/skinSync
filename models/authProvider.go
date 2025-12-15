package models

import "time"

// AuthProviders - different login methods for a user
type AuthProvider struct {
	ID           uint64  `gorm:"primaryKey;autoIncrement"`
	UserID       uint64  `gorm:"index;not null"`
	Provider     string  `gorm:"size:50;not null"` // phone, email, google, apple
	Phone        *string `gorm:"size:20;uniqueIndex:idx_provider_phone"`
	Email        *string `gorm:"size:255;uniqueIndex:idx_provider_email"`
	PasswordHash *string `gorm:"size:255"`
	GoogleUID    *string `gorm:"size:255;uniqueIndex:idx_provider_googleuid"`
	AppleUID     *string `gorm:"size:255;uniqueIndex:idx_provider_appleuid"`
	CreatedAt    time.Time
}
