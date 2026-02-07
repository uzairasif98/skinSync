package controllers

import (
	"net/http"

	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/services"

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
