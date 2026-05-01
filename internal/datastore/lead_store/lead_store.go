// Package leadstore fornece o store de leads do CRM (Colonial NetImóveis).
package leadstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// Lead representa um contato qualificado no CRM do corretor.
type Lead struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	PropertyID *string   `json:"property_id,omitempty"`
	Status     string    `json:"status"` // new|contacted|qualified|proposal|won|lost
	Score      int       `json:"score"`
	Notes      string    `json:"notes"`
	Source     string    `json:"source"` // manual|whatsapp|url_capture|olx|zapimoveis
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateLeadInput define dados para criação de um lead.
type CreateLeadInput struct {
	TenantID   string
	Name       string
	Email      string
	Phone      string
	PropertyID *string
	Status     string
	Score      int
	Notes      string
	Source     string
}

// UpdateLeadInput define campos atualizáveis de um lead.
type UpdateLeadInput struct {
	Name       *string
	Email      *string
	Phone      *string
	PropertyID *string
	Status     *string
	Score      *int
	Notes      *string
	Source     *string
}

// LeadFilters define filtros para listagem de leads.
type LeadFilters struct {
	TenantID   string
	PropertyID *string
	Status     *string
	Phone      *string
	Page       int
	Limit      int
}

// LeadStore define operações de persistência do CRM de leads.
type LeadStore interface {
	t.StoreType

	// Create registra um novo lead.
	Create(ctx context.Context, input *CreateLeadInput) (*Lead, error)

	// GetByID retorna um lead pelo ID, dentro do tenant.
	GetByID(ctx context.Context, tenantID, id string) (*Lead, error)

	// GetByPhone retorna um lead pelo telefone dentro do tenant.
	// Retorna nil, nil se não encontrado.
	GetByPhone(ctx context.Context, tenantID, phone string) (*Lead, error)

	// Update atualiza campos de um lead existente.
	Update(ctx context.Context, tenantID, id string, input *UpdateLeadInput) (*Lead, error)

	// UpdateStatus atualiza apenas o status de um lead.
	UpdateStatus(ctx context.Context, tenantID, id, status string) error

	// AddNote appenda uma nota ao campo notes do lead (prefixo com timestamp).
	AddNote(ctx context.Context, tenantID, id, note string) error

	// ListByTenant retorna leads paginados com filtros opcionais.
	ListByTenant(ctx context.Context, filters *LeadFilters) (*t.PaginatedResult[Lead], error)

	// ListByPropertyID retorna todos os leads vinculados a um imóvel.
	ListByPropertyID(ctx context.Context, tenantID, propertyID string) ([]Lead, error)
}
