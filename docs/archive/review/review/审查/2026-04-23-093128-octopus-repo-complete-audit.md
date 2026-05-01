# Octopus Repo Complete Audit

触发时间：`2026-04-23T09:31:28+08:00`

## 1. Findings

本轮未复现会直接阻断当前后端主流程的 `Critical` 级问题。当前工作区下，后端启动、鉴权、渠道创建、分组创建、API Key 创建和 `/v1/chat/completions` 最小主链都能真实跑通。当前最高风险集中在发布门禁、验证覆盖和文档口径漂移。

### High

1. 发布 workflow 的 tag 路径仍然只做构建与发布，不做测试或 smoke 验证，存在“未经同轮验证直接发版”的真实风险。
   - 证据：`.github/workflows/release.yaml:45-118` 只执行 `bash scripts/build.sh release`、上传 release 和构建推送镜像，没有 `go test`、前端验证或 smoke 步骤。
   - 补充证据：`scripts/build.sh:649-705` 的 `release` 分支只做环境检查、前端构建、价格更新、二进制构建和归档，不包含测试或运行态验证。
   - 对比：`.github/workflows/validation.yaml:46-61` 才是当前仓库里承载 `tsc`、backup/locale 验证、Go 测试和 Linux smoke 的验证入口，但 release workflow 没有在仓库可见配置里显式依赖它。
   - 判断：这不是实现空洞，而是发布门禁缺口。代码今天是可工作的，但 tag 发布路径没有把“已验证”变成强约束。

2. 当前 CI 仍未覆盖一大批已存在测试，以及最近新增的前端无浏览器验证脚本；最近改动最重的 UI 主线仍缺持续化守护。
   - 证据：`.github/workflows/validation.yaml:46-58` 当前只跑 `pnpm exec tsc --noEmit`、`build:static`、`test:locale-consistency`、`test:backup-logic`、`test:backup-component`，以及 `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`。
   - 未进入 workflow 的 Go 测试示例：`internal/client/http_test.go`、`internal/conf/config_test.go`、`internal/helper/fetch_test.go`、`internal/model/channel_test.go`、`internal/model/setting_test.go`、`internal/server/auth/auth_test.go`、`internal/transformer/model/embedding_test.go`、`internal/transformer/outbound/antigravity/messages_test.go`、`internal/transformer/outbound/copilot/chat_test.go`。
   - 未进入 workflow 的近期前端验证脚本示例：`scripts/verify-dynamic-routing-help.mjs`、`scripts/verify-channel-create-flow.mjs`、`scripts/verify-group-create-flow.mjs`、`scripts/verify-ccswitch-flow.mjs`、`scripts/verify-help-hint-accessible.mjs`、`scripts/verify-route-target-copy.mjs`、`scripts/verify-circuit-breaker-help.mjs`。
   - 本地现实：`verify-dynamic-routing-help`、`verify-channel-create-flow`、`verify-group-create-flow`、`verify-ccswitch-flow` 本轮均通过，说明仓库已经有验证资产，但 CI 没持续消费它们。
   - 判断：这会直接削弱最近大面积 UI/设置页改动的回归发现能力，属于真实质量门槛缺口。

### Medium

3. 动态路由需求/实施文档仍明显超前于 shipped 行为，文档里的多模式 AI 路由并不是当前可配置、可运行、可验证的交付物。
   - 证据：`docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:229-237` 仍要求 `shadow-ai`、`hybrid`、`metrics-only`、`strict-mechanism`、`incident-safe`；`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:235-243` 仍以 `hybrid mode`、`metrics-only`、`strict-mechanism` 作为发布前门槛。
   - 实现依据：`internal/model/setting.go:24-29,51-56` 当前只有 `dynamic_routing_health_enabled` 与五个 `race_*_budget` 设置；`internal/task/dynamic.go:39-115` 的后台任务实际只是每日摘要扫描；README 已把口径收口为 `dynamic summary scan`，见 `README.md:176-180` 与 `README_zh.md:165-169`。
   - 判断：这是“文档承诺超前于实现”，不是当前运行时崩溃，但会误导完成度判断和对外交付口径。

4. Windows 验证链已经部分修复，但仓库自带的 `verify-go-env.ps1` 仍不能像 `smoke-win-backend.ps1` 一样自我规避默认 `go-build` 目录权限问题，导致默认 Windows 验证入口不一致。
   - 证据：`scripts/verify-go-env.ps1:26-31,67-69` 只加载 `use-go-env.ps1` 后直接跑 `go build ./...`，没有设置 repo-local `GOCACHE/GOTMPDIR/TEMP/TMP`。
   - 对比：`scripts/smoke-win-backend.ps1:117-121` 会先把 `GOCACHE`、`GOTMPDIR`、`TEMP`、`TMP` 注入到仓库内可写目录，再进入 Go 构建与 smoke。
   - 本地复现：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild` 仍失败，错误是 `C:\Users\李昊桐\AppData\Local\go-build\... Access is denied`；而 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 已能完整通过。
   - 判断：这不是产品主流程故障，但它是仓库验证链自身的可移植性缺口，尤其在 Windows 主机上会持续制造假阻塞。

5. 协议转换核心路径可用，但边界能力仍有真实 TODO，本轮没有找到覆盖这些 TODO 分支的定向验证。
   - 证据：`internal/transformer/inbound/anthropic/messages.go:160-163` 对 `tool_result` 其他结果类型仍标记 `TODO: support other result types`；`internal/transformer/outbound/gemini/messages.go:313-322` 对 `json_schema` 仍保留 `TODO: Convert JSON schema to Gemini schema format if schema is provided`。
   - 已验证现实：repo-local cache 条件下 `go test ./... -count=1` 通过，说明主路径没坏；但这并不等于这些边角协议场景已经闭环。

### Low

6. 顶层 README / README_zh 仍有发布与安装占位符，仓库对外文档尚未收口到可直接执行状态。
   - 证据：`README.md:45-54,71-73,88` 与 `README_zh.md:45-54,70-72,87` 仍包含 `<CURRENT_IMAGE>`、`<CURRENT_REPOSITORY_URL>`、`<CURRENT_ONE_CLICK_INSTALL_SCRIPT>`、`<CURRENT_RELEASES_URL>` 以及 TODO 文本。

7. `web/package.json` 仍保留不可移植的本地 HTTPS 开发脚本。
   - 证据：`web/package.json:7-10` 中 `devs` 脚本硬编码 `/workplace/code/octopus/build/localhost-key.pem` 与 `/workplace/code/octopus/build/localhost.pem`。

8. `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 仍保留过期结论，特别是 backup manifest 真值判断已经与当前代码和测试不一致。
   - 证据：`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:252-253` 仍写 `manifest.contains_secrets` 当前为 `false`。
   - 代码与测试依据：`internal/op/backup.go:162` 已按 `dumpContainsSecrets(d)` 写回真实值；`internal/server/handlers/setting_import_test.go:44-66,148-151` 已验证默认导出 `ContainsSecrets=true`、显式脱敏导出 `ContainsSecrets=false`。

## 2. Completion Assessment

### 总体判断

当前仓库不是“骨架工程”了。后端主链、backup/import 主流程、dynamic summary scan API/UI、Windows smoke 路径，以及近期一批设置页与创建流程的无浏览器验证都是真接线、真可达。当前更像“核心实现基本可用，但发布门禁、验证覆盖和文档收口还没完全闭合”的状态。

### 完成度评估

| 主题 | 评估 | 判断依据 |
| --- | --- | --- |
| 后端启动、HTTP 服务、健康检查、静态资源回退 | 已完成 | `main.go -> cmd/start.go -> internal/server/server.go` 主链真实可达；`/healthz`、静态资源页面和 `octopus healthcheck` 都有实证。 |
| Relay 主路径、failover、race fallback、route-target 动态调优 | 已完成 | `internal/relay/relay.go` 已消费 `shouldEscalateToRace(...)` 与 `runRaceFallbackWindow(...)`；repo-local cache 下 `go test ./... -count=1` 通过，Windows smoke 也通过。 |
| 备份导出 / 导入 / 回滚后端 | 已完成 | `internal/server/handlers/setting.go` 与 `internal/op/backup.go` 已接线 export/import/rollback/preview；backup 脚本与相关 Go 测试能证明主流程真实存在。 |
| Dynamic routing shipped 能力 | 部分完成 | 当前已完成健康开关、race budget、每日 summary scan、统计 API 与设置页展示；不是文档里宣称的完整 AI 多模式体系。 |
| Dynamic routing AI / hybrid / incident-safe 模式 | 未完成 / 文档先行 | 需求和实施文档写了模式体系，但代码里没有对应设置面、状态机和运行时交付。 |
| 前端关键设置页与创建流程的 no-browser 回归资产 | 已完成主线、CI 未收口 | 本地脚本可证明当前实现存在，但 workflow 还没全量接入。 |
| 前端在当前 Windows 主机上的生产构建与 Vitest | 部分完成 / 环境受阻 | `tsc` 和多条单进程脚本通过，但 `next build`、Vitest 仍被宿主 `spawn EPERM` 阻断。 |
| 发布与对外文档收口 | 部分完成 | release workflow 工具链已固定，但发布路径未带验证门禁，README 仍有占位符。 |

### 已完成、部分完成、未完成、疑似空实现

- 已完成：后端启动、鉴权、分组/渠道/API Key 创建链路、relay 主流程、backup/import/rollback 主后端、dynamic summary scan API/UI、Windows smoke 主链。
- 部分完成：动态路由 shipped surface、前端验证主线、协议转换边界能力、Windows 本地验证入口的一致性、发布收口。
- 未完成：动态路由文档承诺中的 AI 多模式体系、README 发布占位符清理、release 路径上的显式验证门禁。
- 疑似空实现或“仅文档存在”：动态路由文档中的 `shadow-ai` / `hybrid` / `metrics-only` / `incident-safe` 目前仍更像规划术语，而不是 shipped 功能。

## 3. Verification Summary

### 已验证项

- 代码与历史上下文核对：`git status --short --branch`、`git branch --all --verbose --no-abbrev`、`git log --oneline --decorate -n 12`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、automation memory、最近审查报告、README/docs/workflow/关键入口代码。
- 后端最小主链：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 通过，已验证 `/healthz`、`/`、`/manifest.json`、登录、channel/group/apikey 创建、`/v1/chat/completions`。
- Go 源码可编译/测试：在 repo-local `GOCACHE/GOTMPDIR/TEMP/TMP` 条件下，`go build ./...` 与 `go test ./... -count=1` 均通过。
- 前端静态与单进程验证：
  - `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过。
  - `node .\scripts\verify-backup-logic.mjs` 通过。
  - `node .\scripts\verify-backup-component.cjs` 通过。
  - `node .\scripts\verify-locale-consistency.mjs` 通过。
  - `node .\scripts\verify-dynamic-routing-help.mjs` 通过。
  - `node .\scripts\verify-channel-create-flow.mjs` 通过。
  - `node .\scripts\verify-group-create-flow.mjs` 通过。
  - `node .\scripts\verify-ccswitch-flow.mjs` 通过。
  - `node .\scripts\sync-web-static.mjs` 通过。

### 失败项或受环境阻塞项

- 直接执行 `go build ./...`、`go test ./... -count=1`：失败，原因不是源码错误，而是宿主默认 `C:\Users\李昊桐\AppData\Local\go-build` 不可写；切换到 repo-local cache 后通过。
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild`：失败，错误同样是默认 `go-build` 目录权限被拒绝；脚本本身未像 smoke 脚本一样自注入 repo-local cache。
- `Set-Location web; node .\node_modules\next\dist\bin\next build`：源码编译和 TypeScript 阶段通过，但最终仍以 `spawn EPERM` 结束。
- `node --require .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts`：仍在 Vite/esbuild 启动阶段触发 `spawn EPERM`。
- `pnpm --dir web ...`：当前主机 `pnpm` 不在 PATH，无法直接执行。
- `corepack pnpm --dir web ...`：当前主机又被 `C:\Users\李昊桐\AppData\Local\node\corepack\lastKnownGood.json` 权限问题阻断。
- `docker --version` / `docker compose version`：当前主机 `docker` 不在 PATH。

### 未验证项

- 本机未运行 Linux 路径：`scripts/smoke-linux-backend.sh` 仅阅读，未在 Linux/WSL 主机上实跑。
- 本机未运行 Docker Compose 路径：因为 `docker` 缺失，无法验证 `scripts/smoke-docker-compose.sh`。
- 本机未拿到可通过的 `next build`/Vitest 结果：当前只能证明源码可编译到后期阶段，以及多条单进程验证脚本通过；不能把宿主 `spawn EPERM` 误记为源码回归。

## 4. Comparison Notes

### 当前工作区 vs `HEAD`

- 当前工作区非常脏。`git diff --stat HEAD` 显示 111 个已跟踪文件差异、`13697` 行新增、`2714` 行删除，另有大量未跟踪测试、脚本、迁移、文档和构建产物。
- 这意味着本轮结论针对的是“当前工作区”，不是单纯的 `HEAD` 提交状态。
- 与上一轮相比，backup 相关验证脚本漂移问题已经被当前工作区修复：`verify-backup-logic.mjs` 与 `verify-backup-component.cjs` 本轮均通过，不再是昨天的高优先级阻塞。

### 当前分支 vs 稳定基线

- 由于本机已有 `origin/dev` ref，本轮稳定基线使用 `origin/dev`。
- `git rev-list --left-right --count origin/dev...HEAD` 显示当前分支相对基线 `0` 个落后、`22` 个领先提交。
- `git diff --name-only origin/dev...HEAD` 表明当前分支相对基线已引入 provider 模板、Copilot/Antigravity、README/CHANGELOG 收口、DocModal、CC Switch、设置页与渠道表单扩展等一大批能力；当前未提交工作区又在这之上叠加了更多 UI、backup、relay、task 和测试变更。

### 代码实现 vs README / docs / 任务说明

- 一致：README/README_zh 已把动态路由真实口径收口为 `dynamic summary scan`，这与 `internal/task/dynamic.go`、`internal/server/handlers/stats.go` 和 `web/src/components/modules/setting/DynamicRouting.tsx` 一致。
- 不一致：动态路由需求/实施文档仍宣称 `shadow-ai` / `hybrid` / `metrics-only` / `incident-safe` 等模式，但代码没有对应设置面与运行时交付。
- 不一致：顶层 README 仍有安装/发布占位符；`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 里 backup manifest 真值判断已经过期。

### 代码实现 vs 测试 / 构建 / 验证覆盖

- 一致：后端主链、backup 主线、dynamic summary scan、近期一批 no-browser 前端验证都是真实现、真接线，不是伪功能。
- 不一致：validation workflow 仍明显窄于仓库现有测试与脚本资产，尤其没有吸纳最近新增的多条 UI 验证脚本。
- 不一致：release workflow 可直接打 tag 发版，但工作流本身不带任何测试/smoke 门禁。
- 环境受限而非源码已证伪：当前 Windows 主机上 `next build`、Vitest、`pnpm`、`corepack pnpm`、Docker 都受到宿主权限/安装状态影响；这会阻断本地验证，但不等于源代码必然错误。

## 5. Top Next Actions

1. 给发布路径补真正的验证门禁。
   - 最直接的方式是让 tag release 显式依赖通过的 validation job，或者在 `release.yaml` 内至少补上 `go test`、前端验证与 smoke 的最小链路；否则 tag 推送仍可能直接发布未验证产物。

2. 扩大 validation workflow 覆盖面，优先纳入最近新增的前端 no-browser 验证脚本和遗漏的 Go 测试包。
   - 至少应把 `verify-dynamic-routing-help`、`verify-channel-create-flow`、`verify-group-create-flow`、`verify-ccswitch-flow` 这类当前主线改动最密集的脚本纳入 CI，并重新评估是否把 `go test ./... -count=1` 作为主验证口径。

3. 明确动态路由的对外交付口径。
   - 要么把 `docs/DYNAMIC_ROUTING_REQUIREMENTS*.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN*.md` 下调到当前 shipped 的 `summary scan + health switch + race budget` 现实范围；要么继续补实现，把文档里承诺的模式真正接线并加验证。

### 需要优先处理的前三项

- `release.yaml` 缺少发版前验证门禁。
- `validation.yaml` 未覆盖大量现有测试和近期前端验证脚本。
- 动态路由 docs 与 shipped 行为严重不一致。

### 建议下一步动作

- 工程侧：补齐 release/validation 两条 workflow 的分工与门禁关系。
- 产品文档侧：清理动态路由 overclaim 和 README 占位符。
- Windows 可移植性侧：把 `verify-go-env.ps1` 对齐到 `smoke-win-backend.ps1` 的 repo-local cache 策略，并记录 `next build`/Vitest/corepack 的宿主权限前置条件。

---

## 中文摘要

1. 本次触发时间
   - `2026-04-23T09:31:28+08:00`

2. 做了哪些检查、运行了哪些命令
   - 检查了仓库结构、`git status`、分支基线、最近提交、automation memory、最近审查报告、README/docs/workflow、后端入口、relay、dynamic routing、backup/import、前端设置页接线与测试分布。
   - 实际运行了：`go build ./...`、`go test ./... -count=1`、`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\verify-backup-logic.mjs`、`node .\scripts\verify-backup-component.cjs`、`node .\scripts\verify-locale-consistency.mjs`、`node .\scripts\verify-dynamic-routing-help.mjs`、`node .\scripts\verify-channel-create-flow.mjs`、`node .\scripts\verify-group-create-flow.mjs`、`node .\scripts\verify-ccswitch-flow.mjs`、`node .\scripts\sync-web-static.mjs`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild`、`Set-Location web; node .\node_modules\next\dist\bin\next build`、`node --require .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts`、`docker --version`、`docker compose version`。

3. 修改了哪些文件
   - 新增审查报告：`docs/review/审查/2026-04-23-093128-octopus-repo-complete-audit.md`
   - 新增可视化报告：`docs/review/审查/2026-04-23-093128-octopus-repo-complete-audit.html`

4. 发现了什么问题
   - 当前没有复现后端主链路的致命断裂，后端最小主流程是真实可达的。
   - 最高风险已经从“功能未接线”转成“发布门禁不足 + CI 覆盖不够 + 动态路由文档超前”。
   - `verify-go-env.ps1` 仍未像 `smoke-win-backend.ps1` 那样自修复 Windows 默认 Go cache 权限问题。
   - 协议转换边界仍有真实 TODO，README 仍有占位符，`web/package.json` 仍有硬编码本地开发路径。
   - 本机前端生产构建与 Vitest 依旧受 `spawn EPERM` 阻断，`pnpm/corepack` 与 Docker 也受宿主环境限制。

5. 本次结果是成功、跳过还是失败
   - 成功。审计已完成，`md + html` 报告已落盘；失败的命令都已定位为具体仓库问题或具体宿主限制，而不是中途放弃。

6. 是否需要我手动介入
   - 需要。
   - 工程决策上，需要决定 release workflow 是否强制依赖 validation 结果，以及 validation workflow 要扩到什么范围。
   - 环境层面，如果你希望在这台 Windows 主机上拿到完整本地发布证明，还需要你手动处理 `pnpm/corepack` 写权限、`next/vitest spawn EPERM`、以及 Docker 缺失问题。
