package controllers

import (
	"net/http"
	"strconv"

	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/labstack/echo/v4"
)

// GetAllClinicsHandler returns all active clinics
func GetAllClinicsHandler(c echo.Context) error {
	clinics, err := services.GetAllClinics()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch clinics",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinics fetched successfully",
		Data:      clinics,
	})
}

// GetAllDoctorsHandler returns all doctors
func GetAllDoctorsHandler(c echo.Context) error {
	doctors, err := services.GetAllDoctors()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch doctors",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "doctors fetched successfully",
		Data:      doctors,
	})
}

// GetTreatmentsByClinicHandler returns treatments offered by a clinic
func GetTreatmentsByClinicHandler(c echo.Context) error {
	clinicID, err := strconv.ParseUint(c.Param("clinicId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid clinic_id",
		})
	}

	treatments, err := services.GetTreatmentsByClinic(clinicID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch treatments",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "treatments fetched successfully",
		Data:      treatments,
	})
}

// GetClinicsByTreatmentHandler returns clinics offering a treatment
func GetClinicsByTreatmentHandler(c echo.Context) error {
	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid treatment_id",
		})
	}

	clinics, err := services.GetClinicsByTreatment(uint(treatmentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch clinics",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinics fetched successfully",
		Data:      clinics,
	})
}

// GetTreatmentsByDoctorHandler returns treatments a doctor can perform
func GetTreatmentsByDoctorHandler(c echo.Context) error {
	doctorID, err := strconv.ParseUint(c.Param("doctorId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid doctor_id",
		})
	}

	treatments, err := services.GetTreatmentsByDoctor(doctorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch treatments",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "treatments fetched successfully",
		Data:      treatments,
	})
}

// GetDoctorsByClinicAndTreatmentHandler returns doctors for a clinic+treatment combo
func GetDoctorsByClinicAndTreatmentHandler(c echo.Context) error {
	clinicID, err := strconv.ParseUint(c.Param("clinicId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid clinic_id",
		})
	}

	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid treatment_id",
		})
	}

	doctors, err := services.GetDoctorsByClinicAndTreatment(clinicID, uint(treatmentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch doctors",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "doctors fetched successfully",
		Data:      doctors,
	})
}

// GetTreatmentsByDoctorAndClinicHandler returns treatments a doctor can perform at a specific clinic
func GetTreatmentsByDoctorAndClinicHandler(c echo.Context) error {
	doctorID, err := strconv.ParseUint(c.Param("doctorId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid doctor_id",
		})
	}

	clinicID, err := strconv.ParseUint(c.Param("clinicId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid clinic_id",
		})
	}

	treatments, err := services.GetTreatmentsByDoctorAndClinic(doctorID, clinicID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch treatments",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "treatments fetched successfully",
		Data:      treatments,
	})
}

// GetClinicsByDoctorAndTreatmentHandler returns clinics where doctor does a treatment
func GetClinicsByDoctorAndTreatmentHandler(c echo.Context) error {
	doctorID, err := strconv.ParseUint(c.Param("doctorId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid doctor_id",
		})
	}

	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "invalid treatment_id",
		})
	}

	clinics, err := services.GetClinicsByDoctorAndTreatment(doctorID, uint(treatmentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{
			IsSuccess: false,
			Message:   "failed to fetch clinics",
		})
	}

	return c.JSON(http.StatusOK, resdto.BaseResponse{
		IsSuccess: true,
		Message:   "clinics fetched successfully",
		Data:      clinics,
	})
}
