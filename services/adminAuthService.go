package services

import (
	"errors"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

// GenerateAdminJWT generates JWT token for admin/clinic users with role claim
func GenerateAdminJWT(email string, adminID uint64, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":    email,
		"admin_id": adminID,
		"role":     role,
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// AdminRegister creates a new admin or clinic user
func AdminRegister(req reqdto.AdminRegisterRequest) (resdto.BaseResponse, error) {
	db := config.DB
	if db == nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "database not initialized"}, errors.New("database not initialized")
	}

	// Find role by name
	var role models.Role
	if err := db.Where("name = ?", req.RoleName).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resdto.BaseResponse{IsSuccess: false, Message: "invalid role name"}, errors.New("invalid role name")
		}
		return resdto.BaseResponse{IsSuccess: false, Message: "error finding role"}, err
	}

	// Check if email already exists
	var existingUser models.AdminUser
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "email already exists"}, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to hash password"}, err
	}

	// Create admin user
	adminUser := models.AdminUser{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		RoleID:       role.ID,
		Status:       "active",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return resdto.BaseResponse{IsSuccess: false, Message: "failed to create admin user"}, err
	}

	// Load role for response
	db.Preload("Role").First(&adminUser, adminUser.ID)

	return resdto.BaseResponse{
		IsSuccess: true,
		Message:   "admin user created successfully",
		Data:      adminUser,
	}, nil
}

// AdminLogin authenticates admin/clinic user and returns JWT tokens
func AdminLogin(req reqdto.AdminLoginRequest) (*resdto.AdminLoginResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// Find admin user by email with role
	var adminUser models.AdminUser
	if err := db.Preload("Role").Where("email = ?", req.Email).First(&adminUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check if user is active
	if adminUser.Status != "active" {
		return nil, errors.New("account is inactive")
	}

	// Verify password
	if !CheckPasswordHash(req.Password, adminUser.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Update last login
	now := time.Now()
	adminUser.LastLogin = &now
	db.Save(&adminUser)

	// Generate tokens (include role name in JWT)
	accessToken, err := GenerateAdminJWT(adminUser.Email, adminUser.ID, adminUser.Role.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Prepare response
	resp := &resdto.AdminLoginResponse{}
	resp.IsSuccess = true
	resp.Message = "login successful"
	resp.Data.AccessToken = accessToken
	resp.Data.RefreshToken = refreshToken
	resp.Data.AccessExpiresAt = time.Now().Add(accessTokenDuration).Unix()
	resp.Data.RefreshExpiresAt = time.Now().Add(refreshTokenDuration).Unix()
	resp.Data.AdminUser = adminUser

	return resp, nil
}

// GetAdminUserByID retrieves admin user by ID
func GetAdminUserByID(adminID uint64) (*models.AdminUser, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var adminUser models.AdminUser
	if err := db.Preload("Role").First(&adminUser, adminID).Error; err != nil {
		return nil, err
	}

	return &adminUser, nil
}
