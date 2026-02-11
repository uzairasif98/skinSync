package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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
		// clinic tables
		&models.Clinic{},
		&models.ClinicRole{},
		&models.ClinicPermission{},
		&models.ClinicRolePermission{},
		&models.ClinicUser{},
		&models.ClinicTreatment{},
		&models.ClinicUserTreatment{},
		&models.ClinicSideArea{},
		&models.ClinicUserSideArea{},
	); err != nil {
		// attempt to close DB on migration error
		if cerr := CloseDB(); cerr != nil {
			log.Printf("also failed to close DB after migration error: %v", cerr)
		}
		log.Fatalf("Failed to migrate database: %v", err)
	}
	SeedOnboardingData()
	SeedRBACData()
	SeedClinicRoles()
	SeedDiscoveryData()
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
		{"clinics.create", "Create/Register clinics"},
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
				"clinics.view", "clinics.create", "clinics.edit", "clinics.delete",
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

// SeedClinicRoles creates predefined clinic roles, permissions, and mappings (idempotent)
func SeedClinicRoles() {
	db := DB
	if db == nil {
		log.Println("DB not initialized, skipping clinic roles seeding")
		return
	}

	// Define clinic-specific permissions
	clinicPermissions := []struct {
		Name        string
		Description string
	}{
		// Staff Management
		{"staff.view", "View clinic staff"},
		{"staff.create", "Create/Register clinic staff"},
		{"staff.edit", "Edit clinic staff"},
		{"staff.delete", "Delete clinic staff"},

		// Appointment Management
		{"appointments.view", "View appointments"},
		{"appointments.create", "Create appointments"},
		{"appointments.edit", "Edit appointments"},
		{"appointments.delete", "Delete/Cancel appointments"},

		// Patient Management
		{"patients.view", "View patient records"},
		{"patients.create", "Create patient records"},
		{"patients.edit", "Edit patient records"},
		{"patients.delete", "Delete patient records"},

		// Treatment Records
		{"treatment_records.view", "View treatment records"},
		{"treatment_records.create", "Create treatment records"},
		{"treatment_records.edit", "Edit treatment records"},

		// Clinic Settings
		{"clinic.view", "View clinic settings"},
		{"clinic.edit", "Edit clinic settings"},

		// Area/Treatment Management
		{"areas.edit", "Edit treatment areas and pricing"},

		// Reports
		{"reports.view", "View reports and analytics"},
		{"reports.export", "Export reports"},

		// Profile
		{"profile.view", "View own profile"},
		{"profile.edit", "Edit own profile"},
	}

	// Create clinic permissions (idempotent)
	clinicPermMap := make(map[string]uint64)
	for _, p := range clinicPermissions {
		var perm models.ClinicPermission
		desc := p.Description
		db.Where("name = ?", p.Name).FirstOrCreate(&perm, models.ClinicPermission{
			Name:        p.Name,
			Description: &desc,
		})
		clinicPermMap[p.Name] = perm.ID
	}

	// Define clinic roles with their permissions
	clinicRoles := []struct {
		Name        string
		Description string
		Permissions []string
	}{
		{
			Name:        models.ClinicRoleOwner,
			Description: "Clinic owner with full access to clinic management",
			Permissions: []string{
				"staff.view", "staff.create", "staff.edit", "staff.delete",
				"appointments.view", "appointments.create", "appointments.edit", "appointments.delete",
				"patients.view", "patients.create", "patients.edit", "patients.delete",
				"treatment_records.view", "treatment_records.create", "treatment_records.edit",
				"clinic.view", "clinic.edit",
				"areas.edit",
				"reports.view", "reports.export",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        models.ClinicRoleManager,
			Description: "Clinic manager with administrative access",
			Permissions: []string{
				"staff.view", "staff.create", "staff.edit",
				"appointments.view", "appointments.create", "appointments.edit", "appointments.delete",
				"patients.view", "patients.create", "patients.edit",
				"treatment_records.view",
				"clinic.view",
				"areas.edit",
				"reports.view",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        models.ClinicRoleDoctor,
			Description: "Doctor/Physician at the clinic",
			Permissions: []string{
				"appointments.view", "appointments.edit",
				"patients.view", "patients.edit",
				"treatment_records.view", "treatment_records.create", "treatment_records.edit",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        models.ClinicRoleInjector,
			Description: "Injector/Aesthetician performing treatments",
			Permissions: []string{
				"appointments.view",
				"patients.view",
				"treatment_records.view", "treatment_records.create",
				"profile.view", "profile.edit",
			},
		},
		{
			Name:        models.ClinicRoleReceptionist,
			Description: "Front desk staff handling appointments",
			Permissions: []string{
				"appointments.view", "appointments.create", "appointments.edit",
				"patients.view", "patients.create",
				"profile.view", "profile.edit",
			},
		},
	}

	// Create clinic roles and assign permissions (idempotent)
	for _, r := range clinicRoles {
		var role models.ClinicRole
		db.Where("name = ?", r.Name).FirstOrCreate(&role, models.ClinicRole{
			Name:        r.Name,
			Description: r.Description,
		})

		// Remove existing role permissions
		db.Where("role_id = ?", role.ID).Delete(&models.ClinicRolePermission{})

		// Add permissions to role
		for _, permName := range r.Permissions {
			if permID, ok := clinicPermMap[permName]; ok {
				db.Create(&models.ClinicRolePermission{
					RoleID:       role.ID,
					PermissionID: permID,
				})
			}
		}
	}

	log.Println("Clinic roles and permissions seeded successfully")
}

// SeedDiscoveryData seeds sample clinics, doctors, and treatment mappings (idempotent)
func SeedDiscoveryData() {
	db := DB
	if db == nil {
		log.Println("DB not initialized, skipping discovery seeding")
		return
	}

	// Get role IDs
	var ownerRole, doctorRole models.ClinicRole
	db.Where("name = ?", models.ClinicRoleOwner).First(&ownerRole)
	db.Where("name = ?", models.ClinicRoleDoctor).First(&doctorRole)

	if ownerRole.ID == 0 || doctorRole.ID == 0 {
		log.Println("Clinic roles not found, skipping discovery seeding")
		return
	}

	// Get all treatments
	var treatments []models.Treatment
	db.Find(&treatments)
	if len(treatments) == 0 {
		log.Println("No treatments found, skipping discovery seeding")
		return
	}

	// Build treatment map by name
	treatmentMap := make(map[string]uint)
	for _, t := range treatments {
		treatmentMap[t.Name] = t.ID
	}

	// Define sample clinics
	sampleClinics := []struct {
		Name    string
		Email   string
		Phone   string
		Address string
	}{
		{"GlowUp Aesthetics", "info@glowup.com", "+1-555-0101", "456 Beauty Ave, Los Angeles, CA 90001"},
		{"SkinPerfect Clinic", "hello@skinperfect.com", "+1-555-0202", "789 Wellness Blvd, Miami, FL 33101"},
		{"Elite Derma Center", "contact@elitederma.com", "+1-555-0303", "321 Medical Plaza, New York, NY 10001"},
	}

	// Hash a default password for sample users
	defaultHash, err := hashPasswordForSeed("SeedPass@123")
	if err != nil {
		log.Printf("Failed to hash seed password: %v", err)
		return
	}

	for _, sc := range sampleClinics {
		// Create clinic (idempotent)
		var clinic models.Clinic
		db.Where("email = ?", sc.Email).FirstOrCreate(&clinic, models.Clinic{
			Name:    sc.Name,
			Email:   sc.Email,
			Phone:   sc.Phone,
			Address: sc.Address,
			Status:  "active",
		})

		// Create owner for clinic (idempotent)
		ownerEmail := "owner@" + sc.Email[strings.Index(sc.Email, "@")+1:]
		var owner models.ClinicUser
		db.Where("email = ? AND clinic_id = ?", ownerEmail, clinic.ID).FirstOrCreate(&owner, models.ClinicUser{
			ClinicID:     clinic.ID,
			Email:        ownerEmail,
			PasswordHash: defaultHash,
			Name:         "Owner of " + sc.Name,
			RoleID:       ownerRole.ID,
			Status:       "active",
		})

		// Create 2 doctors per clinic
		for i, docName := range []string{"Dr. Sarah Johnson", "Dr. Michael Chen"} {
			docEmail := fmt.Sprintf("doctor%d@%s", i+1, sc.Email[strings.Index(sc.Email, "@")+1:])
			var doc models.ClinicUser
			db.Where("email = ? AND clinic_id = ?", docEmail, clinic.ID).FirstOrCreate(&doc, models.ClinicUser{
				ClinicID:     clinic.ID,
				Email:        docEmail,
				PasswordHash: defaultHash,
				Name:         docName,
				RoleID:       doctorRole.ID,
				Status:       "active",
			})

			// Assign all treatments to doctors
			for _, t := range treatments {
				var ut models.ClinicUserTreatment
				db.Where("clinic_user_id = ? AND treatment_id = ?", doc.ID, t.ID).
					FirstOrCreate(&ut, models.ClinicUserTreatment{
						ClinicUserID: doc.ID,
						TreatmentID:  t.ID,
					})
			}
		}

		// Assign all treatments to clinic
		for _, t := range treatments {
			price := 250.0 + float64(clinic.ID*50) + float64(t.ID*25) // varied pricing
			var ct models.ClinicTreatment
			db.Where("clinic_id = ? AND treatment_id = ?", clinic.ID, t.ID).
				FirstOrCreate(&ct, models.ClinicTreatment{
					ClinicID:    clinic.ID,
					TreatmentID: t.ID,
					Price:       &price,
					Status:      "active",
				})
		}
	}

	log.Println("Discovery sample data seeded successfully")
}

// hashPasswordForSeed hashes password using bcrypt (for seeding only)
func hashPasswordForSeed(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
