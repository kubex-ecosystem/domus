package wappmessagestore

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

const tableName = "limalar.wapp_messages"

// pgWappMessageStore implementa WappMessageStore usando PGExecutor.
type pgWappMessageStore struct {
	exec execution.PGExecutor
}

// NewPGWappMessageStore cria uma instância de WappMessageStore para Postgres.
func NewPGWappMessageStore(exec execution.PGExecutor) WappMessageStore {
	return &pgWappMessageStore{exec: exec}
}

// ── Save ──────────────────────────────────────────────────────────────────────

func (s *pgWappMessageStore) Save(ctx context.Context, input *SaveMessageInput) error {
	if input == nil {
		return fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.TenantID == "" || input.Phone == "" {
		return fmt.Errorf("tenant_id and phone are required: %v", t.ErrInvalidInput)
	}

	var pid *string
	if input.PropertyID != "" {
		pid = &input.PropertyID
	}

	_, err := s.exec.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, phone, property_id, direction, role, content, score)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, tableName),
		input.TenantID, input.Phone, pid,
		input.Direction, input.Role, input.Content, input.Score,
	)
	if err != nil {
		return fmt.Errorf("failed to save wapp message: %v", err)
	}
	return nil
}

// ── ListByPhone ───────────────────────────────────────────────────────────────

func (s *pgWappMessageStore) ListByPhone(ctx context.Context, tenantID, phone string) ([]WappMessage, error) {
	if tenantID == "" || phone == "" {
		return nil, fmt.Errorf("tenantID and phone are required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, COALESCE(property_id::text, ''), tenant_id, phone, direction, role, content, score, created_at
		FROM %s
		WHERE tenant_id = $1 AND phone = $2
		ORDER BY created_at ASC
		LIMIT 100
	`, tableName), tenantID, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by phone: %v", err)
	}
	defer rows.Close()

	var msgs []WappMessage
	for rows.Next() {
		var m WappMessage
		if err := rows.Scan(
			&m.ID, &m.PropertyID, &m.TenantID, &m.Phone,
			&m.Direction, &m.Role, &m.Content, &m.Score, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return msgs, nil
}

// ── ListThreads ───────────────────────────────────────────────────────────────

func (s *pgWappMessageStore) ListThreads(ctx context.Context, tenantID string) ([]ThreadSummary, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenantID is required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, `
		SELECT
			m.phone,
			m.score,
			m.role             AS last_role,
			m.content          AS last_message,
			m.created_at       AS last_message_at,
			m.property_id,
			COALESCE(p.type, '')         AS property_type,
			COALESCE(p.neighborhood, '') AS property_neighborhood,
			cnt.message_count
		FROM (
			SELECT DISTINCT ON (phone) *
			FROM limalar.wapp_messages
			WHERE tenant_id = $1
			ORDER BY phone, created_at DESC
		) m
		LEFT JOIN limalar.properties p ON p.id = m.property_id
		LEFT JOIN (
			SELECT phone, COUNT(*) AS message_count
			FROM limalar.wapp_messages
			WHERE tenant_id = $1
			GROUP BY phone
		) cnt ON cnt.phone = m.phone
		ORDER BY m.created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list threads: %v", err)
	}
	defer rows.Close()

	var threads []ThreadSummary
	for rows.Next() {
		var thread ThreadSummary
		var propID *string
		if err := rows.Scan(
			&thread.Phone, &thread.Score, &thread.LastRole, &thread.LastMessage,
			&thread.LastMessageAt, &propID,
			&thread.PropertyType, &thread.PropertyNeighborhood, &thread.MessageCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan thread: %v", err)
		}
		if propID != nil {
			thread.PropertyID = *propID
		}
		threads = append(threads, thread)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return threads, nil
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgWappMessageStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[WappMessage](), "pg_wapp_message_store", nil
}

func (s *pgWappMessageStore) GetName() string {
	return "pg_wapp_message_store"
}

func (s *pgWappMessageStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgWappMessageStore) Close() error {
	return nil
}
