package router

import (
	"os"
	"skinSync/controllers"
	"skinSync/middlewares"

	"github.com/labstack/echo/v4"
)

func RequestHandler(e *echo.Echo) {
	server := os.Getenv("SERVER_ADDRESS")
	router := e.Group("/api")
	SetupRoutes(router)
	e.Logger.Fatal(e.Start(server))
}

func SetupRoutes(e *echo.Group) {
	e.Static("/", os.Getenv("STATIC_DIR"))
	public := e.Group("")
	{
		public.POST("/login", controllers.Login)
		// public masters for frontend to render step UI
		public.GET("/onboarding/masters", controllers.GetOnboardingMastersHandler)

	}

	// Protected routes with authentication and logging

	// current user's saved onboarding selections (requires token)

	api := e.Group("/v1", middlewares.AuthMiddleware)
	{
		api.POST("/auth/refresh", controllers.RefreshTokenHandler)
		// onboarding protected endpoints
		api.POST("/onboarding/answer", controllers.SaveOnboardingHandler)
		api.GET("/onboarding/user", controllers.GetUserOnboardingHandler)

	}

}
