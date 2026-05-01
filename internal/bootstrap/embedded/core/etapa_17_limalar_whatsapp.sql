-- ============================================================================
-- ETAPA 17: Limalar — Mensagens WhatsApp Threaded
-- ============================================================================
-- Cria a tabela limalar.wapp_messages: histórico de conversas por telefone,
-- com role (lead/ai/human) e direction (inbound/outbound).
--
-- Diferente de public.whatsapp_messages (payload bruto do webhook Meta),
-- esta tabela representa a visão de thread processada para a UI do corretor.
--
-- Depende de: etapa_16 (limalar.properties deve existir para a FK opcional)
-- ============================================================================
\echo 'ETAPA 17: Criando limalar.wapp_messages (histórico threaded de conversas)...'

CREATE TABLE IF NOT EXISTS limalar.wapp_messages (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID        REFERENCES limalar.properties(id) ON DELETE SET NULL,
    tenant_id   UUID        NOT NULL,
    phone       TEXT        NOT NULL,
    direction   TEXT        NOT NULL DEFAULT 'inbound',   -- 'inbound' | 'outbound'
    role        TEXT        NOT NULL DEFAULT 'lead',      -- 'lead' | 'ai' | 'human'
    content     TEXT        NOT NULL DEFAULT '',
    score       INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.wapp_messages criada'

-- Índice principal: busca de thread por telefone + tenant, ordenada por tempo
CREATE INDEX IF NOT EXISTS idx_limalar_wapp_msg_phone
    ON limalar.wapp_messages (phone, tenant_id, created_at DESC);

-- Índice para listagem de todas as conversas de um tenant
CREATE INDEX IF NOT EXISTS idx_limalar_wapp_msg_tenant
    ON limalar.wapp_messages (tenant_id, created_at DESC);

-- Índice parcial para busca reversa por imóvel (property_id opcional)
CREATE INDEX IF NOT EXISTS idx_limalar_wapp_msg_property
    ON limalar.wapp_messages (property_id) WHERE property_id IS NOT NULL;
\echo '  Índices de limalar.wapp_messages criados'

\echo 'ETAPA 17 concluída com sucesso!'
