-- ============================================================================
-- ETAPA 9: Auth Sessions (refresh tokens)
-- ============================================================================
-- Cria tabela de sessões para controle de refresh tokens nativos do BE
-- ============================================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
-- Garante que estamos no schema correto
SET search_path TO public;
\ echo 'ETAPA 9: Criando tabela auth_sessions...' CREATE TABLE IF NOT EXISTS auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  refresh_token_hash TEXT NOT NULL,
  user_agent TEXT,
  ip TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_refresh ON auth_sessions(refresh_token_hash);
\ echo 'Tabela auth_sessions criada'