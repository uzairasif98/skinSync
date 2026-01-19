package middlewares

import (
	"net/http"
	"skinSync/dto/response"
	"skinSync/permissions"

	"github.com/labstack/echo/v4"
)

// RequirePermission middleware checks if admin has specific permission
func RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get admin_id from context (set by AdminAuthMiddleware)
			adminIDFloat, ok := c.Get("admin_id").(float64)
			if !ok {
				return c.JSON(http.StatusUnauthorized, response.BaseResponse{
					IsSuccess: false,
					Message:   "admin_id not found in context",
				})
			}

			adminID := uint64(adminIDFloat)

			// Check permission
			hasPermission, err := permissions.HasPermission(adminID, permission)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.BaseResponse{
					IsSuccess: false,
					Message:   "error checking permissions",
				})
			}

			if !hasPermission {
				return c.JSON(http.StatusForbidden, response.BaseResponse{
					IsSuccess: false,
					Message:   "insufficient permissions: " + permission + " required",
				})
			}

			return next(c)
		}
	}
}
