#!/usr/bin/env bash
# shellcheck disable=SC2155

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace
set -o functrace
shopt -s inherit_errexit

# ENVVARS
declare -a _GNYX_ARGS=()
# Assumo o ROOT_DIR como padrão o diretório do projeto
# com fallback pro fixo GNYX_DIR
_GNYX_DIR="/ALL/KUBEX/SHOWCASE/projects/gnyx"
_ROOT_DIR="${_ROOT_DIR:-${_GNYX_DIR}}"
_SCRIPT_DIR="${_ROOT_DIR}/support"
_BINARY_NAME="${_BINARY_NAME:-gnyx}"
_APP_NAME="${_APP_NAME:-}"
_PID=

# _GNYX_UI_DIR="${_GNYX_DIR}/frontend"
# _GNYX_UI_BIN="${_GNYX_UI_DIR}/dist"

__source_script_if_needed() {
  local _check_declare="${1:-}"
  local _script_path="${2:-}"
  # shellcheck disable=SC2065
  if test -z "$(declare -f "${_check_declare:-}")" >/dev/null; then
    # shellcheck source=/dev/null
    source "${_script_path:-}" || {
      echo "Error: Could not source ${_script_path:-}. Please ensure it exists." >&2
      return 1
    }
  fi
  return 0
}
__source_script_if_needed "show_summary" "${_SCRIPT_DIR:-}/config.sh" || exit 1
__source_script_if_needed "apply_manifest" "${_SCRIPT_DIR:-}/apply_manifest.sh" || exit 1
__source_script_if_needed "get_current_shell" "${_SCRIPT_DIR:-}/utils.sh" || exit 1

stop() {
  kill "${_PID}" 2>/dev/null || {
    log error "Error traying to kill gnyx process" true
  }
  return 0
}

die() {
  trap - EXIT HUP INT QUIT ABRT ALRM TERM
  stop 2>/dev/null || true
  clear_build_artifacts 2>/dev/null || true
  clear_script_cache 2>/dev/null || true
  exit 0
}

trap "die;" EXIT

_gnyx_pid() {
  # Nada fica na 5000 se não achar o BE
  local _pid=$(pgrep -f "${_BINARY_NAME}" 2>/dev/null || lsof -Fp -i:5000 | tr -d 'p' 2>/dev/null || echo "")
  if [[ -n "$_pid" ]]; then
    if [[ -n "${_PID:-}" && "${_PID:-}" != "$_pid" ]]; then
      log warn "PID already set to ${_PID:-}, but found $_pid. Killing $_pid"
      kill -9 "${_PID:-}" || {
        log error "Failed to kill gnyx process" true
        return 1
      }
    fi
    _PID="$_pid" && echo "$_PID"
    return 0
  fi
  return 1
}

start() {
  # todo: sanityze args here
  local _start_cmd=("${_BINARY_NAME}" 'gateway' 'up' "$@")

  if _gnyx_pid; then
    # only gets here if _gnyx_pid found a PID
    log warn "Gnyx is already running with PID ${_PID}"
    return 0
  fi

  "${_start_cmd[@]}" &

  _gnyx_pid || {
    log error "Failed to start gnyx" true
    return 1
  }

  log info "Gnyx started with PID ${_PID}"
  return 0
}

restart() {
  _gnyx_pid || {
    log error "Gnyx is not running"
    return 1
  }

  kill -9 "${_PID}" 2>/dev/null || true
  start_gnyx "${@}" || {
    log error "Failed to start GNyx in the restart"
    return 1
  }
  return 0
}

build() {
  cd "${_ROOT_DIR}" || {
    log error "Failed to change directory to ${_ROOT_DIR}" true
    return 1
  }
  go fmt ./... 2>/dev/null || {
    log error "Failed to format go code" true
    return 1
  }
  go vet ./... 2>/dev/null || {
    log error "Failed to vet go code" true
    return 1
  }
  go build -v ./... 2>/dev/null || {
    log error "Failed to build gnyx" true
    return 1
  }
  go mod tidy 2>/dev/null || {
    log error "Failed to tidy go modules" true
    return 1
  }
  make build-dev 2>/dev/null || {
    log error "Failed to build gnyx ui" true
    return 1
  }
  log success "Gnyx built successfully"
  return 0
}

build_ui() {
  pnpm --dir "${_GNYX_UI_DIR}" build || {
    log error "Failed to build gnyx ui" true
    return 1
  }
  return 0
}

broker() {

  return $?
}

manager() {

  # START ITERATION
  while true; do
    # HOLD ARGS/CMDS
    local _exec_args=()

    case "${_exec_args[0]}" in
    #######################################
    # UI - UI AND OTHERS
    build-ui)
      build_gnyx_ui "${_exec_args[@]}"
      ;;

    #######################################
    # GNYX - MAIN PACKAGE CMDS
    start | stop | restart | build)
      # concat action with target to create a function name
      broker "${_exec_args[0]}" "${_exec_args[@]}"
      ;;

    #######################################
    # UNKNOWN
    *)
      log error "Invalid argument: ${_exec_args[0]}" true
      ;;
    esac

    # LOOP COMMAND ERROR HANDLER
    if ! $?; then
      log error "Command failed: ${_exec_args[0]}" true
      return 1
    fi

    # CLEAR COLLECTED ARGS FOR NEXT ITERATION
    if [[ -v _exec_args ]]; then
      unset _exec_args
    fi

    # WAIT FOR NEXT COMMAND
    if [[ -n $_PID || $(_gnyx_pid) ]]; then
      sleep 1
      continue
    fi
  done
}

manager "$@" || exit 1
