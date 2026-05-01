# Octopus 完整审计报告

- 本次触发时间：2026-04-29 12:55:34 +08:00
- 仓库路径：`D:\GPT-codex\octopus_repo`
- 审计分支：`feat/erguotou`
- 审计提交：`bfa27aecbec14329a43550c0d5409aefea257af0`
- 对比基线：`origin/dev`

## 1. Findings

### Critical

本轮未确认当前工作区仍成立的 `Critical` 级问题。

判断依据：
- 在把 Go 缓存重定向到仓库可写目录后，`go build ./...`、`go test ./... -count=1`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`、`node scripts/build-web-static.mjs`、以及整组 repo-local no-browser 校验脚本均通过。
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 也通过，说明当前工作区的最小后端主路径真实可达。

### High

1. `Antigravity` 被 README 和前端文案描述为“内建默认 OAuth / 开箱即用”，但当前后端实现并没有内建 `client_id/client_secret`，默认路径会直接失败。

判断依据：
- `README.md:285-298` 明确写着 `Octopus includes built-in defaults for Copilot and Antigravity OAuth login, so it works out-of-the-box.`，但同一张表又把 `OCTOPUS_ANTIGRAVITY_CLIENT_ID` 和 `OCTOPUS_ANTIGRAVITY_CLIENT_SECRET` 标成必需。
- `README_zh.md:276-289` 也写着“不配置也可直接使用”，同时又把同两项标为“必需”。
- 前端文案 `web/public/locale/en.json:790-797` / `web/public/locale/zh-Hans.json:790-797` 继续宣称 “Uses built-in default OAuth app (Gemini CLI)” / “使用内置默认 OAuth 应用（Gemini CLI）”。
- 但后端 `internal/server/handlers/antigravity.go:107-137` 只会从 `OCTOPUS_ANTIGRAVITY_CLIENT_ID` / `ANTIGRAVITY_CLIENT_ID` 与对应 secret 环境变量读取值，并没有任何默认 `client_id/client_secret`。
- `internal/server/handlers/antigravity.go:166-170` 在缺少 `client_id` 时直接返回 `400`；`internal/server/handlers/antigravity.go:280-286` 在 callback 流程里也会因为缺少 `client_id/client_secret` 把 session 标记为 failed。

影响：
- 这是一个真实用户路径失配，而不是纯文案问题。按当前实现，新部署用户若按 README / UI 理解去直接发起 Antigravity OAuth，默认不会成功。
- 它不会打断 Octopus 核心网关主线，但会让当前宣传的 Antigravity OAuth 主路径默认不可用。

### Medium

2. AI Profile 的 rich typed payload 已真实生成、存储并预览，但运行时主消费仍基本只抽执行配置字段，rich domain payload 仍处于“可审阅资料”而非“真实下游契约”的状态。

判断依据：
- `internal/op/ai_automation_executor.go:384-442` 与 `:1345-1369` 已为不同 domain 生成 `domain_payload`，包括 `grouping_suggestions`、`candidate_channel_model_mappings`、`price_items`、`classification`、`issues`、`suggested_actions` 等结构。
- `internal/op/ai_automation_profile_typed.go:77-126` 会把这些 typed payload 落到分域表；`internal/op/ai_automation_profile_typed.go:347-403` 也能从 typed payload 中抽配置。
- 但真正影响运行时配置来源切换的主路径仍是 `internal/op/ai_automation.go:844-893`；它最终只消费 `base_url`、`api_key`、`channel_type`、`model`、`use_local_default`。
- 前端 `web/src/components/modules/ai-automation/index.tsx:229-310` 主要做 `domain_payload` 或 legacy `content_json` 的结构化预览与摘要，不存在与 `grouping_suggestions / price_items / issues` 等 rich payload 对应的真实后续应用链。
- 需求文档 `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:73-97` 把 AI Profile 定义成结构化、可预览、可校验、可后续消费的结果单元；当前版本至少在“下游真实消费”这一层仍是部分完成。

影响：
- 这不是空实现。任务、历史、typed payload、预览、激活、数据库持久化都是真的。
- 但“AI 生成的分组/价格/分类/健康检查结果已形成后续业务闭环”这一说法在当前代码下仍然偏强，更准确的状态是：结构化产物已落地，真实下游消费仍有限。

3. 协议转换的高级边界仍保留真实 TODO 分支，核心路径虽已通过测试，但音频、prediction、JSON schema、复杂 `tool_result` 和 image-generation 边界能力仍缺直接证明。

判断依据：
- `internal/transformer/model/model.go:142-143` 仍保留 audio TODO，`:208-209` 保留 prediction TODO，`:618-619` 保留 `JSONSchema` TODO，`:861-866` 保留 image-generation tool request body TODO。
- `internal/transformer/inbound/anthropic/messages.go:160-175` 在 `tool_result` 分支仍写着 `TODO: support other result types`。
- `internal/transformer/outbound/gemini/messages.go:336-345` 在 `json_schema` 分支仍写着 `TODO: Convert JSON schema to Gemini schema format if schema is provided`。
- 本轮虽然 `go test ./internal/transformer/...` 与 `go test ./... -count=1` 全绿，但重新搜索 `internal/transformer/**/*test.go` 后，没有找到直接锚定 `json_schema`、复杂 `tool_result`、prediction、audio 这些 TODO 分支的专门回归证据。

影响：
- 当前主流 OpenAI / Anthropic / Gemini / Copilot / Antigravity 核心路径没有被这些 TODO 直接打断。
- 但“多协议无缝转换”在高级边角场景上仍更接近“主链可用、边界部分完成”，发布说明不宜把这些未收口边界描述成完全闭环。

### Low

4. `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 仍保留已过时的“当前环境无法直接复跑”结论，会误导接手人低估当前工作区的可验证程度。

判断依据：
- 文档 `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:448-457` 仍写着“当前环境下也无法直接复跑：本机 `go` 不在 PATH、本机 `pnpm` 未正确装好链接”，并据此得出“缺一次系统验收收口”。
- 但本轮已实际通过 `scripts/use-go-env.ps1`、`scripts/use-node-env.ps1` 与 repo-local Node/Go 入口完成 `go build ./...`、`go test ./... -count=1`、`tsc --noEmit`、`build-web-static.mjs`、`smoke-win-backend.ps1` 和整组 no-browser 校验。
- 这说明“机器裸 PATH 不完整”仍是事实，但“因此当前无法直接复跑”已经不再准确。

影响：
- 这是文档一致性问题，不会直接破坏运行时。
- 但它会误导后续审计、交接和优先级判断，让接手人继续沿用已经失效的环境结论。

5. 工作区噪音仍然很高，`.gitignore` 过窄，已经显著放大对比和审计成本。

判断依据：
- 根目录 `.gitignore:1-5` 当前只有 5 行：`data`、`build`、`static/out/*`、`!static/out/README.md`、`.vscode`。
- 本轮 `git status --porcelain=v1` 共有 `425` 行，其中 `287` 行是未跟踪文件，包含大量 `.tmp*`、cache、patch、日志、`.next`、`node_modules`、`.tools` 等本地产物。
- 当前工作区相对 `HEAD` 的摘要仍是 `138` 个已跟踪变更项外加 `287` 个未跟踪项，这让“真实业务差异”和“环境产物”很难快速分离。

影响：
- 这不是运行时缺陷，但会持续制造审计噪音，提高误提交本地产物、误判回归和遗漏真实差异的概率。

## 2. Completion Assessment

总体判断：

- 如果按 Octopus 核心网关主线看，当前仓库完成度依然较高，约 `90%-92%`：启动链、管理 API、中继 API、通道/分组/日志/统计、协议主路径、构建链、静态导出链，以及 Windows 本地最小后端冒烟都已真实可达并通过本轮复核。
- 如果按最近这批扩展能力看，完成度更适合表述为 `78%-82%`：Copilot / Antigravity / providers 模板库 / AI 自动化 typed payload / 备份增强 / 验证工作流都已具备真实主体，但仍有若干“能力存在、承诺偏强或边界未收口”的部分完成项。

已完成：

- `cmd/start.go -> internal/server/server.go -> router/handlers` 的主入口真实可达；`go build ./...` 与 `go test ./... -count=1` 已在当前工作区通过。
- 前端 `tsc --noEmit`、`build-web-static.mjs`、以及 repo-local no-browser 验证链全部通过，说明当前工作区不是“看起来实现了但基本构建已坏”的状态。
- `Copilot` 与 `Antigravity` 后端 handler、transformer、模型发现与前端表单接线都是真实实现，不是 UI-only 壳子。
- providers 预设库的远端源已经固定到 immutable commit raw URL，并带有 source / commit response header 与 fallback 逻辑。
- AI 自动化任务、历史、typed payload、数据库恢复、Profile 详情与激活主链都已存在并通过后端测试。
- 备份 / 导入 / rollback 主链是真实可用实现，且 no-browser 验证已经覆盖其主要逻辑契约。
- `scripts/smoke-win-backend.ps1` 已证明最小本地后端烟测主链真实可用：`/`、`/manifest.json`、`/healthz`、登录、建 channel/group/apikey 和 `/v1/chat/completions` 转发都通过。

部分完成：

- AI Profile rich payload 已保存和展示，但真实下游消费仍主要限于执行配置字段。
- 多 AI split mode 仍是前端 lane fan-out 到多个普通 `/api/v1/ai/tasks`，不是服务端 orchestrator graph；不过这一点已被当前需求与实现计划明确承认。
- 协议转换高级边界仍有 TODO，核心路径可用但边角能力未完全闭环。
- 备份高级迁移工具仍明确 partial；这一点在 UI 和 docs 中都已保守表述。

未完成 / 疑似空实现：

- 本轮没有再确认到“路由已声明但主流程完全不调用”的纯空实现。
- 最接近“伪承诺”的是 Antigravity 默认 OAuth 文案：它不是路由空实现，而是 UI / README 宣称存在“内建默认 OAuth app”，但后端默认实现并未提供。

## 3. Verification Summary

已验证项：

- Git / 仓库基线：`git status --short --branch`、`git branch -a --verbose --no-abbrev`、`git log --oneline --decorate -n 20`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`git diff --stat origin/dev...HEAD`。
- Go 构建：在把 `GOCACHE/GOTMPDIR/GOMODCACHE` 重定向到仓库 `.tmp-*` 可写目录后，`go build ./...` 通过。
- Go 定向测试：`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1` 通过。
- Go 全量测试：`go test ./... -count=1` 通过。
- Windows 后端最小冒烟：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 通过。
- 前端静态类型检查：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过。
- 前端静态导出：`node scripts/build-web-static.mjs` 通过。
- no-browser 验证链：
  - `verify-locale-consistency.mjs`
  - `verify-home-layout.mjs`
  - `verify-channel-create-flow.mjs`
  - `verify-channel-presentation.mjs`
  - `verify-group-create-flow.mjs`
  - `verify-llm-price-boundary.mjs`
  - `verify-backup-logic.mjs`
  - `verify-backup-component.cjs`
  - `verify-help-hint-accessible.mjs`
  - `verify-dynamic-routing-help.mjs`
  - `verify-circuit-breaker-help.mjs`
  - `verify-model-probe-help.mjs`
  - `verify-setting-info-logic.mjs`
  - `verify-ai-config-profile-summary.mjs`
  - `verify-ai-automation-learning-focus.mjs`
  - `verify-ccswitch-flow.mjs`
  - `verify-route-target-copy.mjs`
  全部通过。

未验证项：

- Vitest：
  - `node web/node_modules/vitest/vitest.mjs run --config web/vitest.config.ts`
  - `node --require scripts/vitest-no-spawn.cjs ...`
  - `node ... vitest.mjs run --config web/vitest.config.ts --configLoader runner`
  三条路径都仍报 `spawn EPERM`。说明当前宿主环境禁止 Vitest 所需的子进程 / worker / Vite config loader 运行，本轮无法在本机证明 `pnpm run test` 的真实结果。
- Linux backend smoke：本轮未跑 `scripts/smoke-linux-backend.sh`，且 `wsl -l -v` 当前返回 `Wsl/EnumerateDistros/Service/E_ACCESSDENIED`。
- Docker compose smoke：本轮未跑 `scripts/smoke-docker-compose.sh`，当前 `docker --version` 不可用。
- 手工前端验收：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md` 本轮未做人眼验收。

执行过程中已自行解决的问题：

- `scripts/use-go-env.ps1` 首次执行时继承了不可写的 `D:\DevCache\go-build`，导致 `go build` 失败；本轮通过显式重定向 `GOCACHE/GOTMPDIR/GOMODCACHE` 到仓库可写目录解决。
- `build-web-static.mjs` 首次执行被陈旧 `web/.next/lock` 卡住；本轮确认无活跃构建依赖后，安全移开锁文件并成功重跑构建。
- `pnpm` 不在 PATH，但本轮通过 repo-local Node 与 `web/node_modules` 直接入口完成了 `tsc`、Vitest 旁路探测和所有 no-browser 校验，避免因为 shell 链接缺失而中断审计。

## 4. Comparison Notes

当前工作区 vs `HEAD`：

- 当前工作区依然非常脏：`git status --porcelain=v1` 共 `425` 行，其中 `287` 行为未跟踪文件。
- 但与早期多轮审计不同，本轮没有再确认到“当前工作区本身已打断构建 / Go gate”的硬回归。相反，Go 全量测试、前端 `tsc`、静态导出、Windows 后端烟测与 no-browser 验证都已在当前工作区通过。
- 本轮仅有两个执行层阻塞：一次是不可写 `GOCACHE`，一次是陈旧 `.next/lock`；两者都已被主线程自行处理，不属于源码回归。

当前分支 vs 主分支 / 稳定基线：

- `origin/dev...HEAD = 0 behind / 22 ahead`。
- 这说明已提交分支本身不是“严重落后上游的残破分叉”；当前最大的管理风险仍是巨量未提交工作区差异，而不是分支落后问题。

代码实现 vs README / docs / 任务说明：

- 当前 README、validation/release workflow、AI 自动化需求与实现计划，对多 AI split mode 的“前端 lane fan-out”定位已经基本一致，不再存在早前那种“文档把它写成后端 orchestrator”的强错位。
- 备份高级迁移仍为 partial 这一点，代码、测试与 docs 也已经统一口径。
- 当前最明显的文档 / UI / 代码不一致点是 Antigravity：README 和前端提示宣称默认 OAuth app 已内建，但后端实际仍要求显式环境变量。
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 的环境结论已经过时，仍会把当前工作区描述成“无法直接复跑”。

代码实现 vs 测试 / 构建 / 验证覆盖：

- Go 主链、本地 no-browser 校验、静态导出链与 Windows 后端最小冒烟已经与当前工作区实现相一致，且都已通过。
- CI 定义里的 Vitest 仍未能在本机证明，原因不是单个测试失败，而是宿主禁止其所需的进程 / worker 池创建。
- Docker / Linux smoke 与手工前端验收本轮都未执行，因此“当前实现可本机通过的自动化链”与“全部发布链”之间仍有未验证空白。

## 5. Top Next Actions

1. 修正 Antigravity 默认 OAuth 承诺。
要求：二选一，且必须同步对齐 README、前端文案、后端行为。
- 方案 A：真正提供内建默认 `client_id/client_secret` 或等效默认 app 注入，让“开箱即用”变成真。
- 方案 B：撤回 README / UI 的“built-in default OAuth app / works out-of-the-box”口径，明确要求管理员先配置环境变量。

2. 明确 AI Profile rich payload 的真实产品边界。
要求：至少选一条落地。
- 方案 A：为 `grouping_suggestions / price_items / classification / issues / suggested_actions` 接上至少一条真实后端 validator 或 downstream consumer。
- 方案 B：维持现状，但把产品与文档表述收敛为“结构化预览与审阅结果”，不要继续暗示已形成完整自动执行闭环。

3. 收口协议转换高级 TODO 与验证空白。
要求：优先覆盖 `json_schema`、Anthropic 复杂 `tool_result`、prediction / audio / image-generation` 相关分支。
- 短期至少补直接回归测试或明确标注“暂不支持”。
- 若要继续对外强调“无缝协议转换”，需要先把这些 TODO 边界和测试证明补齐。

建议下一步动作：

- 更新 `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 中已经过时的“当前无法直接复跑”结论，避免继续误导接手人。
- 补一轮 `.gitignore` 清理，把 `.tmp*`、cache、patch、日志、`.next`、`.tools` 等本地产物纳入忽略，降低后续审查噪音。
- 在具备可 fork/worker 的宿主上补跑 Vitest；在具备 Docker / Linux 环境时补跑 `smoke-linux-backend.sh` 与 `smoke-docker-compose.sh`。

## 本次运行中文摘要

1. 本次触发时间：`2026-04-29 12:55:34 +08:00`
2. 做了哪些检查、运行了哪些命令：读取 automation memory、git 状态 / 分支差异 / 最近提交 / 上轮审查；抽查 `cmd/start.go`、`internal/server/server.go`、`internal/server/handlers/{providers,copilot,antigravity,setting}.go`、`internal/op/ai_automation*.go`、`internal/op/backup.go`、`internal/relay/dynamic_mode.go`、`internal/transformer/*`、README 与 workflow；运行了 `go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1`、`go test ./... -count=1`、`powershell -File .\scripts\smoke-win-backend.ps1`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`、`node scripts/build-web-static.mjs`、17 条 no-browser 校验脚本，并三次旁路尝试 Vitest。
3. 修改了哪些文件：新增 `docs/review/审查/2026-04-29-125534-octopus-repo-complete-audit.md`、`docs/review/审查/2026-04-29-125534-octopus-repo-complete-audit.html`，更新 automation memory `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`；验证过程中刷新了被忽略的 `static/out` 导出产物；本次无业务代码变更。
4. 发现了什么问题：发现 1 个 `High` 问题，即 Antigravity 被宣传为内建默认 OAuth / 开箱即用，但后端默认实现并不提供对应默认值；发现 2 个 `Medium` 问题，即 AI Profile rich payload 仍主要停留在结构化预览层，以及协议转换高级边界仍有真实 TODO；发现 2 个 `Low` 问题，即 `CURRENT_STATUS_AND_PLAN` 仍保留过时环境结论、`.gitignore` 过窄导致工作区噪音过高。
5. 本次结果是成功、跳过还是失败：`成功`。本轮完整审计已落盘，且当前工作区的 Go 全量测试、Windows 后端烟测、前端 `tsc`、静态导出与 no-browser 自动化链均已通过。
6. 是否需要我手动介入：`需要`。优先是两类介入：一类是产品 / 文档决策（Antigravity 默认 OAuth 到底补实现还是收文案、AI Profile rich payload 到底做真实 consumer 还是收缩承诺）；另一类是环境支持（可用 Vitest / Linux / Docker 宿主或 CI）以补齐剩余证明链。
