#!/usr/bin/env bash
# ==========================================
# KUBEX META-SEEDER: Schema Extractor
# ==========================================
# Extrai estrutura completa do PostgreSQL para JSON
# Saída: schema.json com tabelas, colunas, tipos, relações, constraints, enums
# ==========================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${1:-${SCRIPT_DIR}/types}"
SCHEMA_FILE="${OUTPUT_DIR}/schema.json"

# Cores
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
  echo -e "${BLUE} [EXTRACT]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[EXTRACT]${NC} $1"
}

log_error() {
  echo -e "${RED}[EXTRACT]${NC} $1"
}

# Validar DATABASE_URL
if [ -z "${DATABASE_URL:-}" ]; then
  log_error "DATABASE_URL não definida"
  exit 1
fi

# Criar diretório de saída
mkdir -p "$OUTPUT_DIR"

log_info "Extraindo schema do PostgreSQL..."

# ==========================================
# EXTRAIR ENUMS
# ==========================================
log_info "Extraindo enums..."

ENUMS_JSON=$(psql "$DATABASE_URL" -tAc "
SELECT json_agg(
  json_build_object(
    'name', t.typname,
    'values', array_agg(e.enumlabel ORDER BY e.enumsortorder)
  )
)
FROM pg_type t
JOIN pg_enum e ON t.oid = e.enumtypid
JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = 'public'
GROUP BY t.typname
ORDER BY t.typname;
" 2>/dev/null || echo "[]")

# ==========================================
# EXTRAIR TABELAS E COLUNAS
# ==========================================
log_info "Extraindo tabelas e colunas..."

TABLES_JSON=$(psql "$DATABASE_URL" -tAc "
SELECT json_agg(
  json_build_object(
    'table_name', c.table_name,
    'columns', (
      SELECT json_agg(
        json_build_object(
          'column_name', column_name,
          'data_type', data_type,
          'udt_name', udt_name,
          'is_nullable', is_nullable,
          'column_default', column_default,
          'character_maximum_length', character_maximum_length
        ) ORDER BY ordinal_position
      )
      FROM information_schema.columns cols
      WHERE cols.table_schema = 'public'
        AND cols.table_name = c.table_name
    )
  )
)
FROM (
  SELECT DISTINCT table_name
  FROM information_schema.columns
  WHERE table_schema = 'public'
  ORDER BY table_name
) c;
" 2>/dev/null || echo "[]")

# ==========================================
# EXTRAIR CONSTRAINTS (PKs, FKs, Uniques)
# ==========================================
log_info "Extraindo constraints..."

CONSTRAINTS_JSON=$(psql "$DATABASE_URL" -tAc "
SELECT json_agg(
  json_build_object(
    'constraint_name', tc.constraint_name,
    'table_name', tc.table_name,
    'constraint_type', tc.constraint_type,
    'column_name', kcu.column_name,
    'foreign_table_name', ccu.table_name,
    'foreign_column_name', ccu.column_name
  )
)
FROM information_schema.table_constraints tc
LEFT JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
  AND tc.table_schema = kcu.table_schema
LEFT JOIN information_schema.constraint_column_usage ccu
  ON ccu.constraint_name = tc.constraint_name
  AND ccu.table_schema = tc.table_schema
WHERE tc.table_schema = 'public'
  AND tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE')
ORDER BY tc.table_name, tc.constraint_type;
" 2>/dev/null || echo "[]")

# ==========================================
# MONTAR JSON FINAL
# ==========================================
log_info "Montando schema.json..."

cat > "$SCHEMA_FILE" <<EOF
{
  "extracted_at": "$(date -Iseconds)",
  "database": "${DATABASE_URL//:*@/:***@}",
  "enums": ${ENUMS_JSON:-[]},
  "tables": ${TABLES_JSON:-[]},
  "constraints": ${CONSTRAINTS_JSON:-[]}
}
EOF

# Validar JSON
if command -v jq &> /dev/null; then
  if jq empty "$SCHEMA_FILE" 2>/dev/null; then
    log_success "Schema extraído: $SCHEMA_FILE"

    # Estatísticas
    ENUM_COUNT=$(jq '.enums | length' "$SCHEMA_FILE" 2>/dev/null || echo 0)
    TABLE_COUNT=$(jq '.tables | length' "$SCHEMA_FILE" 2>/dev/null || echo 0)
    CONSTRAINT_COUNT=$(jq '.constraints | length' "$SCHEMA_FILE" 2>/dev/null || echo 0)

    log_info "📊 Enums: $ENUM_COUNT | Tabelas: $TABLE_COUNT | Constraints: $CONSTRAINT_COUNT"
  else
    log_error "JSON inválido gerado"
    exit 1
  fi
else
  log_success "Schema extraído (jq não disponível para validação): $SCHEMA_FILE"
fi
