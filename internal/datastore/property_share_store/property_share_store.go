// Package propertysharesstore fornece o store de compartilhamentos de imóveis (Colonial NetImóveis).
package propertysharesstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// ShareRecord representa um evento de compartilhamento de imóvel.
type ShareRecord struct {
	ID             string    `json:"id"`
	PropertyID     string    `json:"property_id"`
	TenantID       string    `json:"tenant_id"`
	Channel        string    `json:"channel"` // "whatsapp" | "instagram" | "link"
	RecipientPhone string    `json:"recipient_phone,omitempty"`
	LinkToken      string    `json:"link_token,omitempty"`
	WappMessageID  string    `json:"wapp_message_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ShareSummary agrega contagens e eventos recentes de compartilhamento de um imóvel.
type ShareSummary struct {
	Total     int            `json:"total"`
	ByChannel map[string]int `json:"by_channel"`
	Recent    []ShareRecord  `json:"recent"`
}

// LogShareInput define dados para registro de um compartilhamento.
type LogShareInput struct {
	PropertyID     string
	TenantID       string
	Channel        string
	RecipientPhone string
	LinkToken      string
	WappMessageID  string
}

// PropertyShareStore define operações de persistência de compartilhamentos.
type PropertyShareStore interface {
	t.StoreType

	// LogShare registra um evento de compartilhamento e retorna o ID gerado.
	LogShare(ctx context.Context, input *LogShareInput) (string, error)

	// ListShares retorna os 20 compartilhamentos mais recentes de um imóvel com totais por canal.
	ListShares(ctx context.Context, tenantID, propertyID string) (*ShareSummary, error)

	// FindByPhone retorna o compartilhamento WhatsApp mais recente para um telefone.
	// Usado pelo LeadProcessor para correlacionar respostas a um imóvel específico.
	// Retorna (nil, nil) se não encontrado.
	FindByPhone(ctx context.Context, phone string) (*ShareRecord, error)
}
