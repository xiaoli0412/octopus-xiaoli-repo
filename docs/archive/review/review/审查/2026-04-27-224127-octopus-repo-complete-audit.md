# Octopus 完整审计报告

审计时间：2026-04-27 22:41:27 +08:00  
仓库：`D:\GPT-codex\octopus_repo`  
分支：`feat/erguotou`  
`HEAD`：`bfa27ae`  
相对 `origin/dev`：`0 behind / 22 ahead`  
审计开始时工作区相对 `HEAD`：`137 files changed, 18525 insertions(+), 3984 deletions(-)`  
说明：本次无业务代码变更，仅新增审查产物与 automation memory。

## 1. Findings

- `Critical` 前端验证/发布总闸当前确定性为红，AI 自动化测试 mock 已经脱离 `AIAutomationConfig` contract，导致 workflow 在真正进入构建与 no-browser 验证前就会失败。  
  依据：[`validation.yaml`](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:46) 与 [`release.yaml`](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:44) 都先执行 `pnpm exec tsc --noEmit`；[`build-web-static.mjs`](/D:/GPT-codex/octopus_repo/scripts/build-web-static.mjs:157) 也会先跑同一个 TypeScript preflight。当前 API contract 在 [`ai-automation.ts`](/D:/GPT-codex/octopus_repo/web/src/api/endpoints/ai-automation.ts:23) 已包含 `requested_active_ai_profile` / `active_ai_profile`；但 [`index.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.test.tsx:212) 的 `state.config` 重置对象省略了这两个字段，而 [`index.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.test.tsx:291)、[`index.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.test.tsx:308)、[`AIAutomationSource.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.test.tsx:285)、[`AIAutomationSource.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.test.tsx:318) 又把 `undefined` 写回更严格的推断 shape。实际复现命令 `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 本轮稳定报错。  
  影响：CI 的 validation/release job 还没走到 `build:static` 和 `test:screenshot-no-browser` 就会先红，属于当前工作区的发布阻塞。

- `High` AI task 的取消路径仍然存在竞争窗口，取消后步骤状态可能残留为 `running`，后台 goroutine 也会继续访问已关闭 DB。  
  依据：[`AITaskCancel()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:391) 现在已经调用 `cancelAITaskExecution()` 与 `markAITaskCanceledSteps()`，说明仓库确实尝试过收口；但异步入口 [`AITaskStartAsync()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:152) 仍然以单独 goroutine 持有超时上下文，而执行循环在 [`ensureAITaskRunnable()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:196) 与 [`startAITaskStep()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:204) 之间仍有 race window，取消发生在这个窗口时，步骤状态仍可能被重新写成 `running`。对应的 handler 回归测试 [`TestCancelAITaskHandlerMarksTaskCanceled`](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation_test.go:252) 要求取消返回后所有步骤都不再是 `running`，而实际复现命令 `go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20` 本轮稳定出现 `step collect_context status = "running"`，并伴随 `ai task execute failed (task=1): sql: database is closed`。  
  影响：后台任务取消语义与真实执行语义仍不一致，属于关键路径稳定性问题。

- `High` AI Profile 双轨主线目前只真正接到了“AI 自动化执行配置切换”层，没有实现文档承诺的“AI 生成方案切换/展示/消费”闭环。  
  依据：需求文档要求 AI 生成分组、渠道识别、价格识别、模型归类等建议，并保存为独立方案，同时在设置页切换到明确的 `AI 生成方案`，还要展示方案名称、更新时间、置信度与风险提示，见 [`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:20)、[`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:38)、[`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:45)、[`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:61)、[`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:79)、[`AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:107)。但当前 runtime consumer 在 [`applyActiveAIProfileConfig()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:492) 和 [`extractAIAutomationConfigFromProfile()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:583) 里只抽取 `base_url`、`api_key`、`channel_type`、`model`、`use_local_default`；设置页 [`AIAutomationSource.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:69) 到 [`AIAutomationSource.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:112) 也只是显示来源与 fallback 摘要；AI 自动化页 [`index.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:793) 到 [`index.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:829) 只是用 Profile 计算 effective endpoint/model。  
  影响：这条主线只能算部分完成，当前“AI 生成方案”更像一个命名超前的 UI/文档概念，而不是已被主流程真实消费的配置方案。

- `Medium` 动态路由文档口径已经落后于代码现实，`dynamic_routing_learning_enabled` 在代码里是实装能力，但文档仍把它描述成“下一阶段”。  
  依据：[`DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:144) 到 [`DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:208) 与 [`DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:171) 到 [`DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:188) 反复把该能力写成“下一阶段完成后”；但 relay 实际已经在 [`dynamic_mode.go`](/D:/GPT-codex/octopus_repo/internal/relay/dynamic_mode.go:278) 和 [`dynamic_mode.go`](/D:/GPT-codex/octopus_repo/internal/relay/dynamic_mode.go:301) 读取学习分数并受开关控制。  
  影响：不会立刻造成运行时故障，但会让后续接手者对当前完成度、测试边界和审计结论产生误判。

## 2. Completion Assessment

- 已完成：后端主服务、管理 API 主体、relay/balancer 主流程、动态路由学习真实接线、AI 自动化入口页、AI 配置接口、AI task 基本执行流程、AI Profile 存储与激活接口、设置页来源卡片、较大规模的 Go 与 repo-local no-browser 验证脚本。
- 部分完成：AI Profile 双轨主线的真正 consumer、AI task 取消与关闭收口、前端统一验证链、当前状态文档与动态路由文档同步。
- 未完成：文档承诺的“AI 生成方案”结构化消费闭环、取消后步骤状态完全收口、可在当前宿主完成的完整前端构建闭环。
- 疑似空实现：没有发现完全空壳的大块模块；但“AI 生成方案切换”目前更像 UI/文档命名超前，真实 consumer 仍只限于 AI 自动化执行配置层。
- 综合判断：当前完成度大约在 `78% - 82%`。仓库已经明显超过骨架阶段，但仍未到“主线能力与验证链都足以放心发布”的状态。

## 3. Verification Summary

### 已验证项

- `git status --short --branch`
- `git branch -vv`
- `git log --oneline --decorate -n 12`
- `git rev-list --left-right --count origin/dev...HEAD`
- `git diff --shortstat HEAD`
- `. .\scripts\use-go-env.ps1; & $env:GOEXE build ./...`：通过
- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/op -count=1`：通过
- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/server/middleware -count=1`：通过
- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/task -count=1`：通过
- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/update -count=1`：通过
- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`：稳定失败，真实复现取消后步骤残留为 `running`
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`：稳定失败，定位到 AI 自动化测试 mock 的类型漂移
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-home-layout.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-channel-create-flow.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-channel-presentation.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-group-create-flow.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-llm-price-boundary.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ccswitch-flow.mjs`：通过
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-route-target-copy.mjs`：通过

### 未验证项

- `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/relay/... -count=1`：当前宿主 `net.Listen(tcp4)` / WinSock provider 报 `The requested service provider could not be loaded or initialized`，无法据此直接给出业务回归结论。
- `. .\scripts\use-node-env.ps1; & $env:NODEEXE scripts/build-web-static.mjs`：当前宿主上的 Next/Turbopack 仍会在“creating new process / binding to a port”阶段报 `os error 10106`；这是当前 Windows 宿主限制，不足以直接定性为源码构建错误，但当前也不能把 Windows 本机构建闭环标记为已验证。
- `pnpm --dir web run test:screenshot-no-browser`：本机 `pnpm` 不在 PATH，已改为逐条分解脚本手工验证；因此“统一入口本身”未在本机直接跑通。
- Linux smoke、Docker smoke、browser-grade smoke、vitest 全量回归：本轮未在当前宿主闭环。

## 4. Comparison Notes

- 当前工作区与 `HEAD`：这是一个大体量进行中工作区，而不是干净发布快照。审计开始时相对 `HEAD` 已有 `137 files changed, 18525 insertions(+), 3984 deletions(-)`，意味着必须区分“当前施工面风险”和“已提交主线风险”。
- 当前分支与稳定基线：当前分支相对 `origin/dev` 为 `22 ahead / 0 behind`，说明它承载了长期私有扩展，而不是小补丁分支。
- 代码实现与 README / docs / 任务说明：README 和 workflow 主方向大体一致，但 AI 自动化需求文档对“AI 生成方案”的描述明显超前于当前 consumer；动态路由文档则反过来落后于代码现实。
- 代码实现与测试/构建/验证覆盖：repo-local no-browser 验证脚本大多为绿，Go `op/middleware/task/update` 也为绿；但 workflow 首先依赖的 `tsc` 当前确定性失败，handler 取消测试确定性失败，relay 网络相关测试又受宿主限制，所以“脚本数量很多”并不等于“当前工作区已被完整证明”。

## 5. Top Next Actions

- 需要优先处理的前三项：
- 1. 修复 [`index.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.test.tsx:212) 与 [`AIAutomationSource.test.tsx`](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.test.tsx:285) 这两组 AI 自动化测试 mock 的 contract 漂移，先把 `tsc --noEmit` 恢复为绿，让 validation/release workflow 能继续往下跑。
- 2. 修复 [`AITaskCancel()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:391) 到 [`AITaskStartAsync()`](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:152) 之间的取消 race，保证 cancel 返回后步骤状态与后台执行都真正收口。
- 3. 对齐 AI Profile 主线边界：要么补齐“结构化方案 + 设置页元数据 + 分组/渠道/价格 consumer”的真实闭环，要么收缩文档和 UI 文案，不再把当前实现包装成已完成的“AI 生成方案切换”。
- 建议下一步动作：
- 1. 先修 `tsc` gate，再在 CI 或正常宿主补跑 `build:static`、`test:screenshot-no-browser`、`go test ./internal/server/handlers ./internal/relay/... -count=1`。
- 2. 同步修复 AI task cancel 后，再把 `TestCancelAITaskHandlerMarksTaskCanceled` 提升为稳定守门测试。
- 3. 最后统一同步 [`CURRENT_STATUS_AND_PLAN.zh-CN.md`](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:47)、动态路由文档和 AI 自动化需求文档，减少后续审计反复被文档漂移误导。

## 中文摘要

1. 本次触发时间：2026-04-27 22:41:27 +08:00。
2. 做了哪些检查、运行了哪些命令：读取了 automation memory、审查目录、状态文档、AI 自动化/动态路由需求文档与关键源码；运行了 `git status --short --branch`、`git branch -vv`、`git log --oneline --decorate -n 12`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`go build ./...`、`go test ./internal/op -count=1`、`go test ./internal/server/middleware -count=1`、`go test ./internal/task -count=1`、`go test ./internal/update -count=1`、`go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`、`go test ./internal/relay/... -count=1`、`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\verify-locale-consistency.mjs`、`node .\scripts\verify-home-layout.mjs`、`node .\scripts\verify-channel-create-flow.mjs`、`node .\scripts\verify-channel-presentation.mjs`、`node .\scripts\verify-group-create-flow.mjs`、`node .\scripts\verify-llm-price-boundary.mjs`、`node .\scripts\verify-ccswitch-flow.mjs`、`node .\scripts\verify-route-target-copy.mjs`、`node scripts/build-web-static.mjs`。
3. 修改了哪些文件：新增 [2026-04-27-224127-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-224127-octopus-repo-complete-audit.md:1)、[2026-04-27-224127-octopus-repo-complete-audit.html](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-224127-octopus-repo-complete-audit.html:1)，更新 [memory.md](/C:/Users/李昊桐/.codex/automations/octopus-repo/memory.md:1)。本次无业务代码变更。
4. 发现了什么问题：发现 1 个 `Critical` 问题，即前端 `tsc` gate 已确定性失败并直接卡住 validation/release workflow；发现 2 个 `High` 问题，即 AI task 取消路径仍有 race、AI Profile 双轨主线只做到执行配置切换；发现 1 个 `Medium` 问题，即动态路由文档对 `dynamic_routing_learning_enabled` 的状态表述已落后于代码现实。
5. 本次结果是成功、跳过还是失败：成功。完整审计已完成并落盘，关键问题均已通过源码与实际命令坐实。
6. 是否需要我手动介入：需要。最优先需要你确认 AI Profile 到底要做成“真实方案切换”还是“仅 AI 自动化配置切换”；同时建议尽快在 CI 或网络栈正常的宿主补跑前端构建与 relay/handler 网络相关测试。
