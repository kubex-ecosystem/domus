package wappwebhookstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/faelmori/domus-pulse/internal/execution"
)

// pgWappWebhookStore implementa WappWebhookStore usando PGExecutor.
type pgWappWebhookStore struct {
	exec execution.PGExecutor
}

// NewPGWappWebhookStore cria uma instância de WappWebhookStore para Postgres.
func NewPGWappWebhookStore(exec execution.PGExecutor) WappWebhookStore {
	return &pgWappWebhookStore{exec: exec}
}

// SaveMessage persiste um evento de mensagem cru na tabela public.whatsapp_messages.
func (s *pgWappWebhookStore) SaveMessage(ctx context.Context, input *SaveWebhookMessageInput) error {
	if input == nil || input.ID == "" {
		return fmt.Errorf("SaveWebhookMessageInput com ID é obrigatório")
	}

	payload := input.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}

	const q = `
		INSERT INTO whatsapp_messages (id, phone_number_id, from_number, type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := s.exec.Exec(ctx, q,
		input.ID,
		input.PhoneNumberID,
		input.FromNumber,
		input.Type,
		payload,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("falha ao salvar webhook message %s: %w", input.ID, err)
	}
	return nil
}

// SaveStatus persiste uma atualização de status na tabela public.whatsapp_statuses.
func (s *pgWappWebhookStore) SaveStatus(ctx context.Context, input *SaveWebhookStatusInput) error {
	if input == nil || input.ID == "" || input.Status == "" {
		return fmt.Errorf("SaveWebhookStatusInput com ID e Status são obrigatórios")
	}

	payload := input.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}

	const q = `
		INSERT INTO whatsapp_statuses (id, status, recipient_id, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id, status) DO NOTHING
	`
	_, err := s.exec.Exec(ctx, q,
		input.ID,
		input.Status,
		input.RecipientID,
		payload,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("falha ao salvar webhook status %s/%s: %w", input.ID, input.Status, err)
	}
	return nil
}

// GetType implementa StoreType.
func (s *pgWappWebhookStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[WebhookMessage](), "pg_wapp_webhook_store", nil
}

// GetName implementa StoreType.
func (s *pgWappWebhookStore) GetName() string { return "pg_wapp_webhook_store" }

// Validate implementa StoreType.
func (s *pgWappWebhookStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

// Close implementa StoreType.
func (s *pgWappWebhookStore) Close() error { return nil }
