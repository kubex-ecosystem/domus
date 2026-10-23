package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/extensions"
)

func init() {
	extensions.Global.Register(&GnyxWMSSchema{})
}

type GnyxWMSSchema struct{}

func (s *GnyxWMSSchema) ID() string {
	return "gnyx_wms_schema"
}

func (s *GnyxWMSSchema) Dependencies() []string {
	return nil
}

func (s *GnyxWMSSchema) Statements() []string {
	return []string{
		`CREATE SCHEMA IF NOT EXISTS gnyx_wms`,

		// API keys — one row per integration key issued to an external caller.
		// key_hash stores SHA-256(raw_key) so the raw key is never stored in clear.
		`CREATE TABLE IF NOT EXISTS gnyx_wms.api_keys (
			id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			key_hash       TEXT        NOT NULL UNIQUE,
			tenant_id      UUID,
			label          TEXT        NOT NULL DEFAULT '',
			rate_limit_rpm INTEGER     NOT NULL DEFAULT 60,
			enabled        BOOLEAN     NOT NULL DEFAULT TRUE,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_used_at   TIMESTAMPTZ,
			expires_at     TIMESTAMPTZ
		)`,

		// Per-request audit log — written asynchronously by the telemetry middleware.
		`CREATE TABLE IF NOT EXISTS gnyx_wms.request_logs (
			id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			api_key_id       UUID        REFERENCES gnyx_wms.api_keys(id) ON DELETE SET NULL,
			tenant_id        UUID,
			endpoint         TEXT        NOT NULL DEFAULT '',
			method           TEXT        NOT NULL DEFAULT '',
			status_code      INTEGER     NOT NULL DEFAULT 0,
			latency_ms       INTEGER     NOT NULL DEFAULT 0,
			error_message    TEXT        NOT NULL DEFAULT '',
			sankhya_cb_state TEXT        NOT NULL DEFAULT '',
			nexus_cb_state   TEXT        NOT NULL DEFAULT '',
			requested_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Aggregated hourly telemetry — upserted after each request.
		`CREATE TABLE IF NOT EXISTS gnyx_wms.telemetry_hourly (
			id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id      UUID,
			endpoint       TEXT        NOT NULL DEFAULT '',
			hour_bucket    TIMESTAMPTZ NOT NULL,
			total_requests INTEGER     NOT NULL DEFAULT 0,
			success_count  INTEGER     NOT NULL DEFAULT 0,
			error_count    INTEGER     NOT NULL DEFAULT 0,
			avg_latency_ms INTEGER     NOT NULL DEFAULT 0,
			max_latency_ms INTEGER     NOT NULL DEFAULT 0,
			UNIQUE (tenant_id, endpoint, hour_bucket)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_gnyx_wms_request_logs_tenant_time
			ON gnyx_wms.request_logs (tenant_id, requested_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_wms_request_logs_key_time
			ON gnyx_wms.request_logs (api_key_id, requested_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_wms_telemetry_tenant_hour
			ON gnyx_wms.telemetry_hourly (tenant_id, hour_bucket DESC)`,
	}
}
