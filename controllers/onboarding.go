package controllers

import (
	"net/http"

	admindto "skinSync/dto/request"
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

// AdminCreateQuestionHandler allows admin to create a question with options
func AdminCreateQuestionHandler(c echo.Context) error {
	var req admindto.CreateQuestionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	base, err := services.CreateOnboardingQuestion(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

// AdminAddOptionsHandler allows admin to add options to an existing question
func AdminAddOptionsHandler(c echo.Context) error {
	qid := c.Param("id")
	var req admindto.AddOptionsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	base, err := services.AddOptionsToQuestion(qid, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

// AdminUpdateQuestionHandler updates question text and/or options
func AdminUpdateQuestionHandler(c echo.Context) error {
	qid := c.Param("id")
	var req admindto.UpdateQuestionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	base, err := services.UpdateOnboardingQuestion(qid, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

// AdminDeleteQuestionHandler deletes a question and its options
func AdminDeleteQuestionHandler(c echo.Context) error {
	qid := c.Param("id")
	base, err := services.DeleteOnboardingQuestion(qid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

// AdminDeleteOptionHandler deletes a single option under a question
func AdminDeleteOptionHandler(c echo.Context) error {
	qid := c.Param("qid")
	oid := c.Param("optionId")
	base, err := services.DeleteOnboardingOption(qid, oid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
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

	base, err := services.SaveOnboardingAnswer(userID, req)
	if err != nil {
		// service already provided base with message
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

// SaveProfileHandler saves user's profile information during onboarding
func SaveProfileHandler(c echo.Context) error {
	var req reqdto.UserProfileRequest
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

	base, err := services.SaveOnboardingProfile(userID, req)
	if err != nil {
		// service already provided base with message
		return c.JSON(http.StatusInternalServerError, base)
	}
	return c.JSON(http.StatusOK, base)
}

func GetUserProfileHandler(c echo.Context) error {
	uid := c.Get("user_id")

	var userID uint64
	switch v := uid.(type) {
	case float64:
		userID = uint64(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnauthorized,
				resdto.BaseResponse{IsSuccess: false, Message: "invalid user id"})
		}
		userID = id
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	default:
		return c.JSON(http.StatusUnauthorized,
			resdto.BaseResponse{IsSuccess: false, Message: "invalid user id"})
	}

	resp, err := services.GetOnboardingProfile(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(http.StatusOK, resp)
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
