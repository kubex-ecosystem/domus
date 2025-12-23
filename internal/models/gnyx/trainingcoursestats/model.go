package trainingcoursestats

import (
	"time"
)

type TrainingCourseStats struct {
	CourseID              string     `json:"course_id" gorm:"column:course_id"`
	AverageCompletionRate *float64   `json:"average_completion_rate,omitempty" gorm:"column:average_completion_rate"`
	LastUpdated           *time.Time `json:"last_updated,omitempty" gorm:"column:last_updated"`
}

func (TrainingCourseStats) TableName() string {
	return "gnyx.training_course_stats"
}
