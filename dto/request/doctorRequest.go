package request

// ContactInfo represents contact information for a doctor/injector
type ContactInfo struct {
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone,omitempty"`
}

// TreatmentSideArea represents a treatment with its associated side areas
type TreatmentSideArea struct {
	TreatmentID        uint   `json:"treatment_id" validate:"required"`
	TreatmentsSubSecID []uint `json:"treatments_sub_sec_id" validate:"required,min=1"` // side_area_ids
}

// RegisterDoctorRequest represents the request to register a doctor/injector
type RegisterDoctorRequest struct {
	Role           string              `json:"role" validate:"required,oneof=doctor injector"`
	Name           string              `json:"name" validate:"required,min=2"`
	Image          string              `json:"image,omitempty"`
	Specialization string              `json:"specialization,omitempty"`
	ContactInfo    ContactInfo         `json:"contact_info" validate:"required"`
	Treatments     []TreatmentSideArea `json:"treatments,omitempty"` // optional now
}

// AssignDoctorTreatmentsRequest represents the request to assign treatments to a doctor
type AssignDoctorTreatmentsRequest struct {
	ClinicUserID uint64              `json:"clinic_user_id" validate:"required"`
	Treatments   []TreatmentSideArea `json:"treatments" validate:"required,min=1"`
}
