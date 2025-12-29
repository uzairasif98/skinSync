package services

import (
	"skinSync/config"
	resdto "skinSync/dto/response"
	"skinSync/models"

	"gorm.io/gorm"
)

// GetTreatmentMasters returns all treatment types with categories and treatments
func GetTreatmentMasters() (resdto.TreatmentMastersResponse, error) {
	db := config.DB

	var treatmentTypes []models.TreatmentType

	// Fetch all active treatment types with their categories and treatments
	err := db.Where("is_active = ?", true).
		Order("display_order ASC").
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("display_order ASC")
		}).
		Preload("Categories.Treatments", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("display_order ASC")
		}).
		Find(&treatmentTypes).Error

	if err != nil {
		return resdto.TreatmentMastersResponse{
			IsSuccess: false,
			Message:   "Failed to fetch treatment masters",
		}, err
	}

	// Convert to DTOs
	var data []resdto.TreatmentTypeDTO
	for _, tt := range treatmentTypes {
		typeDTO := resdto.TreatmentTypeDTO{
			ID:          tt.ID,
			Name:        tt.Name,
			Description: tt.Description,
			Categories:  []resdto.TreatmentCategoryDTO{},
		}

		for _, cat := range tt.Categories {
			catDTO := resdto.TreatmentCategoryDTO{
				ID:         cat.ID,
				Name:       cat.Name,
				Treatments: []resdto.TreatmentDTO{},
			}

			for _, treat := range cat.Treatments {
				treatDTO := resdto.TreatmentDTO{
					ID:          treat.ID,
					Name:        treat.Name,
					Description: treat.Description,
					MaxSyringes: treat.MaxSyringes,
				}
				catDTO.Treatments = append(catDTO.Treatments, treatDTO)
			}

			typeDTO.Categories = append(typeDTO.Categories, catDTO)
		}

		data = append(data, typeDTO)
	}

	return resdto.TreatmentMastersResponse{
		IsSuccess: true,
		Message:   "Treatment masters retrieved",
		Data:      data,
	}, nil
}
