package services

import (
	"skinSync/config"
	resdto "skinSync/dto/response"
	"skinSync/models"
)

// GetTreatmentMasters returns all treatments
func GetTreatmentMasters() (resdto.TreatmentMastersResponse, error) {
	db := config.DB

	var treatments []models.Treatment

	err := db.Find(&treatments).Error
	if err != nil {
		return resdto.TreatmentMastersResponse{
			IsSuccess: false,
			Message:   "Failed to fetch treatments",
			Data:      nil,
		}, err
	}

	// Convert to DTOs
	var data []resdto.TreatmentDTO
	for _, t := range treatments {
		data = append(data, resdto.TreatmentDTO{
			ID:          t.ID,
			Name:        t.Name,
			Icon:        t.Icon,
			Description: t.Description,
			IsArea:      t.IsArea,
		})
	}

	return resdto.TreatmentMastersResponse{
		IsSuccess: true,
		Message:   "Treatments retrieved successfully",
		Data:      data,
	}, nil
}

// GetAreasByTreatment returns all areas for a given treatment ID
func GetAreasByTreatment(treatmentID uint) (resdto.AreasResponse, error) {
	db := config.DB

	var areas []models.Area

	err := db.Where("treatment_id = ?", treatmentID).Find(&areas).Error
	if err != nil {
		return resdto.AreasResponse{
			IsSuccess: false,
			Message:   "Failed to fetch areas",
			Data:      nil,
		}, err
	}

	// Convert to DTOs
	var data []resdto.AreaDTO
	for _, a := range areas {
		data = append(data, resdto.AreaDTO{
			ID:          a.ID,
			Name:        a.Name,
			Icon:        a.Icon,
			Description: a.Description,
			IsSideArea:  a.IsSideArea,
		})
	}

	return resdto.AreasResponse{
		IsSuccess: true,
		Message:   "Areas retrieved successfully",
		Data:      data,
	}, nil
}

// GetSideAreas returns all side areas for a given treatment ID and area ID
func GetSideAreas(treatmentID, areaID uint) (resdto.SideAreasResponse, error) {
	db := config.DB

	var sideAreas []models.SideArea

	err := db.Where("treatment_id = ? AND area_id = ?", treatmentID, areaID).Find(&sideAreas).Error
	if err != nil {
		return resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Failed to fetch side areas",
			Data:      nil,
		}, err
	}

	// Convert to DTOs
	var data []resdto.SideAreaDTO
	for _, sa := range sideAreas {
		// Generate syringe options from min to max
		var syringeOptions []int
		for i := sa.MinSyringe; i <= sa.MaxSyringe; i++ {
			syringeOptions = append(syringeOptions, i)
		}

		data = append(data, resdto.SideAreaDTO{
			ID:             sa.ID,
			Name:           sa.Name,
			Icon:           sa.Icon,
			Description:    sa.Description,
			MinSyringe:     sa.MinSyringe,
			MaxSyringe:     sa.MaxSyringe,
			SyringeOptions: syringeOptions,
		})
	}

	return resdto.SideAreasResponse{
		IsSuccess: true,
		Message:   "Side areas retrieved successfully",
		Data:      data,
	}, nil
}
