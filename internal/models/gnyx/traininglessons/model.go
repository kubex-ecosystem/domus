package traininglessons

import (
	"time"
)

type TrainingLessons struct {
	ID          string     `json:"id" gorm:"column:id;primaryKey"`
	Title       string     `json:"title" gorm:"column:title"`
	Slug        string     `json:"slug" gorm:"column:slug"`
	Description *string    `json:"description,omitempty" gorm:"column:description"`
	ContentType string     `json:"content_type" gorm:"column:content_type"`
	ContentURL  *string    `json:"content_url,omitempty" gorm:"column:content_url"`
	ContentText *string    `json:"content_text,omitempty" gorm:"column:content_text"`
	IsRequired  *bool      `json:"is_required,omitempty" gorm:"column:is_required"`
	CreatedAt   *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (TrainingLessons) TableName() string {
	return "gnyx.training_lessons"
}
