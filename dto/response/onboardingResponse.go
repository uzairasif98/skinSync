package response

// DTOs for user's onboarding answers
type SkinTypeAnswer struct {
	QuestionID uint64 `json:"question_id"`
	TypeID     uint64 `json:"type_id"`
}

type ConcernAnswer struct {
	QuestionID uint64 `json:"question_id"`
	ConcernID  uint64 `json:"concern_id"`
}

type LifeStyleAnswer struct {
	QuestionID    uint64 `json:"question_id"`
	DescriptionID uint64 `json:"description_id"`
}

type QuestionAnswer struct {
	QuestionID uint64 `json:"question_id"`
	OptionID   uint64 `json:"option_id"`
}

type GoalAnswer struct {
	QuestionID uint64 `json:"question_id"`
	GoalID     uint64 `json:"goal_id"`
}

type UserOnboardingResponse struct {
	SkinTypes      []SkinTypeAnswer  `json:"skin_types"`
	Concerns       []ConcernAnswer   `json:"concerns"`
	Lifestyles     []LifeStyleAnswer `json:"lifestyles"`
	Questions      []QuestionAnswer  `json:"questions"`
	Goals          []GoalAnswer      `json:"goals"`
	CompletedSteps []string          `json:"completed_steps"`
}
