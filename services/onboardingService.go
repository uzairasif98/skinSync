package services

import (
	"errors"
	"strings"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
)

func GetOnboardingMasters() (*resdto.OnboardingMastersResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("db not initialized")
	}

	var resp resdto.OnboardingMastersResponse
	resp.IsSuccess = true
	resp.Message = "Onboarding masters retrieved successfully"

	if err := db.Preload("Options").Find(&resp.Data.SkinTypes).Error; err != nil {
		return nil, err
	}
	if err := db.Preload("Options").Find(&resp.Data.Concerns).Error; err != nil {
		return nil, err
	}
	if err := db.Preload("Options").Find(&resp.Data.Lifestyles).Error; err != nil {
		return nil, err
	}
	// preload options for questions
	if err := db.Preload("Options").Find(&resp.Data.Questions).Error; err != nil {
		return nil, err
	}
	if err := db.Preload("Options").Find(&resp.Data.Goals).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

// SaveOnboardingAnswer saves user's answers for a specific onboarding step.
func SaveOnboardingAnswer(userID uint64, req reqdto.OnboardingAnswerRequest) error {
	db := config.DB
	if db == nil {
		return errors.New("db not initialized")
	}
	step := strings.ToLower(req.Step)
	switch step {
	case "skin_type", "skintype", "skin_types":
		for _, a := range req.Answers {
			// remove existing answer for this user+question
			if err := db.Where("user_id = ? AND question_id = ?", userID, a.QuestionID).Delete(&models.SkinType{}).Error; err != nil {
				return err
			}
			st := models.SkinType{UserID: userID, QuestionID: a.QuestionID, TypeID: a.OptionID}
			if err := db.Create(&st).Error; err != nil {
				return err
			}
		}
		return nil
	case "concern", "concerns":
		for _, a := range req.Answers {
			if err := db.Where("user_id = ? AND question_id = ?", userID, a.QuestionID).Delete(&models.SkinConcern{}).Error; err != nil {
				return err
			}
			sc := models.SkinConcern{UserID: userID, QuestionID: a.QuestionID, ConcernID: a.OptionID}
			if err := db.Create(&sc).Error; err != nil {
				return err
			}
		}
		return nil
	case "lifestyle", "lifestyles", "life_style", "lifestyledescription":
		for _, a := range req.Answers {
			if err := db.Where("user_id = ? AND question_id = ?", userID, a.QuestionID).Delete(&models.LifeStyleDescription{}).Error; err != nil {
				return err
			}
			ls := models.LifeStyleDescription{UserID: userID, QuestionID: a.QuestionID, DescriptionID: a.OptionID}
			if err := db.Create(&ls).Error; err != nil {
				return err
			}
		}
		return nil
	case "question", "questions", "condition", "conditions":
		for _, a := range req.Answers {
			if err := db.Where("user_id = ? AND question_id = ?", userID, a.QuestionID).Delete(&models.SkinConditionQuestionAnswer{}).Error; err != nil {
				return err
			}
			qa := models.SkinConditionQuestionAnswer{UserID: userID, QuestionID: a.QuestionID, OptionID: a.OptionID}
			if err := db.Create(&qa).Error; err != nil {
				return err
			}
		}
		return nil
	case "goal", "goals":
		for _, a := range req.Answers {
			if err := db.Where("user_id = ? AND question_id = ?", userID, a.QuestionID).Delete(&models.UserSkinGoal{}).Error; err != nil {
				return err
			}
			ug := models.UserSkinGoal{UserID: userID, QuestionID: a.QuestionID, GoalID: a.OptionID}
			if err := db.Create(&ug).Error; err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("unknown onboarding step")
	}
}

// GetUserOnboarding returns all saved answers for a user.
func GetUserOnboarding(userID uint64) (*resdto.UserOnboardingResponse, error) {
	db := config.DB
	if db == nil {
		return nil, errors.New("db not initialized")
	}

	var resp resdto.UserOnboardingResponse
	resp.IsSuccess = true
	resp.Message = "User onboarding data retrieved successfully"

	var sTypes []models.SkinType
	if err := db.Where("user_id = ?", userID).Find(&sTypes).Error; err != nil {
		return nil, err
	}
	for _, s := range sTypes {
		resp.Data.SkinTypes = append(resp.Data.SkinTypes, resdto.SkinTypeAnswer{QuestionID: s.QuestionID, TypeID: s.TypeID})
	}

	var concerns []models.SkinConcern
	if err := db.Where("user_id = ?", userID).Find(&concerns).Error; err != nil {
		return nil, err
	}
	for _, c := range concerns {
		resp.Data.Concerns = append(resp.Data.Concerns, resdto.ConcernAnswer{QuestionID: c.QuestionID, ConcernID: c.ConcernID})
	}

	var lifestyles []models.LifeStyleDescription
	if err := db.Where("user_id = ?", userID).Find(&lifestyles).Error; err != nil {
		return nil, err
	}
	for _, l := range lifestyles {
		resp.Data.Lifestyles = append(resp.Data.Lifestyles, resdto.LifeStyleAnswer{QuestionID: l.QuestionID, DescriptionID: l.DescriptionID})
	}

	var qas []models.SkinConditionQuestionAnswer
	if err := db.Where("user_id = ?", userID).Find(&qas).Error; err != nil {
		return nil, err
	}
	for _, q := range qas {
		resp.Data.Questions = append(resp.Data.Questions, resdto.QuestionAnswer{QuestionID: q.QuestionID, OptionID: q.OptionID})
	}

	var goals []models.UserSkinGoal
	if err := db.Where("user_id = ?", userID).Find(&goals).Error; err != nil {
		return nil, err
	}
	for _, g := range goals {
		resp.Data.Goals = append(resp.Data.Goals, resdto.GoalAnswer{QuestionID: g.QuestionID, GoalID: g.GoalID})
	}

	// mark completed steps (simple heuristic)
	if len(resp.Data.SkinTypes) > 0 {
		resp.Data.CompletedSteps = append(resp.Data.CompletedSteps, "skin_type")
	}
	if len(resp.Data.Concerns) > 0 {
		resp.Data.CompletedSteps = append(resp.Data.CompletedSteps, "concerns")
	}
	if len(resp.Data.Lifestyles) > 0 {
		resp.Data.CompletedSteps = append(resp.Data.CompletedSteps, "lifestyles")
	}
	if len(resp.Data.Questions) > 0 {
		resp.Data.CompletedSteps = append(resp.Data.CompletedSteps, "questions")
	}
	if len(resp.Data.Goals) > 0 {
		resp.Data.CompletedSteps = append(resp.Data.CompletedSteps, "goals")
	}

	return &resp, nil
}
