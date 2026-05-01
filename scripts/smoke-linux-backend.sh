#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SMOKE_PORT="${OCTOPUS_SMOKE_PORT:-18084}"
MOCK_PORT="${OCTOPUS_SMOKE_MOCK_PORT:-19091}"
BASE_URL="http://127.0.0.1:${SMOKE_PORT}"
UNAME_S="$(uname -s)"
SMOKE_ADMIN_USERNAME="${OCTOPUS_SMOKE_ADMIN_USERNAME:-admin}"
SMOKE_ADMIN_PASSWORD="${OCTOPUS_SMOKE_ADMIN_PASSWORD:-admin}"

TEMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/octopus-e2e-XXXXXX")"
CONFIG_PATH="${TEMP_DIR}/config.json"
DB_PATH="${TEMP_DIR}/octopus.db"
SERVER_OUT="${TEMP_DIR}/server.stdout.log"
SERVER_ERR="${TEMP_DIR}/server.stderr.log"
MOCK_SCRIPT="${TEMP_DIR}/mock_upstream.py"
MOCK_OUT="${TEMP_DIR}/mock.stdout.log"
MOCK_ERR="${TEMP_DIR}/mock.stderr.log"
GOEXE="$(go env GOEXE 2>/dev/null || true)"
EXE_PATH="${TEMP_DIR}/octopus-smoke${GOEXE}"

MOCK_PID=""
SERVER_PID=""
HEALTH=""

write_step() {
    printf '\n== %s ==\n' "$1"
}

mask_secret() {
    local value="${1-}"
    if [[ -z "${value}" ]]; then
        printf '<empty>'
        return 0
    fi

    local visible=4
    local length=${#value}
    if (( length < visible )); then
        visible=${length}
    fi

    printf '***%s' "${value: -visible}"
}

require_command() {
    local name="$1"
    local hint="$2"
    if ! command -v "$name" >/dev/null 2>&1; then
        printf 'Missing required command %s. %s\n' "$name" "$hint" >&2
        exit 1
    fi
}

require_linux_runtime() {
    case "${UNAME_S}" in
        Linux)
            return 0
            ;;
        *)
            printf 'scripts/smoke-linux-backend.sh is intended for Linux/WSL shells. Current uname: %s\n' "${UNAME_S}" >&2
            printf 'Use scripts/smoke-win-backend.ps1 for Windows-native local smoke, or run this script on a Linux host/WSL instance.\n' >&2
            exit 1
            ;;
    esac
}

resolve_smoke_static_dir() {
    if [[ -n "${OCTOPUS_SMOKE_STATIC_DIR:-}" ]]; then
        printf '%s\n' "${OCTOPUS_SMOKE_STATIC_DIR}"
        return 0
    fi

    local source_dir="${REPO_ROOT}/web/out"
    local synced_dir="${REPO_ROOT}/static/out"
    local source_index="${source_dir}/index.html"
    local synced_index="${synced_dir}/index.html"

    if [[ -f "${source_index}" && ( ! -f "${synced_index}" || "${source_index}" -nt "${synced_index}" ) ]]; then
        printf '%s\n' "${source_dir}"
        return 0
    fi

    if [[ -d "${synced_dir}" ]]; then
        printf '%s\n' "${synced_dir}"
        return 0
    fi

    if [[ -d "${source_dir}" ]]; then
        printf '%s\n' "${source_dir}"
        return 0
    fi

    printf '%s\n' "${synced_dir}"
}

cleanup() {
    if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
        kill "${SERVER_PID}" >/dev/null 2>&1 || true
        wait "${SERVER_PID}" >/dev/null 2>&1 || true
    fi
    if [[ -n "${MOCK_PID}" ]] && kill -0 "${MOCK_PID}" >/dev/null 2>&1; then
        kill "${MOCK_PID}" >/dev/null 2>&1 || true
        wait "${MOCK_PID}" >/dev/null 2>&1 || true
    fi
}

dump_logs_on_failure() {
    if [[ -f "${SERVER_ERR}" ]]; then
        printf '\n-- server.stderr.log --\n' >&2
        tail -n 80 "${SERVER_ERR}" >&2 || true
    fi
    if [[ -f "${SERVER_OUT}" ]]; then
        printf '\n-- server.stdout.log --\n' >&2
        tail -n 80 "${SERVER_OUT}" >&2 || true
    fi
    if [[ -f "${MOCK_ERR}" ]]; then
        printf '\n-- mock.stderr.log --\n' >&2
        tail -n 80 "${MOCK_ERR}" >&2 || true
    fi
}

trap cleanup EXIT INT TERM

require_linux_runtime

require_command go "Install Go 1.24.4+ and ensure it is on PATH."
require_command curl "Install curl and ensure it is on PATH."

if command -v python3 >/dev/null 2>&1; then
    PYTHON_BIN="$(command -v python3)"
elif command -v python >/dev/null 2>&1; then
    PYTHON_BIN="$(command -v python)"
else
    printf 'Missing required command python3/python. Install Python 3 and ensure it is on PATH.\n' >&2
    exit 1
fi

write_step "Building temporary smoke binary"
(
    cd "${REPO_ROOT}"
    go build -o "${EXE_PATH}" .
)

SMOKE_STATIC_DIR="$(resolve_smoke_static_dir)"

REPO_ROOT="${REPO_ROOT}" DB_PATH="${DB_PATH}" CONFIG_PATH="${CONFIG_PATH}" SMOKE_PORT="${SMOKE_PORT}" SMOKE_STATIC_DIR="${SMOKE_STATIC_DIR}" "${PYTHON_BIN}" - <<'PY'
import json
import os
from pathlib import Path

config = {
    "server": {
        "host": "127.0.0.1",
        "port": int(os.environ["SMOKE_PORT"]),
        "static_dir": os.environ["SMOKE_STATIC_DIR"],
    },
    "database": {
        "type": "sqlite",
        "path": os.environ["DB_PATH"],
    },
    "log": {
        "level": "info",
    },
}

Path(os.environ["CONFIG_PATH"]).write_text(json.dumps(config, indent=2), encoding="utf-8")
PY

cat > "${MOCK_SCRIPT}" <<'PY'
import json
from http.server import BaseHTTPRequestHandler, HTTPServer


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(body.decode("utf-8"))
        except Exception:
            payload = {}
        response = {
            "id": "chatcmpl-mock-1",
            "object": "chat.completion",
            "created": 1713436800,
            "model": payload.get("model") or "gpt-4o-mini",
            "choices": [
                {
                    "index": 0,
                    "message": {"role": "assistant", "content": "mock-ok"},
                    "finish_reason": "stop",
                }
            ],
            "usage": {"prompt_tokens": 5, "completion_tokens": 2, "total_tokens": 7},
        }
        encoded = json.dumps(response).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(encoded)))
        self.end_headers()
        self.wfile.write(encoded)

    def log_message(self, fmt, *args):
        return


HTTPServer(("127.0.0.1", int(__import__("os").environ["MOCK_PORT"])), Handler).serve_forever()
PY

json_get() {
    local key_path="$1"
    "${PYTHON_BIN}" -c 'import json, sys
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

write_step "Starting mock upstream and Octopus server"
MOCK_PORT="${MOCK_PORT}" "${PYTHON_BIN}" "${MOCK_SCRIPT}" >"${MOCK_OUT}" 2>"${MOCK_ERR}" &
MOCK_PID="$!"
sleep 1

(
    cd "${REPO_ROOT}"
    OCTOPUS_ADMIN_USERNAME="${SMOKE_ADMIN_USERNAME}" \
    OCTOPUS_ADMIN_PASSWORD="${SMOKE_ADMIN_PASSWORD}" \
    "${EXE_PATH}" start --config "${CONFIG_PATH}" >"${SERVER_OUT}" 2>"${SERVER_ERR}"
) &
SERVER_PID="$!"

for _ in $(seq 1 80); do
    if "${EXE_PATH}" healthcheck --config "${CONFIG_PATH}" >/dev/null 2>&1; then
        HEALTH="$(curl --silent --show-error --fail --max-time 5 "${BASE_URL}/healthz")"
        break
    fi
    sleep 0.5
done

if [[ -z "${HEALTH}" ]]; then
    dump_logs_on_failure
    printf 'healthz did not respond successfully\n' >&2
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

write_step "Driving minimal management and gateway smoke flow"

LOGIN_RESP="$(curl --silent --show-error --fail --max-time 10 \
    -H 'Content-Type: application/json' \
    -d "$(printf '{"username":"%s","password":"%s","expire":86400}' "${SMOKE_ADMIN_USERNAME}" "${SMOKE_ADMIN_PASSWORD}")" \
    "${BASE_URL}/api/v1/user/login")"
JWT="$(printf '%s' "${LOGIN_RESP}" | json_get 'data.token')"

CHANNEL_RESP="$(curl --silent --show-error --fail --max-time 15 \
    -H "Authorization: Bearer ${JWT}" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"mock-openai-demo\",\"type\":0,\"enabled\":true,\"base_urls\":[{\"url\":\"http://127.0.0.1:${MOCK_PORT}\",\"delay\":0}],\"keys\":[{\"enabled\":true,\"channel_key\":\"mock-upstream-key\"}],\"model\":\"gpt-4o-mini\"}" \
    "${BASE_URL}/api/v1/channel/create")"
CHANNEL_ID="$(printf '%s' "${CHANNEL_RESP}" | json_get 'data.id')"

GROUP_RESP="$(curl --silent --show-error --fail --max-time 15 \
    -H "Authorization: Bearer ${JWT}" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"gpt-4o-mini\",\"mode\":1,\"items\":[{\"channel_id\":${CHANNEL_ID},\"model_name\":\"gpt-4o-mini\",\"priority\":1,\"weight\":1}]}" \
    "${BASE_URL}/api/v1/group/create")"
GROUP_ID="$(printf '%s' "${GROUP_RESP}" | json_get 'data.id')"

APIKEY_RESP="$(curl --silent --show-error --fail --max-time 15 \
    -H "Authorization: Bearer ${JWT}" \
    -H 'Content-Type: application/json' \
    -d '{"name":"smoke-key","enabled":true}' \
    "${BASE_URL}/api/v1/apikey/create")"
GATEWAY_KEY="$(printf '%s' "${APIKEY_RESP}" | json_get 'data.api_key')"

CHAT_RESP="$(curl --silent --show-error --fail --max-time 15 \
    -H "Authorization: Bearer ${GATEWAY_KEY}" \
    -H 'Content-Type: application/json' \
    -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}' \
    "${BASE_URL}/v1/chat/completions")"

printf 'HEALTH=%s\n' "${HEALTH}"
printf 'FRONTEND_TITLE=%s\n' 'Octopus'
printf 'MANIFEST_NAME=%s\n' "${MANIFEST_NAME}"
printf 'STATIC_DIR=%s\n' "${SMOKE_STATIC_DIR}"
printf 'CHANNEL_ID=%s\n' "${CHANNEL_ID}"
printf 'GROUP_ID=%s\n' "${GROUP_ID}"
printf 'GATEWAY_KEY_MASKED=%s\n' "$(mask_secret "${GATEWAY_KEY}")"
printf 'CHAT_RESPONSE=%s\n' "${CHAT_RESP}"
printf 'TEMP_DIR=%s\n' "${TEMP_DIR}"
