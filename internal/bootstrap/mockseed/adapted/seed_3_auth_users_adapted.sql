SELECT seed_utils.set_search_path('public');
-- ============================================================
-- SEED KUBEX - AUTH USERS (Usuários para Teste de Auth)
-- ============================================================
-- Adiciona usuários específicos para testar autenticação
-- Garante que todas as 8 roles tenham pelo menos um usuário
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
-- ADICIONAR USUÁRIO VIEWER (faltante)
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
VALUES (
    '60000000-0000-0000-0000-000000000011',
    'viewer@kubex.world',
    'Test',
    'Viewer',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4326',
    'active',
    now()
  ) ON CONFLICT (id) DO NOTHING;
-- Adicionar membership viewer para Kubex HQ
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000011',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000008',
    true,
    now()
  ) ON CONFLICT (user_id, tenant_id) DO NOTHING;
-- ================
-- USUÁRIOS ADICIONAIS PARA TESTES
-- ================
-- Usuário Multi-Role (para testar múltiplos tenants)
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
VALUES (
    '60000000-0000-0000-0000-000000000012',
    'multi@kubex.world',
    'Multi',
    'Tenant User',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4327',
    'active',
    now()
  );
-- Multi-role user com acesso a múltiplos tenants
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000012',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000003',
    true,
    now()
  ),
  -- manager em HQ
  (
    '60000000-0000-0000-0000-000000000012',
    '30000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000004',
    true,
    now()
  );
-- sales_rep em Sales
-- Usuário Inativo (para testar bloqueio)
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
VALUES (
    '60000000-0000-0000-0000-000000000013',
    'inactive@kubex.world',
    'Inactive',
    'User',
    crypt(
      seed_utils.get_mapped_uuid('kubex123'),
      gen_salt('bf')
    ),
    '+55 11 98765-4328',
    'inactive',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000013',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000008',
    false,
    now()
  );
-- Usuário Pendente de Reset de Senha
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    phone,
    status,
    force_password_reset,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000014',
    'reset@kubex.world',
    'Password',
    'Reset',
    crypt('temp123', gen_salt('bf')),
    '+55 11 98765-4329',
    'active',
    true,
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000014',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000008',
    true,
    now()
  );
-- ================
-- USUÁRIOS DE TESTE POR ROLE (nomenclatura clara)
-- ================
-- Super Admin de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000015',
    'test.superadmin@kubex.world',
    'Test',
    'Super Admin',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000015',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    true,
    now()
  );
-- Admin de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000016',
    'test.admin@kubex.world',
    'Test',
    'Admin',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000016',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    true,
    now()
  );
-- Manager de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000017',
    'test.manager@kubex.world',
    'Test',
    'Manager',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000017',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000003',
    true,
    now()
  );
-- Sales Rep de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000018',
    'test.salesrep@kubex.world',
    'Test',
    'Sales Rep',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000018',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000004',
    true,
    now()
  );
-- Partner Admin de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000019',
    'test.partneradmin@kubex.world',
    'Test',
    'Partner Admin',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000019',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000005',
    true,
    now()
  );
-- Partner de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000020',
    'test.partner@kubex.world',
    'Test',
    'Partner',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000020',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000006',
    true,
    now()
  );
-- Analyst de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000021',
    'test.analyst@kubex.world',
    'Test',
    'Analyst',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000021',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000007',
    true,
    now()
  );
-- Viewer de Teste
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    password_hash,
    status,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000022',
    'test.viewer@kubex.world',
    'Test',
    'Viewer',
    crypt('test123', gen_salt('bf')),
    'active',
    now()
  );
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active,
    created_at
  )
VALUES (
    '60000000-0000-0000-0000-000000000022',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000008',
    true,
    now()
  );
-- ============================================================
-- FIM DO SEED AUTH USERS
-- ============================================================