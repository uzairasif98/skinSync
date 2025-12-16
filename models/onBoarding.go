package models

// Onboarding masters & answers (single-step questions only)

type SkinConditionQuestion struct {
	ID           uint64                        `gorm:"primaryKey;autoIncrement"`
	QuestionText string                        `gorm:"size:255"`
	Options      []SkinConditionQuestionOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}

type SkinConditionQuestionOption struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64 `gorm:"index;not null"`
	OptionText string `gorm:"size:255"`
}

type SkinConditionQuestionAnswer struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	UserID     uint64 `gorm:"index;not null"`
	QuestionID uint64 `gorm:"index;not null"`
	OptionID   uint64 `gorm:"index;not null"`

	Question SkinConditionQuestion `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE"`
}
