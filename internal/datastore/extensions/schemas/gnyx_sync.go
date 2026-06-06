package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/datastore/extensions"
)

func init() {
	extensions.Global.Register(&GnyxSyncSchema{})
}

type GnyxSyncSchema struct{}

func (s *GnyxSyncSchema) ID() string {
	return "gnyx_sync_schema"
}

func (s *GnyxSyncSchema) Dependencies() []string {
	return nil
}

func (s *GnyxSyncSchema) Statements() []string {
	return []string{
		`CREATE SCHEMA IF NOT EXISTS gnyx_sync`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.workspaces (
			id          UUID        PRIMARY KEY,
			tenant_id   UUID        NOT NULL,
			name        TEXT        NOT NULL,
			slug        TEXT        NOT NULL,
			description TEXT        NOT NULL DEFAULT '',
			is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
			created_at  TIMESTAMPTZ NOT NULL,
			updated_at  TIMESTAMPTZ NOT NULL,
			UNIQUE (tenant_id, slug)
		)`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.source_connections (
			id                     UUID        PRIMARY KEY,
			tenant_id              UUID        NOT NULL,
			workspace_id           UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
			name                   TEXT        NOT NULL,
			engine_type            TEXT        NOT NULL,
			host                   TEXT        NOT NULL,
			port                   INTEGER     NOT NULL,
			database_name          TEXT        NOT NULL DEFAULT '',
			username               TEXT        NOT NULL DEFAULT '',
			secret_ref             TEXT        NOT NULL DEFAULT '',
			timeout_ms             INTEGER     NOT NULL DEFAULT 30000,
			is_active              BOOLEAN     NOT NULL DEFAULT TRUE,
			last_validation_status TEXT        NOT NULL DEFAULT '',
			last_validated_at      TIMESTAMPTZ NULL,
			settings               JSONB       NOT NULL DEFAULT '{}'::jsonb,
			created_at             TIMESTAMPTZ NOT NULL,
			updated_at             TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.query_assets (
			id                   UUID        PRIMARY KEY,
			tenant_id            UUID        NOT NULL,
			workspace_id         UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
			source_connection_id UUID        NOT NULL REFERENCES gnyx_sync.source_connections(id) ON DELETE CASCADE,
			name                 TEXT        NOT NULL,
			dialect              TEXT        NOT NULL,
			query_text           TEXT        NOT NULL,
			parameter_schema     JSONB       NOT NULL DEFAULT '{}'::jsonb,
			is_active            BOOLEAN     NOT NULL DEFAULT TRUE,
			last_preview_status  TEXT        NOT NULL DEFAULT '',
			last_preview_at      TIMESTAMPTZ NULL,
			created_at           TIMESTAMPTZ NOT NULL,
			updated_at           TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.destination_targets (
			id                     UUID        PRIMARY KEY,
			tenant_id              UUID        NOT NULL,
			workspace_id           UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
			name                   TEXT        NOT NULL,
			target_type            TEXT        NOT NULL,
			adapter_key            TEXT        NOT NULL,
			config                 JSONB       NOT NULL DEFAULT '{}'::jsonb,
			secret_ref             TEXT        NOT NULL DEFAULT '',
			is_active              BOOLEAN     NOT NULL DEFAULT TRUE,
			last_validation_status TEXT        NOT NULL DEFAULT '',
			last_validated_at      TIMESTAMPTZ NULL,
			created_at             TIMESTAMPTZ NOT NULL,
			updated_at             TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.jobs (
			id                    UUID        PRIMARY KEY,
			tenant_id             UUID        NOT NULL,
			workspace_id          UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
			source_connection_id  UUID        NOT NULL REFERENCES gnyx_sync.source_connections(id) ON DELETE RESTRICT,
			query_asset_id        UUID        NOT NULL REFERENCES gnyx_sync.query_assets(id) ON DELETE RESTRICT,
			destination_target_id UUID        NOT NULL REFERENCES gnyx_sync.destination_targets(id) ON DELETE RESTRICT,
			name                  TEXT        NOT NULL,
			strategy_key          TEXT        NOT NULL,
			trigger_mode          TEXT        NOT NULL,
			cron_expression       TEXT        NOT NULL DEFAULT '',
			is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
			last_run_status       TEXT        NOT NULL DEFAULT '',
			last_run_at           TIMESTAMPTZ NULL,
			last_success_at       TIMESTAMPTZ NULL,
			last_failure_at       TIMESTAMPTZ NULL,
			created_at            TIMESTAMPTZ NOT NULL,
			updated_at            TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gnyx_sync.runs (
			id              UUID        PRIMARY KEY,
			job_id          UUID        NOT NULL REFERENCES gnyx_sync.jobs(id) ON DELETE CASCADE,
			workspace_id    UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
			tenant_id       UUID        NOT NULL,
			status          TEXT        NOT NULL,
			started_at      TIMESTAMPTZ NOT NULL,
			finished_at     TIMESTAMPTZ NULL,
			rows_read       BIGINT      NOT NULL DEFAULT 0,
			rows_written    BIGINT      NOT NULL DEFAULT 0,
			delivery_count  BIGINT      NOT NULL DEFAULT 0,
			summary         TEXT        NOT NULL DEFAULT '',
			error_message   TEXT        NOT NULL DEFAULT '',
			warning_message TEXT        NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_workspaces_tenant
			ON gnyx_sync.workspaces (tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_connections_workspace
			ON gnyx_sync.source_connections (workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_query_assets_workspace
			ON gnyx_sync.query_assets (workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_targets_workspace
			ON gnyx_sync.destination_targets (workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_jobs_workspace
			ON gnyx_sync.jobs (workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_sync_runs_workspace
			ON gnyx_sync.runs (workspace_id, started_at DESC)`,
	}
}
