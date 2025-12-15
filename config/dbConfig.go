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
		// onboarding question/option tables
		&models.SkinTypeQuestion{},
		&models.SkinTypeOption{},
		&models.SkinType{},
		&models.ConcernQuestion{},
		&models.ConcernOption{},
		&models.SkinConcern{},
		&models.LifeStyleQuestion{},
		&models.LifeStyleDescriptionOption{},
		&models.LifeStyleDescription{},
		&models.SkinConditionQuestion{},
		&models.SkinConditionQuestionOption{},
		&models.SkinConditionQuestionAnswer{},
		&models.SkinGoalQuestion{},
		&models.SkinGoal{},
		&models.UserSkinGoal{},
	); err != nil {
		// attempt to close DB on migration error
		if cerr := CloseDB(); cerr != nil {
			log.Printf("also failed to close DB after migration error: %v", cerr)
		}
		log.Fatalf("Failed to migrate database: %v", err)
	}
	SeedOnboardingData()
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

	// Step 1 skin types (as a question with options)
	skinTypes := []string{"Normal", "Oily", "Dry", "Combination", "Sensitive"}
	var stq models.SkinTypeQuestion
	qText := "Which of these best describes your skin type?"
	db.Where("question_text = ?", qText).FirstOrCreate(&stq, models.SkinTypeQuestion{QuestionText: qText})
	for _, name := range skinTypes {
		var o models.SkinTypeOption
		db.Where("question_id = ? AND name = ?", stq.ID, name).FirstOrCreate(&o, models.SkinTypeOption{QuestionID: stq.ID, Name: name})
	}

	// Step 2 concerns (question + options)
	concerns := []string{"Acne", "Aging", "Hyperpigmentation", "Redness", "Sensitivity"}
	var cq models.ConcernQuestion
	cqText := "Which skin concerns do you have?"
	db.Where("question_text = ?", cqText).FirstOrCreate(&cq, models.ConcernQuestion{QuestionText: cqText})
	for _, name := range concerns {
		var o models.ConcernOption
		db.Where("question_id = ? AND name = ?", cq.ID, name).FirstOrCreate(&o, models.ConcernOption{QuestionID: cq.ID, Name: name})
	}

	// Step 3 lifestyles (question + options)
	lifestyles := []string{"Smoker", "High Stress", "Poor Sleep", "Active", "Vegetarian"}
	var lq models.LifeStyleQuestion
	lqText := "Which lifestyle factors apply to you?"
	db.Where("question_text = ?", lqText).FirstOrCreate(&lq, models.LifeStyleQuestion{QuestionText: lqText})
	for _, name := range lifestyles {
		var o models.LifeStyleDescriptionOption
		db.Where("question_id = ? AND name = ?", lq.ID, name).FirstOrCreate(&o, models.LifeStyleDescriptionOption{QuestionID: lq.ID, Name: name})
	}

	// Step 4 questions + options (already in Question/Option form)
	questions := []struct {
		Text    string
		Options []string
	}{
		{"How often do you experience breakouts?", []string{"Never", "Sometimes", "Often", "Always"}},
		{"How sensitive is your skin to new products?", []string{"Not sensitive", "Slightly", "Moderately", "Very"}},
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

	// Step 5 goals (question + options)
	goals := []string{"Hydration", "Brightening", "Reduce Breakouts", "Anti-Age"}
	var gq models.SkinGoalQuestion
	gqText := "What are your skin goals?"
	db.Where("question_text = ?", gqText).FirstOrCreate(&gq, models.SkinGoalQuestion{QuestionText: gqText})
	for _, name := range goals {
		var o models.SkinGoal
		db.Where("question_id = ? AND name = ?", gq.ID, name).FirstOrCreate(&o, models.SkinGoal{QuestionID: gq.ID, Name: name})
	}
}
