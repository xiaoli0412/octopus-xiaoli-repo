# Octopus 项目文档

## 项目概述

**Octopus** 是一个面向个人的 LLM（大语言模型）API 聚合与负载均衡服务。它作为统一网关，提供以下核心功能：

- **多渠道聚合**：通过单一接口连接多个 LLM 提供商渠道（OpenAI、Anthropic、Gemini、火山引擎）
- **负载均衡**：使用多种策略在多个渠道和 API 密钥之间分发请求
- **协议转换**：在不同 API 格式之间无缝转换（OpenAI Chat/Responses、Anthropic Messages、Gemini、Embeddings）
- **智能选择**：从多个基础 URL 中自动选择延迟最低的端点
- **多密钥管理**：每个渠道支持多个 API 密钥，具备智能密钥轮换
- **统计分析**：全面追踪 token 使用量、成本和请求指标
- **自动同步**：自动从提供商同步模型列表和定价信息

## 技术栈

### 后端（Go）
- **Web 框架**：Gin
- **数据库 ORM**：GORM（支持 SQLite、MySQL、PostgreSQL）
- **认证**：JWT（golang-jwt/jwt）
- **配置管理**：Viper（支持环境变量，前缀 `OCTOPUS_`）
- **CLI**：Cobra
- **日志**：Uber Zap
- **Token 计数**：tiktoken-go
- **SSE**：tmaxmax/go-sse

### 前端（Next.js + React）
- **框架**：Next.js 16.0.7（App Router）
- **UI 组件**：Radix UI + Tailwind CSS v4
- **状态管理**：Zustand
- **数据获取**：TanStack Query（React Query）
- **动画**：Framer Motion
- **图表**：Recharts
- **国际化**：next-intl（支持英文、简体中文、繁体中文）
- **构建输出**：静态导出（嵌入 Go 二进制文件）

### 部署
- **Docker**：多平台支持（linux/amd64、linux/arm64、linux/arm/v7、linux/386）

## 项目结构

```
octopus/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令
│   ├── start.go           # 启动服务命令
│   └── ...
├── internal/              # 内部包
│   ├── conf/             # 配置管理
│   ├── server/           # HTTP 服务器
│   ├── relay/            # 请求转发核心逻辑
│   │   ├── balancer/    # 负载均衡策略
│   │   └── transformer/ # 协议转换器
│   ├── handler/          # API 处理器
│   ├── model/            # 数据库模型
│   └── op/               # 业务操作
├── web/                  # 前端（Next.js）
│   ├── src/
│   │   ├── app/         # 页面路由
│   │   ├── components/  # React 组件
│   │   ├── api/         # API 客户端
│   │   ├── lib/         # 工具函数
│   │   ├── provider/    # Context 提供者
│   │   └── route/       # 路由配置
│   └── out/             # 构建输出（静态文件）
├── static/              # 嵌入的静态资源
├── scripts/            # 构建脚本
├── main.go             # 程序入口
└── docker-compose.yml  # Docker 部署配置
```

## 核心架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         Octopus 架构                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐ │
│  │   Frontend   │─────▶│   Server     │─────▶│   Relay      │ │
│  │  (Next.js)   │      │   (Gin)      │      │   Engine     │ │
│  └──────────────┘      └──────────────┘      └──────────────┘ │
│                                │                    │           │
│                                │                    ▼           │
│                                │           ┌──────────────┐    │
│                                │           │  Transformer │    │
│                                │           │   (Adapters) │    │
│                                │           └──────────────┘    │
│                                │                    │           │
│                                ▼                    ▼           │
│                        ┌──────────────┐    ┌──────────────┐   │
│                        │  Handlers    │    │   Outbound   │   │
│                        │  (API)       │    │  Providers   │   │
│                        └──────────────┘    └──────────────┘   │
│                                │                    │           │
│                                ▼                    ▼           │
│                        ┌──────────────────────────────┐        │
│                        │          Operations (op)      │        │
│                        │  - Channel/Group Management  │        │
│                        │  - Statistics Tracking       │        │
│                        │  - Cache Management          │        │
│                        └──────────────────────────────┘        │
│                                │                                │
│                                ▼                                │
│                        ┌──────────────────────────────┐        │
│                        │     Database (GORM)          │        │
│                        │  - SQLite/MySQL/PostgreSQL   │        │
│                        └──────────────────────────────┘        │
│                                                                  │
│  后台任务：                                                      │
│  - 模型同步（channels → models）                                │
│  - 价格同步（models.dev → LLM pricing）                         │
│  - 延迟测量（base URL 健康检查）                                 │
│  - 统计持久化（memory → DB）                                    │
└─────────────────────────────────────────────────────────────────┘
```

## 数据库模型

### 核心表

| 表名 | 说明 |
|------|------|
| `User` | 管理员认证（用户名、密码哈希） |
| `Channel` | LLM 提供商配置（基础 URL、API 密钥、同步设置） |
| `ChannelKey` | 渠道的 API 密钥（状态、成本追踪、限流） |
| `Group` | 负载均衡组（聚合多个渠道） |
| `GroupItem` | 渠道到组的映射（优先级/权重） |
| `LLMInfo` | 模型定价信息 |
| `APIKey` | 客户端 API 密钥 |
| `Setting` | 系统配置键值存储 |
| `StatsTotal/StatsHourly/StatsDaily` | 时间序列统计 |
| `StatsModel/StatsChannel/StatsAPIKey` | 维度特定指标 |
| `RelayLog` | 请求/响应日志 |
| `MigrationRecord` | 数据库迁移版本追踪 |

## API 结构

### 管理 API（`/api/v1/`，需要 JWT）

- `/user/*` - 认证（登录、修改密码）
- `/channel/*` - 渠道 CRUD、启用/禁用、获取模型、同步
- `/group/*` - 组管理、项目操作
- `/model/*` - 模型定价管理
- `/apikey/*` - 客户端 API 密钥管理
- `/stats/*` - 统计查询
- `/log/*` - 请求日志
- `/setting/*` - 系统设置
- `/backup/*` - 备份/恢复

### 中继 API（`/v1/`，需要 API Key）

- `/chat/completions` - OpenAI Chat 格式
- `/responses` - OpenAI Responses 格式
- `/messages` - Anthropic Messages 格式
- `/embeddings` - Embedding API

## 负载均衡策略

| 策略 | 说明 |
|------|------|
| **Round Robin** | 按顺序轮询渠道 |
| **Random** | 随机选择 |
| **Failover** | 基于优先级的故障转移 |
| **Weighted** | 根据权重比例分配 |

## 高级功能

### 熔断器
- 连续失败阈值
- 指数退避冷却
- 半开放状态探测
- 按模型、按密钥追踪

### 会话亲和性
- 粘性会话保持渠道选择
- 可配置会话超时

### 协议转换
- **入站适配器**：OpenAI Chat/Responses、Anthropic、Embeddings
- **出站适配器**：OpenAI、Anthropic、Gemini、火山引擎
- 自动流式响应转换

### 后台任务
- 从渠道同步模型（自动发现）
- 从 models.dev 更新价格
- 基础 URL 延迟测量
- 统计持久化（批量写入）

## 本地资源与子 Agent 协作规则

后续执行任务时，应优先利用本地资源，而不是跳过上下文重新猜测。这里的本地资源至少包括：仓库内 MD 文档、当前阶段 worklog、现有脚本、测试与构建命令、当前线程里的上下文、已经沉淀的本地 skills，以及可复用的记忆上下文。

执行要求：

- 开工前先盘点本轮可直接复用的本地资源，并优先以这些资源建立任务上下文。
- 若本地 skills、已有工作记录或线程上下文已经覆盖相关结论，先继承、核对、补充，不重复发明流程。
- 遇到可以并行的独立子任务时，应积极使用子 agent，先完成任务拆分，再安排并行执行。
- 每个子 agent 在启动前都要明确职责边界，至少说清楚负责的问题范围、文件范围，或验证范围，避免多人改同一区域造成冲突。
- 若当前环境允许并可正常调度，子 agent 默认统一使用 `gpt-5.4`；若因权限、网关或环境原因无法使用该模型，主线程必须记录阻塞原因，再回退到本地资源优先的串行方案或经明确记录的替代模型。
- 主线程负责汇总各子 agent 的结论，互通有无，消除冲突，并把最终采用的结论写回主文档、worklog 或交付说明。
- 若子 agent 不可用、权限受限或模型不可用，主线程要记录阻塞原因，并改用本地资源或串行方式继续推进，不能直接丢失上下文。
## 开发指南

### 环境要求
- Go 1.24.4+
- Node.js 18+
- pnpm

### 开发模式

Windows 若当前 PowerShell 会话无法稳定找到 Go，可先启用仓库内工具链入口：

```powershell
# 为当前 PowerShell 会话解析 Go 工具链并补齐工作区缓存目录
. .\scripts\use-go-env.ps1
```

若当前 PowerShell 会话里的 Node / pnpm / corepack 在宿主机上不稳定，前端相关命令可先启用仓库内 Node 环境入口：

```powershell
# 为当前 PowerShell 会话固定可用的 Node 可执行文件，并将 corepack / pnpm 缓存收口到仓库内 .tmp-tooling/
. .\scripts\use-node-env.ps1
```

```bash
# 启动后端
go run main.go start

# 启动前端（另一终端）
cd web
NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:8080" pnpm run dev
```

也可以优先使用仓库内脚本化入口来校验依赖并启动开发环境：

```powershell
# Windows：检查依赖并分别启动前后端
powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1

# 仅校验命令与版本，不启动服务
powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1 -CheckOnly
```

```bash
# Linux / WSL：检查依赖并启动前后端
chmod +x ./scripts/dev-linux.sh
./scripts/dev-linux.sh

# 仅校验命令与版本，不启动服务
./scripts/dev-linux.sh --check-only
```

### 构建

```bash
# Linux 服务器二进制（轻量路径，仅构建主二进制）
./scripts/build.sh linux-binary

# Linux 服务器完整构建（含前端构建、价格更新、Docker 二进制准备）
./scripts/build.sh linux-server

# 本地 Docker 镜像构建
./scripts/build.sh docker-image

# 全量发布构建
./scripts/build.sh release

# 输出：
# - build/bin/       # 各平台二进制文件
# - build/docker/    # Docker 专用二进制
# - build/archives/  # ZIP 压缩包
```

```powershell
# Windows 构建（含前端构建与 static/out 同步）
powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1
```

### 验证与冒烟

```powershell
# Windows：检查 Go 环境，必要时补齐工作区缓存目录
powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1

# Windows：附加校验入口（格式、测试、构建或 Phase A 检查）
powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoFmt
powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest
powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild
powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -PhaseA

# Windows：本地后端最小冒烟（含 mock upstream、登录、建 channel/group/apikey、网关转发）
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1

# Windows：先做运行态预检，再看状态/停止/执行 healthcheck
powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only

# Windows：查看/停止本地运行中的前后端，或执行 healthcheck
powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status
powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop
powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action healthcheck

# Windows：前端 no-browser 统一验证入口与 TypeScript 校验
. .\scripts\use-node-env.ps1
pnpm --dir web run test
pnpm --dir web run test:settings-no-browser
pnpm --dir web run test:screenshot-no-browser
& $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json

# Windows：前端静态导出与 static/out 同步
pnpm --dir web run build:static

# Windows：browser smoke 静态守门与统一 screenshot 入口
node .\scripts\verify-browser-smoke-wrapper-alignment.mjs
node .\scripts\verify-ai-automation-learning-focus.mjs
node .\scripts\run-frontend-verification-suite.mjs screenshot
```

```bash
# Linux / WSL：本地后端最小冒烟
./scripts/smoke-linux-backend.sh

# Docker Compose 运行时冒烟
./scripts/smoke-docker-compose.sh

```

前端最终验收仍以“运行态 smoke + 人工 checklist”为准，可参考仓库内清单：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`

说明：历史状态、需求、审查与 worklog 文档已迁入 `docs/archive/`，当前主线 UI 任务文档位于 `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`。

### 环境变量

使用 `OCTOPUS_` 前缀设置配置：

```bash
export OCTOPUS_DB_TYPE=sqlite
export OCTOPUS_DB_CONNSTRING=data/octopus.db
export OCTOPUS_LOG_LEVEL=info
```

## 关键文件参考

| 文件 | 说明 |
|------|------|
| [main.go](main.go) | 应用入口 |
| [cmd/start.go](cmd/start.go) | 服务器初始化序列 |
| [internal/server/server.go](internal/server/server.go) | HTTP 服务器设置 |
| [internal/relay/relay.go](internal/relay/relay.go) | 请求路由核心逻辑 |
| [internal/relay/balancer/](internal/relay/balancer/) | 负载均衡实现 |
| [internal/model/](internal/model/) | 数据库模型定义 |
| [internal/transformer/](internal/transformer/) | API 协议适配器 |
| [web/src/app.tsx](web/src/components/app.tsx) | 主 React 应用 |
| [scripts/dev-win.ps1](scripts/dev-win.ps1) | Windows 本地开发启动脚本 |
| [scripts/dev-linux.sh](scripts/dev-linux.sh) | Linux / WSL 本地开发启动脚本 |
| [scripts/build-win.ps1](scripts/build-win.ps1) | Windows 构建脚本 |
| [scripts/build.sh](scripts/build.sh) | 跨平台构建脚本 |
| [scripts/use-go-env.ps1](scripts/use-go-env.ps1) | Windows PowerShell 会话级 Go 工具链入口 |
| [scripts/use-node-env.ps1](scripts/use-node-env.ps1) | Windows PowerShell 会话级 Node / pnpm / corepack 环境入口 |
| [scripts/verify-go-env.ps1](scripts/verify-go-env.ps1) | Windows Go 环境与工作区缓存校验 |
| [scripts/phase-a-check.ps1](scripts/phase-a-check.ps1) | Windows Phase A 聚合校验入口（`verify-go-env.ps1 -PhaseA` 调用） |
| [scripts/build-web-static.mjs](scripts/build-web-static.mjs) | 前端静态导出与 `web/out -> static/out` 同步入口 |
| [scripts/install.sh](scripts/install.sh) | Linux 服务器一键安装入口（优先拉取 GHCR 官方镜像，失败后回退源码 Docker 构建，并支持已知可用 Linux 二进制兜底） |
| [scripts/run-frontend-verification-suite.mjs](scripts/run-frontend-verification-suite.mjs) | 前端 no-browser 聚合验证入口 |
| [scripts/verify-browser-smoke-wrapper-alignment.mjs](scripts/verify-browser-smoke-wrapper-alignment.mjs) | browser smoke wrapper 静态守门 |
| [scripts/verify-ai-automation-learning-focus.mjs](scripts/verify-ai-automation-learning-focus.mjs) | AI 自动化 learning 断言守门 |
| [scripts/smoke-win-backend.ps1](scripts/smoke-win-backend.ps1) | Windows 本地后端冒烟脚本 |
| [scripts/smoke-linux-backend.sh](scripts/smoke-linux-backend.sh) | Linux / WSL 本地后端冒烟脚本 |
| [scripts/smoke-docker-compose.sh](scripts/smoke-docker-compose.sh) | Docker Compose 运行时冒烟脚本 |
| [scripts/runtime-win.ps1](scripts/runtime-win.ps1) | Windows 本地运行时状态/停止/healthcheck 工具 |
| [.github/workflows/validation.yaml](.github/workflows/validation.yaml) | 仓库 CI 验证链入口（前端/后端/运行态 smoke） |

## Docker 部署

```bash
# 使用 docker-compose
docker compose up -d

# 可选：覆盖默认外部端口/数据目录/容器名
OCTOPUS_PORT=1088 \
OCTOPUS_DATA_DIR=./data \
OCTOPUS_CONTAINER_NAME=octopus \
docker compose up -d

# 或直接运行
docker run -d \
  -v /path/to/data:/app/data \
  -p 1088:1088 \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.18.5

# Linux 服务器可直接使用仓库内安装脚本
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | bash

# 可选：通过环境变量覆盖安装脚本使用的外部端口
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | OCTOPUS_PORT=1088 bash

# 若 raw.githubusercontent.com 在目标主机不稳定，可改为两步执行以区分下载与拉镜像问题
curl -fsSL -o install-octopus.sh https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh
bash install-octopus.sh

# 若 GHCR 拉取失败且源码 Docker 构建仍受阻，可提供已知可用的 Linux 二进制作为兜底
OCTOPUS_BINARY_PATH=/root/octopus-linux-amd64 bash install-octopus.sh
```

说明：Docker Hub 作为 Octopus 官方安装来源已废弃。安装脚本当前只接受 GHCR 官方镜像或你显式提供的私有 / 代理镜像地址；镜像拉取失败时会先回退到 `main` 分支源码支撑 Docker 构建，必要时还可基于 `OCTOPUS_BINARY_PATH` 指向的已知可用 Linux 二进制继续构建本地 Docker 镜像。

## 贡献指南

请参考 [CONTRIBUTING_zh.md](CONTRIBUTING_zh.md) 了解如何贡献代码。

## 上游同步指南

本仓库（`xiaoli0412/octopus-xiaoli-repo`）是基于 `bestruirui/octopus` 演化出的独立仓库，包含额外的功能扩展。

### 初始配置（已完成）

```bash
git remote add upstream https://github.com/bestruirui/octopus.git
git fetch upstream
```

### 合并上游更新

```bash
# 拉取上游最新代码
git fetch upstream

# 查看上游有哪些 本分支尚未包含的提交
git log --oneline upstream/dev ^HEAD

# 切到当前开发分支后合并
git checkout main
git merge upstream/dev --no-ff -m "chore: merge upstream dev into main"

# 如果出现冲突，解决冲突后提交
# git add .
# git commit
```

### 智能合并策略

- **优先选策略**：对上游的 bugfix/feature 提交选择性 `cherry-pick` 而非整体 merge，可减少冲突
- **冲突优先级**：凡是 `internal/transformer/outbound/antigravity/`、`internal/server/handlers/antigravity.go` 等 erguotou 新增的文件，解决冲突时优先保留本分支版本
- **定期同步建议**：每当上游发布新 tag（`v*`）时同步一次

### 查看两仓库差异

```bash
git log upstream/dev ^HEAD --oneline        # 上游有、本地没有的提交
git log HEAD ^upstream/dev --oneline        # 本地有、上游没有的提交（您的扩展功能）
```

