package systemmetrics

import (
	"time"
)

type SystemMetrics struct {
	ID          string     `json:"id" gorm:"column:id;primaryKey"`
	MetricName  string     `json:"metric_name" gorm:"column:metric_name"`
	MetricValue float64    `json:"metric_value" gorm:"column:metric_value"`
	MetricType  string     `json:"metric_type" gorm:"column:metric_type"`
	RecordedAt  *time.Time `json:"recorded_at,omitempty" gorm:"column:recorded_at"`
}

func (SystemMetrics) TableName() string {
	return "gnyx.system_metrics"
}
