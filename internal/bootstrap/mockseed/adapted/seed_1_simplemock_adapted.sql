SELECT seed_utils.set_search_path('public');
-- ============================================================
-- SEED KUBEX - V0.1 (Minimal, Realistic, Coherent)
-- ============================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
-- Garante que estamos no schema correto
SET search_path TO public;
-- ============================================================
-- ================
-- ORG PRINCIPAL
-- ================
insert into orgs (id, name, created_at)
values (
      seed_utils.get_mapped_uuid('01JHX4ZCP3G2T4JC3P8VBA5N3A'),
      'Kubex Ecosystem',
      now()
   );
-- ================
-- TENANT PRINCIPAL
-- ================
insert into postgres.tenants (
      id,
      org_id,
      name,
      slug,
      status,
      created_at
   )
values (
      seed_utils.get_mapped_uuid('01JHX4ZH8Z9AGF1AFM7V0E7EAQ'),
      seed_utils.get_mapped_uuid('01JHX4ZCP3G2T4JC3P8VBA5N3A'),
      'Kubex HQ',
      'gnyx-hq',
      'active',
      now()
   );
-- ================
-- PERMISSIONS (ESSENCIAIS)
-- ================
insert into postgres.permissions (id, code, description)
values (
      seed_utils.get_mapped_uuid('01JHX50R3GGVQHH3M8RP6N3N8Y'),
      'dashboard.view',
      'Acessar dashboards'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX50R3MPD8ZT7NPB0D9G7V3'),
      'users.manage',
      'Gerenciar usuários'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX50R3W8HS14QW9PGN4VRE0'),
      'settings.manage',
      'Gerenciar configurações'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX50R3Z0B5ZQ5FJQ96A6BYT'),
      'partners.view',
      'Ver parceiros'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX50R43DW4TN9CP8NVQ7CFR'),
      'leads.view',
      'Visualizar leads'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX50R46MSN74Z8VMEGXSY9D'),
      'leads.manage',
      'Gerenciar leads'
   );
-- ================
-- ROLES BÁSICOS
-- ================
insert into postgres.roles (id, code, label, description)
values (
      seed_utils.get_mapped_uuid('01JHX52373KT7NPKDD6M5WBR4M'),
      'admin',
      seed_utils.get_mapped_uuid('Administrador'),
      'Acesso total'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX5237F45G3P6PRADAH9J8K'),
      'manager',
      'Gestor',
      'Coordenação e supervisão'
   ),
   (
      seed_utils.get_mapped_uuid('01JHX5237J1TQ5J0V8B4CWVZ4N'),
      'viewer',
      seed_utils.get_mapped_uuid('Visualizador'),
      'Acesso mínimo'
   );
-- ================
-- ROLE PERMISSIONS
-- ================
-- Admin -> tudo
insert into postgres.role_permissions (role_id, permission_id)
select seed_utils.get_mapped_uuid('01JHX52373KT7NPKDD6M5WBR4M'),
   id
from postgres.permissions;
-- Manager -> subset
insert into postgres.role_permissions (role_id, permission_id)
select seed_utils.get_mapped_uuid('01JHX5237F45G3P6PRADAH9J8K'),
   id
from postgres.permissions
where code in (
      'dashboard.view',
      'leads.view',
      'partners.view'
   );
-- Viewer -> mínimo
insert into postgres.role_permissions (role_id, permission_id)
values (
      seed_utils.get_mapped_uuid('01JHX5237J1TQ5J0V8B4CWVZ4N'),
      seed_utils.get_mapped_uuid('01JHX50R3GGVQHH3M8RP6N3N8Y')
   );
-- dashboard.view
-- ================
-- USUÁRIOS
-- ================
insert into postgres.users (
      id,
      email,
      full_name,
      password_hash,
      created_at
   )
values -- Usuário principal (você)
   (
      seed_utils.get_mapped_uuid('01JHX54TJEGW95WNH9TEG10M4H'),
      'rafael@kubex.world',
      'Rafael Mori',
      crypt(
         seed_utils.get_mapped_uuid('kubex123'),
         gen_salt('bf')
      ),
      now()
   ),
   -- Usuário secundário (Thiago)
   (
      seed_utils.get_mapped_uuid('01JHX54TJQBV8Z7RPKT4HR8X7N'),
      'thiago@kubex.world',
      'Thiago CTO',
      crypt(
         seed_utils.get_mapped_uuid('kubex123'),
         gen_salt('bf')
      ),
      now()
   );
-- ================
-- MEMBERSHIPS (ligação usuário ↔ tenant ↔ role)
-- ================
insert into postgres.memberships (
      id,
      user_id,
      tenant_id,
      role_id,
      created_at
   )
values (
      seed_utils.get_mapped_uuid('01JHX56GNNGQVPNB6WNXCHQH8W'),
      seed_utils.get_mapped_uuid('01JHX54TJEGW95WNH9TEG10M4H'),
      seed_utils.get_mapped_uuid('01JHX4ZH8Z9AGF1AFM7V0E7EAQ'),
      seed_utils.get_mapped_uuid('01JHX52373KT7NPKDD6M5WBR4M'),
      now()
   ),
   -- admin
   (
      seed_utils.get_mapped_uuid('01JHX56GP2AY39WXD2T3XYR94M'),
      seed_utils.get_mapped_uuid('01JHX54TJQBV8Z7RPKT4HR8X7N'),
      seed_utils.get_mapped_uuid('01JHX4ZH8Z9AGF1AFM7V0E7EAQ'),
      seed_utils.get_mapped_uuid('01JHX5237F45G3P6PRADAH9J8K'),
      now()
   );
-- manager
-- ================
-- TEAMS (mínimo necessário)
-- ================
insert into postgres.teams (id, tenant_id, name, created_at)
values (
      seed_utils.get_mapped_uuid('01JHX58W89V5Z2N4AWR4FSXCT3'),
      seed_utils.get_mapped_uuid('01JHX4ZH8Z9AGF1AFM7V0E7EAQ'),
      'Equipe Comercial',
      now()
   );
-- ================
-- TEAMS MEMBERS
-- ================
insert into postgres.teams_members (team_id, user_id)
values (
      seed_utils.get_mapped_uuid('01JHX58W89V5Z2N4AWR4FSXCT3'),
      seed_utils.get_mapped_uuid('01JHX54TJEGW95WNH9TEG10M4H')
   ),
   (
      seed_utils.get_mapped_uuid('01JHX58W89V5Z2N4AWR4FSXCT3'),
      seed_utils.get_mapped_uuid('01JHX54TJQBV8Z7RPKT4HR8X7N')
   );
-- ================
-- INVITE DE EXEMPLO
-- (pra testar magic link -> signup)
-- ================
insert into postgres.invites (
      id,
      email,
      role_id,
      tenant_id,
      created_by,
      status,
      created_at
   )
values (
      seed_utils.get_mapped_uuid('01JHX5AHDV7H8T6C2911CKGHEC'),
      'novo-user@kubex.world',
      seed_utils.get_mapped_uuid('01JHX5237J1TQ5J0V8B4CWVZ4N'),
      -- viewer
      seed_utils.get_mapped_uuid('01JHX4ZH8Z9AGF1AFM7V0E7EAQ'),
      seed_utils.get_mapped_uuid('01JHX54TJEGW95WNH9TEG10M4H'),
      'pending',
      now()
   );
-- ============================================================