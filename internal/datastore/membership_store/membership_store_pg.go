package membershipstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/faelmori/domus-pulse/internal/execution"
	"github.com/jackc/pgx/v5"
	pglib "github.com/jackc/pgx/v5"
)

type pgMembershipStore struct {
	exec execution.PGExecutor
}

// NewPGMembershipStore cria uma instância de MembershipStore para Postgres.
func NewPGMembershipStore(exec execution.PGExecutor) MembershipStore {
	return &pgMembershipStore{exec: exec}
}

func (s *pgMembershipStore) ListByTenantWithPermissions(ctx context.Context, tenantID string) ([]MemberWithPermissions, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenantID é obrigatório")
	}

	const q = `
		SELECT
			u.id,
			u.email,
			COALESCE(u.name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.status, ''),
			tm.tenant_id,
			tm.role_id,
			COALESCE(r.code, ''),
			COALESCE(r.display_name, ''),
			tm.is_active,
			tm.created_at,
			COALESCE(
				array_agg(DISTINCT p.code ORDER BY p.code)
					FILTER (WHERE rp.value = true AND p.code IS NOT NULL),
				ARRAY[]::text[]
			) AS permissions
		FROM tenant_membership tm
		JOIN "user" u ON u.id = tm.user_id
		JOIN role r ON r.id = tm.role_id
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		LEFT JOIN permission p ON p.id = rp.permission_id
		WHERE tm.tenant_id = $1
		GROUP BY
			u.id, u.email, u.name, u.last_name, u.status,
			tm.tenant_id, tm.role_id, r.code, r.display_name, tm.is_active, tm.created_at
		ORDER BY tm.created_at ASC, u.email ASC`

	rows, err := s.exec.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("membership.ListByTenantWithPermissions(%s): %w", tenantID, err)
	}
	defer rows.Close()

	var result []MemberWithPermissions
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf("membership scan: %w", err)
		}
		result = append(result, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membership.ListByTenantWithPermissions rows: %w", err)
	}
	return result, nil
}

func (s *pgMembershipStore) Upsert(ctx context.Context, input *UpsertMembershipInput) error {
	if input == nil || strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.TenantID) == "" || strings.TrimSpace(input.RoleID) == "" {
		return fmt.Errorf("UpsertMembershipInput com UserID, TenantID e RoleID é obrigatório")
	}

	const q = `
		INSERT INTO tenant_membership (user_id, tenant_id, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET role_id = EXCLUDED.role_id, is_active = EXCLUDED.is_active, updated_at = now()
	`
	_, err := s.exec.Exec(ctx, q,
		input.UserID,
		input.TenantID,
		input.RoleID,
		input.IsActive,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("membership.Upsert user=%s tenant=%s: %w", input.UserID, input.TenantID, err)
	}
	return nil
}

func (s *pgMembershipStore) UpdateRole(ctx context.Context, input *UpdateRoleInput) error {
	if input == nil || strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.TenantID) == "" || strings.TrimSpace(input.RoleID) == "" {
		return fmt.Errorf("UpdateRoleInput com UserID, TenantID e RoleID é obrigatório")
	}

	const q = `
		UPDATE tenant_membership
		SET role_id = $1, updated_at = now()
		WHERE user_id = $2 AND tenant_id = $3
	`
	tag, err := s.exec.Exec(ctx, q, input.RoleID, input.UserID, input.TenantID)
	if err != nil {
		return fmt.Errorf("membership.UpdateRole user=%s tenant=%s: %w", input.UserID, input.TenantID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("membership not found for user %s in tenant %s", input.UserID, input.TenantID)
	}
	return nil
}

func (s *pgMembershipStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[MemberWithPermissions](), "pg_membership_store", nil
}
func (s *pgMembershipStore) GetName() string { return "pg_membership_store" }
func (s *pgMembershipStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgMembershipStore) Close() error { return nil }

func scanMember(row pglib.Row) (*MemberWithPermissions, error) {
	var m MemberWithPermissions
	err := row.Scan(
		&m.ID, &m.Email, &m.Name, &m.LastName, &m.Status,
		&m.TenantID, &m.RoleID, &m.RoleCode, &m.RoleName,
		&m.IsActive, &m.CreatedAt, &m.Permissions,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("scan MemberWithPermissions: %w", err)
	}
	return &m, nil
}
