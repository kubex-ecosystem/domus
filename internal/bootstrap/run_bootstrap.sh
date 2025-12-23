#!/usr/bin/env bash
# shellcheck disable=SC2155

set -euo pipefail

# run_bootstrap.sh
# Applies seed_utils and final seed SQL files in order, with logging and dry-run.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT" || exit 1
_DEFAULT_DSN="postgres://kubex_adm:admin123@localhost:5432/postgres?sslmode=disable"
DSN="${DSN:-${_DEFAULT_DSN:-}}"
ADAPTED_DIR="${REPO_ROOT}/internal/bootstrap/mockseed/adapted/final"
LOG_DIR="${REPO_ROOT}/internal/bootstrap/logs"
DRY_RUN=0

mkdir -p "$LOG_DIR"

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
  
  local logfile="$LOG_DIR/$(basename "$file").log"
  if [[ $DRY_RUN -eq 1 ]]; then
    psql "$DSN" -v ON_ERROR_STOP=1 -f <(sed '1i BEGIN; ' "$file");
  else
    psql "$DSN" -v ON_ERROR_STOP=1 -f "$file" | tee "$logfile"
  fi
}

# Apply seed_utils first
echo "Applying seed_utils..."
psql "$DSN" -v ON_ERROR_STOP=1 -f ./internal/bootstrap/mockseed/helpers/seed_utils.sql | tee "$LOG_DIR/seed_utils.log"

# Apply all final seeds in alphabetical order
for f in "$ADAPTED_DIR"/*_final.sql; do
  [ -e "$f" ] || continue
  echo "Applying $f"
  run_psql "$f"
done

echo "Bootstrap completed. Logs in $LOG_DIR"
