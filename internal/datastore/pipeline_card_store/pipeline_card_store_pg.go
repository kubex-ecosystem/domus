package pipelinecardstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

const cardsTable = "limalar.pipeline_cards"

type pgPipelineCardStore struct {
	exec execution.PGExecutor
}

// NewPGPipelineCardStore cria uma instância de PipelineCardStore para Postgres.
func NewPGPipelineCardStore(exec execution.PGExecutor) PipelineCardStore {
	return &pgPipelineCardStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgPipelineCardStore) Create(ctx context.Context, input *CreatePipelineCardInput) (*PipelineCard, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required: %v", t.ErrInvalidInput)
	}

	col := input.Col
	if col == "" {
		col = "new"
	}

	var card PipelineCard
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, lead_id, property_id, col, position, title, value, score, ai_tip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, lead_id, property_id, col, position, title, value, score, ai_tip, created_at, updated_at
	`, cardsTable),
		input.TenantID, input.LeadID, input.PropertyID,
		col, input.Position, input.Title, input.Value, input.Score, input.AiTip,
	).Scan(
		&card.ID, &card.TenantID, &card.LeadID, &card.PropertyID,
		&card.Col, &card.Position, &card.Title, &card.Value,
		&card.Score, &card.AiTip, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline card: %w", err)
	}
	return &card, nil
}

// ── UpdateColumn ──────────────────────────────────────────────────────────────

func (s *pgPipelineCardStore) UpdateColumn(ctx context.Context, tenantID, id, col string, position int) error {
	if tenantID == "" || id == "" || col == "" {
		return fmt.Errorf("tenantID, id and col are required: %v", t.ErrInvalidInput)
	}

	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET col = $1, position = $2 WHERE id = $3 AND tenant_id = $4
	`, cardsTable), col, position, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update card column: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("card not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── UpdatePosition ────────────────────────────────────────────────────────────

func (s *pgPipelineCardStore) UpdatePosition(ctx context.Context, tenantID, id string, position int) error {
	if tenantID == "" || id == "" {
		return fmt.Errorf("tenantID and id are required: %v", t.ErrInvalidInput)
	}

	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET position = $1 WHERE id = $2 AND tenant_id = $3
	`, cardsTable), position, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update card position: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("card not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *pgPipelineCardStore) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return fmt.Errorf("tenantID and id are required: %v", t.ErrInvalidInput)
	}

	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s WHERE id = $1 AND tenant_id = $2
	`, cardsTable), id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete pipeline card: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("card not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── ListByTenant ──────────────────────────────────────────────────────────────

func (s *pgPipelineCardStore) ListByTenant(ctx context.Context, filters *PipelineCardFilters) ([]PipelineCard, error) {
	if filters == nil {
		filters = &PipelineCardFilters{}
	}

	where, args := buildCardFilters(filters)

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, lead_id, property_id, col, position, title, value, score, ai_tip, created_at, updated_at
		FROM %s
		%s
		ORDER BY col ASC, position ASC
	`, cardsTable, where), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline cards: %w", err)
	}
	defer rows.Close()

	cards := []PipelineCard{}
	for rows.Next() {
		var c PipelineCard
		if err := scanCard(rows, &c); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgPipelineCardStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[PipelineCard](), "pg_pipeline_card_store", nil
}

func (s *pgPipelineCardStore) GetName() string { return "pg_pipeline_card_store" }

func (s *pgPipelineCardStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgPipelineCardStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanCard(row pgx.Row, c *PipelineCard) error {
	return row.Scan(
		&c.ID, &c.TenantID, &c.LeadID, &c.PropertyID,
		&c.Col, &c.Position, &c.Title, &c.Value,
		&c.Score, &c.AiTip, &c.CreatedAt, &c.UpdatedAt,
	)
}

func buildCardFilters(filters *PipelineCardFilters) (string, []any) {
	where := []string{}
	args := []any{}

	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, filters.TenantID)
	}
	if filters.Col != nil && *filters.Col != "" {
		where = append(where, fmt.Sprintf("col = $%d", len(args)+1))
		args = append(args, *filters.Col)
	}
	if filters.LeadID != nil && *filters.LeadID != "" {
		where = append(where, fmt.Sprintf("lead_id = $%d", len(args)+1))
		args = append(args, *filters.LeadID)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}
