-- ============================================================================
-- ETAPA 26: Sift — CRM de captação (Growth Production)
-- ============================================================================
-- Schema sift com as entidades do funil de captação:
--   stage, turma, lead, lead_event, message, escalation, ig_profile, wa_number
--
-- Tenant isolation: todas as tabelas carregam tenant_id sem FK externa ao
-- schema public — schema sift é autônomo, mesma convenção da etapa 18
-- (limalar) e da 25 (cozinha_pass). O preço é integridade referencial de
-- tenant_id/user_id não checada pelo banco; o ganho é o schema poder migrar
-- para outro cluster sem arrastar o núcleo junto.
--
-- Não referencia limalar.* nem public.lead. Decisão registrada em
-- gnyx/.notes/context/sift_vision.md §6.5 (revisão de 2026-08-06).
--
-- Depende de: etapa_7 (public.update_updated_at_column)
-- ============================================================================
SET search_path TO public;
\echo 'ETAPA 26: Criando schema sift (CRM de captação)...'

CREATE SCHEMA IF NOT EXISTS sift;
\echo '  Schema sift criado'

-- ── Estágios do pipeline ─────────────────────────────────────────────────────
-- Estágio é dado, não código: a operação reordena, renomeia e insere sem
-- deploy. Por isso `key` NÃO tem CHECK de lista fechada — uma lista fixa aqui
-- anularia essa decisão logo na primeira vez que quisessem uma coluna nova.
-- O CHECK garante só o formato, para `key` continuar utilizável como
-- identificador no código.
--
-- `name` é o rótulo em pt-BR e muda à vontade. `is_terminal` existe para a UI
-- e as métricas não precisarem reconhecer 'enrolled'/'lost' pelo nome.

CREATE TABLE IF NOT EXISTS sift.stage (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    key         TEXT        NOT NULL CHECK (key ~ '^[a-z][a-z0-9_]*$'),
    name        TEXT        NOT NULL,
    order_index INTEGER     NOT NULL DEFAULT 0,
    color       TEXT        NOT NULL DEFAULT '',
    is_terminal BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Chave natural do estágio dentro do tenant. É por ela que o seed idempotente
-- da camada de aplicação faz upsert — o seed não cabe aqui, porque na hora da
-- migração nenhum tenant real existe ainda.
CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_stage_tenant_key
    ON sift.stage (tenant_id, key);

CREATE INDEX IF NOT EXISTS idx_sift_stage_tenant_order
    ON sift.stage (tenant_id, order_index ASC);

CREATE OR REPLACE TRIGGER update_sift_stage_updated_at
    BEFORE UPDATE ON sift.stage
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Tabela sift.stage criada'

-- ── Turma ────────────────────────────────────────────────────────────────────
-- Tabela fina e representativa: serve de FK e de exibição na landing. Sem CRUD
-- completo, sem gestão de matrícula — isso é Fase 2, se houver.

CREATE TABLE IF NOT EXISTS sift.turma (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID          NOT NULL,
    name          TEXT          NOT NULL,
    period_start  DATE,
    period_end    DATE,
    monthly_price NUMERIC(12,2) CHECK (monthly_price IS NULL OR monthly_price >= 0),
    status        TEXT          NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','open','closed','finished')),
    landing_slug  TEXT          NOT NULL CHECK (landing_slug ~ '^[a-z0-9][a-z0-9-]*$'),
    -- Qual template a landing usa. É coluna, e não decisão de código, porque a
    -- escolha depende do acervo disponível na hora: turma sem foto pronta fica
    -- melhor no template textual do que num de vitrine com buraco no lugar da
    -- imagem. A operação troca sem deploy quando o material chegar.
    landing_template TEXT       NOT NULL DEFAULT 'essencial'
        CHECK (landing_template IN ('essencial','vitrine','cinematica')),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT sift_turma_periodo_coerente
        CHECK (period_start IS NULL OR period_end IS NULL OR period_end >= period_start)
);

-- UNIQUE GLOBAL, não por tenant. A landing em /t/:slug é anônima: o slug da URL
-- é a única informação que o servidor tem para descobrir de quem é a turma.
-- Unicidade por tenant deixaria a rota pública ambígua.
CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_turma_slug
    ON sift.turma (landing_slug);

CREATE INDEX IF NOT EXISTS idx_sift_turma_tenant_status
    ON sift.turma (tenant_id, status, period_start DESC);

CREATE OR REPLACE TRIGGER update_sift_turma_updated_at
    BEFORE UPDATE ON sift.turma
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Tabela sift.turma criada'

-- ── Lead ─────────────────────────────────────────────────────────────────────
-- `captured_by` é quem captou. Não existe em nenhuma tabela de lead do
-- ecossistema, e é dele que sai a atribuição por pessoa.
--
-- `source` é DELIBERADAMENTE NULLABLE. A ferramenta suporta a possibilidade de
-- duas origens, não exige as duas — a comparação entre elas é uma leitura que
-- aparece quando houver dado dos dois lados, nunca um requisito do fluxo. Além
-- disso, um dos objetivos é que uma origem substitua a outra com o tempo.
-- Obrigar classificação em toda entrada seria atrito, e atrito derruba a
-- adesão, que é a régua honesta do piloto.
--
-- O CHECK usa lista fechada (ao contrário de stage.key) porque origem é
-- vocabulário de análise, não configuração da operação. Origem nova entra por
-- migração aditiva, de propósito.

CREATE TABLE IF NOT EXISTS sift.lead (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID        NOT NULL,
    stage_id      UUID        NOT NULL REFERENCES sift.stage(id),
    turma_id      UUID        REFERENCES sift.turma(id) ON DELETE SET NULL,
    name          TEXT        NOT NULL DEFAULT '',
    ig_handle     TEXT        NOT NULL DEFAULT '',
    phone         TEXT        NOT NULL DEFAULT '',
    email         TEXT        NOT NULL DEFAULT '',
    source        TEXT        CHECK (source IS NULL OR source IN
                      ('ad','landing','dm_manual','referral','organic')),
    source_detail TEXT        NOT NULL DEFAULT '',
    captured_by   UUID,
    assigned_to   UUID,
    fit_score     INTEGER     CHECK (fit_score IS NULL OR fit_score BETWEEN 0 AND 100),
    fit_reason    TEXT        NOT NULL DEFAULT '',
    lost_reason   TEXT        CHECK (lost_reason IS NULL OR lost_reason IN
                      ('price','timing','no_fit','ghosted')),
    notes         TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Único campo genuinamente obrigatório: alguma forma de alcançar a pessoa.
    -- Lead sem handle, sem telefone e sem e-mail não é lead — é anotação.
    CONSTRAINT sift_lead_tem_contato
        CHECK (ig_handle <> '' OR phone <> '' OR email <> '')
);

CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_stage
    ON sift.lead (tenant_id, stage_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_created
    ON sift.lead (tenant_id, created_at DESC);

-- Dedupe por handle: o mesmo perfil não deve virar dois cards. Case-insensitive
-- porque o Instagram trata @Fulana e @fulana como o mesmo perfil.
CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_lead_tenant_handle
    ON sift.lead (tenant_id, lower(ig_handle)) WHERE ig_handle <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_lead_tenant_phone
    ON sift.lead (tenant_id, phone) WHERE phone <> '';

-- Sustenta a leitura por origem — que só produz comparação quando há mais de
-- uma origem preenchida, e degrada para métrica de um funil só quando não há.
CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_source
    ON sift.lead (tenant_id, source, created_at DESC) WHERE source IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sift_lead_turma
    ON sift.lead (turma_id) WHERE turma_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sift_lead_assigned
    ON sift.lead (tenant_id, assigned_to) WHERE assigned_to IS NOT NULL;

CREATE OR REPLACE TRIGGER update_sift_lead_updated_at
    BEFORE UPDATE ON sift.lead
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Tabela sift.lead criada'

-- ── Eventos do lead ──────────────────────────────────────────────────────────
-- Toda transição vira evento com timestamp. É o que dá tempo-até-escalar e
-- tempo-em-estágio sem cálculo posterior.
--
-- Aqui a lista É fechada: tipo de evento é vocabulário do código, não da
-- operação. Tipo novo entra por migração aditiva.

CREATE TABLE IF NOT EXISTS sift.lead_event (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID        NOT NULL,
    lead_id       UUID        NOT NULL REFERENCES sift.lead(id) ON DELETE CASCADE,
    type          TEXT        NOT NULL CHECK (type IN
                      ('created','stage_changed','message_in','message_out',
                       'fit_scored','escalated','turma_linked','lost','note_added')),
    body          TEXT        NOT NULL DEFAULT '',
    actor_user_id UUID,
    metadata      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sift_lead_event_lead
    ON sift.lead_event (lead_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sift_lead_event_tenant_type
    ON sift.lead_event (tenant_id, type, created_at DESC);
\echo '  Tabela sift.lead_event criada'

-- ── Mensagens ────────────────────────────────────────────────────────────────
-- Timeline unificada dos dois canais. Existe porque nada no substrato serve:
-- LeadConversationStore é log de turnos de IA e exige property_id NOT NULL
-- (etapa 16); limalar.wapp_messages é preso a `phone` e não tem lead_id; e
-- whatsapp_messages (etapa 13) é log bruto global, sem tenant_id.

CREATE TABLE IF NOT EXISTS sift.message (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    lead_id     UUID        NOT NULL REFERENCES sift.lead(id) ON DELETE CASCADE,
    channel     TEXT        NOT NULL CHECK (channel IN ('ig','wa')),
    direction   TEXT        NOT NULL CHECK (direction IN ('in','out')),
    author      TEXT        NOT NULL DEFAULT 'lead' CHECK (author IN ('lead','ai','human')),
    content     TEXT        NOT NULL DEFAULT '',
    external_id TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sift_message_lead
    ON sift.message (lead_id, created_at ASC);

-- Idempotência do webhook. Meta reentrega evento, e sem isto a mesma mensagem
-- vira várias linhas na timeline. O id externo é único no provedor, então a
-- unicidade é global e não por tenant.
CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_message_external
    ON sift.message (channel, external_id) WHERE external_id <> '';
\echo '  Tabela sift.message criada'

-- ── Escalação ────────────────────────────────────────────────────────────────
-- A transição mais importante do sistema: é onde a conversão sangra hoje.
-- Guarda quem escalou e para quem — sem isso a atribuição não fecha.
--
-- `summary_generated` distingue "não tem resumo" de "resumo vazio": quando a
-- geração falha ou não há chave de IA, a tela cai para o histórico cru. O
-- handoff nunca depende do LLM estar de pé.

CREATE TABLE IF NOT EXISTS sift.escalation (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID        NOT NULL,
    lead_id           UUID        NOT NULL REFERENCES sift.lead(id) ON DELETE CASCADE,
    from_user_id      UUID,
    to_user_id        UUID,
    escalated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    summary_generated BOOLEAN     NOT NULL DEFAULT FALSE,
    summary_text      TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_sift_escalation_lead
    ON sift.escalation (lead_id, escalated_at DESC);

CREATE INDEX IF NOT EXISTS idx_sift_escalation_tenant_to
    ON sift.escalation (tenant_id, to_user_id, escalated_at DESC);
\echo '  Tabela sift.escalation criada'

-- ── Perfil do Instagram ──────────────────────────────────────────────────────
-- Um perfil por lead, daí lead_id como PK. `last_window_at` é o carimbo da
-- última mensagem recebida: é ele que diz se a janela de resposta inbound ainda
-- está aberta. Fora dela não se envia nada — restrição §10.1, sem exceção.

CREATE TABLE IF NOT EXISTS sift.ig_profile (
    lead_id        UUID        PRIMARY KEY REFERENCES sift.lead(id) ON DELETE CASCADE,
    tenant_id      UUID        NOT NULL,
    handle         TEXT        NOT NULL DEFAULT '',
    ig_user_id     TEXT        NOT NULL DEFAULT '',
    last_window_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_ig_profile_user
    ON sift.ig_profile (tenant_id, ig_user_id) WHERE ig_user_id <> '';

CREATE OR REPLACE TRIGGER update_sift_ig_profile_updated_at
    BEFORE UPDATE ON sift.ig_profile
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
\echo '  Tabela sift.ig_profile criada'

-- ── Mapa de número de WhatsApp → tenant ──────────────────────────────────────
-- Preenche a lacuna da etapa 13: whatsapp_messages não tem tenant_id, é log
-- bruto global chaveado por phone_number_id. Sem este mapa, uma mensagem que
-- chega no webhook não tem como ser atribuída a um tenant.
--
-- phone_number_id é único no provedor, por isso é a PK.

CREATE TABLE IF NOT EXISTS sift.wa_number (
    phone_number_id TEXT        PRIMARY KEY,
    tenant_id       UUID        NOT NULL,
    label           TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sift_wa_number_tenant
    ON sift.wa_number (tenant_id);
\echo '  Tabela sift.wa_number criada'

\echo 'ETAPA 26 concluída com sucesso!'
