-- ============================================================================
-- ETAPA 14: GNyx BI Schema
-- ============================================================================
-- Cria o schema gnyx_bi e suas tabelas de artefatos e jobs de exportação.
-- Schema isolado — sem FKs para public, tenant_id é referência lógica apenas.
-- ============================================================================
\echo 'ETAPA 14: Criando schema gnyx_bi (Business Intelligence)...'

CREATE SCHEMA IF NOT EXISTS gnyx_bi;
\echo '  Schema gnyx_bi criado'

-- Artefatos BI: dashboards, relatórios, análises geradas pelo GNyx
CREATE TABLE IF NOT EXISTS gnyx_bi.artifacts (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, slug)
);
\echo '  Tabela gnyx_bi.artifacts criada'

CREATE INDEX IF NOT EXISTS idx_gnyx_bi_artifacts_tenant_updated
    ON gnyx_bi.artifacts (tenant_id, updated_at DESC);
\echo '  Índice idx_gnyx_bi_artifacts_tenant_updated criado'

-- Jobs de exportação assíncronos dos artefatos
CREATE TABLE IF NOT EXISTS gnyx_bi.export_jobs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id   UUID        NOT NULL REFERENCES gnyx_bi.artifacts(id) ON DELETE CASCADE,
    tenant_id     UUID        NOT NULL,
    status        TEXT        NOT NULL DEFAULT 'queued',
    profile       TEXT        NOT NULL DEFAULT 'functional',
    target_kind   TEXT        NOT NULL DEFAULT 'sankhya_jsp_bundle',
    export_input  JSONB       NOT NULL DEFAULT '{}'::jsonb,
    output_path   TEXT        NOT NULL DEFAULT '',
    report_path   TEXT        NOT NULL DEFAULT '',
    error_message TEXT        NOT NULL DEFAULT '',
    created_by    UUID        NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at    TIMESTAMPTZ NULL,
    finished_at   TIMESTAMPTZ NULL
);
\echo '  Tabela gnyx_bi.export_jobs criada'

CREATE INDEX IF NOT EXISTS idx_gnyx_bi_export_jobs_artifact_created
    ON gnyx_bi.export_jobs (artifact_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_gnyx_bi_export_jobs_tenant_created
    ON gnyx_bi.export_jobs (tenant_id, created_at DESC);
\echo '  Índices de gnyx_bi.export_jobs criados'

\echo 'ETAPA 14 concluída com sucesso!'
