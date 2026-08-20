-- ============================================================================
-- ETAPA 27: Sift — massa de demonstração
-- ============================================================================
-- Popula o schema sift com um funil verossímil, para o frontend ter o que
-- mostrar e o que lapidar antes de existir operação real.
--
-- CONFINAMENTO. Tudo aqui pertence ao tenant de demonstração da etapa 8
-- ('20000000-0000-0000-0000-000000000001', Kubex Demo Corp). Como o gateway
-- resolve tenant por membership, quem não for membro desse tenant não vê nada
-- disto — inclusive numa instalação de produção.
--
-- REMOÇÃO. Um comando, e o schema fica intacto:
--
--   DELETE FROM sift.lead  WHERE tenant_id = '20000000-0000-0000-0000-000000000001';
--   DELETE FROM sift.turma WHERE tenant_id = '20000000-0000-0000-0000-000000000001';
--   DELETE FROM sift.stage WHERE tenant_id = '20000000-0000-0000-0000-000000000001';
--   DELETE FROM sift.wa_number WHERE tenant_id = '20000000-0000-0000-0000-000000000001';
--
-- (evento, mensagem, escalação e perfil saem por ON DELETE CASCADE do lead)
--
-- IDEMPOTÊNCIA. Todo id é derivado por md5 de uma chave estável, e todo INSERT
-- termina em ON CONFLICT DO NOTHING — sem arbiter, para cobrir também as
-- unicidades de handle, telefone e slug. Reexecutar não duplica nada.
--
-- SEM CREDENCIAL. Os usuários de demonstração nascem com password_hash NULL:
-- servem para aparecer como nome na tela, não para autenticar. Seed que carrega
-- hash de senha conhecida vira porta dos fundos em produção.
--
-- Depende de: etapa_8 (tenant e roles de demonstração), etapa_26 (schema sift)
-- ============================================================================
SET search_path TO public;
\echo 'ETAPA 27: Populando sift com massa de demonstração...'

-- ── Pessoas da operação ──────────────────────────────────────────────────────
-- Três papéis, porque atribuição e handoff não são demonstráveis com um usuário
-- só: é preciso ver quem captou, quem escalou e quem recebeu.

INSERT INTO "user" (id, email, name, last_name, status, password_hash, force_password_reset)
VALUES
    ('41000000-0000-0000-0000-000000000001', 'camila@sift.demo',  'Camila', 'Prospecção', 'active', NULL, false),
    ('41000000-0000-0000-0000-000000000002', 'renata@sift.demo',  'Renata', 'Instrutora', 'active', NULL, false),
    ('41000000-0000-0000-0000-000000000003', 'sonia@sift.demo',   'Sônia',  'Sócia',      'active', NULL, false)
ON CONFLICT DO NOTHING;

INSERT INTO tenant_membership (user_id, tenant_id, role_id, is_active)
VALUES
    ('41000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', true),
    ('41000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', true),
    ('41000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', true)
ON CONFLICT DO NOTHING;
\echo '  Usuários de demonstração criados'

-- ── Estágios ─────────────────────────────────────────────────────────────────
-- Esta é também a referência de formato para o seed idempotente que a camada de
-- aplicação vai rodar por tenant real (não cabe em DDL: na migração nenhum
-- tenant real existe ainda).

INSERT INTO sift.stage (id, tenant_id, key, name, order_index, color, is_terminal)
VALUES
    (md5('sift-stage-captured')::uuid,        '20000000-0000-0000-0000-000000000001', 'captured',        'Captado',     1, '#94a3b8', false),
    (md5('sift-stage-qualified')::uuid,       '20000000-0000-0000-0000-000000000001', 'qualified',       'Qualificado', 2, '#38bdf8', false),
    (md5('sift-stage-in_conversation')::uuid, '20000000-0000-0000-0000-000000000001', 'in_conversation', 'Em conversa', 3, '#818cf8', false),
    (md5('sift-stage-escalated')::uuid,       '20000000-0000-0000-0000-000000000001', 'escalated',       'Escalado',    4, '#f59e0b', false),
    (md5('sift-stage-proposal')::uuid,        '20000000-0000-0000-0000-000000000001', 'proposal',        'Proposta',    5, '#a855f7', false),
    (md5('sift-stage-enrolled')::uuid,        '20000000-0000-0000-0000-000000000001', 'enrolled',        'Matriculado', 6, '#22c55e', true),
    (md5('sift-stage-lost')::uuid,            '20000000-0000-0000-0000-000000000001', 'lost',            'Perdido',     7, '#ef4444', true),
    (md5('sift-stage-nurturing')::uuid,       '20000000-0000-0000-0000-000000000001', 'nurturing',       'Nutrição',    8, '#14b8a6', false)
ON CONFLICT DO NOTHING;
\echo '  8 estágios criados'

-- ── Turmas ───────────────────────────────────────────────────────────────────
-- Uma aberta (destino da maioria), uma fechada (que justifica a existência dos
-- leads em Nutrição) e uma encerrada (histórico).

-- Uma turma por template, para as três variantes de landing terem o que mostrar.

INSERT INTO sift.turma (id, tenant_id, name, period_start, period_end, monthly_price, status, landing_slug, landing_template)
VALUES
    (md5('sift-turma-out')::uuid, '20000000-0000-0000-0000-000000000001',
     'Imagem Executiva — Outubro/2026',
     CURRENT_DATE + 20, CURRENT_DATE + 110, 890.00,  'open',
     'demo-imagem-executiva-outubro', 'vitrine'),
    (md5('sift-turma-ago')::uuid, '20000000-0000-0000-0000-000000000001',
     'Imagem Executiva — Agosto/2026',
     CURRENT_DATE - 40, CURRENT_DATE + 50,  890.00,  'closed',
     'demo-imagem-executiva-agosto', 'essencial'),
    (md5('sift-turma-jun')::uuid, '20000000-0000-0000-0000-000000000001',
     'Presença Digital — Junho/2026',
     CURRENT_DATE - 130, CURRENT_DATE - 40, 690.00,  'finished',
     'demo-presenca-digital-junho', 'cinematica')
ON CONFLICT DO NOTHING;
\echo '  3 turmas criadas'

-- ── Leads ────────────────────────────────────────────────────────────────────
-- 40 leads com formato de funil (estreita conforme avança), não distribuição
-- uniforme — funil uniforme ensina intuição errada a quem olha a tela.
--
-- As origens vêm misturadas de propósito, com `dm_manual` pesando mais: é a
-- incumbente, responsável por ~95% da receita hoje. E 5% dos leads ficam com
-- `source` NULL — não é lacuna do seed, é o caso que a tela precisa aguentar,
-- já que classificar origem nunca é obrigatório.

INSERT INTO sift.lead (
    id, tenant_id, stage_id, turma_id, name, ig_handle, phone, email,
    source, source_detail, captured_by, assigned_to,
    fit_score, fit_reason, lost_reason, notes, created_at, updated_at
)
SELECT
    md5('sift-demo-lead-' || d.seq)::uuid,
    '20000000-0000-0000-0000-000000000001',
    st.id,
    CASE
        WHEN d.stage_key = 'nurturing' THEN md5('sift-turma-ago')::uuid
        WHEN d.stage_key IN ('enrolled', 'lost') AND d.seq % 3 = 0 THEN md5('sift-turma-jun')::uuid
        WHEN d.stage_key IN ('captured') AND d.seq % 4 = 0 THEN NULL
        ELSE md5('sift-turma-out')::uuid
    END,
    (ARRAY['Adriana','Beatriz','Cristina','Denise','Eliana','Fernanda','Gláucia','Helena',
           'Isabel','Juliana','Kátia','Lúcia','Marta','Neide','Olívia','Patrícia',
           'Regina','Sandra','Tereza','Vera'])[1 + ((d.seq - 1) % 20)]
      || ' ' ||
    (ARRAY['Andrade','Barbosa','Carvalho'])[1 + ((d.seq - 1) / 20)],
    '@' ||
    lower((ARRAY['adriana','beatriz','cristina','denise','eliana','fernanda','glaucia','helena',
                 'isabel','juliana','katia','lucia','marta','neide','olivia','patricia',
                 'regina','sandra','tereza','vera'])[1 + ((d.seq - 1) % 20)])
      || '.' ||
    lower((ARRAY['andrade','barbosa','carvalho'])[1 + ((d.seq - 1) / 20)]),
    -- Telefone só aparece quando a conversa migrou para o WhatsApp. Antes
    -- disso, a pessoa é um handle do Instagram e nada mais — que é exatamente
    -- como o funil funciona hoje.
    CASE WHEN d.stage_key IN ('escalated','proposal','enrolled','lost','nurturing')
         THEN '5531' || (900000000 + d.seq)::text
         ELSE '' END,
    CASE WHEN d.stage_key IN ('proposal','enrolled')
         THEN lower((ARRAY['adriana','beatriz','cristina','denise','eliana','fernanda','glaucia','helena',
                           'isabel','juliana','katia','lucia','marta','neide','olivia','patricia',
                           'regina','sandra','tereza','vera'])[1 + ((d.seq - 1) % 20)]) || '@exemplo.com'
         ELSE '' END,
    -- O multiplicador 7 é coprimo de 20 e existe para espalhar. Sem ele, como
    -- `seq` percorre um estágio de cada vez, cada bloco do funil receberia uma
    -- faixa contígua de origens — e a tela mostraria coisas como "dm_manual
    -- nunca qualifica" ou "inbound sempre perde". Correlação inventada pelo
    -- gerador que quem olha o demo leria como descoberta.
    CASE ((d.seq * 7) % 20)
        WHEN 0 THEN 'dm_manual' WHEN 1 THEN 'dm_manual' WHEN 2 THEN 'dm_manual'
        WHEN 3 THEN 'dm_manual' WHEN 4 THEN 'dm_manual' WHEN 5 THEN 'dm_manual'
        WHEN 6 THEN 'dm_manual' WHEN 7 THEN 'dm_manual'
        WHEN 8 THEN 'landing'   WHEN 9 THEN 'landing'
        WHEN 10 THEN 'landing'  WHEN 11 THEN 'landing'
        WHEN 12 THEN 'ad'       WHEN 13 THEN 'ad'
        WHEN 14 THEN 'ad'       WHEN 15 THEN 'ad'
        WHEN 16 THEN 'referral' WHEN 17 THEN 'referral'
        WHEN 18 THEN 'organic'
        ELSE NULL
    END,
    CASE ((d.seq * 7) % 20)
        WHEN 12 THEN 'Campanha Turma Outubro — público lookalike'
        WHEN 13 THEN 'Campanha Turma Outubro — engajamento IG'
        WHEN 16 THEN 'Indicação de aluna da turma de junho'
        ELSE ''
    END,
    (ARRAY['41000000-0000-0000-0000-000000000001',
           '41000000-0000-0000-0000-000000000001',
           '41000000-0000-0000-0000-000000000003'])[1 + (d.seq % 3)]::uuid,
    CASE WHEN d.stage_key IN ('escalated','proposal','enrolled')
         THEN (ARRAY['41000000-0000-0000-0000-000000000002',
                     '41000000-0000-0000-0000-000000000003'])[1 + (d.seq % 2)]::uuid
         ELSE '41000000-0000-0000-0000-000000000001'::uuid END,
    CASE WHEN d.stage_key = 'captured' THEN NULL
         ELSE 40 + ((d.seq * 7) % 55) END,
    CASE WHEN d.stage_key = 'captured' THEN ''
         ELSE (ARRAY['Empresária, 47, BH — perfil dentro do ICP',
                     'Advogada, 52, Nova Lima — marcador profissional forte',
                     'Médica, 44, BH — engajamento alto no perfil',
                     'Corretora, 58, Contagem — já acompanha há meses',
                     'Fora da faixa etária alvo, mas contexto compatível'])[1 + (d.seq % 5)] END,
    CASE WHEN d.stage_key = 'lost'
         THEN (ARRAY['price','timing','no_fit','ghosted'])[1 + (d.seq % 4)]
         ELSE NULL END,
    CASE WHEN d.stage_key = 'nurturing'
         THEN 'Quis entrar na turma de agosto, que já havia fechado. Retomar quando abrir a próxima.'
         ELSE '' END,
    NOW() - ((d.dias_base + d.n * 2) || ' days')::interval,
    NOW() - ((d.dias_base + d.n * 2) || ' days')::interval
FROM (
    SELECT p.stage_key,
           p.dias_base,
           g AS n,
           row_number() OVER (ORDER BY p.ordem, g) AS seq
    FROM (VALUES
            ('captured',        1,  9,  4),
            ('qualified',       2,  6, 10),
            ('in_conversation', 3,  5, 16),
            ('escalated',       4,  3, 22),
            ('proposal',        5,  3, 28),
            ('enrolled',        6,  4, 44),
            ('lost',            7,  6, 36),
            ('nurturing',       8,  4, 50)
         ) AS p(stage_key, ordem, qtd, dias_base)
    CROSS JOIN LATERAL generate_series(1, p.qtd) AS g
) d
JOIN sift.stage st
  ON st.tenant_id = '20000000-0000-0000-0000-000000000001'
 AND st.key = d.stage_key
ON CONFLICT DO NOTHING;
\echo '  40 leads criados'

-- ── Perfis do Instagram ──────────────────────────────────────────────────────

INSERT INTO sift.ig_profile (lead_id, tenant_id, handle, ig_user_id, last_window_at)
SELECT l.id,
       l.tenant_id,
       l.ig_handle,
       'igid-' || substr(md5(l.id::text), 1, 12),
       -- Janela de resposta viva só para quem falou nas últimas 24h. É o que
       -- separa "posso responder" de "não posso mandar nada".
       CASE WHEN s.key IN ('in_conversation','escalated')
            THEN NOW() - interval '3 hours'
            ELSE l.created_at END
FROM sift.lead l
JOIN sift.stage s ON s.id = l.stage_id
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001'
  AND l.ig_handle <> ''
ON CONFLICT DO NOTHING;
\echo '  Perfis de Instagram vinculados'

-- ── Mensagens ────────────────────────────────────────────────────────────────
-- Começa no Instagram e migra para o WhatsApp no meio da thread. É o ponto do
-- produto: a mesma conversa, os dois canais, uma timeline só.

INSERT INTO sift.message (id, tenant_id, lead_id, channel, direction, author, content, external_id, created_at)
SELECT
    md5('sift-demo-msg-' || l.id::text || '-' || m.i)::uuid,
    l.tenant_id,
    l.id,
    CASE WHEN m.i <= 2 THEN 'ig' ELSE 'wa' END,
    CASE WHEN m.i % 2 = 1 THEN 'in' ELSE 'out' END,
    CASE WHEN m.i % 2 = 1 THEN 'lead' ELSE 'human' END,
    (ARRAY[
        'Oi! Vi o post de vocês, como funciona o curso?',
        'Oi! Que bom que se interessou. A turma é presencial em BH, encontros semanais. Posso te mandar os detalhes no WhatsApp?',
        'Pode sim, meu número é esse mesmo',
        'Perfeito. A turma de outubro começa dia 20, são 12 encontros. Te mando o material e a gente marca uma conversa?'
    ])[m.i],
    'demo-' || substr(md5('sift-demo-msg-' || l.id::text || '-' || m.i), 1, 20),
    l.created_at + (m.i * interval '5 hours')
FROM sift.lead l
JOIN sift.stage s ON s.id = l.stage_id
CROSS JOIN generate_series(1, 4) AS m(i)
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001'
  AND s.key IN ('in_conversation','escalated','proposal','enrolled','lost','nurturing')
  -- as duas primeiras são do Instagram e valem para todos; as do WhatsApp só
  -- para quem já tem telefone
  AND (m.i <= 2 OR l.phone <> '')
ON CONFLICT DO NOTHING;
\echo '  Mensagens criadas (Instagram + WhatsApp na mesma thread)'

-- ── Escalações ───────────────────────────────────────────────────────────────
-- Metade com resumo gerado, metade sem — a tela precisa mostrar as duas, porque
-- o handoff nunca pode depender do LLM estar de pé.

INSERT INTO sift.escalation (id, tenant_id, lead_id, from_user_id, to_user_id, escalated_at, summary_generated, summary_text)
SELECT
    md5('sift-demo-esc-' || l.id::text)::uuid,
    l.tenant_id,
    l.id,
    l.captured_by,
    l.assigned_to,
    l.created_at + interval '2 days',
    (row_number() OVER (ORDER BY l.id)) % 2 = 0,
    CASE WHEN (row_number() OVER (ORDER BY l.id)) % 2 = 0
         THEN 'Interessada na turma de outubro. Trabalha como ' ||
              split_part(l.fit_reason, ',', 1) ||
              '. Já recebeu valor e formato; objeção foi o horário dos encontros. Parou aguardando confirmação de turma noturna.'
         ELSE '' END
FROM sift.lead l
JOIN sift.stage s ON s.id = l.stage_id
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001'
  AND s.key IN ('escalated','proposal','enrolled')
ON CONFLICT DO NOTHING;
\echo '  Escalações criadas'

-- ── Timeline de eventos ──────────────────────────────────────────────────────
-- É o que dá tempo-até-escalar e tempo-em-estágio sem cálculo posterior.

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-created-' || l.id::text)::uuid, l.tenant_id, l.id, 'created',
       'Lead registrado', l.captured_by,
       jsonb_build_object('source', l.source), l.created_at
FROM sift.lead l WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001'
ON CONFLICT DO NOTHING;

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-fit-' || l.id::text)::uuid, l.tenant_id, l.id, 'fit_scored',
       'Fit avaliado: ' || l.fit_score, NULL,
       jsonb_build_object('score', l.fit_score, 'reason', l.fit_reason, 'model', 'demo'),
       l.created_at + interval '20 minutes'
FROM sift.lead l
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001' AND l.fit_score IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-stage-' || l.id::text)::uuid, l.tenant_id, l.id, 'stage_changed',
       'Movido para ' || s.name, l.assigned_to,
       jsonb_build_object('from', 'captured', 'to', s.key),
       l.created_at + interval '1 day'
FROM sift.lead l
JOIN sift.stage s ON s.id = l.stage_id
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001' AND s.key <> 'captured'
ON CONFLICT DO NOTHING;

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-esc-' || e.lead_id::text)::uuid, e.tenant_id, e.lead_id, 'escalated',
       'Escalado para fechamento', e.from_user_id,
       jsonb_build_object('from_user_id', e.from_user_id, 'to_user_id', e.to_user_id),
       e.escalated_at
FROM sift.escalation e
WHERE e.tenant_id = '20000000-0000-0000-0000-000000000001'
ON CONFLICT DO NOTHING;

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-lost-' || l.id::text)::uuid, l.tenant_id, l.id, 'lost',
       'Perdido: ' || l.lost_reason, l.assigned_to,
       jsonb_build_object('reason', l.lost_reason),
       l.created_at + interval '5 days'
FROM sift.lead l
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001' AND l.lost_reason IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO sift.lead_event (id, tenant_id, lead_id, type, body, actor_user_id, metadata, created_at)
SELECT md5('sift-demo-ev-turma-' || l.id::text)::uuid, l.tenant_id, l.id, 'turma_linked',
       'Vinculado a ' || t.name, l.assigned_to,
       jsonb_build_object('turma_id', t.id),
       l.created_at + interval '2 hours'
FROM sift.lead l
JOIN sift.turma t ON t.id = l.turma_id
WHERE l.tenant_id = '20000000-0000-0000-0000-000000000001'
ON CONFLICT DO NOTHING;
\echo '  Timeline de eventos criada'

-- ── Número de WhatsApp da operação ───────────────────────────────────────────

INSERT INTO sift.wa_number (phone_number_id, tenant_id, label)
VALUES ('demo-phone-number-id-0001', '20000000-0000-0000-0000-000000000001', 'Comercial (demonstração)')
ON CONFLICT DO NOTHING;
\echo '  Número de WhatsApp mapeado'

\echo 'ETAPA 27 concluída com sucesso!'
