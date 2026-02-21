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

		// Admin login (Email + Password) - Login is public, Register requires super_admin
		public.POST("/admin/login", controllers.AdminLoginHandler)

		// Clinic login (Email + Password) - Login is public, Register by super_admin
		public.POST("/clinic/login", controllers.ClinicLoginHandler)

		// Clinic password reset (public - no auth needed)
		public.POST("/clinic/forgot-password", controllers.ClinicForgotPasswordHandler)
		public.POST("/clinic/reset-password", controllers.ClinicResetPasswordHandler)

		// Public masters (onboarding only - no auth needed for initial app load)
		public.GET("/onboarding/masters", controllers.GetOnboardingMastersHandler)

		// Treatment tree APIs (public - no auth required)
		public.GET("/treatments/masters", controllers.GetTreatmentMastersHandler)
		public.GET("/treatments/:id/areas", controllers.GetAreasHandler)
		public.GET("/treatments/:treatmentId/areas/:areaId/sideareas", controllers.GetSideAreasHandler)
	}

	// ========== UNIFIED AUTH ROUTES (Any valid token: customer/admin/clinic) ==========
	unified := e.Group("", middlewares.UnifiedAuthMiddleware)
	{
		// Discovery APIs (Clinic ↔ Treatment ↔ Doctor)
		unified.GET("/clinics", controllers.GetAllClinicsHandler)
		unified.GET("/doctors", controllers.GetAllDoctorsHandler)
		unified.GET("/treatments/:treatmentId/clinics", controllers.GetClinicsByTreatmentHandler)
		unified.GET("/doctors/:doctorId/treatments", controllers.GetTreatmentsByDoctorHandler)
		unified.GET("/clinics/:clinicId/treatments/:treatmentId/doctors", controllers.GetDoctorsByClinicAndTreatmentHandler)
		unified.GET("/doctors/:doctorId/clinics/:clinicId/treatments", controllers.GetTreatmentsByDoctorAndClinicHandler)
		unified.GET("/doctors/:doctorId/treatments/:treatmentId/clinics", controllers.GetClinicsByDoctorAndTreatmentHandler)
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

		// Treatment CRUD (super_admin only)
		// TODO: Implement treatment CRUD controllers
		// admin.POST("/treatments", controllers.CreateTreatmentHandler, middlewares.RequirePermission("treatments.edit"))
		// admin.PUT("/treatments/:id", controllers.UpdateTreatmentHandler, middlewares.RequirePermission("treatments.edit"))
		// admin.DELETE("/treatments/:id", controllers.DeleteTreatmentHandler, middlewares.RequirePermission("treatments.delete"))
	}

	// ========== CLINIC ROUTES (Clinic Auth Required) ==========
	clinic := e.Group("/clinic", middlewares.ClinicAuthMiddleware)
	{
		// Staff management (owner only)
		clinic.POST("/users/register", controllers.RegisterClinicUserHandler, middlewares.RequireClinicPermission("staff.create"))

		// Doctor/Injector registration (owner only) - treatments optional
		clinic.POST("/doctors/register", controllers.RegisterDoctorHandler, middlewares.RequireClinicPermission("staff.create"))

		// Assign treatments to existing doctor/injector
		clinic.POST("/doctors/treatments", controllers.AssignDoctorTreatmentsHandler, middlewares.RequireClinicPermission("staff.edit"))

		// List all doctors/injectors for this clinic
		clinic.GET("/doctors", controllers.GetDoctorsHandler, middlewares.RequireClinicPermission("staff.view"))

		// Get full doctor detail by ID
		clinic.GET("/doctors/:id", controllers.GetDoctorDetailHandler, middlewares.RequireClinicPermission("staff.view"))

		// Clinic-managed areas and prices
		// Frontend posts side_area_id (we resolve area/treatment) and optional syringe_size for per-size prices
		clinic.POST("/side-areas", controllers.CreateClinicSideAreasFromSideAreaHandler, middlewares.RequireClinicPermission("areas.edit"))

		// Per-size prices (frontend can include syringe_size in payload)
		clinic.POST("/side-area-prices", controllers.CreateClinicSideAreasFromSideAreaHandler, middlewares.RequireClinicPermission("areas.edit"))

		// Bulk upsert: frontend sends treatment_id + area list
		clinic.POST("/side-areas/bulk", controllers.CreateClinicSideAreasFromAreaHandler, middlewares.RequireClinicPermission("areas.edit"))

		// Update side areas (bulk sync): full replace for a treatment - adds new, updates existing, removes missing
		clinic.PATCH("/side-areas/bulk", controllers.UpdateClinicSideAreasBulkHandler, middlewares.RequireClinicPermission("areas.edit"))

		// Get side areas by treatment ID
		clinic.GET("/side-areas/treatment/:treatmentId", controllers.GetSideAreasByTreatmentHandler)
		// Get all clinic roles
		clinic.GET("/roles", controllers.GetClinicRolesHandler)
		// Get treatments with side area prices for clinic
		clinic.GET("/treatments", controllers.GetTreatmentByClinicHandler)

		// Change password (requires auth)
		clinic.POST("/change-password", controllers.ClinicChangePasswordHandler)

		// Logout (requires auth)
		clinic.POST("/logout", controllers.ClinicLogoutHandler)

		// TODO: Add more clinic endpoints
		// clinic.GET("/profile/me", controllers.GetClinicUserProfileHandler, middlewares.RequireClinicPermission("profile.view"))
		// clinic.GET("/users", controllers.GetClinicUsersHandler, middlewares.RequireClinicPermission("staff.view"))
		// clinic.PUT("/users/:id", controllers.UpdateClinicUserHandler, middlewares.RequireClinicPermission("staff.edit"))
		// clinic.DELETE("/users/:id", controllers.DeleteClinicUserHandler, middlewares.RequireClinicPermission("staff.delete"))
	}

}
