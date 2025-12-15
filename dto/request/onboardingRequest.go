package request

type OnboardingAnswerItem struct {
	QuestionID uint64 `json:"question_id"`
	OptionID   uint64 `json:"option_id"`
}

// OnboardingAnswerRequest represents answers for a specific onboarding step
type OnboardingAnswerRequest struct {
	Step    string                 `json:"step"`
	Answers []OnboardingAnswerItem `json:"answers"`
}
