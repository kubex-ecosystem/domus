package ckprodutostore

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

type pgCKProdutoStore struct {
	exec execution.PGExecutor
}

func NewPGCKProdutoStore(exec execution.PGExecutor) CKProdutoStore {
	return &pgCKProdutoStore{exec: exec}
}

func (s *pgCKProdutoStore) Create(ctx context.Context, input *CreateCKProdutoInput) (*CKProduto, error) {
	if input == nil || strings.TrimSpace(input.Nome) == "" {
		return nil, fmt.Errorf("nome is required: %w", t.ErrInvalidInput)
	}
	now := time.Now().UTC()
	const q = `
		INSERT INTO cozinha_pass.produto (company_id, nome, tipo, preco_venda, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$5)
		RETURNING id, company_id, nome, tipo, preco_venda, created_at, updated_at
	`
	row := s.exec.QueryRow(ctx, q, input.CompanyID, input.Nome, input.Tipo, input.PrecoVenda, now)
	produto, err := scanCKProduto(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create ck produto: %w", err)
	}

	if len(input.Componentes) > 0 {
		if err := s.SetComponentes(ctx, produto.ID, input.Componentes); err != nil {
			return nil, err
		}
		componentes, err := s.ListComponentes(ctx, produto.ID)
		if err != nil {
			return nil, err
		}
		produto.Componentes = componentes
	}
	return produto, nil
}

func (s *pgCKProdutoStore) GetByID(ctx context.Context, id string) (*CKProduto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, nome, tipo, preco_venda, created_at, updated_at
		FROM cozinha_pass.produto WHERE id = $1
	`
	produto, err := scanCKProduto(s.exec.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck produto: %w", err)
	}
	componentes, err := s.ListComponentes(ctx, id)
	if err != nil {
		return nil, err
	}
	produto.Componentes = componentes
	return produto, nil
}

func (s *pgCKProdutoStore) Update(ctx context.Context, input *UpdateCKProdutoInput) (*CKProduto, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if input.Nome != nil {
		sets = append(sets, fmt.Sprintf("nome = $%d", idx)); args = append(args, *input.Nome); idx++
	}
	if input.Tipo != nil {
		sets = append(sets, fmt.Sprintf("tipo = $%d", idx)); args = append(args, *input.Tipo); idx++
	}
	if input.PrecoVenda != nil {
		sets = append(sets, fmt.Sprintf("preco_venda = $%d", idx)); args = append(args, *input.PrecoVenda); idx++
	}

	var produto *CKProduto
	if len(sets) > 0 {
		sets = append(sets, fmt.Sprintf("updated_at = $%d", idx)); args = append(args, time.Now().UTC()); idx++
		args = append(args, input.ID)
		query := fmt.Sprintf(`
			UPDATE cozinha_pass.produto SET %s WHERE id = $%d
			RETURNING id, company_id, nome, tipo, preco_venda, created_at, updated_at
		`, strings.Join(sets, ", "), idx)
		var err error
		produto, err = scanCKProduto(s.exec.QueryRow(ctx, query, args...))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("ck produto not found: %w", t.ErrNotFound)
			}
			return nil, fmt.Errorf("failed to update ck produto: %w", err)
		}
	} else {
		var err error
		produto, err = s.GetByID(ctx, input.ID)
		if err != nil {
			return nil, err
		}
	}

	if input.Componentes != nil {
		if err := s.SetComponentes(ctx, input.ID, input.Componentes); err != nil {
			return nil, err
		}
	}
	componentes, err := s.ListComponentes(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	produto.Componentes = componentes
	return produto, nil
}

func (s *pgCKProdutoStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	tag, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.produto WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ck produto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ck produto not found: %w", t.ErrNotFound)
	}
	return nil
}

func (s *pgCKProdutoStore) List(ctx context.Context, filters *CKProdutoFilters) (*t.PaginatedResult[CKProduto], error) {
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
		if filters.Tipo != nil {
			wheres = append(wheres, fmt.Sprintf("tipo = $%d", idx)); args = append(args, *filters.Tipo); idx++
		}
	}

	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM cozinha_pass.produto %s`, where), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ck produtos: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, company_id, nome, tipo, preco_venda, created_at, updated_at
		FROM cozinha_pass.produto %s ORDER BY nome ASC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ck produtos: %w", err)
	}
	defer rows.Close()

	produtos := []CKProduto{}
	for rows.Next() {
		p, err := scanCKProduto(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ck produto: %w", err)
		}
		produtos = append(produtos, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ck produtos: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[CKProduto]{
		Data: produtos, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func (s *pgCKProdutoStore) SetComponentes(ctx context.Context, produtoID string, componentes []CreateCKComponenteInput) error {
	if _, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.componente_ficha WHERE produto_id = $1`, produtoID); err != nil {
		return fmt.Errorf("failed to clear componentes: %w", err)
	}
	for _, c := range componentes {
		const q = `
			INSERT INTO cozinha_pass.componente_ficha
				(produto_id, ref_tipo, insumo_ref_id, produto_ref_id, quantidade)
			VALUES ($1,$2,$3,$4,$5)
		`
		if _, err := s.exec.Exec(ctx, q, produtoID, c.RefTipo, c.InsumoRefID, c.ProdutoRefID, c.Quantidade); err != nil {
			return fmt.Errorf("failed to insert componente: %w", err)
		}
	}
	return nil
}

func (s *pgCKProdutoStore) ListComponentes(ctx context.Context, produtoID string) ([]CKComponente, error) {
	rows, err := s.exec.Query(ctx, `
		SELECT id, produto_id, ref_tipo, insumo_ref_id, produto_ref_id, quantidade
		FROM cozinha_pass.componente_ficha WHERE produto_id = $1
	`, produtoID)
	if err != nil {
		return nil, fmt.Errorf("failed to list componentes: %w", err)
	}
	defer rows.Close()

	componentes := []CKComponente{}
	for rows.Next() {
		var c CKComponente
		if err := rows.Scan(&c.ID, &c.ProdutoID, &c.RefTipo, &c.InsumoRefID, &c.ProdutoRefID, &c.Quantidade); err != nil {
			return nil, fmt.Errorf("failed to scan componente: %w", err)
		}
		componentes = append(componentes, c)
	}
	return componentes, rows.Err()
}

func scanCKProduto(scanner interface{ Scan(...any) error }) (*CKProduto, error) {
	var p CKProduto
	var createdAt, updatedAt time.Time
	err := scanner.Scan(&p.ID, &p.CompanyID, &p.Nome, &p.Tipo, &p.PrecoVenda, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = &createdAt
	p.UpdatedAt = &updatedAt
	p.Componentes = []CKComponente{}
	return &p, nil
}

func (s *pgCKProdutoStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKProduto](), "pg_ck_produto_store", nil
}
func (s *pgCKProdutoStore) GetName() string { return "pg_ck_produto_store" }
func (s *pgCKProdutoStore) Name() string    { return "ck_produto" }
func (s *pgCKProdutoStore) Validate() error { if s.exec == nil { return errors.New("PGExecutor is nil") }; return nil }
func (s *pgCKProdutoStore) Close() error    { return nil }
