package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"skinSync/models"
)

var DB *gorm.DB

// loadEnv loads the appropriate env file based on ACTIVE_PROFILE.
func loadEnv() error {
	if os.Getenv("ACTIVE_PROFILE") == "PRODUCTION" {
		if err := godotenv.Load("prod.env"); err != nil {
			return fmt.Errorf("error loading prod.env: %w", err)
		}
	} else {
		if err := godotenv.Load("dev.env"); err != nil {
			return fmt.Errorf("error loading dev.env: %w", err)
		}
	}
	return nil
}

// OpenDB returns a new *gorm.DB connection. Caller should close it when done.
func OpenDB() (*gorm.DB, error) {
	if err := loadEnv(); err != nil {
		return nil, err
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm DB: %w", err)
	}
	return gormDB, nil
}

// ConnectDB opens a connection, assigns it to package-level DB, and runs AutoMigrate.
// It logs fatal on unrecoverable errors.
func ConnectDB() {
	var err error

	DB, err = OpenDB()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Run migrations for all models
	if err := DB.AutoMigrate(
		&models.User{},
		&models.AuthProvider{},
		&models.AuthToken{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},
		// Admin/Clinic users
		&models.AdminUser{},
		&models.AdminPermission{},
		// onboarding question/option/answer tables (single-step)
		&models.SkinConditionQuestion{},
		&models.SkinConditionQuestionOption{},
		&models.SkinConditionQuestionAnswer{},
		&models.UserProfile{},
		// treatment tables
		&models.Treatment{},
		&models.Area{},
		&models.SideArea{},
	); err != nil {
		// attempt to close DB on migration error
		if cerr := CloseDB(); cerr != nil {
			log.Printf("also failed to close DB after migration error: %v", cerr)
		}
		log.Fatalf("Failed to migrate database: %v", err)
	}
	SeedOnboardingData()
	SeedRBACData()
	log.Println("Database connection established")
}

// CloseDB closes the underlying sql.DB connection managed by GORM.
func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve sql.DB from gorm DB: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close sql.DB: %w", err)
	}
	DB = nil
	return nil
}

// SeedOnboardingData creates default master data for onboarding (idempotent)
func SeedOnboardingData() {
	db := DB
	if db == nil {
		log.Println("DB not initialized, skipping seeding")
		return
	}

	// Seed only the single onboarding questions and their options (idempotent)
	questions := []struct {
		Text    string
		Options []string
	}{
		{"How often do you experience breakouts?", []string{"Never", "Sometimes", "Often", "Always"}},
		{"How sensitive is your skin to new products?", []string{"Not sensitive", "Slightly", "Moderately", "Very"}},
		{"How would you describe your skin's oiliness?", []string{"Very oily", "Oily", "Normal", "Dry", "Very dry"}},
		{"How often do you use sunscreen?", []string{"Never", "Sometimes", "Always"}},
		{"Do you notice visible pores?", []string{"No", "Yes - small", "Yes - large"}},
		{"How would you rate hyperpigmentation / dark spots?", []string{"None", "Mild", "Moderate", "Severe"}},
		{"How often do you exfoliate?", []string{"Never", "Monthly", "Weekly", "2-3 times/week"}},
		{"How often do you experience redness?", []string{"Never", "Sometimes", "Often", "Always"}},
	}
	for _, q := range questions {
		var qq models.SkinConditionQuestion
		db.Where("question_text = ?", q.Text).FirstOrCreate(&qq, models.SkinConditionQuestion{QuestionText: q.Text})
		// ensure options exist
		for _, opt := range q.Options {
			var o models.SkinConditionQuestionOption
			db.Where("question_id = ? AND option_text = ?", qq.ID, opt).FirstOrCreate(&o, models.SkinConditionQuestionOption{QuestionID: qq.ID, OptionText: opt})
		}
	}
}

// SeedRBACData creates default roles and permissions (idempotent)
func SeedRBACData() {
	db := DB
	if db == nil {
		log.Println("DB not initialized, skipping RBAC seeding")
		return
	}

	// Define permissions
	permissions := []struct {
		Name        string
		Description string
	}{
		// User Management
		{"users.view", "View customer users"},
		{"users.edit", "Edit customer users"},
		{"users.delete", "Delete customer users"},

		// Clinic Management
		{"clinics.view", "View clinics"},
		{"clinics.edit", "Edit clinics"},
		{"clinics.delete", "Delete clinics"},

		// Treatment Management
		{"treatments.view", "View treatments"},
		{"treatments.edit", "Edit treatments"},
		{"treatments.delete", "Delete treatments"},

		// Onboarding Management
		{"onboarding.view", "View onboarding questions"},
		{"onboarding.edit", "Edit onboarding questions"},
		{"onboarding.delete", "Delete onboarding questions"},

		// Analytics
		{"analytics.view", "View analytics and reports"},
		{"analytics.export", "Export reports"},

		// Admin Management
		{"admins.view", "View admin users"},
		{"admins.create", "Create admin users"},
		{"admins.edit", "Edit admin users"},
		{"admins.delete", "Delete admin users"},

		// Appointment Management
		{"appointments.view", "View appointments"},
		{"appointments.edit", "Edit appointments"},
		{"appointments.delete", "Delete appointments"},

		// Profile
		{"profile.view", "View own profile"},
		{"profile.edit", "Edit own profile"},
	}

	// Create permissions (idempotent)
	permMap := make(map[string]uint64)
	for _, p := range permissions {
		var perm models.Permission
		desc := p.Description
		db.Where("name = ?", p.Name).FirstOrCreate(&perm, models.Permission{
			Name:        p.Name,
			Description: &desc,
		})
		permMap[p.Name] = perm.ID
	}

	// Define roles with their permissions
	roles := []struct {
		Name        string
		Description string
		Permissions []string
	}{
		{
			Name:        "super_admin",
			Description: "Super administrator with all permissions",
			Permissions: []string{
				"users.view", "users.edit", "users.delete",
				"clinics.view", "clinics.edit", "clinics.delete",
				"treatments.view", "treatments.edit", "treatments.delete",
				"onboarding.view", "onboarding.edit", "onboarding.delete",
				"analytics.view", "analytics.export",
				"admins.view", "admins.create", "admins.edit", "admins.delete",
				"appointments.view", "appointments.edit", "appointments.delete",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        "admin",
			Description: "Administrator with most permissions",
			Permissions: []string{
				"users.view", "users.edit",
				"clinics.view", "clinics.edit",
				"treatments.view", "treatments.edit",
				"onboarding.view", // view only, no edit/delete
				"analytics.view",
				"admins.view",
				"appointments.view", "appointments.edit",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        "clinic_manager",
			Description: "Clinic manager with clinic-related permissions",
			Permissions: []string{
				"appointments.view", "appointments.edit", "appointments.delete",
				"treatments.view",
				"users.view",
				"analytics.view",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        "clinic_staff",
			Description: "Clinic staff with limited permissions",
			Permissions: []string{
				"appointments.view", "appointments.edit",
				"treatments.view",
				"profile.view", "profile.edit",
			},
		},
	}

	// Create roles and assign permissions (idempotent)
	for _, r := range roles {
		var role models.Role
		desc := r.Description
		db.Where("name = ?", r.Name).FirstOrCreate(&role, models.Role{
			Name:        r.Name,
			Description: &desc,
		})

		// Remove existing role permissions
		db.Where("role_id = ?", role.ID).Delete(&models.RolePermission{})

		// Add permissions to role
		for _, permName := range r.Permissions {
			if permID, ok := permMap[permName]; ok {
				db.Create(&models.RolePermission{
					RoleID:       role.ID,
					PermissionID: permID,
				})
			}
		}
	}

	log.Println("RBAC data seeded successfully")
}
