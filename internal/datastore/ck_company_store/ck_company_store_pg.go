package ckcompanystore

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

type pgCKCompanyStore struct {
	exec execution.PGExecutor
}

func NewPGCKCompanyStore(exec execution.PGExecutor) CKCompanyStore {
	return &pgCKCompanyStore{exec: exec}
}

func (s *pgCKCompanyStore) Create(ctx context.Context, input *CreateCKCompanyInput) (*CKCompany, error) {
	if input == nil || strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("name is required: %w", t.ErrInvalidInput)
	}

	activeStatus := input.ActiveStatus
	if activeStatus == "" {
		activeStatus = "active"
	}
	brandColor := input.BrandColor
	if brandColor == "" {
		brandColor = "#f97316"
	}

	const q = `
		INSERT INTO cozinha_pass.company (name, logo, brand_color, cuisine, active_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, logo, brand_color, cuisine, active_status, created_at
	`
	row := s.exec.QueryRow(ctx, q,
		input.Name, input.Logo, brandColor, input.Cuisine, activeStatus, time.Now().UTC(),
	)
	company, err := scanCKCompany(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create ck company: %w", err)
	}

	if len(input.Channels) > 0 {
		if err := s.SetChannels(ctx, company.ID, input.Channels); err != nil {
			return nil, err
		}
		company.Channels = input.Channels
	}
	return company, nil
}

func (s *pgCKCompanyStore) GetByID(ctx context.Context, id string) (*CKCompany, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, name, logo, brand_color, cuisine, active_status, created_at
		FROM cozinha_pass.company WHERE id = $1
	`
	row := s.exec.QueryRow(ctx, q, id)
	company, err := scanCKCompany(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck company: %w", err)
	}
	channels, err := s.GetChannels(ctx, id)
	if err != nil {
		return nil, err
	}
	company.Channels = channels
	return company, nil
}

func (s *pgCKCompanyStore) Update(ctx context.Context, input *UpdateCKCompanyInput) (*CKCompany, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Logo != nil {
		sets = append(sets, fmt.Sprintf("logo = $%d", idx))
		args = append(args, *input.Logo)
		idx++
	}
	if input.BrandColor != nil {
		sets = append(sets, fmt.Sprintf("brand_color = $%d", idx))
		args = append(args, *input.BrandColor)
		idx++
	}
	if input.Cuisine != nil {
		sets = append(sets, fmt.Sprintf("cuisine = $%d", idx))
		args = append(args, *input.Cuisine)
		idx++
	}
	if input.ActiveStatus != nil {
		sets = append(sets, fmt.Sprintf("active_status = $%d", idx))
		args = append(args, *input.ActiveStatus)
		idx++
	}

	var company *CKCompany
	if len(sets) > 0 {
		args = append(args, input.ID)
		query := fmt.Sprintf(`
			UPDATE cozinha_pass.company SET %s
			WHERE id = $%d
			RETURNING id, name, logo, brand_color, cuisine, active_status, created_at
		`, strings.Join(sets, ", "), idx)
		row := s.exec.QueryRow(ctx, query, args...)
		var err error
		company, err = scanCKCompany(row)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("ck company not found: %w", t.ErrNotFound)
			}
			return nil, fmt.Errorf("failed to update ck company: %w", err)
		}
	} else {
		var err error
		company, err = s.GetByID(ctx, input.ID)
		if err != nil {
			return nil, err
		}
	}

	if input.Channels != nil {
		if err := s.SetChannels(ctx, input.ID, input.Channels); err != nil {
			return nil, err
		}
		company.Channels = input.Channels
	} else {
		channels, err := s.GetChannels(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		company.Channels = channels
	}
	return company, nil
}

func (s *pgCKCompanyStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	tag, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.company WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ck company: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ck company not found: %w", t.ErrNotFound)
	}
	return nil
}

func (s *pgCKCompanyStore) List(ctx context.Context, filters *CKCompanyFilters) (*t.PaginatedResult[CKCompany], error) {
	page, limit := 1, 20
	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
		}
	}
	offset := (page - 1) * limit

	wheres := []string{}
	args := []any{}
	idx := 1

	if filters != nil {
		if filters.Name != nil && strings.TrimSpace(*filters.Name) != "" {
			wheres = append(wheres, fmt.Sprintf("name ILIKE $%d", idx))
			args = append(args, "%"+*filters.Name+"%")
			idx++
		}
		if filters.ActiveStatus != nil {
			wheres = append(wheres, fmt.Sprintf("active_status = $%d", idx))
			args = append(args, *filters.ActiveStatus)
			idx++
		}
	}

	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM cozinha_pass.company %s`, where), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ck companies: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, name, logo, brand_color, cuisine, active_status, created_at
		FROM cozinha_pass.company %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ck companies: %w", err)
	}
	defer rows.Close()

	companies := []CKCompany{}
	for rows.Next() {
		c, err := scanCKCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ck company: %w", err)
		}
		companies = append(companies, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ck companies: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[CKCompany]{
		Data: companies, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func (s *pgCKCompanyStore) SetChannels(ctx context.Context, companyID string, channels []string) error {
	if _, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.company_channel WHERE company_id = $1`, companyID); err != nil {
		return fmt.Errorf("failed to clear channels: %w", err)
	}
	for _, ch := range channels {
		if _, err := s.exec.Exec(ctx,
			`INSERT INTO cozinha_pass.company_channel (company_id, channel) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			companyID, ch,
		); err != nil {
			return fmt.Errorf("failed to insert channel %s: %w", ch, err)
		}
	}
	return nil
}

func (s *pgCKCompanyStore) GetChannels(ctx context.Context, companyID string) ([]string, error) {
	rows, err := s.exec.Query(ctx,
		`SELECT channel FROM cozinha_pass.company_channel WHERE company_id = $1 ORDER BY channel`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	defer rows.Close()

	channels := []string{}
	for rows.Next() {
		var ch string
		if err := rows.Scan(&ch); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func scanCKCompany(scanner interface{ Scan(...any) error }) (*CKCompany, error) {
	var c CKCompany
	var createdAt time.Time
	err := scanner.Scan(&c.ID, &c.Name, &c.Logo, &c.BrandColor, &c.Cuisine, &c.ActiveStatus, &createdAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = &createdAt
	c.Channels = []string{}
	return &c, nil
}

func (s *pgCKCompanyStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKCompany](), "pg_ck_company_store", nil
}
func (s *pgCKCompanyStore) GetName() string { return "pg_ck_company_store" }
func (s *pgCKCompanyStore) Name() string    { return "ck_company" }
func (s *pgCKCompanyStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgCKCompanyStore) Close() error { return nil }
