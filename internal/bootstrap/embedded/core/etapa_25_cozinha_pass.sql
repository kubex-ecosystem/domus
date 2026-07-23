-- ============================================================================
-- ETAPA 25: Cozinha Pass — KDS Multi-Tenant
-- ============================================================================
-- Schema cozinha_pass com todas as entidades do Kitchen Display System:
--   company, company_channel, insumo, produto, componente_ficha,
--   orders, order_item, message, movimentacao_estoque
-- ============================================================================
SET search_path TO public;
\echo 'ETAPA 25: Criando schema cozinha_pass (Cozinha Pass — KDS)...'

CREATE SCHEMA IF NOT EXISTS cozinha_pass;
\echo '  Schema cozinha_pass criado'

-- ── Empresa locatária (tenant do KDS) ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.company (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    logo          TEXT,
    brand_color   TEXT        NOT NULL DEFAULT '#f97316',
    cuisine       TEXT        NOT NULL DEFAULT '',
    active_status TEXT        NOT NULL DEFAULT 'active'
        CHECK (active_status IN ('active','paused','closed')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela cozinha_pass.company criada'

CREATE TABLE IF NOT EXISTS cozinha_pass.company_channel (
    company_id UUID NOT NULL REFERENCES cozinha_pass.company(id) ON DELETE CASCADE,
    channel    TEXT NOT NULL
        CHECK (channel IN ('whatsapp','instagram','email','sms','ifood','99','manual')),
    PRIMARY KEY (company_id, channel)
);
\echo '  Tabela cozinha_pass.company_channel criada'

-- ── Insumos (matéria-prima) ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.insumo (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID          NOT NULL REFERENCES cozinha_pass.company(id) ON DELETE CASCADE,
    nome           TEXT          NOT NULL,
    unidade        TEXT          NOT NULL CHECK (unidade IN ('un','kg','g','L','ml')),
    estoque        NUMERIC(14,4) NOT NULL DEFAULT 0 CHECK (estoque >= 0),
    estoque_minimo NUMERIC(14,4) NOT NULL DEFAULT 0 CHECK (estoque_minimo >= 0),
    custo_unitario NUMERIC(12,4) NOT NULL DEFAULT 0 CHECK (custo_unitario >= 0),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ck_insumo_company ON cozinha_pass.insumo (company_id);
\echo '  Tabela cozinha_pass.insumo criada'

-- ── Produtos (processados e finais) ──────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.produto (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID          NOT NULL REFERENCES cozinha_pass.company(id) ON DELETE CASCADE,
    nome        TEXT          NOT NULL,
    tipo        TEXT          NOT NULL CHECK (tipo IN ('processado','final')),
    preco_venda NUMERIC(12,2) CHECK (preco_venda IS NULL OR preco_venda >= 0),
    CONSTRAINT preco_so_final CHECK (tipo = 'final' OR preco_venda IS NULL),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ck_produto_company ON cozinha_pass.produto (company_id);
\echo '  Tabela cozinha_pass.produto criada'

-- ── Ficha técnica — BOM recursivo ────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.componente_ficha (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    produto_id     UUID          NOT NULL REFERENCES cozinha_pass.produto(id) ON DELETE CASCADE,
    ref_tipo       TEXT          NOT NULL CHECK (ref_tipo IN ('insumo','produto')),
    insumo_ref_id  UUID          REFERENCES cozinha_pass.insumo(id)   ON DELETE RESTRICT,
    produto_ref_id UUID          REFERENCES cozinha_pass.produto(id)  ON DELETE RESTRICT,
    quantidade     NUMERIC(14,4) NOT NULL CHECK (quantidade > 0),
    CONSTRAINT ref_coerente CHECK (
        (ref_tipo = 'insumo'  AND insumo_ref_id  IS NOT NULL AND produto_ref_id IS NULL) OR
        (ref_tipo = 'produto' AND produto_ref_id IS NOT NULL AND insumo_ref_id  IS NULL)
    ),
    CONSTRAINT sem_auto_ref CHECK (produto_ref_id IS NULL OR produto_ref_id <> produto_id)
);
CREATE INDEX IF NOT EXISTS idx_ck_componente_produto ON cozinha_pass.componente_ficha (produto_id);
\echo '  Tabela cozinha_pass.componente_ficha criada'

-- ── Pedidos ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.orders (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID          NOT NULL REFERENCES cozinha_pass.company(id) ON DELETE CASCADE,
    ticket      TEXT          NOT NULL,
    client_name TEXT          NOT NULL,
    channel     TEXT          NOT NULL
        CHECK (channel IN ('whatsapp','instagram','email','sms','ifood','99','manual')),
    status      TEXT          NOT NULL DEFAULT 'new'
        CHECK (status IN ('new','preparing','ready','delivered','cancelled')),
    destination TEXT,
    note        TEXT,
    total_value NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, ticket)
);
CREATE INDEX IF NOT EXISTS idx_ck_orders_company_status  ON cozinha_pass.orders (company_id, status);
CREATE INDEX IF NOT EXISTS idx_ck_orders_company_created ON cozinha_pass.orders (company_id, created_at DESC);
\echo '  Tabela cozinha_pass.orders criada'

CREATE TABLE IF NOT EXISTS cozinha_pass.order_item (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES cozinha_pass.orders(id)  ON DELETE CASCADE,
    produto_id UUID          REFERENCES cozinha_pass.produto(id)          ON DELETE SET NULL,
    name       TEXT          NOT NULL,
    quantity   NUMERIC(14,4) NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL,
    notes      TEXT
);
CREATE INDEX IF NOT EXISTS idx_ck_order_item_order ON cozinha_pass.order_item (order_id);
\echo '  Tabela cozinha_pass.order_item criada'

-- ── Inbox de mensagens ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.message (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID        NOT NULL REFERENCES cozinha_pass.company(id) ON DELETE CASCADE,
    order_id    UUID        REFERENCES cozinha_pass.orders(id)           ON DELETE SET NULL,
    channel     TEXT        NOT NULL
        CHECK (channel IN ('whatsapp','instagram','email','sms','ifood','99','manual')),
    client_name TEXT        NOT NULL,
    content     TEXT        NOT NULL,
    direction   TEXT        NOT NULL CHECK (direction IN ('inbound','outbound')),
    read        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ck_message_company_order   ON cozinha_pass.message (company_id, order_id);
CREATE INDEX IF NOT EXISTS idx_ck_message_company_created ON cozinha_pass.message (company_id, created_at DESC);
\echo '  Tabela cozinha_pass.message criada'

-- ── Livro-razão de movimentações de estoque ──────────────────────────────────

CREATE TABLE IF NOT EXISTS cozinha_pass.movimentacao_estoque (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    insumo_id  UUID          NOT NULL REFERENCES cozinha_pass.insumo(id)  ON DELETE CASCADE,
    order_id   UUID          REFERENCES cozinha_pass.orders(id)           ON DELETE SET NULL,
    tipo       TEXT          NOT NULL CHECK (tipo IN ('baixa','entrada','ajuste','estorno')),
    quantidade NUMERIC(14,4) NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ck_movimentacao_insumo
    ON cozinha_pass.movimentacao_estoque (insumo_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ck_movimentacao_baixa_idempotente
    ON cozinha_pass.movimentacao_estoque (order_id, insumo_id) WHERE tipo = 'baixa';
\echo '  Tabela cozinha_pass.movimentacao_estoque criada'

\echo 'ETAPA 25 concluída com sucesso!'
