package tenantsubscriptions

import (
	"time"
)

type TenantSubscriptions struct {
	ID              *string    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantID        string     `json:"tenant_id" gorm:"column:tenant_id"`
	PlanID          string     `json:"plan_id" gorm:"column:plan_id"`
	StartDate       time.Time  `json:"start_date" gorm:"column:start_date"`
	EndDate         *time.Time `json:"end_date,omitempty" gorm:"column:end_date"`
	Status          string     `json:"status" gorm:"column:status"`
	PaymentMethodID *string    `json:"payment_method_id,omitempty" gorm:"column:payment_method_id"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (TenantSubscriptions) TableName() string {
	return "gnyx.tenant_subscriptions"
}
