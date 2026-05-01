-- ============================================================================
-- ETAPA 18: Limalar — Pipeline CRM (Leads + Cards)
-- ============================================================================
-- Cria as tabelas do módulo CRM:
--   - limalar.leads          : contatos qualificados (pessoas interessadas)
--   - limalar.pipeline_cards : cards do kanban de vendas por lead/imóvel
--
-- Tenant isolation: ambas as tabelas usam tenant_id sem FK externa ao schema
-- public — schema limalar é autônomo.
--
-- Depende de: etapa_16 (limalar.properties deve existir para FKs opcionais)
-- ============================================================================
\echo 'ETAPA 18: Criando limalar.leads e limalar.pipeline_cards (CRM Pipeline)...'

-- ── Leads ────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.leads (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID        NOT NULL,
    name              TEXT        NOT NULL DEFAULT '',
    email             TEXT        NOT NULL DEFAULT '',
    phone             TEXT        NOT NULL DEFAULT '',
    property_id       UUID        REFERENCES limalar.properties(id) ON DELETE SET NULL,
    status            TEXT        NOT NULL DEFAULT 'new',
    -- 'new' | 'contacted' | 'qualified' | 'proposal' | 'won' | 'lost'
    score             INTEGER     NOT NULL DEFAULT 0,
    notes             TEXT        NOT NULL DEFAULT '',
    source            TEXT        NOT NULL DEFAULT 'manual',
    -- 'manual' | 'whatsapp' | 'url_capture' | 'olx' | 'zapimoveis'
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.leads criada'

-- phone é o identificador natural de leads via WhatsApp
CREATE UNIQUE INDEX IF NOT EXISTS idx_limalar_leads_phone_tenant
    ON limalar.leads (phone, tenant_id)
    WHERE phone != '';

CREATE INDEX IF NOT EXISTS idx_limalar_leads_tenant_status
    ON limalar.leads (tenant_id, status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_leads_tenant_score
    ON limalar.leads (tenant_id, score DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_leads_property
    ON limalar.leads (property_id) WHERE property_id IS NOT NULL;

CREATE OR REPLACE TRIGGER update_limalar_leads_updated_at
    BEFORE UPDATE ON limalar.leads
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Índices e trigger de limalar.leads criados'

-- ── Pipeline Cards ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.pipeline_cards (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    lead_id     UUID        REFERENCES limalar.leads(id) ON DELETE CASCADE,
    property_id UUID        REFERENCES limalar.properties(id) ON DELETE SET NULL,
    col         TEXT        NOT NULL DEFAULT 'new',
    -- 'new' | 'contacted' | 'proposal' | 'won' | 'lost'
    position    INTEGER     NOT NULL DEFAULT 0,
    title       TEXT        NOT NULL DEFAULT '',
    value       BIGINT      NOT NULL DEFAULT 0,   -- centavos
    score       INTEGER     NOT NULL DEFAULT 0,
    ai_tip      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.pipeline_cards criada'

CREATE INDEX IF NOT EXISTS idx_limalar_pipeline_tenant_col
    ON limalar.pipeline_cards (tenant_id, col, position ASC);

CREATE INDEX IF NOT EXISTS idx_limalar_pipeline_lead
    ON limalar.pipeline_cards (lead_id) WHERE lead_id IS NOT NULL;

CREATE OR REPLACE TRIGGER update_limalar_pipeline_cards_updated_at
    BEFORE UPDATE ON limalar.pipeline_cards
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Índices e trigger de limalar.pipeline_cards criados'

\echo 'ETAPA 18 concluída com sucesso!'
