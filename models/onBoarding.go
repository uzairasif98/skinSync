package models

// Onboarding masters & junctions (steps)
//Step 1
type SkinTypeQuestion struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement"`
	QuestionText string           `gorm:"size:255"`
	Options      []SkinTypeOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
type SkinTypeOption struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64 `gorm:"index;not null"`
	Name       string `gorm:"size:255;uniqueIndex"`
}

type SkinType struct {
	ID         uint64           `gorm:"primaryKey;autoIncrement"`
	UserID     uint64           `gorm:"index;not null"`
	QuestionID uint64           `gorm:"index;not null"`
	TypeID     uint64           `gorm:"index;not null"`
	Question   SkinTypeQuestion `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE"`
}

//Step 2
type ConcernQuestion struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement"`
	QuestionText string          `gorm:"size:255"`
	Options      []ConcernOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
type ConcernOption struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64 `gorm:"index;not null"`
	Name       string `gorm:"size:255;uniqueIndex"`
}

type SkinConcern struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	UserID     uint64 `gorm:"index;not null"`
	QuestionID uint64 `gorm:"index;not null"`
	ConcernID  uint64 `gorm:"index;not null"`

	Question ConcernQuestion `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE"`
}

// Step 3
type LifeStyleQuestion struct {
	ID           uint64                       `gorm:"primaryKey;autoIncrement"`
	QuestionText string                       `gorm:"size:255"`
	Options      []LifeStyleDescriptionOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
type LifeStyleDescriptionOption struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64 `gorm:"index;not null"`
	Name       string `gorm:"size:255;uniqueIndex"`
}

type LifeStyleDescription struct {
	ID            uint64            `gorm:"primaryKey;autoIncrement"`
	UserID        uint64            `gorm:"index;not null"`
	QuestionID    uint64            `gorm:"index;not null"`
	DescriptionID uint64            `gorm:"index;not null"`
	Question      LifeStyleQuestion `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE"`
}

// Questions & options (step 4)

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

// Goals (step 5)
type SkinGoalQuestion struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement"`
	QuestionText string     `gorm:"size:255"`
	Options      []SkinGoal `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}

type SkinGoal struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	QuestionID uint64 `gorm:"index;not null"`
	Name       string `gorm:"size:255;uniqueIndex"`
}

type UserSkinGoal struct {
	ID         uint64           `gorm:"primaryKey;autoIncrement"`
	UserID     uint64           `gorm:"index;not null"`
	QuestionID uint64           `gorm:"index;not null"`
	GoalID     uint64           `gorm:"index;not null"`
	Question   SkinGoalQuestion `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE"`
}
