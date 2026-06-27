<div align="center">

<img src="web/public/logo.svg" alt="Octopus Logo" width="120" height="120">

### Octopus

**为个人打造的简单、美观、优雅的 LLM API 聚合与负载均衡服务**

简体中文 | [English](README.md)

</div>


## ✨ 特性

### 核心

- 🔀 **多渠道聚合** - 支持接入多个 LLM 供应商渠道，统一管理
- 🔑 **多 Key 支持** - 单渠道支持配置多 Key，具备智能轮换
- ⚡ **智能优选** - 单渠道多端点，智能选择延迟最小的端点请求
- ⚖️ **负载均衡** - 支持轮询、随机、故障转移、加权四种分发模式
- 🔄 **协议互转** - 支持 OpenAI Chat / OpenAI Responses / Anthropic Messages / Gemini / Embeddings 互相转换
- 💰 **价格同步** - 自动从 models.dev 等公开数据源同步模型价格
- 🔃 **模型同步** - 自动与渠道同步可用模型列表
- 📊 **数据统计** - 全面的请求统计、Token 消耗、费用追踪
- 🗄️ **多数据库支持** - 支持 SQLite、MySQL、PostgreSQL

### v1.23 亮点

- 🦀 **Rust FFI 进入生产热路径** - Token 计数、SSE 流解析、负载均衡选择、流式响应聚合均可走 Rust FFI 核心；非 Rust 构建仍可通过环境开关回退到 Go 实现
- 🔗 **上游深链接 V2** - 深度对接 New API / sub2API / OpenAI Compatible 上游：真实端点探测、模型/Key/分组发现、余额/订阅追踪、上游 Key 创建、自动签到、余额预警、模型连通性测试，并一键同步到本地渠道/分组/价格
- 🚦 **API Key 限流** - 为每个分发 Key 配置 RPM（每分钟请求数）、TPM（每分钟 Token 数）、Daily（每日请求数）配额，基于内存令牌桶与滑动窗口中间件实现
- 🐳 **Docker 优先部署与更新** - 一键安装脚本优先拉取 GHCR 镜像，失败时回退到源码 Docker 构建；容器内展示一键复制更新命令 `docker compose pull && docker compose up -d`；小内存服务器还可使用已知可用二进制构建本地镜像
- 🌐 **多语言 UI** - Web 管理台支持英文、简体中文、繁体中文

### 更多扩展（相对上游项目）

- 🤖 **GitHub Copilot OAuth** - 一键 GitHub Copilot OAuth Device Flow 登录
- 🌌 **Antigravity（Google Gemini Code Assist）** - OAuth Web Flow 集成，自动获取项目 ID
- 🧪 **模型测试 UI** - 保存渠道前可先测试模型连通性；支持批量测试，429 视为通过
- 🔌 **内置供应商模板** - 20+ 预设供应商配置（OpenAI、Anthropic、Gemini、智谱、火山引擎、Copilot 等）
- 📋 **CC Switch 集成** - 在 UI 中直接生成 `ccswitch://` 深链接，一键将供应商配置导入 Claude / Codex / Gemini CLI 工具
- 🎯 **Zen 渠道** - 根据模型名前缀自动选择协议路由
- ⚙️ **API 基础地址设置** - 配置对外可访问的基础地址，用于生成 curl 示例和引导内容


## 🏗️ 架构

```mermaid
flowchart LR
    subgraph Client
        A[OpenAI SDK / Claude Code / Codex / Gemini CLI]
    end
    subgraph Octopus
        B[Next.js UI] --> C[Gin Server]
        C --> D[Relay Engine]
        D --> E[Rust FFI Core<br/>token count / JSON / SSE]
        D --> F[Transformer Adapters]
        F --> G[Outbound Providers]
        C --> H[Operations Layer]
        H --> I[(Database)]
    end
    subgraph Upstream
        J[New API / sub2API / OpenAI Compatible]
    end
    subgraph Providers
        K[OpenAI / Anthropic / Gemini / Volcengine / Copilot]
    end
    A -->|API Key + /v1/*| C
    G --> K
    H -->|probe / sync / checkin| J
```

关键组件：

- **Relay Engine** - 将请求路由到选定的渠道/分组，应用限流并记录用量。
- **Rust FFI Core**（`rust/core/` + `internal/rustbridge/`）- 负责 token 计数、JSON 解析/转换、SSE aggregate。
- **Transformer Adapters** - 入站适配器（OpenAI Chat/Responses、Anthropic、Embeddings）与出站适配器（OpenAI、Anthropic、Gemini、火山引擎、Copilot、Antigravity、Zen）。
- **Upstream Gateway** - 探测/同步上游站点、管理上游 Key、自动签到、余额告警。
- **Load Balancer** - 轮询、随机、故障转移、加权。
- **Statistics & Logs** - 内存聚合后按配置周期批量写入数据库。


## 🚀 快速开始

### 🐳 Docker（推荐）

一键安装：

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | bash
```

脚本默认使用 `1088` 作为外部端口，启动前会先探测端口占用；如果 `1088` 已被占用，非交互场景会自动切换到可用端口并继续安装。未显式设置 `OCTOPUS_DATA_DIR` 时，脚本会优先复用现有 `octopus` 容器的 `/app/data` 挂载目录，避免升级时误生成空库。脚本会先尝试拉取 GHCR 官方镜像；如果 GHCR 返回 `denied` 或不可达，会回退到对应 release tag 的本地源码支撑 Docker 构建；如果服务器侧源码 Docker 构建仍被阻塞，还可以基于已知可用的 Linux 二进制继续构建本地 Docker 镜像。Docker Hub 已不再作为官方安装来源。

如果你的 SSH 主机访问 `raw.githubusercontent.com` 不稳定，优先改成“两步执行”：

```bash
curl -fsSL -o install-octopus.sh https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh
bash install-octopus.sh
```

非交互自定义端口示例：

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | OCTOPUS_PORT=1088 bash
```

如果所在网络对 GHCR 有限制，也可以显式指定一个可达的私有镜像或镜像代理：

```bash
OCTOPUS_IMAGE=registry.example.com/octopus-xiaoli-repo:v1.23.1 bash install-octopus.sh
```

如果 GHCR 不通，且服务器侧源码 Docker 构建仍然失败，也可以直接提供一个已知可用的 Linux 二进制，再让安装脚本继续构建本地 Docker 镜像：

```bash
OCTOPUS_BINARY_PATH=/root/octopus-linux-amd64 bash install-octopus.sh
```

只有在确认当前服务器可以拉取 GHCR 包时，才建议直接使用 `docker run`：

```bash
docker run -d --name octopus -v /path/to/data:/app/data -p 1088:1088 ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.23.1
```

如果你仍然希望手动安装，也可以继续使用仓库内 compose：

```bash
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo
docker compose up -d --build
```

如果直接使用当前仓库中的 compose 文件，默认持久化目录为 `./data`。升级已有服务器时，应先把 `OCTOPUS_DATA_DIR` 指向真实旧数据目录；一键安装脚本在能检查到现有 `octopus` 容器挂载时会自动处理这一点。

也可以不修改 compose 文件本身，直接覆盖默认运行参数：

```bash
OCTOPUS_PORT=1088 \
OCTOPUS_DATA_DIR=./build/compose-smoke-data \
OCTOPUS_CONTAINER_NAME=octopus-smoke \
docker compose up -d
```

现在容器启动链默认不会在 `PUID` / `PGID` 指定非 root 运行、但挂载数据目录不可写时静默退回 root。正确做法是修复宿主机卷权限；如果必须临时兼容旧环境，可显式设置 `ALLOW_ROOT_FALLBACK_ON_DATA_DIR_ERROR=true`，但应在卷权限修复后尽快移除。


### 📦 从 Release 下载

从 [Releases](https://github.com/xiaoli0412/octopus-xiaoli-repo/releases) 下载对应平台的二进制文件，然后运行：

```bash
./octopus start
```

### 🛠️ 源码运行

**环境要求：**
- Go 1.24.4+
- Node.js 18+
- pnpm
- Rust 工具链（cargo、rustc）- 用于编译 Rust FFI 核心

```bash
# 克隆项目
git clone https://github.com/xiaoli0412/octopus-xiaoli-repo.git
cd octopus-xiaoli-repo
# 构建前端并将导出产物同步到 static/out
cd web && pnpm install && pnpm run build:static && cd ..
# 构建 Rust FFI 核心
./scripts/build.sh rust-core
# 启动后端服务
go run main.go start 
```

> 💡 **提示**：前端构建产物会被嵌入到 Go 二进制文件中，所以必须先构建前端再启动后端。

**Windows 构建**

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1
```

该脚本会构建前端、将静态导出产物同步到 `static/out`、编译 Rust FFI 核心，并生成 `build/bin/octopus-windows-amd64.exe`。

**Windows 开发模式**

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1
```

**Linux 开发模式**

```bash
chmod +x ./scripts/dev-linux.sh
./scripts/dev-linux.sh
```

**Linux 服务器二进制（轻量路径）**

```bash
chmod +x ./scripts/build.sh
./scripts/build.sh linux-binary
```

该路径仅构建 `build/bin/octopus-linux-x86_64`，跳过前端与价格更新流程，适合只需要 Linux 服务器二进制时使用。

**Linux 后端烟测**

```bash
chmod +x ./scripts/smoke-linux-backend.sh
./scripts/smoke-linux-backend.sh
```

该脚本会构建临时二进制、启动本地 mock upstream，并验证最小主链路：
- `/`
- `/manifest.json`
- `/healthz`
- `/api/v1/user/login`
- `/api/v1/channel/create`
- `/api/v1/group/create`
- `/api/v1/apikey/create`
- `/v1/chat/completions`

**Docker Compose 运行态烟测**

```bash
chmod +x ./scripts/smoke-docker-compose.sh
./scripts/smoke-docker-compose.sh
```

该脚本会对仓库内的 `docker-compose.yml` 做一条低污染的运行态验收链。

**仓库验证工作流**

仓库中还提供了 `.github/workflows/validation.yaml`，用于在本地缺少 Linux 运行环境或 Docker 时继续覆盖主要验收链：
- `go build ./...`
- Go 测试
- 前端 `pnpm exec tsc --noEmit`
- Linux 后端烟测 `scripts/smoke-linux-backend.sh`
- Docker Compose 运行态烟测 `scripts/smoke-docker-compose.sh`

**前端手工验收清单**

复杂 UI 的最终验收口径仍是“运行态 smoke + 手工 checklist”。
仓库内固定清单位于 `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`，对外宣称最终 UI 验收完成前，应至少覆盖桌面端与 `375px` 宽度。

**备份 / 导入契约**

- `/api/v1/setting/export` 在未传 `include_secrets` 时默认导出可直接恢复的全量快照
- 只有显式 `include_secrets=false` 才导出脱敏快照
- 动态路由后台任务的真实口径是 `dynamic summary scan`，只产出摘要，不会持久化写回新阈值，也不会改写用户路由顺序

**开发模式**

```bash
cd web && pnpm install && NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:1088" pnpm run dev
## 新建终端，启动后端服务
go run main.go start
## 访问前端地址
http://localhost:3000
```

### 🔐 初始账户

首次启动后，Octopus 会创建初始管理员账户：

- **用户名**：默认是 `admin`，也可通过 `OCTOPUS_ADMIN_USERNAME` 指定
- **密码**：默认随机生成并打印在启动日志中，也可通过 `OCTOPUS_ADMIN_PASSWORD` 指定

如果没有显式提供 `OCTOPUS_ADMIN_PASSWORD`，请通过 `docker logs octopus` 或服务启动日志查看生成的初始密码。
使用生成密码或环境变量指定的启动密码首次登录后，Octopus 会阻止继续访问其余管理台功能，直到你完成一次管理员密码修改。
首次登录时用户名可选修改；如果不改，仍然保留 `admin`，但密码必须修改。

> ⚠️ **安全提示**：请在首次登录后立即修改初始密码。

### 📝 配置文件

配置文件默认位于 `data/config.json`，首次启动时自动生成。

**完整配置示例：**

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

**配置项说明：**

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.host` | 监听地址 | `0.0.0.0` |
| `server.port` | 服务端口 | `8080` |
| `server.static_dir` | 优先使用的本地静态资源目录；不可用时回退到内嵌静态资源 | `static/out` |
| `database.type` | 数据库类型 | `sqlite` |
| `database.path` | 数据库连接地址 | `data/data.db` |
| `log.level` | 日志级别 | `info` |

> **注意**：`config.json` 中的默认端口为 `8080`。如果通过 Docker（安装脚本或 docker-compose）部署，外部端口默认映射为 `1088`。

对于 `octopus healthcheck` CLI，如果配置里的监听地址是 `0.0.0.0` 或 `::` 这类仅适合服务端绑定的通配地址，命令会自动改用 `127.0.0.1` 作为探活目标。

**数据库配置：**

支持三种数据库：

| 类型 | `database.type` | `database.path` 格式 |
|------|-----------------|---------------------|
| SQLite | `sqlite` | `data/data.db` |
| MySQL | `mysql` | `user:password@tcp(host:port)/dbname` |
| PostgreSQL | `postgres` | `postgresql://user:password@host:port/dbname?sslmode=disable` |

**MySQL 配置示例：**

```json
{
  "database": {
    "type": "mysql",
    "path": "root:password@tcp(127.0.0.1:3306)/octopus"
  }
}
```

**PostgreSQL 配置示例：**

```json
{
  "database": {
    "type": "postgres",
    "path": "postgresql://user:password@localhost:5432/octopus?sslmode=disable"
  }
}
```

> 💡 **提示**：MySQL 和 PostgreSQL 需要先手动创建数据库，程序会自动创建表结构。

**环境变量：**

所有配置项均可通过环境变量覆盖，格式为 `OCTOPUS_` + 配置路径（用 `_` 连接）：

| 环境变量 | 对应配置项 |
|----------|-----------|
| `OCTOPUS_SERVER_PORT` | `server.port` |
| `OCTOPUS_SERVER_HOST` | `server.host` |
| `OCTOPUS_SERVER_STATIC_DIR` | `server.static_dir` |
| `OCTOPUS_DATABASE_TYPE` | `database.type` |
| `OCTOPUS_DATABASE_PATH` | `database.path` |
| `OCTOPUS_LOG_LEVEL` | `log.level` |
| `OCTOPUS_GITHUB_PAT` | 用于获取最新版本时的速率限制（可选） |
| `OCTOPUS_RELAY_MAX_SSE_EVENT_SIZE` | 最大 SSE 事件大小（可选） |

### OAuth 环境变量覆盖（可选）

Octopus 对 Copilot 登录内置了默认值，可直接使用。
Antigravity OAuth 需要先配置 `OCTOPUS_ANTIGRAVITY_CLIENT_ID` 和 `OCTOPUS_ANTIGRAVITY_CLIENT_SECRET`，然后才能启动网页授权流程。
如有需要，仍可通过环境变量覆盖各项 OAuth 配置：

| 环境变量 | 默认值 |
|----------|--------|
| `OCTOPUS_COPILOT_CLIENT_ID` | `151ef1b1b0345b2351ca` |
| `OCTOPUS_COPILOT_SCOPE` | `copilot` |
| `OCTOPUS_COPILOT_DEVICE_CODE_URL` | `https://github.com/login/device/code` |
| `OCTOPUS_COPILOT_ACCESS_TOKEN_URL` | `https://github.com/login/oauth/access_token` |
| `OCTOPUS_ANTIGRAVITY_CLIENT_ID` | *（必需，通过环境变量设置）* |
| `OCTOPUS_ANTIGRAVITY_CLIENT_SECRET` | *（必需，通过环境变量设置）* |
| `OCTOPUS_ANTIGRAVITY_AUTHORIZE_URL` | `https://accounts.google.com/o/oauth2/v2/auth` |
| `OCTOPUS_ANTIGRAVITY_TOKEN_URL` | `https://oauth2.googleapis.com/token` |
| `OCTOPUS_ANTIGRAVITY_SCOPE` | `https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile` |

说明：
- Antigravity 使用 Octopus 自己的回调地址（`/api/v1/channel/antigravity/oauth/callback`）和轮询流程。
- 同时兼容 `ANTIGRAVITY_*` / `COPILOT_*` 别名变量。


## ⚠️ 升级注意

### 升级到 v1.23.1

1. **Rust FFI Core**：v1.23 将 Rust FFI 核心（`rust/core/`）延伸至 relay 流式、负载均衡选择和流式聚合。从源码构建时请确保已安装 Rust 工具链。Release 预构建二进制已内嵌 Rust 库。
2. **数据库迁移**：v1.23 包含新的数据库迁移。升级前请备份数据目录，程序将在启动时自动执行迁移。
3. **Docker 更新**：如果当前以 Docker 方式运行，请使用 Web UI 版本信息卡中展示的更新命令：
   ```bash
   docker compose pull && docker compose up -d
   ```
   容器内已禁用一键自更新，避免错误地替换容器内二进制。
4. **API Key 限流**：升级后，现有 API Key 默认不限流。请前往**设置 → API Key** 配置 RPM / TPM / Daily 配额。
5. **上游网关**：上游探测层已增强多 endpoint 回退、凭证降级、Cookie 回退等能力。已有上游站点将在下次刷新时被重新探测。
6. **先备份**：重大升级前，请使用 `/api/v1/setting/export`（或设置页备份功能）导出全量快照。


## 📖 功能说明

### 📡 渠道管理

渠道是连接 LLM 供应商的基础配置单元。

**Base URL 说明：**

程序会根据渠道类型自动补全 API 路径，您只需填写基础 URL 即可：

| 渠道类型 | 自动补全路径 | 填写 URL | 完整请求地址示例 |
|----------|-------------|----------|-----------------|
| OpenAI Chat | `/chat/completions` | `https://api.openai.com/v1` | `https://api.openai.com/v1/chat/completions` |
| OpenAI Responses | `/responses` | `https://api.openai.com/v1` | `https://api.openai.com/v1/responses` |
| Anthropic | `/messages` | `https://api.anthropic.com/v1` | `https://api.anthropic.com/v1/messages` |
| Gemini | `/models/:model:generateContent` | `https://generativelanguage.googleapis.com/v1beta` | `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent` |

> 💡 **提示**：填写 Base URL 时无需包含具体的 API 端点路径，程序会自动处理。

---

### 📁 分组管理

分组用于将多个渠道聚合为一个统一的对外模型名称。

**核心概念：**

- **分组名称** 即程序对外暴露的模型名称
- 调用 API 时，将请求中的 `model` 参数设置为分组名称即可

**负载均衡模式：**

| 模式 | 说明 |
|------|------|
| 🔄 **轮询** | 每次请求依次切换到下一个渠道 |
| 🎲 **随机** | 每次请求随机选择一个可用渠道 |
| 🛡️ **故障转移** | 优先使用高优先级渠道，仅当其故障时才切换到低优先级渠道 |
| ⚖️ **加权分配** | 根据渠道设置的权重比例分配请求 |

> 💡 **示例**：创建分组名称为 `gpt-4o`，将多个供应商的 GPT-4o 渠道加入该分组，即可通过统一的 `model: gpt-4o` 访问所有渠道。

---

### 💰 价格管理

管理系统中的模型价格信息。

**数据来源：**

- 系统会定期从 [models.dev](https://github.com/sst/models.dev) 同步更新模型价格数据
- 当创建渠道时，若渠道包含的模型不在 models.dev 中，系统会自动在此页面创建该模型的价格信息
- 也支持手动创建 models.dev 中已存在的模型，用于自定义价格

**价格优先级：**

| 优先级 | 来源 | 说明 |
|:------:|------|------|
| 🥇 高 | 本页面 | 用户在价格管理页面设置的价格 |
| 🥈 低 | models.dev | 自动同步的默认价格 |

> 💡 **提示**：如需覆盖某个模型的默认价格，只需在价格管理页面为其设置自定义价格即可。

---

### ⚙️ 设置

系统全局配置项。

**统计保存周期（分钟）：**

由于程序涉及大量统计项目，若每次请求都直接写入数据库会影响读写性能。因此程序采用以下策略：

- 统计数据先保存在 **内存** 中
- 按设定的周期 **定期批量写入** 数据库

> ⚠️ **重要提示**：退出程序时，请使用正常的关闭方式（如 `Ctrl+C` 或发送 `SIGTERM` 信号），以确保内存中的统计数据能正确写入数据库。**请勿使用 `kill -9` 等强制终止方式**，否则可能导致统计数据丢失。

---

## 🔌 客户端接入

### OpenAI SDK

```python
from openai import OpenAI
import os

client = OpenAI(   
    base_url="http://127.0.0.1:1088/v1",
    api_key="sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg", 
)
completion = client.chat.completions.create(
    model="octopus-openai",  # 填写正确的分组名称
    messages = [
        {"role": "user", "content": "Hello"},
    ],
)
print(completion.choices[0].message.content)
```

### Claude Code

编辑 `~/.claude/settings.json`

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

编辑 `~/.codex/config.toml`

```toml
model = "octopus-codex" # 填写正确的分组名称

model_provider = "octopus"

[model_providers.octopus]
name = "octopus"
base_url = "http://127.0.0.1:1088/v1"
```

编辑 `~/.codex/auth.json`

```json
{
  "OPENAI_API_KEY": "sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg"
}
```


### CC Switch（一键导入 CLI 工具）

在 Web 界面点击 **API 文档** 按钮 → **CC Switch** tab，即可生成深链接，直接将 Octopus 供应商配置导入 CLI 工具。

支持工具：**Claude Code**、**Codex**、**Gemini CLI**

Claude Code 模式还支持为 Haiku / Sonnet / Opus 分别配置不同的模型映射。

---

## 🤝 致谢

- 🙏 [looplj/axonhub](https://github.com/looplj/axonhub) - 本项目的 LLM API 适配模块直接源自该仓库的实现
- 📊 [sst/models.dev](https://github.com/sst/models.dev) - AI 模型数据库，提供模型价格数据
