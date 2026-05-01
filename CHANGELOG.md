# CHANGELOG

本文件记录 [xiaoli0412/octopus-xiaoli-repo](https://github.com/xiaoli0412/octopus-xiaoli-repo) 相对于上游 [bestruirui/octopus](https://github.com/bestruirui/octopus) 的所有新增功能、修复与变更。

---

## [Unreleased] - 基于 main 分支

### ✨ 新增功能

#### GitHub Copilot 支持（OAuth Device Flow）
- `internal/server/handlers/copilot.go`：实现 GitHub Copilot OAuth Device Flow 完整流程，包括获取设备码、轮询授权状态
- 渠道类型 6（`OutboundTypeGithubCopilot`）：注册至 outbound 路由，使用 OpenAI Chat 格式 + Bearer token
- 在渠道表单（`Form.tsx`）中添加 GitHub Copilot 认证入口，显示验证码及状态提示

#### Antigravity（Google Gemini Code Assist OAuth）支持
- `internal/server/handlers/antigravity.go`：实现 Antigravity OAuth Web Flow，包括 Start / Poll 两个接口
- `internal/transformer/outbound/antigravity/messages.go`：全新 Antigravity outbound 适配器，完整实现 `cloudcode-pa.googleapis.com` 协议：
  - 自动调用 `POST /v1internal:loadCodeAssist` 获取 `cloudaicompanionProject`（managed project ID），并通过 `sync.Map` 缓存
  - 将标准 Gemini 请求包装为 `{project, model, user_prompt_id, request}` 格式
  - 支持流式（SSE）和非流式响应，均解包 `{response: ...}` 外层结构
  - 自动设置 `X-Goog-Api-Client` 和 `Client-Metadata` 必需 headers
  - Key 格式支持 `<token>` 或 `<token>|<projectId>` 两种形式
- `internal/helper/fetch.go`：新增 `fetchAntigravityModels`，通过 `loadCodeAssist` 获取项目 ID 后调用 `retrieveUserQuota` 获取可用模型列表
- `providers.json`：Antigravity base_url 更新为 `https://cloudcode-pa.googleapis.com`
- 前端默认 base_url 同步更新

#### Providers API 与渠道模型测试
- `internal/server/handlers/providers.go`：新增 `GET /api/v1/providers` 接口，从 GitHub 远程获取或 fallback 至内置 `providers.json` 供应商列表
- `providers.json`：内置主流供应商配置（OpenAI、Anthropic、Gemini、DeepSeek、OpenRouter、智谱、火山引擎等共 20+ 供应商）
- `internal/server/handlers/channel.go`：
  - 新增 `POST /api/v1/channel/test-models`：对已有渠道测试指定模型的连通性和响应正确性（支持 429 视为通过）
  - 新增 `POST /api/v1/channel/test-models-by-config`：根据临时配置（无需保存渠道）测试模型，用于渠道创建前验证
- 渠道表单集成模型测试 UI，可逐个或批量验证模型

#### 模型选择交互优化
- 渠道表单新增"从列表选择模型"对话框（`ModelSelectDialog`）：
  - 自动获取该渠道所有可用模型（`/fetch-models`）
  - 支持搜索过滤、批量勾选
  - 区分"自动获取"和"手动配置"两类模型，统一显示为 `bg-primary` 徽章
  - 每次打开对话框时基于 `allModels` 重建已选状态，避免残留
- 修复对话框内滚动条被全局 `scrollbar-width: none` 隐藏的问题，添加 `.dialog-model-scrollbar` CSS 重写

#### API 使用文档弹窗（DocModal）增强
- 新增 curl 示例代码生成（支持 OpenAI Chat / Responses / Anthropic 三种格式）
- 新增模型/分组下拉选择，可同时展示分组和"渠道:模型"格式
- 新增分组名合法性检查（包含空格或冒号时给出警告）
- 新增 **CC Switch** 导入 Tab：
  - 支持 `claude` / `codex` / `gemini` 三种 CLI 工具
  - `claude` 模式下支持配置 Haiku / Sonnet / Opus 分模型映射
  - 生成 `ccswitch://v1/import` 深链接并打开新窗口

#### 全局设置增强
- 新增 API Base URL 配置项（`setting_key: api_base_url`），用于生成文档中的 curl 示例地址
- 前端自动读取该设置并展示在 DocModal 中

#### Zen 渠道类型（类型 8）
- 注册 `OutboundTypeZen`，按模型名前缀动态路由到对应 outbound（claude-\* → Anthropic, gpt-\* → OpenAI Responses, gemini-\* → Gemini, 其他 → OpenAI Chat）
- 适用于连接 OpenCode Zen 等多协议聚合端点

### 🐛 Bug 修复

- **`fix: 模型测试 429 视为通过`**：`testChannelModels` 和 `testChannelModelsByConfig` 中将 HTTP 429（Rate Limited）视为渠道可用，`Passed=true` 并附带说明文字
- **`fix: relay.go model not supported 错误提示`**：将错误消息改为 `"model not allowed for this API key"`，更准确描述 API Key 权限限制问题
- **`fix: locale 文件路径`**：修正 locale 文件名（`zh_hans.json` → `zh-Hans.json`，`zh_hant.json` → `zh-Hant.json`）并更新 provider
- **`fix: Antigravity TLS SNI 错误`**：还原了错误的域名映射代码，将 base_url 正确指向 `cloudcode-pa.googleapis.com`
- **`fix: GLM/Zhipu 1210 参数有误`**：通过 debug 日志写文件定位根因，确认 `metadata` 字段（来自 Anthropic 协议透传的 `user_id`）导致 GLM OpenAI-compat 端点返回 1210 错误；正确解法为将渠道改用 GLM Anthropic 兼容接口（`/api/anthropic/v1`），而非修改请求参数
- **`fix: mobile navbar 分隔线方向`**：底部导航栏（mobile）的分隔线从水平线（`w-6 h-px`）改为垂直线（`w-px h-6`），在横排布局中视觉正确

### 🔧 配置与基础设施

- `CORS 中间件`：允许本地开发来源（`localhost:3000`、`127.0.0.1:3000` 等），方便前后端分离开发
- `providers.json`：更新 Antigravity base URL、新增 Zen 渠道配置；新增 `质谱 AI (Anthropic 兼容)` 和 `质谱 Coding Plan (Anthropic 兼容)` 两条入口（base_url 均为 `https://open.bigmodel.cn/api/anthropic/v1`，channel_type=2），方便直接使用 Anthropic 兼容接口
- `docker-compose.yml`：镜像来源更新为 `xiaoli0412/octopus-xiaoli-repo`
- `CLAUDE.md`：添加上游同步操作指引和完整项目文档
- `CI/CD`：`release.yaml` 新增 tag 推送触发条件（`tags: v*`），在 `main` 分支上推送 `vX.Y.Z` tag 即可触发完整 release 构建流程（编译 + GitHub Release + Docker 镜像推送）

### 🌐 国际化

- 新增 `doc.*` 翻译命名空间（API 文档、curl 示例、CC Switch 相关 UI）
- 修正 locale 文件命名规范（使用连字符而非下划线）
- 补充 `channel.models.*` 命名空间（模型选择对话框空状态文案）

---

## 上游版本记录（参考）

上游 `bestruirui/octopus` 最新版本为 `v0.9.25`，本分支基于上游 `dev` 分支的 `88eb9cc`（合并点）进行开发。

如需查看上游变更日志，请访问：https://github.com/bestruirui/octopus/releases
