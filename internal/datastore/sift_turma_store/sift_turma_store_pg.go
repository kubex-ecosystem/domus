package siftturmastore

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

type pgSiftTurmaStore struct {
	exec execution.PGExecutor
}

func NewPGSiftTurmaStore(exec execution.PGExecutor) SiftTurmaStore {
	return &pgSiftTurmaStore{exec: exec}
}

const turmaCols = `id, tenant_id, name, period_start, period_end, monthly_price,
	status, landing_slug, landing_template, created_at, updated_at`

func (s *pgSiftTurmaStore) GetBySlug(ctx context.Context, slug string) (*Turma, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, nil
	}
	row := s.exec.QueryRow(ctx, `
		SELECT `+turmaCols+` FROM sift.turma WHERE landing_slug = $1
	`, slug)
	turma, err := scanTurma(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sift turma by slug: %w", err)
	}
	return turma, nil
}

func (s *pgSiftTurmaStore) GetByID(ctx context.Context, tenantID, id string) (*Turma, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(id) == "" {
		return nil, nil
	}
	// tenant_id no WHERE, não só no filtro da aplicação: sem isso um id
	// adivinhado lê turma de outro tenant.
	row := s.exec.QueryRow(ctx, `
		SELECT `+turmaCols+` FROM sift.turma WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	turma, err := scanTurma(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sift turma: %w", err)
	}
	return turma, nil
}

func (s *pgSiftTurmaStore) List(ctx context.Context, tenantID string) ([]Turma, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	rows, err := s.exec.Query(ctx, `
		SELECT `+turmaCols+` FROM sift.turma WHERE tenant_id = $1
		ORDER BY period_start DESC NULLS LAST, name ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sift turmas: %w", err)
	}
	defer rows.Close()

	turmas := []Turma{}
	for rows.Next() {
		turma, err := scanTurma(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sift turma: %w", err)
		}
		turmas = append(turmas, *turma)
	}
	return turmas, rows.Err()
}

func (s *pgSiftTurmaStore) Create(ctx context.Context, input *CreateTurmaInput) (*Turma, error) {
	if input == nil || strings.TrimSpace(input.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("name is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.LandingSlug) == "" {
		return nil, fmt.Errorf("landing_slug is required: %w", t.ErrInvalidInput)
	}

	status := input.Status
	if strings.TrimSpace(status) == "" {
		status = "draft"
	}
	template := input.LandingTemplate
	if strings.TrimSpace(template) == "" {
		template = "essencial"
	}

	row := s.exec.QueryRow(ctx, `
		INSERT INTO sift.turma
			(tenant_id, name, period_start, period_end, monthly_price, status, landing_slug, landing_template)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+turmaCols,
		input.TenantID, input.Name, input.PeriodStart, input.PeriodEnd,
		input.MonthlyPrice, status, input.LandingSlug, template,
	)
	turma, err := scanTurma(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create sift turma: %w", err)
	}
	return turma, nil
}

func scanTurma(scanner interface{ Scan(...any) error }) (*Turma, error) {
	var turma Turma
	var createdAt, updatedAt time.Time
	if err := scanner.Scan(
		&turma.ID, &turma.TenantID, &turma.Name, &turma.PeriodStart, &turma.PeriodEnd,
		&turma.MonthlyPrice, &turma.Status, &turma.LandingSlug, &turma.LandingTemplate,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	turma.CreatedAt = &createdAt
	turma.UpdatedAt = &updatedAt
	return &turma, nil
}

func (s *pgSiftTurmaStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Turma](), "pg_sift_turma_store", nil
}
func (s *pgSiftTurmaStore) GetName() string { return "pg_sift_turma_store" }
func (s *pgSiftTurmaStore) Name() string    { return "sift_turma" }
func (s *pgSiftTurmaStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgSiftTurmaStore) Close() error { return nil }
