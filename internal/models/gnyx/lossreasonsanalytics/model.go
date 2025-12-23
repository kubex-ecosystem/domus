package lossreasonsanalytics

import (
	"time"
)

type LossReasonsAnalytics struct {
	ID             string     `json:"id" gorm:"column:id;primaryKey"`
	Month          string     `json:"month" gorm:"column:month"`
	Reason         string     `json:"reason" gorm:"column:reason"`
	TotalValueLost *float64   `json:"total_value_lost,omitempty" gorm:"column:total_value_lost"`
	Percentage     *float64   `json:"percentage,omitempty" gorm:"column:percentage"`
	CreatedAt      *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (LossReasonsAnalytics) TableName() string {
	return "gnyx.loss_reasons_analytics"
}
