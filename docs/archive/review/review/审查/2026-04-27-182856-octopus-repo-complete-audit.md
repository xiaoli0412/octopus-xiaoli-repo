# Octopus 完整审计报告

审计时间：2026-04-27 18:28:56 +08:00  
仓库：`D:\GPT-codex\octopus_repo`  
分支：`feat/erguotou`  
`HEAD`：`bfa27ae`（`v0.1.3`）  
相对 `origin/dev`：`0 behind / 22 ahead`  
当前工作区相对 `HEAD`：`136 files changed, 18381 insertions(+), 3977 deletions(-)`  
说明：本次无业务代码变更，仅新增审查产物与 automation memory。

## 1. Findings

- `Critical` Windows 更新失败路径不安全，既可能在重启失败前先执行关闭动作，也会在当前实现下直接触发空指针 panic。
  依据：[internal/update/core.go](/D:/GPT-codex/octopus_repo/internal/update/core.go:107) 的 `restartExecutable()` 先调用 `shutdown.Shutdown()`，再尝试在 Windows 上 `startReplacementProcess()`；而 [internal/utils/shutdown/shutdown.go](/D:/GPT-codex/octopus_repo/internal/utils/shutdown/shutdown.go:46) 到 [internal/utils/shutdown/shutdown.go](/D:/GPT-codex/octopus_repo/internal/utils/shutdown/shutdown.go:52) 的 `Shutdown()` 假定 `ilog` 已初始化，失败路径没有保护。直接复现命令是 `go test ./internal/update -run TestRestartExecutableWindowsDoesNotExitWhenRestartFails -count=1`，本轮稳定失败并把栈打到 `shutdown.Shutdown()`。这意味着 Windows 自更新关键路径在“替换进程未能成功拉起”时并不安全，属于发布阻塞级问题。

- `High` AI Profile 双轨主线只实现了“AI 自动化执行配置切换”，没有实现需求文档承诺的“AI 生成方案切换/展示/消费”。
  依据：[docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:22) 到 [docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:26) 要求 AI 生成的分组、渠道识别、价格识别等建议保存为独立 AI Profile，并在设置页显式切换；[docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:110) 还明确要求展示方案名称、更新时间、置信度和风险提示。但当前运行时消费只发生在 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:19)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:484) 和 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:548)，它们只从 Profile 内容里抽取 `base_url`、`api_key`、`channel_type`、`model`、`use_local_default`。执行器保存的结果内容主要是 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:294) 到 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:355) 的 `raw_output`、`summary`、`config`、`runtime`、`tool_execution`，没有形成一个能被设置页或业务主流程消费的“分组/渠道/价格方案对象”。前端设置入口 [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:21) 到 [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:71) 也只支持模式切换、显示 `activeProfileID` 和跳转 AI 中心，缺少文档承诺的方案元数据与显式方案选择。因此这条主线目前只能算部分完成，而且存在“名字像方案切换，实际只是 AI 自动化 endpoint 配置切换”的认知风险。

- `High` AI task 取消与异步生命周期没有收口，取消后后台 goroutine 仍可能继续跑，步骤状态也会停留在 `running`。
  依据：[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:142) 到 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:153) 的 `AITaskStartAsync()` 用 `context.Background()` 启动 goroutine，与 server/test 生命周期脱钩；[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:385) 到 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:397) 的 `AITaskCancel()` 只更新 task 行的 `status/finished_at/progress`，并没有取消正在执行的上下文；而执行器步骤收口依赖 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:190) 到 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:205) 的运行期检查。直接复现命令是 `go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`，本轮稳定出现 `step collect_context status = "running"`，并伴随 `ai task execute failed (task=1): sql: database is closed`。这说明取消语义与异步执行语义当前并不一致，属于关键后台流程风险。

- `Medium` 当前宿主上的验证闭环弱于文档和工作区体量所需的置信度，部分关键路径无法在本机完整证明。
  依据：本轮 `scripts/verify-go-env.ps1 -GoBuild` 通过，但直接运行 targeted Go 测试时必须手工补齐 `LOCALAPPDATA`、`GOCACHE`、`GOTMPDIR` 才能工作；Codex 自带 `node.exe` 运行 `tsc` 仍会触发 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`；尝试调用外部 Node 时又报“拒绝访问”；同线程已验证 `internal/relay` 与部分 `internal/server/handlers` 测试仍受本机 `net.Listen(tcp4)` / WinSock provider 限制，Linux/Docker smoke 也无法在当前宿主闭环。这里更像环境限制，不宜直接算作仓库逻辑回归，但它会让“前端构建/网络相关行为/跨平台 smoke”仍处于未完全验证状态，属于发布风险而不是产品缺陷。

## 2. Completion Assessment

- 已完成：后端服务主链路、管理 API 主体、relay/balancer 主流程、AI 自动化中心入口、AI 配置接口、AI task 基本执行流程、AI Profile 存储与手动激活接口、动态路由学习状态接口、设置页配置来源卡片、较大规模的 Go/前端测试与验证脚本补充。
- 部分完成：AI Profile 双轨主线的真正 consumer、AI 结果结构化方案对象、AI task 取消/关闭收口、跨平台验证闭环、与当前实现一致的状态文档。
- 未完成：文档承诺的“设置页展示并选择 AI 方案元数据”“AI 方案对分组/渠道/价格等建议的真实消费闭环”“可证明的取消后步骤收口”。
- 疑似空实现或伪接线：严格来说不是“完全空实现”，而是“主线只接到了 AI 自动化执行配置层，没有接到文档承诺的 AI 生成方案层”。
- 综合判断：当前仓库完成度大约在 `80% - 85%`。已经明显超过骨架阶段，但还没有达到“关键路径、主流程接线、验证闭环都可放心发布”的状态。

## 3. Verification Summary

### 已验证项

- `git status --short --branch`：确认当前位于 `feat/erguotou`，工作区相对 `HEAD` 是大体量脏树。
- `git branch -vv`、`git log --oneline --decorate -n 12`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`：确认 `HEAD` 为 `bfa27ae`，相对 `origin/dev` 为 `0 behind / 22 ahead`，当前工作区相对 `HEAD` 为 `136 files changed, 18381 insertions(+), 3977 deletions(-)`。
- `scripts/verify-go-env.ps1 -GoBuild`：通过，说明当前工作区至少还能完成后端构建。
- `go test ./internal/update -run TestRestartExecutableWindowsDoesNotExitWhenRestartFails -count=1`：稳定失败，真实复现 Windows 更新重启路径的 panic。
- `go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`：稳定失败，真实复现 AI task 取消后 step 仍为 `running` 且后台协程继续访问已关闭 DB。
- 同线程较早阶段已验证的 repo-local 无浏览器脚本：`verify-setting-info-logic.mjs`、`verify-locale-consistency.mjs`、`verify-home-layout.mjs`、`verify-channel-create-flow.mjs`、`verify-channel-presentation.mjs`、`verify-group-create-flow.mjs`、`verify-llm-price-boundary.mjs`、`verify-help-hint-accessible.mjs`、`verify-backup-logic.mjs`、`verify-backup-component.cjs`、`verify-dynamic-routing-help.mjs`、`verify-circuit-breaker-help.mjs`、`verify-model-probe-help.mjs`、`verify-ccswitch-flow.mjs`、`verify-route-target-copy.mjs` 已通过。
- 同线程较早阶段已验证 `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过；这说明 TypeScript 类型层面没有立即暴露阻塞，但当前宿主上的 Node 启动链不稳定，后续复跑仍建议换宿主或走 CI。

### 未验证项

- 当前宿主上没有完成一次可信的前端完整构建/测试闭环；`tsc` 在本轮复跑时被宿主 Node 环境拦住，`vitest` 与 `build-web-static` 也受同类环境问题影响。
- `internal/relay` 与部分 `internal/server/handlers` 依赖监听 socket 的用例仍受本机 WinSock provider 限制，不能在本次直接用来判定业务回归与否。
- Linux smoke、Docker smoke、跨平台 shell 验证在当前宿主上未闭环，原因是 `bash`/Docker 能力缺失或受限，而不是仓库逻辑已被证伪。

## 4. Comparison Notes

- 当前工作区与 `HEAD`：这是一个大体量进行中工作区，核心增量集中在 AI 自动化、动态路由、设置页、验证脚本和测试补强。审计时必须区分“已提交主分支扩展能力”和“当前未提交施工面”。
- 当前分支与稳定基线：当前分支是 `feat/erguotou`，`HEAD` 为 `bfa27ae`，相对 `origin/dev` 已经 `22 ahead`、`0 behind`。这说明它不是轻量偏移，而是明显带有私有扩展能力的长期分支。
- 代码实现与 README / docs / 任务说明：README 与主实现方向大体一致，但 AI 自动化需求文档和当前代码的一致性不足，最大缺口在“AI 生成方案”的真实 consumer 与设置页元数据展示；这也是本次 High finding 的来源。
- 代码实现与测试/构建/验证覆盖：后端构建和不少 repo-local 验证脚本是真实可跑的，两条 targeted 失败测试也给出了强证据；但当前宿主不足以证明完整前端构建、部分网络相关 Go 用例、Linux/Docker smoke，从而让发布信心仍然依赖额外验证环境。

## 5. Top Next Actions

- 需要优先处理的前三项：
- 1. 修复 Windows 自更新重启路径，至少要让 [internal/update/core.go](/D:/GPT-codex/octopus_repo/internal/update/core.go:107) 的重启失败不会在关闭逻辑前触发 panic，也不要在确认替换进程无法拉起前就把当前实例带入关闭态。
- 2. 明确 AI Profile 双轨主线的 contract：如果它真的是“AI 生成方案”，就需要定义结构化 Profile schema、设置页元数据展示和真实 consumer；如果不是，就应下调文档/UI 表述，避免继续把 endpoint 配置切换包装成“方案切换”。
- 3. 为 AI task 引入真实的可取消上下文和任务级生命周期管理，保证 `cancel`、server shutdown、测试 teardown 都能收口步骤状态和后台 goroutine。
- 建议下一步动作：
- 1. 先修第 1 项，再补一个不依赖全局 logger 初始化的单测，确保 Windows 更新失败路径可重复验证。
- 2. 然后处理第 2 项，先做产品/实现边界决策，再决定是补 consumer 还是收缩需求。
- 3. 再处理第 3 项，并在一台网络栈与 Node 工具链正常的宿主或 CI 上补跑 `tsc`、`vitest`、`build-web-static`、`go test ./internal/relay ./internal/server/handlers -count=1`。

## 中文摘要

1. 本次触发时间：2026-04-27 18:28:56 +08:00。
2. 做了哪些检查、运行了哪些命令：读取了 automation memory、审查目录和仓库关键文档；运行了 `git status --short --branch`、`git branch -vv`、`git log --oneline --decorate -n 12`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`scripts/verify-go-env.ps1 -GoBuild`、`go test ./internal/update -run TestRestartExecutableWindowsDoesNotExitWhenRestartFails -count=1`、`go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`；并复核了 AI 自动化、更新、设置页、需求文档的关键代码位点。同线程较早阶段已经完成并沿用了多条 repo-local 验证脚本和一次 TypeScript `tsc` 验证结果。
3. 修改了哪些文件：新增 [2026-04-27-182856-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-182856-octopus-repo-complete-audit.md:1)、[2026-04-27-182856-octopus-repo-complete-audit.html](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-182856-octopus-repo-complete-audit.html:1)，更新 [memory.md](/C:/Users/李昊桐/.codex/automations/octopus-repo/memory.md:1)。本次无业务代码变更。
4. 发现了什么问题：发现 1 个 `Critical` 问题，即 Windows 更新失败路径会先进入关闭逻辑并可直接 panic；发现 2 个 `High` 问题，即 AI Profile 双轨主线只做到了执行配置切换、AI task 取消与后台生命周期未收口；发现 1 个 `Medium` 问题，即当前宿主无法完整闭环前端/网络相关验证，发布信心仍依赖额外环境。
5. 本次结果是成功、跳过还是失败：成功。完整审计已完成并落盘，且关键风险已经通过源码和 targeted tests 双重坐实。
6. 是否需要我手动介入：需要。优先需要你确认 AI Profile 到底要做成“真实方案切换”还是“仅 AI 自动化配置切换”，同时建议尽快安排一台环境正常的宿主或 CI 补跑前端/网络相关验证。
