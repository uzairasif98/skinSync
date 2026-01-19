package middlewares

import (
	"net/http"
	"skinSync/dto/response"

	"github.com/labstack/echo/v4"
)

// AdminOnly middleware ensures only admin role can access
func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, ok := c.Get("admin_role").(string)
		if !ok {
			return c.JSON(http.StatusForbidden, response.BaseResponse{
				IsSuccess: false,
				Message:   "role not found in token",
			})
		}

		if role != "admin" {
			return c.JSON(http.StatusForbidden, response.BaseResponse{
				IsSuccess: false,
				Message:   "admin access required",
			})
		}

		return next(c)
	}
}

// ClinicOnly middleware allows both clinic and admin roles
func ClinicOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, ok := c.Get("admin_role").(string)
		if !ok {
			return c.JSON(http.StatusForbidden, response.BaseResponse{
				IsSuccess: false,
				Message:   "role not found in token",
			})
		}

		if role != "clinic" && role != "admin" {
			return c.JSON(http.StatusForbidden, response.BaseResponse{
				IsSuccess: false,
				Message:   "clinic or admin access required",
			})
		}

		return next(c)
	}
}
