package pendingaccessstore

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

// pgPendingAccessStore implementa PendingAccessStore usando PGExecutor.
type pgPendingAccessStore struct {
	exec execution.PGExecutor
}

// NewPGPendingAccessStore cria uma instância de PendingAccessStore para Postgres.
func NewPGPendingAccessStore(exec execution.PGExecutor) PendingAccessStore {
	return &pgPendingAccessStore{exec: exec}
}

// Create insere uma nova solicitação ou atualiza se já existir.
func (s *pgPendingAccessStore) Create(ctx context.Context, input *CreatePendingAccessRequestInput) (*PendingAccessRequest, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Provider) == "" {
		return nil, fmt.Errorf("email and provider are required: %v", t.ErrInvalidInput)
	}

	status := StatusPending
	if input.Status != nil && strings.TrimSpace(string(*input.Status)) != "" {
		status = *input.Status
	}

	query := `
		INSERT INTO pending_access_requests (
			email, provider, provider_user_id, name, avatar_url, status,
			requester_ip, requester_user_agent, tenant_id, role_code, metadata,
			reviewed_by, reviewed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (email, provider, status)
		DO UPDATE SET
			provider_user_id = EXCLUDED.provider_user_id,
			name = EXCLUDED.name,
			avatar_url = EXCLUDED.avatar_url,
			requester_ip = EXCLUDED.requester_ip,
			requester_user_agent = EXCLUDED.requester_user_agent,
			tenant_id = EXCLUDED.tenant_id,
			role_code = EXCLUDED.role_code,
			metadata = EXCLUDED.metadata,
			reviewed_by = EXCLUDED.reviewed_by,
			reviewed_at = EXCLUDED.reviewed_at,
			updated_at = now()
		RETURNING id, email, provider, provider_user_id, name, avatar_url, status,
		          requester_ip, requester_user_agent, tenant_id, role_code, metadata,
		          reviewed_by, reviewed_at, created_at, updated_at
	`

	createdAt := time.Now().UTC()
	row := s.exec.QueryRow(ctx, query,
		strings.ToLower(strings.TrimSpace(input.Email)),
		strings.TrimSpace(input.Provider),
		input.ProviderUserID,
		input.Name,
		input.AvatarURL,
		status,
		input.RequesterIP,
		input.RequesterUserAgent,
		input.TenantID,
		input.RoleCode,
		normalizeJSON(input.Metadata),
		input.ReviewedBy,
		input.ReviewedAt,
		createdAt,
	)

	req, err := scanPendingAccess(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending access request: %v", err)
	}
	return req, nil
}

// GetByID busca solicitação por ID.
func (s *pgPendingAccessStore) GetByID(ctx context.Context, id string) (*PendingAccessRequest, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	const q = `
		SELECT id, email, provider, provider_user_id, name, avatar_url, status,
		       requester_ip, requester_user_agent, tenant_id, role_code, metadata,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM pending_access_requests
		WHERE id = $1
	`

	row := s.exec.QueryRow(ctx, q, id)
	req, err := scanPendingAccess(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pending access request by id: %v", err)
	}
	return req, nil
}

// Update atualiza status e campos de revisão.
func (s *pgPendingAccessStore) Update(ctx context.Context, input *UpdatePendingAccessRequestInput) (*PendingAccessRequest, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("update input with ID is required: %v", t.ErrInvalidInput)
	}

	updates := []string{}
	args := []any{}
	idx := 1

	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", idx))
		args = append(args, *input.Status)
		idx++
	}
	if input.ReviewedBy != nil {
		updates = append(updates, fmt.Sprintf("reviewed_by = $%d", idx))
		args = append(args, input.ReviewedBy)
		idx++
	}
	if input.ReviewedAt != nil {
		updates = append(updates, fmt.Sprintf("reviewed_at = $%d", idx))
		args = append(args, input.ReviewedAt)
		idx++
	}

	if len(updates) == 0 {
		return s.GetByID(ctx, input.ID)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now().UTC())
	idx++

	args = append(args, input.ID)

	query := fmt.Sprintf(`
		UPDATE pending_access_requests
		SET %s
		WHERE id = $%d
		RETURNING id, email, provider, provider_user_id, name, avatar_url, status,
		          requester_ip, requester_user_agent, tenant_id, role_code, metadata,
		          reviewed_by, reviewed_at, created_at, updated_at
	`, strings.Join(updates, ", "), idx)

	row := s.exec.QueryRow(ctx, query, args...)
	req, err := scanPendingAccess(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("pending access request not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update pending access request: %v", err)
	}
	return req, nil
}

// List retorna solicitações paginadas.
func (s *pgPendingAccessStore) List(ctx context.Context, filters *PendingAccessFilters) (*t.PaginatedResult[PendingAccessRequest], error) {
	if filters == nil {
		filters = &PendingAccessFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := calcOffset(page, limit)

	where, args := buildFilters(filters)
	query := fmt.Sprintf(`
		SELECT id, email, provider, provider_user_id, name, avatar_url, status,
		       requester_ip, requester_user_agent, tenant_id, role_code, metadata,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM pending_access_requests
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending access requests: %v", err)
	}
	defer rows.Close()

	requests := []PendingAccessRequest{}
	for rows.Next() {
		req, err := scanPendingAccess(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending access request: %v", err)
		}
		requests = append(requests, *req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	total, err := s.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &t.PaginatedResult[PendingAccessRequest]{
		Data:       requests,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calcTotalPages(total, limit),
	}, nil
}

// Count retorna total de solicitações com filtros.
func (s *pgPendingAccessStore) Count(ctx context.Context, filters *PendingAccessFilters) (int64, error) {
	if filters == nil {
		filters = &PendingAccessFilters{}
	}

	where, args := buildFilters(filters)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM pending_access_requests %s`, where)

	var total int64
	err := s.exec.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending access requests: %v", err)
	}
	return total, nil
}

func (s *pgPendingAccessStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[PendingAccessRequest](), "pg_pending_access_store", nil
}

func (s *pgPendingAccessStore) GetName() string {
	return "pg_pending_access_store"
}

func (s *pgPendingAccessStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgPendingAccessStore) Close() error {
	return nil
}

// Helpers ----------------------------------------------------------------

func scanPendingAccess(row pgx.Row) (*PendingAccessRequest, error) {
	var req PendingAccessRequest
	var status string
	var metadata []byte
	err := row.Scan(
		&req.ID,
		&req.Email,
		&req.Provider,
		&req.ProviderUserID,
		&req.Name,
		&req.AvatarURL,
		&status,
		&req.RequesterIP,
		&req.RequesterUserAgent,
		&req.TenantID,
		&req.RoleCode,
		&metadata,
		&req.ReviewedBy,
		&req.ReviewedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	req.Status = PendingAccessStatus(status)
	req.Metadata = metadata
	return &req, nil
}

func buildFilters(filters *PendingAccessFilters) (string, []any) {
	where := []string{}
	args := []any{}

	add := func(cond string, val any) {
		where = append(where, cond)
		args = append(args, val)
	}

	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		add(fmt.Sprintf("email = $%d", len(args)+1), strings.ToLower(strings.TrimSpace(*filters.Email)))
	}
	if filters.Provider != nil && strings.TrimSpace(*filters.Provider) != "" {
		add(fmt.Sprintf("provider = $%d", len(args)+1), strings.TrimSpace(*filters.Provider))
	}
	if filters.Status != nil && strings.TrimSpace(string(*filters.Status)) != "" {
		add(fmt.Sprintf("status = $%d", len(args)+1), *filters.Status)
	}
	if filters.TenantID != nil && strings.TrimSpace(*filters.TenantID) != "" {
		add(fmt.Sprintf("tenant_id = $%d", len(args)+1), strings.TrimSpace(*filters.TenantID))
	}
	if filters.RoleCode != nil && strings.TrimSpace(*filters.RoleCode) != "" {
		add(fmt.Sprintf("role_code = $%d", len(args)+1), strings.TrimSpace(*filters.RoleCode))
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizeLimit(limit, def, max int) int {
	if limit <= 0 {
		return def
	}
	if limit > max {
		return max
	}
	return limit
}

func calcOffset(page, limit int) int {
	return (page - 1) * limit
}

func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 1
	}
	pages := int(total) / limit
	if int64(pages*limit) < total {
		pages++
	}
	if pages <= 0 {
		pages = 1
	}
	return pages
}

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
