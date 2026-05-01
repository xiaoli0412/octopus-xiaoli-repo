#!/bin/sh
set -e

PUID=${PUID:-10001}
PGID=${PGID:-10001}
DATA_DIR=${DATA_DIR:-/app/data}

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
    chown -R "$PUID:$PGID" "$DATA_DIR"
fi

cd /app

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
