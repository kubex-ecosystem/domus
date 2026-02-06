-- ============================================================================
-- ETAPA 10: Pending Access Requests
-- ============================================================================
-- Cria tabela para solicitações de acesso pendentes (ex: OAuth)
-- ============================================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
-- Garante que estamos no schema correto
SET search_path TO public;
\ echo 'ETAPA 10: Criando pending_access_requests...' CREATE TABLE IF NOT EXISTS pending_access_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL,
    provider TEXT NOT NULL,
    provider_user_id TEXT,
    name TEXT,
    avatar_url TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    requester_ip TEXT,
    requester_user_agent TEXT,
    tenant_id UUID,
    role_code TEXT,
    metadata JSONB,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
ALTER TABLE pending_access_requests
ADD CONSTRAINT pending_access_requests_unique UNIQUE (email, provider, status);
CREATE INDEX IF NOT EXISTS idx_pending_access_requests_status ON pending_access_requests (status);
CREATE INDEX IF NOT EXISTS idx_pending_access_requests_created_at ON pending_access_requests (created_at DESC);
\ echo '  Tabela pending_access_requests criada' \ echo 'ETAPA 10 concluída com sucesso!'