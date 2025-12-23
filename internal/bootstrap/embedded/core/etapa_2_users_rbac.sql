-- ============================================================================
-- ETAPA 2: Users + RBAC
-- ============================================================================
-- Cria tabelas de usuários e sistema RBAC (roles, permissions)
-- ============================================================================

-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public

-- Garante que estamos no schema correto

SET search_path TO public;

\echo '🚀 ETAPA 2: Criando usuários e sistema RBAC...'

-- Usuários
CREATE TABLE IF NOT EXISTS "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Identificação (email case-insensitive)
    email CITEXT UNIQUE NOT NULL,
    name TEXT,
    last_name TEXT,
    password_hash TEXT,

    -- Contato
    phone TEXT,
    avatar_url TEXT,

    -- Estado
    status TEXT,
    force_password_reset BOOLEAN DEFAULT false,
    last_login TIMESTAMPTZ,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela user criada'

-- Roles (papéis do sistema)
CREATE TABLE IF NOT EXISTS role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT,

    is_system_role BOOLEAN DEFAULT false,

    -- Hierarquia
    parent_role_id UUID REFERENCES role(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela role criada'

-- Permissions (permissões granulares)
CREATE TABLE IF NOT EXISTS permission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT,

    -- Agrupamento
    category TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela permission criada'

-- Role ↔ Permission (N:N)
CREATE TABLE IF NOT EXISTS role_permission (
    role_id UUID NOT NULL REFERENCES role(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permission(id) ON DELETE CASCADE,

    value BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id)
);

\echo '  ✅ Tabela role_permission criada'
\echo '✨ ETAPA 2 concluída com sucesso!'
