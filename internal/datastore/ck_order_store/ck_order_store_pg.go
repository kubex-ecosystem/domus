package ckorderstore

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

type pgCKOrderStore struct {
	exec execution.PGExecutor
}

func NewPGCKOrderStore(exec execution.PGExecutor) CKOrderStore {
	return &pgCKOrderStore{exec: exec}
}

func (s *pgCKOrderStore) Create(ctx context.Context, input *CreateCKOrderInput) (*CKOrder, error) {
	if input == nil || strings.TrimSpace(input.CompanyID) == "" {
		return nil, fmt.Errorf("company_id is required: %w", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Ticket) == "" {
		return nil, fmt.Errorf("ticket is required: %w", t.ErrInvalidInput)
	}

	now := time.Now().UTC()
	const q = `
		INSERT INTO cozinha_pass.orders
			(company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id)
		VALUES ($1,$2,$3,$4,'new',$5,$6,$7,$8,$8,$9)
		RETURNING id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
	`
	row := s.exec.QueryRow(ctx, q,
		input.CompanyID, input.Ticket, input.ClientName, input.Channel,
		input.Destination, input.Note, input.TotalValue, now, input.ExternalID,
	)
	order, err := scanCKOrder(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create ck order: %w", err)
	}

	for _, item := range input.Items {
		const iq = `
			INSERT INTO cozinha_pass.order_item (order_id, produto_id, name, quantity, unit_price, notes)
			VALUES ($1,$2,$3,$4,$5,$6)
			RETURNING id, order_id, produto_id, name, quantity, unit_price, notes
		`
		irow := s.exec.QueryRow(ctx, iq, order.ID, item.ProdutoID, item.Name, item.Quantity, item.UnitPrice, item.Notes)
		oi, err := scanCKOrderItem(irow)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
		order.Items = append(order.Items, *oi)
	}
	return order, nil
}

func (s *pgCKOrderStore) GetByID(ctx context.Context, id string) (*CKOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
		FROM cozinha_pass.orders WHERE id = $1
	`
	row := s.exec.QueryRow(ctx, q, id)
	order, err := scanCKOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck order: %w", err)
	}
	items, err := s.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (s *pgCKOrderStore) GetByTicket(ctx context.Context, companyID, ticket string) (*CKOrder, error) {
	if strings.TrimSpace(companyID) == "" || strings.TrimSpace(ticket) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
		FROM cozinha_pass.orders WHERE company_id = $1 AND ticket = $2
	`
	row := s.exec.QueryRow(ctx, q, companyID, ticket)
	order, err := scanCKOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck order by ticket: %w", err)
	}
	items, err := s.ListItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (s *pgCKOrderStore) GetByExternalID(ctx context.Context, companyID, externalID string) (*CKOrder, error) {
	if strings.TrimSpace(companyID) == "" || strings.TrimSpace(externalID) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
		FROM cozinha_pass.orders WHERE company_id = $1 AND external_id = $2
	`
	row := s.exec.QueryRow(ctx, q, companyID, externalID)
	order, err := scanCKOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck order by external_id: %w", err)
	}
	items, err := s.ListItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (s *pgCKOrderStore) UpdateStatus(ctx context.Context, input *UpdateCKOrderStatusInput) (*CKOrder, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	if !statusValido(input.Status) {
		return nil, fmt.Errorf("status %q: %w", input.Status, t.ErrInvalidInput)
	}

	// Tudo numa transação: o FOR UPDATE serializa duas mudanças simultâneas no
	// mesmo pedido, e a baixa só fica gravada se o status também ficar.
	tx, err := s.exec.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx for ck order status: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var atual, companyID string
	err = tx.QueryRow(ctx,
		`SELECT status, company_id FROM cozinha_pass.orders WHERE id = $1 FOR UPDATE`,
		input.ID,
	).Scan(&atual, &companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ck order not found: %w", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to lock ck order: %w", err)
	}
	if !TransicaoValida(atual, input.Status) {
		return nil, fmt.Errorf("%s → %s: %w", atual, input.Status, ErrTransicaoInvalida)
	}

	const q = `
		UPDATE cozinha_pass.orders SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
	`
	order, err := scanCKOrder(tx.QueryRow(ctx, q, input.Status, time.Now().UTC(), input.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to update ck order status: %w", err)
	}

	if input.Status == "delivered" && atual != "delivered" {
		if _, err := tx.Exec(ctx, baixaEntregaSQL, input.ID, companyID); err != nil {
			return nil, fmt.Errorf("failed to apply stock baixa: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit ck order status: %w", err)
	}
	return order, nil
}

// baixaEntregaSQL dá baixa no estoque de um pedido entregue ($1 = pedido,
// $2 = empresa). Mesma regra do front (explodirInsumos em utils/producao.ts):
//
//   - explode cada item ligado a produto pela ficha técnica, recursivamente,
//     multiplicando as quantidades; um produto que já está no caminho é
//     ignorado (guarda de ciclo), e a profundidade para em 32;
//   - soma o consumo por insumo, só de insumos da própria empresa;
//   - grava uma movimentação 'baixa' por insumo. O índice único parcial
//     (order_id, insumo_id) WHERE tipo='baixa' torna isso idempotente: o que
//     já tinha baixa não entra de novo, e só o que entrou desconta estoque;
//   - o estoque nunca fica negativo (GREATEST(0, ...)), como no front.
const baixaEntregaSQL = `
WITH RECURSIVE explode (produto_id, fator, caminho) AS (
	SELECT oi.produto_id, oi.quantity::numeric, ARRAY[oi.produto_id]
	FROM cozinha_pass.order_item oi
	WHERE oi.order_id = $1 AND oi.produto_id IS NOT NULL
	UNION ALL
	SELECT cf.produto_ref_id, (e.fator * cf.quantidade)::numeric, e.caminho || cf.produto_ref_id
	FROM explode e
	JOIN cozinha_pass.componente_ficha cf ON cf.produto_id = e.produto_id
	WHERE cf.ref_tipo = 'produto'
	  AND NOT (cf.produto_ref_id = ANY (e.caminho))
	  AND cardinality(e.caminho) < 32
),
consumo AS (
	SELECT cf.insumo_ref_id AS insumo_id, SUM(e.fator * cf.quantidade) AS quantidade
	FROM explode e
	JOIN cozinha_pass.componente_ficha cf ON cf.produto_id = e.produto_id
	JOIN cozinha_pass.insumo i ON i.id = cf.insumo_ref_id AND i.company_id = $2
	WHERE cf.ref_tipo = 'insumo'
	GROUP BY cf.insumo_ref_id
),
lancado AS (
	INSERT INTO cozinha_pass.movimentacao_estoque (insumo_id, order_id, tipo, quantidade, created_at)
	SELECT insumo_id, $1, 'baixa', quantidade, NOW()
	FROM consumo
	WHERE quantidade > 0
	ON CONFLICT (order_id, insumo_id) WHERE tipo = 'baixa' DO NOTHING
	RETURNING insumo_id, quantidade
)
UPDATE cozinha_pass.insumo i
SET estoque = GREATEST(0, i.estoque - l.quantidade), updated_at = NOW()
FROM lancado l
WHERE i.id = l.insumo_id
`

func (s *pgCKOrderStore) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	tag, err := s.exec.Exec(ctx, `DELETE FROM cozinha_pass.orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ck order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ck order not found: %w", t.ErrNotFound)
	}
	return nil
}

func (s *pgCKOrderStore) List(ctx context.Context, filters *CKOrderFilters) (*t.PaginatedResult[CKOrder], error) {
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
		if filters.CompanyIDs != nil {
			if len(filters.CompanyIDs) == 0 {
				wheres = append(wheres, "FALSE")
			} else {
				wheres = append(wheres, fmt.Sprintf("company_id = ANY($%d::uuid[])", idx))
				args = append(args, filters.CompanyIDs)
				idx++
			}
		}
		if filters.CompanyID != nil {
			wheres = append(wheres, fmt.Sprintf("company_id = $%d", idx))
			args = append(args, *filters.CompanyID)
			idx++
		}
		if filters.Status != nil {
			wheres = append(wheres, fmt.Sprintf("status = $%d", idx))
			args = append(args, *filters.Status)
			idx++
		}
		if filters.Channel != nil {
			wheres = append(wheres, fmt.Sprintf("channel = $%d", idx))
			args = append(args, *filters.Channel)
			idx++
		}
	}

	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM cozinha_pass.orders %s`, where), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ck orders: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at, external_id
		FROM cozinha_pass.orders %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ck orders: %w", err)
	}
	defer rows.Close()

	orders := []CKOrder{}
	for rows.Next() {
		o, err := scanCKOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ck order: %w", err)
		}
		orders = append(orders, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ck orders: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[CKOrder]{
		Data: orders, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func (s *pgCKOrderStore) ListItems(ctx context.Context, orderID string) ([]CKOrderItem, error) {
	rows, err := s.exec.Query(ctx, `
		SELECT id, order_id, produto_id, name, quantity, unit_price, notes
		FROM cozinha_pass.order_item WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list order items: %w", err)
	}
	defer rows.Close()

	items := []CKOrderItem{}
	for rows.Next() {
		item, err := scanCKOrderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func scanCKOrder(scanner interface{ Scan(...any) error }) (*CKOrder, error) {
	var o CKOrder
	var createdAt, updatedAt time.Time
	err := scanner.Scan(
		&o.ID, &o.CompanyID, &o.Ticket, &o.ClientName, &o.Channel,
		&o.Status, &o.Destination, &o.Note, &o.TotalValue, &createdAt, &updatedAt, &o.ExternalID,
	)
	if err != nil {
		return nil, err
	}
	o.CreatedAt = &createdAt
	o.UpdatedAt = &updatedAt
	o.Items = []CKOrderItem{}
	return &o, nil
}

func scanCKOrderItem(scanner interface{ Scan(...any) error }) (*CKOrderItem, error) {
	var i CKOrderItem
	err := scanner.Scan(&i.ID, &i.OrderID, &i.ProdutoID, &i.Name, &i.Quantity, &i.UnitPrice, &i.Notes)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (s *pgCKOrderStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKOrder](), "pg_ck_order_store", nil
}
func (s *pgCKOrderStore) GetName() string { return "pg_ck_order_store" }
func (s *pgCKOrderStore) Name() string    { return "ck_order" }
func (s *pgCKOrderStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgCKOrderStore) Close() error { return nil }
