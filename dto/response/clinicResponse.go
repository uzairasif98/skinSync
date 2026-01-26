package response

// RegisterClinicResponse represents clinic registration response
type RegisterClinicResponse struct {
	BaseResponse
	Data *RegisterClinicData `json:"data,omitempty"`
}

// RegisterClinicData contains registered clinic info
type RegisterClinicData struct {
	ClinicID       uint64 `json:"clinic_id"`
	ClinicName     string `json:"clinic_name"`
	ClinicEmail    string `json:"clinic_email"`
	OwnerEmail     string `json:"owner_email"`
	OwnerPassword  string `json:"owner_password"` // Plain password (only returned once, for testing)
	Status         string `json:"status"`
}

// ClinicDTO represents clinic info for responses
type ClinicDTO struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
	Logo    string `json:"logo,omitempty"`
	Status  string `json:"status"`
}

// ClinicUserDTO represents clinic user info for responses
type ClinicUserDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}
