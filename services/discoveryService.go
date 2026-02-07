package services

import (
	"skinSync/config"
	"skinSync/models"
)

// DoctorDTO represents a doctor in discovery responses
type DoctorDTO struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ClinicID uint64 `json:"clinic_id"`
}

// ClinicWithPriceDTO represents a clinic with optional treatment price
type ClinicWithPriceDTO struct {
	ID      uint64   `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Phone   string   `json:"phone,omitempty"`
	Address string   `json:"address,omitempty"`
	Logo    string   `json:"logo,omitempty"`
	Status  string   `json:"status"`
	Price   *float64 `json:"price,omitempty"`
}

// TreatmentWithPriceDTO represents a treatment with optional clinic-specific price
type TreatmentWithPriceDTO struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
}

// GetAllClinics returns all active clinics
func GetAllClinics() ([]models.Clinic, error) {
	db := config.DB
	var clinics []models.Clinic
	if err := db.Where("status = ?", "active").Find(&clinics).Error; err != nil {
		return nil, err
	}
	return clinics, nil
}

// GetAllDoctors returns all clinic users with doctor role
func GetAllDoctors() ([]DoctorDTO, error) {
	db := config.DB

	var clinicUsers []models.ClinicUser
	if err := db.Preload("Role").
		Joins("JOIN clinic_roles ON clinic_roles.id = clinic_users.role_id").
		Where("clinic_roles.name = ? AND clinic_users.status = ?",
			models.ClinicRoleDoctor, "active").
		Find(&clinicUsers).Error; err != nil {
		return nil, err
	}

	doctors := make([]DoctorDTO, 0, len(clinicUsers))
	for _, u := range clinicUsers {
		doctors = append(doctors, DoctorDTO{
			ID:       u.ID,
			Name:     u.Name,
			Email:    u.Email,
			Role:     u.Role.Name,
			ClinicID: u.ClinicID,
		})
	}
	return doctors, nil
}

// GetTreatmentsByClinic returns treatments offered by a specific clinic
func GetTreatmentsByClinic(clinicID uint64) ([]TreatmentWithPriceDTO, error) {
	db := config.DB

	var clinicTreatments []models.ClinicTreatment
	if err := db.Preload("Treatment").
		Where("clinic_id = ? AND status = ?", clinicID, "active").
		Find(&clinicTreatments).Error; err != nil {
		return nil, err
	}

	treatments := make([]TreatmentWithPriceDTO, 0, len(clinicTreatments))
	for _, ct := range clinicTreatments {
		treatments = append(treatments, TreatmentWithPriceDTO{
			ID:          ct.Treatment.ID,
			Name:        ct.Treatment.Name,
			Icon:        ct.Treatment.Icon,
			Description: ct.Treatment.Description,
			Price:       ct.Price,
		})
	}
	return treatments, nil
}

// GetClinicsByTreatment returns clinics that offer a specific treatment
func GetClinicsByTreatment(treatmentID uint) ([]ClinicWithPriceDTO, error) {
	db := config.DB

	var clinicTreatments []models.ClinicTreatment
	if err := db.Preload("Clinic").
		Where("treatment_id = ? AND status = ?", treatmentID, "active").
		Find(&clinicTreatments).Error; err != nil {
		return nil, err
	}

	clinics := make([]ClinicWithPriceDTO, 0, len(clinicTreatments))
	for _, ct := range clinicTreatments {
		if ct.Clinic.Status == "active" {
			clinics = append(clinics, ClinicWithPriceDTO{
				ID:      ct.Clinic.ID,
				Name:    ct.Clinic.Name,
				Email:   ct.Clinic.Email,
				Phone:   ct.Clinic.Phone,
				Address: ct.Clinic.Address,
				Logo:    ct.Clinic.Logo,
				Status:  ct.Clinic.Status,
				Price:   ct.Price,
			})
		}
	}
	return clinics, nil
}

// GetTreatmentsByDoctor returns treatments a doctor can perform
func GetTreatmentsByDoctor(doctorID uint64) ([]TreatmentWithPriceDTO, error) {
	db := config.DB

	var userTreatments []models.ClinicUserTreatment
	if err := db.Preload("Treatment").
		Where("clinic_user_id = ?", doctorID).
		Find(&userTreatments).Error; err != nil {
		return nil, err
	}

	treatments := make([]TreatmentWithPriceDTO, 0, len(userTreatments))
	for _, ut := range userTreatments {
		treatments = append(treatments, TreatmentWithPriceDTO{
			ID:          ut.Treatment.ID,
			Name:        ut.Treatment.Name,
			Icon:        ut.Treatment.Icon,
			Description: ut.Treatment.Description,
		})
	}
	return treatments, nil
}

// GetDoctorsByClinicAndTreatment returns doctors who can perform a treatment at a specific clinic
func GetDoctorsByClinicAndTreatment(clinicID uint64, treatmentID uint) ([]DoctorDTO, error) {
	db := config.DB

	var clinicUsers []models.ClinicUser
	if err := db.Preload("Role").
		Joins("JOIN clinic_user_treatments ON clinic_user_treatments.clinic_user_id = clinic_users.id").
		Joins("JOIN clinic_roles ON clinic_roles.id = clinic_users.role_id").
		Where("clinic_users.clinic_id = ? AND clinic_user_treatments.treatment_id = ? AND clinic_users.status = ? AND clinic_roles.name = ?",
			clinicID, treatmentID, "active", models.ClinicRoleDoctor).
		Find(&clinicUsers).Error; err != nil {
		return nil, err
	}

	doctors := make([]DoctorDTO, 0, len(clinicUsers))
	for _, u := range clinicUsers {
		doctors = append(doctors, DoctorDTO{
			ID:       u.ID,
			Name:     u.Name,
			Email:    u.Email,
			Role:     u.Role.Name,
			ClinicID: u.ClinicID,
		})
	}
	return doctors, nil
}

// GetTreatmentsByDoctorAndClinic returns treatments a doctor can perform at a specific clinic
func GetTreatmentsByDoctorAndClinic(doctorID uint64, clinicID uint64) ([]TreatmentWithPriceDTO, error) {
	db := config.DB

	// Find the doctor at this clinic
	var clinicUser models.ClinicUser
	if err := db.Where("id = ? AND clinic_id = ? AND status = ?", doctorID, clinicID, "active").
		First(&clinicUser).Error; err != nil {
		return nil, err
	}

	// Get treatments this doctor can perform
	var userTreatments []models.ClinicUserTreatment
	if err := db.Preload("Treatment").
		Where("clinic_user_id = ?", clinicUser.ID).
		Find(&userTreatments).Error; err != nil {
		return nil, err
	}

	// Cross-reference with clinic's active treatments to get prices
	treatments := make([]TreatmentWithPriceDTO, 0, len(userTreatments))
	for _, ut := range userTreatments {
		var ct models.ClinicTreatment
		if err := db.Where("clinic_id = ? AND treatment_id = ? AND status = ?",
			clinicID, ut.TreatmentID, "active").First(&ct).Error; err != nil {
			continue // clinic doesn't offer this treatment
		}
		treatments = append(treatments, TreatmentWithPriceDTO{
			ID:          ut.Treatment.ID,
			Name:        ut.Treatment.Name,
			Icon:        ut.Treatment.Icon,
			Description: ut.Treatment.Description,
			Price:       ct.Price,
		})
	}
	return treatments, nil
}

// GetClinicsByDoctorAndTreatment returns clinics where a doctor can perform a specific treatment
func GetClinicsByDoctorAndTreatment(doctorID uint64, treatmentID uint) ([]ClinicWithPriceDTO, error) {
	db := config.DB

	// Get all clinic_user records for this email (doctor at multiple clinics)
	var doctor models.ClinicUser
	if err := db.First(&doctor, doctorID).Error; err != nil {
		return nil, err
	}

	// Find all clinic_user records with same email that have this treatment
	var clinicUsers []models.ClinicUser
	if err := db.Preload("Clinic").
		Joins("JOIN clinic_user_treatments ON clinic_user_treatments.clinic_user_id = clinic_users.id").
		Where("clinic_users.email = ? AND clinic_user_treatments.treatment_id = ? AND clinic_users.status = ?",
			doctor.Email, treatmentID, "active").
		Find(&clinicUsers).Error; err != nil {
		return nil, err
	}

	// Also check that the clinic actually offers this treatment
	clinics := make([]ClinicWithPriceDTO, 0)
	for _, u := range clinicUsers {
		if u.Clinic.Status != "active" {
			continue
		}
		// Check if clinic offers this treatment
		var ct models.ClinicTreatment
		if err := db.Where("clinic_id = ? AND treatment_id = ? AND status = ?",
			u.ClinicID, treatmentID, "active").First(&ct).Error; err != nil {
			continue
		}
		clinics = append(clinics, ClinicWithPriceDTO{
			ID:      u.Clinic.ID,
			Name:    u.Clinic.Name,
			Email:   u.Clinic.Email,
			Phone:   u.Clinic.Phone,
			Address: u.Clinic.Address,
			Logo:    u.Clinic.Logo,
			Status:  u.Clinic.Status,
			Price:   ct.Price,
		})
	}
	return clinics, nil
}
