# 2026-04-28 18:31:06 Octopus Repo Complete Audit

## 1. Findings

### High

1. AI Profile 的“结构化方案”主合同仍未真正闭环，当前运行时真实消费的大多仍只是执行配置字段，而不是各域可复用的运营方案。
   文件/模块：`docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:20-23,61-69,77-97`，`internal/op/ai_automation_executor.go:408-450`，`internal/op/ai_automation_profile_typed.go:201-255,323-379`，`internal/op/ai_automation.go:748-889`，`web/src/components/modules/ai-automation/index.tsx:2447-2457`。
   判断依据：需求文档要求 AI 生成的分组、渠道识别、价格识别、模型归类、配置健康检查都应保存为可预览、可校验、可后续消费的结构化 AI Profile；但执行器当前生成的 profile 主体仍主要是 `summary`、`raw_output`、`config`、`runtime`、`tool_execution*` 等通用包装字段。typed payload 虽然存在，但它只是从这些通用字段做回填；而当前执行器并没有为 `grouping_suggestions`、`price_items`、`classification`、`issues` 等域字段产出稳定内容。真正被运行时读取的仍只有 `base_url / api_key / channel_type / model / use_local_default`。前端 Profile 预览也只是直接展示 `domain_payload` 或 legacy `content_json` 的 JSON 快照。
   风险：README / docs / UI 容易让人误以为“AI 方案中心”已经具备可复用的结构化运营方案闭环，但当前更接近“AI 结果审计快照 + 执行配置切换器”。

### Medium

2. “多 AI / planner-executor-reviewer / router mesh” 当前本质上是前端批量发起多个普通任务，不是真正的后端 orchestrator。
   文件/模块：`web/src/components/modules/ai-automation/index.tsx:1423-1495`，`internal/server/handlers/ai_automation.go:18-30,106-117`，`internal/op/ai_automation.go:289-341`，`internal/op/ai_automation_executor.go:141-239`。
   判断依据：前端提供了 `single / planner_executor / planner_executor_reviewer / router_mesh` 等 split mode 和 lane 策略；但真正执行时只是按 lane 组装多个 `POST /api/v1/ai/tasks` 请求，顺序或并行地提交多个普通任务。后端没有 lane graph、依赖关系、聚合器、共享上下文状态机或“reviewer 审核后再推进 executor”的服务端编排模型，仍是单任务执行器。
   风险：UI 语义明显强于后端真实能力，用户可能误以为存在服务端协调、互审或守护链路，实际只是前端 fan-out + 列表展示。

3. 备份/导入主链是真实现，但“高级迁移能力”仍显式处于部分完成状态，不能算 fully done。
   文件/模块：`docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:183-184`，`web/src/components/modules/setting/Backup.tsx:271-272,301,915-932`。
   判断依据：需求文档明确写了“Backup advanced migration remains partially complete”；前端也仍直接对用户展示 `Advanced migration tooling still pending` 和“仍需手动处理的迁移能力”区域。这说明主链可用，但高级迁移工具、回滚域细化和剩余手工迁移项还没有完全收口。
   风险：如果状态文档或交付说明把备份主线描述成“完全完成”，会高估当前迁移能力边界。

### Low

4. `config_health` 与 `config_health_check` 的命名在 docs 与代码之间仍有轻微漂移。
   文件/模块：`docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:77-83,179`，`internal/model/ai_automation.go:14,56`，`web/src/components/modules/ai-automation/index.tsx:257,385,687`。
   判断依据：需求文档的 Profile 类型列表仍写 `config_health`，但代码和前端任务类型统一使用 `config_health_check`，并只在需求补充说明里声明前者为 legacy alias。
   风险：不会直接破坏运行时，但会持续增加检索、维护和测试命名噪音。

5. 仓库忽略规则过窄，当前工作区混入了大量宿主生成产物，显著放大了审计噪音和误提交风险。
   文件/模块：`.gitignore:1-5`，`git status --short --branch` 当前输出。
   判断依据：`.gitignore` 当前只忽略 `data`、`build`、`static/out/*` 和 `.vscode`，没有覆盖 `.next`、`node_modules`、`.gocache`、`.gomodcache`、`.tmp*`、`%SystemDrive%` 等本仓常见本地产物；这些内容已在当前工作区以 untracked 形式大量出现。
   风险：真实业务改动和宿主临时产物混在一起，会持续干扰 code review、审计、PR diff 和误提交控制。

## 2. Completion Assessment

### 完成度评估

- 已完成：服务启动主链、静态资源构建同步、AI 自动化基础配置/API、AI task 基本执行链、AI Profile 保存/激活/回退、动态路由学习查询与重置入口、备份导出/导入/回滚主链、`healthcheck` CLI、CI 中的 `tsc / build:static / vitest / no-browser / go build / targeted go test / Linux smoke / docker smoke` 门禁定义。
- 部分完成：AI Profile 结构化消费闭环、多 lane AI 编排语义、备份高级迁移工具、全宿主可复现的浏览器/Vitest/Linux/Docker 证明链。
- 未完成：把 `grouping / channel_recognition / price_recognition / model_classification / config_health_check` 这些域真正接到 downstream consumer/validator；将 multi-lane 语义提升为真实后端 orchestrator，或明确收缩为 batch launcher。
- 疑似空实现/语义强于实现：没有发现完全不可达的核心主流程空壳；但 multi-lane 控制台属于“展示语义明显强于后端能力”，AI Profile 各域 typed payload 也仍偏“回填式结构化”。

### 主流程可达性判断

- `cmd/start.go:27-50` 会按 `conf.Load -> db.InitDB -> op.InitCache -> server.Start -> task.Init` 真实拉起主流程，不是空壳启动。
- `internal/task/init.go:23-69` 真实注册了价格更新、基础 URL 延迟、动态路由摘要扫描、模型同步、统计保存与日志落库任务。
- `op.InitCache()` 阶段也会接入 AI task 恢复链，而不是只创建表结构不消费。

## 3. Verification Summary

### 已验证项

- `git status --short --branch`
- `git log --oneline --decorate -n 12`
- `git branch -a --verbose --no-abbrev`
- `git rev-list --left-right --count origin/dev...HEAD` -> `0 22`
- `git diff --shortstat HEAD` -> `137 files changed, 19537 insertions(+), 4008 deletions(-)`
- `. .\scripts\use-go-env.ps1; go build ./...` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./cmd -count=1` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./internal/op -count=1` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./internal/task -count=1` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./internal/update -count=1` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./internal/server/handlers -run 'Test(AIAutomationConfigHandlers|AIAutomationConfigHandlerReturnsExplicitFallbackSemantics|AITaskHandlerRejectsWhenDisabled|CancelAITaskHandlerMarksTaskCanceled|DynamicRouteLearningHandlersListAndReset|AIProfileActivateHandlerSwitchesSettingsOnly)$' -count=1` -> 通过
- `. .\scripts\use-go-env.ps1; go test ./internal/server/middleware -count=1` -> 通过
- `. .\scripts\use-node-env.ps1; node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` -> 通过
- `. .\scripts\use-node-env.ps1; node .\scripts\build-web-static.mjs` -> 通过
- `. .\scripts\use-node-env.ps1; node ...` 批量运行 no-browser/logic 校验 -> 全部通过：`verify-locale-consistency`、`verify-home-layout`、`verify-channel-create-flow`、`verify-channel-presentation`、`verify-group-create-flow`、`verify-llm-price-boundary`、`verify-backup-logic`、`verify-backup-component`、`verify-help-hint-accessible`、`verify-dynamic-routing-help`、`verify-circuit-breaker-help`、`verify-model-probe-help`、`verify-setting-info-logic`、`verify-ai-config-profile-summary`、`verify-ai-automation-learning-focus`、`verify-ccswitch-flow`、`verify-route-target-copy`。

### 未验证项

- `. .\scripts\use-go-env.ps1; go test ./internal/server/handlers -count=1`：当前宿主的 `net.Listen(tcp4)` / WinSock service-provider 初始化失败会阻断依赖本地监听 socket 的测试用例，因此不能把这组失败直接定性为仓库回归。
- `. .\scripts\use-go-env.ps1; go test ./internal/relay/... -count=1`：同样受本机 `listen tcp4 127.0.0.1:0` service-provider 初始化失败影响。
- `. .\scripts\use-node-env.ps1; web\node_modules\.bin\vitest.cmd run`：当前宿主在 `vite/esbuild` 加载 `web/vitest.config.ts` 阶段报 `spawn EPERM`，无法在本机复现 CI 里的 Vitest 结果。
- `bash scripts/smoke-linux-backend.sh`：当前主机 `wsl -l -v` 返回 `Wsl/EnumerateDistros/Service/E_ACCESSDENIED`，无法本机执行 Linux smoke。
- `bash scripts/smoke-docker-compose.sh` / `docker compose`：当前主机 `docker` 不可用，无法本机执行 Docker runtime smoke。

## 4. Comparison Notes

### 当前工作区 vs `HEAD`

- 当前工作区相对 `HEAD` 仍是大体量未提交改动：`137 files changed, 19537 insertions(+), 4008 deletions(-)`。
- 关键变化集中在 `internal/op`、`internal/relay`、`internal/server/handlers`、`web/src/components/modules/*`、CI workflow、README 和大量新增测试/脚本。
- 工作区里还混入了大量未忽略的宿主产物，导致 `git status` 噪音显著偏大。

### 当前分支 vs 最近稳定基线

- 当前位于 `feat/erguotou`，`HEAD` 为 `bfa27ae`。
- 相对 `origin/dev` 为 `0 behind / 22 ahead`，说明当前分支在上游 `dev` 之上已有较大功能扩展面。

### 代码实现 vs README / docs / 任务说明

- README、`validation.yaml`、`release.yaml`、`build-web-static.mjs` 与当前可通过的 `tsc/build:static/no-browser` 链路基本一致，较前几轮已明显收口。
- 最大一致性缺口仍集中在 AI Profile 语义：文档把它描述成结构化 AI 方案载体，但当前代码真实主消费主要落在执行配置字段与审计快照。
- 备份文案反而比较诚实：文档和 UI 都明确承认高级迁移能力仍在补齐，不应对外宣称 fully done。

### 代码实现 vs 测试 / 构建 / 验证覆盖

- 当前 CI workflow 已把 `pnpm run test`、`pnpm run test:screenshot-no-browser`、`go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1`、Linux smoke 和 Docker smoke 都列为正式 gate。
- 但在当前 Windows 宿主上，full `handlers/relay` 网络型 Go 测试、Vitest、Linux smoke、Docker smoke 仍无法本机复现，因此本轮“已验证项”更偏向 targeted tests + no-browser + build chain，而不是全矩阵闭环。

## 5. Top Next Actions

### 需要优先处理的前三项

1. 决定 AI Profile 到底是“执行配置切换器 + 审计快照”还是“真正可消费的结构化方案”。
   若选择后者，应为 `grouping / channel_recognition / price_recognition / model_classification / config_health_check` 至少各接上一条真实 downstream consumer 或 validator；若短期做不到，应同步收缩 docs / UI 承诺。

2. 对多 lane AI 控制台做一次语义决策。
   要么把 UI/文档明确收缩为“批量任务启动器”，要么补一个后端 orchestrator：lane 依赖、聚合结果、review/guard 真正拦截执行、共享上下文和服务端审计模型。

3. 继续把备份高级迁移保持在“部分完成”口径，并用健康宿主补齐剩余证明链。
   具体包括：Linux smoke、Docker compose smoke、Vitest、full `handlers/relay` 网络测试，以及浏览器级证据；同时补宽 `.gitignore`，把宿主产物从工作区里剥离出去。

### 建议下一步动作

- 先做架构/产品决策：AI Profile 与 multi-lane 两条语义主线不能长期靠 UI 话术兜着。
- 再做一次验证宿主收口：优先在 CI、可用 Docker/WSL 的 Linux/Windows 环境或更宽松的本机上复跑 full matrix。
- 完成后再更新状态文档，避免把“主链可用但范围未收口”的部分提前写成 fully done。

## 本次中文摘要

1. 本次触发时间：2026-04-28 18:31:06 +08:00。
2. 做了哪些检查、运行了哪些命令：读取了 automation memory、`README`、`validation.yaml`、`release.yaml`、`CURRENT_STATUS_AND_PLAN`、`FRONTEND_UI_MAINLINE_STATUS`、AI 自动化需求文档、核心 Go/前端源码；运行了 `git status --short --branch`、`git log --oneline --decorate -n 12`、`git branch -a --verbose --no-abbrev`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`go build ./...`、`go test ./cmd`、`go test ./internal/op`、`go test ./internal/task`、`go test ./internal/update`、定向 `internal/server/handlers` / `internal/server/middleware` 测试、`tsc --noEmit`、`build-web-static.mjs` 以及整组 repo-local no-browser / logic 校验脚本；同时确认了 `docker` 不可用、`wsl -l -v` 为 `E_ACCESSDENIED`、Vitest 为 `esbuild spawn EPERM`、full `handlers/relay` 为 WinSock provider 阻塞。
3. 修改了哪些文件：
   - `docs/review/审查/2026-04-28-183106-octopus-repo-complete-audit.md`
   - `docs/review/审查/2026-04-28-183106-octopus-repo-complete-audit.html`
   - `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`
4. 发现了什么问题：本轮最重要的问题是 AI Profile 仍缺少真正的结构化 consumer/validator 主闭环；多 lane AI 控制台当前只是前端 fan-out 普通任务；备份高级迁移仍显式 pending；`config_health` 命名有轻微漂移；`.gitignore` 过窄导致大量宿主产物混入工作区。
5. 本次结果是成功、跳过还是失败：成功；但有明确的环境阻塞项，导致 full `handlers/relay`、Vitest、Linux smoke、Docker smoke 只能记为未验证而不是失败。
6. 是否需要我手动介入：需要。优先是两类介入：
   - 架构/产品决策：AI Profile 是继续做真正结构化方案，还是收缩为配置切换器；multi-lane 是实现后端 orchestrator，还是收缩 UI 语义。
   - 环境支持：如需补齐全量证明链，需要可运行 Docker / WSL / Vitest 的宿主或直接参考 CI 结果。
