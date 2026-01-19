package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser represents admin and clinic users in the system
type AdminUser struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	RoleID       uint64         `gorm:"not null" json:"role_id"`
	Role         Role           `gorm:"foreignKey:RoleID" json:"role"`
	Status       string         `gorm:"size:20;default:'active'" json:"status"`
	LastLogin    *time.Time     `json:"last_login,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AdminUser
func (AdminUser) TableName() string {
	return "admin_users"
}
