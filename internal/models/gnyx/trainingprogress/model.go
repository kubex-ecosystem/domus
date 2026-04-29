package trainingprogress

import (
	"time"
)

type TrainingProgress struct {
	ID             string     `json:"id" gorm:"column:id;primaryKey"`
	IsCompleted    *bool      `json:"is_completed,omitempty" gorm:"column:is_completed"`
	CompletionDate *time.Time `json:"completion_date,omitempty" gorm:"column:completion_date"`
	CreatedAt      *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (TrainingProgress) TableName() string {
	return "gnyx.training_progress"
}
