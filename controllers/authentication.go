package controllers

import (
	"net/http"

	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/services"

	"github.com/labstack/echo/v4"
)

func Login(c echo.Context) error {
	var req reqdto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	// For email provider, call SendOTP and return BaseResponse
	if req.Provider == "email" {
		if req.Email == nil {
			return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: "email required"})
		}
		otpReq := reqdto.SendOTPRequest{
			Email:      *req.Email,
			DeviceInfo: req.DeviceInfo,
			IPAddress:  req.IPAddress,
		}
		resp, err := services.SendOTP(otpReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, resp)
		}
		return c.JSON(http.StatusOK, resp)
	}

	// For other providers (phone, google, apple), use existing login flow
	resp, err := services.Login(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
func RefreshTokenHandler(c echo.Context) error {
	var req reqdto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp, err := services.RefreshToken(req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

// SendOTPHandler handles OTP generation and sending
func SendOTPHandler(c echo.Context) error {
	var req reqdto.SendOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: "email is required"})
	}

	resp, err := services.SendOTP(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resp)
	}
	return c.JSON(http.StatusOK, resp)
}

// VerifyOTPHandler handles OTP verification and login
func VerifyOTPHandler(c echo.Context) error {
	var req reqdto.VerifyOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}

	if req.Email == "" || req.OTP == "" {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: "email and otp are required"})
	}

	resp, err := services.VerifyOTP(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, resdto.BaseResponse{IsSuccess: false, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
