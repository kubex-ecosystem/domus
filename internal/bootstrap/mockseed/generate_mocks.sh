#!/usr/bin/env bash
# ==========================================
# KUBEX META-SEEDER: Mock Generator
# ==========================================
# Gera mocks TypeScript a partir dos dados reais do PostgreSQL
# Entrada: schema.json + DATABASE_URL
# Saída: mocks/*.ts
# ==========================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${1:-${SCRIPT_DIR}/types}"
SCHEMA_FILE="${OUTPUT_DIR}/schema.json"
MOCKS_DIR="${OUTPUT_DIR}/mocks"

# Cores
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
  echo -e "${BLUE}ℹ️  [GEN-MOCK]${NC} $1"
}

log_success() {
  echo -e "${GREEN}✅ [GEN-MOCK]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}⚠️  [GEN-MOCK]${NC} $1"
}

log_error() {
  echo -e "${RED}❌ [GEN-MOCK]${NC} $1"
}

# Validar DATABASE_URL
if [ -z "${DATABASE_URL:-}" ]; then
  log_error "DATABASE_URL não definida"
  exit 1
fi

# Validar schema.json
if [ ! -f "$SCHEMA_FILE" ]; then
  log_error "schema.json não encontrado: $SCHEMA_FILE"
  log_info "Execute primeiro: ./extract_schema.sh"
  exit 1
fi

# Validar jq
if ! command -v jq &> /dev/null; then
  log_error "jq não encontrado"
  exit 1
fi

# Criar diretório de mocks
mkdir -p "$MOCKS_DIR"

log_info "Gerando mocks TypeScript..."

# ==========================================
# CONVERTER VALOR PG PARA TS
# ==========================================
format_ts_value() {
  local value=$1
  local data_type=$2

  if [ "$value" = "null" ] || [ -z "$value" ]; then
    echo "null"
    return
  fi

  case "$data_type" in
    "uuid"|"text"|"character varying"|"citext"|"timestamp"*|"date")
      # Escapar strings e datas
      echo "\"$(echo "$value" | sed 's/"/\\"/g')\""
      ;;
    "integer"|"bigint"|"smallint"|"numeric"|"decimal"|"real"|"double precision")
      echo "$value"
      ;;
    "boolean")
      echo "$value"
      ;;
    "json"|"jsonb")
      # JSON já está formatado
      echo "$value"
      ;;
    *)
      echo "\"$value\""
      ;;
  esac
}

# ==========================================
# GERAR INDEX DE MOCKS
# ==========================================
MOCKS_INDEX="${MOCKS_DIR}/index.ts"
cat > "$MOCKS_INDEX" <<EOF
/**
 * Mocks do banco Kubex
 * Gerado automaticamente em $(date -Iseconds)
 * NÃO EDITE MANUALMENTE - dados reais do PostgreSQL
 */

EOF

# ==========================================
# PROCESSAR CADA TABELA
# ==========================================
TABLE_COUNT=$(jq '.tables | length' "$SCHEMA_FILE")
PROCESSED=0

jq -r '.tables[].table_name' "$SCHEMA_FILE" | while read -r TABLE_NAME; do
  log_info "Processando tabela: $TABLE_NAME"

  # Converter para PascalCase
  TYPE_NAME=$(echo "$TABLE_NAME" | sed -r 's/(^|_)([a-z])/\U\2/g')

  # Obter dados da tabela (limite de 1000 registros por segurança)
  TABLE_DATA=$(psql "$DATABASE_URL" -tAc "
    SELECT json_agg(row_to_json(t))
    FROM (SELECT * FROM \"$TABLE_NAME\" LIMIT 1000) t;
  " 2>/dev/null || echo "null")

  if [ "$TABLE_DATA" = "null" ] || [ -z "$TABLE_DATA" ]; then
    log_warning "  Tabela vazia: $TABLE_NAME"
    MOCK_FILE="${MOCKS_DIR}/${TABLE_NAME}.ts"

    cat > "$MOCK_FILE" <<EOF
import { ${TYPE_NAME} } from '../${TABLE_NAME}';

export const ${TABLE_NAME}Mocks: ${TYPE_NAME}[] = [];
EOF

    echo "export * from './${TABLE_NAME}';" >> "$MOCKS_INDEX"
    continue
  fi

  # Contar registros
  RECORD_COUNT=$(echo "$TABLE_DATA" | jq 'length' 2>/dev/null || echo 0)

  MOCK_FILE="${MOCKS_DIR}/${TABLE_NAME}.ts"

  cat > "$MOCK_FILE" <<EOF
import { ${TYPE_NAME} } from '../${TABLE_NAME}';

/**
 * Mocks para tabela: $TABLE_NAME
 * Total de registros: $RECORD_COUNT
 */
export const ${TABLE_NAME}Mocks: ${TYPE_NAME}[] =
EOF

  # Formatar JSON com indentação
  echo "$TABLE_DATA" | jq '.' >> "$MOCK_FILE"

  echo ";" >> "$MOCK_FILE"

  # Adicionar ao index
  echo "export * from './${TABLE_NAME}';" >> "$MOCKS_INDEX"

  log_success "  ✓ ${TABLE_NAME}.ts ($RECORD_COUNT registros)"
  ((PROCESSED++))
done

log_success "✨ Mocks gerados: $MOCKS_DIR/ ($PROCESSED tabelas)"

# ==========================================
# GERAR HELPER DE ACESSO
# ==========================================
HELPER_FILE="${MOCKS_DIR}/helpers.ts"

cat > "$HELPER_FILE" <<EOF
/**
 * Helpers para trabalhar com mocks
 * Gerado automaticamente em $(date -Iseconds)
 */

import * as mocks from './index';

/**
 * Retorna todos os mocks de uma tabela específica
 */
export function getMocks<T>(tableName: string): T[] {
  const mockKey = \`\${tableName}Mocks\`;
  return (mocks as any)[mockKey] || [];
}

/**
 * Busca um mock por ID
 */
export function findById<T extends { id: any }>(
  tableName: string,
  id: any
): T | undefined {
  const items = getMocks<T>(tableName);
  return items.find((item) => item.id === id);
}

/**
 * Filtra mocks por propriedade
 */
export function filterBy<T>(
  tableName: string,
  key: keyof T,
  value: any
): T[] {
  const items = getMocks<T>(tableName);
  return items.filter((item) => item[key] === value);
}
EOF

echo "export * from './helpers';" >> "$MOCKS_INDEX"

log_success "Helpers criados: $HELPER_FILE"
log_info "📁 Mocks disponíveis em: $MOCKS_DIR"
