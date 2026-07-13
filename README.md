<div align="center">

<img src="web/public/logo.svg" alt="Octopus Logo" width="120" height="120">

# Octopus

### Enterprise-Grade LLM API Aggregation, Load Balancing & Governance Platform

[![Release](https://img.shields.io/github/v/release/xiaoli0412/octopus-xiaoli-repo?style=flat-square)](https://github.com/xiaoli0412/octopus-xiaoli-repo/releases)
[![License](https://img.shields.io/github/license/xiaoli0412/octopus-xiaoli-repo?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://github.com/xiaoli0412/octopus-xiaoli-repo/pkgs/container/octopus-xiaoli-repo)

English | [简体中文](README_zh.md)

</div>

---

## Overview

Octopus is a production-ready LLM API gateway that aggregates multiple upstream providers (OpenAI, Anthropic, Gemini, Copilot, etc.) behind a single unified API surface. It provides intelligent load balancing, circuit breaking, health-aware routing, fine-grained access control, full audit logging, and comprehensive observability — designed for teams that need enterprise-grade reliability for their LLM infrastructure.

### Why Octopus?

| Capability | What It Means |
|---|---|
| **Multi-Provider Aggregation** | Connect dozens of LLM channels under one API key. Clients see a single endpoint; Octopus handles routing, failover, and protocol conversion transparently. |
| **Health-Aware Routing** | Active health probes + circuit breakers + latency-aware selection ensure requests go to the fastest healthy upstream. No more manual failover. |
| **Enterprise Security** | Per-key channel/group/capability/IP-CIDR permissions, JWT dual-key rotation, audit logging with before/after diff, `/metrics` auth, and read-only container hardening. |
| **Full Observability** | Prometheus metrics, OpenTelemetry tracing, structured logging with rotation, and a built-in audit log UI — everything you need for production monitoring. |
| **Operational Excellence** | Automated SQLite backups, one-click data export, request replay for debugging, K8s deployment templates, and zero-downtime rolling upgrades. |
| **Protocol Conversion** | Seamless translation between OpenAI Chat/Responses, Anthropic Messages, Gemini, and Embeddings APIs — clients can use any SDK against any upstream. |

---

## Architecture

```mermaid
flowchart TB
    subgraph Clients
        A[OpenAI SDK] & B[Claude Code] & C[Codex CLI] & D[Gemini CLI]
    end

    subgraph Octopus Gateway
        E[Next.js Admin UI] --> F[Gin HTTP Server]
        F --> G[Auth & Rate Limit Middleware]
        G --> H[Relay Engine]
        H --> I[Error Classifier & Retry]
        I --> J[Load Balancer]
        J --> K[Health-Aware Selector]
        H --> L[Transformer Adapters]
        L --> M[Outbound Providers]
        H --> N[Response Cache]
        F --> O[Operations Layer]
        O --> P[(SQLite / MySQL / PostgreSQL)]
        O --> Q[Backup Task]
        O --> R[Cost Alert Webhook]
        F --> S[Metrics & Tracing]
    end

    subgraph Upstream Providers
        T[OpenAI] & U[Anthropic] & V[Gemini] & W[Copilot] & X[Volcengine]
    end

    A & B & C & D -->|API Key + /v1/*| F
    M --> T & U & V & W & X
    S -->|Prometheus / OTLP| Grafana[Monitoring Stack]
```

### Core Components

| Component | Location | Responsibility |
|---|---|---|
| **Relay Engine** | `internal/relay/` | Request routing, error classification, differentiated retry (429 backoff / 401-403 failover / 5xx failover / timeout failover) |
| **Load Balancer** | `internal/relay/balancer/` | Round-robin, random, failover, weighted, `least_latency`, `health_aware` modes with circuit breaker |
| **Health Checker** | `internal/op/health_check_task.go` | Background HEAD + optional LLM probe, auto-recovery on circuit breaker |
| **Transformer** | `internal/transformer/` | Inbound (OpenAI/Anthropic/Gemini) and outbound (OpenAI/Anthropic/Gemini/Copilot/Volcengine) protocol conversion |
| **Rust FFI Core** | `rust/core/` + `internal/rustbridge/` | Token counting, SSE stream parsing, balancer selection, streaming aggregation — Go fallbacks available |
| **Observability** | `internal/observability/` | Prometheus metrics, OTel tracing, audit logging |
| **Operations** | `internal/op/` | Channel/group/APIKey management, stats, backup, cost alerts, response cache |
| **Admin UI** | `web/` | Next.js 16 dashboard with audit log viewer, batch operations, CSV export, hotkey support |

---

## Quick Start

### Docker (Recommended)

```bash
# One-line install
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | bash
```

Or manual docker-compose:

```bash
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo
docker compose up -d
```

The service will be available at `http://localhost:1088`.

**Custom configuration:**

```bash
OCTOPUS_PORT=1088 \
OCTOPUS_DATA_DIR=/data/octopus \
docker compose up -d
```

**Direct GHCR image:**

```bash
docker run -d --name octopus \
  -v /path/to/data:/app/data \
  -p 1088:1088 \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:latest
```

### Kubernetes

Production K8s templates are included in `deploy/k8s/`:

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/pvc.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

Includes liveness/readiness probes, resource limits, `runAsNonRoot`, `readOnlyRootFilesystem`, and `cap_drop: ALL`.

### Binary

Download from [Releases](https://github.com/xiaoli0412/octopus-xiaoli-repo/releases):

```bash
./octopus start
```

### Build from Source

**Requirements:** Go 1.25.0+, Node.js 22+, pnpm 10+, Rust toolchain

```bash
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo

# Build frontend
cd web && pnpm install && pnpm run build:static && cd ..

# Build Rust FFI core
./scripts/build.sh rust-core

# Start
go run main.go start
```

---

## Bootstrap Credentials

On first launch, Octopus creates an administrator account:

- **Username:** `admin` (or `OCTOPUS_ADMIN_USERNAME`)
- **Password:** Auto-generated, printed in startup logs (or set via `OCTOPUS_ADMIN_PASSWORD`)

```bash
docker logs octopus 2>&1 | grep "bootstrap"
```

After first login, a mandatory password change is required before the management console becomes accessible.

---

## Configuration

### Configuration File

Located at `data/config.json`, auto-generated on first start:

```json
{
  "server": { "host": "0.0.0.0", "port": 8080, "static_dir": "static/out" },
  "database": { "type": "sqlite", "path": "data/data.db" },
  "log": { "level": "info" }
}
```

### Environment Variables

All config options can be overridden via `OCTOPUS_` prefixed environment variables:

| Variable | Description | Default |
|---|---|---|
| `OCTOPUS_SERVER_HOST` | Listen address | `0.0.0.0` |
| `OCTOPUS_SERVER_PORT` | Server port | `8080` |
| `OCTOPUS_DATABASE_TYPE` | `sqlite` / `mysql` / `postgres` | `sqlite` |
| `OCTOPUS_DATABASE_PATH` | DB connection string | `data/data.db` |
| `OCTOPUS_LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `info` |
| `OCTOPUS_LOG_FILE` | Log file path (enables lumberjack rotation) | *(stderr only)* |
| `OCTOPUS_SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `30s` |
| `OCTOPUS_ADMIN_USERNAME` | Initial admin username | `admin` |
| `OCTOPUS_ADMIN_PASSWORD` | Initial admin password | *(auto-generated)* |
| `OCTOPUS_GITHUB_PAT` | GitHub PAT for version checks | — |

### Database Support

| Type | Connection String Format |
|---|---|
| SQLite | `data/data.db` |
| MySQL | `user:password@tcp(host:port)/dbname` |
| PostgreSQL | `postgresql://user:password@host:port/dbname?sslmode=disable` |

> MySQL and PostgreSQL require manual database creation. Tables are auto-migrated on startup.

---

## Enterprise Features

### Security & Access Control

- **API Key Permission Granularity:** Per-key `AllowedChannels`, `AllowedGroups`, `AllowedCapabilities` (chat/embedding/response/message), and `AllowedIPCIDRs` — four-dimensional access control.
- **JWT Dual-Key Rotation:** Primary + secondary secrets for zero-downtime key rotation via `POST /api/v1/user/rotate-secret`.
- **Metrics Authentication:** `/metrics` endpoint supports Bearer token + IP allowlist (`metrics_auth_token`, `metrics_ip_allowlist`).
- **Audit Logging:** All sensitive operations (login, CRUD on channels/groups/APIKeys, config changes) are logged with before/after JSON diff. Sensitive fields are redacted.
- **Container Hardening:** `read_only: true`, `no-new-privileges`, `cap_drop: ALL`, non-root user (UID 10001), tmpfs for `/tmp`.

### Reliability & Resilience

- **Error Classification & Differentiated Retry:** 429 → exponential backoff + jitter (same key, max 2 retries); 401/403 → mark key invalid + failover; 5xx → immediate failover; 408/504 → timeout failover; other 4xx → direct return.
- **Circuit Breaker:** Per-channel circuit breaker with configurable thresholds. Auto-recovery on successful health probe.
- **Health-Aware Routing:** `least_latency` and `health_aware` balancer modes aggregate success rate, latency, and circuit breaker state to select the optimal channel.
- **Active Health Probes:** Background task (5-min interval, configurable) performs HEAD probes on base URLs with optional LLM 1-token probes. Results feed back into routing decisions.
- **Response Cache:** Non-streaming requests support SHA256-keyed LRU cache with TTL and concurrent-safe access. `X-Octopus-Cache: HIT/MISS` header indicates cache status.
- **Graceful Shutdown:** Context + timeout (default 30s), concurrent cleanup of background tasks, in-memory stats flushed to DB before exit.

### Observability

- **Prometheus Metrics** (`/metrics`): `relay_requests_total`, `relay_duration_seconds` (histogram), `channel_health`, `circuit_breaker_state`, `token_throughput_total`, `http_client_pool_idle`, and more.
- **OpenTelemetry Tracing:** Full relay span chain (`relay.handle` → `balancer.pick` → `outbound.forward` → `stream.aggregate`), OTLP export, W3C TraceContext propagation.
- **Structured Logging:** Configurable level, optional file output with lumberjack rotation (MaxSize/MaxBackups/MaxAge).
- **Audit Log UI:** Web panel with time/user/action/resource filtering, before/after JSON diff viewer with field-level highlighting.

### Operations & Maintenance

- **Automated Backups:** SQLite database auto-backed up to `data/backups/backup-<timestamp>.db`. Configurable retention count, old backups auto-deleted.
- **Data Export:** Statistics CSV export by channel/model/APIKey dimension with time range filtering (`GET /api/v1/stats/export`).
- **Batch Operations:** Bulk enable/disable/delete for channels, groups, and API keys (max 100 per batch, each action audit-logged).
- **Request Replay:** `POST /api/v1/log/replay/:id` re-runs a logged request through the relay pipeline for debugging (marked `X-Octopus-Replay: true`, excluded from stats).
- **Cost Alert Webhooks:** Per-API-key cost thresholds (50%/80%/100%) with deduplication. Supports Slack, Feishu, and DingTalk message formats.
- **Version Endpoint:** `GET /version` returns version/commit/buildTime/goVersion (no auth required).

---

## Client Integration

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:1088/v1",
    api_key="sk-octopus-xxxxxxxxxxxx",
)
completion = client.chat.completions.create(
    model="gpt-4o",  # group name
    messages=[{"role": "user", "content": "Hello"}],
)
```

### Claude Code

```json
// ~/.claude/settings.json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:1088",
    "ANTHROPIC_AUTH_TOKEN": "sk-octopus-xxxxxxxxxxxx",
    "ANTHROPIC_MODEL": "claude-sonnet-4-5"
  }
}
```

### Codex CLI

```toml
# ~/.codex/config.toml
model = "gpt-4o"
[model_providers.octopus]
base_url = "http://127.0.0.1:1088/v1"
```

### CC Switch (One-Click Import)

The web UI's **API Docs** → **CC Switch** tab generates deep links that import your Octopus provider config directly into Claude Code, Codex, or Gemini CLI.

---

## Load Balancing Modes

| Mode | Description | Use Case |
|---|---|---|
| **Round Robin** | Cycles through channels sequentially | Equal-cost providers |
| **Random** | Random selection | Simple distribution |
| **Failover** | Priority-ordered, falls back on failure | Primary + backup setup |
| **Weighted** | Distribution based on configured weights | Mixed-capacity providers |
| **Least Latency** | Selects lowest-latency channel | Performance-critical workloads |
| **Health Aware** | Aggregates success rate + latency + circuit state | Maximum reliability |

---

## Upgrade Guide

### Upgrading to v1.25

1. **Backup:** Use Settings → Backup or `GET /api/v1/setting/export` to take a full snapshot.
2. **Docker:** `docker compose pull && docker compose up -d`
3. **Database Migrations:** Run automatically on startup. New migrations: audit log table, API key permissions, backup metadata.
4. **API Key Permissions:** Existing keys have no restrictions by default (backward compatible). Configure per-key permissions in Settings → API Key.
5. **Health Check Task:** Enabled by default (5-min interval). Configure in Settings → System.
6. **Metrics Auth:** If `metrics_auth_token` is not set, `/metrics` remains open (backward compatible).

### Version Compatibility

- v1.25 is backward compatible with v1.24 data.
- Rust FFI core is bundled in release binaries. Source builds require Rust toolchain.
- Go 1.25.0+ required for source builds.

---

## Monitoring & Alerting

### Prometheus Metrics

```
relay_requests_total{channel,group,model,status}
relay_duration_seconds_bucket{channel,group,model}
channel_health{channel}
circuit_breaker_state{channel}
token_throughput_total{type,prompt,completion}
http_client_pool_idle
```

### Grafana Dashboard

Import the Prometheus datasource and create panels for:
- Request rate and error rate by channel
- P50/P95/P99 latency distribution
- Circuit breaker state heatmap
- Token throughput trends
- Cache hit rate

### Alerting Rules

```yaml
# Prometheus Alertmanager
- alert: ChannelCircuitOpen
  expr: circuit_breaker_state == 1
  for: 5m
  labels: { severity: critical }

- alert: HighErrorRate
  expr: rate(relay_requests_total{status="error"}[5m]) > 0.1
  for: 10m
  labels: { severity: warning }
```

---

## Project Structure

```
octopus-xiaoli-repo/
├── cmd/                    # CLI entrypoints (start, healthcheck, version)
├── internal/
│   ├── conf/               # Configuration
│   ├── db/migrate/         # Database migrations (001-020)
│   ├── model/              # Data models
│   ├── op/                 # Operations layer (channels, groups, APIKeys, stats, backup, alerts)
│   ├── relay/              # Relay engine, balancer, error classifier
│   ├── rustbridge/         # Rust FFI bindings
│   ├── server/             # HTTP server, handlers, middleware
│   ├── observability/      # Metrics, tracing, audit
│   ├── task/               # Background tasks
│   ├── transformer/        # Protocol conversion adapters
│   └── utils/              # Shared utilities
├── rust/core/              # Rust FFI core library
├── web/                    # Next.js admin UI
├── deploy/k8s/             # Kubernetes deployment templates
├── scripts/                # Build, deploy, smoke test scripts
└── docker-compose.yml      # Docker Compose configuration
```

---

## Development

```bash
# Frontend dev server (port 3000)
cd web && pnpm install && pnpm run dev

# Backend dev server (port 1088)
go run main.go start

# Run tests
go test ./internal/op ./internal/server/handlers ./internal/relay/... -count=1
cd web && pnpm run test

# Windows build
powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1

# Linux binary
./scripts/build.sh linux-binary
```

---

## Acknowledgments

- [looplj/axonhub](https://github.com/looplj/axonhub) — LLM API adaptation module
- [sst/models.dev](https://github.com/sst/models.dev) — AI model pricing database
- [bestruirui/octopus](https://github.com/bestruirui/octopus) — Original project

---

## License

[MIT](LICENSE)
