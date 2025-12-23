package trainingcourses

import (
	"time"
)

type TrainingCourses struct {
	ID                  string     `json:"id" gorm:"column:id;primaryKey"`
	Title               string     `json:"title" gorm:"column:title"`
	Slug                string     `json:"slug" gorm:"column:slug"`
	Description         *string    `json:"description,omitempty" gorm:"column:description"`
	InstructorName      *string    `json:"instructor_name,omitempty" gorm:"column:instructor_name"`
	InstructorAvatarURL *string    `json:"instructor_avatar_url,omitempty" gorm:"column:instructor_avatar_url"`
	CourseImageURL      *string    `json:"course_image_url,omitempty" gorm:"column:course_image_url"`
	Category            *string    `json:"category,omitempty" gorm:"column:category"`
	DifficultyLevel     *string    `json:"difficulty_level,omitempty" gorm:"column:difficulty_level"`
	IsActive            *bool      `json:"is_active,omitempty" gorm:"column:is_active"`
	CreatedAt           *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (TrainingCourses) TableName() string {
	return "gnyx.training_courses"
}
