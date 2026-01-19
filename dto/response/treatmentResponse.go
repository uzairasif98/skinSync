package response

// TreatmentDTO represents a single treatment
type TreatmentDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	IsArea      bool   `json:"is_area"`
}

// TreatmentMastersResponse for GET /api/treatments/masters
type TreatmentMastersResponse struct {
	IsSuccess bool           `json:"is_success"`
	Message   string         `json:"message"`
	Data      []TreatmentDTO `json:"data"`
}

// AreaDTO represents a single area
type AreaDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	IsSideArea  bool   `json:"is_sidearea"`
}

// AreasResponse for GET /api/treatments/:id/areas
type AreasResponse struct {
	IsSuccess bool      `json:"is_success"`
	Message   string    `json:"message"`
	Data      []AreaDTO `json:"data"`
}

// SideAreaDTO represents a single side area
type SideAreaDTO struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon,omitempty"`
	Description    string `json:"description,omitempty"`
	MinSyringe     int    `json:"min_syringe"`
	MaxSyringe     int    `json:"max_syringe"`
	SyringeOptions []int  `json:"syringe_options"`
}

// SideAreasResponse for GET /api/treatments/:treatmentId/areas/:areaId/sideareas
type SideAreasResponse struct {
	IsSuccess bool          `json:"is_success"`
	Message   string        `json:"message"`
	Data      []SideAreaDTO `json:"data"`
}
