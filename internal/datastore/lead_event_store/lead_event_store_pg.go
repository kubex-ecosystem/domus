package leadeventstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

const eventsTable = "limalar.lead_events"

type pgLeadEventStore struct {
	exec execution.PGExecutor
}

// NewPGLeadEventStore cria uma instância de LeadEventStore para Postgres.
func NewPGLeadEventStore(exec execution.PGExecutor) LeadEventStore {
	return &pgLeadEventStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgLeadEventStore) Create(ctx context.Context, input *CreateLeadEventInput) (*LeadEvent, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.TenantID == "" || input.LeadID == "" || input.Type == "" {
		return nil, fmt.Errorf("tenant_id, lead_id and type are required: %v", t.ErrInvalidInput)
	}

	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}

	var ev LeadEvent
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, lead_id, type, body, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, lead_id, type, body, metadata, created_at
	`, eventsTable),
		input.TenantID, input.LeadID, input.Type, input.Body, metadata,
	).Scan(&ev.ID, &ev.TenantID, &ev.LeadID, &ev.Type, &ev.Body, &ev.Metadata, &ev.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead event: %w", err)
	}
	return &ev, nil
}

// ── ListByLead ────────────────────────────────────────────────────────────────

func (s *pgLeadEventStore) ListByLead(ctx context.Context, tenantID, leadID string) ([]LeadEvent, error) {
	if tenantID == "" || leadID == "" {
		return nil, fmt.Errorf("tenantID and leadID are required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, lead_id, type, body, metadata, created_at
		FROM %s
		WHERE lead_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
		LIMIT 100
	`, eventsTable), leadID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list lead events: %w", err)
	}
	defer rows.Close()

	events := []LeadEvent{}
	for rows.Next() {
		var ev LeadEvent
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.LeadID, &ev.Type, &ev.Body, &ev.Metadata, &ev.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lead event: %w", err)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgLeadEventStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[LeadEvent](), "pg_lead_event_store", nil
}

func (s *pgLeadEventStore) GetName() string { return "pg_lead_event_store" }

func (s *pgLeadEventStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgLeadEventStore) Close() error { return nil }
