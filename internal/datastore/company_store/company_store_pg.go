package companystore

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

// pgCompanyStore implementa CompanyStore usando PGExecutor.
type pgCompanyStore struct {
	exec execution.PGExecutor
}

// NewPGCompanyStore cria uma instância de CompanyStore para Postgres.
func NewPGCompanyStore(exec execution.PGExecutor) CompanyStore {
	return &pgCompanyStore{exec: exec}
}

// Create insere uma nova empresa.
func (s *pgCompanyStore) Create(ctx context.Context, input *CreateCompanyInput) (*Company, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("name is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Slug) == "" {
		return nil, fmt.Errorf("slug is required: %v", t.ErrInvalidInput)
	}

	const q = `
		INSERT INTO gnyx.companies (
			name, slug, is_trial, is_active, domain, phone, address, plan_expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, slug, created_at, updated_at, plan_expires_at, is_trial, is_active, domain, phone, address
	`

	now := time.Now().UTC()
	row := s.exec.QueryRow(ctx, q,
		input.Name,
		strings.ToLower(strings.TrimSpace(input.Slug)),
		input.IsTrial,
		input.IsActive,
		input.Domain,
		input.Phone,
		input.Address,
		input.PlanExpiresAt,
		now,
	)

	company, err := scanCompany(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %v", err)
	}
	return company, nil
}

// GetByID busca empresa por ID.
func (s *pgCompanyStore) GetByID(ctx context.Context, id string) (*Company, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	const q = `
		SELECT id, name, slug, created_at, updated_at, plan_expires_at, is_trial, is_active, domain, phone, address
		FROM gnyx.companies
		WHERE id = $1
	`

	row := s.exec.QueryRow(ctx, q, id)
	company, err := scanCompany(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get company by id: %v", err)
	}
	return company, nil
}

// GetBySlug busca empresa por slug (case-insensitive).
func (s *pgCompanyStore) GetBySlug(ctx context.Context, slug string) (*Company, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, nil
	}

	const q = `
		SELECT id, name, slug, created_at, updated_at, plan_expires_at, is_trial, is_active, domain, phone, address
		FROM gnyx.companies
		WHERE slug = $1
	`

	row := s.exec.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(slug)))
	company, err := scanCompany(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get company by slug: %v", err)
	}
	return company, nil
}

// Update atualiza campos da empresa conforme input.
func (s *pgCompanyStore) Update(ctx context.Context, input *UpdateCompanyInput) (*Company, error) {
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
	if input.Slug != nil {
		updates = append(updates, fmt.Sprintf("slug = $%d", idx))
		args = append(args, strings.ToLower(strings.TrimSpace(*input.Slug)))
		idx++
	}
	if input.IsTrial != nil {
		updates = append(updates, fmt.Sprintf("is_trial = $%d", idx))
		args = append(args, input.IsTrial)
		idx++
	}
	if input.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, input.IsActive)
		idx++
	}
	if input.Domain != nil {
		updates = append(updates, fmt.Sprintf("domain = $%d", idx))
		args = append(args, input.Domain)
		idx++
	}
	if input.Phone != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", idx))
		args = append(args, input.Phone)
		idx++
	}
	if input.Address != nil {
		updates = append(updates, fmt.Sprintf("address = $%d", idx))
		args = append(args, input.Address)
		idx++
	}
	if input.PlanExpiresAt != nil {
		updates = append(updates, fmt.Sprintf("plan_expires_at = $%d", idx))
		args = append(args, input.PlanExpiresAt)
		idx++
	}

	if len(updates) == 0 {
		// Nenhum campo para atualizar, retorna empresa atual
		return s.GetByID(ctx, input.ID)
	}

	// Sempre atualiza updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now().UTC())
	idx++

	// Adiciona ID como último argumento
	args = append(args, input.ID)

	query := fmt.Sprintf(`
		UPDATE gnyx.companies
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, created_at, updated_at, plan_expires_at, is_trial, is_active, domain, phone, address
	`, strings.Join(updates, ", "), idx)

	row := s.exec.QueryRow(ctx, query, args...)
	company, err := scanCompany(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("company not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update company: %v", err)
	}
	return company, nil
}

// Delete remove uma empresa (hard delete).
func (s *pgCompanyStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %v", t.ErrInvalidInput)
	}

	const q = `DELETE FROM gnyx.companies WHERE id = $1`

	tag, err := s.exec.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("failed to delete company: %v", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("company not found: %v", t.ErrNotFound)
	}

	return nil
}

// List retorna empresas paginadas com filtros opcionais.
func (s *pgCompanyStore) List(ctx context.Context, filters *CompanyFilters) (*t.PaginatedResult[Company], error) {
	page := 1
	limit := 20

	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
		}
	}

	offset := (page - 1) * limit

	// Construir WHERE clause
	whereClauses := []string{}
	args := []any{}
	idx := 1

	if filters != nil {
		if filters.Name != nil && strings.TrimSpace(*filters.Name) != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", idx))
			args = append(args, "%"+*filters.Name+"%")
			idx++
		}
		if filters.Slug != nil && strings.TrimSpace(*filters.Slug) != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("slug = $%d", idx))
			args = append(args, strings.ToLower(strings.TrimSpace(*filters.Slug)))
			idx++
		}
		if filters.IsActive != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", idx))
			args = append(args, *filters.IsActive)
			idx++
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM gnyx.companies %s`, whereClause)
	var total int64
	if err := s.exec.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count companies: %v", err)
	}

	// Query paginada
	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, created_at, updated_at, plan_expires_at, is_trial, is_active, domain, phone, address
		FROM gnyx.companies
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, idx, idx+1)

	rows, err := s.exec.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list companies: %v", err)
	}
	defer rows.Close()

	companies := []Company{}
	for rows.Next() {
		company, err := scanCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company: %v", err)
		}
		companies = append(companies, *company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating companies: %v", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &t.PaginatedResult[Company]{
		Data:       companies,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Count retorna o total de empresas (com filtros opcionais).
func (s *pgCompanyStore) Count(ctx context.Context, filters *CompanyFilters) (int64, error) {
	whereClauses := []string{}
	args := []any{}
	idx := 1

	if filters != nil {
		if filters.Name != nil && strings.TrimSpace(*filters.Name) != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", idx))
			args = append(args, "%"+*filters.Name+"%")
			idx++
		}
		if filters.Slug != nil && strings.TrimSpace(*filters.Slug) != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("slug = $%d", idx))
			args = append(args, strings.ToLower(strings.TrimSpace(*filters.Slug)))
			idx++
		}
		if filters.IsActive != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", idx))
			args = append(args, *filters.IsActive)
			idx++
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM gnyx.companies %s`, whereClause)
	var total int64
	if err := s.exec.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count companies: %v", err)
	}

	return total, nil
}

// scanCompany é um helper para fazer scan de Company de pgx.Row ou pgx.Rows.
func scanCompany(scanner interface {
	Scan(dest ...any) error
}) (*Company, error) {
	var c Company
	err := scanner.Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.PlanExpiresAt,
		&c.IsTrial,
		&c.IsActive,
		&c.Domain,
		&c.Phone,
		&c.Address,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetType implementa StoreType.
func (s *pgCompanyStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Company](), "pg_company_store", nil
}

// GetName implementa StoreType.
func (s *pgCompanyStore) GetName() string {
	return "pg_company_store"
}

// Validate implementa StoreType.
func (s *pgCompanyStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

// Close implementa StoreType.
func (s *pgCompanyStore) Close() error {
	// PGExecutor é gerenciado pelo pool, não precisa cleanup aqui
	return nil
}

// Name implementa StoreType.
func (s *pgCompanyStore) Name() string {
	return "company"
}
