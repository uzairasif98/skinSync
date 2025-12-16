package controllers

import (
	"net/http"

	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetOnboardingMastersHandler(c echo.Context) error {
	m, err := services.GetOnboardingMasters()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, m)
}

// SaveOnboardingHandler saves user's selection(s) for a given onboarding step
func SaveOnboardingHandler(c echo.Context) error {
	var req reqdto.OnboardingAnswerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	// extract user id from context (set by auth middleware)
	uid := c.Get("user_id")
	var userID uint64
	switch v := uid.(type) {
	case float64:
		userID = uint64(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{IsSuccess: false, Message: "invalid user id in token"})
		}
		userID = id
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	default:
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{IsSuccess: false, Message: "invalid user id in token"})
	}

	if err := services.SaveOnboardingAnswer(userID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resdto.BaseResponse{IsSuccess: true, Message: "saved"})
}

// GetUserOnboardingHandler returns user's saved onboarding answers and progress
func GetUserOnboardingHandler(c echo.Context) error {
	uid := c.Get("user_id")
	var userID uint64
	switch v := uid.(type) {
	case float64:
		userID = uint64(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{IsSuccess: false, Message: "invalid user id in token"})
		}
		userID = id
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	default:
		return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{IsSuccess: false, Message: "invalid user id in token"})
	}

	resp, err := services.GetUserOnboarding(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
