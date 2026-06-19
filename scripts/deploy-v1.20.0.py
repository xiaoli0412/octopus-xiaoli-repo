#!/usr/bin/env python3
"""Deploy octopus v1.20.0 to remote servers with zero data loss.

The servers are managed via docker compose in /root/octopus-xiaoli-repo.
This script pulls the latest source and rebuilds the container, preserving
all existing volume mounts.
"""

import argparse
import sys

import paramiko

COMPOSE_DIR = "/root/octopus-xiaoli-repo"
DEPLOY_TIMEOUT = 1200  # seconds; docker builds can be slow on small servers


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


def deploy_compose(host: str, info: dict) -> dict:
    """Pull latest source in the compose directory and rebuild the container."""
    script = f"""set -e
cd "{COMPOSE_DIR}"

echo "=== Current git status ==="
git status --short

echo "=== Pulling latest source ==="
git fetch origin
git reset --hard origin/main

echo "=== Building container (output buffered to /tmp/octopus-build.log) ==="
if ! docker compose build > /tmp/octopus-build.log 2>&1; then
    echo "=== Build failed; last 80 lines ==="
    tail -n 80 /tmp/octopus-build.log
    exit 1
fi
echo "=== Build succeeded; starting container ==="
docker compose up -d

echo "=== Waiting for healthcheck ==="
for i in $(seq 1 60); do
    if docker ps --filter "name=octopus" --filter "health=healthy" --format '{{.Names}}' | grep -q "octopus"; then
        echo "Container healthy"
        exit 0
    fi
    sleep 3
done
echo "Timeout waiting for healthy"
docker logs --tail 30 octopus
exit 1
"""
    out, err, code = run_ssh(host, info["user"], info["password"], info["port"], script, timeout=DEPLOY_TIMEOUT)
    return {"out": out, "err": err, "code": code}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", required=True, help="Target host IP")
    parser.add_argument("--port", type=int, default=22, help="SSH port")
    parser.add_argument("--user", required=True, help="SSH user")
    parser.add_argument("--password", required=True, help="SSH password")
    parser.add_argument("--inspect-only", action="store_true", help="Only inspect current state")
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

    print("--- Deploying v1.20.0 via compose ---")
    result = deploy_compose(args.host, info)
    print(result["out"])
    if result["err"]:
        print("STDERR:", result["err"])
    if result["code"] != 0:
        print(f"DEPLOYMENT FAILED on {args.host}", file=sys.stderr)
        sys.exit(1)
    print(f"--- {args.host} deployed successfully ---")


if __name__ == "__main__":
    main()
