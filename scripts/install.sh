#!/usr/bin/env bash

set -euo pipefail

DEFAULT_PORT="8080"
DEFAULT_DATA_DIR="./data"
DEFAULT_CONTAINER_NAME="octopus"
DEFAULT_REPO_URL="https://github.com/xiaoli0412/octopus-xiaoli-repo.git"
DEFAULT_REPO_DIR="octopus-xiaoli-repo"

OCTOPUS_PORT_INPUT="${OCTOPUS_PORT:-${DEFAULT_PORT}}"
OCTOPUS_DATA_DIR_INPUT="${OCTOPUS_DATA_DIR:-${DEFAULT_DATA_DIR}}"
OCTOPUS_CONTAINER_NAME_INPUT="${OCTOPUS_CONTAINER_NAME:-${DEFAULT_CONTAINER_NAME}}"
REPO_URL="${OCTOPUS_REPO_URL:-${DEFAULT_REPO_URL}}"
REPO_DIR="${OCTOPUS_REPO_DIR:-${DEFAULT_REPO_DIR}}"

write_info() {
    printf '[INFO] %s\n' "$1"
}

write_warn() {
    printf '[WARN] %s\n' "$1" >&2
}

fail() {
    printf '[ERROR] %s\n' "$1" >&2
    exit 1
}

require_command() {
    local name="$1"
    local hint="$2"
    if ! command -v "$name" >/dev/null 2>&1; then
        fail "Missing required command '$name'. ${hint}"
    fi
}

is_valid_port() {
    local port="$1"
    [[ "$port" =~ ^[0-9]+$ ]] || return 1
    (( port >= 1 && port <= 65535 ))
}

port_in_use() {
    local port="$1"

    if command -v ss >/dev/null 2>&1; then
        ss -ltn "sport = :${port}" | tail -n +2 | grep -q .
        return
    fi

    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
        return
    fi

    if command -v netstat >/dev/null 2>&1; then
        netstat -ltn 2>/dev/null | awk '{print $4}' | grep -E "(^|[:.])${port}$" >/dev/null 2>&1
        return
    fi

    return 1
}

prompt_for_port() {
    local suggested_port="$1"
    local chosen_port=""

    while true; do
        printf 'Port %s is already in use. Enter a different external port [%s]: ' "$DEFAULT_PORT" "$suggested_port"
        IFS= read -r chosen_port
        chosen_port="${chosen_port:-$suggested_port}"

        if ! is_valid_port "$chosen_port"; then
            write_warn "Port must be an integer between 1 and 65535."
            continue
        fi

        if port_in_use "$chosen_port"; then
            write_warn "Port ${chosen_port} is also in use. Choose another port."
            continue
        fi

        printf '%s' "$chosen_port"
        return 0
    done
}

resolve_external_port() {
    local port="$1"

    if ! is_valid_port "$port"; then
        fail "Invalid OCTOPUS_PORT: ${port}. Expected an integer between 1 and 65535."
    fi

    if [[ "$port" != "$DEFAULT_PORT" ]]; then
        if port_in_use "$port"; then
            fail "Requested OCTOPUS_PORT ${port} is already in use."
        fi
        printf '%s' "$port"
        return 0
    fi

    if port_in_use "$port"; then
        if [[ ! -t 0 ]]; then
            fail "Port ${port} is already in use. Re-run with OCTOPUS_PORT=<new-port>, for example OCTOPUS_PORT=1008."
        fi

        write_warn "Detected an existing listener on port ${port}."
        prompt_for_port "1008"
        return 0
    fi

    printf '%s' "$port"
}

require_command git "Install Git and ensure it is on PATH."
require_command docker "Install Docker and ensure docker compose is available."

if ! docker compose version >/dev/null 2>&1; then
    fail "docker compose is required. Install a Docker version that includes the compose plugin."
fi

EXTERNAL_PORT="$(resolve_external_port "$OCTOPUS_PORT_INPUT")"

if [[ -e "$REPO_DIR" && ! -d "$REPO_DIR" ]]; then
    fail "Target path exists and is not a directory: ${REPO_DIR}"
fi

if [[ ! -d "$REPO_DIR/.git" ]]; then
    write_info "Cloning repository into ${REPO_DIR}"
    git clone "$REPO_URL" "$REPO_DIR"
else
    write_info "Repository already exists in ${REPO_DIR}; reusing current checkout"
fi

cd "$REPO_DIR"

write_info "Starting Octopus with external port ${EXTERNAL_PORT}"
OCTOPUS_PORT="$EXTERNAL_PORT" \
OCTOPUS_DATA_DIR="$OCTOPUS_DATA_DIR_INPUT" \
OCTOPUS_CONTAINER_NAME="$OCTOPUS_CONTAINER_NAME_INPUT" \
docker compose up -d --build

write_info "Octopus is starting"
write_info "UI: http://127.0.0.1:${EXTERNAL_PORT}"
write_info "Container: ${OCTOPUS_CONTAINER_NAME_INPUT}"
write_info "Data dir: ${OCTOPUS_DATA_DIR_INPUT}"
