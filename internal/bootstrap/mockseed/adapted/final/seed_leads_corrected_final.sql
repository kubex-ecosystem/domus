BEGIN;

SELECT seed_utils.set_search_path('public');


-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public

-- Garante que estamos no schema correto

SET search_path TO public;

-- ============================================================

-- LEADS CORRIGIDOS (partner_id)
INSERT INTO lead (id, tenant_id, pipeline_id, stage_id, company_name, contact_name, contact_email, contact_phone, value, status, assigned_to, partner_id, created_at) VALUES
  ('B0000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000004', 'Global Tech Inc', 'Jane Smith', 'jane@globaltech.com', '+1 555-5678', 250000.00, seed_utils.get_mapped_uuid('negotiating'), '60000000-0000-0000-0000-000000000003', 'A0000000-0000-0000-0000-000000000001', now() - interval '1 week'),
  ('B0000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000002', 'Enterprise Solutions', 'Alice Brown', 'alice@enterprise.com', '+1 555-3456', 500000.00, 'qualified', '60000000-0000-0000-0000-000000000003', 'A0000000-0000-0000-0000-000000000002', now() - interval '4 days'),
  ('B0000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000003', '80000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000012', 'LATAM Corp', 'Diego Martinez', 'diego@latamcorp.com', '+52 55 1111-2222', 180000.00, 'in_progress', '60000000-0000-0000-0000-000000000007', 'A0000000-0000-0000-0000-000000000003', now() - interval '5 days'),
  ('B0000000-0000-0000-0000-000000000008', '30000000-0000-0000-0000-000000000003', '80000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000014', 'SaaS Enterprise', 'Gabriel Costa', 'gabriel@saasent.io', '+55 21 3333-4444', 320000.00, 'won', '60000000-0000-0000-0000-000000000006', 'A0000000-0000-0000-0000-000000000005', now() - interval '2 weeks')
ON CONFLICT (id) DO NOTHING;


COMMIT;
