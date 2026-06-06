#!/usr/bin/env bash

set -euo pipefail

DEFAULT_PORT="1088"
DEFAULT_DATA_DIR="./data"
DEFAULT_CONTAINER_NAME="octopus"
DEFAULT_IMAGE="ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.19.4"
DEFAULT_REPO_REF="${DEFAULT_IMAGE##*:}"

OCTOPUS_PORT_INPUT="${OCTOPUS_PORT:-${DEFAULT_PORT}}"
OCTOPUS_DATA_DIR_INPUT="${OCTOPUS_DATA_DIR:-${DEFAULT_DATA_DIR}}"
OCTOPUS_CONTAINER_NAME_INPUT="${OCTOPUS_CONTAINER_NAME:-${DEFAULT_CONTAINER_NAME}}"
OCTOPUS_IMAGE_INPUT="${OCTOPUS_IMAGE:-${DEFAULT_IMAGE}}"

DOCKER_RUN_SECURITY_ARGS=(
    --read-only
    --tmpfs /tmp:rw,noexec,nosuid,size=64m
    --security-opt no-new-privileges:true
    --cap-drop ALL
    --cap-add CHOWN
    --cap-add SETGID
    --cap-add SETUID
)

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

is_disabled_dockerhub_image() {
    local image="$1"
    case "$image" in
        docker.io/*|index.docker.io/*|xiaoli0412/octopus-xiaoli-repo|xiaoli0412/octopus-xiaoli-repo:*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

prepare_repo_checkout() {
    local repo_url="$1"
    local repo_dir="$2"
    local repo_ref="$3"

    if ! command -v git >/dev/null 2>&1; then
        write_warn "Missing required command 'git'. Install Git and ensure it is on PATH."
        return 1
    fi

    if [[ ! -d "$repo_dir/.git" ]]; then
        write_info "Cloning repository into ${repo_dir}"
        if ! git clone "$repo_url" "$repo_dir"; then
            write_warn "Failed to clone ${repo_url} into ${repo_dir}."
            return 1
        fi
    fi

    if ! git -C "$repo_dir" fetch --tags origin; then
        write_warn "Failed to fetch ${repo_ref} from ${repo_url}."
        return 1
    fi
    if ! git -C "$repo_dir" checkout -f "$repo_ref"; then
        write_warn "Failed to checkout ${repo_ref} in ${repo_dir}."
        return 1
    fi
}

build_and_start_from_source() {
    local repo_url="$1"
    local repo_ref="$2"
    local repo_dir="$3"
    local external_port="$4"
    local data_dir="$5"
    local container_name="$6"

    if ! prepare_repo_checkout "$repo_url" "$repo_dir" "$repo_ref"; then
        write_warn "Source-build fallback checkout did not complete."
        return 1
    fi

    write_warn "Falling back to local Docker source build because image pull failed."
    if ! (
        cd "$repo_dir"
        OCTOPUS_PORT="$external_port" \
        OCTOPUS_DATA_DIR="$data_dir" \
        OCTOPUS_CONTAINER_NAME="$container_name" \
        docker compose build \
          --build-arg GOPROXY="${OCTOPUS_GOPROXY:-https://proxy.golang.org,direct}" \
          --build-arg GOSUMDB="${OCTOPUS_GOSUMDB:-sum.golang.org}" \
          --build-arg NPM_CONFIG_REGISTRY="${OCTOPUS_NPM_CONFIG_REGISTRY:-https://registry.npmjs.org/}"

        OCTOPUS_PORT="$external_port" \
        OCTOPUS_DATA_DIR="$data_dir" \
        OCTOPUS_CONTAINER_NAME="$container_name" \
        docker compose up -d
    ); then
        write_warn "Local Docker source build did not complete successfully."
        return 1
    fi

    if ! verify_running_container "$container_name" \
        "Source-build container failed to stay running. Showing the latest logs for ${container_name}:"; then
        write_warn "Octopus source-build container exited right after startup. The installer will continue to the next fallback when available."
        return 1
    fi
}

verify_running_container() {
    local container_name="$1"
    local log_hint="$2"

    if [[ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null || printf 'false')" == "true" ]]; then
        return 0
    fi

    write_warn "$log_hint"
    docker logs --tail 80 "$container_name" >&2 || true
    return 1
}

build_and_start_from_uploaded_binary() {
    local binary_path="$1"
    local external_port="$2"
    local data_dir="$3"
    local container_name="$4"
    local image_tag="$5"

    if [[ ! -f "$binary_path" ]]; then
        fail "OCTOPUS_BINARY_PATH points to a missing file: ${binary_path}"
    fi

    require_command mktemp "Install coreutils or provide a writable temporary directory."

    local temp_dir
    temp_dir="$(mktemp -d)"
    cp "$binary_path" "$temp_dir/octopus"
    cp "$(dirname "$0")/dockerfiles/entrypoint.sh" "$temp_dir/entrypoint.sh"
    sed -i 's/\r$//' "$temp_dir/entrypoint.sh"

    cat > "$temp_dir/Dockerfile" <<'EOF'
FROM debian:bookworm-slim
ENV TZ=Asia/Shanghai \
    PUID=10001 \
    PGID=10001 \
    DATA_DIR=/app/data \
    OCTOPUS_SERVER_HOST=0.0.0.0 \
    OCTOPUS_SERVER_PORT=1088 \
    OCTOPUS_DATABASE_TYPE=sqlite \
    OCTOPUS_DATABASE_PATH=/app/data/data.db \
    OCTOPUS_LOG_LEVEL=info
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata gosu && \
    ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata && \
    groupadd --system --gid 10001 octopus && \
    useradd --system --uid 10001 --gid 10001 --home-dir /app --shell /usr/sbin/nologin octopus && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /app/data && \
    chown -R 10001:10001 /app
COPY octopus /app/octopus
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /app/octopus /entrypoint.sh
EXPOSE 1088
CMD ["/bin/sh", "/entrypoint.sh"]
EOF

    write_warn "Falling back to a local Docker image built from the provided binary because image pull and source build are unavailable."
    docker build -t "$image_tag" "$temp_dir"
    rm -rf "$temp_dir"

    mkdir -p "$data_dir"
    chown -R 10001:10001 "$data_dir" || true

    docker rm -f "$container_name" >/dev/null 2>&1 || true
    local container_id
    container_id="$(docker run -d \
        --name "$container_name" \
        --restart unless-stopped \
        -p "${external_port}:1088" \
        -e OCTOPUS_SERVER_HOST=0.0.0.0 \
        -e OCTOPUS_SERVER_PORT=1088 \
        -e OCTOPUS_DATABASE_TYPE=sqlite \
        -e OCTOPUS_DATABASE_PATH=/app/data/data.db \
        -e OCTOPUS_LOG_LEVEL=info \
        -e DATA_DIR=/app/data \
        -e PUID=10001 \
        -e PGID=10001 \
        -v "${data_dir}:/app/data" \
        "${DOCKER_RUN_SECURITY_ARGS[@]}" \
        "$image_tag")"

    if ! verify_running_container "$container_name" \
        "Binary-fallback container failed to stay running. Showing the latest logs for ${container_name}:"; then
        fail "Octopus binary-fallback container exited right after startup. Fix the data directory permissions or set ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR=true explicitly if you need a temporary compatibility escape hatch."
    fi
}

validate_image_source() {
    local image="$1"
    if is_disabled_dockerhub_image "$image"; then
        fail "Docker Hub installation is disabled for Octopus. Use the official GHCR image or set OCTOPUS_IMAGE to a reachable private or mirrored registry."
    fi
}

pull_image() {
    local image="$1"

    validate_image_source "$image"
    write_info "Pulling Docker image ${image}"
    if docker pull "$image"; then
        printf '%s' "$image"
        return 0
    fi

    return 1
}

warn_if_raw_download_host_is_unstable() {
    if [[ -t 0 ]]; then
        return 0
    fi

    if command -v curl >/dev/null 2>&1; then
        if ! curl -I --silent --show-error --location --max-time 10 https://raw.githubusercontent.com/ >/dev/null 2>&1; then
            write_warn "raw.githubusercontent.com is unreachable from this host right now. If the one-liner download fails, download scripts/install.sh elsewhere and upload it first."
        fi
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
warn_if_raw_download_host_is_unstable

REPO_URL="${OCTOPUS_REPO_URL:-https://github.com/xiaoli0412/octopus-xiaoli-repo.git}"
REPO_REF="${OCTOPUS_REPO_REF:-${DEFAULT_REPO_REF}}"
REPO_DIR="${OCTOPUS_REPO_DIR:-octopus-xiaoli-repo}"
BINARY_PATH="${OCTOPUS_BINARY_PATH:-}"
BINARY_IMAGE_TAG="${OCTOPUS_BINARY_IMAGE_TAG:-octopus-local:installer-fallback}"

if [[ ! -t 0 ]]; then
    write_info "Non-interactive install detected. The installer will try GHCR first, then fall back to a local source build from ${REPO_REF}."
fi

PULLED_IMAGE=""
if ! PULLED_IMAGE="$(pull_image "$OCTOPUS_IMAGE_INPUT")"; then
    if ! build_and_start_from_source "$REPO_URL" "$REPO_REF" "$REPO_DIR" "$EXTERNAL_PORT" "$OCTOPUS_DATA_DIR_INPUT" "$OCTOPUS_CONTAINER_NAME_INPUT"; then
        if [[ -n "$BINARY_PATH" ]]; then
            build_and_start_from_uploaded_binary "$BINARY_PATH" "$EXTERNAL_PORT" "$OCTOPUS_DATA_DIR_INPUT" "$OCTOPUS_CONTAINER_NAME_INPUT" "$BINARY_IMAGE_TAG"
        else
            fail "Unable to pull the GHCR image and the source-backed Docker build did not complete. Re-run with OCTOPUS_BINARY_PATH=<local-octopus-binary> to build a local Docker image from a known-good binary."
        fi
    fi

    write_info "Octopus is starting"
    write_info "UI: http://127.0.0.1:${EXTERNAL_PORT}"
    write_info "Container: ${OCTOPUS_CONTAINER_NAME_INPUT}"
    write_info "Data dir: ${OCTOPUS_DATA_DIR_INPUT}"
    write_info "Source build ref: ${REPO_REF}"
    if [[ -n "$BINARY_PATH" ]]; then
        write_info "Binary fallback: ${BINARY_PATH}"
    fi
    exit 0
fi

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
    "${DOCKER_RUN_SECURITY_ARGS[@]}" \
    "$PULLED_IMAGE")"

if ! verify_running_container "$OCTOPUS_CONTAINER_NAME_INPUT" \
    "Container failed to stay running. Showing the latest logs for ${OCTOPUS_CONTAINER_NAME_INPUT}:"; then
    fail "Octopus container exited right after startup. Check the logs above, especially data directory permissions or registry/image overrides."
fi

write_info "Octopus is starting"
write_info "UI: http://127.0.0.1:${EXTERNAL_PORT}"
write_info "Container: ${OCTOPUS_CONTAINER_NAME_INPUT}"
write_info "Data dir: ${OCTOPUS_DATA_DIR_INPUT}"
write_info "Image: ${PULLED_IMAGE}"
