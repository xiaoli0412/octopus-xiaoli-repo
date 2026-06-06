#!/bin/sh
set -e

PUID=${PUID:-10001}
PGID=${PGID:-10001}
DATA_DIR=${DATA_DIR:-/app/data}
ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR=${ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR:-false}

is_truthy() {
    case "$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')" in
        1|true|yes|on)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

can_write_data_dir_as_target() {
    if [ "$PUID" = 0 ] && [ "$PGID" = 0 ]; then
        return 0
    fi

    if command -v su-exec >/dev/null 2>&1; then
        su-exec "$PUID:$PGID" sh -c "test -w \"$DATA_DIR\""
        return $?
    fi

    if command -v gosu >/dev/null 2>&1; then
        gosu "$PUID:$PGID" sh -c "test -w \"$DATA_DIR\""
        return $?
    fi

    return 1
}

target_runtime_is_writable() {
    if [ ! -d "$DATA_DIR" ]; then
        return 1
    fi

    if can_write_data_dir_as_target; then
        return 0
    fi

    return 1
}

case "$PUID" in
    ''|*[!0-9]*)
        echo 'Error: PUID must be a non-negative integer.' >&2
        exit 1
        ;;
esac

case "$PGID" in
    ''|*[!0-9]*)
        echo 'Error: PGID must be a non-negative integer.' >&2
        exit 1
        ;;
esac

if [ "$(id -u)" = 0 ] && { [ "$PUID" != 0 ] || [ "$PGID" != 0 ]; }; then
    if ! chown -R "$PUID:$PGID" "$DATA_DIR"; then
        echo "Warning: unable to change ownership of $DATA_DIR; checking existing write permissions instead." >&2
    fi
fi

cd /app

if [ "$(id -u)" = 0 ] && { [ "$PUID" != 0 ] || [ "$PGID" != 0 ]; } && ! target_runtime_is_writable; then
    if is_truthy "$ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR"; then
        echo "Warning: $DATA_DIR is not writable for the configured runtime user; ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR is enabled, so Octopus will start as root." >&2
        echo 'Hint: fix the host volume ownership or mount a writable data directory to restore unprivileged runtime and remove the fallback override.' >&2
        exec ./octopus start
    fi

    echo "Error: $DATA_DIR is not writable for the configured runtime user; refusing to start Octopus as root by default." >&2
    echo 'Hint: fix the host volume ownership or mount a writable data directory. Set ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR=true only as a temporary compatibility escape hatch.' >&2
    exit 1
fi

if command -v su-exec >/dev/null 2>&1; then
    exec su-exec "$PUID:$PGID" ./octopus start
elif command -v gosu >/dev/null 2>&1; then
    exec gosu "$PUID:$PGID" ./octopus start
else
    if [ "$PUID" != 0 ] || [ "$PGID" != 0 ]; then
        echo 'Error: neither su-exec nor gosu is available; refusing to start Octopus as root when PUID/PGID requests an unprivileged runtime.' >&2
        exit 1
    fi
    exec ./octopus start
fi
