-- ============================================================================
-- ETAPA 11: Integration Engine (O Motor Kubex)
-- ============================================================================
-- Cria as tabelas do motor de integração vinculadas ao tenant
-- ============================================================================
SET search_path TO public;
\echo 'ETAPA 11: Criando Integration Engine...'
-- 1. Integration Configs (O Cofre das Conexões: MSSQL, APIs, IMAP)
CREATE TABLE IF NOT EXISTS integration_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    partner_id UUID REFERENCES partner(id) ON DELETE SET NULL, -- Linka com a sua tabela partner existente!
    type TEXT NOT NULL, -- Ex: 'MSSQL_ERP', 'REST_API', 'IMAP'
    name TEXT NOT NULL,
    settings JSONB NOT NULL, -- IPs, portas, chaves SSH, tudo dinâmico aqui
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
\echo '  Tabela integration_config criada'
-- 2. Sync Jobs (A Agenda do GNyx)
CREATE TABLE IF NOT EXISTS sync_job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    config_id UUID NOT NULL REFERENCES integration_config(id) ON DELETE CASCADE,
    task_name TEXT NOT NULL, -- Ex: 'SyncNewOrdersFastChannel'
    cron_expression TEXT,
    is_active BOOLEAN DEFAULT true,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
\echo '  Tabela sync_job criada'
-- 3. Queries (O Arsenal de SQL do Cliente)
CREATE TABLE IF NOT EXISTS integration_query (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    name TEXT NOT NULL, -- Ex: 'CheckClientERP'
    sql_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
\echo '  Tabela integration_query criada'
-- 4. Triggers e Índices (Seguindo o seu padrão)
CREATE INDEX IF NOT EXISTS idx_sync_job_tenant ON sync_job(tenant_id);
CREATE INDEX IF NOT EXISTS idx_integration_config_type ON integration_config(type);
CREATE INDEX IF NOT EXISTS idx_integration_config_created_at ON integration_config(created_at);
CREATE INDEX IF NOT EXISTS idx_integration_config_tenant ON integration_config(tenant_id);
CREATE INDEX IF NOT EXISTS idx_integration_query_tenant ON integration_query(tenant_id);
\echo '  Índices criados para tenant_id' -- Triggers para atualizar updated_at automaticamente
CREATE TRIGGER update_integration_config_updated_at BEFORE UPDATE ON integration_config FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sync_job_updated_at BEFORE UPDATE ON sync_job FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_integration_query_updated_at BEFORE UPDATE ON integration_query FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
\echo 'ETAPA 11 concluída com sucesso!'