package models

import (
	"time"

	"gorm.io/gorm"
)

// TreatmentType represents main treatment types (Dermal Fillers, Botox)
type TreatmentType struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:100;not null;unique" json:"name"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	DisplayOrder int            `gorm:"default:0" json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Categories   []TreatmentCategory `gorm:"foreignKey:TreatmentTypeID" json:"categories,omitempty"`
}

func (TreatmentType) TableName() string {
	return "treatment_types"
}

// TreatmentCategory represents categories within each treatment type
type TreatmentCategory struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TreatmentTypeID uint           `gorm:"not null;index" json:"treatment_type_id"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description,omitempty"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	DisplayOrder    int            `gorm:"default:0" json:"display_order"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	TreatmentType   TreatmentType  `gorm:"foreignKey:TreatmentTypeID;constraint:OnDelete:CASCADE" json:"-"`
	Treatments      []Treatment    `gorm:"foreignKey:CategoryID" json:"treatments,omitempty"`
}

func (TreatmentCategory) TableName() string {
	return "treatment_categories"
}

// Treatment represents individual treatment areas
type Treatment struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CategoryID   uint           `gorm:"not null;index" json:"category_id"`
	Name         string         `gorm:"size:200;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	MaxSyringes  *int           `gorm:"default:null" json:"max_syringes,omitempty"` // Only for Dermal Fillers
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	DisplayOrder int            `gorm:"default:0" json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Category     TreatmentCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Treatment) TableName() string {
	return "treatments"
}
