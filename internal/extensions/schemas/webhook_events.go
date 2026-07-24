package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/extensions"
)

func init() {
	extensions.Global.Register(&WebhookEventsSchema{})
}

// WebhookEventsSchema é a tabela de auditoria/idempotência da engine de webhook
// do gnyx (internal/services/webhook). Registra cada evento recebido uma única
// vez (event_id é a chave) com seu ciclo de vida — recebido → processando →
// concluído/falhou — para não reprocessar um evento já concluído e para deixar
// rastro do que aconteceu com cada entrega.
type WebhookEventsSchema struct{}

func (s *WebhookEventsSchema) ID() string {
	return "webhook_events_schema"
}

func (s *WebhookEventsSchema) Dependencies() []string {
	return nil
}

func (s *WebhookEventsSchema) Statements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS webhook_events (
			event_id     TEXT        PRIMARY KEY,
			source       TEXT        NOT NULL DEFAULT '',
			event_type   TEXT        NOT NULL DEFAULT '',
			payload      JSONB       NOT NULL DEFAULT '{}'::jsonb,
			headers      JSONB       NOT NULL DEFAULT '{}'::jsonb,
			status       TEXT        NOT NULL,
			error        TEXT        NOT NULL DEFAULT '',
			received_at  TIMESTAMPTZ NOT NULL,
			processed_at TIMESTAMPTZ NULL,
			retry_count  INTEGER     NOT NULL DEFAULT 0
		)`,
		// Consultas operacionais típicas: "o que falhou?" e "o que chegou de tal
		// fonte, mais recente primeiro?".
		`CREATE INDEX IF NOT EXISTS idx_webhook_events_status
			ON webhook_events (status)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_events_source_received
			ON webhook_events (source, received_at DESC)`,
	}
}
