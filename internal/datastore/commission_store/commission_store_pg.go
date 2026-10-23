package commissionstore

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

const commissionsTable = "limalar.commissions"

type pgCommissionStore struct {
	exec execution.PGExecutor
}

// NewPGCommissionStore cria uma instância de CommissionStore para Postgres.
func NewPGCommissionStore(exec execution.PGExecutor) CommissionStore {
	return &pgCommissionStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgCommissionStore) Create(ctx context.Context, input *CreateCommissionInput) (*Commission, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.TenantID == "" || input.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required: %v", t.ErrInvalidInput)
	}

	status := input.Status
	if status == "" {
		status = "pending"
	}

	var c Commission
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, user_id, property_id, gross_value, pct, status, notes, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, user_id, property_id, gross_value, pct, net_value,
		          status, notes, paid_at, created_at, updated_at
	`, commissionsTable),
		input.TenantID, input.UserID, input.PropertyID, input.GrossValue,
		input.Pct, status, input.Notes, input.PaidAt,
	).Scan(
		&c.ID, &c.TenantID, &c.UserID, &c.PropertyID,
		&c.GrossValue, &c.Pct, &c.NetValue,
		&c.Status, &c.Notes, &c.PaidAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create commission: %w", err)
	}
	return &c, nil
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func (s *pgCommissionStore) GetByID(ctx context.Context, tenantID, id string) (*Commission, error) {
	if tenantID == "" || id == "" {
		return nil, fmt.Errorf("tenantID and id are required: %v", t.ErrInvalidInput)
	}

	var c Commission
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, user_id, property_id, gross_value, pct, net_value,
		       status, notes, paid_at, created_at, updated_at
		FROM %s WHERE id = $1 AND tenant_id = $2
	`, commissionsTable), id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.UserID, &c.PropertyID,
		&c.GrossValue, &c.Pct, &c.NetValue,
		&c.Status, &c.Notes, &c.PaidAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("commission not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commission: %w", err)
	}
	return &c, nil
}

// ── List ──────────────────────────────────────────────────────────────────────

func (s *pgCommissionStore) List(ctx context.Context, filters *CommissionFilters) (*t.PaginatedResult[Commission], error) {
	if filters == nil {
		filters = &CommissionFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildCommissionFilters(filters)

	q := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, property_id, gross_value, pct, net_value,
		       status, notes, paid_at, created_at, updated_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, commissionsTable, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list commissions: %w", err)
	}
	defer rows.Close()

	commissions := []Commission{}
	for rows.Next() {
		var c Commission
		if err := scanCommission(rows, &c); err != nil {
			return nil, fmt.Errorf("failed to scan commission: %w", err)
		}
		commissions = append(commissions, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	countWhere, countArgs := buildCommissionFilters(filters)
	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, commissionsTable, countWhere), countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count commissions: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &t.PaginatedResult[Commission]{
		Data:       commissions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func (s *pgCommissionStore) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	if tenantID == "" || id == "" || status == "" {
		return fmt.Errorf("tenantID, id and status are required: %v", t.ErrInvalidInput)
	}

	var q string
	var args []any
	if status == "received" {
		q = fmt.Sprintf(`UPDATE %s SET status = $1, paid_at = NOW() WHERE id = $2 AND tenant_id = $3`, commissionsTable)
	} else {
		q = fmt.Sprintf(`UPDATE %s SET status = $1 WHERE id = $2 AND tenant_id = $3`, commissionsTable)
	}
	args = []any{status, id, tenantID}

	tag, err := s.exec.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to update commission status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("commission not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── SumByPeriod ───────────────────────────────────────────────────────────────

func (s *pgCommissionStore) SumByPeriod(ctx context.Context, tenantID, userID string, from, to time.Time) (int64, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenantID is required: %v", t.ErrInvalidInput)
	}

	args := []any{tenantID, from, to}
	userClause := ""
	if userID != "" {
		args = append(args, userID)
		userClause = fmt.Sprintf("AND user_id = $%d", len(args))
	}

	var sum int64
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT COALESCE(SUM(net_value), 0)
		FROM %s
		WHERE tenant_id = $1
		  AND status = 'received'
		  AND paid_at BETWEEN $2 AND $3
		  %s
	`, commissionsTable, userClause), args...).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("failed to sum commissions: %w", err)
	}
	return sum, nil
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgCommissionStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Commission](), "pg_commission_store", nil
}

func (s *pgCommissionStore) GetName() string { return "pg_commission_store" }

func (s *pgCommissionStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgCommissionStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanCommission(row pgx.Row, c *Commission) error {
	return row.Scan(
		&c.ID, &c.TenantID, &c.UserID, &c.PropertyID,
		&c.GrossValue, &c.Pct, &c.NetValue,
		&c.Status, &c.Notes, &c.PaidAt, &c.CreatedAt, &c.UpdatedAt,
	)
}

func buildCommissionFilters(filters *CommissionFilters) (string, []any) {
	where := []string{}
	args := []any{}

	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, filters.TenantID)
	}
	if filters.UserID != nil && *filters.UserID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, *filters.UserID)
	}
	if filters.PropertyID != nil && *filters.PropertyID != "" {
		where = append(where, fmt.Sprintf("property_id = $%d", len(args)+1))
		args = append(args, *filters.PropertyID)
	}
	if filters.Status != nil && *filters.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filters.Status)
	}
	if filters.From != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, *filters.From)
	}
	if filters.To != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", len(args)+1))
		args = append(args, *filters.To)
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
