package models

import (
	"time"
)

// Clinic represents a clinic/business entity
type Clinic struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"size:255;not null" json:"name"`
	Email     string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Phone     string     `gorm:"size:50" json:"phone,omitempty"`
	Address   string     `gorm:"type:text" json:"address,omitempty"`
	Logo      string     `gorm:"size:500" json:"logo,omitempty"`
	Status    string     `gorm:"size:20;default:'active'" json:"status"` // active, inactive, suspended
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	// Relationships
	Users []ClinicUser `gorm:"foreignKey:ClinicID" json:"users,omitempty"`
}

func (Clinic) TableName() string {
	return "clinics"
}

// ClinicUser represents users belonging to a clinic
type ClinicUser struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64     `gorm:"not null;index" json:"clinic_id"`
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	Role         string     `gorm:"size:50;not null" json:"role"` // owner, doctor, injector, receptionist
	Status       string     `gorm:"size:20;default:'active'" json:"status"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	// Relationships
	Clinic Clinic `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
}

func (ClinicUser) TableName() string {
	return "clinic_users"
}

// Clinic user roles
const (
	ClinicRoleOwner        = "owner"
	ClinicRoleDoctor       = "doctor"
	ClinicRoleInjector     = "injector"
	ClinicRoleReceptionist = "receptionist"
)
