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

// SideAreaPriceItem represents price info sent per side area by FE
type SideAreaPriceItem struct {
	SideAreaID  uint     `json:"side_area_id"`
	Price       *float64 `json:"price,omitempty"`
	SyringeSize int      `json:"syringe_size,omitempty"`
}

// BulkSideAreaRequest is the payload shape FE will send
type BulkSideAreaRequest struct {
	TreatmentID    uint                `json:"treatment_id"`
	TreatmentPrice float64             `json:"treatment_price"`
	SideAreas      []SideAreaPriceItem `json:"side_area"`
}

// BulkSideAreaResponseItem represents a side area in the response
type BulkSideAreaResponseItem struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	PerSyringePrice *float64 `json:"per_syringe_price,omitempty"`
}

// BulkSideAreaResponse represents the response for bulk side area upsert
type BulkSideAreaResponse struct {
	ID          uint                       `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Price       float64                    `json:"treatment_price"`
	SideAreas   []BulkSideAreaResponseItem `json:"side_areas"`
}

// CreateClinicSideAreasBulk creates a treatment with side areas for a clinic.
// Returns error if the treatment already exists for this clinic or if any side area already exists.
func CreateClinicSideAreasBulk(req BulkSideAreaRequest, clinicID uint64) (*BulkSideAreaResponse, error) {
	if len(req.SideAreas) == 0 {
		return nil, fmt.Errorf("side_area list is empty")
	}
	db := config.DB

	// Check if treatment already exists for this clinic
	var existingCT models.ClinicTreatment
	if err := db.Where("clinic_id = ? AND treatment_id = ?", clinicID, req.TreatmentID).
		First(&existingCT).Error; err == nil {
		return nil, fmt.Errorf("treatment already exists for this clinic")
	}

	// Get treatment details
	var treatment models.Treatment
	if err := db.First(&treatment, req.TreatmentID).Error; err != nil {
		return nil, fmt.Errorf("treatment id %d not found", req.TreatmentID)
	}

	// Check if any side areas already exist for this clinic+treatment
	for _, it := range req.SideAreas {
		var existing models.ClinicSideArea
		if err := db.Where("clinic_id = ? AND treatment_id = ? AND side_area_id = ?",
			clinicID, req.TreatmentID, it.SideAreaID).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("side_area %d already exists for this clinic and treatment", it.SideAreaID)
		}
	}

	// Create treatment in clinic_treatments
	ct := models.ClinicTreatment{
		ClinicID:    clinicID,
		TreatmentID: req.TreatmentID,
		Status:      "active",
	}
	if req.TreatmentPrice != 0.0 {
		ct.Price = &req.TreatmentPrice
	}
	if err := db.Create(&ct).Error; err != nil {
		return nil, fmt.Errorf("failed to create clinic treatment: %w", err)
	}

	var treatmentPrice float64
	if ct.Price != nil {
		treatmentPrice = *ct.Price
	}

	// Create side areas
	var rows []models.ClinicSideArea
	var responseItems []BulkSideAreaResponseItem

	for _, it := range req.SideAreas {
		var sideArea models.SideArea
		if err := db.First(&sideArea, it.SideAreaID).Error; err != nil {
			return nil, fmt.Errorf("side_area id %d not found", it.SideAreaID)
		}

		if sideArea.TreatmentID != req.TreatmentID {
			return nil, fmt.Errorf("side_area %d does not belong to treatment %d", it.SideAreaID, req.TreatmentID)
		}

		row := models.ClinicSideArea{
			ClinicID:    uint(clinicID),
			TreatmentID: sideArea.TreatmentID,
			AreaID:      sideArea.AreaID,
			SideAreaID:  sideArea.ID,
			SyringeSize: it.SyringeSize,
			Price:       it.Price,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		rows = append(rows, row)

		responseItems = append(responseItems, BulkSideAreaResponseItem{
			ID:              sideArea.ID,
			Name:            sideArea.Name,
			PerSyringePrice: it.Price,
		})
	}

	if err := db.Create(&rows).Error; err != nil {
		return nil, err
	}

	return &BulkSideAreaResponse{
		ID:          treatment.ID,
		Name:        treatment.Name,
		Description: treatment.Description,
		Price:       treatmentPrice,
		SideAreas:   responseItems,
	}, nil
}

// UpdateClinicSideAreasBulk does a full sync for a treatment:
// - Updates price for side areas in the request
// - Adds new side areas not yet in the DB
// - Removes side areas from DB that are NOT in the request (for that treatment + clinic)
func UpdateClinicSideAreasBulk(req BulkSideAreaRequest, clinicID uint64) (*BulkSideAreaResponse, error) {
	if req.TreatmentID == 0 {
		return nil, fmt.Errorf("treatment_id is required")
	}
	db := config.DB

	// Upsert treatment in clinic_treatments (always, so no duplicate rows)
	ct := models.ClinicTreatment{
		ClinicID:    clinicID,
		TreatmentID: req.TreatmentID,
		Status:      "active",
	}
	if req.TreatmentPrice != 0.0 {
		ct.Price = &req.TreatmentPrice
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "treatment_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "status", "updated_at"}),
	}).Create(&ct).Error; err != nil {
		return nil, fmt.Errorf("failed to upsert clinic treatment: %w", err)
	}

	// Fetch the saved treatment price from DB to return in response
	var savedCT models.ClinicTreatment
	db.Where("clinic_id = ? AND treatment_id = ?", clinicID, req.TreatmentID).First(&savedCT)
	var treatmentPrice float64
	if savedCT.Price != nil {
		treatmentPrice = *savedCT.Price
	}

	// Get treatment details
	var treatment models.Treatment
	if err := db.First(&treatment, req.TreatmentID).Error; err != nil {
		return nil, fmt.Errorf("treatment id %d not found", req.TreatmentID)
	}

	// Delete side areas for this clinic+treatment that are NOT in the request
	if len(req.SideAreas) > 0 {
		ids := make([]uint, 0, len(req.SideAreas))
		for _, it := range req.SideAreas {
			ids = append(ids, it.SideAreaID)
		}
		db.Where("clinic_id = ? AND treatment_id = ? AND side_area_id NOT IN ?", clinicID, req.TreatmentID, ids).
			Delete(&models.ClinicSideArea{})
	} else {
		// If empty list, remove all side areas for this treatment
		db.Where("clinic_id = ? AND treatment_id = ?", clinicID, req.TreatmentID).
			Delete(&models.ClinicSideArea{})

		return &BulkSideAreaResponse{
			ID:          treatment.ID,
			Name:        treatment.Name,
			Description: treatment.Description,
			Price:       treatmentPrice,
			SideAreas:   []BulkSideAreaResponseItem{},
		}, nil
	}

	// Upsert the requested side areas
	var rows []models.ClinicSideArea
	var responseItems []BulkSideAreaResponseItem

	for _, it := range req.SideAreas {
		var sideArea models.SideArea
		if err := db.First(&sideArea, it.SideAreaID).Error; err != nil {
			return nil, fmt.Errorf("side_area id %d not found", it.SideAreaID)
		}

		if sideArea.TreatmentID != req.TreatmentID {
			return nil, fmt.Errorf("side_area %d does not belong to treatment %d", it.SideAreaID, req.TreatmentID)
		}

		row := models.ClinicSideArea{
			ClinicID:    uint(clinicID),
			TreatmentID: sideArea.TreatmentID,
			AreaID:      sideArea.AreaID,
			SideAreaID:  sideArea.ID,
			SyringeSize: it.SyringeSize,
			Price:       it.Price,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		rows = append(rows, row)

		responseItems = append(responseItems, BulkSideAreaResponseItem{
			ID:              sideArea.ID,
			Name:            sideArea.Name,
			PerSyringePrice: it.Price,
		})
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "treatment_id"}, {Name: "area_id"}, {Name: "side_area_id"}, {Name: "syringe_size"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "status", "updated_at"}),
	}).Create(&rows).Error; err != nil {
		return nil, err
	}

	return &BulkSideAreaResponse{
		ID:          treatment.ID,
		Name:        treatment.Name,
		Description: treatment.Description,
		Price:       treatmentPrice,
		SideAreas:   responseItems,
	}, nil
}

// UpdateSideAreaPayload is the payload for updating individual side areas
type UpdateSideAreaPayload struct {
	SideAreaID  uint     `json:"side_area_id" validate:"required"`
	Price       *float64 `json:"price,omitempty"`
	SyringeSize *int     `json:"syringe_size,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// UpdateClinicSideAreasIndividual updates price/status for individual clinic side areas
func UpdateClinicSideAreasIndividual(payload []UpdateSideAreaPayload, clinicID uint64) error {
	if len(payload) == 0 {
		return fmt.Errorf("payload is empty")
	}
	db := config.DB

	for _, p := range payload {
		query := db.Model(&models.ClinicSideArea{}).
			Where("clinic_id = ? AND side_area_id = ?", clinicID, p.SideAreaID)

		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		if p.Price != nil {
			updates["price"] = *p.Price
		}
		if p.Status != "" {
			updates["status"] = p.Status
		}
		if p.SyringeSize != nil {
			query = query.Where("syringe_size = ?", *p.SyringeSize)
		}

		result := query.Updates(updates)
		if result.RowsAffected == 0 {
			return fmt.Errorf("clinic side area with side_area_id %d not found for this clinic", p.SideAreaID)
		}
	}

	return nil
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

// // GetTreatmentsByClinic returns all treatments offered by a clinic with their side area prices
func GetTreatmentByClinic(clinicID uint64) (map[string]interface{}, error) {
	db := config.DB

	// Get all clinic treatments assigned to this clinic
	var clinicTreatments []models.ClinicTreatment
	if err := db.Where("clinic_id = ? AND status = ?", clinicID, "active").
		Preload("Treatment").
		Find(&clinicTreatments).Error; err != nil {
		return nil, err
	}

	if len(clinicTreatments) == 0 {
		return map[string]interface{}{
			"is_success": true,
			"message":    "Treatments retrieved successfully",
			"data":       []BulkSideAreaResponse{},
		}, nil
	}

	var treatments []BulkSideAreaResponse

	for _, ct := range clinicTreatments {
		// Get all side areas for this treatment and clinic
		var clinicSideAreas []models.ClinicSideArea
		if err := db.Where("clinic_id = ? AND treatment_id = ? AND status = ?", clinicID, ct.TreatmentID, "active").
			Find(&clinicSideAreas).Error; err != nil {
			continue
		}

		var sideAreas []BulkSideAreaResponseItem
		for _, csa := range clinicSideAreas {
			// Get side area name
			var sideArea models.SideArea
			if err := db.First(&sideArea, csa.SideAreaID).Error; err != nil {
				continue
			}
			sideAreas = append(sideAreas, BulkSideAreaResponseItem{
				ID:              sideArea.ID,
				Name:            sideArea.Name,
				PerSyringePrice: csa.Price,
			})
		}

		treatments = append(treatments, BulkSideAreaResponse{
			ID:          ct.Treatment.ID,
			Name:        ct.Treatment.Name,
			Description: ct.Treatment.Description,
			Price: func() float64 {
				if ct.Price != nil {
					return *ct.Price
				} else {
					return 0
				}
			}(),
			SideAreas: sideAreas,
		})
	}

	return map[string]interface{}{
		"is_success": true,
		"message":    "Treatments retrieved successfully",
		"data":       treatments,
	}, nil
}

// GetClinicRoles returns all clinic roles
func GetClinicRoles() ([]models.ClinicRole, error) {
	db := config.DB
	var roles []models.ClinicRole
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ==================== CLINIC PASSWORD MANAGEMENT ====================

// ClinicForgotPassword generates an OTP and sends it to the clinic user's email
func ClinicForgotPassword(email string) error {
	db := config.DB

	// Check if any clinic user with this email exists
	var clinicUser models.ClinicUser
	if err := db.Where("email = ? AND status = ?", email, "active").First(&clinicUser).Error; err != nil {
		// Don't reveal whether email exists or not
		return nil
	}

	// Reuse existing OTP system with a prefix to separate from login OTPs
	otpKey := "clinic_reset:" + email

	// Check rate limiting
	otpMutex.Lock()
	existingOTP, exists := otpStore[otpKey]
	if exists && time.Since(existingOTP.LastSentAt) < getResendCooldown() {
		otpMutex.Unlock()
		return errors.New("please wait before requesting a new OTP")
	}
	otpMutex.Unlock()

	// Generate OTP
	otpCode, err := GenerateOTP()
	if err != nil {
		return errors.New("failed to generate OTP")
	}

	// Store OTP with 15 minute expiry
	otpMutex.Lock()
	otpStore[otpKey] = OTPData{
		Code:       otpCode,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		LastSentAt: time.Now(),
		Attempts:   0,
	}
	otpMutex.Unlock()

	// Send email
	if err := utils.SendPasswordResetEmail(email, otpCode); err != nil {
		otpMutex.Lock()
		delete(otpStore, otpKey)
		otpMutex.Unlock()
		return errors.New("failed to send reset email")
	}

	return nil
}

// ClinicResetPassword verifies OTP and sets a new password
func ClinicResetPassword(email, otp, newPassword string) error {
	db := config.DB

	otpKey := "clinic_reset:" + email

	// Get OTP from store
	otpMutex.Lock()
	otpData, exists := otpStore[otpKey]
	otpMutex.Unlock()

	if !exists {
		return errors.New("OTP not found, please request a new one")
	}

	// Check expiry
	if time.Now().After(otpData.ExpiresAt) {
		otpMutex.Lock()
		delete(otpStore, otpKey)
		otpMutex.Unlock()
		return errors.New("OTP expired, please request a new one")
	}

	// Check max attempts
	if otpData.Attempts >= getMaxOTPAttempts() {
		otpMutex.Lock()
		delete(otpStore, otpKey)
		otpMutex.Unlock()
		return errors.New("too many failed attempts, please request a new OTP")
	}

	// Verify OTP
	if otpData.Code != otp {
		otpMutex.Lock()
		otpData.Attempts++
		otpStore[otpKey] = otpData
		otpMutex.Unlock()
		remaining := getMaxOTPAttempts() - otpData.Attempts
		return fmt.Errorf("invalid OTP, %d attempts remaining", remaining)
	}

	// OTP verified — delete it
	otpMutex.Lock()
	delete(otpStore, otpKey)
	otpMutex.Unlock()

	// Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password for ALL clinic user records with this email
	if err := db.Model(&models.ClinicUser{}).
		Where("email = ?", email).
		Update("password_hash", hashedPassword).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// ClinicChangePassword verifies old password and sets a new one (authenticated)
func ClinicChangePassword(clinicUserID uint64, oldPassword, newPassword string) error {
	db := config.DB

	// Get the clinic user
	var clinicUser models.ClinicUser
	if err := db.First(&clinicUser, clinicUserID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !CheckPasswordHash(oldPassword, clinicUser.PasswordHash) {
		return errors.New("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password for ALL records with this email (multi-clinic support)
	if err := db.Model(&models.ClinicUser{}).
		Where("email = ?", clinicUser.Email).
		Update("password_hash", hashedPassword).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}
