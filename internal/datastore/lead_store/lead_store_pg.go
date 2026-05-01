package leadstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
	"github.com/jackc/pgx/v5"
)

const leadsTable = "limalar.leads"

type pgLeadStore struct {
	exec execution.PGExecutor
}

// NewPGLeadStore cria uma instância de LeadStore para Postgres.
func NewPGLeadStore(exec execution.PGExecutor) LeadStore {
	return &pgLeadStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgLeadStore) Create(ctx context.Context, input *CreateLeadInput) (*Lead, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required: %v", t.ErrInvalidInput)
	}

	status := input.Status
	if status == "" {
		status = "new"
	}
	source := input.Source
	if source == "" {
		source = "manual"
	}

	var lead Lead
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, name, email, phone, property_id, status, score, notes, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
	`, leadsTable),
		input.TenantID, input.Name, input.Email, input.Phone, input.PropertyID,
		status, input.Score, input.Notes, source,
	).Scan(
		&lead.ID, &lead.TenantID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.PropertyID, &lead.Status, &lead.Score, &lead.Notes, &lead.Source,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}
	return &lead, nil
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func (s *pgLeadStore) GetByID(ctx context.Context, tenantID, id string) (*Lead, error) {
	if tenantID == "" || id == "" {
		return nil, fmt.Errorf("tenantID and id are required: %v", t.ErrInvalidInput)
	}

	var lead Lead
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
		FROM %s WHERE id = $1 AND tenant_id = $2
	`, leadsTable), id, tenantID).Scan(
		&lead.ID, &lead.TenantID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.PropertyID, &lead.Status, &lead.Score, &lead.Notes, &lead.Source,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("lead not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}
	return &lead, nil
}

// ── GetByPhone ────────────────────────────────────────────────────────────────

func (s *pgLeadStore) GetByPhone(ctx context.Context, tenantID, phone string) (*Lead, error) {
	if tenantID == "" || phone == "" {
		return nil, fmt.Errorf("tenantID and phone are required: %v", t.ErrInvalidInput)
	}

	var lead Lead
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
		FROM %s WHERE phone = $1 AND tenant_id = $2
	`, leadsTable), phone, tenantID).Scan(
		&lead.ID, &lead.TenantID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.PropertyID, &lead.Status, &lead.Score, &lead.Notes, &lead.Source,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lead by phone: %w", err)
	}
	return &lead, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *pgLeadStore) Update(ctx context.Context, tenantID, id string, input *UpdateLeadInput) (*Lead, error) {
	if tenantID == "" || id == "" || input == nil {
		return nil, fmt.Errorf("tenantID, id and input are required: %v", t.ErrInvalidInput)
	}

	sets := []string{}
	args := []any{}

	if input.Name != nil {
		args = append(args, *input.Name)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}
	if input.Email != nil {
		args = append(args, *input.Email)
		sets = append(sets, fmt.Sprintf("email = $%d", len(args)))
	}
	if input.Phone != nil {
		args = append(args, *input.Phone)
		sets = append(sets, fmt.Sprintf("phone = $%d", len(args)))
	}
	if input.PropertyID != nil {
		args = append(args, input.PropertyID)
		sets = append(sets, fmt.Sprintf("property_id = $%d", len(args)))
	}
	if input.Status != nil {
		args = append(args, *input.Status)
		sets = append(sets, fmt.Sprintf("status = $%d", len(args)))
	}
	if input.Score != nil {
		args = append(args, *input.Score)
		sets = append(sets, fmt.Sprintf("score = $%d", len(args)))
	}
	if input.Notes != nil {
		args = append(args, *input.Notes)
		sets = append(sets, fmt.Sprintf("notes = $%d", len(args)))
	}
	if input.Source != nil {
		args = append(args, *input.Source)
		sets = append(sets, fmt.Sprintf("source = $%d", len(args)))
	}

	if len(sets) == 0 {
		return s.GetByID(ctx, tenantID, id)
	}

	args = append(args, id, tenantID)
	q := fmt.Sprintf(`
		UPDATE %s SET %s
		WHERE id = $%d AND tenant_id = $%d
		RETURNING id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
	`, leadsTable, strings.Join(sets, ", "), len(args)-1, len(args))

	var lead Lead
	err := s.exec.QueryRow(ctx, q, args...).Scan(
		&lead.ID, &lead.TenantID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.PropertyID, &lead.Status, &lead.Score, &lead.Notes, &lead.Source,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("lead not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update lead: %w", err)
	}
	return &lead, nil
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func (s *pgLeadStore) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	if tenantID == "" || id == "" || status == "" {
		return fmt.Errorf("tenantID, id and status are required: %v", t.ErrInvalidInput)
	}

	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET status = $1 WHERE id = $2 AND tenant_id = $3
	`, leadsTable), status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update lead status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("lead not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── AddNote ───────────────────────────────────────────────────────────────────

func (s *pgLeadStore) AddNote(ctx context.Context, tenantID, id, note string) error {
	if tenantID == "" || id == "" || note == "" {
		return fmt.Errorf("tenantID, id and note are required: %v", t.ErrInvalidInput)
	}

	prefix := fmt.Sprintf("[%s] ", time.Now().UTC().Format("2006-01-02 15:04"))
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET notes = CASE WHEN notes = '' THEN $1 ELSE notes || E'\n' || $1 END
		WHERE id = $2 AND tenant_id = $3
	`, leadsTable), prefix+note, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to add note to lead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("lead not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── ListByTenant ──────────────────────────────────────────────────────────────

func (s *pgLeadStore) ListByTenant(ctx context.Context, filters *LeadFilters) (*t.PaginatedResult[Lead], error) {
	if filters == nil {
		filters = &LeadFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildLeadFilters(filters)

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
		FROM %s
		%s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, leadsTable, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads: %w", err)
	}
	defer rows.Close()

	leads := []Lead{}
	for rows.Next() {
		var l Lead
		if err := scanLead(rows, &l); err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	countWhere, countArgs := buildLeadFilters(filters)
	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, leadsTable, countWhere), countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count leads: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &t.PaginatedResult[Lead]{
		Data:       leads,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ── ListByPropertyID ──────────────────────────────────────────────────────────

func (s *pgLeadStore) ListByPropertyID(ctx context.Context, tenantID, propertyID string) ([]Lead, error) {
	if tenantID == "" || propertyID == "" {
		return nil, fmt.Errorf("tenantID and propertyID are required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, name, email, phone, property_id, status, score, notes, source, created_at, updated_at
		FROM %s WHERE property_id = $1 AND tenant_id = $2
		ORDER BY updated_at DESC
	`, leadsTable), propertyID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads by property: %w", err)
	}
	defer rows.Close()

	var leads []Lead
	for rows.Next() {
		var l Lead
		if err := scanLead(rows, &l); err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, l)
	}
	return leads, rows.Err()
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgLeadStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Lead](), "pg_lead_store", nil
}

func (s *pgLeadStore) GetName() string { return "pg_lead_store" }

func (s *pgLeadStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgLeadStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanLead(row pgx.Row, l *Lead) error {
	return row.Scan(
		&l.ID, &l.TenantID, &l.Name, &l.Email, &l.Phone,
		&l.PropertyID, &l.Status, &l.Score, &l.Notes, &l.Source,
		&l.CreatedAt, &l.UpdatedAt,
	)
}

func buildLeadFilters(filters *LeadFilters) (string, []any) {
	where := []string{}
	args := []any{}

	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, filters.TenantID)
	}
	if filters.PropertyID != nil && *filters.PropertyID != "" {
		where = append(where, fmt.Sprintf("property_id = $%d", len(args)+1))
		args = append(args, *filters.PropertyID)
	}
	if filters.Status != nil && *filters.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filters.Status)
	}
	if filters.Phone != nil && *filters.Phone != "" {
		where = append(where, fmt.Sprintf("phone = $%d", len(args)+1))
		args = append(args, *filters.Phone)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
