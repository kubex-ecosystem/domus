package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/datastore/extensions"
)

func init() {
	extensions.Global.Register(&GnyxBISchema{})
}

type GnyxBISchema struct{}

func (s *GnyxBISchema) ID() string {
	return "gnyx_bi_schema"
}

func (s *GnyxBISchema) Dependencies() []string {
	return nil
}

func (s *GnyxBISchema) Statements() []string {
	return []string{
		`CREATE SCHEMA IF NOT EXISTS gnyx_bi`,
		`CREATE TABLE IF NOT EXISTS gnyx_bi.artifacts (
			id               UUID        PRIMARY KEY,
			tenant_id        UUID        NOT NULL,
			slug             TEXT        NOT NULL,
			title            TEXT        NOT NULL,
			description      TEXT        NOT NULL DEFAULT '',
			domain           TEXT        NOT NULL DEFAULT '',
			status           TEXT        NOT NULL DEFAULT 'draft',
			source_mode      TEXT        NOT NULL DEFAULT '',
			source_ref       JSONB       NOT NULL DEFAULT '{}'::jsonb,
			generation       JSONB       NOT NULL DEFAULT '{}'::jsonb,
			plan             JSONB       NOT NULL DEFAULT '{}'::jsonb,
			dashboard_schema JSONB       NOT NULL DEFAULT '{}'::jsonb,
			runtime_summary  JSONB       NOT NULL DEFAULT '{}'::jsonb,
			delivery         JSONB       NOT NULL DEFAULT '{}'::jsonb,
			tags             JSONB       NOT NULL DEFAULT '[]'::jsonb,
			created_by       UUID        NULL,
			updated_by       UUID        NULL,
			created_at       TIMESTAMPTZ NOT NULL,
			updated_at       TIMESTAMPTZ NOT NULL,
			UNIQUE (tenant_id, slug)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_bi_artifacts_tenant_updated
			ON gnyx_bi.artifacts (tenant_id, updated_at DESC)`,
		`CREATE TABLE IF NOT EXISTS gnyx_bi.export_jobs (
			id           UUID        PRIMARY KEY,
			artifact_id  UUID        NOT NULL,
			tenant_id    UUID        NOT NULL,
			status       TEXT        NOT NULL DEFAULT 'queued',
			profile      TEXT        NOT NULL DEFAULT 'functional',
			target_kind  TEXT        NOT NULL DEFAULT 'sankhya_jsp_bundle',
			export_input JSONB       NOT NULL DEFAULT '{}'::jsonb,
			output_path  TEXT        NOT NULL DEFAULT '',
			report_path  TEXT        NOT NULL DEFAULT '',
			error_message TEXT       NOT NULL DEFAULT '',
			created_by   UUID        NULL,
			created_at   TIMESTAMPTZ NOT NULL,
			started_at   TIMESTAMPTZ NULL,
			finished_at  TIMESTAMPTZ NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_bi_export_jobs_artifact_created
			ON gnyx_bi.export_jobs (artifact_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gnyx_bi_export_jobs_tenant_created
			ON gnyx_bi.export_jobs (tenant_id, created_at DESC)`,
	}
}
