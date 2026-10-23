package siftleadstore

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

type pgSiftLeadStore struct {
	exec execution.PGExecutor
}

func NewPGSiftLeadStore(exec execution.PGExecutor) SiftLeadStore {
	return &pgSiftLeadStore{exec: exec}
}

const leadCols = `id, tenant_id, stage_id, turma_id, name, ig_handle, phone, email,
	source, source_detail, captured_by, assigned_to, fit_score, fit_reason,
	lost_reason, notes, created_at, updated_at`

func (s *pgSiftLeadStore) Create(ctx context.Context, input *CreateLeadInput) (*Lead, error) {
	if input == nil || strings.TrimSpace(input.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.StageID) == "" {
		return nil, fmt.Errorf("stage_id is required: %w", t.ErrInvalidInput)
	}
	// Mesma regra do CHECK sift_lead_tem_contato, verificada antes de ir ao
	// banco para o erro chegar nomeado em vez de como violação de constraint.
	// Lead sem forma de alcançar a pessoa não é lead, é anotação.
	if strings.TrimSpace(input.IGHandle) == "" &&
		strings.TrimSpace(input.Phone) == "" &&
		strings.TrimSpace(input.Email) == "" {
		return nil, fmt.Errorf("lead needs at least one of ig_handle, phone or email: %w", t.ErrInvalidInput)
	}

	row := s.exec.QueryRow(ctx, `
		INSERT INTO sift.lead
			(tenant_id, stage_id, turma_id, name, ig_handle, phone, email,
			 source, source_detail, captured_by, assigned_to, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING `+leadCols,
		input.TenantID, input.StageID, input.TurmaID, input.Name,
		strings.TrimSpace(input.IGHandle), strings.TrimSpace(input.Phone), strings.TrimSpace(input.Email),
		input.Source, input.SourceDetail, input.CapturedBy, input.AssignedTo, input.Notes,
	)
	lead, err := scanLead(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create sift lead: %w", err)
	}
	return lead, nil
}

func (s *pgSiftLeadStore) GetByID(ctx context.Context, tenantID, id string) (*Lead, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(id) == "" {
		return nil, nil
	}
	row := s.exec.QueryRow(ctx, `
		SELECT `+leadCols+` FROM sift.lead WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	lead, err := scanLead(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sift lead: %w", err)
	}
	return lead, nil
}

func (s *pgSiftLeadStore) FindDuplicate(ctx context.Context, tenantID, igHandle, phone string) (*Lead, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	handle := strings.TrimSpace(igHandle)
	ph := strings.TrimSpace(phone)
	if handle == "" && ph == "" {
		return nil, nil
	}

	// lower(ig_handle) espelha exatamente o índice único parcial
	// idx_sift_lead_tenant_handle — se a consulta divergisse do índice, a
	// checagem passaria e o INSERT ainda assim estouraria.
	row := s.exec.QueryRow(ctx, `
		SELECT `+leadCols+` FROM sift.lead
		WHERE tenant_id = $1
		  AND ( ($2 <> '' AND lower(ig_handle) = lower($2))
		     OR ($3 <> '' AND phone = $3) )
		ORDER BY created_at ASC
		LIMIT 1
	`, tenantID, handle, ph)
	lead, err := scanLead(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find duplicate sift lead: %w", err)
	}
	return lead, nil
}

func (s *pgSiftLeadStore) MoveToStage(ctx context.Context, tenantID, id, stageID string) (*Lead, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tenant_id and id are required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(stageID) == "" {
		return nil, fmt.Errorf("stage_id is required: %w", t.ErrInvalidInput)
	}

	// updated_at explícito: a extensão Go do schema não instala os triggers de
	// updated_at (só o caminho do .sql instala), e tempo-em-estágio é métrica
	// do produto — não pode depender de qual caminho de DDL rodou.
	//
	// O stage_id é conferido contra o mesmo tenant na subconsulta: sem isso um
	// stage_id de outro tenant entraria e o lead ficaria apontando para uma
	// coluna que a operação não vê.
	row := s.exec.QueryRow(ctx, `
		UPDATE sift.lead
		SET stage_id = $3, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		  AND EXISTS (SELECT 1 FROM sift.stage st WHERE st.id = $3 AND st.tenant_id = $2)
		RETURNING `+leadCols,
		id, tenantID, stageID,
	)
	lead, err := scanLead(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to move sift lead to stage: %w", err)
	}
	return lead, nil
}

func (s *pgSiftLeadStore) List(ctx context.Context, filters *LeadFilters) (*t.PaginatedResult[Lead], error) {
	if filters == nil || strings.TrimSpace(filters.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	wheres := []string{"tenant_id = $1"}
	args := []any{filters.TenantID}
	idx := 2
	add := func(clause string, v any) {
		wheres = append(wheres, fmt.Sprintf(clause, idx))
		args = append(args, v)
		idx++
	}
	if filters.StageID != nil {
		add("stage_id = $%d", *filters.StageID)
	}
	if filters.TurmaID != nil {
		add("turma_id = $%d", *filters.TurmaID)
	}
	if filters.Source != nil {
		add("source = $%d", *filters.Source)
	}
	if filters.Assigned != nil {
		add("assigned_to = $%d", *filters.Assigned)
	}
	where := "WHERE " + strings.Join(wheres, " AND ")

	var total int64
	if err := s.exec.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM sift.lead %s`, where), args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count sift leads: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT `+leadCols+` FROM sift.lead %s
		ORDER BY updated_at DESC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sift leads: %w", err)
	}
	defer rows.Close()

	leads := []Lead{}
	for rows.Next() {
		lead, err := scanLead(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sift lead: %w", err)
		}
		leads = append(leads, *lead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sift leads: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[Lead]{
		Data: leads, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func scanLead(scanner interface{ Scan(...any) error }) (*Lead, error) {
	var l Lead
	var createdAt, updatedAt time.Time
	if err := scanner.Scan(
		&l.ID, &l.TenantID, &l.StageID, &l.TurmaID, &l.Name, &l.IGHandle, &l.Phone, &l.Email,
		&l.Source, &l.SourceDetail, &l.CapturedBy, &l.AssignedTo, &l.FitScore, &l.FitReason,
		&l.LostReason, &l.Notes, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	l.CreatedAt = &createdAt
	l.UpdatedAt = &updatedAt
	return &l, nil
}

func (s *pgSiftLeadStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Lead](), "pg_sift_lead_store", nil
}
func (s *pgSiftLeadStore) GetName() string { return "pg_sift_lead_store" }
func (s *pgSiftLeadStore) Name() string    { return "sift_lead" }
func (s *pgSiftLeadStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgSiftLeadStore) Close() error { return nil }
