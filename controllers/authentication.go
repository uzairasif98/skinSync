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
