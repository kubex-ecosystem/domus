package userpreferences

import (
	"time"
)

type UserPreferences struct {
	ID                  *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserID              string    `json:"user_id" gorm:"column:user_id"`
	PreferenceType      string    `json:"preference_type" gorm:"column:preference_type"`
	PreferenceValueType string    `json:"preference_value_type" gorm:"column:preference_value_type"`
	PreferenceKey       string    `json:"preference_key" gorm:"column:preference_key"`
	PreferenceValue     *string   `json:"preference_value,omitempty" gorm:"column:preference_value"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (UserPreferences) TableName() string {
	return "gnyx.user_preferences"
}
