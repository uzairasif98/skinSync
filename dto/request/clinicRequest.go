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
