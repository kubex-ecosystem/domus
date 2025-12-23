-- ============================================================================
-- ETAPA 3: Memberships + Foreign Keys Circulares
-- ============================================================================
-- Cria tabelas de memberships e adiciona FKs circulares
-- ============================================================================

-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public

-- Garante que estamos no schema correto

SET search_path TO public;

\echo '🚀 ETAPA 3: Criando memberships e FKs circulares...'

-- Tenant Membership
CREATE TABLE IF NOT EXISTS tenant_membership (
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES role(id),

    is_active BOOLEAN DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,

    PRIMARY KEY (user_id, tenant_id)
);

\echo '  ✅ Tabela tenant_membership criada'

-- Team Membership
CREATE TABLE IF NOT EXISTS team_membership (
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES team(id) ON DELETE CASCADE,

    role_id UUID REFERENCES role(id),

    is_active BOOLEAN DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,

    PRIMARY KEY (user_id, team_id)
);

\echo '  ✅ Tabela team_membership criada'

-- Adicionar FK circular: team.created_by → user
ALTER TABLE team
ADD CONSTRAINT team_created_by_fkey
FOREIGN KEY (created_by) REFERENCES "user"(id) ON DELETE SET NULL;

\echo '  ✅ FK circular team.created_by adicionada'
\echo '✨ ETAPA 3 concluída com sucesso!'
