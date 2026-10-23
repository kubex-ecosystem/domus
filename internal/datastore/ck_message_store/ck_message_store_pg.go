package ckmessagestore

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

type pgCKMessageStore struct {
	exec execution.PGExecutor
}

func NewPGCKMessageStore(exec execution.PGExecutor) CKMessageStore {
	return &pgCKMessageStore{exec: exec}
}

func (s *pgCKMessageStore) Create(ctx context.Context, input *CreateCKMessageInput) (*CKMessage, error) {
	if input == nil || strings.TrimSpace(input.Content) == "" {
		return nil, fmt.Errorf("content is required: %w", t.ErrInvalidInput)
	}
	const q = `
		INSERT INTO cozinha_pass.message
			(company_id, order_id, channel, client_name, content, direction, read, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, company_id, order_id, channel, client_name, content, direction, read, created_at
	`
	row := s.exec.QueryRow(ctx, q,
		input.CompanyID, input.OrderID, input.Channel,
		input.ClientName, input.Content, input.Direction, input.Read, time.Now().UTC(),
	)
	return scanCKMessage(row)
}

func (s *pgCKMessageStore) GetByID(ctx context.Context, id string) (*CKMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	const q = `
		SELECT id, company_id, order_id, channel, client_name, content, direction, read, created_at
		FROM cozinha_pass.message WHERE id = $1
	`
	msg, err := scanCKMessage(s.exec.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ck message: %w", err)
	}
	return msg, nil
}

func (s *pgCKMessageStore) List(ctx context.Context, filters *CKMessageFilters) (*t.PaginatedResult[CKMessage], error) {
	page, limit := 1, 50
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
		if filters.OrderID != nil {
			wheres = append(wheres, fmt.Sprintf("order_id = $%d", idx))
			args = append(args, *filters.OrderID)
			idx++
		}
		if filters.Channel != nil {
			wheres = append(wheres, fmt.Sprintf("channel = $%d", idx))
			args = append(args, *filters.Channel)
			idx++
		}
		if filters.Direction != nil {
			wheres = append(wheres, fmt.Sprintf("direction = $%d", idx))
			args = append(args, *filters.Direction)
			idx++
		}
		if filters.Unread != nil && *filters.Unread {
			wheres = append(wheres, "read = false AND direction = 'inbound'")
		}
	}

	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if err := s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM cozinha_pass.message %s`, where), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ck messages: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, company_id, order_id, channel, client_name, content, direction, read, created_at
		FROM cozinha_pass.message %s ORDER BY created_at ASC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ck messages: %w", err)
	}
	defer rows.Close()

	messages := []CKMessage{}
	for rows.Next() {
		m, err := scanCKMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ck message: %w", err)
		}
		messages = append(messages, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ck messages: %w", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &t.PaginatedResult[CKMessage]{
		Data: messages, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	}, nil
}

func (s *pgCKMessageStore) MarkRead(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required: %w", t.ErrInvalidInput)
	}
	tag, err := s.exec.Exec(ctx, `UPDATE cozinha_pass.message SET read = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark message read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ck message not found: %w", t.ErrNotFound)
	}
	return nil
}

func (s *pgCKMessageStore) MarkConversationRead(ctx context.Context, companyID, channel, clientName string, orderID *string) error {
	var err error
	if orderID != nil {
		_, err = s.exec.Exec(ctx, `
			UPDATE cozinha_pass.message SET read = true
			WHERE company_id = $1 AND order_id = $2 AND direction = 'inbound'
		`, companyID, *orderID)
	} else {
		_, err = s.exec.Exec(ctx, `
			UPDATE cozinha_pass.message SET read = true
			WHERE company_id = $1 AND channel = $2 AND client_name = $3
			  AND order_id IS NULL AND direction = 'inbound'
		`, companyID, channel, clientName)
	}
	if err != nil {
		return fmt.Errorf("failed to mark conversation read: %w", err)
	}
	return nil
}

func scanCKMessage(scanner interface{ Scan(...any) error }) (*CKMessage, error) {
	var m CKMessage
	var createdAt time.Time
	err := scanner.Scan(
		&m.ID, &m.CompanyID, &m.OrderID, &m.Channel,
		&m.ClientName, &m.Content, &m.Direction, &m.Read, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	m.CreatedAt = &createdAt
	return &m, nil
}

func (s *pgCKMessageStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKMessage](), "pg_ck_message_store", nil
}
func (s *pgCKMessageStore) GetName() string { return "pg_ck_message_store" }
func (s *pgCKMessageStore) Name() string    { return "ck_message" }
func (s *pgCKMessageStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgCKMessageStore) Close() error { return nil }
