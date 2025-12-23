package subscriptionplans

import (
	"encoding/json"
	"time"
)

type SubscriptionPlans struct {
	ID           *string         `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name         string          `json:"name" gorm:"column:name"`
	Description  *string         `json:"description,omitempty" gorm:"column:description"`
	Price        float64         `json:"price" gorm:"column:price"`
	BillingCycle string          `json:"billing_cycle" gorm:"column:billing_cycle"`
	Features     json.RawMessage `json:"features" gorm:"column:features"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (SubscriptionPlans) TableName() string {
	return "gnyx.subscription_plans"
}
