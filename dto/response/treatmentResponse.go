package response

// TreatmentMastersResponse for GET /api/treatments/masters
type TreatmentMastersResponse struct {
	IsSuccess bool            `json:"is_success"`
	Message   string          `json:"message"`
	Data      []TreatmentTypeDTO `json:"data"`
}

// TreatmentTypeDTO represents treatment type with categories
type TreatmentTypeDTO struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Categories  []TreatmentCategoryDTO `json:"categories"`
}

// TreatmentCategoryDTO represents category with treatments
type TreatmentCategoryDTO struct {
	ID         uint           `json:"id"`
	Name       string         `json:"name"`
	Treatments []TreatmentDTO `json:"treatments"`
}

// TreatmentDTO represents individual treatment
type TreatmentDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MaxSyringes *int   `json:"max_syringes,omitempty"`
}
