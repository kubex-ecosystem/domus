-- ============================================================================
-- ETAPA 21: Limalar — Agenda de Visitas
-- ============================================================================
-- Cria a tabela limalar.visit_events: agenda de visitas do corretor.
--
-- Tenant isolation: tenant_id sem FK externa ao schema public.
-- agent_id identifica o corretor responsável (sem FK — schema limalar autônomo).
--
-- Depende de: etapa_16 (limalar.properties), etapa_18 (limalar.leads)
-- ============================================================================
\echo 'ETAPA 21: Criando limalar.visit_events (Agenda de Visitas)...'

CREATE TABLE IF NOT EXISTS limalar.visit_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID        NOT NULL,
    agent_id     UUID        NOT NULL,
    -- corretor responsável (user_id, sem FK para manter schema autônomo)
    lead_id      UUID        REFERENCES limalar.leads(id) ON DELETE SET NULL,
    property_id  UUID        REFERENCES limalar.properties(id) ON DELETE SET NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'scheduled',
    -- scheduled|confirmed|completed|cancelled|no_show
    notes        TEXT        NOT NULL DEFAULT '',
    address      TEXT        NOT NULL DEFAULT '',
    -- endereço do imóvel visitado (desnormalizado para exportar para Maps/Waze)
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.visit_events criada'

-- Busca principal: agenda do dia / semana para um corretor
CREATE INDEX IF NOT EXISTS idx_limalar_visit_events_agent_day
    ON limalar.visit_events (tenant_id, agent_id, scheduled_at);

-- Busca por lead (timeline de visitas de um lead)
CREATE INDEX IF NOT EXISTS idx_limalar_visit_events_lead
    ON limalar.visit_events (lead_id) WHERE lead_id IS NOT NULL;

-- Busca por imóvel
CREATE INDEX IF NOT EXISTS idx_limalar_visit_events_property
    ON limalar.visit_events (property_id) WHERE property_id IS NOT NULL;

CREATE OR REPLACE TRIGGER update_limalar_visit_events_updated_at
    BEFORE UPDATE ON limalar.visit_events
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

\echo '  Índices e trigger de limalar.visit_events criados'
\echo 'ETAPA 21 concluída com sucesso!'
