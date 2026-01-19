package middlewares

import (
	"net/http"
	"skinSync/dto/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// AdminAuthMiddleware validates JWT token for admin/clinic users
func AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "authorization header missing",
			})
		}

		// Check if header starts with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "invalid authorization header format",
			})
		}

		// Extract token from "Bearer <token>"
		tokenString := authHeader[7:]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "invalid token claims",
			})
		}

		// Check if token has admin_id and role (admin JWT format)
		adminID, hasAdminID := claims["admin_id"]
		role, hasRole := claims["role"]

		if !hasAdminID || !hasRole {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "invalid admin token",
			})
		}

		// Set admin context
		c.Set("admin_id", adminID)
		c.Set("admin_role", role)
		c.Set("email", claims["email"])

		return next(c)
	}
}
