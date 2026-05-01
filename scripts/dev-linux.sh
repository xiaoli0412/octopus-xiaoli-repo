#!/usr/bin/env bash

set -euo pipefail

BACKEND_ONLY=0
FRONTEND_ONLY=0
INSTALL_FRONTEND=0
CHECK_ONLY=0
API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:8080}"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --backend-only)
            BACKEND_ONLY=1
            shift
            ;;
        --frontend-only)
            FRONTEND_ONLY=1
            shift
            ;;
        --install-frontend)
            INSTALL_FRONTEND=1
            shift
            ;;
        --check-only)
            CHECK_ONLY=1
            shift
            ;;
        --api-base-url)
            API_BASE_URL="${2:?missing value for --api-base-url}"
            shift 2
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

if [[ $BACKEND_ONLY -eq 1 && $FRONTEND_ONLY -eq 1 ]]; then
    echo "Cannot use --backend-only and --frontend-only together." >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WEB_DIR="${REPO_ROOT}/web"

if [[ ! -f "${REPO_ROOT}/main.go" ]]; then
    echo "Repository root not detected: ${REPO_ROOT}" >&2
    exit 1
fi

if [[ ! -f "${WEB_DIR}/package.json" ]]; then
    echo "Web directory not detected: ${WEB_DIR}" >&2
    exit 1
fi

require_command() {
    local name="$1"
    local hint="$2"
    if ! command -v "${name}" >/dev/null 2>&1; then
        echo "Missing required command '${name}'. ${hint}" >&2
        exit 1
    fi
}

version_core_from_text() {
    local text="$1"
    if [[ "$text" =~ ([0-9]+)\.([0-9]+)(\.([0-9]+))? ]]; then
        local major="${BASH_REMATCH[1]}"
        local minor="${BASH_REMATCH[2]}"
        local patch="${BASH_REMATCH[4]:-0}"
        printf '%s.%s.%s\n' "$major" "$minor" "$patch"
        return 0
    fi

    echo "Unable to parse version from: $text" >&2
    return 1
}

assert_minimum_version() {
    local tool_name="$1"
    local actual_text="$2"
    local minimum_version="$3"
    local actual_version
    local actual_major actual_minor actual_patch
    local minimum_major minimum_minor minimum_patch

    actual_version="$(version_core_from_text "$actual_text")" || exit 1
    IFS='.' read -r actual_major actual_minor actual_patch <<< "$actual_version"
    IFS='.' read -r minimum_major minimum_minor minimum_patch <<< "$minimum_version"

    if (( actual_major < minimum_major )) || \
       (( actual_major == minimum_major && actual_minor < minimum_minor )) || \
       (( actual_major == minimum_major && actual_minor == minimum_minor && actual_patch < minimum_patch )); then
        echo "${tool_name} version ${actual_version} is too old. Required: ${minimum_version} or newer." >&2
        exit 1
    fi

    echo "[OK] ${tool_name} version ${actual_version}"
}

echo "== Checking required tools =="

if [[ $FRONTEND_ONLY -eq 0 ]]; then
    require_command go "Install Go 1.24.4+ and ensure it is on PATH."
    assert_minimum_version "Go" "$(go version)" "1.24.4"
fi

if [[ $BACKEND_ONLY -eq 0 ]]; then
    require_command node "Install Node.js 18+ and ensure it is on PATH."
    require_command pnpm "Install pnpm and ensure it is on PATH."
    assert_minimum_version "Node.js" "$(node --version)" "18.0.0"
    assert_minimum_version "pnpm" "$(pnpm --version)" "7.0.0"
fi

if [[ $CHECK_ONLY -eq 1 ]]; then
    echo "== Check-only summary =="
    if [[ $FRONTEND_ONLY -eq 0 ]]; then
        echo "Backend command: cd '${REPO_ROOT}' && go run main.go start"
    fi
    if [[ $BACKEND_ONLY -eq 0 ]]; then
        echo "Frontend command: cd '${WEB_DIR}' && NEXT_PUBLIC_API_BASE_URL='${API_BASE_URL}' pnpm run dev"
    fi
    exit 0
fi

if [[ $BACKEND_ONLY -eq 0 ]]; then
    if [[ $INSTALL_FRONTEND -eq 1 || ! -d "${WEB_DIR}/node_modules" ]]; then
        echo "== Installing frontend dependencies =="
        (cd "${WEB_DIR}" && pnpm install --frozen-lockfile)
    fi
fi

start_backend() {
    echo "== Starting backend =="
    (
        cd "${REPO_ROOT}"
        exec go run main.go start
    )
}

start_frontend() {
    echo "== Starting frontend =="
    (
        cd "${WEB_DIR}"
        export NEXT_PUBLIC_API_BASE_URL="${API_BASE_URL}"
        exec pnpm run dev
    )
}

if [[ $BACKEND_ONLY -eq 1 ]]; then
    start_backend
    exit 0
fi

if [[ $FRONTEND_ONLY -eq 1 ]]; then
    start_frontend
    exit 0
fi

start_backend &
BACKEND_PID=$!
trap 'kill ${BACKEND_PID} >/dev/null 2>&1 || true' EXIT

start_frontend
