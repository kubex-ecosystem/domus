// Package leadeventstore fornece o store de eventos da timeline de leads (Colonial NetImóveis).
package leadeventstore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// LeadEvent representa uma entrada na timeline de atividades de um lead.
type LeadEvent struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	LeadID    string          `json:"lead_id"`
	Type      string          `json:"type"`     // created|status_changed|score_updated|note_added|whatsapp_message|ai_qualified|pipeline_moved
	Body      string          `json:"body"`     // texto legível por humano
	Metadata  json.RawMessage `json:"metadata"` // dados extras (old_status, new_status, score, etc.)
	CreatedAt time.Time       `json:"created_at"`
}

// CreateLeadEventInput define os dados para criar um evento.
type CreateLeadEventInput struct {
	TenantID string
	LeadID   string
	Type     string
	Body     string
	Metadata json.RawMessage // nil → {}
}

// LeadEventStore define operações de persistência da timeline de leads.
type LeadEventStore interface {
	t.StoreType

	// Create registra um novo evento na timeline do lead.
	Create(ctx context.Context, input *CreateLeadEventInput) (*LeadEvent, error)

	// ListByLead retorna todos os eventos de um lead, ordenados do mais recente ao mais antigo.
	ListByLead(ctx context.Context, tenantID, leadID string) ([]LeadEvent, error)
}
