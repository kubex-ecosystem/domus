-- ============================================================================
-- ETAPA 12: External Metadata Registry
-- ============================================================================
-- Cria a camada genérica de governança para metadados externos carregados
-- no PostgreSQL usado pelo Domus, sem acoplar datasets específicos ao core.
-- ============================================================================
SET search_path TO public;
\echo 'ETAPA 12: Criando external metadata registry...'

CREATE TABLE IF NOT EXISTS external_metadata_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_system TEXT NOT NULL,
    domain TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    dataset_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    manifest JSONB NOT NULL DEFAULT '{}'::jsonb,
    row_count BIGINT,
    last_loaded_at TIMESTAMPTZ,
    load_mode TEXT NOT NULL DEFAULT 'full_refresh',
    status TEXT NOT NULL DEFAULT 'ready',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    UNIQUE (source_system, domain, schema_name, dataset_name)
);
\echo '  Tabela external_metadata_registry criada'

CREATE INDEX IF NOT EXISTS idx_external_metadata_registry_source_domain
    ON external_metadata_registry(source_system, domain);
CREATE INDEX IF NOT EXISTS idx_external_metadata_registry_schema_dataset
    ON external_metadata_registry(schema_name, dataset_name);
CREATE INDEX IF NOT EXISTS idx_external_metadata_registry_status
    ON external_metadata_registry(status);
CREATE INDEX IF NOT EXISTS idx_external_metadata_registry_last_loaded
    ON external_metadata_registry(last_loaded_at DESC);
\echo '  Índices da external_metadata_registry criados'

CREATE TRIGGER update_external_metadata_registry_updated_at
    BEFORE UPDATE ON external_metadata_registry
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
\echo '  Trigger da external_metadata_registry criado'

\echo 'ETAPA 12 concluída com sucesso!'
