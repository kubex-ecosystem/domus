package companymetrics

import (
	"time"
)

type CompanyMetrics struct {
	ID                 string     `json:"id" gorm:"column:id;primaryKey"`
	Month              string     `json:"month" gorm:"column:month"`
	TotalPipelineValue *float64   `json:"total_pipeline_value,omitempty" gorm:"column:total_pipeline_value"`
	WonValue           *float64   `json:"won_value,omitempty" gorm:"column:won_value"`
	LostValue          *float64   `json:"lost_value,omitempty" gorm:"column:lost_value"`
	ConversionRate     *float64   `json:"conversion_rate,omitempty" gorm:"column:conversion_rate"`
	AverageDealSize    *float64   `json:"average_deal_size,omitempty" gorm:"column:average_deal_size"`
	LastCalculatedAt   *time.Time `json:"last_calculated_at,omitempty" gorm:"column:last_calculated_at"`
	CreatedAt          *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (CompanyMetrics) TableName() string {
	return "gnyx.company_metrics"
}
