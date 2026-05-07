#!/usr/bin/env bash

set -euo pipefail

DEFAULT_PORT="1088"
DEFAULT_DATA_DIR="./data"
DEFAULT_CONTAINER_NAME="octopus"
DEFAULT_IMAGE="ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.17"
DEFAULT_IMAGE_FALLBACK=""

OCTOPUS_PORT_INPUT="${OCTOPUS_PORT:-${DEFAULT_PORT}}"
OCTOPUS_DATA_DIR_INPUT="${OCTOPUS_DATA_DIR:-${DEFAULT_DATA_DIR}}"
OCTOPUS_CONTAINER_NAME_INPUT="${OCTOPUS_CONTAINER_NAME:-${DEFAULT_CONTAINER_NAME}}"
OCTOPUS_IMAGE_INPUT="${OCTOPUS_IMAGE:-${DEFAULT_IMAGE}}"
OCTOPUS_IMAGE_FALLBACK_INPUT="${OCTOPUS_IMAGE_FALLBACK:-${DEFAULT_IMAGE_FALLBACK}}"

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

pull_image_with_fallback() {
    local primary_image="$1"
    local fallback_image="$2"

    write_info "Pulling Docker image ${primary_image}"
    if docker pull "$primary_image"; then
        printf '%s' "$primary_image"
        return 0
    fi

    if [[ -n "$fallback_image" && "$fallback_image" != "$primary_image" ]]; then
        write_warn "Failed to pull ${primary_image}; retrying with fallback image ${fallback_image}."
        if docker pull "$fallback_image"; then
            printf '%s' "$fallback_image"
            return 0
        fi
    fi

    fail "Unable to pull Octopus image. Check access to GHCR or Docker Hub, or re-run with OCTOPUS_IMAGE=<reachable-image>."
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
            local fallback_port=""
            fallback_port="$(find_noninteractive_fallback_port)" || true

            if [[ -n "$fallback_port" ]]; then
                write_warn "Port ${port} is already in use; auto-selected ${fallback_port} for non-interactive install."
                printf '%s' "$fallback_port"
                return 0
            fi

            fail "Port ${port} is already in use and no fallback port is free. Re-run with OCTOPUS_PORT=<new-port>."
        fi

        write_warn "Detected an existing listener on port ${port}."
        prompt_for_port "1008"
        return 0
    fi

    printf '%s' "$port"
}

find_noninteractive_fallback_port() {
    local candidate
    for candidate in 18080 18086 28080; do
        if ! port_in_use "$candidate"; then
            printf '%s' "$candidate"
            return 0
        fi
    done

    return 1
}

require_command docker "Install Docker and ensure docker compose is available."

if ! docker info >/dev/null 2>&1; then
    fail "Docker daemon is not available. Start Docker and try again."
fi

EXTERNAL_PORT="$(resolve_external_port "$OCTOPUS_PORT_INPUT")"

if [[ ! -t 0 ]]; then
    write_info "Non-interactive install detected. If this host cannot access raw.githubusercontent.com or GHCR reliably, download scripts/install.sh locally first or set OCTOPUS_IMAGE to a reachable registry mirror."
fi

PULLED_IMAGE="$(pull_image_with_fallback "$OCTOPUS_IMAGE_INPUT" "$OCTOPUS_IMAGE_FALLBACK_INPUT")"

write_info "Removing any existing container named ${OCTOPUS_CONTAINER_NAME_INPUT}"
docker rm -f "$OCTOPUS_CONTAINER_NAME_INPUT" >/dev/null 2>&1 || true

mkdir -p "$OCTOPUS_DATA_DIR_INPUT"

write_info "Starting Octopus container with external port ${EXTERNAL_PORT}"
CONTAINER_ID="$(docker run -d \
    --name "$OCTOPUS_CONTAINER_NAME_INPUT" \
    --restart unless-stopped \
    -p "${EXTERNAL_PORT}:1088" \
    -e OCTOPUS_SERVER_HOST=0.0.0.0 \
    -e OCTOPUS_SERVER_PORT=1088 \
    -e OCTOPUS_DATABASE_TYPE=sqlite \
    -e OCTOPUS_DATABASE_PATH=/app/data/data.db \
    -e OCTOPUS_LOG_LEVEL=info \
    -e DATA_DIR=/app/data \
    -e PUID=10001 \
    -e PGID=10001 \
    -v "${OCTOPUS_DATA_DIR_INPUT}:/app/data" \
    "$PULLED_IMAGE")"

if [[ "$(docker inspect -f '{{.State.Running}}' "$CONTAINER_ID" 2>/dev/null || printf 'false')" != "true" ]]; then
    write_warn "Container failed to stay running. Showing the latest logs for ${OCTOPUS_CONTAINER_NAME_INPUT}:"
    docker logs --tail 80 "$OCTOPUS_CONTAINER_NAME_INPUT" >&2 || true
    fail "Octopus container exited right after startup. Check the logs above, especially data directory permissions or registry/image overrides."
fi

write_info "Octopus is starting"
write_info "UI: http://127.0.0.1:${EXTERNAL_PORT}"
write_info "Container: ${OCTOPUS_CONTAINER_NAME_INPUT}"
write_info "Data dir: ${OCTOPUS_DATA_DIR_INPUT}"
write_info "Image: ${PULLED_IMAGE}"
