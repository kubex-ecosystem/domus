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
			(company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at)
		VALUES ($1,$2,$3,$4,'new',$5,$6,$7,$8,$8)
		RETURNING id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at
	`
	row := s.exec.QueryRow(ctx, q,
		input.CompanyID, input.Ticket, input.ClientName, input.Channel,
		input.Destination, input.Note, input.TotalValue, now,
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
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at
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
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at
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

func (s *pgCKOrderStore) UpdateStatus(ctx context.Context, input *UpdateCKOrderStatusInput) (*CKOrder, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	const q = `
		UPDATE cozinha_pass.orders SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at
	`
	row := s.exec.QueryRow(ctx, q, input.Status, time.Now().UTC(), input.ID)
	order, err := scanCKOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ck order not found: %w", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update ck order status: %w", err)
	}
	return order, nil
}

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
		SELECT id, company_id, ticket, client_name, channel, status, destination, note, total_value, created_at, updated_at
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
		&o.Status, &o.Destination, &o.Note, &o.TotalValue, &createdAt, &updatedAt,
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
