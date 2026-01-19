package controllers

import (
	"net/http"

	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/labstack/echo/v4"
)

// AdminRegisterHandler handles admin/clinic user registration
func AdminRegisterHandler(c echo.Context) error {
	var req reqdto.AdminRegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Name == "" || req.RoleName == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "email, password, name, and role_name are required",
		})
	}

	resp, err := services.AdminRegister(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusCreated, resp)
}

// AdminLoginHandler handles admin/clinic user login
func AdminLoginHandler(c echo.Context) error {
	var req reqdto.AdminLoginRequest
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

	resp, err := services.AdminLogin(req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetAdminMeHandler returns current admin user info
func GetAdminMeHandler(c echo.Context) error {
	// Get admin_id from context (set by AdminAuthMiddleware)
	adminID, ok := c.Get("admin_id").(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "admin_id not found in context",
		})
	}

	adminUser, err := services.GetAdminUserByID(uint64(adminID))
	if err != nil {
		return c.JSON(http.StatusNotFound, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "admin user not found",
		})
	}

	return c.JSON(http.StatusOK, resdto.AdminInfoResponse{
		BaseResponse: resdto.BaseResponse{
			IsSuccess: true,
			Message:   "",
		},
		Data: *adminUser,
	})
}
