package trainingbadges

import (
	"time"
)

type TrainingBadges struct {
	ID               string     `json:"id" gorm:"column:id;primaryKey"`
	BadgeType        string     `json:"badge_type" gorm:"column:badge_type"`
	BadgeName        string     `json:"badge_name" gorm:"column:badge_name"`
	BadgeDescription *string    `json:"badge_description,omitempty" gorm:"column:badge_description"`
	BadgeIcon        *string    `json:"badge_icon,omitempty" gorm:"column:badge_icon"`
	EarnedDate       *time.Time `json:"earned_date,omitempty" gorm:"column:earned_date"`
	CreatedAt        *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (TrainingBadges) TableName() string {
	return "public.training_badges"
}
