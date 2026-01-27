package middlewares

import (
	"net/http"
	"strings"

	resdto "skinSync/dto/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

var clinicJwtSecret = []byte("skinSync") // Same secret as admin JWT

// ClinicAuthMiddleware validates JWT token for clinic users
func ClinicAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return clinicJwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid or expired token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid token claims",
			})
		}

		// Extract clinic user info from claims
		clinicUserID, ok := claims["clinic_user_id"].(float64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid clinic_user_id in token",
			})
		}

		clinicID, ok := claims["clinic_id"].(float64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid clinic_id in token",
			})
		}

		email, ok := claims["email"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid email in token",
			})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, resdto.BaseResponse{
				IsSuccess: false,
				Message:   "invalid role in token",
			})
		}

		// Set context values for downstream handlers
		c.Set("clinic_user_id", clinicUserID)
		c.Set("clinic_id", clinicID)
		c.Set("email", email)
		c.Set("role", role)

		return next(c)
	}
}
