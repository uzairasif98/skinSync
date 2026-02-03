package middlewares

import (
	"net/http"
	"skinSync/dto/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// UnifiedAuthMiddleware accepts any valid JWT token (customer, admin, or clinic)
// and sets appropriate context values based on the token type
func UnifiedAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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

		// Determine token type and set context values accordingly
		if adminID, hasAdminID := claims["admin_id"]; hasAdminID {
			// Admin token
			c.Set("user_type", "admin")
			c.Set("admin_id", adminID)
			c.Set("admin_role", claims["role"])
			c.Set("email", claims["email"])
		} else if clinicUserID, hasClinicUserID := claims["clinic_user_id"]; hasClinicUserID {
			// Clinic token
			c.Set("user_type", "clinic")
			c.Set("clinic_user_id", clinicUserID)
			c.Set("clinic_id", claims["clinic_id"])
			c.Set("role", claims["role"])
			c.Set("email", claims["email"])
		} else if userID, hasUserID := claims["user_id"]; hasUserID {
			// Customer token
			c.Set("user_type", "customer")
			c.Set("user_id", userID)
			c.Set("email", claims["email"])
		} else {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{
				IsSuccess: false,
				Message:   "invalid token type",
			})
		}

		return next(c)
	}
}
