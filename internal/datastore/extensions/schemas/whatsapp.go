package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/datastore/extensions"
)

func init() {
	extensions.Global.Register(&WhatsappSchema{})
}

type WhatsappSchema struct{}

func (s *WhatsappSchema) ID() string {
	return "whatsapp_schema"
}

func (s *WhatsappSchema) Dependencies() []string {
	return nil
}

func (s *WhatsappSchema) Statements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS whatsapp_messages (
			id              TEXT        PRIMARY KEY,
			phone_number_id TEXT        NOT NULL DEFAULT '',
			from_number     TEXT        NOT NULL DEFAULT '',
			type            TEXT        NOT NULL DEFAULT '',
			payload         JSONB       NOT NULL DEFAULT '{}'::jsonb,
			created_at      TIMESTAMPTZ NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_phone_created
			ON whatsapp_messages (phone_number_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS whatsapp_statuses (
			id           TEXT        NOT NULL,
			status       TEXT        NOT NULL,
			recipient_id TEXT        NOT NULL DEFAULT '',
			payload      JSONB       NOT NULL DEFAULT '{}'::jsonb,
			created_at   TIMESTAMPTZ NOT NULL,
			PRIMARY KEY (id, status)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_whatsapp_statuses_recipient
			ON whatsapp_statuses (recipient_id, created_at DESC)`,
	}
}
