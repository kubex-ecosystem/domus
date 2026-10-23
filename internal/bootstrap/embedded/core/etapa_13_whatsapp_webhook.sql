-- ============================================================================
-- ETAPA 13: WhatsApp Webhook Storage
-- ============================================================================
-- Cria tabelas de armazenamento bruto de webhooks Meta (mensagens e status).
-- Essas tabelas vivem em public pois são infra de plataforma — não pertencem
-- a nenhum domínio de produto específico.
-- ============================================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
SET search_path TO public;
\echo 'ETAPA 13: Criando storage de webhooks WhatsApp (Meta Cloud API)...'

-- Mensagens brutas recebidas via webhook Meta
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id              TEXT        PRIMARY KEY,
    phone_number_id TEXT        NOT NULL DEFAULT '',
    from_number     TEXT        NOT NULL DEFAULT '',
    type            TEXT        NOT NULL DEFAULT '',
    payload         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela whatsapp_messages criada'

CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_phone_created
    ON whatsapp_messages (phone_number_id, created_at DESC);
\echo '  Índice idx_whatsapp_messages_phone_created criado'

-- Atualizações de status de entrega (delivered, read, failed)
CREATE TABLE IF NOT EXISTS whatsapp_statuses (
    id           TEXT        NOT NULL,
    status       TEXT        NOT NULL,
    recipient_id TEXT        NOT NULL DEFAULT '',
    payload      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, status)
);
\echo '  Tabela whatsapp_statuses criada'

CREATE INDEX IF NOT EXISTS idx_whatsapp_statuses_recipient
    ON whatsapp_statuses (recipient_id, created_at DESC);
\echo '  Índice idx_whatsapp_statuses_recipient criado'

\echo 'ETAPA 13 concluída com sucesso!'
