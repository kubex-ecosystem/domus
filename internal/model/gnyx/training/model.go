package training

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// TrainingCourseStat espelha training_course_stats.
type TrainingCourseStat struct {
	CourseID              gnyx.UUID       `json:"course_id" db:"course_id"`
	TotalLessons          *int64          `json:"total_lessons,omitempty" db:"total_lessons"`
	TotalEnrollments      *int64          `json:"total_enrollments,omitempty" db:"total_enrollments"`
	TotalCompletions      *int64          `json:"total_completions,omitempty" db:"total_completions"`
	AverageCompletionRate *float64        `json:"average_completion_rate,omitempty" db:"average_completion_rate"`
	LastUpdated           *gnyx.Timestamp `json:"last_updated,omitempty" db:"last_updated"`
}

// TrainingCourse representa training_courses.
type TrainingCourse struct {
	ID                  gnyx.UUID       `json:"id" db:"id"`
	Title               string          `json:"title" db:"title"`
	Slug                string          `json:"slug" db:"slug"`
	Description         *string         `json:"description,omitempty" db:"description"`
	InstructorName      *string         `json:"instructor_name,omitempty" db:"instructor_name"`
	InstructorAvatarURL *string         `json:"instructor_avatar_url,omitempty" db:"instructor_avatar_url"`
	CourseImageURL      *string         `json:"course_image_url,omitempty" db:"course_image_url"`
	Category            *string         `json:"category,omitempty" db:"category"`
	EstimatedDuration   *int64          `json:"estimated_duration,omitempty" db:"estimated_duration"`
	DifficultyLevel     *string         `json:"difficulty_level,omitempty" db:"difficulty_level"`
	IsActive            *bool           `json:"is_active,omitempty" db:"is_active"`
	CompanyID           *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	CreatedBy           *gnyx.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt           *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// TrainingLesson representa training_lessons.
type TrainingLesson struct {
	ID                gnyx.UUID       `json:"id" db:"id"`
	CourseID          *gnyx.UUID      `json:"course_id,omitempty" db:"course_id"`
	Title             string          `json:"title" db:"title"`
	Slug              string          `json:"slug" db:"slug"`
	Description       *string         `json:"description,omitempty" db:"description"`
	LessonOrder       int64           `json:"lesson_order" db:"lesson_order"`
	ContentType       string          `json:"content_type" db:"content_type"`
	ContentURL        *string         `json:"content_url,omitempty" db:"content_url"`
	ContentText       *string         `json:"content_text,omitempty" db:"content_text"`
	EstimatedDuration *int64          `json:"estimated_duration,omitempty" db:"estimated_duration"`
	IsRequired        *bool           `json:"is_required,omitempty" db:"is_required"`
	CreatedAt         *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt         *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// TrainingProgress espelha training_progress.
type TrainingProgress struct {
	ID             gnyx.UUID       `json:"id" db:"id"`
	UserID         *gnyx.UUID      `json:"user_id,omitempty" db:"user_id"`
	CourseID       *gnyx.UUID      `json:"course_id,omitempty" db:"course_id"`
	LessonID       *gnyx.UUID      `json:"lesson_id,omitempty" db:"lesson_id"`
	IsCompleted    *bool           `json:"is_completed,omitempty" db:"is_completed"`
	CompletionDate *gnyx.Timestamp `json:"completion_date,omitempty" db:"completion_date"`
	TimeSpent      *int64          `json:"time_spent,omitempty" db:"time_spent"`
	CreatedAt      *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}
