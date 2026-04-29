package activityparticipants

import (
	"time"
)

type ActivityParticipants struct {
	ID        string     `json:"id" gorm:"column:id;primaryKey"`
	Role      *string    `json:"role,omitempty" gorm:"column:role"`
	Status    *string    `json:"status,omitempty" gorm:"column:status"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (ActivityParticipants) TableName() string {
	return "gnyx.activity_participants"
}
