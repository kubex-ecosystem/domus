-- seed_utils.sql
-- Utilities for deterministic mapping of textual IDs to UUIDs and session schema control

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS seed_utils AUTHORIZATION current_user;

CREATE TABLE IF NOT EXISTS seed_utils.id_map (
  source_text text PRIMARY KEY,
  mapped_uuid uuid NOT NULL,
  namespace uuid NOT NULL DEFAULT '11111111-1111-1111-1111-111111111111'::uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Deterministic mapping using UUID v5 (namespace + source_text)
CREATE OR REPLACE FUNCTION seed_utils.get_mapped_uuid(
  p_source_text text,
  p_namespace uuid DEFAULT '11111111-1111-1111-1111-111111111111'::uuid
) RETURNS uuid
LANGUAGE plpgsql AS $$
DECLARE
  v_uuid uuid := uuid_generate_v5(p_namespace, p_source_text);
  v_existing uuid;
BEGIN
  LOOP
    BEGIN
      INSERT INTO seed_utils.id_map(source_text, mapped_uuid, namespace)
      VALUES (p_source_text, v_uuid, p_namespace);
      RETURN v_uuid;
    EXCEPTION WHEN unique_violation THEN
      SELECT mapped_uuid INTO v_existing FROM seed_utils.id_map WHERE source_text = p_source_text;
      IF v_existing IS NOT NULL THEN
        RETURN v_existing;
      END IF;
      -- otherwise loop and retry (very unlikely)
    END;
  END LOOP;
END;
$$;

-- Helper to deterministically set the search_path for seeds (keeps public as fallback)
CREATE OR REPLACE FUNCTION seed_utils.set_search_path(p_schema text) RETURNS void
LANGUAGE plpgsql AS $$
BEGIN
  EXECUTE format('SET search_path TO %I, public', p_schema);
END;
$$;

-- Convenience: show mapping for a given source_text
CREATE OR REPLACE FUNCTION seed_utils.show_mapping(p_source_text text) RETURNS TABLE(source_text text, mapped_uuid uuid, namespace uuid, created_at timestamptz) AS $$
  SELECT source_text, mapped_uuid, namespace, created_at FROM seed_utils.id_map WHERE source_text = p_source_text;
$$ LANGUAGE sql STABLE;

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_seed_utils_id_map_namespace ON seed_utils.id_map(namespace, mapped_uuid);
