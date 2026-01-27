package services

import (
	"errors"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
	"skinSync/utils"

	"gorm.io/gorm"
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
		Role:         models.ClinicRoleOwner,
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

	// Find clinic user by email with clinic relation
	var clinicUser models.ClinicUser
	if err := db.Preload("Clinic").Where("email = ?", req.Email).First(&clinicUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check if user is active
	if clinicUser.Status != "active" {
		return nil, errors.New("account is inactive")
	}

	// Check if clinic is active
	if clinicUser.Clinic.Status != "active" {
		return nil, errors.New("clinic is inactive or suspended")
	}

	// Verify password
	if !CheckPasswordHash(req.Password, clinicUser.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Update last login
	now := time.Now()
	clinicUser.LastLogin = &now
	db.Save(&clinicUser)

	// Generate tokens (include clinic_id and role in JWT)
	accessToken, err := GenerateClinicJWT(clinicUser.Email, clinicUser.ID, clinicUser.ClinicID, clinicUser.Role)
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
	resp.Data.ClinicUser = resdto.ClinicUserDTO{
		ID:       clinicUser.ID,
		ClinicID: clinicUser.ClinicID,
		Email:    clinicUser.Email,
		Name:     clinicUser.Name,
		Role:     clinicUser.Role,
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

	return resp, nil
}
