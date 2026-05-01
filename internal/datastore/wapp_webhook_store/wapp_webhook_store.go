// Package wappwebhookstore fornece o store de eventos brutos de webhook WhatsApp (Meta Cloud API).
// Persiste na tabela pública whatsapp_messages e whatsapp_statuses, que armazenam
// os payloads JSON crus recebidos do webhook Meta — distinto de limalar.wapp_messages
// que armazena threads estruturadas para a UI.
package wappwebhookstore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// WebhookMessage representa um evento de mensagem recebida via webhook Meta.
type WebhookMessage struct {
	ID            string          `json:"id"`
	PhoneNumberID string          `json:"phone_number_id"`
	FromNumber    string          `json:"from_number"`
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

// WebhookStatus representa uma atualização de status de entrega (delivered, read, failed).
type WebhookStatus struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	RecipientID string          `json:"recipient_id"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SaveWebhookMessageInput define os dados necessários para salvar um evento de mensagem.
type SaveWebhookMessageInput struct {
	ID            string
	PhoneNumberID string
	FromNumber    string
	Type          string
	Payload       json.RawMessage
}

// SaveWebhookStatusInput define os dados necessários para salvar um evento de status.
type SaveWebhookStatusInput struct {
	ID          string
	Status      string
	RecipientID string
	Payload     json.RawMessage
}

// WappWebhookStore define as operações de persistência dos eventos brutos de webhook.
type WappWebhookStore interface {
	t.StoreType

	// SaveMessage persiste um evento de mensagem recebida via webhook.
	// ON CONFLICT (id) DO NOTHING — idempotente por design do webhook Meta.
	SaveMessage(ctx context.Context, input *SaveWebhookMessageInput) error

	// SaveStatus persiste uma atualização de status de entrega.
	// ON CONFLICT (id, status) DO NOTHING — idempotente.
	SaveStatus(ctx context.Context, input *SaveWebhookStatusInput) error
}
