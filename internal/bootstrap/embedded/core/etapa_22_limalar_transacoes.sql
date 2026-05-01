-- ============================================================================
-- ETAPA 22: Limalar — Transações e Documentos
-- ============================================================================
-- Cria as tabelas:
--   limalar.transactions      — negociações imobiliárias com status de fechamento
--   limalar.transaction_steps — checklist de etapas (due-diligence, cartório, etc.)
--   limalar.transaction_docs  — documentos do processo (contrato, matrícula, etc.)
--
-- Depende de: etapa_16 (limalar.properties), etapa_18 (limalar.leads)
-- ============================================================================
\echo 'ETAPA 22: Criando limalar.transactions, transaction_steps, transaction_docs...'

CREATE TABLE IF NOT EXISTS limalar.transactions (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID          NOT NULL,
    property_id    UUID          REFERENCES limalar.properties(id) ON DELETE SET NULL,
    buyer_lead_id  UUID          REFERENCES limalar.leads(id) ON DELETE SET NULL,
    status         TEXT          NOT NULL DEFAULT 'negotiation',
    -- negotiation|contract_signed|due_diligence|financing|closing|completed|cancelled
    commission_pct NUMERIC(5, 2) NOT NULL DEFAULT 6.00,
    closing_date   DATE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.transactions criada'

CREATE TABLE IF NOT EXISTS limalar.transaction_steps (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID        NOT NULL REFERENCES limalar.transactions(id) ON DELETE CASCADE,
    step_code      TEXT        NOT NULL,
    title          TEXT        NOT NULL DEFAULT '',
    status         TEXT        NOT NULL DEFAULT 'pending',
    -- pending|in_progress|done|skipped
    due_date       DATE,
    completed_at   TIMESTAMPTZ,
    assignee_id    UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.transaction_steps criada'

CREATE TABLE IF NOT EXISTS limalar.transaction_docs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID        NOT NULL REFERENCES limalar.transactions(id) ON DELETE CASCADE,
    file_name      TEXT        NOT NULL DEFAULT '',
    file_url       TEXT        NOT NULL DEFAULT '',
    doc_type       TEXT        NOT NULL DEFAULT 'other',
    -- contract|deed|financing_proposal|id_document|other
    extracted_data JSONB       NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.transaction_docs criada'

-- Índices de busca principal
CREATE INDEX IF NOT EXISTS idx_limalar_transactions_tenant_status
    ON limalar.transactions (tenant_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_transactions_property
    ON limalar.transactions (property_id) WHERE property_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_limalar_transactions_buyer
    ON limalar.transactions (buyer_lead_id) WHERE buyer_lead_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_limalar_transaction_steps_tx
    ON limalar.transaction_steps (transaction_id, created_at);

CREATE INDEX IF NOT EXISTS idx_limalar_transaction_docs_tx
    ON limalar.transaction_docs (transaction_id, created_at);

CREATE OR REPLACE TRIGGER update_limalar_transactions_updated_at
    BEFORE UPDATE ON limalar.transactions
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

\echo '  Índices e trigger de limalar.transactions criados'
\echo 'ETAPA 22 concluída com sucesso!'
