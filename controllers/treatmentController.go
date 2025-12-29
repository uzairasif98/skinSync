package controllers

import (
	"net/http"

	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/labstack/echo/v4"
)

// GetTreatmentMastersHandler handles GET /api/treatments/masters
func GetTreatmentMastersHandler(c echo.Context) error {
	resp, err := services.GetTreatmentMasters()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}
