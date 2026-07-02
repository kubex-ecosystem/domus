// Package bmoristore fornece o store de domínio para o módulo BMori (HealthTech).
package bmoristore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// BiometricsRecord representa um registro persistido em bmori.biometrics.
type BiometricsRecord struct {
	ID               string         `json:"id,omitempty"`
	TenantID         string         `json:"tenant_id"`
	ClientID         string         `json:"client_id"`
	PersonalUserID   string         `json:"personal_user_id,omitempty"`
	RecordedAt       time.Time      `json:"recorded_at"`
	SourceDevice     string         `json:"source_device"`
	SourceVendor     string         `json:"source_vendor"`
	RestingHeartRate *int           `json:"resting_heart_rate,omitempty"`
	AvgHeartRate     *int           `json:"avg_heart_rate,omitempty"`
	StepsCount       int            `json:"steps_count"`
	SleepMinutes     int            `json:"sleep_minutes"`
	CaloriesBurned   int            `json:"calories_burned"`
	DistanceMeters   int            `json:"distance_meters"`
	Payload          map[string]any `json:"payload,omitempty"`
	CreatedAt        *time.Time     `json:"created_at,omitempty"`
	UpdatedAt        *time.Time     `json:"updated_at,omitempty"`
}

// ClientFatigueMetrics representa a visão agregada de fadiga de um aluno.
type ClientFatigueMetrics struct {
	ClientID               string     `json:"client_id"`
	TenantID               string     `json:"tenant_id"`
	PersonalUserID         string     `json:"personal_user_id,omitempty"`
	LastRecordedAt         *time.Time `json:"last_recorded_at,omitempty"`
	AvgRestingHeartRate    float64    `json:"avg_resting_heart_rate"`
	AvgHeartRate           float64    `json:"avg_heart_rate"`
	AvgSleepMinutes        float64    `json:"avg_sleep_minutes"`
	AvgStepsCount          float64    `json:"avg_steps_count"`
	AvgCaloriesBurned      float64    `json:"avg_calories_burned"`
	AvgDistanceMeters      float64    `json:"avg_distance_meters"`
	FatigueScore           float64    `json:"fatigue_score"`
	RecordsConsidered      int64      `json:"records_considered"`
	AssessmentWindowInDays int        `json:"assessment_window_in_days"`
}

// HighRiskClient representa um aluno com risco elevado de churn.
type HighRiskClient struct {
	ProfileID        string     `json:"profile_id"`
	TenantID         string     `json:"tenant_id"`
	ClientID         string     `json:"client_id"`
	PersonalUserID   string     `json:"personal_user_id,omitempty"`
	RiskScore        float64    `json:"risk_score"`
	RiskLevel        string     `json:"risk_level"`
	Status           string     `json:"status"`
	LastBiometricsAt *time.Time `json:"last_biometrics_at,omitempty"`
	LastAssessmentAt time.Time  `json:"last_assessment_at"`
	Notes            string     `json:"notes"`
}

// ChurnRiskProfile representa um perfil persistido em bmori.churn_risk_profiles.
type ChurnRiskProfile struct {
	ID                   string           `json:"id"`
	TenantID             string           `json:"tenant_id"`
	ClientID             string           `json:"client_id"`
	PersonalUserID       string           `json:"personal_user_id,omitempty"`
	RiskScore            float64          `json:"risk_score"`
	RiskLevel            string           `json:"risk_level"`
	Status               string           `json:"status"`
	AssessmentWindowDays int              `json:"assessment_window_days"`
	LastBiometricsAt     *time.Time       `json:"last_biometrics_at,omitempty"`
	LastAssessmentAt     time.Time        `json:"last_assessment_at"`
	Factors              []map[string]any `json:"factors,omitempty"`
	RecommendedActions   []map[string]any `json:"recommended_actions,omitempty"`
	Notes                string           `json:"notes"`
	CreatedAt            *time.Time       `json:"created_at,omitempty"`
	UpdatedAt            *time.Time       `json:"updated_at,omitempty"`
}

// UpsertChurnRiskProfileInput define os campos atualizáveis do perfil preditivo.
type UpsertChurnRiskProfileInput struct {
	TenantID             string
	ClientID             string
	PersonalUserID       string
	RiskScore            float64
	RiskLevel            string
	Status               string
	AssessmentWindowDays int
	LastBiometricsAt     *time.Time
	Factors              []map[string]any
	RecommendedActions   []map[string]any
	Notes                string
}

// BMoriStore expõe o SIP público do módulo BMori para consumo interno.
type BMoriStore interface {
	t.StoreType

	SaveBiometrics(ctx context.Context, input *BiometricsRecord) error
	GetClientFatigueMetrics(ctx context.Context, clientID string) (*ClientFatigueMetrics, error)
	ListRecentBiometrics(ctx context.Context, clientID string, limit int) ([]BiometricsRecord, error)
	GetHighRiskClients(ctx context.Context, tenantID string) ([]HighRiskClient, error)
	UpsertChurnRiskProfile(ctx context.Context, input *UpsertChurnRiskProfileInput) (*ChurnRiskProfile, error)
}
