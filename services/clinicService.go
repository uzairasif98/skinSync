package services

import (
	"errors"
	"fmt"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
	"skinSync/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegisterClinic creates a new clinic and its owner (super_admin only)
func RegisterClinic(req reqdto.RegisterClinicRequest) (resdto.RegisterClinicResponse, error) {
	db := config.DB
	if db == nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "database not initialized"},
		}, errors.New("database not initialized")
	}

	// Check if clinic email already exists
	var existingClinic models.Clinic
	if err := db.Where("email = ?", req.ClinicEmail).First(&existingClinic).Error; err == nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "clinic email already exists"},
		}, errors.New("clinic email already exists")
	}

	// Check if owner email already exists in clinic_users
	var existingUser models.ClinicUser
	if err := db.Where("email = ?", req.OwnerEmail).First(&existingUser).Error; err == nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "owner email already exists"},
		}, errors.New("owner email already exists")
	}

	// Generate secure password for owner
	plainPassword, err := utils.GenerateSecurePassword()
	if err != nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "failed to generate password"},
		}, err
	}

	// Hash password
	hashedPassword, err := HashPassword(plainPassword)
	if err != nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "failed to hash password"},
		}, err
	}

	// Get owner role ID from clinic_roles table
	var ownerRole models.ClinicRole
	if err := db.Where("name = ?", models.ClinicRoleOwner).First(&ownerRole).Error; err != nil {
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "owner role not found - please run database seeding"},
		}, errors.New("owner role not found in clinic_roles table")
	}

	// Start transaction
	tx := db.Begin()

	// Create clinic
	clinic := models.Clinic{
		Name:    req.ClinicName,
		Email:   req.ClinicEmail,
		Phone:   req.ClinicPhone,
		Address: req.ClinicAddress,
		Logo:    req.ClinicLogo,
		Status:  "active",
	}

	if err := tx.Create(&clinic).Error; err != nil {
		tx.Rollback()
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "failed to create clinic"},
		}, err
	}

	// Create clinic owner
	clinicOwner := models.ClinicUser{
		ClinicID:     clinic.ID,
		Email:        req.OwnerEmail,
		PasswordHash: hashedPassword,
		Name:         req.OwnerName,
		RoleID:       ownerRole.ID,
		Status:       "active",
	}

	if err := tx.Create(&clinicOwner).Error; err != nil {
		tx.Rollback()
		return resdto.RegisterClinicResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "failed to create clinic owner"},
		}, err
	}

	// Commit transaction
	tx.Commit()

	// Send email with credentials to owner
	emailErr := utils.SendClinicCredentialsEmail(
		clinicOwner.Email,
		clinicOwner.Name,
		clinic.Name,
		plainPassword,
	)

	message := "Clinic registered successfully. Credentials sent to owner email."
	if emailErr != nil {
		// Log the error but don't fail the registration
		message = "Clinic registered successfully. Email sending failed - please share credentials manually."
	}

	return resdto.RegisterClinicResponse{
		BaseResponse: resdto.BaseResponse{
			IsSuccess: true,
			Message:   message,
		},
		Data: &resdto.RegisterClinicData{
			ClinicID:      clinic.ID,
			ClinicName:    clinic.Name,
			ClinicEmail:   clinic.Email,
			OwnerEmail:    clinicOwner.Email,
			OwnerPassword: plainPassword, // Return plain password (for backup if email fails)
			Status:        clinic.Status,
		},
	}, nil
}

// ClinicLogin authenticates clinic user and returns JWT tokens
func ClinicLogin(req reqdto.ClinicLoginRequest) (*resdto.ClinicLoginResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Find all clinic users with this email
	var clinicUsers []models.ClinicUser
	if err := db.Preload("Clinic").Preload("Role").Where("email = ?", req.Email).Find(&clinicUsers).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}

	if len(clinicUsers) == 0 {
		return nil, errors.New("invalid email or password")
	}

	// Verify password against first record (all records share same password)
	if !CheckPasswordHash(req.Password, clinicUsers[0].PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Filter only active users at active clinics
	var activeUsers []models.ClinicUser
	for _, u := range clinicUsers {
		if u.Status == "active" && u.Clinic.Status == "active" {
			activeUsers = append(activeUsers, u)
		}
	}

	if len(activeUsers) == 0 {
		return nil, errors.New("account is inactive or clinic is suspended")
	}

	// If multiple clinics and no clinic_id provided → return clinic selection list
	if len(activeUsers) > 1 && req.ClinicID == nil {
		clinicList := make([]resdto.ClinicSelectionDTO, 0, len(activeUsers))
		for _, u := range activeUsers {
			clinicList = append(clinicList, resdto.ClinicSelectionDTO{
				ID:   u.Clinic.ID,
				Name: u.Clinic.Name,
				Logo: u.Clinic.Logo,
				Role: u.Role.Name,
			})
		}

		resp := &resdto.ClinicLoginResponse{}
		resp.IsSuccess = true
		resp.Message = "multiple clinics found, please select one"
		resp.Data.RequiresClinicSelection = true
		resp.Data.Clinics = clinicList
		return resp, nil
	}

	// Determine which user to log in
	var clinicUser models.ClinicUser
	if req.ClinicID != nil {
		// Find user at specific clinic
		found := false
		for _, u := range activeUsers {
			if u.ClinicID == *req.ClinicID {
				clinicUser = u
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("user not found at this clinic")
		}
	} else {
		// Single clinic - login directly
		clinicUser = activeUsers[0]
	}

	// Update last login
	now := time.Now()
	clinicUser.LastLogin = &now
	db.Save(&clinicUser)

	// Generate tokens
	accessToken, err := GenerateClinicJWT(clinicUser.Email, clinicUser.ID, clinicUser.ClinicID, clinicUser.Role.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Prepare response
	resp := &resdto.ClinicLoginResponse{}
	resp.IsSuccess = true
	resp.Message = "login successful"
	resp.Data.AccessToken = accessToken
	resp.Data.RefreshToken = refreshToken
	resp.Data.AccessExpiresAt = time.Now().Add(accessTokenDuration).Unix()
	resp.Data.RefreshExpiresAt = time.Now().Add(refreshTokenDuration).Unix()
	clinicUserDTO := resdto.ClinicUserDTO{
		ID:       clinicUser.ID,
		ClinicID: clinicUser.ClinicID,
		Email:    clinicUser.Email,
		Name:     clinicUser.Name,
		Role:     clinicUser.Role.Name,
		Status:   clinicUser.Status,
		Clinic: &resdto.ClinicDTO{
			ID:      clinicUser.Clinic.ID,
			Name:    clinicUser.Clinic.Name,
			Email:   clinicUser.Clinic.Email,
			Phone:   clinicUser.Clinic.Phone,
			Address: clinicUser.Clinic.Address,
			Logo:    clinicUser.Clinic.Logo,
			Status:  clinicUser.Clinic.Status,
		},
	}
	resp.Data.ClinicUser = &clinicUserDTO

	return resp, nil
}

// RegisterClinicUser creates a new clinic user (staff) - only owner can do this
func RegisterClinicUser(req reqdto.RegisterClinicUserRequest, clinicID uint64) (*resdto.RegisterClinicUserResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Validate the role exists
	var role models.ClinicRole
	if err := db.First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid role_id")
		}
		return nil, err
	}

	// Prevent creating another owner
	if role.Name == models.ClinicRoleOwner {
		return nil, errors.New("cannot create another owner for the clinic")
	}

	// Check if email already exists at THIS clinic
	var existingUser models.ClinicUser
	if err := db.Where("email = ? AND clinic_id = ?", req.Email, clinicID).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists at this clinic")
	}

	// Check if email exists at another clinic - reuse password
	var existingElsewhere models.ClinicUser
	var plainPassword string
	var hashedPassword string

	if err := db.Where("email = ?", req.Email).First(&existingElsewhere).Error; err == nil {
		// Same person at another clinic - reuse password hash
		hashedPassword = existingElsewhere.PasswordHash
		plainPassword = "" // Don't return password since it's the same as existing
	} else {
		// New user - generate password
		var genErr error
		plainPassword, genErr = utils.GenerateSecurePassword()
		if genErr != nil {
			return nil, errors.New("failed to generate password")
		}
		var hashErr error
		hashedPassword, hashErr = HashPassword(plainPassword)
		if hashErr != nil {
			return nil, errors.New("failed to hash password")
		}
	}

	// Create clinic user
	clinicUser := models.ClinicUser{
		ClinicID:     clinicID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		RoleID:       req.RoleID,
		Status:       "active",
	}

	if err := db.Create(&clinicUser).Error; err != nil {
		return nil, errors.New("failed to create clinic user")
	}

	// Get clinic name for email
	var clinic models.Clinic
	db.First(&clinic, clinicID)

	message := "Clinic user registered successfully."
	if plainPassword != "" {
		// New user - send email with credentials
		emailErr := utils.SendClinicCredentialsEmail(
			clinicUser.Email,
			clinicUser.Name,
			clinic.Name,
			plainPassword,
		)
		message = "Clinic user registered successfully. Credentials sent to email."
		if emailErr != nil {
			message = "Clinic user registered successfully. Email sending failed - please share credentials manually."
		}
	} else {
		message = "Clinic user registered successfully. User already exists - same password as existing account."
	}

	return &resdto.RegisterClinicUserResponse{
		BaseResponse: resdto.BaseResponse{
			IsSuccess: true,
			Message:   message,
		},
		Data: &resdto.RegisterClinicUserData{
			ID:       clinicUser.ID,
			ClinicID: clinicUser.ClinicID,
			Email:    clinicUser.Email,
			Name:     clinicUser.Name,
			Role:     role.Name,
			Password: plainPassword,
			Status:   clinicUser.Status,
		},
	}, nil
}

// RegisterDoctorWithTreatments creates a doctor/injector and assigns them to specific side areas
func RegisterDoctorWithTreatments(req reqdto.RegisterDoctorRequest, clinicID uint64) (*resdto.RegisterClinicUserResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Map role name to role_id
	var roleName string
	if req.Role == "doctor" {
		roleName = models.ClinicRoleDoctor
	} else if req.Role == "injector" {
		roleName = models.ClinicRoleInjector
	} else {
		return nil, errors.New("invalid role - must be 'doctor' or 'injector'")
	}

	// Get role ID
	var role models.ClinicRole
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, errors.New("role not found in database")
	}

	// Check if email already exists at THIS clinic
	var existingUser models.ClinicUser
	if err := db.Where("email = ? AND clinic_id = ?", req.ContactInfo.Email, clinicID).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists at this clinic")
	}

	// Check if email exists at another clinic - reuse password
	var existingElsewhere models.ClinicUser
	var plainPassword string
	var hashedPassword string

	if err := db.Where("email = ?", req.ContactInfo.Email).First(&existingElsewhere).Error; err == nil {
		// Same person at another clinic - reuse password hash
		hashedPassword = existingElsewhere.PasswordHash
		plainPassword = "" // Don't return password since it's the same as existing
	} else {
		// New user - generate password
		var genErr error
		plainPassword, genErr = utils.GenerateSecurePassword()
		if genErr != nil {
			return nil, errors.New("failed to generate password")
		}
		var hashErr error
		hashedPassword, hashErr = HashPassword(plainPassword)
		if hashErr != nil {
			return nil, errors.New("failed to hash password")
		}
	}

	// Start transaction
	tx := db.Begin()

	// Create clinic user
	clinicUser := models.ClinicUser{
		ClinicID:     clinicID,
		Email:        req.ContactInfo.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		RoleID:       role.ID,
		Status:       "active",
	}

	if err := tx.Create(&clinicUser).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create clinic user")
	}

	// Process treatments and side areas
	var sideAreaRecords []models.ClinicUserSideArea

	for _, treatment := range req.Treatments {
		// For each side area ID in the treatment
		for _, sideAreaID := range treatment.TreatmentsSubSecID {
			// Lookup side area to get area_id and treatment_id
			var sideArea models.SideArea
			if err := tx.First(&sideArea, sideAreaID).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("side_area id %d not found", sideAreaID)
			}

			// Verify the side area belongs to the specified treatment
			if sideArea.TreatmentID != treatment.TreatmentID {
				tx.Rollback()
				return nil, fmt.Errorf("side_area %d does not belong to treatment %d", sideAreaID, treatment.TreatmentID)
			}

			// Create ClinicUserSideArea record
			sideAreaRecord := models.ClinicUserSideArea{
				ClinicUserID: clinicUser.ID,
				ClinicID:     clinicID,
				TreatmentID:  sideArea.TreatmentID,
				AreaID:       sideArea.AreaID,
				SideAreaID:   sideArea.ID,
			}

			sideAreaRecords = append(sideAreaRecords, sideAreaRecord)
		}
	}

	// Batch insert side area records
	if len(sideAreaRecords) > 0 {
		if err := tx.Create(&sideAreaRecords).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to assign side areas to user")
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction")
	}

	// Get clinic name for email
	var clinic models.Clinic
	db.First(&clinic, clinicID)

	message := "Doctor registered successfully."
	if plainPassword != "" {
		// New user - send email with credentials
		emailErr := utils.SendClinicCredentialsEmail(
			clinicUser.Email,
			clinicUser.Name,
			clinic.Name,
			plainPassword,
		)
		message = "Doctor registered successfully. Credentials sent to email."
		if emailErr != nil {
			message = "Doctor registered successfully. Email sending failed - please share credentials manually."
		}
	} else {
		message = "Doctor registered successfully. User already exists - same password as existing account."
	}

	return &resdto.RegisterClinicUserResponse{
		BaseResponse: resdto.BaseResponse{
			IsSuccess: true,
			Message:   message,
		},
		Data: &resdto.RegisterClinicUserData{
			ID:       clinicUser.ID,
			ClinicID: clinicUser.ClinicID,
			Email:    clinicUser.Email,
			Name:     clinicUser.Name,
			Role:     role.Name,
			Password: plainPassword,
			Status:   clinicUser.Status,
		},
	}, nil
}

// SideAreaPayload is the payload shape when frontend sends side_area_id
type SideAreaPayload struct {
	ClinicID    uint64   `json:"clinic_id"`
	SideAreaID  uint     `json:"side_area_id"`
	SyringeSize int      `json:"syringe_size,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// UpsertClinicSideAreasFromSideArea accepts payloads where frontend sends side_area_id.
// It looks up the SideArea to get AreaID and TreatmentID and upserts into clinic_side_areas.
func UpsertClinicSideAreasFromSideArea(payload []SideAreaPayload) error {
	if len(payload) == 0 {
		return nil
	}

	db := config.DB

	var rows []models.ClinicSideArea
	for _, p := range payload {
		// ensure side area exists and fetch its area_id and treatment_id
		var sa models.SideArea
		if err := db.First(&sa, p.SideAreaID).Error; err != nil {
			return fmt.Errorf("side_area id %d not found: %w", p.SideAreaID, err)
		}

		row := models.ClinicSideArea{
			ClinicID:    uint(p.ClinicID),
			TreatmentID: sa.TreatmentID,
			AreaID:      sa.AreaID,
			SideAreaID:  sa.ID,
			SyringeSize: p.SyringeSize,
			Price:       p.Price,
			Status:      p.Status,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		rows = append(rows, row)
	}

	// upsert on (clinic_id, treatment_id, area_id, side_area_id, syringe_size)
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "treatment_id"}, {Name: "area_id"}, {Name: "side_area_id"}, {Name: "syringe_size"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "status", "updated_at"}),
	}).Create(&rows).Error
}

// (Removed separate syringe price upsert) Use UpsertClinicSideAreasFromSideArea with SyringeSize to store per-size prices in clinic_side_areas.

// AreaPriceItem represents price info sent per area by FE
type AreaPriceItem struct {
	AreaID uint     `json:"area_id"`
	Price  *float64 `json:"price,omitempty"`
	// optional syringe size if FE wants to set per-size price for all side areas
	SyringeSize int `json:"syringe_size,omitempty"`
}

// AreaPriceRequest is the payload shape FE will send
type AreaPriceRequest struct {
	TreatmentID uint            `json:"treatment_id"`
	Areas       []AreaPriceItem `json:"area"`
}

// UpsertClinicSideAreasFromAreaRequest accepts a payload where FE sends treatment_id and a list
// of areas with prices. For each area it finds all SideAreas and upserts ClinicSideArea rows
// (one row per SideArea). SyringeSize from the item is applied to all side areas (0 if omitted).
func UpsertClinicSideAreasFromAreaRequest(req AreaPriceRequest, clinicID uint64) error {
	if len(req.Areas) == 0 {
		return nil
	}
	db := config.DB

	var rows []models.ClinicSideArea
	for _, it := range req.Areas {
		var sideAreas []models.SideArea
		if err := db.Where("treatment_id = ? AND area_id = ?", req.TreatmentID, it.AreaID).Find(&sideAreas).Error; err != nil {
			return err
		}

		for _, sa := range sideAreas {
			row := models.ClinicSideArea{
				ClinicID:    uint(clinicID),
				TreatmentID: sa.TreatmentID,
				AreaID:      sa.AreaID,
				SideAreaID:  sa.ID,
				SyringeSize: it.SyringeSize,
				Price:       it.Price,
				Status:      "active",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "treatment_id"}, {Name: "area_id"}, {Name: "side_area_id"}, {Name: "syringe_size"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "status", "updated_at"}),
	}).Create(&rows).Error
}

// GetSideAreasByTreatment returns all side areas for a given treatment ID
func GetSideAreasByTreatment(treatmentID uint) (resdto.SideAreasResponse, error) {
	db := config.DB

	var sideAreas []models.SideArea

	err := db.Where("treatment_id = ?", treatmentID).Find(&sideAreas).Error
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

// GetTreatmentsByClinic returns all treatments offered by a clinic with their side area prices
func GetTreatmentByClinic(clinicID uint64) (resdto.GetClinicTreatmentsResponse, error) {
	db := config.DB

	// Get all clinic treatments
	var clinicTreatments []models.ClinicTreatment
	err := db.Where("clinic_id = ? AND status = ?", clinicID, "active").
		Preload("Treatment").
		Find(&clinicTreatments).Error
	if err != nil {
		return resdto.GetClinicTreatmentsResponse{
			BaseResponse: resdto.BaseResponse{IsSuccess: false, Message: "Failed to fetch treatments"},
		}, err
	}

	var treatmentDTOs []resdto.ClinicTreatmentDTO

	// For each treatment, get areas and side areas with prices
	for _, ct := range clinicTreatments {
		// Get areas for this treatment
		var areas []models.Area
		err := db.Where("treatment_id = ?", ct.TreatmentID).Find(&areas).Error
		if err != nil {
			continue
		}

		var areaDTOs []resdto.ClinicAreaDTO

		for _, area := range areas {
			// Get side areas with clinic prices
			var clinicSideAreas []models.ClinicSideArea
			err := db.Where("clinic_id = ? AND treatment_id = ? AND area_id = ? AND status = ?",
				clinicID, ct.TreatmentID, area.ID, "active").
				Find(&clinicSideAreas).Error
			if err != nil {
				continue
			}

			var sideAreaDTOs []resdto.ClinicSideAreaPriceDTO

			for _, csa := range clinicSideAreas {
				// Get side area details
				var sideArea models.SideArea
				err := db.Where("id = ?", csa.SideAreaID).First(&sideArea).Error
				if err != nil {
					continue
				}

				// Generate syringe options
				var syringeOptions []int
				for i := sideArea.MinSyringe; i <= sideArea.MaxSyringe; i++ {
					syringeOptions = append(syringeOptions, i)
				}

				sideAreaDTOs = append(sideAreaDTOs, resdto.ClinicSideAreaPriceDTO{
					ID:          csa.ID,
					SideAreaID:  sideArea.ID,
					Name:        sideArea.Name,
					Icon:        sideArea.Icon,
					Description: sideArea.Description,
					MinSyringe:  sideArea.MinSyringe,
					MaxSyringe:  sideArea.MaxSyringe,
					Price:       csa.Price,
					Status:      csa.Status,
				})
			}

			if len(sideAreaDTOs) > 0 {
				areaDTOs = append(areaDTOs, resdto.ClinicAreaDTO{
					ID:          area.ID,
					AreaID:      area.ID,
					Name:        area.Name,
					Icon:        area.Icon,
					Description: area.Description,
					SideAreas:   sideAreaDTOs,
				})
			}
		}

		treatmentDTOs = append(treatmentDTOs, resdto.ClinicTreatmentDTO{
			ID:          ct.TreatmentID,
			Name:        ct.Treatment.Name,
			Icon:        ct.Treatment.Icon,
			Description: ct.Treatment.Description,
			Price:       ct.Price,
			Areas:       areaDTOs,
			Status:      ct.Status,
		})
	}

	return resdto.GetClinicTreatmentsResponse{
		BaseResponse: resdto.BaseResponse{IsSuccess: true, Message: "Treatments retrieved successfully"},
		Data:         treatmentDTOs,
	}, nil
}
