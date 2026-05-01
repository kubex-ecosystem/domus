// Package pipelinecardstore fornece o store de cards do kanban de pipeline CRM (Colonial NetImóveis).
package pipelinecardstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// PipelineCard representa um card no kanban de vendas do corretor.
type PipelineCard struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	LeadID     *string   `json:"lead_id,omitempty"`
	PropertyID *string   `json:"property_id,omitempty"`
	Col        string    `json:"col"`      // 'new'|'contacted'|'proposal'|'won'|'lost'
	Position   int       `json:"position"` // ordem dentro da coluna
	Title      string    `json:"title"`
	Value      int64     `json:"value"` // centavos
	Score      int       `json:"score"`
	AiTip      string    `json:"ai_tip"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreatePipelineCardInput define dados para criação de um card.
type CreatePipelineCardInput struct {
	TenantID   string
	LeadID     *string
	PropertyID *string
	Col        string
	Position   int
	Title      string
	Value      int64
	Score      int
	AiTip      string
}

// PipelineCardFilters define filtros para listagem de cards.
type PipelineCardFilters struct {
	TenantID string
	Col      *string
	LeadID   *string
}

// PipelineCardStore define operações de persistência do kanban de pipeline.
type PipelineCardStore interface {
	t.StoreType

	// Create registra um novo card no pipeline.
	Create(ctx context.Context, input *CreatePipelineCardInput) (*PipelineCard, error)

	// UpdateColumn move um card para outra coluna e ajusta a posição.
	UpdateColumn(ctx context.Context, tenantID, id, col string, position int) error

	// UpdatePosition reordena um card dentro da mesma coluna.
	UpdatePosition(ctx context.Context, tenantID, id string, position int) error

	// Delete remove um card do pipeline.
	Delete(ctx context.Context, tenantID, id string) error

	// ListByTenant retorna todos os cards de um tenant, opcionalmente filtrados.
	ListByTenant(ctx context.Context, filters *PipelineCardFilters) ([]PipelineCard, error)
}
