package services

import (
	"errors"
	"strings"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"
)

type OnboardingMasters struct {
	Questions []models.SkinConditionQuestion `json:"questions"` // preloaded with Options
}

func GetOnboardingMasters() (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}
	var masters OnboardingMasters
	// preload options for questions
	if err := db.Preload("Options").Find(&masters.Questions).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}
	base = resdto.BaseResponse{IsSuccess: true, Message: "", Data: masters}
	return base, nil
}

// CreateOnboardingQuestion creates a question and its options.
func CreateOnboardingQuestion(req reqdto.CreateQuestionRequest) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	q := models.SkinConditionQuestion{QuestionText: req.QuestionText}
	if err := db.Create(&q).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	// insert options
	for _, optText := range req.Options {
		o := models.SkinConditionQuestionOption{QuestionID: q.ID, OptionText: optText}
		if err := db.Create(&o).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
	}

	// reload question with options
	if err := db.Preload("Options").First(&q, q.ID).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "created", Data: q}
	return base, nil
}

// AddOptionsToQuestion adds options to an existing question by id (string id accepted)
func AddOptionsToQuestion(qidStr string, req reqdto.AddOptionsRequest) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	// parse qid
	var q models.SkinConditionQuestion
	if err := db.First(&q, qidStr).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "question not found"}
		return base, err
	}

	for _, optText := range req.Options {
		o := models.SkinConditionQuestionOption{QuestionID: q.ID, OptionText: optText}
		if err := db.Create(&o).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
	}

	if err := db.Preload("Options").First(&q, q.ID).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "options added", Data: q}
	return base, nil
}

// UpdateOnboardingQuestion updates question text and optionally replaces or adds options.
func UpdateOnboardingQuestion(qidStr string, req reqdto.UpdateQuestionRequest) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	var q models.SkinConditionQuestion
	if err := db.Preload("Options").First(&q, qidStr).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "question not found"}
		return base, err
	}

	// update question text if provided
	if req.QuestionText != nil {
		q.QuestionText = *req.QuestionText
		if err := db.Save(&q).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
	}

	// handle options
	if req.ReplaceOptions {
		// delete existing options
		if err := db.Where("question_id = ?", q.ID).Delete(&models.SkinConditionQuestionOption{}).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
		// insert unique options
		seen := make(map[string]bool)
		for _, optText := range req.Options {
			if strings.TrimSpace(optText) == "" {
				continue
			}
			if seen[optText] {
				continue
			}
			seen[optText] = true
			o := models.SkinConditionQuestionOption{QuestionID: q.ID, OptionText: optText}
			if err := db.Create(&o).Error; err != nil {
				base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
				return base, err
			}
		}
	} else if len(req.Options) > 0 {
		// add options that don't already exist
		for _, optText := range req.Options {
			if strings.TrimSpace(optText) == "" {
				continue
			}
			var existing models.SkinConditionQuestionOption
			if err := db.Where("question_id = ? AND option_text = ?", q.ID, optText).First(&existing).Error; err == nil {
				// exists: skip (idempotent)
				continue
			}
			o := models.SkinConditionQuestionOption{QuestionID: q.ID, OptionText: optText}
			if err := db.Create(&o).Error; err != nil {
				base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
				return base, err
			}
		}
	}

	// reload
	if err := db.Preload("Options").First(&q, q.ID).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "updated", Data: q}
	return base, nil
}

// DeleteOnboardingQuestion deletes a question and its options.
func DeleteOnboardingQuestion(qidStr string) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	var q models.SkinConditionQuestion
	if err := db.First(&q, qidStr).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "question not found"}
		return base, err
	}

	if err := db.Delete(&q).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "deleted"}
	return base, nil
}

// DeleteOnboardingOption deletes a single option under a question.
func DeleteOnboardingOption(qidStr string, oidStr string) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	var o models.SkinConditionQuestionOption
	if err := db.Where("question_id = ? AND id = ?", qidStr, oidStr).First(&o).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "option not found"}
		return base, err
	}

	if err := db.Delete(&o).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "option deleted"}
	return base, nil
}

// SaveOnboardingAnswer saves user's answers for a specific onboarding step.
func SaveOnboardingAnswer(userID uint64, req reqdto.OnboardingAnswerRequest) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}

	// Accept a few possible step names; if provided and unknown, return error.
	step := strings.ToLower(strings.TrimSpace(req.Step))
	if step != "" && step != "onboarding" && step != "questions" && step != "answers" && step != "onboard" {
		base = resdto.BaseResponse{IsSuccess: false, Message: "unknown onboarding step"}
		return base, errors.New("unknown onboarding step")
	}

	// Group options by question_id so we can replace per-question answers
	byQ := make(map[uint64][]uint64)
	for _, a := range req.Answers {
		byQ[a.QuestionID] = append(byQ[a.QuestionID], a.OptionID)
	}

	for qID, opts := range byQ {
		// delete existing answers for this user + question
		if err := db.Where("user_id = ? AND question_id = ?", userID, qID).Delete(&models.SkinConditionQuestionAnswer{}).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
		// insert new answers (one record per selected option)
		for _, opt := range opts {
			qa := models.SkinConditionQuestionAnswer{UserID: userID, QuestionID: qID, OptionID: opt}
			if err := db.Create(&qa).Error; err != nil {
				base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
				return base, err
			}
		}
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "saved"}
	return base, nil
}

// GetUserOnboarding returns all saved answers for a user.
func GetUserOnboarding(userID uint64) (resdto.BaseResponse, error) {
	var base resdto.BaseResponse
	db := config.DB
	if db == nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: "db not initialized"}
		return base, errors.New("db not initialized")
	}
	var resp resdto.UserOnboardingResponse
	var qas []models.SkinConditionQuestionAnswer
	if err := db.Where("user_id = ?", userID).Find(&qas).Error; err != nil {
		base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
		return base, err
	}

	// collect unique question IDs and option IDs preserving occurrence order
	questionOrder := make([]uint64, 0)
	seenQuestions := make(map[uint64]bool)
	optionIDs := make([]uint64, 0, len(qas))
	for _, q := range qas {
		if !seenQuestions[q.QuestionID] {
			seenQuestions[q.QuestionID] = true
			questionOrder = append(questionOrder, q.QuestionID)
		}
		optionIDs = append(optionIDs, q.OptionID)
	}

	// fetch option texts
	optionTextMap := make(map[uint64]string)
	if len(optionIDs) > 0 {
		var opts []models.SkinConditionQuestionOption
		if err := db.Where("id IN ?", optionIDs).Find(&opts).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
		for _, o := range opts {
			optionTextMap[o.ID] = o.OptionText
		}
	}

	// fetch question texts
	questionTextMap := make(map[uint64]string)
	if len(questionOrder) > 0 {
		var qMasters []models.SkinConditionQuestion
		if err := db.Where("id IN ?", questionOrder).Find(&qMasters).Error; err != nil {
			base = resdto.BaseResponse{IsSuccess: false, Message: err.Error()}
			return base, err
		}
		for _, qm := range qMasters {
			questionTextMap[qm.ID] = qm.QuestionText
		}
	}

	// group options per question
	optionsByQuestion := make(map[uint64][]resdto.OptionAnswer)
	for _, q := range qas {
		oa := resdto.OptionAnswer{OptionID: q.OptionID}
		if txt, ok := optionTextMap[q.OptionID]; ok {
			oa.OptionText = txt
		}
		// avoid duplicates
		exists := false
		for _, ex := range optionsByQuestion[q.QuestionID] {
			if ex.OptionID == oa.OptionID {
				exists = true
				break
			}
		}
		if !exists {
			optionsByQuestion[q.QuestionID] = append(optionsByQuestion[q.QuestionID], oa)
		}
	}

	// build grouped response preserving question order
	for _, qid := range questionOrder {
		qg := resdto.QuestionGroup{QuestionID: qid, QuestionText: questionTextMap[qid], Options: optionsByQuestion[qid]}
		resp.Questions = append(resp.Questions, qg)
	}

	resp.CompletedSteps = []string{}
	if len(resp.Questions) > 0 {
		resp.CompletedSteps = append(resp.CompletedSteps, "onboarding")
	}

	base = resdto.BaseResponse{IsSuccess: true, Message: "", Data: resp}
	return base, nil
}
