package response

// OptionAnswer represents a selected option for a question with display text
type OptionAnswer struct {
	OptionID   uint64 `json:"option_id"`
	OptionText string `json:"option_text,omitempty"`
}

// QuestionGroup groups selected options under a question and includes question text
type QuestionGroup struct {
	QuestionID   uint64         `json:"question_id"`
	QuestionText string         `json:"question_text,omitempty"`
	Options      []OptionAnswer `json:"options"`
}

// UserOnboardingResponse contains grouped questions (single onboarding step)
type UserOnboardingResponse struct {
	Questions      []QuestionGroup `json:"questions"`
	CompletedSteps []string        `json:"completed_steps"`
}
