-- ============================================================
-- SEED KUBEX - V0.1 (Minimal, Realistic, Coherent)
-- ============================================================
-- Seed minimalista para testes básicos
-- Estrutura: 1 org, 1 tenant, 2 users, 3 roles, 6 permissions
-- ============================================================
SET search_path TO public;
-- ================
-- ORG PRINCIPAL
-- ================
INSERT INTO org (id, name, created_at)
VALUES (
    '10000000-0000-0000-0000-000000000001',
    'Kubex PRM',
    now()
  ) ON CONFLICT (id) DO NOTHING;
-- ================
-- TENANT PRINCIPAL
-- ================
INSERT INTO tenant (
    id,
    org_id,
    name,
    slug,
    plan,
    is_active,
    created_at
  )
VALUES (
    '20000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000001',
    'Kubex HQ',
    'gnyx-hq',
    'enterprise',
    true,
    now()
  ) ON CONFLICT (id) DO NOTHING;
-- ================
-- PERMISSIONS (ESSENCIAIS)
-- ================
INSERT INTO permission (
    id,
    code,
    display_name,
    description,
    category,
    created_at
  )
VALUES (
    '30000000-0000-0000-0000-000000000001',
    'dashboard.view',
    'Visualizar Dashboard',
    'Acessar dashboards',
    'dashboard',
    now()
  ),
  (
    '30000000-0000-0000-0000-000000000002',
    'users.manage',
    'Gerenciar Usuários',
    'Gerenciar usuários',
    'users',
    now()
  ),
  (
    '30000000-0000-0000-0000-000000000003',
    'settings.manage',
    'Gerenciar Configurações',
    'Gerenciar configurações',
    'settings',
    now()
  ),
  (
    '30000000-0000-0000-0000-000000000004',
    'partners.view',
    'Visualizar Parceiros',
    'Ver parceiros',
    'partners',
    now()
  ),
  (
    '30000000-0000-0000-0000-000000000005',
    'leads.view',
    'Visualizar Leads',
    'Visualizar leads',
    'leads',
    now()
  ),
  (
    '30000000-0000-0000-0000-000000000006',
    'leads.manage',
    'Gerenciar Leads',
    'Gerenciar leads',
    'leads',
    now()
  ) ON CONFLICT (code) DO NOTHING;
-- ================
-- ROLES BÁSICOS
-- ================
INSERT INTO role (
    id,
    code,
    display_name,
    description,
    is_system_role,
    created_at
  )
VALUES (
    '40000000-0000-0000-0000-000000000001',
    'admin',
    'Administrador',
    'Acesso total',
    true,
    now()
  ),
  (
    '40000000-0000-0000-0000-000000000002',
    'manager',
    'Gestor',
    'Coordenação e supervisão',
    true,
    now()
  ),
  (
    '40000000-0000-0000-0000-000000000003',
    'viewer',
    'Visualizador',
    'Acesso mínimo',
    true,
    now()
  ) ON CONFLICT (code) DO NOTHING;
-- ================
-- ROLE PERMISSIONS
-- ================
-- Admin -> tudo
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000001',
  id,
  true,
  now()
FROM permission ON CONFLICT DO NOTHING;
-- Manager -> subset
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000002',
  id,
  true,
  now()
FROM permission
WHERE code IN ('dashboard.view', 'leads.view', 'partners.view') ON CONFLICT DO NOTHING;
-- Viewer -> mínimo
INSERT INTO role_permission (role_id, permission_id, value, created_at)
VALUES (
    '40000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000001',
    true,
    now()
  ) ON CONFLICT DO NOTHING;
-- ================
-- USUÁRIOS
-- ================
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES -- Usuário principal (Rafael)
  (
    '50000000-0000-0000-0000-000000000001',
    'rafael@gnyx.app',
    'Rafael',
    'Mori',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  ),
  -- Usuário secundário (Thiago)
  (
    '50000000-0000-0000-0000-000000000002',
    'thiago@gnyx.app',
    'Thiago',
    'CTO',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  ) ON CONFLICT (id) DO NOTHING;
-- ================
-- TENANT MEMBERSHIPS (ligação usuário ↔ tenant ↔ role)
-- ================
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '50000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    true,
    now()
  ),
  -- Rafael: admin
  (
    '50000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    true,
    now()
  ) -- Thiago: manager
  ON CONFLICT (user_id, tenant_id) DO NOTHING;
-- ================
-- TEAMS (mínimo necessário)
-- ================
INSERT INTO team (
    id,
    tenant_id,
    name,
    description,
    is_default,
    is_active,
    created_by,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Equipe Comercial',
    'Equipe principal de vendas',
    true,
    true,
    '50000000-0000-0000-0000-000000000001',
    now()
  ) ON CONFLICT (id) DO NOTHING;
-- ================
-- TEAM MEMBERSHIPS
-- ================
INSERT INTO team_membership (user_id, team_id, role_id, is_active, created_at)
VALUES (
    '50000000-0000-0000-0000-000000000001',
    '60000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    true,
    now()
  ),
  (
    '50000000-0000-0000-0000-000000000002',
    '60000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    true,
    now()
  ) ON CONFLICT (user_id, team_id) DO NOTHING;
-- ============================================================
-- FIM DO SEED V0.1
-- ============================================================