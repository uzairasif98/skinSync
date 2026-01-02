package models

import (
	"time"
)

// Treatment represents main treatment types (Dermal Fillers, Botox, etc.)
type Treatment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;unique" json:"name"`
	Icon        string    `gorm:"size:255" json:"icon,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	IsArea      bool      `gorm:"default:false" json:"is_area"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Areas       []Area    `gorm:"foreignKey:TreatmentID" json:"areas,omitempty"`
}

func (Treatment) TableName() string {
	return "treatments"
}

// Area represents areas within each treatment
type Area struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TreatmentID uint       `gorm:"not null;index" json:"treatment_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Icon        string     `gorm:"size:255" json:"icon,omitempty"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	IsSideArea  bool       `gorm:"default:false" json:"is_sidearea"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Treatment   Treatment  `gorm:"foreignKey:TreatmentID;constraint:OnDelete:CASCADE" json:"-"`
	SideAreas   []SideArea `gorm:"foreignKey:AreaID" json:"side_areas,omitempty"`
}

func (Area) TableName() string {
	return "areas"
}

// SideArea represents specific side areas within an area
type SideArea struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TreatmentID uint      `gorm:"not null;index" json:"treatment_id"`
	AreaID      uint      `gorm:"not null;index" json:"area_id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Icon        string    `gorm:"size:255" json:"icon,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	MinSyringe  int       `gorm:"default:1" json:"min_syringe"`
	MaxSyringe  int       `gorm:"default:1" json:"max_syringe"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Treatment   Treatment `gorm:"foreignKey:TreatmentID;constraint:OnDelete:CASCADE" json:"-"`
	Area        Area      `gorm:"foreignKey:AreaID;constraint:OnDelete:CASCADE" json:"-"`
}

func (SideArea) TableName() string {
	return "side_areas"
}
