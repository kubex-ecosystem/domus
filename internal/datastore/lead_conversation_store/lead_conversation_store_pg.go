package leadconversationstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
	"github.com/jackc/pgx/v5"
)

const tableName = "limalar.lead_conversations"

// pgLeadConversationStore implementa LeadConversationStore usando PGExecutor.
type pgLeadConversationStore struct {
	exec execution.PGExecutor
}

// NewPGLeadConversationStore cria uma instância de LeadConversationStore para Postgres.
func NewPGLeadConversationStore(exec execution.PGExecutor) LeadConversationStore {
	return &pgLeadConversationStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgLeadConversationStore) Create(ctx context.Context, input *CreateConversationInput) (*LeadConversation, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	if input.PropertyID == "" || input.TenantID == "" || input.Phone == "" {
		return nil, fmt.Errorf("property_id, tenant_id and phone are required: %v", t.ErrInvalidInput)
	}

	status := input.Status
	if status == "" {
		status = "pending"
	}

	var conv LeadConversation
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(property_id, tenant_id, phone, inbound_message, ai_reply, score, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, property_id, tenant_id, phone, inbound_message, ai_reply, score, status, created_at
	`, tableName),
		input.PropertyID, input.TenantID, input.Phone,
		input.InboundMessage, input.AIReply, input.Score, status,
	).Scan(
		&conv.ID, &conv.PropertyID, &conv.TenantID, &conv.Phone,
		&conv.InboundMessage, &conv.AIReply, &conv.Score, &conv.Status, &conv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead conversation: %v", err)
	}
	return &conv, nil
}

// ── List ──────────────────────────────────────────────────────────────────────

func (s *pgLeadConversationStore) List(ctx context.Context, filters *ConversationFilters) (*t.PaginatedResult[LeadConversation], error) {
	if filters == nil {
		filters = &ConversationFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildFilters(filters)

	q := fmt.Sprintf(`
		SELECT id, property_id, tenant_id, phone, inbound_message, ai_reply, score, status, created_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, tableName, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list lead conversations: %v", err)
	}
	defer rows.Close()

	convs := []LeadConversation{}
	for rows.Next() {
		var c LeadConversation
		if err := scanConversation(rows, &c); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %v", err)
		}
		convs = append(convs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	countWhere, countArgs := buildFilters(filters)
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableName, countWhere)
	var total int64
	if err := s.exec.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count conversations: %v", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &t.PaginatedResult[LeadConversation]{
		Data:       convs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ── ListByProperty ────────────────────────────────────────────────────────────

func (s *pgLeadConversationStore) ListByProperty(ctx context.Context, tenantID, propertyID string) ([]LeadConversation, error) {
	if tenantID == "" || propertyID == "" {
		return nil, fmt.Errorf("tenantID and propertyID are required: %v", t.ErrInvalidInput)
	}

	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, property_id, tenant_id, phone, inbound_message, ai_reply, score, status, created_at
		FROM %s
		WHERE property_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
	`, tableName), propertyID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations by property: %v", err)
	}
	defer rows.Close()

	var convs []LeadConversation
	for rows.Next() {
		var c LeadConversation
		if err := scanConversation(rows, &c); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %v", err)
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func (s *pgLeadConversationStore) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	if tenantID == "" || id == "" || status == "" {
		return fmt.Errorf("tenantID, id and status are required: %v", t.ErrInvalidInput)
	}

	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET status = $1 WHERE id = $2 AND tenant_id = $3
	`, tableName), status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("conversation not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgLeadConversationStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[LeadConversation](), "pg_lead_conversation_store", nil
}

func (s *pgLeadConversationStore) GetName() string {
	return "pg_lead_conversation_store"
}

func (s *pgLeadConversationStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgLeadConversationStore) Close() error {
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanConversation(row pgx.Row, c *LeadConversation) error {
	return row.Scan(
		&c.ID, &c.PropertyID, &c.TenantID, &c.Phone,
		&c.InboundMessage, &c.AIReply, &c.Score, &c.Status, &c.CreatedAt,
	)
}

func buildFilters(filters *ConversationFilters) (string, []any) {
	where := []string{}
	args := []any{}

	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, filters.TenantID)
	}
	if filters.PropertyID != nil && *filters.PropertyID != "" {
		where = append(where, fmt.Sprintf("property_id = $%d", len(args)+1))
		args = append(args, *filters.PropertyID)
	}
	if filters.Phone != nil && *filters.Phone != "" {
		where = append(where, fmt.Sprintf("phone = $%d", len(args)+1))
		args = append(args, *filters.Phone)
	}
	if filters.Status != nil && *filters.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filters.Status)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
