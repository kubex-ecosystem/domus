#!/usr/bin/env bash
# ==========================================
# KUBEX META-SEEDER: Bidirectional Sync
# ==========================================
# Sincroniza dados entre PostgreSQL e Mocks TypeScript
# Modos: diff, dry-run, pg-to-mock, mock-to-pg
# ==========================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${1:-${SCRIPT_DIR}/types}"
MOCKS_DIR="${OUTPUT_DIR}/mocks"
SCHEMA_FILE="${OUTPUT_DIR}/schema.json"

MODE="${2:-diff}"  # diff | dry-run | pg-to-mock | mock-to-pg

# Cores
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() {
  echo -e "${BLUE}ℹ️  [SYNC]${NC} $1"
}

log_success() {
  echo -e "${GREEN}✅ [SYNC]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}⚠️  [SYNC]${NC} $1"
}

log_error() {
  echo -e "${RED}❌ [SYNC]${NC} $1"
}

log_diff() {
  echo -e "${CYAN}🔄 [DIFF]${NC} $1"
}

# Validações
if [ -z "${DATABASE_URL:-}" ]; then
  log_error "DATABASE_URL não definida"
  exit 1
fi

if [ ! -f "$SCHEMA_FILE" ]; then
  log_error "schema.json não encontrado: $SCHEMA_FILE"
  exit 1
fi

if [ ! -d "$MOCKS_DIR" ]; then
  log_error "Diretório de mocks não encontrado: $MOCKS_DIR"
  log_info "Execute primeiro: ./generate_mocks.sh"
  exit 1
fi

if ! command -v jq &> /dev/null; then
  log_error "jq não encontrado"
  exit 1
fi

# ==========================================
# EXIBIR MODO
# ==========================================
case "$MODE" in
  diff)
    log_info "Modo: DIFF (apenas comparação)"
    ;;
  dry-run)
    log_info "Modo: DRY-RUN (simulação de mudanças)"
    ;;
  pg-to-mock)
    log_warning "Modo: PG → MOCK (irá sobrescrever mocks)"
    ;;
  mock-to-pg)
    log_warning "Modo: MOCK → PG (irá modificar banco de dados)"
    ;;
  *)
    log_error "Modo inválido: $MODE"
    log_info "Modos válidos: diff | dry-run | pg-to-mock | mock-to-pg"
    exit 1
    ;;
esac

# ==========================================
# FUNÇÕES DE COMPARAÇÃO
# ==========================================

compare_table() {
  local table_name=$1

  log_info "Comparando tabela: $table_name"

  # Dados do PG
  local pg_data=$(psql "$DATABASE_URL" -tAc "
    SELECT json_agg(row_to_json(t))
    FROM (SELECT * FROM \"$table_name\" ORDER BY id LIMIT 1000) t;
  " 2>/dev/null || echo "null")

  local pg_count=0
  if [ "$pg_data" != "null" ]; then
    pg_count=$(echo "$pg_data" | jq 'length' 2>/dev/null || echo 0)
  fi

  # Dados dos mocks
  local mock_file="${MOCKS_DIR}/${table_name}.ts"
  local mock_count=0

  if [ -f "$mock_file" ]; then
    # Extrair array TypeScript (remover export, import, comentários)
    local mock_data=$(grep -v '^import' "$mock_file" | \
                      grep -v '^export' | \
                      grep -v '^/\*' | \
                      grep -v '^ \*' | \
                      grep -v '^;' | \
                      sed 's/^export const .*= *//' | \
                      sed 's/;$//' | \
                      tr -d '\n' | \
                      jq '.' 2>/dev/null || echo "[]")

    mock_count=$(echo "$mock_data" | jq 'length' 2>/dev/null || echo 0)
  fi

  # Calcular diferença
  local diff=$((pg_count - mock_count))

  if [ $diff -eq 0 ]; then
    log_success "  ✓ $table_name: PG=$pg_count | Mock=$mock_count (sincronizado)"
  elif [ $diff -gt 0 ]; then
    log_diff "  ↑ $table_name: PG=$pg_count | Mock=$mock_count (+$diff no PG)"
  else
    log_diff "  ↓ $table_name: PG=$pg_count | Mock=$mock_count (${diff#-} no Mock)"
  fi

  echo "$table_name,$pg_count,$mock_count,$diff"
}

# ==========================================
# SINCRONIZAR PG → MOCK
# ==========================================

sync_pg_to_mock() {
  local table_name=$1

  log_info "Sincronizando PG → Mock: $table_name"

  if [ "$MODE" = "dry-run" ]; then
    log_info "  [DRY-RUN] Mock seria regenerado"
    return
  fi

  # Chamar generate_mocks.sh para essa tabela específica
  log_info "  Regenerando mock..."

  # Aqui poderíamos implementar lógica específica ou apenas chamar o gerador completo
  log_warning "  Implemente lógica específica ou execute: ./generate_mocks.sh"
}

# ==========================================
# SINCRONIZAR MOCK → PG
# ==========================================

sync_mock_to_pg() {
  local table_name=$1

  log_warning "Sincronizando Mock → PG: $table_name"

  local mock_file="${MOCKS_DIR}/${table_name}.ts"

  if [ ! -f "$mock_file" ]; then
    log_error "  Mock não encontrado: $mock_file"
    return 1
  fi

  # Extrair dados do mock
  local mock_data=$(grep -v '^import' "$mock_file" | \
                    grep -v '^export' | \
                    grep -v '^/\*' | \
                    grep -v '^ \*' | \
                    grep -v '^;' | \
                    sed 's/^export const .*= *//' | \
                    sed 's/;$//' | \
                    tr -d '\n')

  if [ "$MODE" = "dry-run" ]; then
    log_info "  [DRY-RUN] Dados seriam inseridos no PG"
    return
  fi

  log_error "  ⚠️  ATENÇÃO: Sincronização Mock → PG pode causar conflitos"
  log_error "  Esta funcionalidade requer implementação cuidadosa de:"
  log_error "    - Resolução de conflitos de ID"
  log_error "    - Validação de constraints"
  log_error "    - Merge vs Replace strategy"
  log_warning "  Use com cautela em ambiente de desenvolvimento apenas"
}

# ==========================================
# EXECUTAR COMPARAÇÃO
# ==========================================

log_info "Iniciando sincronização..."
echo ""

# CSV para relatório
REPORT_FILE="${OUTPUT_DIR}/sync_report_$(date +%Y%m%d_%H%M%S).csv"
echo "table_name,pg_count,mock_count,diff" > "$REPORT_FILE"

# Processar todas as tabelas
jq -r '.tables[].table_name' "$SCHEMA_FILE" | while read -r table_name; do
  result=$(compare_table "$table_name")
  echo "$result" >> "$REPORT_FILE"

  # Executar sincronização se não for apenas diff
  if [ "$MODE" = "pg-to-mock" ]; then
    sync_pg_to_mock "$table_name"
  elif [ "$MODE" = "mock-to-pg" ]; then
    sync_mock_to_pg "$table_name"
  fi
done

echo ""
log_success "Relatório gerado: $REPORT_FILE"

# Estatísticas finais
TOTAL_TABLES=$(wc -l < "$REPORT_FILE")
((TOTAL_TABLES--)) # Remover header

log_info "📊 Estatísticas:"
log_info "   Tabelas analisadas: $TOTAL_TABLES"

if [ "$MODE" = "diff" ] || [ "$MODE" = "dry-run" ]; then
  SYNCED=$(awk -F',' '$4 == 0' "$REPORT_FILE" | wc -l)
  OUT_OF_SYNC=$(awk -F',' '$4 != 0 && NR > 1' "$REPORT_FILE" | wc -l)

  log_info "   Sincronizadas: $SYNCED"
  log_info "   Dessincronizadas: $OUT_OF_SYNC"

  if [ $OUT_OF_SYNC -gt 0 ]; then
    echo ""
    log_warning "Para sincronizar, execute:"
    log_info "   PG → Mock: ./sync_data.sh [dir] pg-to-mock"
    log_info "   Mock → PG: ./sync_data.sh [dir] mock-to-pg (cuidado!)"
  fi
fi

echo ""
log_success "✨ Sincronização concluída!"
