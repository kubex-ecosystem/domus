-- ============================================================================
-- ETAPA 16: Limalar — Schema de Domínio (Colonial NetImóveis)
-- ============================================================================
-- Cria o schema limalar e as tabelas principais do domínio imobiliário:
--   - properties          : cadastro de imóveis
--   - property_shares     : rastreamento de compartilhamentos
--   - lead_conversations  : qualificação de leads por IA por imóvel
--
-- Tenant isolation: todas as tabelas possuem tenant_id como coluna de
-- particionamento lógico (sem FK para public.tenant — schema autônomo).
-- ============================================================================
\echo 'ETAPA 16: Criando schema limalar (Colonial NetImóveis — domínio imobiliário)...'

CREATE SCHEMA IF NOT EXISTS limalar;
\echo '  Schema limalar criado'

-- ── Imóveis ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.properties (
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
);
\echo '  Tabela limalar.properties criada'

CREATE INDEX IF NOT EXISTS idx_limalar_props_tenant_updated
    ON limalar.properties (tenant_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_props_tenant_transaction
    ON limalar.properties (tenant_id, transaction);
\echo '  Índices de limalar.properties criados'

-- Trigger de updated_at — reutiliza a função definida em public (etapa 7)
CREATE OR REPLACE TRIGGER update_limalar_properties_updated_at
    BEFORE UPDATE ON limalar.properties
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Trigger updated_at de limalar.properties criado'

-- ── Compartilhamentos de imóveis ──────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.property_shares (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id     UUID        NOT NULL REFERENCES limalar.properties(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL,
    channel         TEXT        NOT NULL DEFAULT 'link',
    recipient_phone TEXT        NOT NULL DEFAULT '',
    link_token      TEXT        NOT NULL DEFAULT '',
    wapp_message_id TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.property_shares criada'

CREATE INDEX IF NOT EXISTS idx_limalar_shares_property
    ON limalar.property_shares (property_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_shares_tenant
    ON limalar.property_shares (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_shares_token
    ON limalar.property_shares (link_token) WHERE link_token != '';

CREATE INDEX IF NOT EXISTS idx_limalar_shares_phone
    ON limalar.property_shares (recipient_phone) WHERE recipient_phone != '';
\echo '  Índices de limalar.property_shares criados'

-- ── Conversas de qualificação de leads (por imóvel) ───────────────────────────

CREATE TABLE IF NOT EXISTS limalar.lead_conversations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id     UUID        NOT NULL REFERENCES limalar.properties(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL,
    phone           TEXT        NOT NULL,
    inbound_message TEXT        NOT NULL,
    ai_reply        TEXT        NOT NULL DEFAULT '',
    score           INT         NOT NULL DEFAULT 0,
    status          TEXT        NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.lead_conversations criada'

CREATE INDEX IF NOT EXISTS idx_limalar_lead_conv_property
    ON limalar.lead_conversations (property_id, tenant_id);

CREATE INDEX IF NOT EXISTS idx_limalar_lead_conv_phone
    ON limalar.lead_conversations (phone);

CREATE INDEX IF NOT EXISTS idx_limalar_lead_conv_tenant_created
    ON limalar.lead_conversations (tenant_id, created_at DESC);
\echo '  Índices de limalar.lead_conversations criados'

\echo 'ETAPA 16 concluída com sucesso!'
