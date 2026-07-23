BEGIN;

SET search_path TO public, bmori;

-- ----------------------------------------------------------------------------
-- BMori seed: alunos + biometria + churn para validação do Attention List real
-- ----------------------------------------------------------------------------

-- 1) Garante role "client" (aluno) para os fluxos de convite e RBAC.
INSERT INTO role (code, display_name, description, is_system_role)
VALUES ('client', 'Aluno', 'Acesso do aluno ao app BMori', true)
ON CONFLICT (code) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  description = EXCLUDED.description;

-- 2) Define tenant alvo (primeiro tenant ativo).
WITH target_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_active = true
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO "user" (id, email, name, last_name, password_hash, status, created_at)
VALUES
  (
    '71000000-0000-0000-0000-000000000001',
    'coach.bmori@kubex.world',
    'Marina',
    'Costa',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  ),
  (
    '71000000-0000-0000-0000-000000000011',
    'ana.aluno@kubex.world',
    'Ana',
    'Oliveira',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  ),
  (
    '71000000-0000-0000-0000-000000000012',
    'bruno.aluno@kubex.world',
    'Bruno',
    'Melo',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  ),
  (
    '71000000-0000-0000-0000-000000000013',
    'livia.aluno@kubex.world',
    'Livia',
    'Reis',
    crypt('kubex123', gen_salt('bf')),
    'active',
    now()
  )
ON CONFLICT (email) DO UPDATE
SET
  name = EXCLUDED.name,
  last_name = EXCLUDED.last_name,
  status = EXCLUDED.status;

-- 3) Vincula memberships (coach e alunos).
WITH target_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_active = true
  ORDER BY created_at ASC
  LIMIT 1
),
role_map AS (
  SELECT code, id
  FROM role
  WHERE code IN ('coach', 'trainer', 'manager', 'admin', 'client', 'viewer')
)
INSERT INTO tenant_membership (user_id, tenant_id, role_id, is_active, created_at, updated_at)
SELECT
  u.id AS user_id,
  t.id AS tenant_id,
  CASE
    WHEN u.email = 'coach.bmori@kubex.world'
      THEN COALESCE(
        (SELECT id FROM role_map WHERE code = 'coach'),
        (SELECT id FROM role_map WHERE code = 'trainer'),
        (SELECT id FROM role_map WHERE code = 'manager'),
        (SELECT id FROM role_map WHERE code = 'viewer')
      )
    ELSE COALESCE(
      (SELECT id FROM role_map WHERE code = 'client'),
      (SELECT id FROM role_map WHERE code = 'viewer')
    )
  END AS role_id,
  true,
  now(),
  now()
FROM "user" u
CROSS JOIN target_tenant t
WHERE u.email IN (
  'coach.bmori@kubex.world',
  'ana.aluno@kubex.world',
  'bruno.aluno@kubex.world',
  'livia.aluno@kubex.world'
)
ON CONFLICT (user_id, tenant_id) DO UPDATE
SET
  role_id = EXCLUDED.role_id,
  is_active = true,
  updated_at = now();

-- 4) Limpa dados BMori anteriores destes alunos para evitar duplicidade.
DELETE FROM bmori.biometrics
WHERE student_user_id IN (
  '71000000-0000-0000-0000-000000000011'::uuid,
  '71000000-0000-0000-0000-000000000012'::uuid,
  '71000000-0000-0000-0000-000000000013'::uuid
);

DELETE FROM bmori.churn_risk_profiles
WHERE student_user_id IN (
  '71000000-0000-0000-0000-000000000011'::uuid,
  '71000000-0000-0000-0000-000000000012'::uuid,
  '71000000-0000-0000-0000-000000000013'::uuid
);

-- 5) Insere biometria recente para três alunos.
WITH target_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_active = true
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO bmori.biometrics (
  tenant_id,
  student_user_id,
  personal_user_id,
  recorded_at,
  source_device,
  source_vendor,
  resting_heart_rate,
  avg_heart_rate,
  steps_count,
  sleep_minutes,
  calories_burned,
  distance_meters,
  payload
)
SELECT
  t.id,
  x.student_user_id,
  '71000000-0000-0000-0000-000000000001'::uuid,
  x.recorded_at,
  x.source_device,
  x.source_vendor,
  x.resting_heart_rate,
  x.avg_heart_rate,
  x.steps_count,
  x.sleep_minutes,
  x.calories_burned,
  x.distance_meters,
  x.payload::jsonb
FROM target_tenant t
CROSS JOIN (
  VALUES
    (
      '71000000-0000-0000-0000-000000000011'::uuid,
      now() - interval '1 day',
      'whoop',
      'whoop',
      63,
      84,
      5200,
      305,
      620,
      4500,
      '{"hrv_ms": 38, "sleep_score": 53, "trend": "down"}'
    ),
    (
      '71000000-0000-0000-0000-000000000012'::uuid,
      now() - interval '1 day',
      'apple_watch',
      'apple',
      58,
      79,
      7600,
      390,
      710,
      5800,
      '{"hrv_ms": 54, "sleep_score": 71, "trend": "stable"}'
    ),
    (
      '71000000-0000-0000-0000-000000000013'::uuid,
      now() - interval '1 day',
      'oura_ring',
      'oura',
      52,
      74,
      9300,
      445,
      790,
      6800,
      '{"hrv_ms": 68, "sleep_score": 83, "trend": "up"}'
    )
) AS x(
  student_user_id,
  recorded_at,
  source_device,
  source_vendor,
  resting_heart_rate,
  avg_heart_rate,
  steps_count,
  sleep_minutes,
  calories_burned,
  distance_meters,
  payload
);

-- 6) Insere perfis de churn para alimentar a Attention List.
WITH target_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_active = true
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO bmori.churn_risk_profiles (
  tenant_id,
  student_user_id,
  personal_user_id,
  risk_score,
  risk_level,
  status,
  assessment_window_days,
  last_biometrics_at,
  last_assessment_at,
  factors,
  recommended_actions,
  notes
)
SELECT
  t.id,
  x.student_user_id,
  '71000000-0000-0000-0000-000000000001'::uuid,
  x.risk_score,
  x.risk_level,
  'active',
  14,
  now() - interval '1 day',
  now(),
  x.factors::jsonb,
  x.recommended_actions::jsonb,
  x.notes
FROM target_tenant t
CROSS JOIN (
  VALUES
    (
      '71000000-0000-0000-0000-000000000011'::uuid,
      88.0,
      'critical',
      '[{"factor":"hrv_drop","weight":0.45},{"factor":"poor_sleep","weight":0.35}]',
      '[{"action":"active_recovery"},{"action":"coach_checkin"}]',
      'HRV em queda contínua e sono insuficiente por 7 dias.'
    ),
    (
      '71000000-0000-0000-0000-000000000012'::uuid,
      62.0,
      'high',
      '[{"factor":"sleep_instability","weight":0.30},{"factor":"adherence_drop","weight":0.22}]',
      '[{"action":"nutrition_review"},{"action":"weekly_checkin"}]',
      'Oscilação de recuperação e adesão abaixo da meta.'
    ),
    (
      '71000000-0000-0000-0000-000000000013'::uuid,
      28.0,
      'low',
      '[{"factor":"stable_biometrics","weight":0.10}]',
      '[{"action":"maintain_plan"}]',
      'Engajamento estável com sinais fisiológicos positivos.'
    )
) AS x(student_user_id, risk_score, risk_level, factors, recommended_actions, notes)
ON CONFLICT (tenant_id, student_user_id) DO UPDATE
SET
  personal_user_id = EXCLUDED.personal_user_id,
  risk_score = EXCLUDED.risk_score,
  risk_level = EXCLUDED.risk_level,
  status = EXCLUDED.status,
  assessment_window_days = EXCLUDED.assessment_window_days,
  last_biometrics_at = EXCLUDED.last_biometrics_at,
  last_assessment_at = EXCLUDED.last_assessment_at,
  factors = EXCLUDED.factors,
  recommended_actions = EXCLUDED.recommended_actions,
  notes = EXCLUDED.notes,
  updated_at = now();

COMMIT;
