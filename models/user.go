package models

import (
	"time"

	"gorm.io/gorm"
)

// Users
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PrimaryEmail *string        `gorm:"size:255;uniqueIndex" json:"primary_email"`
	PrimaryPhone *string        `gorm:"size:20;uniqueIndex" json:"primary_phone"`
	Status       string         `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	AuthProviders []AuthProvider `gorm:"constraint:OnDelete:CASCADE"`
	AuthTokens    []AuthToken    `gorm:"constraint:OnDelete:CASCADE"`
	Roles         []Role         `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
}
