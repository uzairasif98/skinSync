package request

// CreateQuestionRequest is the payload to create a question with options
type CreateQuestionRequest struct {
	QuestionText string   `json:"question_text"`
	Options      []string `json:"options"`
}

// AddOptionsRequest is the payload to add options to an existing question
type AddOptionsRequest struct {
	Options []string `json:"options"`
}

// UpdateQuestionRequest is the payload to update question text and optionally options
type UpdateQuestionRequest struct {
	QuestionText   *string  `json:"question_text,omitempty"`
	Options        []string `json:"options,omitempty"`
	ReplaceOptions bool     `json:"replace_options,omitempty"` // if true, replace all existing options with provided list
}
