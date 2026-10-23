// Package wappmessagestore fornece o store de mensagens WhatsApp threaded (Colonial NetImóveis).
package wappmessagestore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// WappMessage representa uma mensagem individual em uma thread WhatsApp.
type WappMessage struct {
	ID         string    `json:"id"`
	PropertyID string    `json:"property_id,omitempty"`
	TenantID   string    `json:"tenant_id"`
	Phone      string    `json:"phone"`
	Direction  string    `json:"direction"` // "inbound" | "outbound"
	Role       string    `json:"role"`      // "lead" | "ai" | "human"
	Content    string    `json:"content"`
	Score      int       `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}

// ThreadSummary representa o estado mais recente de uma conversa com um lead.
type ThreadSummary struct {
	Phone                string    `json:"phone"`
	Score                int       `json:"score"`
	LastRole             string    `json:"last_role"`
	LastMessage          string    `json:"last_message"`
	LastMessageAt        time.Time `json:"last_message_at"`
	MessageCount         int       `json:"message_count"`
	PropertyID           string    `json:"property_id,omitempty"`
	PropertyType         string    `json:"property_type,omitempty"`
	PropertyNeighborhood string    `json:"property_neighborhood,omitempty"`
}

// SaveMessageInput define dados para salvar uma mensagem.
type SaveMessageInput struct {
	TenantID   string
	Phone      string
	PropertyID string // pode ser vazio
	Direction  string // "inbound" | "outbound"
	Role       string // "lead" | "ai" | "human"
	Content    string
	Score      int
}

// WappMessageStore define operações de persistência de mensagens WhatsApp.
type WappMessageStore interface {
	t.StoreType

	// Save insere um registro de mensagem.
	Save(ctx context.Context, input *SaveMessageInput) error

	// ListByPhone retorna até 100 mensagens de um telefone, ordenadas por data ASC.
	ListByPhone(ctx context.Context, tenantID, phone string) ([]WappMessage, error)

	// ListThreads retorna um ThreadSummary por número único, ordenado por atividade mais recente.
	ListThreads(ctx context.Context, tenantID string) ([]ThreadSummary, error)
}
