package invitestore

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

// pgInviteStore implementa InviteStore usando PGExecutor.
type pgInviteStore struct {
	exec       execution.PGExecutor
	defaultTTL time.Duration
}

// NewPGInviteStore cria uma instância de InviteStore para Postgres.
func NewPGInviteStore(exec execution.PGExecutor) InviteStore {
	return &pgInviteStore{
		exec:       exec,
		defaultTTL: 7 * 24 * time.Hour, // 7 dias padrão
	}
}

// Create insere um novo convite.
func (s *pgInviteStore) Create(ctx context.Context, input *CreateInvitationInput) (*Invitation, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	table := tableName(input.Type)
	expiresAt := s.computeExpiry(input.ExpiresAt)

	const qTpl = `
		INSERT INTO %s (
			name, email, role, token, tenant_id, team_id, invited_by, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8)
		RETURNING id, name, email, role, token, tenant_id, team_id, invited_by,
		          status, expires_at, accepted_at, created_at, updated_at
	`

	query := fmt.Sprintf(qTpl, table)
	row := s.exec.QueryRow(ctx, query,
		input.Name,
		strings.ToLower(strings.TrimSpace(input.Email)),
		input.Role,
		input.Token,
		input.TenantID,
		input.TeamID,
		input.InvitedBy,
		expiresAt,
	)

	inv, err := scanInvitation(row, input.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %v", err)
	}
	return inv, nil
}

// GetByID busca convite por ID e tipo.
func (s *pgInviteStore) GetByID(ctx context.Context, id string, invType InvitationType) (*Invitation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	table := tableName(invType)
	query := fmt.Sprintf(`
		SELECT id, name, email, role, token, tenant_id, team_id, invited_by,
		       status, expires_at, accepted_at, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, table)

	row := s.exec.QueryRow(ctx, query, id)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invitation by id: %v", err)
	}
	return inv, nil
}

// GetByToken busca convite por token em ambas as tabelas.
func (s *pgInviteStore) GetByToken(ctx context.Context, token string) (*Invitation, error) {
	if strings.TrimSpace(token) == "" {
		return nil, nil
	}

	// Tenta partner_invitation primeiro
	if inv, err := s.findByToken(ctx, "partner_invitation", TypePartner, token); err == nil {
		return inv, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Depois internal_invitation
	inv, err := s.findByToken(ctx, "internal_invitation", TypeInternal, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

// Update atualiza campos do convite.
func (s *pgInviteStore) Update(ctx context.Context, input *UpdateInvitationInput) (*Invitation, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("update input with ID is required: %v", t.ErrInvalidInput)
	}

	table := tableName(input.Type)
	updates := []string{}
	args := []any{}
	idx := 1

	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", idx))
		args = append(args, *input.Status)
		idx++

		// Se status = accepted e AcceptedAt não foi informado, seta automaticamente
		if *input.Status == StatusAccepted && input.AcceptedAt == nil {
			now := time.Now().UTC()
			input.AcceptedAt = &now
		}
	}

	if input.AcceptedAt != nil {
		updates = append(updates, fmt.Sprintf("accepted_at = $%d", idx))
		args = append(args, input.AcceptedAt.UTC())
		idx++
	}

	if input.ExpiresAt != nil {
		updates = append(updates, fmt.Sprintf("expires_at = $%d", idx))
		args = append(args, input.ExpiresAt.UTC())
		idx++
	}

	if len(updates) == 0 {
		// Nenhum campo para atualizar, retorna convite atual
		return s.GetByID(ctx, input.ID, input.Type)
	}

	// Sempre atualiza updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now().UTC())
	idx++

	// Adiciona ID como último argumento
	args = append(args, input.ID)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id = $%d
		RETURNING id, name, email, role, token, tenant_id, team_id, invited_by,
		          status, expires_at, accepted_at, created_at, updated_at
	`, table, strings.Join(updates, ", "), idx)

	row := s.exec.QueryRow(ctx, query, args...)
	inv, err := scanInvitation(row, input.Type)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invitation not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update invitation: %v", err)
	}
	return inv, nil
}

// Accept marca o convite como aceito usando transação.
func (s *pgInviteStore) Accept(ctx context.Context, token string) (*Invitation, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("token is required: %v", t.ErrInvalidInput)
	}

	// Tenta accept em partner_invitation
	if inv, err := s.acceptInTable(ctx, "partner_invitation", TypePartner, token); err == nil {
		return inv, nil
	} else if !errors.Is(err, pgx.ErrNoRows) && !errors.Is(err, t.ErrNotFound) {
		return nil, err
	}

	// Tenta accept em internal_invitation
	return s.acceptInTable(ctx, "internal_invitation", TypeInternal, token)
}

// Revoke altera o status para revoked.
func (s *pgInviteStore) Revoke(ctx context.Context, id string, invType InvitationType) error {
	status := StatusRevoked
	_, err := s.Update(ctx, &UpdateInvitationInput{
		ID:     id,
		Type:   invType,
		Status: &status,
	})
	return err
}

// Delete remove o convite (hard delete).
func (s *pgInviteStore) Delete(ctx context.Context, id string, invType InvitationType) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %v", t.ErrInvalidInput)
	}

	table := tableName(invType)
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, table)

	tag, err := s.exec.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invitation not found: %v", t.ErrNotFound)
	}
	return nil
}

// List retorna convites paginados com filtros.
func (s *pgInviteStore) List(ctx context.Context, filters *InvitationFilters) (*t.PaginatedResult[Invitation], error) {
	if filters == nil || filters.Type == nil {
		return nil, fmt.Errorf("filters with Type are required: %v", t.ErrInvalidInput)
	}

	table := tableName(*filters.Type)
	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := calcOffset(page, limit)

	where, args := buildInviteFilters(filters)

	query := fmt.Sprintf(`
		SELECT id, name, email, role, token, tenant_id, team_id, invited_by,
		       status, expires_at, accepted_at, created_at, updated_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, table, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %v", err)
	}
	defer rows.Close()

	invitations := []Invitation{}
	for rows.Next() {
		inv, err := scanInvitation(rows, *filters.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %v", err)
		}
		invitations = append(invitations, *inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	// Count total
	total, err := s.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &t.PaginatedResult[Invitation]{
		Data:       invitations,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calcTotalPages(total, limit),
	}, nil
}

// Count retorna total de convites com filtros.
func (s *pgInviteStore) Count(ctx context.Context, filters *InvitationFilters) (int64, error) {
	if filters == nil || filters.Type == nil {
		return 0, fmt.Errorf("filters with Type are required: %v", t.ErrInvalidInput)
	}

	table := tableName(*filters.Type)
	where, args := buildInviteFilters(filters)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table, where)

	var total int64
	err := s.exec.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count invitations: %v", err)
	}
	return total, nil
}

func (s *pgInviteStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Invitation](), "pg_invite_store", nil
}

func (s *pgInviteStore) GetName() string {
	return "pg_invite_store"
}

func (s *pgInviteStore) Validate() error {
	if s.exec == nil {
		return fmt.Errorf("executor is nil")
	}
	if s.defaultTTL <= 0 {
		return fmt.Errorf("defaultTTL must be positive")
	}
	return nil
}

func (s *pgInviteStore) Close() error {
	// PGExecutor é gerenciado pelo pool, não precisa cleanup aqui
	return nil
}

// Helpers ----------------------------------------------------------------

func (s *pgInviteStore) computeExpiry(explicit *time.Time) time.Time {
	if explicit != nil && !explicit.IsZero() {
		return explicit.UTC()
	}
	return time.Now().UTC().Add(s.defaultTTL)
}

func (s *pgInviteStore) findByToken(ctx context.Context, table string, invType InvitationType, token string) (*Invitation, error) {
	query := fmt.Sprintf(`
		SELECT id, name, email, role, token, tenant_id, team_id, invited_by,
		       status, expires_at, accepted_at, created_at, updated_at
		FROM %s
		WHERE token = $1
	`, table)

	row := s.exec.QueryRow(ctx, query, token)
	return scanInvitation(row, invType)
}

func (s *pgInviteStore) acceptInTable(ctx context.Context, table string, invType InvitationType, token string) (*Invitation, error) {
	tx, err := s.exec.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// SELECT FOR UPDATE
	query := fmt.Sprintf(`
		SELECT id, name, email, role, token, tenant_id, team_id, invited_by,
		       status, expires_at, accepted_at, created_at, updated_at
		FROM %s
		WHERE token = $1
		FOR UPDATE
	`, table)

	row := tx.QueryRow(ctx, query, token)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invitation not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to lock invitation: %v", err)
	}

	// Validações
	if inv.Status != StatusPending {
		return nil, fmt.Errorf("invitation is not pending (status=%s)", inv.Status)
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("invitation expired")
	}

	// Update status
	now := time.Now().UTC()
	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, accepted_at = $2, updated_at = $2
		WHERE id = $3
	`, table)

	if _, err := tx.Exec(ctx, updateQuery, StatusAccepted, now, inv.ID); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Atualiza objeto em memória
	inv.Status = StatusAccepted
	inv.AcceptedAt = &now
	inv.UpdatedAt = &now

	return inv, nil
}

func scanInvitation(row pgx.Row, invType InvitationType) (*Invitation, error) {
	var inv Invitation
	err := row.Scan(
		&inv.ID,
		&inv.Name,
		&inv.Email,
		&inv.Role,
		&inv.Token,
		&inv.TenantID,
		&inv.TeamID,
		&inv.InvitedBy,
		&inv.Status,
		&inv.ExpiresAt,
		&inv.AcceptedAt,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	inv.Type = invType
	return &inv, nil
}

func tableName(t InvitationType) string {
	if t == TypePartner {
		return "partner_invitation"
	}
	return "internal_invitation"
}

func validateCreateInput(input *CreateInvitationInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Token) == "" {
		return fmt.Errorf("token is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Role) == "" {
		return fmt.Errorf("role is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" {
		return fmt.Errorf("tenant_id is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.InvitedBy) == "" {
		return fmt.Errorf("invited_by is required: %v", t.ErrInvalidInput)
	}
	if input.Type != TypePartner && input.Type != TypeInternal {
		return fmt.Errorf("invalid type: %v", t.ErrInvalidInput)
	}
	return nil
}

func buildInviteFilters(filters *InvitationFilters) (string, []any) {
	if filters == nil {
		return "", nil
	}

	where := []string{}
	args := []any{}

	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		where = append(where, fmt.Sprintf("LOWER(email) = $%d", len(args)+1))
		args = append(args, strings.ToLower(strings.TrimSpace(*filters.Email)))
	}

	if filters.TenantID != nil && strings.TrimSpace(*filters.TenantID) != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, *filters.TenantID)
	}

	if filters.TeamID != nil && strings.TrimSpace(*filters.TeamID) != "" {
		where = append(where, fmt.Sprintf("team_id = $%d", len(args)+1))
		args = append(args, *filters.TeamID)
	}

	if filters.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filters.Status)
	}

	if filters.InvitedBy != nil && strings.TrimSpace(*filters.InvitedBy) != "" {
		where = append(where, fmt.Sprintf("invited_by = $%d", len(args)+1))
		args = append(args, *filters.InvitedBy)
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

func calcOffset(page, limit int) int {
	return (page - 1) * limit
}

func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}
	if pages == 0 && total > 0 {
		return 1
	}
	return pages
}
