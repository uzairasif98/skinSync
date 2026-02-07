package request

// RegisterClinicRequest represents clinic registration request by super_admin
type RegisterClinicRequest struct {
	// Clinic info
	ClinicName    string `json:"clinic_name" validate:"required,min=2"`
	ClinicEmail   string `json:"clinic_email" validate:"required,email"`
	ClinicPhone   string `json:"clinic_phone"`
	ClinicAddress string `json:"clinic_address"`
	ClinicLogo    string `json:"clinic_logo"`

	// Owner info
	OwnerName  string `json:"owner_name" validate:"required,min=2"`
	OwnerEmail string `json:"owner_email" validate:"required,email"`
}

// ClinicLoginRequest represents clinic user login request
type ClinicLoginRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	ClinicID *uint64 `json:"clinic_id,omitempty"` // Optional: required when user belongs to multiple clinics
}

// RegisterClinicUserRequest represents request to register a new clinic user (staff)
type RegisterClinicUserRequest struct {
	Email  string `json:"email" validate:"required,email"`
	Name   string `json:"name" validate:"required,min=2"`
	RoleID uint64 `json:"role_id" validate:"required"`
}
