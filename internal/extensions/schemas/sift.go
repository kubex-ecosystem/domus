package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/extensions"
)

func init() {
	extensions.Global.Register(&SiftSchema{})
}

// SiftSchema espelha `bootstrap/embedded/core/etapa_26_sift.sql`.
//
// Os dois existem porque há dois caminhos vivos de DDL no domus, e um schema
// que só cobre um deles nasce ausente para metade dos consumidores:
//
//   - os backends (dockerstack/remotestack) leem `bootstrap.manifest.json` e
//     aplicam os `etapa_*.sql` — é o caminho do `domus database migrate`;
//   - `extensions.Global.ApplyAll` (cmd/cli/database.go) e `Apply`
//     (client/client.go) aplicam estas extensões Go.
//
// Precedente: `cozinha_pass.go` e `etapa_25_cozinha_pass.sql` são mantidos em
// sincronia pelo mesmo motivo. A duplicação é deliberada e tem um custo real —
// as duas cópias podem divergir em silêncio. Ao mexer numa, mexa na outra.
//
// UMA DIFERENÇA QUE NÃO É DESCUIDO: os triggers de `updated_at` ficam só no
// `.sql`, como já acontece em `cozinha_pass.go` e `limalar.go`. Aqui não há
// como declarar dependência da etapa 7 (que cria `public.update_updated_at_column`),
// e um `CREATE TRIGGER` sobre função inexistente derrubaria a extensão inteira.
// Por isso os stores do Sift gravam `updated_at = NOW()` explicitamente no
// UPDATE: a corretude não depende de qual caminho de DDL rodou.
type SiftSchema struct{}

func (s *SiftSchema) ID() string {
	return "sift_schema"
}

func (s *SiftSchema) Dependencies() []string {
	return nil
}

func (s *SiftSchema) Statements() []string {
	return []string{
		`CREATE SCHEMA IF NOT EXISTS sift`,

		// Estágio é dado, não código: a operação reordena, renomeia e insere sem
		// deploy. `key` não tem lista fechada de propósito — só formato, para
		// continuar utilizável como identificador no código.
		`CREATE TABLE IF NOT EXISTS sift.stage (
			id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id   UUID        NOT NULL,
			key         TEXT        NOT NULL CHECK (key ~ '^[a-z][a-z0-9_]*$'),
			name        TEXT        NOT NULL,
			order_index INTEGER     NOT NULL DEFAULT 0,
			color       TEXT        NOT NULL DEFAULT '',
			is_terminal BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Chave natural do estágio dentro do tenant — é por ela que o seed
		// idempotente da camada de aplicação faz upsert.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_stage_tenant_key
			ON sift.stage (tenant_id, key)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_stage_tenant_order
			ON sift.stage (tenant_id, order_index ASC)`,

		`CREATE TABLE IF NOT EXISTS sift.turma (
			id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id        UUID          NOT NULL,
			name             TEXT          NOT NULL,
			period_start     DATE,
			period_end       DATE,
			monthly_price    NUMERIC(12,2) CHECK (monthly_price IS NULL OR monthly_price >= 0),
			status           TEXT          NOT NULL DEFAULT 'draft'
				CHECK (status IN ('draft','open','closed','finished')),
			landing_slug     TEXT          NOT NULL CHECK (landing_slug ~ '^[a-z0-9][a-z0-9-]*$'),
			landing_template TEXT          NOT NULL DEFAULT 'essencial'
				CHECK (landing_template IN ('essencial','vitrine','cinematica')),
			created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			CONSTRAINT sift_turma_periodo_coerente
				CHECK (period_start IS NULL OR period_end IS NULL OR period_end >= period_start)
		)`,

		// UNIQUE GLOBAL, não por tenant: /t/:slug é anônima, e o slug é a única
		// informação que o servidor tem para descobrir de quem é a turma.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_turma_slug
			ON sift.turma (landing_slug)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_turma_tenant_status
			ON sift.turma (tenant_id, status, period_start DESC)`,

		// `source` é deliberadamente nullable: a ferramenta suporta a
		// possibilidade de duas origens, não exige as duas. Obrigar
		// classificação em toda entrada é atrito, e atrito derruba a adesão.
		`CREATE TABLE IF NOT EXISTS sift.lead (
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
			CONSTRAINT sift_lead_tem_contato
				CHECK (ig_handle <> '' OR phone <> '' OR email <> '')
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_stage
			ON sift.lead (tenant_id, stage_id, updated_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_created
			ON sift.lead (tenant_id, created_at DESC)`,

		// Dedupe por handle: o mesmo perfil não deve virar dois cards.
		// Case-insensitive porque o Instagram trata @Fulana e @fulana como o
		// mesmo perfil.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_lead_tenant_handle
			ON sift.lead (tenant_id, lower(ig_handle)) WHERE ig_handle <> ''`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_lead_tenant_phone
			ON sift.lead (tenant_id, phone) WHERE phone <> ''`,

		// Sustenta a leitura por origem — que só produz comparação quando há
		// mais de uma origem preenchida, e degrada para métrica de um funil só
		// quando não há.
		`CREATE INDEX IF NOT EXISTS idx_sift_lead_tenant_source
			ON sift.lead (tenant_id, source, created_at DESC) WHERE source IS NOT NULL`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_turma
			ON sift.lead (turma_id) WHERE turma_id IS NOT NULL`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_assigned
			ON sift.lead (tenant_id, assigned_to) WHERE assigned_to IS NOT NULL`,

		// Aqui a lista É fechada: tipo de evento é vocabulário do código, não
		// configuração da operação.
		`CREATE TABLE IF NOT EXISTS sift.lead_event (
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
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_event_lead
			ON sift.lead_event (lead_id, created_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_lead_event_tenant_type
			ON sift.lead_event (tenant_id, type, created_at DESC)`,

		`CREATE TABLE IF NOT EXISTS sift.message (
			id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id   UUID        NOT NULL,
			lead_id     UUID        NOT NULL REFERENCES sift.lead(id) ON DELETE CASCADE,
			channel     TEXT        NOT NULL CHECK (channel IN ('ig','wa')),
			direction   TEXT        NOT NULL CHECK (direction IN ('in','out')),
			author      TEXT        NOT NULL DEFAULT 'lead' CHECK (author IN ('lead','ai','human')),
			content     TEXT        NOT NULL DEFAULT '',
			external_id TEXT        NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_message_lead
			ON sift.message (lead_id, created_at ASC)`,

		// Idempotência do webhook: a Meta reentrega evento, e sem isto a mesma
		// mensagem vira várias linhas na timeline.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_message_external
			ON sift.message (channel, external_id) WHERE external_id <> ''`,

		// `summary_generated` distingue "não tem resumo" de "resumo vazio": o
		// handoff cai para o histórico cru quando o LLM não está de pé.
		`CREATE TABLE IF NOT EXISTS sift.escalation (
			id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id         UUID        NOT NULL,
			lead_id           UUID        NOT NULL REFERENCES sift.lead(id) ON DELETE CASCADE,
			from_user_id      UUID,
			to_user_id        UUID,
			escalated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			summary_generated BOOLEAN     NOT NULL DEFAULT FALSE,
			summary_text      TEXT        NOT NULL DEFAULT ''
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_escalation_lead
			ON sift.escalation (lead_id, escalated_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_escalation_tenant_to
			ON sift.escalation (tenant_id, to_user_id, escalated_at DESC)`,

		// `last_window_at` é o carimbo da última mensagem recebida: é ele que
		// diz se a janela de resposta inbound ainda está aberta.
		`CREATE TABLE IF NOT EXISTS sift.ig_profile (
			lead_id        UUID        PRIMARY KEY REFERENCES sift.lead(id) ON DELETE CASCADE,
			tenant_id      UUID        NOT NULL,
			handle         TEXT        NOT NULL DEFAULT '',
			ig_user_id     TEXT        NOT NULL DEFAULT '',
			last_window_at TIMESTAMPTZ,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sift_ig_profile_user
			ON sift.ig_profile (tenant_id, ig_user_id) WHERE ig_user_id <> ''`,

		// Preenche a lacuna da etapa 13: whatsapp_messages é log bruto global,
		// sem tenant_id. Sem este mapa, mensagem que chega no webhook não tem
		// como ser atribuída a um tenant.
		`CREATE TABLE IF NOT EXISTS sift.wa_number (
			phone_number_id TEXT        PRIMARY KEY,
			tenant_id       UUID        NOT NULL,
			label           TEXT        NOT NULL DEFAULT '',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sift_wa_number_tenant
			ON sift.wa_number (tenant_id)`,
	}
}
