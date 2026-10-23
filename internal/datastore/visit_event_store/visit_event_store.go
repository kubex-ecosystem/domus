// Package visiteventstore fornece o store de agenda de visitas (Colonial NetImóveis).
package visiteventstore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// VisitEvent representa uma visita agendada pelo corretor.
type VisitEvent struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	AgentID     string    `json:"agent_id"`
	LeadID      *string   `json:"lead_id,omitempty"`
	PropertyID  *string   `json:"property_id,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"` // scheduled|confirmed|completed|cancelled|no_show
	Notes       string    `json:"notes"`
	Address     string    `json:"address"` // desnormalizado para export Maps/Waze
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateVisitEventInput define dados para criar uma visita.
type CreateVisitEventInput struct {
	TenantID    string
	AgentID     string
	LeadID      *string
	PropertyID  *string
	ScheduledAt time.Time
	Status      string
	Notes       string
	Address     string
}

// VisitEventFilters define filtros para listagem de visitas.
type VisitEventFilters struct {
	TenantID string
	AgentID  *string
	LeadID   *string
	Status   *string
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

// VisitEventStore define operações de persistência da agenda de visitas.
type VisitEventStore interface {
	t.StoreType

	// Create registra uma nova visita na agenda.
	Create(ctx context.Context, input *CreateVisitEventInput) (*VisitEvent, error)

	// GetByID retorna uma visita pelo ID, dentro do tenant.
	GetByID(ctx context.Context, tenantID, id string) (*VisitEvent, error)

	// List retorna visitas paginadas com filtros opcionais.
	List(ctx context.Context, filters *VisitEventFilters) (*t.PaginatedResult[VisitEvent], error)

	// UpdateStatus atualiza o status de uma visita.
	UpdateStatus(ctx context.Context, tenantID, id, status string) error

	// ListByDay retorna todas as visitas de um corretor em um determinado dia,
	// ordenadas por scheduled_at ASC.
	ListByDay(ctx context.Context, tenantID, agentID string, day time.Time) ([]VisitEvent, error)
}
