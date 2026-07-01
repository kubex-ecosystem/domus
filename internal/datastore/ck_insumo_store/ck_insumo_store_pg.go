package ckinsumostore

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

type pgCKInsumoStore struct {
	exec execution.PGExecutor
}

func NewPGCKInsumoStore(exec execution.PGExecutor) CKInsumoStore {
	return &pgCKInsumoStore{exec: exec}
}

func (s *pgCKInsumoStore) Create(ctx context.Context, input *CreateCKInsumoInput) (*CKInsumo, error) {
	if input == nil || strings.TrimSpace(input.Nome) == "" {
		return nil, fmt.Errorf("nome is required: %w", t.ErrInvalidInput)
	}
	now := time.Now().UTC()
	const q = `
		INSERT INTO cozinha_pass.insumo
			(company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$7)
		RETURNING id, company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at
	`
	row := s.exec.QueryRow(ctx, q,
		input.CompanyID, input.Nome, input.Unidade,
		input.Estoque, input.EstoqueMinimo, input.CustoUnitario, now,
	)
	return scanCKInsumo(row)
}

func (s *pgCKInsumoStore) GetByID(ctx context.Context, id string) (*CKInsumo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at
		FROM cozinha_pass.insumo WHERE id = $1
	`
	insumo, err := scanCKInsumo(s.exec.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck insumo: %w", err)
	}
	return insumo, nil
}

func (s *pgCKInsumoStore) Update(ctx context.Context, input *UpdateCKInsumoInput) (*CKInsumo, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if input.Nome != nil {
		sets = append(sets, fmt.Sprintf("nome = $%d", idx)); args = append(args, *input.Nome); idx++
	}
	if input.Unidade != nil {
		sets = append(sets, fmt.Sprintf("unidade = $%d", idx)); args = append(args, *input.Unidade); idx++
	}
	if input.Estoque != nil {
		sets = append(sets, fmt.Sprintf("estoque = $%d", idx)); args = append(args, *input.Estoque); idx++
	}
	if input.EstoqueMinimo != nil {
		sets = append(sets, fmt.Sprintf("estoque_minimo = $%d", idx)); args = append(args, *input.EstoqueMinimo); idx++
	}
	if input.CustoUnitario != nil {
		sets = append(sets, fmt.Sprintf("custo_unitario = $%d", idx)); args = append(args, *input.CustoUnitario); idx++
	}

	if len(sets) == 0 {
		return s.GetByID(ctx, input.ID)
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", idx)); args = append(args, time.Now().UTC()); idx++
	args = append(args, input.ID)

	query := fmt.Sprintf(`
		UPDATE cozinha_pass.insumo SET %s WHERE id = $%d
		RETURNING id, company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at
	`, strings.Join(sets, ", "), idx)

	insumo, err := scanCKInsumo(s.exec.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ck insumo not found: %w", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update ck insumo: %w", err)
	}
	return insumo, nil
}

func (s *pgCKInsumoStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	tag, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.insumo WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ck insumo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ck insumo not found: %w", t.ErrNotFound)
	}
	return nil
}

func (s *pgCKInsumoStore) List(ctx context.Context, filters *CKInsumoFilters) (*t.PaginatedResult[CKInsumo], error) {
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
		if filters.CompanyID != nil {
			wheres = append(wheres, fmt.Sprintf("company_id = $%d", idx)); args = append(args, *filters.CompanyID); idx++
		}
		if filters.BaixoEstoque != nil && *filters.BaixoEstoque {
			wheres = append(wheres, "estoque <= estoque_minimo")
		}
	}

	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM cozinha_pass.insumo %s`, where), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ck insumos: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at
		FROM cozinha_pass.insumo %s ORDER BY nome ASC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ck insumos: %w", err)
	}
	defer rows.Close()

	insumos := []CKInsumo{}
	for rows.Next() {
		i, err := scanCKInsumo(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ck insumo: %w", err)
		}
		insumos = append(insumos, *i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ck insumos: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[CKInsumo]{
		Data: insumos, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func (s *pgCKInsumoStore) AjustarEstoque(ctx context.Context, id string, delta float64) (*CKInsumo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	const q = `
		UPDATE cozinha_pass.insumo
		SET estoque = GREATEST(0, estoque + $1), updated_at = $2
		WHERE id = $3
		RETURNING id, company_id, nome, unidade, estoque, estoque_minimo, custo_unitario, created_at, updated_at
	`
	insumo, err := scanCKInsumo(s.exec.QueryRow(ctx, q, delta, time.Now().UTC(), id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ck insumo not found: %w", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to adjust estoque: %w", err)
	}
	return insumo, nil
}

func scanCKInsumo(scanner interface{ Scan(...any) error }) (*CKInsumo, error) {
	var i CKInsumo
	var createdAt, updatedAt time.Time
	err := scanner.Scan(
		&i.ID, &i.CompanyID, &i.Nome, &i.Unidade,
		&i.Estoque, &i.EstoqueMinimo, &i.CustoUnitario,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	i.CreatedAt = &createdAt
	i.UpdatedAt = &updatedAt
	return &i, nil
}

func (s *pgCKInsumoStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKInsumo](), "pg_ck_insumo_store", nil
}
func (s *pgCKInsumoStore) GetName() string { return "pg_ck_insumo_store" }
func (s *pgCKInsumoStore) Name() string    { return "ck_insumo" }
func (s *pgCKInsumoStore) Validate() error { if s.exec == nil { return errors.New("PGExecutor is nil") }; return nil }
func (s *pgCKInsumoStore) Close() error    { return nil }
