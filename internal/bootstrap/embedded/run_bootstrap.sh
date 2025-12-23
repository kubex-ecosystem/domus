#!/usr/bin/env bash
# ==========================================
# CANALIZE BOOTSTRAP EXECUTOR v3.0
# ==========================================
# Versão: 3.0.0
# Data: 2025-11-10
# Autores: Rafael Mori (Desenvolvedor) + Claude Code (Anthropic)
# ==========================================
# Descrição:
#   Script executor do Schema Híbrido Kubex v3.0
#   Executa todas as 8 etapas em ordem, com validação e logging.
# ==========================================

set -euo pipefail  # Exit on error, undefined vars, pipe failures

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
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ==========================================
# FUNÇÕES AUXILIARES
# ==========================================

log_info() {
  echo -e "${BLUE}ℹ️  [INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
  echo -e "${GREEN}✅ [SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
  echo -e "${YELLOW}⚠️  [WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
  echo -e "${RED}❌ [ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_step() {
  echo -e "${PURPLE}🚀 [STEP $1]${NC} $2" | tee -a "$LOG_FILE"
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
  if ! command -v psql &> /dev/null; then
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
  if ! psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
    log_error "Não foi possível conectar ao banco de dados."
    exit 1
  fi

  log_success "Pré-requisitos validados"
}

# ==========================================
# EXECUTAR ETAPA
# ==========================================

execute_step() {
  local step_num=$1
  local step_name=$2
  local step_file=$3
  local step_start=$(date +%s)

  log_step "$step_num" "$step_name"

  local full_path="${SCRIPT_DIR}/${step_file}"

  if [ ! -f "$full_path" ]; then
    log_error "Arquivo não encontrado: $full_path"
    return 1
  fi

  # Executar SQL com output redirecionado para log
  if psql "${DATABASE_URL:-}" -f "$full_path" >> "$LOG_FILE" 2>&1; then
    local step_end=$(date +%s)
    local duration=$((step_end - step_start))
    log_success "Etapa $step_num concluída em ${duration}s"
    echo "$step_num,$step_name,$duration,success" >> "${LOG_DIR}/execution_summary.csv"
    return 0
  else
    local step_end=$(date +%s)
    local duration=$((step_end - step_start))
    log_error "Falha na etapa $step_num após ${duration}s"
    echo "$step_num,$step_name,$duration,failed" >> "${LOG_DIR}/execution_summary.csv"
    return 1
  fi
}

# ==========================================
# VALIDAÇÕES PÓS-EXECUÇÃO
# ==========================================

validate_installation() {
  log_info "Validando instalação..."

  local errors=0

  # Verificar tabelas críticas
  local required_tables=("org" "tenant" "team" "user" "role" "permission" "role_permission" "tenant_membership" "team_membership" "partner_invitation" "internal_invitation" "pipeline" "pipeline_stage" "partner" "lead" "commission" "clawback")

  for table in "${required_tables[@]}"; do
    if psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM pg_tables WHERE schemaname='public' AND tablename='$table'" | grep -q "1"; then
      log_info "✓ Tabela '$table' criada"
    else
      log_error "✗ Tabela '$table' NÃO encontrada"
      ((errors++))
    fi
  done

  # Verificar funções críticas
  local required_functions=("update_updated_at_column")

  for func in "${required_functions[@]}"; do
    if psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM pg_proc WHERE proname='$func'" | grep -q -E "[1-9]"; then
      log_info "✓ Função '$func' criada"
    else
      log_error "✗ Função '$func' NÃO encontrada"
      ((errors++))
    fi
  done

  # Verificar roles do sistema
  local roles_count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM role WHERE is_system_role = true")
  if [ "$roles_count" -eq 8 ]; then
    log_info "✓ 8 roles do sistema criadas"
  else
    log_warning "⚠ Esperado 8 roles, encontrado: $roles_count"
  fi

  # Verificar permissions
  local permissions_count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM permission")
  if [ "$permissions_count" -eq 35 ]; then
    log_info "✓ 35 permissions criadas"
  else
    log_warning "⚠ Esperado 35 permissions, encontrado: $permissions_count"
  fi

  if [ $errors -eq 0 ]; then
    log_success "Validação concluída sem erros"
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

  cat > "$JSON_LOG" <<EOF
{
  "execution_timestamp": "$(date -Iseconds)",
  "database_url": "${DATABASE_URL//:*@/:***@}",
  "total_duration_seconds": $total_duration,
  "status": "$status",
  "log_file": "$LOG_FILE",
  "steps_executed": $(wc -l < "${LOG_DIR}/execution_summary.csv"),
  "manifest_version": "3.0.0",
  "schema_version": "hybrid-v1.0",
  "features": {
    "multi_tenancy": true,
    "rbac_hierarchy": true,
    "invites": true,
    "business_entities": true,
    "saas_tiers": true
  }
}
EOF

  log_info "Relatório JSON gerado: $JSON_LOG"
}

# ==========================================
# FUNÇÃO PRINCIPAL
# ==========================================

main() {
  local start_time=$(date +%s)

  # Criar diretório de logs
  mkdir -p "$LOG_DIR"

  # Header
  print_header "🚀 CANALIZE SCHEMA HÍBRIDO v3.0 - INICIANDO"

  log_info "Timestamp: $(date)"
  log_info "Database: ${DATABASE_URL//:*@/:***@}"
  log_info "Log: $LOG_FILE"

  # Validar pré-requisitos
  validate_prerequisites

  # Criar CSV de resumo
  echo "step,name,duration_seconds,status" > "${LOG_DIR}/execution_summary.csv"

  # Executar etapas conforme manifest
  execute_step 1 "Extensions + Multi-Tenancy" "core/etapa_1_extensions_tenancy.sql" || exit 1
  execute_step 2 "Users + RBAC" "core/etapa_2_users_rbac.sql" || exit 1
  execute_step 3 "Memberships + FKs Circulares" "core/etapa_3_memberships.sql" || exit 1
  execute_step 4 "Invites" "core/etapa_4_invites.sql" || exit 1
  execute_step 5 "Business Entities" "core/etapa_5_business_entities.sql" || exit 1
  execute_step 6 "Índices de Performance" "core/etapa_6_indices.sql" || exit 1
  execute_step 7 "Triggers" "core/etapa_7_triggers.sql" || exit 1
  execute_step 8 "Seed Data" "core/etapa_8_seed_data.sql" || exit 1
  execute_step 9 "Seed Data" "core/etapa_9_auth_sessions.sql" || exit 1
  execute_step 10 "Seed Data" "core/etapa_10_pending_access_requests.sql" || exit 1

  # Validar instalação
  if validate_installation; then
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))

    print_header "✅ BOOTSTRAP CONCLUÍDO COM SUCESSO"
    log_success "Tempo total: ${total_duration}s"
    log_info "Log completo: $LOG_FILE"

    generate_json_report "$total_duration" "success"
    exit 0
  else
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))

    print_header "⚠️  BOOTSTRAP CONCLUÍDO COM AVISOS"
    log_warning "Tempo total: ${total_duration}s"
    log_warning "Revise o log: $LOG_FILE"

    generate_json_report "$total_duration" "success_with_warnings"
    exit 0
  fi
}

# ==========================================
# EXECUTAR
# ==========================================

main "$@"
