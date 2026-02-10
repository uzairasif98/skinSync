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

// ClinicRole represents roles specific to clinic users (separate from admin roles)
type ClinicRole struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ClinicRole) TableName() string {
	return "clinic_roles"
}

// ClinicUser represents users belonging to a clinic
type ClinicUser struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64     `gorm:"not null;uniqueIndex:idx_email_clinic" json:"clinic_id"`
	Email        string     `gorm:"size:255;not null;uniqueIndex:idx_email_clinic" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	RoleID       uint64     `gorm:"not null" json:"role_id"`
	Status       string     `gorm:"size:20;default:'active'" json:"status"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	// Relationships
	Clinic Clinic     `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	Role   ClinicRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (ClinicUser) TableName() string {
	return "clinic_users"
}

// Clinic role name constants
const (
	ClinicRoleOwner        = "owner"
	ClinicRoleManager      = "manager"
	ClinicRoleDoctor       = "doctor"
	ClinicRoleInjector     = "injector"
	ClinicRoleReceptionist = "receptionist"
)

// ClinicPermission represents permissions specific to clinic users
type ClinicPermission struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description *string `gorm:"size:255" json:"description,omitempty"`
}

func (ClinicPermission) TableName() string {
	return "clinic_permissions"
}

// ClinicRolePermission maps clinic roles to clinic permissions
type ClinicRolePermission struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint64 `gorm:"not null;index:idx_clinic_role_perm,unique" json:"role_id"`
	PermissionID uint64 `gorm:"not null;index:idx_clinic_role_perm,unique" json:"permission_id"`

	// Relationships
	Role       ClinicRole       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"-"`
	Permission ClinicPermission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ClinicRolePermission) TableName() string {
	return "clinic_role_permissions"
}

// ClinicTreatment maps which treatments a clinic offers
type ClinicTreatment struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID    uint64    `gorm:"not null;uniqueIndex:idx_clinic_treatment" json:"clinic_id"`
	TreatmentID uint      `gorm:"not null;uniqueIndex:idx_clinic_treatment" json:"treatment_id"`
	Price       *float64  `json:"price,omitempty"`
	Status      string    `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Clinic    Clinic    `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	Treatment Treatment `gorm:"foreignKey:TreatmentID;constraint:OnDelete:CASCADE" json:"treatment,omitempty"`
}

func (ClinicTreatment) TableName() string {
	return "clinic_treatments"
}

// ClinicUserTreatment maps which treatments a doctor can perform
type ClinicUserTreatment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicUserID uint64    `gorm:"not null;uniqueIndex:idx_user_treatment" json:"clinic_user_id"`
	TreatmentID  uint      `gorm:"not null;uniqueIndex:idx_user_treatment" json:"treatment_id"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	ClinicUser ClinicUser `gorm:"foreignKey:ClinicUserID;constraint:OnDelete:CASCADE" json:"clinic_user,omitempty"`
	Treatment  Treatment  `gorm:"foreignKey:TreatmentID;constraint:OnDelete:CASCADE" json:"treatment,omitempty"`
}

func (ClinicUserTreatment) TableName() string {
	return "clinic_user_treatments"
}

// ClinicSideArea links a clinic with a side area (under a treatment and area)
type ClinicSideArea struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ClinicID    uint      `gorm:"index:idx_clinic_treatment_area_side,unique" json:"clinic_id"`
	TreatmentID uint      `gorm:"index:idx_clinic_treatment_area_side,unique" json:"treatment_id"`
	AreaID      uint      `gorm:"index:idx_clinic_treatment_area_side,unique" json:"area_id"`
	SideAreaID  uint      `gorm:"index:idx_clinic_treatment_area_side,unique" json:"side_area_id"`
	SyringeSize int       `gorm:"index:idx_clinic_treatment_area_side,unique" json:"syringe_size,omitempty"`
	Price       *float64  `json:"price,omitempty"`
	Status      string    `gorm:"size:20;default:'active'" json:"status,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	SideArea SideArea `gorm:"foreignKey:SideAreaID" json:"-"`
}

func (ClinicSideArea) TableName() string {
	return "clinic_side_areas"
}
