package activities

import (
	"time"
)

type Activities struct {
	ID              string     `json:"id" gorm:"column:id;primaryKey"`
	Title           string     `json:"title" gorm:"column:title"`
	Description     *string    `json:"description,omitempty" gorm:"column:description"`
	Type            string     `json:"type" gorm:"column:type"`
	AssignedTo      string     `json:"assigned_to" gorm:"column:assigned_to"`
	CreatedBy       string     `json:"created_by" gorm:"column:created_by"`
	CompanyID       string     `json:"company_id" gorm:"column:company_id"`
	StartTime       time.Time  `json:"start_time" gorm:"column:start_time"`
	EndTime         *time.Time `json:"end_time,omitempty" gorm:"column:end_time"`
	Status          *string    `json:"status,omitempty" gorm:"column:status"`
	Priority        *string    `json:"priority,omitempty" gorm:"column:priority"`
	Location        *string    `json:"location,omitempty" gorm:"column:location"`
	Notes           *string    `json:"notes,omitempty" gorm:"column:notes"`
	IsRecurring     *bool      `json:"is_recurring,omitempty" gorm:"column:is_recurring"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	PostponedFrom   *time.Time `json:"postponed_from,omitempty" gorm:"column:postponed_from"`
	PostponedReason *string    `json:"postponed_reason,omitempty" gorm:"column:postponed_reason"`
	CreatedAt       *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (Activities) TableName() string {
	return "gnyx.activities"
}
