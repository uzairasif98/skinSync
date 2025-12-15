package middlewares

import (
	"net/http"
	"skinSync/dto/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("skinSync")

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{Message: "Authorization header missing"})
		}

		// Extract token from "Bearer <token>"
		tokenString := authHeader[len("Bearer "):]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{Message: "invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, response.BaseResponse{Message: "invalid token claims"})
		}

		c.Set("email", claims["email"])
		c.Set("user_id", claims["user_id"])

		return next(c)
	}
}
