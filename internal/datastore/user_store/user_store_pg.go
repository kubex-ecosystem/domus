package userstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/model/gnyx/user"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// pgUserStore implementa UserStore usando PGExecutor.
type pgUserStore struct {
	exec       execution.PGExecutor
	tableOnce  sync.Once
	tableName  string
	tableCheck error
}

// NewPGUserStore cria uma instância de UserStore para Postgres.
func NewPGUserStore(exec execution.PGExecutor) UserStore {
	return &pgUserStore{exec: exec}
}

func (s *pgUserStore) resolveTableName(ctx context.Context) string {
	s.tableOnce.Do(func() {
		if ctx == nil {
			ctx = context.Background()
		}

		var name *string
		err := s.exec.QueryRow(ctx, `SELECT to_regclass('public."user"')`).Scan(&name)
		if err == nil && name != nil && *name != "" {
			s.tableName = `"user"`
			return
		}

		name = nil
		err = s.exec.QueryRow(ctx, `SELECT to_regclass('public.users')`).Scan(&name)
		if err == nil && name != nil && *name != "" {
			s.tableName = "users"
			return
		}

		// Default to legacy table name when discovery fails.
		s.tableName = `"user"`
		s.tableCheck = err
	})
	if s.tableName == "" {
		s.tableName = `"user"`
	}
	return s.tableName
}

// Create insere um novo usuário.
func (s *pgUserStore) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Email) == "" {
		return nil, fmt.Errorf("email is required: %v", t.ErrInvalidInput)
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`
		INSERT INTO %s (
			email, name, last_name, password_hash, phone, avatar_url,
			status, force_password_reset, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, name, last_name, password_hash, phone, avatar_url,
		          status, force_password_reset, last_login, created_at, updated_at
	`, tableName)

	now := time.Now().UTC()
	row := s.exec.QueryRow(ctx, q,
		strings.ToLower(strings.TrimSpace(input.Email)),
		input.Name,
		input.LastName,
		input.PasswordHash,
		input.Phone,
		input.AvatarURL,
		input.Status,
		input.ForcePasswordReset,
		now,
	)

	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return user, nil
}

// GetByID busca usuário por ID.
func (s *pgUserStore) GetByID(ctx context.Context, id string) (*User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`
		SELECT id, email, name, last_name, password_hash, phone, avatar_url,
		       status, force_password_reset, last_login, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, tableName)

	row := s.exec.QueryRow(ctx, q, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	return user, nil
}

// GetByEmail busca usuário por email (case-insensitive via CITEXT).
func (s *pgUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, nil
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`
		SELECT id, email, name, last_name, password_hash, phone, avatar_url,
		       status, force_password_reset, last_login, created_at, updated_at
		FROM %s
		WHERE email = $1
	`, tableName)

	row := s.exec.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(email)))
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}
	return user, nil
}

// Update atualiza campos do usuário conforme input.
func (s *pgUserStore) Update(ctx context.Context, input *UpdateUserInput) (*User, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("update input with ID is required: %v", t.ErrInvalidInput)
	}

	updates := []string{}
	args := []any{}
	idx := 1

	if input.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", idx))
		args = append(args, input.Name)
		idx++
	}
	if input.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", idx))
		args = append(args, input.LastName)
		idx++
	}
	if input.PasswordHash != nil {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", idx))
		args = append(args, input.PasswordHash)
		idx++
	}
	if input.Phone != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", idx))
		args = append(args, input.Phone)
		idx++
	}
	if input.AvatarURL != nil {
		updates = append(updates, fmt.Sprintf("avatar_url = $%d", idx))
		args = append(args, input.AvatarURL)
		idx++
	}
	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", idx))
		args = append(args, input.Status)
		idx++
	}
	if input.ForcePasswordReset != nil {
		updates = append(updates, fmt.Sprintf("force_password_reset = $%d", idx))
		args = append(args, *input.ForcePasswordReset)
		idx++
	}

	if len(updates) == 0 {
		// Nenhum campo para atualizar, retorna usuário atual
		return s.GetByID(ctx, input.ID)
	}

	// Sempre atualiza updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now().UTC())
	idx++

	// Adiciona ID como último argumento
	args = append(args, input.ID)

	tableName := s.resolveTableName(ctx)
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id = $%d
		RETURNING id, email, name, last_name, password_hash, phone, avatar_url,
		          status, force_password_reset, last_login, created_at, updated_at
	`, tableName, strings.Join(updates, ", "), idx)

	row := s.exec.QueryRow(ctx, query, args...)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return user, nil
}

// UpdatePassword atualiza apenas o password_hash.
func (s *pgUserStore) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(passwordHash) == "" {
		return fmt.Errorf("userID and passwordHash are required: %v", t.ErrInvalidInput)
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`
		UPDATE %s
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`, tableName)

	tag, err := s.exec.Exec(ctx, q, passwordHash, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %v", t.ErrNotFound)
	}
	return nil
}

// UpdateLastLogin atualiza o timestamp do último login.
func (s *pgUserStore) UpdateLastLogin(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("userID is required: %v", t.ErrInvalidInput)
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`
		UPDATE %s
		SET last_login = $1, updated_at = $1
		WHERE id = $2
	`, tableName)

	now := time.Now().UTC()
	tag, err := s.exec.Exec(ctx, q, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %v", t.ErrNotFound)
	}
	return nil
}

// Delete remove o usuário (hard delete).
func (s *pgUserStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %v", t.ErrInvalidInput)
	}

	tableName := s.resolveTableName(ctx)
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, tableName)

	tag, err := s.exec.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %v", t.ErrNotFound)
	}
	return nil
}

// List retorna usuários paginados com filtros.
func (s *pgUserStore) List(ctx context.Context, filters *UserFilters) (*t.PaginatedResult[User], error) {
	if filters == nil {
		filters = &UserFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := calcOffset(page, limit)

	where, args := buildUserFilters(filters)

	tableName := s.resolveTableName(ctx)
	query := fmt.Sprintf(`
		SELECT id, email, name, last_name, password_hash, phone, avatar_url,
		       status, force_password_reset, last_login, created_at, updated_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, tableName, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	// Count total
	total, err := s.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &t.PaginatedResult[User]{
		Data:       users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calcTotalPages(total, limit),
	}, nil
}

// Count retorna total de usuários com filtros.
func (s *pgUserStore) Count(ctx context.Context, filters *UserFilters) (int64, error) {
	if filters == nil {
		filters = &UserFilters{}
	}

	where, args := buildUserFilters(filters)
	tableName := s.resolveTableName(ctx)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableName, where)

	var total int64
	err := s.exec.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %v", err)
	}
	return total, nil
}

func (s *pgUserStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[user.User](), "pg_user_store", nil
}

func (s *pgUserStore) GetName() string {
	return "pg_user_store"
}

func (s *pgUserStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgUserStore) Close() error {
	// PGExecutor é gerenciado pelo pool, não precisa cleanup aqui
	return nil
}

// Helpers ----------------------------------------------------------------

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&u.LastName,
		&u.PasswordHash,
		&u.Phone,
		&u.AvatarURL,
		&u.Status,
		&u.ForcePasswordReset,
		&u.LastLogin,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func buildUserFilters(filters *UserFilters) (string, []any) {
	if filters == nil {
		return "", nil
	}

	where := []string{}
	args := []any{}

	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		where = append(where, fmt.Sprintf("email = $%d", len(args)+1))
		args = append(args, strings.ToLower(strings.TrimSpace(*filters.Email)))
	}

	if filters.Status != nil && strings.TrimSpace(*filters.Status) != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filters.Status)
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
