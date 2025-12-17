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

// UserProfile - stores user profile details
type UserProfile struct {
	UserProfileID uint64 `gorm:"primaryKey;autoIncrement"`

	UserID uint64 `gorm:"not null;index:idx_user_id"`

	Name string `gorm:"size:100;not null"`

	PhoneNumber *string `gorm:"size:20"`

	EmailAddress *string `gorm:"size:150;unique"`

	Location *string `gorm:"size:550"`

	Bio *string `gorm:"type:text"`

	ProfileImagePath *string `gorm:"size:255"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Relationship - UserProfile belongs to User
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}
