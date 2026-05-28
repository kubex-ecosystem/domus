package rolestore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
)

type pgRoleStore struct {
	exec execution.PGExecutor
}

// NewPGRoleStore cria uma instância de RoleStore para Postgres.
func NewPGRoleStore(exec execution.PGExecutor) RoleStore {
	return &pgRoleStore{exec: exec}
}

func (s *pgRoleStore) GetByCode(ctx context.Context, code string) (*Role, error) {
	if strings.TrimSpace(code) == "" {
		return nil, nil
	}
	const q = `
		SELECT r.id, r.code, r.display_name, COALESCE(r.description,''), COALESCE(r.is_system_role,false),
		       COUNT(rp.permission_id) FILTER (WHERE rp.value = true) AS permission_count
		FROM role r
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		WHERE r.code = $1
		GROUP BY r.id, r.code, r.display_name, r.description, r.is_system_role
	`
	row := s.exec.QueryRow(ctx, q, strings.TrimSpace(code))
	role, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("role.GetByCode(%s): %w", code, err)
	}
	return role, nil
}

func (s *pgRoleStore) GetByID(ctx context.Context, id string) (*Role, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT r.id, r.code, r.display_name, COALESCE(r.description,''), COALESCE(r.is_system_role,false),
		       COUNT(rp.permission_id) FILTER (WHERE rp.value = true) AS permission_count
		FROM role r
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		WHERE r.id = $1
		GROUP BY r.id, r.code, r.display_name, r.description, r.is_system_role
	`
	row := s.exec.QueryRow(ctx, q, strings.TrimSpace(id))
	role, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("role.GetByID(%s): %w", id, err)
	}
	return role, nil
}

func (s *pgRoleStore) List(ctx context.Context) ([]Role, error) {
	const q = `
		SELECT r.id, r.code, r.display_name, COALESCE(r.description,''), COALESCE(r.is_system_role,false),
		       COUNT(rp.permission_id) FILTER (WHERE rp.value = true) AS permission_count
		FROM role r
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		GROUP BY r.id, r.code, r.display_name, r.description, r.is_system_role
		ORDER BY r.code ASC
	`
	rows, err := s.exec.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("role.List: %w", err)
	}
	defer rows.Close()

	var result []Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("role.List scan: %w", err)
		}
		result = append(result, *role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("role.List rows: %w", err)
	}
	return result, nil
}

func (s *pgRoleStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Role](), "pg_role_store", nil
}
func (s *pgRoleStore) GetName() string { return "pg_role_store" }
func (s *pgRoleStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgRoleStore) Close() error { return nil }

func scanRole(row pgx.Row) (*Role, error) {
	var r Role
	err := row.Scan(&r.ID, &r.Code, &r.DisplayName, &r.Description, &r.IsSystemRole, &r.PermissionCount)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
