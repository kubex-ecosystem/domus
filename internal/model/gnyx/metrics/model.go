package metrics

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// SystemLogLevel mirrors severity levels used for system logging.
type SystemLogLevel string

const (
	SystemLogLevelInfo     SystemLogLevel = "info"
	SystemLogLevelWarning  SystemLogLevel = "warning"
	SystemLogLevelError    SystemLogLevel = "error"
	SystemLogLevelCritical SystemLogLevel = "critical"
)

// SystemLog espelha a tabela system_logs.
type SystemLog struct {
	ID        gnyx.UUID       `json:"id" db:"id"`
	Level     SystemLogLevel  `json:"level" db:"level"`
	Message   string          `json:"message" db:"message"`
	Category  string          `json:"category" db:"category"`
	Metadata  gnyx.JSONValue  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
}

// SystemMetric representa system_metrics.
type SystemMetric struct {
	ID          gnyx.UUID       `json:"id" db:"id"`
	MetricName  string          `json:"metric_name" db:"metric_name"`
	MetricValue float64         `json:"metric_value" db:"metric_value"`
	MetricType  MetricType      `json:"metric_type" db:"metric_type"`
	Metadata    gnyx.JSONValue  `json:"metadata,omitempty" db:"metadata"`
	RecordedAt  *gnyx.Timestamp `json:"recorded_at,omitempty" db:"recorded_at"`
}

// CompanyMetric espelha company_metrics.
type CompanyMetric struct {
	ID                 gnyx.UUID       `json:"id" db:"id"`
	CompanyID          *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	Month              string          `json:"month" db:"month"`
	TotalLeads         *int64          `json:"total_leads,omitempty" db:"total_leads"`
	ActiveLeads        *int64          `json:"active_leads,omitempty" db:"active_leads"`
	WonLeads           *int64          `json:"won_leads,omitempty" db:"won_leads"`
	LostLeads          *int64          `json:"lost_leads,omitempty" db:"lost_leads"`
	TotalPipelineValue *float64        `json:"total_pipeline_value,omitempty" db:"total_pipeline_value"`
	WonValue           *float64        `json:"won_value,omitempty" db:"won_value"`
	LostValue          *float64        `json:"lost_value,omitempty" db:"lost_value"`
	ConversionRate     *float64        `json:"conversion_rate,omitempty" db:"conversion_rate"`
	AverageDealSize    *float64        `json:"average_deal_size,omitempty" db:"average_deal_size"`
	AverageSalesCycle  *int64          `json:"average_sales_cycle,omitempty" db:"average_sales_cycle"`
	LastCalculatedAt   *gnyx.Timestamp `json:"last_calculated_at,omitempty" db:"last_calculated_at"`
	CreatedAt          *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt          *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// LossReasonAnalytics representa loss_reasons_analytics.
type LossReasonAnalytics struct {
	ID             gnyx.UUID       `json:"id" db:"id"`
	CompanyID      *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	Month          string          `json:"month" db:"month"`
	Reason         string          `json:"reason" db:"reason"`
	Count          *int64          `json:"count,omitempty" db:"count"`
	TotalValueLost *float64        `json:"total_value_lost,omitempty" db:"total_value_lost"`
	Percentage     *float64        `json:"percentage,omitempty" db:"percentage"`
	CreatedAt      *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// RealTimeMetricsView representa a view real_time_metrics.
type RealTimeMetricsView struct {
	CompanyID      gnyx.UUID `json:"company_id" db:"company_id"`
	TotalLeads     int64     `json:"total_leads" db:"total_leads"`
	ActiveLeads    int64     `json:"active_leads" db:"active_leads"`
	WonLeads       int64     `json:"won_leads" db:"won_leads"`
	LostLeads      int64     `json:"lost_leads" db:"lost_leads"`
	PipelineValue  float64   `json:"pipeline_value" db:"pipeline_value"`
	WonValue       float64   `json:"won_value" db:"won_value"`
	LostValue      float64   `json:"lost_value" db:"lost_value"`
	ConversionRate float64   `json:"conversion_rate" db:"conversion_rate"`
	AvgDealSize    float64   `json:"avg_deal_size" db:"avg_deal_size"`
}
