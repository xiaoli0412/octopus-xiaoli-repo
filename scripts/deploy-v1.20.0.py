#!/usr/bin/env python3
"""Deploy octopus to remote servers with zero data loss.

The servers are managed via docker compose in /root/octopus-xiaoli-repo.
By default this script pulls the latest docker-compose.yml from origin/main
and starts the GHCR pre-built image, preserving all existing volume mounts.
Use --build to force a local build as a fallback.
"""

import argparse
import sys

import paramiko

COMPOSE_DIR = "/root/octopus-xiaoli-repo"
PULL_TIMEOUT = 300  # seconds; image pull is usually fast
BUILD_TIMEOUT = 1200  # seconds; local builds can be slow on small servers


def run_ssh(host: str, user: str, password: str, port: int, command: str, timeout: int = 300):
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=host, port=port, username=user, password=password, timeout=20, banner_timeout=20)
    try:
        stdin, stdout, stderr = client.exec_command(command, timeout=timeout)
        out = stdout.read().decode("utf-8", errors="replace")
        err = stderr.read().decode("utf-8", errors="replace")
        return out, err, stdout.channel.recv_exit_status()
    finally:
        client.close()


def inspect_container(host: str, info: dict) -> dict:
    cmd = (
        "docker ps --format 'table {{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}' ; "
        "echo '---MOUNTS---' ; "
        "docker inspect --format='{{range .Mounts}}{{printf \"%s -> %s\\n\" .Source .Destination}}{{end}}' octopus 2>/dev/null || echo 'NO_OCTOPUS_CONTAINER' ; "
        "echo '---PORT---' ; "
        "docker inspect --format='{{range $p, $_ := .NetworkSettings.Ports}}{{range .}}{{printf \"%s/%s -> %s\\n\" .HostPort (index (split $p \"/\") 0) .HostIp}}{{end}}{{end}}' octopus 2>/dev/null || echo 'NO_OCTOPUS_CONTAINER'"
    )
    out, err, code = run_ssh(host, info["user"], info["password"], info["port"], cmd)
    return {"out": out, "err": err, "code": code}


def deploy_compose(host: str, info: dict, build_local: bool = False) -> dict:
    """Pull latest source in the compose directory and start the container."""
    if build_local:
        deploy_steps = f"""echo "=== Building container locally (output buffered to /tmp/octopus-build.log) ==="
if ! docker compose build > /tmp/octopus-build.log 2>&1; then
    echo "=== Build failed; last 80 lines ==="
    tail -n 80 /tmp/octopus-build.log
    exit 1
fi
echo "=== Build succeeded; starting container ==="
docker compose up -d"""
        timeout = BUILD_TIMEOUT
    else:
        deploy_steps = f"""echo "=== Pulling latest GHCR image ==="
docker compose pull

echo "=== Starting container ==="
docker compose up -d"""
        timeout = PULL_TIMEOUT

    script = f"""set -e
cd "{COMPOSE_DIR}"

echo "=== Current git status ==="
git status --short

echo "=== Pulling latest compose/source ==="
git fetch origin
git reset --hard origin/main

{deploy_steps}

echo "=== Waiting for healthcheck ==="
for i in $(seq 1 120); do
    status=$(docker inspect --format='{{.State.Health.Status}}' octopus 2>/dev/null || echo '')
    if [ "$status" = "healthy" ]; then
        echo "Container healthy"
        exit 0
    fi
    echo "... health status: ${status:-unknown} (iteration $i)"
    sleep 3
done
echo "Timeout waiting for healthy"
docker logs --tail 30 octopus
exit 1
"""
    out, err, code = run_ssh(host, info["user"], info["password"], info["port"], script, timeout=timeout)
    return {"out": out, "err": err, "code": code}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", required=True, help="Target host IP")
    parser.add_argument("--port", type=int, default=22, help="SSH port")
    parser.add_argument("--user", required=True, help="SSH user")
    parser.add_argument("--password", required=True, help="SSH password")
    parser.add_argument("--inspect-only", action="store_true", help="Only inspect current state")
    parser.add_argument("--build", action="store_true", help="Force local docker compose build instead of pulling GHCR image")
    args = parser.parse_args()

    info = {"port": args.port, "user": args.user, "password": args.password}

    print(f"\n========== {args.host} ==========")
    print("--- Current state ---")
    state = inspect_container(args.host, info)
    print(state["out"])
    if state["err"]:
        print("STDERR:", state["err"])

    if args.inspect_only:
        return

    print(f"--- Deploying via compose (mode: {'local-build' if args.build else 'ghcr-pull'}) ---")
    result = deploy_compose(args.host, info, build_local=args.build)
    print(result["out"])
    if result["err"]:
        print("STDERR:", result["err"])
    if result["code"] != 0:
        print(f"DEPLOYMENT FAILED on {args.host}", file=sys.stderr)
        sys.exit(1)
    print(f"--- {args.host} deployed successfully ---")


if __name__ == "__main__":
    main()
