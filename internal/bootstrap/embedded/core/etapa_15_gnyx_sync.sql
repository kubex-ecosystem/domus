-- ============================================================================
-- ETAPA 15: GNyx Sync Schema
-- ============================================================================
-- Cria o schema gnyx_sync com workspaces, conexões, query assets,
-- destinos, jobs e runs de sincronização de dados.
-- Schema isolado — tenant_id é referência lógica, sem FK para public.
-- ============================================================================
\echo 'ETAPA 15: Criando schema gnyx_sync (Data Synchronization)...'

CREATE SCHEMA IF NOT EXISTS gnyx_sync;
\echo '  Schema gnyx_sync criado'

-- Workspaces agrupam recursos de sincronização por tenant
CREATE TABLE IF NOT EXISTS gnyx_sync.workspaces (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, slug)
);
\echo '  Tabela gnyx_sync.workspaces criada'

-- Conexões com fontes de dados externas
CREATE TABLE IF NOT EXISTS gnyx_sync.source_connections (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela gnyx_sync.source_connections criada'

-- Query assets: queries parametrizadas associadas a uma fonte
CREATE TABLE IF NOT EXISTS gnyx_sync.query_assets (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela gnyx_sync.query_assets criada'

-- Destinos de entrega dos dados sincronizados
CREATE TABLE IF NOT EXISTS gnyx_sync.destination_targets (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela gnyx_sync.destination_targets criada'

-- Jobs de sincronização: orquestram source → query → destination
CREATE TABLE IF NOT EXISTS gnyx_sync.jobs (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela gnyx_sync.jobs criada'

-- Runs: histórico de execuções de cada job
CREATE TABLE IF NOT EXISTS gnyx_sync.runs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID        NOT NULL REFERENCES gnyx_sync.jobs(id) ON DELETE CASCADE,
    workspace_id    UUID        NOT NULL REFERENCES gnyx_sync.workspaces(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL,
    status          TEXT        NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ NULL,
    rows_read       BIGINT      NOT NULL DEFAULT 0,
    rows_written    BIGINT      NOT NULL DEFAULT 0,
    delivery_count  BIGINT      NOT NULL DEFAULT 0,
    summary         TEXT        NOT NULL DEFAULT '',
    error_message   TEXT        NOT NULL DEFAULT '',
    warning_message TEXT        NOT NULL DEFAULT ''
);
\echo '  Tabela gnyx_sync.runs criada'

-- Índices de performance
CREATE INDEX IF NOT EXISTS idx_gnyx_sync_workspaces_tenant
    ON gnyx_sync.workspaces (tenant_id);

CREATE INDEX IF NOT EXISTS idx_gnyx_sync_connections_workspace
    ON gnyx_sync.source_connections (workspace_id);

CREATE INDEX IF NOT EXISTS idx_gnyx_sync_query_assets_workspace
    ON gnyx_sync.query_assets (workspace_id);

CREATE INDEX IF NOT EXISTS idx_gnyx_sync_targets_workspace
    ON gnyx_sync.destination_targets (workspace_id);

CREATE INDEX IF NOT EXISTS idx_gnyx_sync_jobs_workspace
    ON gnyx_sync.jobs (workspace_id);

CREATE INDEX IF NOT EXISTS idx_gnyx_sync_runs_workspace
    ON gnyx_sync.runs (workspace_id, started_at DESC);
\echo '  Índices de gnyx_sync criados (6)'

\echo 'ETAPA 15 concluída com sucesso!'
