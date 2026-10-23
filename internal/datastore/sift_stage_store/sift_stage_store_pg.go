package siftstagestore

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

type pgSiftStageStore struct {
	exec execution.PGExecutor
}

func NewPGSiftStageStore(exec execution.PGExecutor) SiftStageStore {
	return &pgSiftStageStore{exec: exec}
}

const stageCols = `id, tenant_id, key, name, order_index, color, is_terminal, created_at, updated_at`

func (s *pgSiftStageStore) List(ctx context.Context, tenantID string) ([]Stage, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	rows, err := s.exec.Query(ctx, `
		SELECT `+stageCols+`
		FROM sift.stage WHERE tenant_id = $1
		ORDER BY order_index ASC, name ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sift stages: %w", err)
	}
	defer rows.Close()

	stages := []Stage{}
	for rows.Next() {
		st, err := scanStage(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sift stage: %w", err)
		}
		stages = append(stages, *st)
	}
	return stages, rows.Err()
}

func (s *pgSiftStageStore) GetByKey(ctx context.Context, tenantID, key string) (*Stage, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(key) == "" {
		return nil, nil
	}
	row := s.exec.QueryRow(ctx, `
		SELECT `+stageCols+`
		FROM sift.stage WHERE tenant_id = $1 AND key = $2
	`, tenantID, key)
	st, err := scanStage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sift stage: %w", err)
	}
	return st, nil
}

func (s *pgSiftStageStore) Upsert(ctx context.Context, tenantID string, inputs []UpsertStageInput) ([]Stage, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %w", t.ErrInvalidInput)
	}
	if len(inputs) == 0 {
		return []Stage{}, nil
	}

	// Tudo ou nada: um seed que aplica metade deixa o tenant com pipeline
	// incompleto e leads sem estágio para apontar.
	tx, err := s.exec.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx for sift stage upsert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// updated_at explícito, e não por trigger: a extensão Go do schema não
	// instala os triggers (só o caminho do .sql instala), então depender deles
	// tornaria a corretude função de qual caminho de DDL rodou.
	const q = `
		INSERT INTO sift.stage (tenant_id, key, name, order_index, color, is_terminal)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			name        = EXCLUDED.name,
			order_index = EXCLUDED.order_index,
			color       = EXCLUDED.color,
			is_terminal = EXCLUDED.is_terminal,
			updated_at  = NOW()
		RETURNING ` + stageCols

	stages := make([]Stage, 0, len(inputs))
	for _, in := range inputs {
		if strings.TrimSpace(in.Key) == "" {
			return nil, fmt.Errorf("stage key is required: %w", t.ErrInvalidInput)
		}
		name := in.Name
		if strings.TrimSpace(name) == "" {
			name = in.Key
		}
		row := tx.QueryRow(ctx, q, tenantID, in.Key, name, in.OrderIndex, in.Color, in.IsTerminal)
		st, err := scanStage(row)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert sift stage %q: %w", in.Key, err)
		}
		stages = append(stages, *st)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit sift stage upsert: %w", err)
	}
	return stages, nil
}

func scanStage(scanner interface{ Scan(...any) error }) (*Stage, error) {
	var st Stage
	var createdAt, updatedAt time.Time
	if err := scanner.Scan(
		&st.ID, &st.TenantID, &st.Key, &st.Name,
		&st.OrderIndex, &st.Color, &st.IsTerminal, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	st.CreatedAt = &createdAt
	st.UpdatedAt = &updatedAt
	return &st, nil
}

func (s *pgSiftStageStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Stage](), "pg_sift_stage_store", nil
}
func (s *pgSiftStageStore) GetName() string { return "pg_sift_stage_store" }
func (s *pgSiftStageStore) Name() string    { return "sift_stage" }
func (s *pgSiftStageStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgSiftStageStore) Close() error { return nil }
