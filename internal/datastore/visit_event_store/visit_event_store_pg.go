package visiteventstore

import (
	"context"
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
	visitTable = "limalar.visit_events"
	visitCols  = "id, tenant_id, agent_id, lead_id, property_id, scheduled_at, status, notes, address, created_at, updated_at"
)

type pgVisitEventStore struct {
	exec execution.PGExecutor
}

func NewPGVisitEventStore(exec execution.PGExecutor) VisitEventStore {
	return &pgVisitEventStore{exec: exec}
}

func (s *pgVisitEventStore) Create(ctx context.Context, input *CreateVisitEventInput) (*VisitEvent, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	status := input.Status
	if status == "" {
		status = "scheduled"
	}
	var v VisitEvent
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, agent_id, lead_id, property_id, scheduled_at, status, notes, address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s
	`, visitTable, visitCols),
		input.TenantID, input.AgentID, input.LeadID, input.PropertyID,
		input.ScheduledAt, status, input.Notes, input.Address,
	).Scan(&v.ID, &v.TenantID, &v.AgentID, &v.LeadID, &v.PropertyID,
		&v.ScheduledAt, &v.Status, &v.Notes, &v.Address, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create visit event: %w", err)
	}
	return &v, nil
}

func (s *pgVisitEventStore) GetByID(ctx context.Context, tenantID, id string) (*VisitEvent, error) {
	var v VisitEvent
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM %s WHERE id = $1 AND tenant_id = $2
	`, visitCols, visitTable), id, tenantID).Scan(
		&v.ID, &v.TenantID, &v.AgentID, &v.LeadID, &v.PropertyID,
		&v.ScheduledAt, &v.Status, &v.Notes, &v.Address, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("visit event not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get visit event: %w", err)
	}
	return &v, nil
}

func (s *pgVisitEventStore) List(ctx context.Context, filters *VisitEventFilters) (*t.PaginatedResult[VisitEvent], error) {
	if filters == nil {
		filters = &VisitEventFilters{}
	}
	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildVisitFilters(filters)
	q := fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY scheduled_at ASC LIMIT $%d OFFSET $%d`,
		visitCols, visitTable, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list visit events: %w", err)
	}
	defer rows.Close()

	visits := []VisitEvent{}
	for rows.Next() {
		var v VisitEvent
		if err := rows.Scan(&v.ID, &v.TenantID, &v.AgentID, &v.LeadID, &v.PropertyID,
			&v.ScheduledAt, &v.Status, &v.Notes, &v.Address, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}

	cWhere, cArgs := buildVisitFilters(filters)
	var total int64
	_ = s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, visitTable, cWhere), cArgs...).Scan(&total)

	tp := int(total) / limit
	if int(total)%limit != 0 {
		tp++
	}
	return &t.PaginatedResult[VisitEvent]{Data: visits, Total: total, Page: page, Limit: limit, TotalPages: tp}, nil
}

func (s *pgVisitEventStore) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3
	`, visitTable), status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update visit status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("visit event not found: %v", t.ErrNotFound)
	}
	return nil
}

func (s *pgVisitEventStore) ListByDay(ctx context.Context, tenantID, agentID string, day time.Time) ([]VisitEvent, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE tenant_id = $1 AND agent_id = $2
		  AND scheduled_at >= $3 AND scheduled_at < $4
		  AND status NOT IN ('cancelled', 'no_show')
		ORDER BY scheduled_at ASC
	`, visitCols, visitTable), tenantID, agentID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list visits by day: %w", err)
	}
	defer rows.Close()

	var visits []VisitEvent
	for rows.Next() {
		var v VisitEvent
		if err := rows.Scan(&v.ID, &v.TenantID, &v.AgentID, &v.LeadID, &v.PropertyID,
			&v.ScheduledAt, &v.Status, &v.Notes, &v.Address, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

func (s *pgVisitEventStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[VisitEvent](), "pg_visit_event_store", nil
}
func (s *pgVisitEventStore) GetName() string { return "pg_visit_event_store" }
func (s *pgVisitEventStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgVisitEventStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildVisitFilters(f *VisitEventFilters) (string, []any) {
	var where []string
	var args []any
	if f.TenantID != "" {
		args = append(args, f.TenantID)
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)))
	}
	if f.AgentID != nil {
		args = append(args, *f.AgentID)
		where = append(where, fmt.Sprintf("agent_id = $%d", len(args)))
	}
	if f.LeadID != nil {
		args = append(args, *f.LeadID)
		where = append(where, fmt.Sprintf("lead_id = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if f.From != nil {
		args = append(args, *f.From)
		where = append(where, fmt.Sprintf("scheduled_at >= $%d", len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		where = append(where, fmt.Sprintf("scheduled_at < $%d", len(args)))
	}
	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(p int) int {
	if p < 1 {
		return 1
	}
	return p
}

func normalizeLimit(l, def, max int) int {
	if l <= 0 {
		return def
	}
	if l > max {
		return max
	}
	return l
}

// suppress unused import
var _ = time.Now
