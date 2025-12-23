package paymentmethods

import (
	"encoding/json"
	"time"
)

type PaymentMethods struct {
	ID          *string         `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name        string          `json:"name" gorm:"column:name"`
	Description *string         `json:"description,omitempty" gorm:"column:description"`
	Provider    string          `json:"provider" gorm:"column:provider"`
	Details     json.RawMessage `json:"details" gorm:"column:details"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (PaymentMethods) TableName() string {
	return "gnyx.payment_methods"
}
