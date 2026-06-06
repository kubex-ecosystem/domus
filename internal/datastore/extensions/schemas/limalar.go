package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/datastore/extensions"
)

func init() {
	extensions.Global.Register(&LimalarSchema{})
}

type LimalarSchema struct{}

func (s *LimalarSchema) ID() string {
	return "limalar_schema"
}

func (s *LimalarSchema) Dependencies() []string {
	return nil
}

func (s *LimalarSchema) Statements() []string {
	return []string{
		`CREATE SCHEMA IF NOT EXISTS limalar`,
		`CREATE TABLE IF NOT EXISTS limalar.properties (
			id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id     UUID        NOT NULL,
			type          TEXT        NOT NULL DEFAULT '',
			transaction   TEXT        NOT NULL DEFAULT 'venda',
			neighborhood  TEXT        NOT NULL DEFAULT '',
			address       TEXT        NOT NULL DEFAULT '',
			price         BIGINT      NOT NULL DEFAULT 0,
			area          INTEGER     NOT NULL DEFAULT 0,
			bedrooms      INTEGER     NOT NULL DEFAULT 0,
			suites        INTEGER     NOT NULL DEFAULT 0,
			bathrooms     INTEGER     NOT NULL DEFAULT 0,
			parking       INTEGER     NOT NULL DEFAULT 0,
			description   TEXT        NOT NULL DEFAULT '',
			highlights    TEXT        NOT NULL DEFAULT '',
			contact_phone TEXT        NOT NULL DEFAULT '',
			images        JSONB       NOT NULL DEFAULT '[]'::jsonb,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_limalar_props_tenant_updated
			ON limalar.properties (tenant_id, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_limalar_props_tenant_transaction
			ON limalar.properties (tenant_id, transaction)`,
		`CREATE TABLE IF NOT EXISTS limalar.property_shares (
			id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			property_id      UUID        NOT NULL REFERENCES limalar.properties(id) ON DELETE CASCADE,
			tenant_id        UUID        NOT NULL,
			channel          TEXT        NOT NULL DEFAULT 'link',
			recipient_phone  TEXT        NOT NULL DEFAULT '',
			link_token       TEXT        NOT NULL DEFAULT '',
			wapp_message_id  TEXT        NOT NULL DEFAULT '',
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_limalar_shares_property
			ON limalar.property_shares (property_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_limalar_shares_tenant
			ON limalar.property_shares (tenant_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_limalar_shares_token
			ON limalar.property_shares (link_token) WHERE link_token != ''`,
	}
}
