-- ============================================================================
-- ETAPA 19: Limalar — Módulo Financeiro (Comissões)
-- ============================================================================
-- Cria a tabela limalar.commissions: rastreamento de comissões do corretor.
--
-- Tenant isolation: tenant_id sem FK externa ao schema public.
-- user_id identifica o corretor responsável (sem FK — evita acoplamento com
-- public.users para manter o schema limalar autônomo).
--
-- Depende de: etapa_16 (limalar.properties para a FK opcional)
-- ============================================================================
\echo 'ETAPA 19: Criando limalar.commissions (Módulo Financeiro)...'

CREATE TABLE IF NOT EXISTS limalar.commissions (
    id          UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID           NOT NULL,
    user_id     UUID           NOT NULL,
    property_id UUID           REFERENCES limalar.properties(id) ON DELETE SET NULL,
    gross_value BIGINT         NOT NULL DEFAULT 0,
    -- valor bruto da venda em centavos
    pct         NUMERIC(5, 2)  NOT NULL DEFAULT 0,
    -- percentual de comissão (ex: 6.00 para 6%)
    net_value   BIGINT         GENERATED ALWAYS AS (ROUND(gross_value * pct / 100)) STORED,
    -- valor líquido da comissão em centavos (coluna gerada)
    status      TEXT           NOT NULL DEFAULT 'pending',
    -- 'pending' | 'received' | 'overdue'
    notes       TEXT           NOT NULL DEFAULT '',
    paid_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.commissions criada'

-- Busca principal: listar comissões de um tenant por status e data
CREATE INDEX IF NOT EXISTS idx_limalar_commissions_tenant_status
    ON limalar.commissions (tenant_id, status, created_at DESC);

-- Busca por corretor
CREATE INDEX IF NOT EXISTS idx_limalar_commissions_user
    ON limalar.commissions (tenant_id, user_id, created_at DESC);

-- Busca por imóvel
CREATE INDEX IF NOT EXISTS idx_limalar_commissions_property
    ON limalar.commissions (property_id) WHERE property_id IS NOT NULL;

CREATE OR REPLACE TRIGGER update_limalar_commissions_updated_at
    BEFORE UPDATE ON limalar.commissions
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Índices e trigger de limalar.commissions criados'

\echo 'ETAPA 19 concluída com sucesso!'
