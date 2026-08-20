package siftleadeventstore

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

type pgSiftLeadEventStore struct {
	exec execution.PGExecutor
}

func NewPGSiftLeadEventStore(exec execution.PGExecutor) SiftLeadEventStore {
	return &pgSiftLeadEventStore{exec: exec}
}

const eventCols = `id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at`

func (s *pgSiftLeadEventStore) Create(ctx context.Context, input *CreateLeadEventInput) (*LeadEvent, error) {
	if input == nil || strings.TrimSpace(input.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.LeadID) == "" {
		return nil, fmt.Errorf("lead_id is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Type) == "" {
		return nil, fmt.Errorf("type is required: %w", t.ErrInvalidInput)
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to encode lead event metadata: %w", err)
	}

	// O lead é conferido contra o mesmo tenant na subconsulta: sem isso um
	// lead_id de outro tenant entraria na trilha deste, e a FK do banco não
	// pegaria — ela só garante que o lead existe, não de quem ele é.
	row := s.exec.QueryRow(ctx, `
		INSERT INTO sift.lead_event (tenant_id, lead_id, type, body, actor_user_id, metadata)
		SELECT $1, $2, $3, $4, $5, $6::jsonb
		WHERE EXISTS (SELECT 1 FROM sift.lead l WHERE l.id = $2 AND l.tenant_id = $1)
		RETURNING `+eventCols,
		input.TenantID, input.LeadID, input.Type, input.Body, input.ActorUserID, string(raw),
	)
	ev, err := scanLeadEvent(row)
	if err != nil {
		// Sem linha aqui significa que o lead não é deste tenant (ou não
		// existe). Não confirmar existência de id alheio: o chamador recebe o
		// mesmo (nil,nil) nos dois casos.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to create sift lead event: %w", err)
	}
	return ev, nil
}

func (s *pgSiftLeadEventStore) ListByLead(ctx context.Context, tenantID, leadID string) ([]LeadEvent, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(leadID) == "" {
		return nil, fmt.Errorf("tenant_id and lead_id are required: %w", t.ErrInvalidInput)
	}
	rows, err := s.exec.Query(ctx, `
		SELECT `+eventCols+` FROM sift.lead_event
		WHERE tenant_id = $1 AND lead_id = $2
		ORDER BY created_at ASC
	`, tenantID, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sift lead events: %w", err)
	}
	defer rows.Close()

	events := []LeadEvent{}
	for rows.Next() {
		ev, err := scanLeadEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sift lead event: %w", err)
		}
		events = append(events, *ev)
	}
	return events, rows.Err()
}

func scanLeadEvent(scanner interface{ Scan(...any) error }) (*LeadEvent, error) {
	var ev LeadEvent
	var createdAt time.Time
	var raw []byte
	if err := scanner.Scan(
		&ev.ID, &ev.TenantID, &ev.LeadID, &ev.Type, &ev.Body,
		&ev.ActorUserID, &raw, &createdAt,
	); err != nil {
		return nil, err
	}
	ev.CreatedAt = &createdAt
	ev.Metadata = map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &ev.Metadata); err != nil {
			return nil, fmt.Errorf("failed to decode lead event metadata: %w", err)
		}
	}
	return &ev, nil
}

func (s *pgSiftLeadEventStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[LeadEvent](), "pg_sift_lead_event_store", nil
}
func (s *pgSiftLeadEventStore) GetName() string { return "pg_sift_lead_event_store" }
func (s *pgSiftLeadEventStore) Name() string    { return "sift_lead_event" }
func (s *pgSiftLeadEventStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgSiftLeadEventStore) Close() error { return nil }
