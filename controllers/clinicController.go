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
