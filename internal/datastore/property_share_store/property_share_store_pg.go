package propertysharesstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
	"github.com/jackc/pgx/v5"
)

const tableName = "limalar.property_shares"

// pgPropertyShareStore implementa PropertyShareStore usando PGExecutor.
type pgPropertyShareStore struct {
	exec execution.PGExecutor
}

// NewPGPropertyShareStore cria uma instância de PropertyShareStore para Postgres.
func NewPGPropertyShareStore(exec execution.PGExecutor) PropertyShareStore {
	return &pgPropertyShareStore{exec: exec}
}

// ── LogShare ──────────────────────────────────────────────────────────────────

func (s *pgPropertyShareStore) LogShare(ctx context.Context, input *LogShareInput) (string, error) {
	if input == nil {
		return "", fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.PropertyID == "" || input.TenantID == "" {
		return "", fmt.Errorf("property_id and tenant_id are required: %v", t.ErrInvalidInput)
	}

	var id string
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(property_id, tenant_id, channel, recipient_phone, link_token, wapp_message_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tableName),
		input.PropertyID, input.TenantID, input.Channel,
		input.RecipientPhone, input.LinkToken, input.WappMessageID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to log share: %v", err)
	}
	return id, nil
}

// ── ListShares ────────────────────────────────────────────────────────────────

func (s *pgPropertyShareStore) ListShares(ctx context.Context, tenantID, propertyID string) (*ShareSummary, error) {
	if tenantID == "" || propertyID == "" {
		return nil, fmt.Errorf("tenantID and propertyID are required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, channel, recipient_phone, link_token, wapp_message_id, created_at
		FROM %s
		WHERE property_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
		LIMIT 20
	`, tableName), propertyID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shares: %v", err)
	}
	defer rows.Close()

	summary := &ShareSummary{
		ByChannel: make(map[string]int),
		Recent:    []ShareRecord{},
	}
	for rows.Next() {
		var r ShareRecord
		r.PropertyID = propertyID
		r.TenantID = tenantID
		if err := rows.Scan(
			&r.ID, &r.Channel, &r.RecipientPhone,
			&r.LinkToken, &r.WappMessageID, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan share: %v", err)
		}
		summary.Recent = append(summary.Recent, r)
		summary.ByChannel[r.Channel]++
		summary.Total++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return summary, nil
}

// ── FindByPhone ───────────────────────────────────────────────────────────────

func (s *pgPropertyShareStore) FindByPhone(ctx context.Context, phone string) (*ShareRecord, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone is required: %v", t.ErrInvalidInput)
	}

	var r ShareRecord
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, property_id, tenant_id, channel, recipient_phone, link_token, wapp_message_id, created_at
		FROM %s
		WHERE recipient_phone = $1 AND channel = 'whatsapp'
		ORDER BY created_at DESC
		LIMIT 1
	`, tableName), phone).Scan(
		&r.ID, &r.PropertyID, &r.TenantID, &r.Channel,
		&r.RecipientPhone, &r.LinkToken, &r.WappMessageID, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find share by phone: %v", err)
	}
	return &r, nil
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgPropertyShareStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[ShareRecord](), "pg_property_share_store", nil
}

func (s *pgPropertyShareStore) GetName() string {
	return "pg_property_share_store"
}

func (s *pgPropertyShareStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgPropertyShareStore) Close() error {
	return nil
}
