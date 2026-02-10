package response

// RegisterClinicResponse represents clinic registration response
type RegisterClinicResponse struct {
	BaseResponse
	Data *RegisterClinicData `json:"data,omitempty"`
}

// RegisterClinicData contains registered clinic info
type RegisterClinicData struct {
	ClinicID      uint64 `json:"clinic_id"`
	ClinicName    string `json:"clinic_name"`
	ClinicEmail   string `json:"clinic_email"`
	OwnerEmail    string `json:"owner_email"`
	OwnerPassword string `json:"owner_password"` // Plain password (only returned once, for testing)
	Status        string `json:"status"`
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
	ID       uint64     `json:"id"`
	ClinicID uint64     `json:"clinic_id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Role     string     `json:"role"`
	Status   string     `json:"status"`
	Clinic   *ClinicDTO `json:"clinic,omitempty"`
}

// ClinicLoginResponse represents clinic login response
type ClinicLoginResponse struct {
	BaseResponse
	Data ClinicLoginData `json:"data"`
}

// ClinicLoginData contains login tokens and user info
type ClinicLoginData struct {
	AccessToken             string               `json:"access_token,omitempty"`
	RefreshToken            string               `json:"refresh_token,omitempty"`
	AccessExpiresAt         int64                `json:"access_expires_at,omitempty"`
	RefreshExpiresAt        int64                `json:"refresh_expires_at,omitempty"`
	ClinicUser              *ClinicUserDTO       `json:"clinic_user,omitempty"`
	RequiresClinicSelection bool                 `json:"requires_clinic_selection,omitempty"`
	Clinics                 []ClinicSelectionDTO `json:"clinics,omitempty"`
}

// ClinicSelectionDTO represents clinic info shown during multi-clinic login selection
type ClinicSelectionDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo,omitempty"`
	Role string `json:"role"`
}

// RegisterClinicUserResponse represents clinic user registration response
type RegisterClinicUserResponse struct {
	BaseResponse
	Data *RegisterClinicUserData `json:"data,omitempty"`
}

// RegisterClinicUserData contains registered clinic user info
type RegisterClinicUserData struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Password string `json:"password"` // Plain password (only returned once)
	Status   string `json:"status"`
}

// ClinicSideAreaPriceDTO represents a side area with its price for a clinic
type ClinicSideAreaPriceDTO struct {
	ID          uint     `json:"id"`
	SideAreaID  uint     `json:"side_area_id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Description string   `json:"description,omitempty"`
	MinSyringe  int      `json:"min_syringe"`
	MaxSyringe  int      `json:"max_syringe"`
	Price       *float64 `json:"price,omitempty"`
	Status      string   `json:"status"`
}

// ClinicAreaDTO represents an area with its side areas and prices
type ClinicAreaDTO struct {
	ID          uint                     `json:"id"`
	AreaID      uint                     `json:"area_id"`
	Name        string                   `json:"name"`
	Icon        string                   `json:"icon,omitempty"`
	Description string                   `json:"description,omitempty"`
	SideAreas   []ClinicSideAreaPriceDTO `json:"side_areas,omitempty"`
}

// ClinicTreatmentDTO represents a treatment with areas and side areas priced for a clinic
type ClinicTreatmentDTO struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Icon        string          `json:"icon,omitempty"`
	Description string          `json:"description,omitempty"`
	Price       *float64        `json:"price,omitempty"` // Treatment-level price (if any)
	Areas       []ClinicAreaDTO `json:"areas,omitempty"`
	Status      string          `json:"status"`
}

// GetClinicTreatmentsResponse represents response for getting treatments by clinic
type GetClinicTreatmentsResponse struct {
	BaseResponse
	Data []ClinicTreatmentDTO `json:"data"`
}
