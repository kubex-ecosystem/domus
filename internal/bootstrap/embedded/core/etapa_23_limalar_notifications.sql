-- ============================================================================
-- ETAPA 23: Limalar — Centro de Notificações
-- ============================================================================
-- Cria as tabelas do sistema de notificações:
--   - limalar.notifications      : notificações geradas pelo sistema
--   - limalar.notification_prefs : preferências por evento/canal (mute-by-default)
--
-- Filosofia Calm UX: enabled = false por padrão em notification_prefs.
-- Usuário opta por receber, não precisa optar por silenciar.
--
-- Depende de: etapa_16 (schema limalar já existe)
-- ============================================================================
\echo 'ETAPA 23: Criando limalar.notifications e limalar.notification_prefs...'

-- ── Notificações ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.notifications (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    type            TEXT        NOT NULL,
    -- 'lead_qualified' | 'whatsapp_reply' | 'pipeline_moved'
    -- | 'commission_received' | 'score_updated' | 'system'
    urgency         TEXT        NOT NULL DEFAULT 'normal',
    -- 'low' | 'normal' | 'high'
    title           TEXT        NOT NULL DEFAULT '',
    body            TEXT        NOT NULL DEFAULT '',
    reference_id    UUID,
    reference_type  TEXT,
    -- 'lead' | 'commission' | 'pipeline_card' | 'property'
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
\echo '  Tabela limalar.notifications criada'

CREATE INDEX IF NOT EXISTS idx_limalar_notif_user_unread
    ON limalar.notifications (user_id, tenant_id, created_at DESC)
    WHERE read_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_limalar_notif_user_all
    ON limalar.notifications (user_id, tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_limalar_notif_reference
    ON limalar.notifications (reference_id, reference_type)
    WHERE reference_id IS NOT NULL;
\echo '  Índices de limalar.notifications criados'

-- ── Preferências de notificação ───────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS limalar.notification_prefs (
    id               UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID  NOT NULL,
    tenant_id        UUID  NOT NULL,
    event_type       TEXT  NOT NULL,
    channel          TEXT  NOT NULL DEFAULT 'in_app',
    -- 'in_app' | 'email' | 'whatsapp'
    enabled          BOOL  NOT NULL DEFAULT false,
    -- mute-by-default: usuário precisa optar por ativar
    digest_frequency TEXT  NOT NULL DEFAULT 'realtime',
    -- 'realtime' | 'daily' | 'weekly' | 'never'
    UNIQUE (user_id, tenant_id, event_type, channel)
);
\echo '  Tabela limalar.notification_prefs criada'

CREATE INDEX IF NOT EXISTS idx_limalar_notif_prefs_user
    ON limalar.notification_prefs (user_id, tenant_id);
\echo '  Índices de limalar.notification_prefs criados'

\echo 'ETAPA 23 concluída com sucesso!'
