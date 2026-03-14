#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2207
# ==========================================
# KUBEX META-SEEDER v1.0.0
# ==========================================
# Versão: 1.0.0
# Data: 2025-12-02
# Autores: Rafael Mori + Claude Code (Anthropic)
# ==========================================
# Descrição:
#   Meta-Seeder Kubex - Sistema completo de seeding avançado
#   Capacidades:
#     - Seed hydration (dados iniciais)
#     - Extração de schema PostgreSQL
#     - Geração de TypeScript types
#     - Geração de mocks TypeScript
#     - Sincronização bidirecional PG ↔ Mocks
# ==========================================
# Uso:
#   ./run_bootstrap.sh [comando] [opções]
#
# Comandos:
#   seed              - Executa seed hydration (padrão)
#   extract           - Extrai schema do PG para JSON
#   generate-types    - Gera TypeScript types
#   generate-mocks    - Gera mocks TypeScript
#   sync [modo]       - Sincroniza dados (diff|dry-run|pg-to-mock|mock-to-pg)
#   full-pipeline     - Executa pipeline completo
#   help              - Exibe esta ajuda
#
# Exemplos:
#   ./run_bootstrap.sh seed
#   ./run_bootstrap.sh extract /path/to/output
#   ./run_bootstrap.sh full-pipeline
#   ./run_bootstrap.sh sync diff
# ==========================================

set -euo pipefail # Exit on error, undefined vars, pipe failures

# ==========================================
# CONFIGURAÇÕES
# ==========================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${SCRIPT_DIR}/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="${LOG_DIR}/bootstrap_${TIMESTAMP}.log"
JSON_LOG="${LOG_DIR}/bootstrap_${TIMESTAMP}.json"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
# CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ==========================================
# FUNÇÕES AUXILIARES
# ==========================================

log_info() {
  echo -e "${BLUE} [INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
  echo -e "${YELLOW} [WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_step() {
  echo -e "${PURPLE}[STEP $1]${NC} $2" | tee -a "$LOG_FILE"
}

print_header() {
  echo "" | tee -a "$LOG_FILE"
  echo "========================================" | tee -a "$LOG_FILE"
  echo "$1" | tee -a "$LOG_FILE"
  echo "========================================" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
}

# ==========================================
# VALIDAÇÕES PRÉ-EXECUÇÃO
# ==========================================

validate_prerequisites() {
  log_info "Validando pré-requisitos..."

  # Verificar se psql está disponível
  if ! command -v psql &>/dev/null; then
    log_error "psql não encontrado. Instale PostgreSQL client."
    exit 1
  fi

  # Verificar variável de ambiente DATABASE_URL
  if [ -z "${DATABASE_URL:-}" ]; then
    log_error "Variável DATABASE_URL não definida."
    log_info "Defina: export DATABASE_URL='postgres://user:pass@host:port/dbname'"
    exit 1
  fi

  # Testar conexão
  if ! psql "$DATABASE_URL" -c "SELECT 1" &>/dev/null; then
    log_error "Não foi possível conectar ao banco de dados."
    exit 1
  fi

  log_success "Pré-requisitos validados"
}

# ==========================================
# EXECUTAR SEED HYDRATION
# ==========================================

hydration_seed_exec() {
  local step_num=$1
  local step_name=$2
  local step_file=$3
  local step_start=$(date +%s)

  log_step "$step_num" "$step_name"

  local full_path="${step_file}"

  if [ ! -f "$full_path" ]; then
    log_error "Arquivo não encontrado: $full_path"
    return 1
  fi

  # Executar SQL capturando stdout e stderr separadamente
  local sql_output
  local sql_exit_code
  sql_output=$(psql "${DATABASE_URL:-}" -v ON_ERROR_STOP=1 -f "$full_path" 2>&1)
  sql_exit_code=$?

  # Log do output
  echo "$sql_output" >>"$LOG_FILE"

  local step_end=$(date +%s)
  local duration=$((step_end - step_start))

  if [ $sql_exit_code -eq 0 ]; then
    # Verificar se há erros no output (mesmo com exit code 0)
    if echo "$sql_output" | grep -qE "^(ERROR|FATAL):"; then
      log_error "Erros SQL detectados no seed $step_num após ${duration}s"
      echo "$sql_output" | grep -E "^(ERROR|FATAL):" | head -5 | while read -r line; do
        log_error "  $line"
      done
      echo "$step_num,$step_name,$duration,failed" >>"${LOG_DIR}/execution_summary_seed.csv"
      return 1
    fi
    log_success "Hydration de seed $step_num concluída em ${duration}s"
    echo "$step_num,$step_name,$duration,success" >>"${LOG_DIR}/execution_summary_seed.csv"
    return 0
  else
    log_error "Falha no processo de hydration do seed $step_num após ${duration}s (exit code: $sql_exit_code)"
    # Mostrar as primeiras linhas de erro
    echo "$sql_output" | grep -E "^(ERROR|FATAL):" | head -5 | while read -r line; do
      log_error "  $line"
    done
    echo "$step_num,$step_name,$duration,failed" >>"${LOG_DIR}/execution_summary_seed.csv"
    return 1
  fi
}

# ==========================================
# VALIDAÇÕES PÓS-EXECUÇÃO
# ==========================================

validate_installation() {
  local seed_file=$1

  log_info "Validando dados inseridos..."

  local errors=0
  local warnings=0

  # Tabelas principais que devem ter dados após o seed
  local core_tables=("org" "tenant" "role" "permission" "role_permission" "\"user\"" "tenant_membership")

  for table in "${core_tables[@]}"; do
    local count
    count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $table" 2>/dev/null)
    if [ $? -ne 0 ]; then
      log_warning "Tabela $table não existe ou erro ao consultar"
      ((warnings++))
    elif [ "$count" -gt 0 ]; then
      log_info "✓ $table: $count registros"
    else
      log_warning "$table: vazia"
      ((warnings++))
    fi
  done

  # Tabelas opcionais (não é erro se estiverem vazias)
  local optional_tables=("team" "team_membership" "pipeline" "pipeline_stage" "partner" "lead")

  for table in "${optional_tables[@]}"; do
    local count
    count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $table" 2>/dev/null)
    if [ $? -eq 0 ] && [ "$count" -gt 0 ]; then
      log_info "✓ $table: $count registros"
    fi
  done

  if [ $errors -eq 0 ]; then
    if [ $warnings -gt 0 ]; then
      log_warning "Validação concluída com $warnings avisos"
    else
      log_success "Validação concluída sem erros"
    fi
    return 0
  else
    log_error "Validação encontrou $errors erros"
    return 1
  fi
}

# ==========================================
# GERAR RELATÓRIO JSON
# ==========================================

generate_json_report() {
  local total_duration=$1
  local status=$2
  local seed_file=$3

  # Tabelas conhecidas do schema
  local known_tables=("org" "tenant" "role" "permission" "role_permission" "\"user\"" "tenant_membership" "team" "team_membership" "pipeline" "pipeline_stage" "partner" "lead")

  local json_tables=""
  local total_records=0
  local tables_with_data=0

  for table in "${known_tables[@]}"; do
    local count
    count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $table" 2>/dev/null) || count="0"
    # Garantir que count é um número
    count="${count:-0}"
    count="${count// /}"
    if [[ "$count" =~ ^[0-9]+$ ]] && [ "$count" != "0" ]; then
      local clean_table="${table//\"/}"
      json_tables+="      \"$clean_table\": $count,"$'\n'
      total_records=$((total_records + count))
      tables_with_data=$((tables_with_data + 1))
    fi
  done
  json_tables="${json_tables%,$'\n'}" # Remove última vírgula e newline

  cat >"$JSON_LOG" <<EOF
{
  "execution_timestamp": "$(date -Iseconds)",
  "database_url": "${DATABASE_URL//:*@/:***@}",
  "total_duration_seconds": $total_duration,
  "status": "$status",
  "log_file": "$LOG_FILE",
  "seed_file": "$seed_file",
  "manifest_version": "0.0.1",
  "schema_version": "hybrid-v1.0",
  "features": {
    "seed_hydration": true
  },
  "summary": {
    "tables_with_data": $tables_with_data,
    "total_records": $total_records
  },
  "tables_seeded": {
$json_tables
  }
}
EOF

  log_info "Relatório JSON gerado: $JSON_LOG"
}

# ==========================================
# FUNÇÃO PRINCIPAL
# ==========================================

# TODO: permitir múltiplas seeds e execução em paralelo com relatórios,
# indempotência, deduplicaçõa, etc..
main() {
  local seed_pattern="${1:-fulldata}" # Padrão: fulldata (seed completo)

  # Buscar apenas no diretório core/ para evitar arquivos adaptados
  local seeds_to_use=()
  while IFS= read -r -d '' file; do
    seeds_to_use+=("$file")
  done < <(find "${SCRIPT_DIR}/core" -maxdepth 1 -name "*${seed_pattern}*.sql" -type f -print0 2>/dev/null | sort -z)

  if [ ${#seeds_to_use[@]} -eq 0 ]; then
    log_error "Seed '$seed_pattern' não encontrada no diretório core/."
    log_info "Seeds disponíveis:"
    ls -1 "${SCRIPT_DIR}/core/"*.sql 2>/dev/null | xargs -I{} basename {} | sed 's/^/  - /'
    exit 1
  fi

  local start_time=$(date +%s)

  # Criar diretório de logs
  mkdir -p "$LOG_DIR"

  # Header
  print_header "KUBEX HYDRATION v0.0.1 - INICIANDO"

  log_info "Timestamp: $(date)"
  log_info "Database: ${DATABASE_URL//:*@/:***@}"
  log_info "Log: $LOG_FILE"

  # Validar pré-requisitos
  validate_prerequisites

  # Criar CSV de resumo
  echo "step,name,duration_seconds,status" >"${LOG_DIR}/execution_summary_seed.csv"

  local step_counter=1
  for seed_file in "${seeds_to_use[@]}"; do
    log_info "Usando seed: $seed_file"
    hydration_seed_exec "$step_counter" "$(basename "$seed_file")" "$seed_file"
    ((step_counter++))

    # Validar instalação
    if validate_installation "$seed_file"; then
      local end_time=$(date +%s)
      local total_duration=$((end_time - start_time))

      print_header "BOOTSTRAP CONCLUÍDO COM SUCESSO"
      log_success "Tempo total: ${total_duration}s"
      log_info "Log completo: $LOG_FILE"

      generate_json_report "$total_duration" "success" "$seed_file"
      exit 0
    else
      local end_time=$(date +%s)
      local total_duration=$((end_time - start_time))

      print_header " BOOTSTRAP CONCLUÍDO COM AVISOS"
      log_warning "Tempo total: ${total_duration}s"
      log_warning "Revise o log: $LOG_FILE"

      generate_json_report "$total_duration" "success_with_warnings" "$seed_file"
      exit 0
    fi

  done
}

# ==========================================
# NOVAS FUNCIONALIDADES - META-SEEDER
# ==========================================

show_help() {
  print_header "🧙‍♂️ KUBEX META-SEEDER v1.0.0"

  echo "Uso: ./run_bootstrap.sh [comando] [opções]"
  echo ""
  echo "Comandos disponíveis:"
  echo ""
  echo "  seed [pattern]         Executa seed hydration (padrões: simplemock, fulldata, auth_users, leads)"
  echo "  extract [output_dir]   Extrai schema PostgreSQL para JSON"
  echo "  generate-types [dir]   Gera TypeScript types a partir do schema"
  echo "  generate-mocks [dir]   Gera mocks TypeScript dos dados reais"
  echo "  sync [dir] [mode]      Sincroniza PG ↔ Mocks"
  echo "                         Modos: diff | dry-run | pg-to-mock | mock-to-pg"
  echo "  full-pipeline [dir]    Executa pipeline completo"
  echo "  help                   Exibe esta ajuda"
  echo ""
  echo "Exemplos:"
  echo ""
  echo "  # Seed básico"
  echo "  ./run_bootstrap.sh seed"
  echo ""
  echo "  # Pipeline completo de geração"
  echo "  ./run_bootstrap.sh full-pipeline"
  echo ""
  echo "  # Gerar apenas types em diretório customizado"
  echo "  ./run_bootstrap.sh extract /tmp/output"
  echo "  ./run_bootstrap.sh generate-types /tmp/output"
  echo ""
  echo "  # Comparar PG vs Mocks"
  echo "  ./run_bootstrap.sh sync types diff"
  echo ""
}

run_extract() {
  local output_dir="${1:-${SCRIPT_DIR}/types}"

  print_header "🔍 EXTRAINDO SCHEMA DO POSTGRESQL"

  if [ ! -f "${SCRIPT_DIR}/extract_schema.sh" ]; then
    log_error "extract_schema.sh não encontrado"
    exit 1
  fi

  bash "${SCRIPT_DIR}/extract_schema.sh" "$output_dir"
}

run_generate_types() {
  local output_dir="${1:-${SCRIPT_DIR}/types}"

  print_header "GERANDO TYPESCRIPT TYPES"

  if [ ! -f "${SCRIPT_DIR}/generate_types.sh" ]; then
    log_error "generate_types.sh não encontrado"
    exit 1
  fi

  bash "${SCRIPT_DIR}/generate_types.sh" "$output_dir"
}

run_generate_mocks() {
  local output_dir="${1:-${SCRIPT_DIR}/types}"

  print_header "🎭 GERANDO MOCKS TYPESCRIPT"

  if [ ! -f "${SCRIPT_DIR}/generate_mocks.sh" ]; then
    log_error "generate_mocks.sh não encontrado"
    exit 1
  fi

  bash "${SCRIPT_DIR}/generate_mocks.sh" "$output_dir"
}

run_sync() {
  local output_dir="${1:-${SCRIPT_DIR}/types}"
  local mode="${2:-diff}"

  print_header "🔄 SINCRONIZANDO PG ↔ MOCKS"

  if [ ! -f "${SCRIPT_DIR}/sync_data.sh" ]; then
    log_error "sync_data.sh não encontrado"
    exit 1
  fi

  bash "${SCRIPT_DIR}/sync_data.sh" "$output_dir" "$mode"
}

run_full_pipeline() {
  local output_dir="${1:-${SCRIPT_DIR}/types}"

  print_header "EXECUTANDO PIPELINE COMPLETO"

  local pipeline_start=$(date +%s)

  log_info "Output: $output_dir"
  echo ""

  # Etapa 1: Extrair schema
  log_step 1 "Extraindo schema PostgreSQL"
  run_extract "$output_dir"
  echo ""

  # Etapa 2: Gerar types
  log_step 2 "Gerando TypeScript types"
  run_generate_types "$output_dir"
  echo ""

  # Etapa 3: Gerar mocks
  log_step 3 "Gerando mocks TypeScript"
  run_generate_mocks "$output_dir"
  echo ""

  # Etapa 4: Comparar
  log_step 4 "Comparando PG vs Mocks"
  run_sync "$output_dir" "diff"
  echo ""

  local pipeline_end=$(date +%s)
  local pipeline_duration=$((pipeline_end - pipeline_start))

  print_header "PIPELINE CONCLUÍDO"
  log_success "Tempo total: ${pipeline_duration}s"
  log_info "Arquivos gerados em: $output_dir"
  echo ""
  log_info "Próximos passos:"
  log_info "  - Copie os types para seu projeto frontend"
  log_info "  - Use os mocks para desenvolvimento local"
  log_info "  - Execute './run_bootstrap.sh sync diff' para comparar"
}

# ==========================================
# ROTEADOR DE COMANDOS
# ==========================================

route_command() {
  local command="${1:-seed}"
  shift || true

  case "$command" in
  seed)
    main "$@"
    ;;
  extract)
    run_extract "$@"
    ;;
  generate-types)
    run_generate_types "$@"
    ;;
  generate-mocks)
    run_generate_mocks "$@"
    ;;
  sync)
    run_sync "$@"
    ;;
  full-pipeline)
    run_full_pipeline "$@"
    ;;
  help | --help | -h)
    show_help
    ;;
  *)
    log_error "Comando desconhecido: $command"
    echo ""
    show_help
    exit 1
    ;;
  esac
}

# ==========================================
# EXECUTAR
# ==========================================

route_command "$@"
