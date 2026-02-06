#!/usr/bin/env bash
# ==========================================
# KUBEX META-SEEDER: TypeScript Generator
# ==========================================
# Gera TypeScript types a partir do schema.json
# Entrada: schema.json
# Saída: types/*.ts
# ==========================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${1:-${SCRIPT_DIR}/types}"
SCHEMA_FILE="${OUTPUT_DIR}/schema.json"

# Cores
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
  echo -e "${BLUE} [GEN-TS]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[GEN-TS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW} [GEN-TS]${NC} $1"
}

log_error() {
  echo -e "${RED}[GEN-TS]${NC} $1"
}

# Validar schema.json
if [ ! -f "$SCHEMA_FILE" ]; then
  log_error "schema.json não encontrado em: $SCHEMA_FILE"
  log_info "Execute primeiro: ./extract_schema.sh"
  exit 1
fi

# Validar jq
if ! command -v jq &> /dev/null; then
  log_error "jq não encontrado. Instale: apt-get install jq"
  exit 1
fi

log_info "Gerando TypeScript types de: $SCHEMA_FILE"

# ==========================================
# MAPEAMENTO PG -> TS TYPES
# ==========================================
pg_to_ts_type() {
  local pg_type=$1
  local udt_name=$2

  case "$pg_type" in
    "uuid") echo "string" ;;
    "text"|"character varying"|"varchar"|"citext") echo "string" ;;
    "integer"|"bigint"|"smallint") echo "number" ;;
    "numeric"|"decimal"|"real"|"double precision") echo "number" ;;
    "boolean") echo "boolean" ;;
    "timestamp with time zone"|"timestamp without time zone"|"timestamptz"|"date") echo "Date | string" ;;
    "json"|"jsonb") echo "Record<string, any>" ;;
    "USER-DEFINED")
      # Mapear tipos USER-DEFINED conhecidos
      case "$udt_name" in
        "citext") echo "string" ;;
        "bool") echo "boolean" ;;
        "int4"|"int8") echo "number" ;;
        *) echo "$udt_name" ;;
      esac
      ;;
    *) echo "any /* $pg_type */" ;;
  esac
}

# ==========================================
# GERAR ENUMS
# ==========================================
log_info "Gerando enums..."

ENUM_COUNT=$(jq '.enums | length' "$SCHEMA_FILE")

if [ "$ENUM_COUNT" -gt 0 ]; then
  ENUMS_FILE="${OUTPUT_DIR}/enums.ts"

  cat > "$ENUMS_FILE" <<EOF
/**
 * Enums extraídos do PostgreSQL
 * Gerado automaticamente em $(date -Iseconds)
 * NÃO EDITE MANUALMENTE
 */

EOF

  jq -r '.enums[] | "export enum \(.name | ascii_upcase) {\n  \(.values | map("  \(. | ascii_upcase) = \"\(.)\"") | join(",\n"))\n}\n"' "$SCHEMA_FILE" >> "$ENUMS_FILE"

  log_success "Enums gerados: $ENUMS_FILE ($ENUM_COUNT enums)"
else
  log_warning "Nenhum enum encontrado"
fi

# ==========================================
# GERAR TYPES POR TABELA
# ==========================================
log_info "Gerando types por tabela..."

TABLE_COUNT=$(jq '.tables | length' "$SCHEMA_FILE")

if [ "$TABLE_COUNT" -eq 0 ]; then
  log_error "Nenhuma tabela encontrada no schema.json"
  exit 1
fi

# Criar index.ts para exports
INDEX_FILE="${OUTPUT_DIR}/index.ts"
cat > "$INDEX_FILE" <<EOF
/**
 * Types do banco Kubex
 * Gerado automaticamente em $(date -Iseconds)
 * NÃO EDITE MANUALMENTE
 */

EOF

# Adicionar export de enums se existir
if [ "$ENUM_COUNT" -gt 0 ]; then
  echo "export * from './enums';" >> "$INDEX_FILE"
fi

# Processar cada tabela
jq -c '.tables[]' "$SCHEMA_FILE" | while read -r table; do
  TABLE_NAME=$(echo "$table" | jq -r '.table_name')

  # Converter snake_case para PascalCase
  TYPE_NAME=$(echo "$TABLE_NAME" | sed -r 's/(^|_)([a-z])/\U\2/g')

  TYPE_FILE="${OUTPUT_DIR}/${TABLE_NAME}.ts"

  cat > "$TYPE_FILE" <<EOF
/**
 * Type para tabela: $TABLE_NAME
 * Gerado automaticamente em $(date -Iseconds)
 */

export interface ${TYPE_NAME} {
EOF

  # Adicionar cada coluna
  echo "$table" | jq -c '.columns[]' | while read -r column; do
    COL_NAME=$(echo "$column" | jq -r '.column_name')
    DATA_TYPE=$(echo "$column" | jq -r '.data_type')
    UDT_NAME=$(echo "$column" | jq -r '.udt_name')
    IS_NULLABLE=$(echo "$column" | jq -r '.is_nullable')

    TS_TYPE=$(pg_to_ts_type "$DATA_TYPE" "$UDT_NAME")

    # Adicionar ? se nullable
    if [ "$IS_NULLABLE" = "YES" ]; then
      echo "  ${COL_NAME}?: ${TS_TYPE};" >> "$TYPE_FILE"
    else
      echo "  ${COL_NAME}: ${TS_TYPE};" >> "$TYPE_FILE"
    fi
  done

  echo "}" >> "$TYPE_FILE"

  # Adicionar ao index
  echo "export * from './${TABLE_NAME}';" >> "$INDEX_FILE"

  log_info "  ✓ ${TABLE_NAME}.ts"
done

log_success "Types gerados: $OUTPUT_DIR/ ($TABLE_COUNT tabelas)"

# ==========================================
# GERAR METADATA
# ==========================================
METADATA_FILE="${OUTPUT_DIR}/metadata.ts"

cat > "$METADATA_FILE" <<EOF
/**
 * Metadata do schema
 * Gerado automaticamente em $(date -Iseconds)
 */

export const SCHEMA_METADATA = {
  extractedAt: '$(jq -r '.extracted_at' "$SCHEMA_FILE")',
  totalTables: $TABLE_COUNT,
  totalEnums: $ENUM_COUNT,
  tables: $(jq -c '[.tables[].table_name]' "$SCHEMA_FILE")
} as const;
EOF

echo "export * from './metadata';" >> "$INDEX_FILE"

log_success "Geração completa!"
log_info "📁 Arquivos em: $OUTPUT_DIR"
