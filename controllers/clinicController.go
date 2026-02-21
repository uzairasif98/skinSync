package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// RegisterClinicHandler handles clinic registration by super_admin
func RegisterClinicHandler(c echo.Context) error {
	var req reqdto.RegisterClinicRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Validate required fields
	if req.ClinicName == "" || req.ClinicEmail == "" || req.OwnerName == "" || req.OwnerEmail == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_name, clinic_email, owner_name, and owner_email are required",
		})
	}

	resp, err := services.RegisterClinic(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusCreated, resp)
}

// ClinicLoginHandler handles clinic user login
func ClinicLoginHandler(c echo.Context) error {
	var req reqdto.ClinicLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "email and password are required",
		})
	}

	resp, err := services.ClinicLogin(req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// RegisterClinicUserHandler handles clinic user (staff) registration by owner
func RegisterClinicUserHandler(c echo.Context) error {
	// Get clinic_id from context (set by ClinicAuthMiddleware)
	clinicIDFloat, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDFloat)

	var req reqdto.RegisterClinicUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Validate required fields
	if req.Email == "" || req.Name == "" || req.RoleID == 0 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "email, name, and role_id are required",
		})
	}

	resp, err := services.RegisterClinicUser(req, clinicID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// RegisterDoctorHandler handles doctor/injector registration with treatment side areas
func RegisterDoctorHandler(c echo.Context) error {
	// Get clinic_id from context (set by ClinicAuthMiddleware)
	clinicIDFloat, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDFloat)

	var req reqdto.RegisterDoctorRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Validate required fields
	if req.Role == "" || req.Name == "" || req.ContactInfo.Email == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "role, name, and contact_info.email are required",
		})
	}

	// Validate role is doctor or injector
	if req.Role != "doctor" && req.Role != "injector" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "role must be 'doctor' or 'injector'",
		})
	}

	resp, err := services.RegisterDoctorWithTreatments(req, clinicID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// AssignDoctorTreatmentsHandler handles POST /clinic/doctors/treatments
func AssignDoctorTreatmentsHandler(c echo.Context) error {
	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	var req reqdto.AssignDoctorTreatmentsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload: " + err.Error(),
		})
	}

	if req.ClinicUserID == 0 || len(req.Treatments) == 0 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_user_id and treatments are required",
		})
	}

	if err := services.AssignDoctorTreatments(req, clinicID); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "treatments assigned successfully",
	})
}

// GetDoctorsHandler handles GET /clinic/doctors
func GetDoctorsHandler(c echo.Context) error {
	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	doctors, err := services.GetDoctorsByClinic(clinicID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch doctors: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_success": true,
		"message":    "doctors retrieved successfully",
		"data":       doctors,
	})
}

// GetDoctorDetailHandler handles GET /clinic/doctors/:id
func GetDoctorDetailHandler(c echo.Context) error {
	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	doctorID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid doctor id",
		})
	}

	detail, err := services.GetDoctorDetailByID(doctorID, clinicID)
	if err != nil {
		return c.JSON(http.StatusNotFound, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_success": true,
		"message":    "doctor detail retrieved successfully",
		"data":       detail,
	})
}

// CreateClinicSideAreasFromSideAreaHandler handles POST /clinic/side-areas
// Body: [{ "clinic_id":1, "side_area_id":4, "price":50.0, "status":"active" }, ...]
// Frontend sends side_area_id; handler will lookup the side-area to get area_id and treatment_id
func CreateClinicSideAreasFromSideAreaHandler(c echo.Context) error {
	var payload []services.SideAreaPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload: " + err.Error(),
		})
	}

	// fill clinic_id from context if middleware provides it
	if clinicIDf, ok := c.Get("clinic_id").(float64); ok {
		clinicID := uint64(clinicIDf)
		for i := range payload {
			if payload[i].ClinicID == 0 {
				payload[i].ClinicID = clinicID
			}
		}
	}

	// call service to upsert (service will resolve side-area -> area/treatment)
	if err := services.UpsertClinicSideAreasFromSideArea(payload); err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to save clinic side areas: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinic side areas saved",
	})
}

// CreateClinicSideAreasFromAreaHandler handles POST /clinic/side-areas/bulk
// Body: { "treatments": [ { "treatment_id":6, "side_area": [ { "side_area_id":8, "price":150.0 }, ... ] }, ... ] }
func CreateClinicSideAreasFromAreaHandler(c echo.Context) error {
	var req services.BulkSideAreaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload: " + err.Error(),
		})
	}

	// get clinic id from context (ClinicAuthMiddleware)
	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	resp, err := services.CreateClinicSideAreasBulk(req, clinicID)
	if err != nil {
		// Check if it's an "already exists" error
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			return c.JSON(http.StatusConflict, resdto.BaseResponse{
				IsSuccess: false,
				Message:   errMsg,
			})
		}
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to save clinic side areas: " + errMsg,
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"is_success": true,
		"message":    "clinic side areas created",
		"data":       resp,
	})
}

// UpdateClinicSideAreasBulkHandler handles PUT /clinic/side-areas/bulk
// Full sync: updates existing, adds new, removes side areas not in the request
func UpdateClinicSideAreasBulkHandler(c echo.Context) error {
	var req services.BulkSideAreaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload: " + err.Error(),
		})
	}

	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	resp, err := services.UpdateClinicSideAreasBulk(req, clinicID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to update clinic side areas: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_success": true,
		"message":    "clinic side areas updated",
		"data":       resp,
	})
}

// UpdateClinicSideAreasHandler handles PUT /clinic/side-areas
// Updates price/status of individual side area records
func UpdateClinicSideAreasHandler(c echo.Context) error {
	var payload []services.UpdateSideAreaPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload: " + err.Error(),
		})
	}

	clinicIDf, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDf)

	if err := services.UpdateClinicSideAreasIndividual(payload, clinicID); err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to update: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinic side areas updated",
	})
}

// GetSideAreasByTreatmentHandler handles GET /clinic/side-areas/treatment/:treatmentId
func GetSideAreasByTreatmentHandler(c echo.Context) error {
	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Invalid treatment ID",
			Data:      nil,
		})
	}

	resp, err := services.GetSideAreasByTreatment(uint(treatmentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Failed to fetch side areas",
			Data:      nil,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// GetClinicRolesHandler handles GET /clinic/roles
func GetClinicRolesHandler(c echo.Context) error {
	roles, err := services.GetClinicRoles()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "Failed to fetch roles",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_success": true,
		"message":    "Roles retrieved successfully",
		"data":       roles,
	})
}

// GetTreatmentByClinicHandler handles GET /clinic/treatments
func GetTreatmentByClinicHandler(c echo.Context) error {
	// Get clinic_id from context (set by ClinicAuthMiddleware)
	clinicIDFloat, ok := c.Get("clinic_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_id not found in context",
		})
	}
	clinicID := uint64(clinicIDFloat)

	resp, err := services.GetTreatmentByClinic(clinicID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "Failed to fetch treatments",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// ClinicForgotPasswordHandler handles POST /clinic/forgot-password
func ClinicForgotPasswordHandler(c echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload",
		})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "email is required",
		})
	}

	if err := services.ClinicForgotPassword(req.Email); err != nil {
		return c.JSON(http.StatusTooManyRequests, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Always return success (don't reveal if email exists)
	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "if this email is registered, a password reset OTP has been sent",
	})
}

// ClinicResetPasswordHandler handles POST /clinic/reset-password
func ClinicResetPasswordHandler(c echo.Context) error {
	var req struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload",
		})
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "email, otp, and new_password are required",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "password must be at least 6 characters",
		})
	}

	if err := services.ClinicResetPassword(req.Email, req.OTP, req.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "password reset successfully",
	})
}

// ClinicChangePasswordHandler handles POST /clinic/change-password (requires auth)
func ClinicChangePasswordHandler(c echo.Context) error {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid payload",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "old_password and new_password are required",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "new password must be at least 6 characters",
		})
	}

	clinicUserIDf, ok := c.Get("clinic_user_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "clinic_user_id not found in context",
		})
	}
	clinicUserID := uint64(clinicUserIDf)

	if err := services.ClinicChangePassword(clinicUserID, req.OldPassword, req.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "password changed successfully",
	})
}

// ClinicLogoutHandler handles POST /clinic/logout
func ClinicLogoutHandler(c echo.Context) error {
	// Get the raw token from the Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid authorization header",
		})
	}
	tokenString := parts[1]

	// Parse token to get expiry time
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("skinSync"), nil
	})

	var expiresAt time.Time
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expiresAt = time.Unix(int64(exp), 0)
		} else {
			expiresAt = time.Now().Add(24 * time.Hour)
		}
	} else {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	// Add token to blacklist
	services.BlacklistToken(tokenString, expiresAt)

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "logged out successfully",
	})
}
