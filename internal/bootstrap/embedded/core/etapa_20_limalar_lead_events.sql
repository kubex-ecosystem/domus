-- ============================================================================
-- ETAPA 20: Limalar — Timeline de Eventos de Leads
-- ============================================================================
-- Cria a tabela limalar.lead_events: log cronológico de atividades por lead.
-- Alimentada por: criação, mudança de status, notas, mensagens WhatsApp,
-- qualificação IA e qualquer outro evento relevante do ciclo do lead.
--
-- Depende de: etapa_18 (limalar.leads)
-- ============================================================================
\echo 'ETAPA 20: Criando limalar.lead_events (timeline de atividades)...'

CREATE TABLE IF NOT EXISTS limalar.lead_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    lead_id     UUID        NOT NULL REFERENCES limalar.leads(id) ON DELETE CASCADE,
    type        TEXT        NOT NULL DEFAULT 'note',
    -- 'created' | 'status_changed' | 'score_updated' | 'note_added'
    -- | 'whatsapp_message' | 'ai_qualified' | 'pipeline_moved'
    body        TEXT        NOT NULL DEFAULT '',
    metadata    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.lead_events criada'

-- Busca principal: timeline cronológica de um lead
CREATE INDEX IF NOT EXISTS idx_limalar_lead_events_lead
    ON limalar.lead_events (lead_id, created_at DESC);

-- Busca por tipo de evento (filtragem por categoria)
CREATE INDEX IF NOT EXISTS idx_limalar_lead_events_tenant_type
    ON limalar.lead_events (tenant_id, type, created_at DESC);

\echo '  Índices de limalar.lead_events criados'

\echo 'ETAPA 20 concluída com sucesso!'
