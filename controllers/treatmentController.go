package controllers

import (
	"net/http"
	"strconv"

	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/labstack/echo/v4"
)

// GetTreatmentMastersHandler handles GET /api/treatments/masters
func GetTreatmentMastersHandler(c echo.Context) error {
	resp, err := services.GetTreatmentMasters()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.TreatmentMastersResponse{
			IsSuccess: false,
			Message:   "Failed to fetch treatments",
			Data:      nil,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// GetAreasHandler handles GET /api/treatments/:id/areas
func GetAreasHandler(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.AreasResponse{
			IsSuccess: false,
			Message:   "Invalid treatment ID",
			Data:      nil,
		})
	}

	resp, err := services.GetAreasByTreatment(uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.AreasResponse{
			IsSuccess: false,
			Message:   "Failed to fetch areas",
			Data:      nil,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// GetSideAreasHandler handles GET /api/treatments/:treatmentId/areas/:areaId/sideareas
func GetSideAreasHandler(c echo.Context) error {
	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Invalid treatment ID",
			Data:      nil,
		})
	}

	areaID, err := strconv.ParseUint(c.Param("areaId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Invalid area ID",
			Data:      nil,
		})
	}

	resp, err := services.GetSideAreas(uint(treatmentID), uint(areaID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.SideAreasResponse{
			IsSuccess: false,
			Message:   "Failed to fetch side areas",
			Data:      nil,
		})
	}
	return c.JSON(http.StatusOK, resp)
}
