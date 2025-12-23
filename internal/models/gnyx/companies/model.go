package companies

import (
	"time"
)

type Companies struct {
	ID            string     `json:"id" gorm:"column:id;primaryKey"`
	Name          string     `json:"name" gorm:"column:name"`
	Slug          string     `json:"slug" gorm:"column:slug"`
	CreatedAt     *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
	PlanExpiresAt *time.Time `json:"plan_expires_at,omitempty" gorm:"column:plan_expires_at"`
	IsTrial       *bool      `json:"is_trial,omitempty" gorm:"column:is_trial"`
	IsActive      *bool      `json:"is_active,omitempty" gorm:"column:is_active"`
	Domain        *string    `json:"domain,omitempty" gorm:"column:domain"`
	Phone         *string    `json:"phone,omitempty" gorm:"column:phone"`
	Address       *string    `json:"address,omitempty" gorm:"column:address"`
}

func (Companies) TableName() string {
	return "gnyx.companies"
}
