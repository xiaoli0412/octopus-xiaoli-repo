# CHANGELOG

本文件记录 [xiaoli0412/octopus-xiaoli-repo](https://github.com/xiaoli0412/octopus-xiaoli-repo) 相对于上游 [bestruirui/octopus](https://github.com/bestruirui/octopus) 的所有新增功能、修复与变更。

---

## 1.18.5 - 2026-05-10

### ✨ 核心更新

- `AI 自动化` 继续完成从旧任务控制台到治理总控台的换代：前后端治理会话模型、预览与应用链路、策略方案与回滚点继续收口，主页面进一步压缩为更贴近运维面板的工作台布局。
- 首页统计区完成运维化重排：`趋势 / 排行榜 / 令牌明细` 统一切到更紧凑的面板式交互，补齐 `金额 / 次数 / 令牌` 切换、图表类型切换、按渠道/模型/分发密钥/供应商密钥的中文维度展示，并修复 `window.12h` 一类 locale key 直出问题。
- `自动化设置` 补齐真正可用的运行入口配置：支持本机默认地址开关、请求地址、请求密钥、请求模型、并发与分发策略联动保存，并修复“选择项保存为 UI 值而不是真实 API key”的错位问题。

### 🔧 稳定性与发布收口

- backup / import / rollback、AI governance、stats token breakdown 与相关 handler / op / no-browser 验证链继续扩展，形成这次 `1.18.5` 的完整后端与前端收口版本。
- 版本源、前端展示版本、Debian Docker 构建参数、安装脚本默认镜像、README / AGENTS / 手工验收文档中的示例 tag 已统一同步到 `1.18.5` / `v1.18.5`。

### Release Highlights (EN)

- `AI Automation` continues the replacement of the old task console with a governance-first control center. The governance session model, preview/apply flow, saved strategy profiles, rollback points, and the main operator layout were all tightened again for this patch line.
- The home stats area was rebuilt into a more operations-oriented surface: `Trend`, `Rank`, and `Token Breakdown` now expose cost/request/token switching, chart-mode switching, Chinese-only dimension labels, and fixed the leaked locale-key issue such as `window.12h` rendering directly in the UI.
- The automation settings dialog now persists usable runtime configuration for local-default endpoint mode, request base URL, request API key, request model, and dispatch/runtime policy controls. It also fixes the earlier mismatch where a UI selection value could be stored instead of the real API key.
- Release-facing version sources are now synchronized to `1.18.5` / `v1.18.5` across runtime constants, frontend display, Docker build defaults, installer defaults, and release documentation.

## 1.18 - 2026-05-09

### ✨ 核心更新

- `AI Automation` 正式切换到治理工作台口径：后端新增 `overview / sessions / apply-runs / strategy-profiles / expert-presets / learning/summary` 聚合接口，前端重建为单页治理会话工作台，保留 `manual / ai_profile` 非破坏式边界，并把设置页入口收敛为轻量治理摘要卡。
- 退役旧版 AI 自动化 v1 路径：历史 `/api/v1/ai/config`、`/tasks`、`/profiles`、`/prompt-templates` 等旧入口统一返回 `410 Gone`，明确要求切换到治理会话 endpoints，避免旧 UI 与旧脚本继续误用半废弃协议。
- Backup / rollback 契约继续补强：补齐 AI automation 相关快照内容、replace-prune 与 compatibility 细节、回滚 compare/preview 信息，以及更完整的后端测试覆盖，确保治理数据与备份链路一起进入正式发布面。

### 🔧 发布与验证

- 修复并对齐前端发布守门：`verify-ai-config-profile-summary.mjs` 已从旧 `config-source-logic.ts` 依赖迁移到当前治理版设置卡与 API 契约校验，`test:screenshot-no-browser` 再次恢复可发布状态。
- 本轮已通过的关键门槛包括 `go build ./...`、治理与 backup 相关 Go tests、`tsc --noEmit`、`pnpm --dir web run build:static`，以及完整 `pnpm --dir web run test:screenshot-no-browser`。
- 正式发布版本升级为 `1.18` / `v1.18`，`VERSION`、后端版本常量、前端展示版本、Debian Docker 构建默认版本、安装脚本默认镜像，以及 README / 手工验收 / 仓库说明中的示例 tag 已统一同步。

### Release Highlights (EN)

- `AI Automation` now ships as a governance workspace instead of the earlier task-console shape. The backend exposes `overview`, `sessions`, `apply-runs`, `strategy-profiles`, `expert-presets`, and `learning/summary`, while the frontend is rebuilt around governance sessions with the `manual / ai_profile` safety boundary preserved.
- Legacy AI automation v1 endpoints are now explicitly retired with `410 Gone` responses so old task/profile/template routes cannot silently drift against the new governance contract.
- Backup and rollback coverage was extended to include AI automation state, richer replace-prune and compatibility details, and stronger backend test coverage, so governance data is releasable together with the backup chain.
- The frontend release gate was realigned by moving `verify-ai-config-profile-summary.mjs` off the deleted `config-source-logic.ts` helper and onto the current governance settings card plus API contract. `test:screenshot-no-browser`, `build:static`, `tsc`, governance-related Go tests, and `go build ./...` all pass again in this release run.
- The formal release version is now `1.18` / `v1.18`, synchronized across runtime version constants, frontend display, Debian Docker build defaults, installer defaults, and release-facing documentation.

## 1.17.1 - 2026-05-07

### 🐛 发布后修复

- 废弃 Docker Hub 官方安装路径：`scripts/install.sh` 不再接受 `docker.io/xiaoli0412/octopus-xiaoli-repo`，安装入口统一收敛到 GHCR 官方镜像、显式提供的私有 / 代理镜像，或镜像拉取失败后的源码支撑 Docker 构建，避免失效安装方案继续误导用户。
- 进一步补齐 GHCR 失败时的 Docker 安装收口：安装脚本现在会优先走 `GHCR -> 源码 Docker 构建 -> 指定 Linux 二进制构建本地 Docker 镜像` 的三级回退链，把本次线上验证过的成功路径收口成正式安装入口。
- 修复 `screenshot-no-browser` 中 Backup 回滚预览契约漂移：补齐 rollback replace-prune 兼容性详情、结构化 guidance/signal 文案，以及 `backup-rollback-preview-warnings-list` 测试选择器，对齐 `Backup.tsx`、`backup-logic.ts`、no-browser verifier 与 setting mock。
- 修复 Debian Docker 构建阶段缺少 `scripts/build-web-static.mjs` 导致的 `MODULE_NOT_FOUND`，`Dockerfile.debian` 现已在 `web-builder` 阶段显式复制该脚本，保证 `pnpm run build:static` 可在 CI 和 compose smoke 中执行。
- 修复 API key 登录态链路：`/api/v1/apikey/login` 现在返回结构化认证状态，前端会在服务端校验成功后再持久化 API key session，并把 `enabled / expire_at / supported_models / auth_mode` 同步到 dashboard 与 auth store。
- 修复 AI automation 任务取消与失败收口：为运行中任务补充跟踪态判断、失败后强制取消入口、取消等待和 terminal status 保护，减少后台执行与取消状态错位。

### ✨ 完成度整合

- 将其他自动化已完成但未发布的 backup rollback 细节收口一并纳入：回滚预览现在会展示 replace-prune 清理对象、结构化 rollback guidance、compatibility details 与更明确的网络错误提示。
- API key dashboard 增加会话状态徽章和认证模式展示，四语 locale 同步补齐。
- OAuth override URL 校验补到 Copilot / Antigravity handler 与测试，避免错误或带凭据 URL 在服务端静默流入后续链路。

### Release Highlights (EN)

- Fixed the Backup rollback preview contract drift that broke `test:screenshot-no-browser`, including the rollback replace-prune diagnostics, structured guidance/signals, and the updated `backup-rollback-preview-warnings-list` selectors.
- Removed Docker Hub from the official install path. `scripts/install.sh` now only accepts the GHCR release image, an explicitly provided private or mirrored registry image, or the existing source-backed Docker build fallback when image pulls fail.
- Fixed Debian Docker builds by copying `scripts/build-web-static.mjs` into the web builder stage so `pnpm run build:static` no longer fails with `MODULE_NOT_FOUND`.
- Fixed the API key auth session flow so `/api/v1/apikey/login` returns structured auth status and the frontend only persists API key auth after server validation.
- Folded in already-finished backup rollback, dashboard, and OAuth URL validation improvements, and released them together as the `1.17.1` patch line.

## 1.17 - 2026-05-07

### ✨ 核心更新

- 备份 / 导入 / 回滚链路继续补强：设置页 Backup 工作台新增更完整的 scope 摘要、preview invalidation、route diff compare、风险提示与回滚元信息展示，导入与回滚流程更可解释。
- 后端设置、分组、渠道、日志与路由目标相关操作补齐更严格的参数校验和测试覆盖，包含 AI automation base URL 校验、query param 处理、import preview 与任务初始化链路收口。
- 前端与脚本验证同步扩展到 backup 相关 no-browser / browser smoke 守门，减少“页面有改动但验证链没有跟上”的漂移。

### 🔧 发布与兼容性

- 发布版本升级为 `1.17`，`VERSION`、后端版本常量、前端展示版本、Docker Debian 构建默认版本和安装脚本默认镜像统一同步。
- README 与手工验收清单同步更新到 `v1.17` 正式发布口径。

### Release Highlights (EN)

- Backup / import / rollback flows were expanded with clearer scope summaries, preview invalidation, route diff compare views, risk hints, and richer rollback metadata in the Settings backup workspace.
- Backend validation and test coverage were tightened across settings, groups, channels, logs, route targets, import preview, and AI automation base URL handling.
- Frontend verification scripts were extended around backup-related no-browser and browser smoke checks so release coverage stays aligned with the shipped UI changes.
- The formal release version is now `1.17`, with synchronized version sources across runtime constants, frontend display, Docker build defaults, installer defaults, and release-facing docs.

## 1.16.4 - 2026-05-03

### 🐛 安装与发布修复

- `scripts/install.sh` 现在默认使用 `1088` 作为开放端口，并在非交互场景下自动选择可用外部端口。
- 一键安装直接拉取 `v1.16.4` 正式版 Docker 镜像，不再依赖服务器本地编译。
- Debian Dockerfile 继续保留给 CI 和镜像发布流程使用，避免正式版构建口径和安装口径冲突。

### 🔧 版本与文档

- 发布版本升级为 `1.16.4`，`VERSION`、后端/前端版本常量和安装说明保持一致。
- README 与手工验收清单同步补充了新的安装和构建行为说明。

## 1.16(beta) - 2026-05-02

### ✨ AI 自动化中心重建

- `AI Automation` 按 `单页内二级视图 + 单 AI 主链优先 + 右侧状态栏 + 固定高度工作区` 重建，顶层路由仍保持 `ai`。
- 首屏主链收敛为 `任务类型 -> 输入 -> 执行模型/来源 -> 运行 -> 当前结果`，结果默认摘要视图，并保留 `对比 / 原始` 切换。
- `AI Profile` 继续主链化：结果区可一键启用新产出的 Profile，但默认不自动切换运行来源；`manual / ai_profile` 保护链继续保留。
- 多 AI 能力保留为二级“高级调度区”，支持 `split mode / dispatch mode / 并发数 / lane 模型覆盖 / lane 来源覆盖`，不再抢占首屏。
- 模板、工具、快照、历史、学习全部收敛到固定容器内滚动区域，移除大块说明卡、展开撑高和宣传式文案。

### 🔧 前端契约与验证

- `web/src/components/modules/ai-automation/` 已拆分为稳定模块：页面壳、主链、状态栏、结果面板、Profile 面板、高级调度、资产区、学习区、历史区和共享逻辑。
- 结果消费逻辑补强为优先读取 `result_payload`，并在仅返回 `result_json` 时安全回退解析，稳定支持 `摘要 / 对比 / 原始 / 工具执行 / Profile 产物 / 保护动作`。
- 四套语言 `zh-Hans / zh-Hant / ja / en` 已同步到新工作台结构，避免 key path 泄漏、英文 fallback 和混语。
- AI 自动化页面测试、设置页联动测试、学习焦点静态守门已按新结构重写，保持单 AI 主链、多 AI 调度、Profile 激活、学习区、历史区与脱敏 API key 覆盖。

### 📦 发布与版本

- 当前发布版本升级为 `1.16(beta)`，Git tag / Release 名称同步为 `v1.16-beta`。
- 后端默认版本、前端展示版本、仓库级 `VERSION` 版本源、Debian Docker 构建默认版本注入统一同步到 `1.16(beta)`。
- `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md` 与 `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md` 已改写为本轮 AI 自动化整页重建口径和验收项。

## 1.12.5(beta) - 2026-05-01

### 🐛 Bug 修复

- `Install flow`：新增 `scripts/install.sh`，安装前会探测默认外部端口 `8080` 是否被占用；若已占用，则提示用户改为新的外部端口，例如 `1008`。
- `Install flow`：保留服务默认监听端口 `8080`，仅在安装和 compose 暴露层处理外部端口冲突，避免与现有配置和运行时语义脱节。

### 🔧 发布与版本

- 当前发布版本升级为 `1.12.5(beta)`。
- 后端默认版本、前端版本显示、仓库级 `VERSION` 版本源同步到 `1.12.5(beta)`。
- README / README_zh 的 Docker 快速开始改为优先引导使用带端口探测的安装脚本，并补充非交互自定义端口示例。

## 1.12(beta) - 2026-05-01

### 🐛 Bug 修复

- `Responses stream fallback`：补齐 `response.output_text.done` 的 outbound consumer fallback，避免缺少 `output_text.delta` 时 assistant 文本静默丢失。
- `Responses stream fallback`：补齐 `message output_item.done` 的 message text fallback，并保持与已消费 `delta` 的去重语义一致。

### 🔧 发布与版本

- 新增仓库级 `VERSION` 版本源，当前发布版本固定为 `1.12(beta)`。
- 后端默认版本、前端版本展示、Windows/Linux 构建注入、Debian Docker 构建注入统一同步到 `1.12(beta)`。
- `release` workflow 现在区分展示版本与 Docker tag：GitHub Release 继续使用 `v1.12-beta` 一类 tag，应用显示保持 `1.12(beta)`，Docker tag 自动转成兼容格式。
- 设置页版本信息增加规范化比较与展示，避免 `1.12(beta)` 与 `v1.12-beta` 被误判为不同版本。

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
