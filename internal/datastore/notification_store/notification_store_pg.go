package notificationstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	notifTable = "limalar.notifications"
	prefTable  = "limalar.notification_prefs"
	notifCols  = "id, tenant_id, user_id, type, urgency, title, body, reference_id, reference_type, read_at, created_at"
)

// ── NotificationStore pg ─────────────────────────────────────────────────────

type pgNotificationStore struct {
	exec execution.PGExecutor
}

// NewPGNotificationStore cria uma instância de NotificationStore para Postgres.
func NewPGNotificationStore(exec execution.PGExecutor) NotificationStore {
	return &pgNotificationStore{exec: exec}
}

func (s *pgNotificationStore) Create(ctx context.Context, input *CreateNotificationInput) (*Notification, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	urgency := input.Urgency
	if urgency == "" {
		urgency = "normal"
	}
	var n Notification
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, user_id, type, urgency, title, body, reference_id, reference_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s
	`, notifTable, notifCols),
		input.TenantID, input.UserID, input.Type, urgency,
		input.Title, input.Body, input.ReferenceID, input.ReferenceType,
	).Scan(&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Urgency,
		&n.Title, &n.Body, &n.ReferenceID, &n.ReferenceType, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	return &n, nil
}

func (s *pgNotificationStore) ListByUser(ctx context.Context, filters *NotificationFilters) (*t.PaginatedResult[Notification], error) {
	if filters == nil {
		filters = &NotificationFilters{}
	}
	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildNotifFilters(filters)
	q := fmt.Sprintf(`
		SELECT %s FROM %s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, notifCols, notifTable, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifs := []Notification{}
	for rows.Next() {
		var n Notification
		if err := scanNotif(rows, &n); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}

	cWhere, cArgs := buildNotifFilters(filters)
	var total int64
	_ = s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, notifTable, cWhere), cArgs...).Scan(&total)

	tp := int(total) / limit
	if int(total)%limit != 0 {
		tp++
	}
	return &t.PaginatedResult[Notification]{Data: notifs, Total: total, Page: page, Limit: limit, TotalPages: tp}, nil
}

func (s *pgNotificationStore) MarkRead(ctx context.Context, tenantID, userID, id string) error {
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND tenant_id = $3 AND read_at IS NULL
	`, notifTable), id, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification not found or already read: %v", t.ErrNotFound)
	}
	return nil
}

func (s *pgNotificationStore) MarkAllRead(ctx context.Context, tenantID, userID string) error {
	_, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET read_at = NOW() WHERE user_id = $1 AND tenant_id = $2 AND read_at IS NULL
	`, notifTable), userID, tenantID)
	return err
}

func (s *pgNotificationStore) UnreadCount(ctx context.Context, tenantID, userID string) (int64, error) {
	var count int64
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE user_id = $1 AND tenant_id = $2 AND read_at IS NULL
	`, notifTable), userID, tenantID).Scan(&count)
	return count, err
}

func (s *pgNotificationStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Notification](), "pg_notification_store", nil
}
func (s *pgNotificationStore) GetName() string { return "pg_notification_store" }
func (s *pgNotificationStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgNotificationStore) Close() error { return nil }

// ── NotificationPrefStore pg ──────────────────────────────────────────────────

type pgNotificationPrefStore struct {
	exec execution.PGExecutor
}

// NewPGNotificationPrefStore cria uma instância de NotificationPrefStore para Postgres.
func NewPGNotificationPrefStore(exec execution.PGExecutor) NotificationPrefStore {
	return &pgNotificationPrefStore{exec: exec}
}

func (s *pgNotificationPrefStore) Upsert(ctx context.Context, input *UpsertNotificationPrefInput) (*NotificationPref, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	freq := input.DigestFrequency
	if freq == "" {
		freq = "realtime"
	}
	ch := input.Channel
	if ch == "" {
		ch = "in_app"
	}
	var p NotificationPref
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (user_id, tenant_id, event_type, channel, enabled, digest_frequency)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, tenant_id, event_type, channel) DO UPDATE
		SET enabled = EXCLUDED.enabled, digest_frequency = EXCLUDED.digest_frequency
		RETURNING id, user_id, tenant_id, event_type, channel, enabled, digest_frequency
	`, prefTable),
		input.UserID, input.TenantID, input.EventType, ch, input.Enabled, freq,
	).Scan(&p.ID, &p.UserID, &p.TenantID, &p.EventType, &p.Channel, &p.Enabled, &p.DigestFrequency)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert notification pref: %w", err)
	}
	return &p, nil
}

func (s *pgNotificationPrefStore) Get(ctx context.Context, userID, tenantID, eventType, channel string) (*NotificationPref, error) {
	var p NotificationPref
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, user_id, tenant_id, event_type, channel, enabled, digest_frequency
		FROM %s WHERE user_id = $1 AND tenant_id = $2 AND event_type = $3 AND channel = $4
	`, prefTable), userID, tenantID, eventType, channel).Scan(
		&p.ID, &p.UserID, &p.TenantID, &p.EventType, &p.Channel, &p.Enabled, &p.DigestFrequency)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification pref: %w", err)
	}
	return &p, nil
}

func (s *pgNotificationPrefStore) ListByUser(ctx context.Context, userID, tenantID string) ([]NotificationPref, error) {
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, tenant_id, event_type, channel, enabled, digest_frequency
		FROM %s WHERE user_id = $1 AND tenant_id = $2 ORDER BY event_type, channel
	`, prefTable), userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification prefs: %w", err)
	}
	defer rows.Close()
	var prefs []NotificationPref
	for rows.Next() {
		var p NotificationPref
		if err := rows.Scan(&p.ID, &p.UserID, &p.TenantID, &p.EventType, &p.Channel, &p.Enabled, &p.DigestFrequency); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}

func (s *pgNotificationPrefStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[NotificationPref](), "pg_notification_pref_store", nil
}
func (s *pgNotificationPrefStore) GetName() string { return "pg_notification_pref_store" }
func (s *pgNotificationPrefStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgNotificationPrefStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanNotif(row pgx.Row, n *Notification) error {
	return row.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Urgency,
		&n.Title, &n.Body, &n.ReferenceID, &n.ReferenceType, &n.ReadAt, &n.CreatedAt)
}

func buildNotifFilters(f *NotificationFilters) (string, []any) {
	var where []string
	var args []any
	if f.TenantID != "" {
		args = append(args, f.TenantID)
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)))
	}
	if f.UserID != "" {
		args = append(args, f.UserID)
		where = append(where, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if f.UnreadOnly {
		where = append(where, "read_at IS NULL")
	}
	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(p int) int {
	if p < 1 {
		return 1
	}
	return p
}

func normalizeLimit(l, def, max int) int {
	if l <= 0 {
		return def
	}
	if l > max {
		return max
	}
	return l
}

// suppress unused import
var _ = time.Now
