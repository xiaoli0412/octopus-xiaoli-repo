<div align="center">

<img src="web/public/logo.svg" alt="Octopus Logo" width="120" height="120">

### Octopus

**A Simple, Beautiful, and Elegant LLM API Aggregation & Load Balancing Service for Individuals**

 English | [简体中文](README_zh.md)

</div>


## ✨ Features

- 🔀 **Multi-Channel Aggregation** - Connect multiple LLM provider channels with unified management
- 🔑 **Multi-Key Support** - Support multiple API keys for a single channel
- ⚡ **Smart Selection** - Multiple endpoints per channel, smart selection of the endpoint with the shortest delay
- ⚖️ **Load Balancing** - Automatic request distribution for stable and efficient service
- 🔄 **Protocol Conversion** - Seamless conversion between OpenAI Chat / OpenAI Responses / Anthropic API formats
- 💰 **Price Sync** - Automatic model pricing updates
- 🔃 **Model Sync** - Automatic synchronization of available model lists with channels
- 📊 **Analytics** - Comprehensive request statistics, token consumption, and cost tracking
- 🎨 **Elegant UI** - Clean and beautiful web management panel
- 🗄️ **Multi-Database Support** - Support for SQLite, MySQL, PostgreSQL

### ✨ Extra Features (vs upstream)

- 🤖 **GitHub Copilot OAuth** - One-click GitHub Copilot OAuth Device Flow login, no manual token management
- 🌌 **Antigravity (Google Gemini Code Assist)** - OAuth Web Flow integration with automatic project ID retrieval and Gemini request wrapping
- 🧪 **Model Testing UI** - Test channel model connectivity before saving; supports batch testing, 429 treated as pass
- 🔌 **Built-in Providers** - 20+ pre-configured provider templates (OpenAI, Anthropic, Gemini, Zhipu, Volcengine, Copilot, etc.) for one-click channel creation
- 📋 **CC Switch Integration** - Generate `ccswitch://` deep links to import provider config into Claude / Codex / Gemini CLI tools directly from the UI
- 🎯 **Zen Channel** - Smart protocol routing: auto-selects Anthropic / OpenAI Responses / Gemini / OpenAI Chat based on model name prefix
- ⚙️ **API Base URL Setting** - Configure the externally accessible base URL for generated curl examples and client config instructions


## 🚀 Quick Start

### 🐳 Docker

Run directly:

```bash
docker run -d --name octopus -v /path/to/data:/app/data -p 1088:1088 ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.18.5
```

Or use the install script. It keeps `1088` as the default external port, probes the port before startup, auto-switches to a free port in non-interactive runs, tries the official GHCR image first, falls back to a local source-backed Docker build from the matching release tag when GHCR is unreachable, and can finally build a local Docker image from a known-good Linux binary when the server-side source build is still blocked. Docker Hub is no longer an official install source:

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | bash
```

If your SSH host has unstable access to `raw.githubusercontent.com`, prefer the two-step flow below so you can tell whether the failure is in script download or image pulling:

```bash
curl -fsSL -o install-octopus.sh https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh
bash install-octopus.sh
```

Non-interactive custom port example:

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | OCTOPUS_PORT=1088 bash
```

If your network restricts GHCR, you can still pin a reachable private or mirrored registry image explicitly:

```bash
OCTOPUS_IMAGE=registry.example.com/octopus-xiaoli-repo:v1.18.5 bash install-octopus.sh
```

If GHCR is blocked and the server-side source build still fails, provide a known-good Linux binary and let the installer build a local Docker image from it:

```bash
OCTOPUS_BINARY_PATH=/root/octopus-linux-amd64 bash install-octopus.sh
```

If you prefer the manual path, you can still clone the repository and run compose directly:

```bash
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo
docker compose up -d --build
```

If you use the compose file in this repository directly, the default persistent data directory is `./data`.
The Debian Docker build remains available for CI and image publishing, while the install script prefers the prebuilt release image and only falls back to a source-backed Docker build when image pulling fails.
You can also override the default compose runtime parameters without editing the file itself:

```bash
OCTOPUS_PORT=1088 \
OCTOPUS_DATA_DIR=./build/compose-smoke-data \
OCTOPUS_CONTAINER_NAME=octopus-smoke \
docker compose up -d
```

The container startup path now refuses to silently fall back to a root runtime when `PUID` / `PGID` requests an unprivileged user but the mounted data directory is not writable. Fix the host volume ownership instead. If you need a temporary compatibility escape hatch for an existing deployment, set `ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR=true` explicitly and remove it after repairing the mount permissions.


### 📦 Download from Release

Download the binary for your platform from [Releases](https://github.com/xiaoli0412/octopus-xiaoli-repo/releases), then run:

```bash
./octopus start
```

### 🛠️ Build from Source

**Requirements:**
- Go 1.24.4
- Node.js 18+
- pnpm

```bash
# Clone the repository
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo
# Build frontend and sync the exported assets into static/out
cd web && pnpm install && pnpm run build:static && cd ..
# Start the backend service
go run main.go start 
```

> 💡 **Tip**: The frontend build artifacts are embedded into the Go binary, so you must build the frontend before starting the backend.

**Windows Build**

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1
```

This script will build the frontend, sync the exported assets into `static/out`, and generate `build/bin/octopus-windows-amd64.exe`.

**Windows Development Mode**

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1
```

**Linux Development Mode**

```bash
chmod +x ./scripts/dev-linux.sh
./scripts/dev-linux.sh
```

**Linux Server Binary Only**

```bash
chmod +x ./scripts/build.sh
./scripts/build.sh linux-binary
```

This path builds only `build/bin/octopus-linux-x86_64` and skips the frontend and price-update pipeline when you only need a Linux server binary.

**Linux Backend Smoke**

```bash
chmod +x ./scripts/smoke-linux-backend.sh
./scripts/smoke-linux-backend.sh
```

This smoke path builds a temporary local binary, starts a local mock upstream, and verifies the minimal backend flow end-to-end:
- `/`
- `/manifest.json`
- `/healthz`
- `/api/v1/user/login`
- `/api/v1/channel/create`
- `/api/v1/group/create`
- `/api/v1/apikey/create`
- `/v1/chat/completions`

**Docker Compose Runtime Smoke**

```bash
chmod +x ./scripts/smoke-docker-compose.sh
./scripts/smoke-docker-compose.sh
```

This smoke path verifies the repository `docker-compose.yml` end-to-end with a temporary, parameterized runtime:
- `docker compose up -d --build`
- `/healthz`
- `/`
- `/manifest.json`
- compose teardown and cleanup

**Repository Validation Workflow**

The repository also includes a GitHub Actions validation workflow at `.github/workflows/validation.yaml`.
It is intended to cover the remaining Milestone 6 acceptance chain in CI when local Docker or Linux runtime is not available:
- `go build ./...`
- `go test ./internal/op -count=1`
- frontend `pnpm exec tsc --noEmit`
- Linux backend smoke via `scripts/smoke-linux-backend.sh`
- Docker compose runtime smoke via `scripts/smoke-docker-compose.sh`

**Manual Frontend Checklist**

Complex UI acceptance remains `runtime smoke + manual checklist`, not browser E2E automation.
The repository-tracked checklist lives at `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md` and should be run on both desktop and `375px` width before claiming final UI acceptance.
The current UI rollback-and-recovery mainline task is tracked in `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`.

**Backup / Import Contract**

- `/api/v1/setting/export` defaults to a full restore-ready snapshot when `include_secrets` is omitted
- only explicit `include_secrets=false` produces a redacted snapshot
- the daily background dynamic-routing task is a `dynamic summary scan`; it does not persist new thresholds or reorder user routing

**Development Mode**

```bash
cd web && pnpm install && NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:1088" pnpm run dev
## Open a new terminal, start the backend service
go run main.go start
## Access the frontend at
http://localhost:3000
```

### 🔐 Bootstrap Credentials

After first launch, Octopus creates the initial administrator with:

- **Username**: `admin` by default, or `OCTOPUS_ADMIN_USERNAME` if set
- **Password**: `admin` by default, or `OCTOPUS_ADMIN_PASSWORD` if set

If you do not provide `OCTOPUS_ADMIN_PASSWORD`, the built-in bootstrap credentials are `admin / admin`.
After the first successful login with the built-in password, Octopus blocks the rest of the management console until you set a new administrator password.
Changing the username during that first-login flow is optional; changing the password is mandatory.

### 📝 Configuration File

The configuration file is located at `data/config.json` by default and is automatically generated on first startup.

**Complete Configuration Example:**

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "static_dir": "static/out"
  },
  "database": {
    "type": "sqlite",
    "path": "data/data.db"
  },
  "log": {
    "level": "info"
  }
}
```

**Configuration Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `server.host` | Listen address | `0.0.0.0` |
| `server.port` | Server port | `8080` |
| `server.static_dir` | Preferred local static asset directory; falls back to embedded assets when unavailable | `static/out` |
| `database.type` | Database type | `sqlite` |
| `database.path` | Database connection string | `data/data.db` |
| `log.level` | Log level | `info` |

For the `octopus healthcheck` CLI, wildcard listen addresses such as `0.0.0.0` and `::` are automatically mapped to `127.0.0.1` so the probe uses a client-reachable target.

**Database Configuration:**

Three database types are supported:

| Type | `database.type` | `database.path` Format |
|------|-----------------|-----------------------|
| SQLite | `sqlite` | `data/data.db` |
| MySQL | `mysql` | `user:password@tcp(host:port)/dbname` |
| PostgreSQL | `postgres` | `postgresql://user:password@host:port/dbname?sslmode=disable` |

**MySQL Configuration Example:**

```json
{
  "database": {
    "type": "mysql",
    "path": "root:password@tcp(127.0.0.1:3306)/octopus"
  }
}
```

**PostgreSQL Configuration Example:**

```json
{
  "database": {
    "type": "postgres",
    "path": "postgresql://user:password@localhost:5432/octopus?sslmode=disable"
  }
}
```

> 💡 **Tip**: MySQL and PostgreSQL require manual database creation. The application will automatically create the table structure.

### 🌐 Environment Variables

All configuration options can be overridden via environment variables using the format `OCTOPUS_` + configuration path (joined with `_`):

| Environment Variable | Configuration Option |
|---------------------|---------------------|
| `OCTOPUS_SERVER_PORT` | `server.port` |
| `OCTOPUS_SERVER_HOST` | `server.host` |
| `OCTOPUS_SERVER_STATIC_DIR` | `server.static_dir` |
| `OCTOPUS_DATABASE_TYPE` | `database.type` |
| `OCTOPUS_DATABASE_PATH` | `database.path` |
| `OCTOPUS_LOG_LEVEL` | `log.level` |
| `OCTOPUS_GITHUB_PAT` | For rate limiting when getting the latest version (optional) |
| `OCTOPUS_RELAY_MAX_SSE_EVENT_SIZE` | Maximum SSE event size (optional) |

### OAuth Environment Overrides (Optional)

Octopus includes built-in defaults for Copilot OAuth login.
Antigravity OAuth requires `OCTOPUS_ANTIGRAVITY_CLIENT_ID` and `OCTOPUS_ANTIGRAVITY_CLIENT_SECRET` before the web flow can start.
You can still override any OAuth value via environment variables.

| Environment Variable | Default |
|---------------------|---------|
| `OCTOPUS_COPILOT_CLIENT_ID` | `151ef1b1b0345b2351ca` |
| `OCTOPUS_COPILOT_SCOPE` | `copilot` |
| `OCTOPUS_COPILOT_DEVICE_CODE_URL` | `https://github.com/login/device/code` |
| `OCTOPUS_COPILOT_ACCESS_TOKEN_URL` | `https://github.com/login/oauth/access_token` |
| `OCTOPUS_ANTIGRAVITY_CLIENT_ID` | *(required, set via environment variable)* |
| `OCTOPUS_ANTIGRAVITY_CLIENT_SECRET` | *(required, set via environment variable)* |
| `OCTOPUS_ANTIGRAVITY_AUTHORIZE_URL` | `https://accounts.google.com/o/oauth2/v2/auth` |
| `OCTOPUS_ANTIGRAVITY_TOKEN_URL` | `https://oauth2.googleapis.com/token` |
| `OCTOPUS_ANTIGRAVITY_SCOPE` | `https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile` |

Notes:
- Antigravity uses Octopus's own callback endpoint (`/api/v1/channel/antigravity/oauth/callback`) and polling flow.
- `ANTIGRAVITY_*` / `COPILOT_*` aliases are also supported for compatibility.

## 📸 Screenshots

### 🖥️ Desktop

<div align="center">
<table>
<tr>
<td align="center"><b>Dashboard</b></td>
<td align="center"><b>Channel Management</b></td>
<td align="center"><b>Group Management</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/desktop-home.png" alt="Dashboard" width="400"></td>
<td><img src="web/public/screenshot/desktop-channel.png" alt="Channel" width="400"></td>
<td><img src="web/public/screenshot/desktop-group.png" alt="Group" width="400"></td>
</tr>
<tr>
<td align="center"><b>Price Management</b></td>
<td align="center"><b>Logs</b></td>
<td align="center"><b>Settings</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/desktop-price.png" alt="Price Management" width="400"></td>
<td><img src="web/public/screenshot/desktop-log.png" alt="Logs" width="400"></td>
<td><img src="web/public/screenshot/desktop-setting.png" alt="Settings" width="400"></td>
</tr>
<tr>
<td align="center"><b>Curl usage</b></td>
<td align="center"><b>CC Switch</b></td>
<td align="center"><b> </b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/desktop-api-curl.png" alt="Curl Usage" width="400"></td>
<td><img src="web/public/screenshot/desktop-api-cc.png" alt="CC Switch" width="400"></td>
<td></td>
</tr>
</table>
</div>

### 📱 Mobile

<div align="center">
<table>
<tr>
<td align="center"><b>Home</b></td>
<td align="center"><b>Channel</b></td>
<td align="center"><b>Group</b></td>
<td align="center"><b>Price</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/mobile-home.png" alt="Mobile Home" width="140"></td>
<td><img src="web/public/screenshot/mobile-channel.png" alt="Mobile Channel" width="140"></td>
<td><img src="web/public/screenshot/mobile-group.png" alt="Mobile Group" width="140"></td>
<td><img src="web/public/screenshot/mobile-price.png" alt="Mobile Price" width="140"></td>
</tr>
<tr>
<td align="center"><b>Logs</b></td>
<td align="center"><b>Settings</b></td>
<td align="center"><b>Curl usage</b></td>
<td align="center"><b>CC Switch</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/mobile-log.png" alt="Mobile Logs" width="140"></td>
<td><img src="web/public/screenshot/mobile-setting.png" alt="Mobile Settings" width="140"></td>
<td><img src="web/public/screenshot/mobile-api-curl.png" alt="Mobile Curl Usage" width="140"></td>
<td><img src="web/public/screenshot/mobile-api-cc.png" alt="Mobile CC Switch" width="140"></td>
<td></td>
<td></td>
<td></td>
<td></td>
</tr>
</table>
</div>


## 📖 Documentation

### 📡 Channel Management

Channels are the basic configuration units for connecting to LLM providers.

**Base URL Guide:**

The program automatically appends API paths based on channel type. You only need to provide the base URL:

| Channel Type | Auto-appended Path | Base URL | Full Request URL Example |
|--------------|-------------------|----------|--------------------------|
| OpenAI Chat | `/chat/completions` | `https://api.openai.com/v1` | `https://api.openai.com/v1/chat/completions` |
| OpenAI Responses | `/responses` | `https://api.openai.com/v1` | `https://api.openai.com/v1/responses` |
| Anthropic | `/messages` | `https://api.anthropic.com/v1` | `https://api.anthropic.com/v1/messages` |
| Gemini | `/models/:model:generateContent` | `https://generativelanguage.googleapis.com/v1beta` | `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent` |

> 💡 **Tip**: No need to include specific API endpoint paths in the Base URL - the program handles this automatically.

---

### 📁 Group Management

Groups aggregate multiple channels into a unified external model name.

**Core Concepts:**

- **Group name** is the model name exposed by the program
- When calling the API, set the `model` parameter to the group name

**Load Balancing Modes:**

| Mode | Description |
|------|-------------|
| 🔄 **Round Robin** | Cycles through channels sequentially for each request |
| 🎲 **Random** | Randomly selects an available channel for each request |
| 🛡️ **Failover** | Prioritizes high-priority channels, switches to lower priority only on failure |
| ⚖️ **Weighted** | Distributes requests based on configured channel weights |

> 💡 **Example**: Create a group named `gpt-4o`, add multiple providers' GPT-4o channels to it, then access all channels via a unified `model: gpt-4o`.

---

### 💰 Price Management

Manage model pricing information in the system.

**Data Sources:**

- The system periodically syncs model pricing data from [models.dev](https://github.com/sst/models.dev)
- When creating a channel, if the channel contains models not in models.dev, the system automatically creates pricing information for those models on this page, so this page displays models that haven't had their prices fetched from upstream, allowing users to set prices manually
- Manual creation of models that exist in models.dev is also supported for custom pricing

**Price Priority:**

| Priority | Source | Description |
|:--------:|--------|-------------|
| 🥇 High | This Page | Prices set by user in price management page |
| 🥈 Low | models.dev | Auto-synced default prices |

> 💡 **Tip**: To override a model's default price, simply set a custom price for it in the price management page.

---

### ⚙️ Settings

Global system configuration.

**Statistics Save Interval (minutes):**

Since the program handles numerous statistics, writing to the database on every request would impact read/write performance. The program uses this strategy:

- Statistics are first stored in **memory**
- Periodically **batch-written** to the database at the configured interval

> ⚠️ **Important**: When exiting the program, use proper shutdown methods (like `Ctrl+C` or sending `SIGTERM` signal) to ensure in-memory statistics are correctly written to the database. **Do NOT use `kill -9` or other forced termination methods**, as this may result in statistics data loss.

---

## 🔌 Client Integration

### OpenAI SDK

```python
from openai import OpenAI
import os

client = OpenAI(   
base_url="http://127.0.0.1:1088/v1",
    api_key="sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg", 
)
completion = client.chat.completions.create(
    model="octopus-openai",  # Use the correct group name
    messages = [
        {"role": "user", "content": "Hello"},
    ],
)
print(completion.choices[0].message.content)
```

### Claude Code

Edit `~/.claude/settings.json`

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:1088",
    "ANTHROPIC_AUTH_TOKEN": "sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "ANTHROPIC_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_SMALL_FAST_MODEL": "octopus-haiku-4-5",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "octopus-haiku-4-5"
  }
}
```

### Codex

Edit `~/.codex/config.toml`

```toml
model = "octopus-codex" # Use the correct group name

model_provider = "octopus"

[model_providers.octopus]
name = "octopus"
base_url = "http://127.0.0.1:1088/v1"
```

Edit `~/.codex/auth.json`

```json
{
  "OPENAI_API_KEY": "sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg"
}
```

### CC Switch (One-click CLI Import)

In the web UI, click the **API Docs** button → **CC Switch** tab to generate a deep link that imports your Octopus provider config directly into CLI tools.

Supported tools: **Claude Code**, **Codex**, **Gemini CLI**

For Claude Code, you can also configure separate model mappings for Haiku / Sonnet / Opus roles.

---

## 🤝 Acknowledgments

- 🙏 [looplj/axonhub](https://github.com/looplj/axonhub) - The LLM API adaptation module in this project is directly derived from this repository
- 📊 [sst/models.dev](https://github.com/sst/models.dev) - AI model database providing model pricing data

