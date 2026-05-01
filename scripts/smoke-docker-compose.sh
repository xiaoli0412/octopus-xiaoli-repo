#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_PORT="${OCTOPUS_COMPOSE_SMOKE_PORT:-18086}"
COMPOSE_PROJECT="${OCTOPUS_COMPOSE_PROJECT:-octopus-smoke}"
COMPOSE_DATA_DIR="${OCTOPUS_COMPOSE_DATA_DIR:-${REPO_ROOT}/build/compose-smoke-data}"
COMPOSE_CONTAINER_NAME="${OCTOPUS_COMPOSE_CONTAINER_NAME:-octopus-smoke}"
BASE_URL="http://127.0.0.1:${COMPOSE_PORT}"

write_step() {
    printf '\n== %s ==\n' "$1"
}

require_command() {
    local name="$1"
    local hint="$2"
    if ! command -v "$name" >/dev/null 2>&1; then
        printf 'Missing required command %s. %s\n' "$name" "$hint" >&2
        exit 1
    fi
}

json_get() {
    local key_path="$1"
    python3 -c 'import json, sys
data = json.load(sys.stdin)
value = data
for part in sys.argv[1].split("."):
    value = value[part]
if isinstance(value, (dict, list)):
    print(json.dumps(value, ensure_ascii=False))
else:
    print(value)
' "$key_path"
}

cleanup() {
    local exit_code=$?
    if command -v docker >/dev/null 2>&1; then
        (
            cd "${REPO_ROOT}"
            docker compose -p "${COMPOSE_PROJECT}" down --volumes --remove-orphans >/dev/null 2>&1 || true
        )
    fi
    exit "${exit_code}"
}

trap cleanup EXIT INT TERM

require_command docker "Install Docker Desktop or Docker Engine and ensure docker compose is available."
require_command curl "Install curl and ensure it is on PATH."
require_command python3 "Install Python 3 and ensure it is on PATH."

mkdir -p "${COMPOSE_DATA_DIR}"

write_step "Building and starting docker compose runtime"
(
    cd "${REPO_ROOT}"
    OCTOPUS_PORT="${COMPOSE_PORT}" \
    OCTOPUS_DATA_DIR="${COMPOSE_DATA_DIR}" \
    OCTOPUS_CONTAINER_NAME="${COMPOSE_CONTAINER_NAME}" \
    docker compose -p "${COMPOSE_PROJECT}" up -d --build
)

HEALTH=""
for _ in $(seq 1 120); do
    HEALTH="$(curl --silent --show-error --fail --max-time 5 "${BASE_URL}/healthz" 2>/dev/null || true)"
    if [[ -n "${HEALTH}" ]]; then
        break
    fi
    sleep 2
done

if [[ -z "${HEALTH}" ]]; then
    (
        cd "${REPO_ROOT}"
        docker compose -p "${COMPOSE_PROJECT}" logs --tail=120 >&2 || true
    )
    printf 'docker compose runtime did not become healthy\n' >&2
    exit 1
fi

write_step "Verifying frontend shell and static assets"
ROOT_HTML="$(curl --silent --show-error --fail --max-time 10 "${BASE_URL}/")"
MANIFEST_JSON="$(curl --silent --show-error --fail --max-time 10 "${BASE_URL}/manifest.json")"
if [[ "${ROOT_HTML}" != *"<title>Octopus</title>"* ]]; then
    printf 'frontend shell did not render the expected Octopus title\n' >&2
    exit 1
fi
MANIFEST_NAME="$(printf '%s' "${MANIFEST_JSON}" | json_get 'name')"
if [[ "${MANIFEST_NAME}" != "Octopus" ]]; then
    printf 'manifest name = %s, want Octopus\n' "${MANIFEST_NAME}" >&2
    exit 1
fi

printf 'HEALTH=%s\n' "${HEALTH}"
printf 'FRONTEND_TITLE=%s\n' 'Octopus'
printf 'MANIFEST_NAME=%s\n' "${MANIFEST_NAME}"
printf 'COMPOSE_PROJECT=%s\n' "${COMPOSE_PROJECT}"
printf 'COMPOSE_DATA_DIR=%s\n' "${COMPOSE_DATA_DIR}"
