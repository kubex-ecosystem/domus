package bmoristore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

const (
	biometricsTableName        = "bmori.biometrics"
	churnRiskProfilesTableName = "bmori.churn_risk_profiles"
)

// pgBMoriStore implementa BMoriStore sobre PGExecutor.
type pgBMoriStore struct {
	exec execution.PGExecutor
}

// NewPGBMoriStore cria uma instância de BMoriStore para Postgres.
func NewPGBMoriStore(exec execution.PGExecutor) BMoriStore {
	return &pgBMoriStore{exec: exec}
}

func (s *pgBMoriStore) SaveBiometrics(ctx context.Context, input *BiometricsRecord) error {
	if input == nil {
		return fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" || strings.TrimSpace(input.ClientID) == "" {
		return fmt.Errorf("tenant_id and client_id are required: %v", t.ErrInvalidInput)
	}

	recordedAt := input.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now().UTC()
	}

	payload := input.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal biometrics payload: %v", err)
	}

	var personalUserID *string
	if strings.TrimSpace(input.PersonalUserID) != "" {
		personalUserID = &input.PersonalUserID
	}

	_, err = s.exec.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			tenant_id, student_user_id, personal_user_id, recorded_at,
			source_device, source_vendor, resting_heart_rate, avg_heart_rate,
			steps_count, sleep_minutes, calories_burned, distance_meters, payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, biometricsTableName),
		input.TenantID,
		input.ClientID,
		personalUserID,
		recordedAt,
		input.SourceDevice,
		input.SourceVendor,
		input.RestingHeartRate,
		input.AvgHeartRate,
		input.StepsCount,
		input.SleepMinutes,
		input.CaloriesBurned,
		input.DistanceMeters,
		payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save biometrics: %v", err)
	}
	return nil
}

func (s *pgBMoriStore) GetClientFatigueMetrics(ctx context.Context, clientID string) (*ClientFatigueMetrics, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("clientID is required: %v", t.ErrInvalidInput)
	}

	row := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			student_user_id::text AS client_id,
			tenant_id::text,
			COALESCE(MAX(personal_user_id::text), '') AS personal_user_id,
			MAX(recorded_at) AS last_recorded_at,
			COALESCE(AVG(resting_heart_rate)::double precision, 0) AS avg_resting_heart_rate,
			COALESCE(AVG(avg_heart_rate)::double precision, 0) AS avg_heart_rate,
			COALESCE(AVG(sleep_minutes)::double precision, 0) AS avg_sleep_minutes,
			COALESCE(AVG(steps_count)::double precision, 0) AS avg_steps_count,
			COALESCE(AVG(calories_burned)::double precision, 0) AS avg_calories_burned,
			COALESCE(AVG(distance_meters)::double precision, 0) AS avg_distance_meters,
			COUNT(*) AS records_considered,
			7 AS assessment_window_in_days,
			LEAST(
				100,
				GREATEST(
					0,
					(100 - COALESCE(AVG(sleep_minutes)::double precision / 6, 0))
					+ COALESCE(AVG(avg_heart_rate)::double precision / 3, 0)
					- COALESCE(AVG(steps_count)::double precision / 250, 0)
				)
			) AS fatigue_score
		FROM %s
		WHERE student_user_id = $1
		  AND recorded_at >= NOW() - INTERVAL '7 days'
		GROUP BY student_user_id, tenant_id
		ORDER BY MAX(recorded_at) DESC
		LIMIT 1
	`, biometricsTableName), clientID)

	var metrics ClientFatigueMetrics
	var lastRecordedAt *time.Time
	if err := row.Scan(
		&metrics.ClientID,
		&metrics.TenantID,
		&metrics.PersonalUserID,
		&lastRecordedAt,
		&metrics.AvgRestingHeartRate,
		&metrics.AvgHeartRate,
		&metrics.AvgSleepMinutes,
		&metrics.AvgStepsCount,
		&metrics.AvgCaloriesBurned,
		&metrics.AvgDistanceMeters,
		&metrics.RecordsConsidered,
		&metrics.AssessmentWindowInDays,
		&metrics.FatigueScore,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get client fatigue metrics: %v", err)
	}
	metrics.LastRecordedAt = lastRecordedAt
	return &metrics, nil
}

func (s *pgBMoriStore) ListRecentBiometrics(ctx context.Context, clientID string, limit int) ([]BiometricsRecord, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("clientID is required: %v", t.ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT
			id::text,
			tenant_id::text,
			student_user_id::text,
			COALESCE(personal_user_id::text, ''),
			recorded_at,
			source_device,
			source_vendor,
			resting_heart_rate,
			avg_heart_rate,
			steps_count,
			sleep_minutes,
			calories_burned,
			distance_meters,
			payload,
			created_at,
			updated_at
		FROM %s
		WHERE student_user_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`, biometricsTableName), clientID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent biometrics: %v", err)
	}
	defer rows.Close()

	items := make([]BiometricsRecord, 0, limit)
	for rows.Next() {
		var item BiometricsRecord
		var payloadRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.ClientID,
			&item.PersonalUserID,
			&item.RecordedAt,
			&item.SourceDevice,
			&item.SourceVendor,
			&item.RestingHeartRate,
			&item.AvgHeartRate,
			&item.StepsCount,
			&item.SleepMinutes,
			&item.CaloriesBurned,
			&item.DistanceMeters,
			&payloadRaw,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan biometrics record: %v", err)
		}
		if len(payloadRaw) > 0 {
			if err := json.Unmarshal(payloadRaw, &item.Payload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal biometrics payload: %v", err)
			}
		}
		if item.Payload == nil {
			item.Payload = map[string]any{}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return items, nil
}

func (s *pgBMoriStore) GetHighRiskClients(ctx context.Context, tenantID string) ([]HighRiskClient, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenantID is required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT
			id::text,
			tenant_id::text,
			student_user_id::text AS client_id,
			COALESCE(personal_user_id::text, '') AS personal_user_id,
			risk_score::double precision,
			risk_level,
			status,
			last_biometrics_at,
			last_assessment_at,
			notes
		FROM %s
		WHERE tenant_id = $1
		  AND status = 'active'
		  AND (
				risk_level IN ('high', 'critical')
				OR risk_score >= 70
		  )
		ORDER BY risk_score DESC, last_assessment_at DESC
	`, churnRiskProfilesTableName), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list high risk clients: %v", err)
	}
	defer rows.Close()

	clients := []HighRiskClient{}
	for rows.Next() {
		var item HighRiskClient
		if err := rows.Scan(
			&item.ProfileID,
			&item.TenantID,
			&item.ClientID,
			&item.PersonalUserID,
			&item.RiskScore,
			&item.RiskLevel,
			&item.Status,
			&item.LastBiometricsAt,
			&item.LastAssessmentAt,
			&item.Notes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan high risk client: %v", err)
		}
		clients = append(clients, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return clients, nil
}

func (s *pgBMoriStore) UpsertChurnRiskProfile(ctx context.Context, input *UpsertChurnRiskProfileInput) (*ChurnRiskProfile, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" || strings.TrimSpace(input.ClientID) == "" {
		return nil, fmt.Errorf("tenant_id and client_id are required: %v", t.ErrInvalidInput)
	}

	riskLevel := strings.TrimSpace(input.RiskLevel)
	if riskLevel == "" {
		riskLevel = "low"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	windowDays := input.AssessmentWindowDays
	if windowDays <= 0 {
		windowDays = 14
	}

	factors := input.Factors
	if factors == nil {
		factors = []map[string]any{}
	}
	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal churn factors: %v", err)
	}

	actions := input.RecommendedActions
	if actions == nil {
		actions = []map[string]any{}
	}
	actionsJSON, err := json.Marshal(actions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal churn recommended actions: %v", err)
	}

	var personalUserID *string
	if strings.TrimSpace(input.PersonalUserID) != "" {
		personalUserID = &input.PersonalUserID
	}

	row := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			tenant_id,
			student_user_id,
			personal_user_id,
			risk_score,
			risk_level,
			status,
			assessment_window_days,
			last_biometrics_at,
			last_assessment_at,
			factors,
			recommended_actions,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9, $10, $11)
		ON CONFLICT (tenant_id, student_user_id) DO UPDATE
		SET
			personal_user_id = EXCLUDED.personal_user_id,
			risk_score = EXCLUDED.risk_score,
			risk_level = EXCLUDED.risk_level,
			status = EXCLUDED.status,
			assessment_window_days = EXCLUDED.assessment_window_days,
			last_biometrics_at = EXCLUDED.last_biometrics_at,
			last_assessment_at = NOW(),
			factors = EXCLUDED.factors,
			recommended_actions = EXCLUDED.recommended_actions,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING
			id::text,
			tenant_id::text,
			student_user_id::text,
			COALESCE(personal_user_id::text, ''),
			risk_score::double precision,
			risk_level,
			status,
			assessment_window_days,
			last_biometrics_at,
			last_assessment_at,
			factors,
			recommended_actions,
			notes,
			created_at,
			updated_at
	`, churnRiskProfilesTableName),
		input.TenantID,
		input.ClientID,
		personalUserID,
		input.RiskScore,
		riskLevel,
		status,
		windowDays,
		input.LastBiometricsAt,
		factorsJSON,
		actionsJSON,
		input.Notes,
	)

	var profile ChurnRiskProfile
	var factorsRaw []byte
	var actionsRaw []byte
	if err := row.Scan(
		&profile.ID,
		&profile.TenantID,
		&profile.ClientID,
		&profile.PersonalUserID,
		&profile.RiskScore,
		&profile.RiskLevel,
		&profile.Status,
		&profile.AssessmentWindowDays,
		&profile.LastBiometricsAt,
		&profile.LastAssessmentAt,
		&factorsRaw,
		&actionsRaw,
		&profile.Notes,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to upsert churn risk profile: %v", err)
	}

	if len(factorsRaw) > 0 {
		if err := json.Unmarshal(factorsRaw, &profile.Factors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal churn factors: %v", err)
		}
	}
	if len(actionsRaw) > 0 {
		if err := json.Unmarshal(actionsRaw, &profile.RecommendedActions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal churn recommended actions: %v", err)
		}
	}
	if profile.Factors == nil {
		profile.Factors = []map[string]any{}
	}
	if profile.RecommendedActions == nil {
		profile.RecommendedActions = []map[string]any{}
	}
	return &profile, nil
}

func (s *pgBMoriStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[BiometricsRecord](), "pg_bmori_store", nil
}

func (s *pgBMoriStore) GetName() string {
	return "pg_bmori_store"
}

func (s *pgBMoriStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgBMoriStore) Close() error {
	return nil
}
