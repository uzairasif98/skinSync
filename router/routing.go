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

	// ========== PUBLIC ROUTES (No Auth) ==========
	public := e.Group("")
	{
		// Customer login (OTP)
		public.POST("/login", controllers.Login)
		public.POST("/verify-otp", controllers.VerifyOTPHandler)

		// Admin/Clinic login (Email + Password) - Login is public, Register requires super_admin
		public.POST("/admin/login", controllers.AdminLoginHandler)

		// Public masters
		public.GET("/onboarding/masters", controllers.GetOnboardingMastersHandler)
		public.GET("/treatments/masters", controllers.GetTreatmentMastersHandler)
		public.GET("/treatments/:id/areas", controllers.GetAreasHandler)
		public.GET("/treatments/:treatmentId/areas/:areaId/sideareas", controllers.GetSideAreasHandler)
	}

	// ========== CUSTOMER ROUTES (Customer Auth Required) ==========
	customer := e.Group("/v1", middlewares.AuthMiddleware)
	{
		customer.POST("/auth/refresh", controllers.RefreshTokenHandler)

		// Onboarding endpoints
		customer.PATCH("/onboarding/answer", controllers.SaveOnboardingHandler)
		customer.POST("/onboarding/profile", controllers.SaveProfileHandler)
		customer.GET("/onboarding/fetchprofile", controllers.GetUserProfileHandler)
		customer.GET("/onboarding/user", controllers.GetUserOnboardingHandler)
	}

	// ========== ADMIN ROUTES (Permission-Based) ==========
	admin := e.Group("/admin", middlewares.AdminAuthMiddleware)
	{
		// Profile
		admin.GET("/me", controllers.GetAdminMeHandler, middlewares.RequirePermission("profile.view"))

		// Admin user management (super_admin only)
		admin.POST("/register", controllers.AdminRegisterHandler, middlewares.RequirePermission("admins.create"))
		admin.POST("/verify-password", controllers.VerifyPasswordHandler, middlewares.RequirePermission("admins.create"))

		// Onboarding question management
		admin.POST("/onboarding/question", controllers.AdminCreateQuestionHandler, middlewares.RequirePermission("onboarding.edit"))
		admin.POST("/onboarding/question/:id/options", controllers.AdminAddOptionsHandler, middlewares.RequirePermission("onboarding.edit"))
		admin.PUT("/onboarding/question/:id", controllers.AdminUpdateQuestionHandler, middlewares.RequirePermission("onboarding.edit"))
		admin.DELETE("/onboarding/question/:id", controllers.AdminDeleteQuestionHandler, middlewares.RequirePermission("onboarding.delete"))
		admin.DELETE("/onboarding/question/:qid/options/:optionId", controllers.AdminDeleteOptionHandler, middlewares.RequirePermission("onboarding.delete"))

		// User management (future)
		// admin.GET("/users", controllers.GetUsers, middlewares.RequirePermission("users.view"))
		// admin.PUT("/users/:id", controllers.UpdateUser, middlewares.RequirePermission("users.edit"))
		// admin.DELETE("/users/:id", controllers.DeleteUser, middlewares.RequirePermission("users.delete"))

		// Clinic management (super_admin only)
		admin.POST("/clinic/register", controllers.RegisterClinicHandler, middlewares.RequirePermission("clinics.create"))
	}

	// ========== CLINIC ROUTES (Permission-Based) ==========
	clinic := e.Group("/clinic", middlewares.AdminAuthMiddleware)
	{
		// Profile
		clinic.GET("/me", controllers.GetAdminMeHandler, middlewares.RequirePermission("profile.view"))

		// Appointments (future - clinic staff can view and edit)
		// clinic.GET("/appointments", controllers.GetAppointments, middlewares.RequirePermission("appointments.view"))
		// clinic.POST("/appointments", controllers.CreateAppointment, middlewares.RequirePermission("appointments.edit"))
		// clinic.PUT("/appointments/:id", controllers.UpdateAppointment, middlewares.RequirePermission("appointments.edit"))
	}

}
