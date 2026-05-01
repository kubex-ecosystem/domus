// Package leadconversationstore fornece o store de conversas de qualificação de leads por IA (Colonial NetImóveis).
package leadconversationstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// LeadConversation representa um turno de conversa de qualificação por IA.
type LeadConversation struct {
	ID             string    `json:"id"`
	PropertyID     string    `json:"property_id"`
	TenantID       string    `json:"tenant_id"`
	Phone          string    `json:"phone"`
	InboundMessage string    `json:"inbound_message"`
	AIReply        string    `json:"ai_reply"`
	Score          int       `json:"score"`
	Status         string    `json:"status"` // "pending" | "replied" | "escalated"
	CreatedAt      time.Time `json:"created_at"`
}

// CreateConversationInput define dados para criação de um turno.
type CreateConversationInput struct {
	PropertyID     string
	TenantID       string
	Phone          string
	InboundMessage string
	AIReply        string
	Score          int
	Status         string
}

// ConversationFilters define filtros para listagem de conversas.
type ConversationFilters struct {
	TenantID   string
	PropertyID *string
	Phone      *string
	Status     *string
	Page       int
	Limit      int
}

// LeadConversationStore define operações de persistência de conversas de lead.
type LeadConversationStore interface {
	t.StoreType

	// Create registra um novo turno de qualificação.
	Create(ctx context.Context, input *CreateConversationInput) (*LeadConversation, error)

	// List retorna conversas paginadas com filtros opcionais.
	List(ctx context.Context, filters *ConversationFilters) (*t.PaginatedResult[LeadConversation], error)

	// ListByProperty retorna todas as conversas de um imóvel para um tenant.
	ListByProperty(ctx context.Context, tenantID, propertyID string) ([]LeadConversation, error)

	// UpdateStatus atualiza o status de uma conversa.
	UpdateStatus(ctx context.Context, tenantID, id, status string) error
}
