SELECT seed_utils.set_search_path('public');
-- ============================================================
-- SEED KUBEX - V2.0 (Full Data - Realistic Scenario)
-- ============================================================
-- Cenário completo com múltiplas orgs, tenants, usuários,
-- roles, permissions, teams, leads e estrutura de parceiros
-- ============================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
-- Garante que estamos no schema correto
SET search_path TO public;
-- ================
INSERT INTO org (id, name, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000001',
    'Kubex Holdings',
    now() - interval '6 months'
  ),
  (
    '20000000-0000-0000-0000-000000000002',
    'Tech Partners Network',
    now() - interval '4 months'
  ),
  (
    '20000000-0000-0000-0000-000000000003',
    'Enterprise Solutions Group',
    now() - interval '2 months'
  ) ON CONFLICT (id) DO NOTHING;
-- ================
INSERT INTO tenant (
    id,
    org_id,
    name,
    slug,
    domain,
    plan,
    is_active,
    is_trial,
    trial_ends_at,
    created_at
  )
VALUES (
    '30000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Kubex HQ',
    'gnyx-hq',
    'hq.kubex.world',
    seed_utils.get_mapped_uuid('enterprise'),
    true,
    false,
    null,
    now() - interval '6 months'
  ),
  (
    '30000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000001',
    'Kubex Sales',
    'gnyx-sales',
    'sales.kubex.world',
    seed_utils.get_mapped_uuid('professional'),
    true,
    false,
    null,
    now() - interval '5 months'
  ),
  (
    '30000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000002',
    'TechPartners Brazil',
    'techpartners-br',
    'br.techpartners.com',
    seed_utils.get_mapped_uuid('professional'),
    true,
    false,
    null,
    now() - interval '4 months'
  ),
  (
    '30000000-0000-0000-0000-000000000004',
    '20000000-0000-0000-0000-000000000002',
    'TechPartners LATAM',
    'techpartners-latam',
    'latam.techpartners.com',
    'starter',
    true,
    true,
    now() + interval '1 month',
    now() - interval '3 weeks'
  ),
  (
    '30000000-0000-0000-0000-000000000005',
    '20000000-0000-0000-0000-000000000003',
    'Enterprise Demo',
    'enterprise-demo',
    'demo.enterprise.io',
    seed_utils.get_mapped_uuid('enterprise'),
    true,
    false,
    null,
    now() - interval '2 months'
  ) ON CONFLICT (id) DO NOTHING;
-- ================
-- ROLES
-- ================
INSERT INTO role (
    id,
    code,
    display_name,
    description,
    is_system_role,
    parent_role_id,
    created_at
  )
VALUES -- System Roles
  (
    '40000000-0000-0000-0000-000000000001',
    'super_admin',
    'Super Admin',
    'Acesso total ao sistema',
    true,
    null,
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000002',
    'admin',
    seed_utils.get_mapped_uuid('Administrador'),
    'Administrador do tenant',
    true,
    '40000000-0000-0000-0000-000000000001',
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000003',
    'manager',
    'Gerente',
    'Gerente de equipe',
    true,
    '40000000-0000-0000-0000-000000000002',
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000004',
    'sales_rep',
    'Vendedor',
    'Representante de vendas',
    true,
    '40000000-0000-0000-0000-000000000003',
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000005',
    'partner_admin',
    'Administrador de Parceiro',
    'Administra parceiros',
    true,
    '40000000-0000-0000-0000-000000000002',
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000006',
    'partner',
    'Parceiro',
    'Parceiro comercial',
    true,
    null,
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000007',
    'analyst',
    'Analista',
    'Analista de dados',
    true,
    '40000000-0000-0000-0000-000000000003',
    now() - interval '6 months'
  ),
  (
    '40000000-0000-0000-0000-000000000008',
    'viewer',
    seed_utils.get_mapped_uuid('Visualizador'),
    'Apenas visualização',
    true,
    null,
    now() - interval '6 months'
  );
-- ================
-- PERMISSIONS
-- ================
INSERT INTO permission (
    id,
    code,
    display_name,
    description,
    category,
    created_at
  )
VALUES -- Dashboard
  (
    '50000000-0000-0000-0000-000000000001',
    'dashboard.view',
    'Visualizar Dashboard',
    'Acesso ao dashboard principal',
    'dashboard',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000002',
    'dashboard.analytics',
    'Analytics Avançado',
    'Acesso a analytics avançado',
    'dashboard',
    now() - interval '6 months'
  ),
  -- Users
  (
    '50000000-0000-0000-0000-000000000003',
    'users.view',
    'Visualizar Usuários',
    'Ver lista de usuários',
    'users',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000004',
    'users.create',
    'Criar Usuários',
    'Criar novos usuários',
    'users',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000005',
    'users.edit',
    'Editar Usuários',
    'Editar usuários existentes',
    'users',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000006',
    'users.delete',
    'Deletar Usuários',
    'Remover usuários',
    'users',
    now() - interval '6 months'
  ),
  -- Partners
  (
    '50000000-0000-0000-0000-000000000007',
    'partners.view',
    'Visualizar Parceiros',
    'Ver lista de parceiros',
    'partners',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000008',
    'partners.create',
    'Criar Parceiros',
    'Adicionar parceiros',
    'partners',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000009',
    'partners.edit',
    'Editar Parceiros',
    'Modificar parceiros',
    'partners',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000010',
    'partners.delete',
    'Deletar Parceiros',
    'Remover parceiros',
    'partners',
    now() - interval '6 months'
  ),
  -- Leads
  (
    '50000000-0000-0000-0000-000000000011',
    'leads.view',
    'Visualizar Leads',
    'Ver leads',
    'leads',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000012',
    'leads.create',
    'Criar Leads',
    'Adicionar leads',
    'leads',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000013',
    'leads.edit',
    'Editar Leads',
    'Modificar leads',
    'leads',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000014',
    'leads.delete',
    'Deletar Leads',
    'Remover leads',
    'leads',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000015',
    'leads.assign',
    'Atribuir Leads',
    'Distribuir leads',
    'leads',
    now() - interval '6 months'
  ),
  -- Pipelines
  (
    '50000000-0000-0000-0000-000000000016',
    'pipelines.view',
    'Visualizar Pipelines',
    'Ver pipelines',
    'pipelines',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000017',
    'pipelines.manage',
    'Gerenciar Pipelines',
    'Criar e editar pipelines',
    'pipelines',
    now() - interval '6 months'
  ),
  -- Commissions
  (
    '50000000-0000-0000-0000-000000000018',
    'commissions.view',
    'Visualizar Comissões',
    'Ver comissões',
    seed_utils.get_mapped_uuid('commissions'),
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000019',
    'commissions.approve',
    'Aprovar Comissões',
    'Aprovar pagamentos',
    seed_utils.get_mapped_uuid('commissions'),
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000020',
    'commissions.manage',
    'Gerenciar Comissões',
    'Configurar comissões',
    seed_utils.get_mapped_uuid('commissions'),
    now() - interval '6 months'
  ),
  -- Settings
  (
    '50000000-0000-0000-0000-000000000021',
    'settings.view',
    'Visualizar Configurações',
    'Ver configurações',
    'settings',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000022',
    'settings.edit',
    'Editar Configurações',
    'Modificar configurações',
    'settings',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000023',
    'settings.security',
    'Configurações de Segurança',
    'Gerenciar segurança',
    'settings',
    now() - interval '6 months'
  ),
  -- Teams
  (
    '50000000-0000-0000-0000-000000000024',
    'teams.view',
    'Visualizar Times',
    'Ver times',
    'teams',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000025',
    'teams.manage',
    'Gerenciar Times',
    'Criar e editar times',
    'teams',
    now() - interval '6 months'
  ),
  -- Reports
  (
    '50000000-0000-0000-0000-000000000026',
    'reports.view',
    'Visualizar Relatórios',
    'Ver relatórios',
    'reports',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000027',
    'reports.export',
    'Exportar Relatórios',
    'Baixar relatórios',
    'reports',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000028',
    'reports.advanced',
    'Relatórios Avançados',
    'Relatórios customizados',
    'reports',
    now() - interval '6 months'
  ),
  -- Invitations
  (
    '50000000-0000-0000-0000-000000000029',
    'invites.send',
    'Enviar Convites',
    'Convidar usuários',
    'invites',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000030',
    'invites.manage',
    'Gerenciar Convites',
    'Administrar convites',
    'invites',
    now() - interval '6 months'
  ),
  -- System
  (
    '50000000-0000-0000-0000-000000000031',
    'system.admin',
    'Administração Sistema',
    'Acesso total ao sistema',
    'system',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000032',
    'system.logs',
    'Visualizar Logs',
    'Ver logs do sistema',
    'system',
    now() - interval '6 months'
  ),
  (
    '50000000-0000-0000-0000-000000000033',
    'system.billing',
    'Gerenciar Cobrança',
    'Administrar billing',
    'system',
    now() - interval '6 months'
  );
-- ================
-- ROLE PERMISSIONS
-- ================
-- Super Admin: todas as permissões
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000001',
  id,
  true,
  now()
FROM permission;
-- Admin: quase todas exceto system.admin
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000002',
  id,
  true,
  now()
FROM permission
WHERE code != 'system.admin';
-- Manager: gestão de equipe e leads
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000003',
  id,
  true,
  now()
FROM permission
WHERE category IN (
    'dashboard',
    'leads',
    'partners',
    'teams',
    'reports',
    'users'
  )
  AND code NOT IN (
    'users.delete',
    'partners.delete',
    'leads.delete'
  );
-- Sales Rep: apenas leads e dashboard
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000004',
  id,
  true,
  now()
FROM permission
WHERE category IN ('dashboard', 'leads')
  AND code NOT IN ('leads.delete');
-- Partner Admin: gestão de parceiros
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000005',
  id,
  true,
  now()
FROM permission
WHERE category IN ('dashboard', 'partners', 'leads', 'reports')
  AND code NOT IN ('partners.delete', 'leads.delete');
-- Partner: visualização apenas
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000006',
  id,
  true,
  now()
FROM permission
WHERE code IN ('dashboard.view', 'leads.view', 'reports.view');
-- Analyst: relatórios e visualização
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000007',
  id,
  true,
  now()
FROM permission
WHERE category IN ('dashboard', 'reports')
  OR code IN ('leads.view', 'partners.view', 'users.view');
-- Viewer: apenas visualização básica
INSERT INTO role_permission (role_id, permission_id, value, created_at)
SELECT '40000000-0000-0000-0000-000000000008',
  id,
  true,
  now()
FROM permission
WHERE code IN ('dashboard.view');
-- ================
-- USERS
-- ================
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    phone,
    status,
    created_at
  )
VALUES -- Kubex HQ Users
  (
    '60000000-0000-0000-0000-000000000001',
    'rafael@kubex.world',
    'Rafael',
    'Mori',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4321',
    'active',
    now() - interval '6 months'
  ),
  (
    '60000000-0000-0000-0000-000000000002',
    'thiago@kubex.world',
    'Thiago',
    'Silva',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4322',
    'active',
    now() - interval '6 months'
  ),
  (
    '60000000-0000-0000-0000-000000000003',
    'maria@kubex.world',
    'Maria',
    'Santos',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4323',
    'active',
    now() - interval '5 months'
  ),
  (
    '60000000-0000-0000-0000-000000000004',
    'joao@kubex.world',
    'João',
    'Oliveira',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4324',
    'active',
    now() - interval '5 months'
  ),
  (
    '60000000-0000-0000-0000-000000000005',
    'ana@kubex.world',
    'Ana',
    'Costa',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4325',
    'active',
    now() - interval '4 months'
  ),
  -- TechPartners Users
  (
    '60000000-0000-0000-0000-000000000006',
    'carlos@techpartners.com',
    'Carlos',
    'Ferreira',
    crypt('tech123', gen_salt('bf')),
    '+55 21 98765-1111',
    'active',
    now() - interval '4 months'
  ),
  (
    '60000000-0000-0000-0000-000000000007',
    'beatriz@techpartners.com',
    'Beatriz',
    'Lima',
    crypt('tech123', gen_salt('bf')),
    '+55 21 98765-1112',
    'active',
    now() - interval '3 months'
  ),
  (
    '60000000-0000-0000-0000-000000000008',
    'pedro@techpartners.com',
    'Pedro',
    'Rodrigues',
    crypt('tech123', gen_salt('bf')),
    '+55 21 98765-1113',
    'active',
    now() - interval '3 months'
  ),
  -- Enterprise Demo Users
  (
    '60000000-0000-0000-0000-000000000009',
    'admin@enterprise.io',
    'Demo',
    seed_utils.get_mapped_uuid('Administrator'),
    crypt('demo123', gen_salt('bf')),
    '+1 555-0100',
    'active',
    now() - interval '2 months'
  ),
  (
    '60000000-0000-0000-0000-000000000010',
    'sales@enterprise.io',
    'Demo',
    'Sales Rep',
    crypt('demo123', gen_salt('bf')),
    '+1 555-0101',
    'active',
    now() - interval '2 months'
  );
-- ================
-- TENANT MEMBERSHIPS (USER ↔ TENANT ↔ ROLE)
-- ================
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES -- Kubex HQ
  (
    '60000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    true,
    now() - interval '6 months'
  ),
  -- Rafael: Super Admin
  (
    '60000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    true,
    now() - interval '6 months'
  ),
  -- Thiago: Admin
  (
    '60000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000003',
    true,
    now() - interval '5 months'
  ),
  -- Maria: Manager
  (
    '60000000-0000-0000-0000-000000000004',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000004',
    true,
    now() - interval '5 months'
  ),
  -- João: Sales Rep
  (
    '60000000-0000-0000-0000-000000000005',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000007',
    true,
    now() - interval '4 months'
  ),
  -- Ana: Analyst
  -- Kubex Sales
  (
    '60000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000002',
    true,
    now() - interval '5 months'
  ),
  -- Maria: Admin
  (
    '60000000-0000-0000-0000-000000000004',
    '30000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000004',
    true,
    now() - interval '5 months'
  ),
  -- João: Sales Rep
  -- TechPartners Brazil
  (
    '60000000-0000-0000-0000-000000000006',
    '30000000-0000-0000-0000-000000000003',
    '40000000-0000-0000-0000-000000000002',
    true,
    now() - interval '4 months'
  ),
  -- Carlos: Admin
  (
    '60000000-0000-0000-0000-000000000007',
    '30000000-0000-0000-0000-000000000003',
    '40000000-0000-0000-0000-000000000005',
    true,
    now() - interval '3 months'
  ),
  -- Beatriz: Partner Admin
  (
    '60000000-0000-0000-0000-000000000008',
    '30000000-0000-0000-0000-000000000003',
    '40000000-0000-0000-0000-000000000006',
    true,
    now() - interval '3 months'
  ),
  -- Pedro: Partner
  -- TechPartners LATAM
  (
    '60000000-0000-0000-0000-000000000006',
    '30000000-0000-0000-0000-000000000004',
    '40000000-0000-0000-0000-000000000003',
    true,
    now() - interval '3 weeks'
  ),
  -- Carlos: Manager
  -- Enterprise Demo
  (
    '60000000-0000-0000-0000-000000000009',
    '30000000-0000-0000-0000-000000000005',
    '40000000-0000-0000-0000-000000000002',
    true,
    now() - interval '2 months'
  ),
  -- Admin
  (
    '60000000-0000-0000-0000-000000000010',
    '30000000-0000-0000-0000-000000000005',
    '40000000-0000-0000-0000-000000000004',
    true,
    now() - interval '2 months'
  );
-- Sales Rep
-- ================
-- TEAMS
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
VALUES -- Kubex HQ Teams
  (
    '70000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    'Comercial',
    'Equipe de vendas principal',
    true,
    true,
    '60000000-0000-0000-0000-000000000001',
    now() - interval '6 months'
  ),
  (
    '70000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    'Parcerias',
    'Equipe de gestão de parceiros',
    false,
    true,
    '60000000-0000-0000-0000-000000000001',
    now() - interval '5 months'
  ),
  (
    '70000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000001',
    'Suporte',
    'Equipe de suporte ao cliente',
    false,
    true,
    '60000000-0000-0000-0000-000000000002',
    now() - interval '4 months'
  ),
  -- Kubex Sales Teams
  (
    '70000000-0000-0000-0000-000000000004',
    '30000000-0000-0000-0000-000000000002',
    'Vendas Corporativas',
    'Vendas B2B',
    true,
    true,
    '60000000-0000-0000-0000-000000000003',
    now() - interval '5 months'
  ),
  -- TechPartners Teams
  (
    '70000000-0000-0000-0000-000000000005',
    '30000000-0000-0000-0000-000000000003',
    'Canal de Vendas',
    'Parceiros de venda',
    true,
    true,
    '60000000-0000-0000-0000-000000000006',
    now() - interval '4 months'
  ),
  (
    '70000000-0000-0000-0000-000000000006',
    '30000000-0000-0000-0000-000000000003',
    seed_utils.get_mapped_uuid('Integradores'),
    'Parceiros técnicos',
    false,
    true,
    '60000000-0000-0000-0000-000000000006',
    now() - interval '3 months'
  ),
  -- Enterprise Demo Teams
  (
    '70000000-0000-0000-0000-000000000007',
    '30000000-0000-0000-0000-000000000005',
    'Demo Team',
    'Equipe de demonstração',
    true,
    true,
    '60000000-0000-0000-0000-000000000009',
    now() - interval '2 months'
  );
-- ================
-- TEAM MEMBERSHIPS
-- ================
INSERT INTO team_membership (user_id, team_id, role_id, is_active, created_at)
VALUES -- Kubex HQ Comercial
  (
    '60000000-0000-0000-0000-000000000003',
    '70000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000003',
    true,
    now() - interval '5 months'
  ),
  -- Maria (Manager)
  (
    '60000000-0000-0000-0000-000000000004',
    '70000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000004',
    true,
    now() - interval '5 months'
  ),
  -- João (Sales Rep)
  -- Kubex HQ Parcerias
  (
    '60000000-0000-0000-0000-000000000002',
    '70000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000005',
    true,
    now() - interval '5 months'
  ),
  -- Thiago (Partner Admin)
  -- TechPartners Canal de Vendas
  (
    '60000000-0000-0000-0000-000000000007',
    '70000000-0000-0000-0000-000000000005',
    '40000000-0000-0000-0000-000000000005',
    true,
    now() - interval '3 months'
  ),
  -- Beatriz
  (
    '60000000-0000-0000-0000-000000000008',
    '70000000-0000-0000-0000-000000000005',
    '40000000-0000-0000-0000-000000000006',
    true,
    now() - interval '3 months'
  );
-- Pedro
-- ================
-- PIPELINES
-- ================
INSERT INTO pipeline (
    id,
    tenant_id,
    name,
    description,
    is_default,
    created_at
  )
VALUES (
    '80000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    'Vendas B2B',
    'Pipeline principal de vendas corporativas',
    true,
    now() - interval '6 months'
  ),
  (
    '80000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    'Parcerias',
    'Pipeline de aquisição de parceiros',
    false,
    now() - interval '5 months'
  ),
  (
    '80000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'Canal TechPartners',
    'Pipeline de vendas através de parceiros',
    true,
    now() - interval '4 months'
  );
-- ================
-- PIPELINE STAGES
-- ================
INSERT INTO pipeline_stage (id, pipeline_id, name, order_index, created_at)
VALUES -- Vendas B2B Stages
  (
    '90000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    'Prospecção',
    1,
    now() - interval '6 months'
  ),
  (
    '90000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000001',
    'Qualificação',
    2,
    now() - interval '6 months'
  ),
  (
    '90000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000001',
    'Proposta',
    3,
    now() - interval '6 months'
  ),
  (
    '90000000-0000-0000-0000-000000000004',
    '80000000-0000-0000-0000-000000000001',
    'Negociação',
    4,
    now() - interval '6 months'
  ),
  (
    '90000000-0000-0000-0000-000000000005',
    '80000000-0000-0000-0000-000000000001',
    'Fechado Ganho',
    5,
    now() - interval '6 months'
  ),
  (
    '90000000-0000-0000-0000-000000000006',
    '80000000-0000-0000-0000-000000000001',
    'Fechado Perdido',
    6,
    now() - interval '6 months'
  ),
  -- Parcerias Stages
  (
    '90000000-0000-0000-0000-000000000007',
    '80000000-0000-0000-0000-000000000002',
    'Primeiro Contato',
    1,
    now() - interval '5 months'
  ),
  (
    '90000000-0000-0000-0000-000000000008',
    '80000000-0000-0000-0000-000000000002',
    'Análise de Fit',
    2,
    now() - interval '5 months'
  ),
  (
    '90000000-0000-0000-0000-000000000009',
    '80000000-0000-0000-0000-000000000002',
    'Proposta de Parceria',
    3,
    now() - interval '5 months'
  ),
  (
    '90000000-0000-0000-0000-000000000010',
    '80000000-0000-0000-0000-000000000002',
    'Parceiro Ativo',
    4,
    now() - interval '5 months'
  ),
  -- Canal TechPartners Stages
  (
    '90000000-0000-0000-0000-000000000011',
    '80000000-0000-0000-0000-000000000003',
    'Lead Recebido',
    1,
    now() - interval '4 months'
  ),
  (
    '90000000-0000-0000-0000-000000000012',
    '80000000-0000-0000-0000-000000000003',
    'Em Atendimento',
    2,
    now() - interval '4 months'
  ),
  (
    '90000000-0000-0000-0000-000000000013',
    '80000000-0000-0000-0000-000000000003',
    'Proposta Enviada',
    3,
    now() - interval '4 months'
  ),
  (
    '90000000-0000-0000-0000-000000000014',
    '80000000-0000-0000-0000-000000000003',
    'Fechado',
    4,
    now() - interval '4 months'
  );
-- ================
-- PARTNERS
-- ================
INSERT INTO partner (
    id,
    tenant_id,
    name,
    email,
    phone,
    cnpj,
    tier,
    status,
    primary_contact_user_id,
    created_at
  )
VALUES (
    'A0000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    'TechSolutions Brasil',
    'contato@techsolutions.com.br',
    '+55 11 3333-4444',
    seed_utils.get_mapped_uuid('12345678000100'),
    'platinum',
    'active',
    '60000000-0000-0000-0000-000000000008',
    now() - interval '4 months'
  ),
  (
    'A0000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    'Innovate Partners',
    'hello@innovatepartners.io',
    '+55 21 4444-5555',
    seed_utils.get_mapped_uuid('23456789000111'),
    'gold',
    'active',
    '60000000-0000-0000-0000-000000000007',
    now() - interval '3 months'
  ),
  (
    'A0000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'Channel Master LATAM',
    'latam@channelmaster.com',
    '+52 55 1234-5678',
    seed_utils.get_mapped_uuid('34567890000122'),
    'silver',
    'active',
    null,
    now() - interval '2 months'
  ),
  (
    'A0000000-0000-0000-0000-000000000004',
    '30000000-0000-0000-0000-000000000001',
    'StartupHub Network',
    'network@startuphub.io',
    '+55 11 5555-6666',
    seed_utils.get_mapped_uuid('45678901000133'),
    'bronze',
    'pending',
    null,
    now() - interval '1 month'
  ),
  (
    'A0000000-0000-0000-0000-000000000005',
    '30000000-0000-0000-0000-000000000003',
    'Enterprise Integrations',
    'sales@enterpriseint.com',
    '+55 47 6666-7777',
    seed_utils.get_mapped_uuid('56789012000144'),
    'gold',
    'active',
    null,
    now() - interval '6 weeks'
  );
-- ================
-- LEADS
-- ================
INSERT INTO lead (
    id,
    tenant_id,
    pipeline_id,
    stage_id,
    company_name,
    contact_name,
    contact_email,
    contact_phone,
    value,
    status,
    assigned_to,
    partner_id,
    created_at
  )
VALUES -- Kubex HQ Leads
  (
    'B0000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000003',
    'Acme Corporation',
    'John Doe',
    'john@acme.com',
    '+1 555-1234',
    150000.00,
    'in_progress',
    '60000000-0000-0000-0000-000000000004',
    null,
    now() - interval '2 weeks'
  ),
  (
    'B0000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000004',
    'Global Tech Inc',
    'Jane Smith',
    'jane@globaltech.com',
    '+1 555-5678',
    250000.00,
    seed_utils.get_mapped_uuid('negotiating'),
    '60000000-0000-0000-0000-000000000003',
    '0000000-0000-0000-0000-000000000001',
    now() - interval '1 week'
  ),
  (
    'B0000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000005',
    'Innovate Co',
    'Bob Johnson',
    'bob@innovate.co',
    '+1 555-9012',
    75000.00,
    'won',
    '60000000-0000-0000-0000-000000000004',
    null,
    now() - interval '3 weeks'
  ),
  (
    'B0000000-0000-0000-0000-000000000004',
    '30000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000002',
    'Enterprise Solutions',
    'Alice Brown',
    'alice@enterprise.com',
    '+1 555-3456',
    500000.00,
    'qualified',
    '60000000-0000-0000-0000-000000000003',
    '0000000-0000-0000-0000-000000000002',
    now() - interval '4 days'
  ),
  (
    'B0000000-0000-0000-0000-000000000005',
    '30000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000006',
    'Small Biz Ltd',
    'Charlie Davis',
    'charlie@smallbiz.com',
    '+1 555-7890',
    25000.00,
    'lost',
    '60000000-0000-0000-0000-000000000004',
    null,
    now() - interval '1 month'
  ),
  -- TechPartners Leads
  (
    'B0000000-0000-0000-0000-000000000006',
    '30000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    '90000000-0000-0000-0000-000000000012',
    'LATAM Corp',
    'Diego Martinez',
    'diego@latamcorp.com',
    '+52 55 1111-2222',
    180000.00,
    'in_progress',
    '60000000-0000-0000-0000-000000000007',
    '0000000-0000-0000-0000-000000000003',
    now() - interval '5 days'
  ),
  (
    'B0000000-0000-0000-0000-000000000007',
    '30000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    '90000000-0000-0000-0000-000000000011',
    'Brazil Startups',
    'Fernanda Silva',
    'fernanda@brazilstartups.com.br',
    '+55 11 2222-3333',
    95000.00,
    'new',
    '60000000-0000-0000-0000-000000000008',
    null,
    now() - interval '2 days'
  ),
  (
    'B0000000-0000-0000-0000-000000000008',
    '30000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    '90000000-0000-0000-0000-000000000014',
    'SaaS Enterprise',
    'Gabriel Costa',
    'gabriel@saasent.io',
    '+55 21 3333-4444',
    320000.00,
    'won',
    '60000000-0000-0000-0000-000000000006',
    '0000000-0000-0000-0000-000000000005',
    now() - interval '2 weeks'
  );
-- ============================================================
-- FIM DO SEED V2.0
-- ============================================================