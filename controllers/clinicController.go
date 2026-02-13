package controllers

import (
	"net/http"
	"strconv"

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
	if req.Role == "" || req.Name == "" || req.ContactInfo.Email == "" || len(req.Treatments) == 0 {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "role, name, contact_info.email, and treatments are required",
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
// Body: { "treatment_id":51, "area": [ { "area_id":1, "price":75.0 }, ... ] }
func CreateClinicSideAreasFromAreaHandler(c echo.Context) error {
	var req services.AreaPriceRequest
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

	if err := services.UpsertClinicSideAreasFromAreaRequest(req, clinicID); err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to save clinic side areas from areas: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinic side areas saved",
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

// GetTreatmentsByClinicHandler handles GET /clinic/treatments
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
		return c.JSON(http.StatusInternalServerError, resdto.GetClinicTreatmentsResponse{
			BaseResponse: resdto.BaseResponse{
				IsSuccess: false,
				Message:   "Failed to fetch treatments",
			},
		})
	}

	return c.JSON(http.StatusOK, resp)
}
