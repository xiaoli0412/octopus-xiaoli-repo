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

## 开发指南

### 环境要求
- Go 1.23+
- Node.js 18+
- pnpm

### 开发模式

```bash
# 启动后端
go run main.go start

# 启动前端（另一终端）
cd web
NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:8080" pnpm run dev
```

### 构建

```bash
# 跨平台构建
./scripts/build.sh

# 输出：
# - build/bin/       # 各平台二进制文件
# - build/docker/    # Docker 专用二进制
# - build/archives/  # ZIP 压缩包
```

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
| [scripts/build.sh](scripts/build.sh) | 跨平台构建脚本 |

## Docker 部署

```bash
# 使用 docker-compose
docker compose up -d

# 或直接运行
docker run -d \
  -v /path/to/data:/app/data \
  -p 8080:8080 \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:latest
```

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

