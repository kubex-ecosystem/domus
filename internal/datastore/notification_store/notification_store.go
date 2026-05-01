// Package notificationstore fornece os stores de notificações (Colonial NetImóveis).
// Dois stores coexistem no mesmo pacote por serem domínio estreitamente acoplado:
//   - NotificationStore     : CRUD de notificações + mark-read
//   - NotificationPrefStore : preferências de canal/evento por usuário (mute-by-default)
package notificationstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// ── Notification ──────────────────────────────────────────────────────────────

// Notification representa uma notificação gerada pelo sistema para um usuário.
type Notification struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	UserID        string     `json:"user_id"`
	Type          string     `json:"type"`    // lead_qualified|whatsapp_reply|pipeline_moved|commission_received|score_updated|system
	Urgency       string     `json:"urgency"` // low|normal|high
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	ReferenceID   *string    `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateNotificationInput define dados para criar uma notificação.
type CreateNotificationInput struct {
	TenantID      string
	UserID        string
	Type          string
	Urgency       string
	Title         string
	Body          string
	ReferenceID   *string
	ReferenceType *string
}

// NotificationFilters define filtros para listagem de notificações.
type NotificationFilters struct {
	TenantID   string
	UserID     string
	UnreadOnly bool
	Page       int
	Limit      int
}

// NotificationStore define operações de persistência de notificações.
type NotificationStore interface {
	t.StoreType

	// Create registra uma nova notificação.
	Create(ctx context.Context, input *CreateNotificationInput) (*Notification, error)

	// ListByUser retorna notificações paginadas de um usuário.
	ListByUser(ctx context.Context, filters *NotificationFilters) (*t.PaginatedResult[Notification], error)

	// MarkRead marca uma notificação específica como lida.
	MarkRead(ctx context.Context, tenantID, userID, id string) error

	// MarkAllRead marca todas as notificações não lidas do usuário como lidas.
	MarkAllRead(ctx context.Context, tenantID, userID string) error

	// UnreadCount retorna o número de notificações não lidas do usuário.
	UnreadCount(ctx context.Context, tenantID, userID string) (int64, error)
}

// ── NotificationPref ──────────────────────────────────────────────────────────

// NotificationPref representa a preferência de notificação de um usuário.
// enabled = false por padrão — Calm UX: mute-by-default.
type NotificationPref struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	TenantID        string `json:"tenant_id"`
	EventType       string `json:"event_type"`
	Channel         string `json:"channel"`          // in_app|email|whatsapp
	Enabled         bool   `json:"enabled"`          // false = mute
	DigestFrequency string `json:"digest_frequency"` // realtime|daily|weekly|never
}

// UpsertNotificationPrefInput define dados para inserir ou atualizar uma preferência.
type UpsertNotificationPrefInput struct {
	UserID          string
	TenantID        string
	EventType       string
	Channel         string
	Enabled         bool
	DigestFrequency string
}

// NotificationPrefStore define operações de persistência de preferências de notificação.
type NotificationPrefStore interface {
	t.StoreType

	// Upsert insere ou atualiza a preferência de notificação.
	Upsert(ctx context.Context, input *UpsertNotificationPrefInput) (*NotificationPref, error)

	// Get retorna a preferência para um evento/canal específico.
	Get(ctx context.Context, userID, tenantID, eventType, channel string) (*NotificationPref, error)

	// ListByUser retorna todas as preferências de um usuário em um tenant.
	ListByUser(ctx context.Context, userID, tenantID string) ([]NotificationPref, error)
}
