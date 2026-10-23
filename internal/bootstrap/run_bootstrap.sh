#!/usr/bin/env bash
# shellcheck disable=SC2155

set -euo pipefail

# run_bootstrap.sh
# Applies seed_utils and final seed SQL files in order, with logging and dry-run.

_DATABASE_URL="postgres://kubex_adm:admin123@127.0.0.1:5432/postgres?sslmode=disable"
DSN="${DATABASE_URL:-${_DATABASE_URL:-}}"

# Testa conexão com o banco de dados
psql "$DSN" -c "SELECT 1;" || {
  echo "Erro ao conectar ao banco de dados"
  exit 1
}

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT" || exit 1

declare -A DIRSMAP=(
  ["adapted"]="${REPO_ROOT}/internal/bootstrap/mockseed/adapted/final"
  ["helpers"]="${REPO_ROOT}/internal/bootstrap/mockseed/helpers"
  ["logs"]="${REPO_ROOT}/internal/bootstrap/logs"
)

DRY_RUN=0

for k in "${!DIRSMAP[@]}"; do
  if [[ ! -d "${DIRSMAP[$k]}" ]]; then
    echo "Creating directory ${DIRSMAP[$k]}"
    mkdir -p "${DIRSMAP[$k]}"
  fi
done

usage() {
  cat <<EOF
Usage: $0 [--dry-run]
  --dry-run    Validate SQL files only by running in a transaction and rolling back
EOF
}

if [[ ${1:-} == "--dry-run" ]]; then
  DRY_RUN=1
fi

run_psql() {
  local file="$1"

  local logfile="${DIRSMAP["logs"]}/$(basename "${file}").log"
  if [[ $DRY_RUN -eq 1 ]]; then
    psql "$DSN" -v ON_ERROR_STOP=1 -f <(sed '1i BEGIN; ' "${file}")
  else
    psql "$DSN" -v ON_ERROR_STOP=1 -f "${file}" | tee "${logfile}"
  fi
}

# Apply seed_utils first
echo "Applying seed_utils..."
psql "$DSN" -v ON_ERROR_STOP=1 -f "${DIRSMAP["helpers"]}/seed_utils.sql" | tee "${DIRSMAP["logs"]}/seed_utils.log"

# Apply all final seeds in alphabetical order
for f in "${DIRSMAP["adapted"]}"/*_final.sql; do
  [ -e "$f" ] || continue
  echo "Applying $f"
  run_psql "$f"
done

echo "Bootstrap completed. Logs in ${DIRSMAP["logs"]}"
