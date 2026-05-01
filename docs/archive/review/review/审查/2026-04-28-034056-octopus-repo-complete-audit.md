# Octopus 完整审计报告

审计时间：2026-04-28 03:40:56 +08:00  
仓库：`D:\GPT-codex\octopus_repo`  
分支：`feat/erguotou`  
`HEAD`：`bfa27ae`  
稳定基线：`origin/dev`  

本轮先读取了 automation memory、仓库结构、`git` 状态、最近提交和核心文档，再按当前工作区、`HEAD`、`origin/dev`、README/docs、构建/测试覆盖做交叉核查。结论和上一轮相比有明显更新：此前的 `tsc` gate 和 AI task cancel 主问题，在当前工作区与 repo-local 环境脚本下已经不再是本轮主 finding。

## 1. Findings

本轮没有确认到新的 `Critical` 级代码回归；当前最重的问题集中在 `High` / `Medium`。

- `High` AI Profile 仍没有形成需求承诺的“结构化建议”闭环，当前保存的内容主要是原始输出与执行元数据，后端真正能消费的仍只有少量执行配置字段。  
  依据：需求文档要求 `grouping / channel_recognition / price_recognition / model_classification / config_health` 等 Profile 保存“结构化内容”，且必须能被前端预览和后端校验，不能只是不可解析自然语言，见 [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:83) 和 [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:97)。但当前生成逻辑把 `raw_output`、`summary`、`tool_execution_summary`、`config/runtime` 直接写入结果与 Profile 内容，见 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:364) 和 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:382)；后端真正解析 Profile 时只抽取 `base_url / api_key / channel_type / model / use_local_default`，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:593)；前端 Profile 预览也只是直接渲染 `content_json`，见 [index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:2265)。这说明“AI 生成方案”目前更像元数据封装，而不是可被后续链路真正消费的结构化运营方案。

- `Medium` 文档承诺的“历史任务区/历史任务列表”还没有真正实现，当前 API 和前端只覆盖“创建任务 + 轮询单个任务 + 取消当前任务”。  
  依据：需求要求 AI 自动化中心展示历史任务区，实施计划也明确列了“历史任务列表”，见 [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:69) 和 [AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md:118)。但当前 handler 只有 `POST /tasks`、`GET /tasks/:id`、`POST /tasks/:id/cancel`，没有 list/history 接口，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:27)；前端状态也只持有 `latestTask = taskQuery.data || createTask.data`，见 [index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:833)。这意味着任务“历史”在当前实现里仍是缺口，不是主流程已接线但没展示的小问题。

- `Medium` AI task 的配置快照和运行态仍只存在进程内内存，`AITask` 行被写进数据库，但任务真正执行所需的 snapshot/context 没有持久化，因此进程异常退出后会留下语义不完整的 `pending/running` 任务。  
  依据：`AITaskCreate()` 在入库后只把 `ConfigSnapshot` 放进进程内 `sync.Map`，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:326)；运行时快照容器本身也是内存态 `aiTaskRuntimeConfig sync.Map`，见 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:43) 和 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:503)；任务结束后还会主动删除，见 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:158) 和 [ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:525)。与此同时，模型层里的 `ConfigSnapshot` 只存在创建请求，不在 `AITask` 持久化字段中，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go:196)。这会让数据库里的任务记录看起来可追踪，但在 crash / 非优雅重启后并不具备真正的恢复语义。

- `Medium` `AIProfileActivate()` 只把新 Profile 标成 `active`，不会清理旧的 `active` 状态，数据库状态与真正运行时来源会逐渐漂移。  
  依据：激活逻辑只设置 `config_source_mode`、`active_ai_profile_id`，然后单独把当前 Profile 行更新为 `active`，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:459) 和 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:473)。前端判断“当前正在使用哪个 Profile”时，实际依赖的是 `resolvedActiveProfileID` 与所选 `id` 是否相等，而不是数据库里有多少个 `status=active` 的行，见 [AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:72) 和 [AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:120)。这不会立刻打断运行，但会让 API 消费者、后续统计和人工排查面对一个会累积脏状态的 Profile 表。

- `Medium` 测试/构建覆盖与代码实现仍不完全一致：仓库已经有 Vitest 组件测试和 `internal/update` 测试，但 validation/release workflow 都没有把它们纳入 gate。  
  依据：当前 CI gate 只跑 `pnpm exec tsc --noEmit`、`pnpm run build:static`、`pnpm run test:screenshot-no-browser` 和选定的 Go 包，见 [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:48) 和 [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:57)，release workflow 同样如此，见 [release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:46) 和 [release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:55)。但仓库内已经存在 [AIAutomationSource.test.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.test.tsx:1)、[index.test.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.test.tsx:1) 等前端组件测试，以及 [update_test.go](/D:/GPT-codex/octopus_repo/internal/update/update_test.go:1)。这意味着“测试存在”不等于“发布前会被执行”。

- `Low` AI 配置健康检查的命名在 docs 与代码之间仍有轻微漂移，需求文档写 `config_health`，代码与前端统一使用 `config_health_check`。  
  依据：需求文档列的是 `config_health`，见 [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:83)；而后端 task/profile 常量和前端任务类型都使用 `config_health_check`，见 [ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go:14)、[ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go:44) 和 [index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:252)。这更像一致性瑕疵，但会给后续检索、文档维护和测试命名带来噪音。

## 2. Completion Assessment

- 已完成：
  - 多渠道聚合、核心 relay/balancer 主流程、Copilot/Antigravity/provider 相关扩展、基础设置与管理 API 主链。
  - AI 自动化中心的页面骨架、AI 配置接口、模型发现、任务创建/轮询/取消、Profile 激活、动态路由学习展示与开关。
  - repo-local 的 Go / Node 环境脚本、TypeScript 检查、静态导出和大部分 no-browser 验证脚本。

- 部分完成：
  - AI Profile 双轨主线已经能切换“执行配置来源”，但还没有把分组/价格/模型归类等 AI 建议做成真正可消费的结构化方案。
  - AI task 的数据库记录、进度步骤和取消主路径已经可用，但仍不具备可靠的跨进程恢复语义。
  - 任务中心已经有“当前任务 + 结果 + Profile 预览”，但没有历史任务列表闭环。

- 未完成：
  - 文档承诺的历史任务区。
  - 各类 AI Profile 的结构化内容 schema、校验和真实 consumer。
  - 把现有 Vitest / `internal/update` 测试纳入正式 CI gate。

- 疑似空实现：
  - 本轮没有发现“完全没接线但假装已执行”的旧问题；上一轮关于 `profile_activate` / `snapshot_guard` 的伪接线，在当前工作区已不再成立。
  - 当前更像“实现了任务壳层和配置切换壳层，但结构化运营语义仍然停留在半成品”。

- 综合完成度判断：
  - 整体仓库主线完成度约 `84% - 88%`。
  - 其中 AI 自动化中心作为单独子主线，更接近 `70% - 75%`：框架和 UI 已经比较完整，但需求最核心的“结构化建议 + 历史 + 可持续任务语义”还没有收口。

## 3. Verification Summary

### 已验证项

- Git / 基线 / 文档：
  - `git status --short --branch`
  - `git branch --show-current`
  - `git branch -vv`
  - `git branch -a`
  - `git log --oneline -n 12`
  - `git rev-list --left-right --count origin/dev...HEAD`
  - `git diff --shortstat HEAD`
  - `git diff --stat origin/dev...HEAD`
  - 抽查了 `README.md`、`CURRENT_STATUS_AND_PLAN.zh-CN.md`、AI 自动化需求/实施计划、review/worklog 历史记录。

- Go 构建 / 测试：
  - `. .\scripts\use-go-env.ps1; go build ./...`：通过。
  - `. .\scripts\use-go-env.ps1; go test ./internal/op -count=1`：通过。
  - `. .\scripts\use-go-env.ps1; go test ./internal/server/middleware -count=1`：通过。
  - `. .\scripts\use-go-env.ps1; go test ./internal/task -count=1`：通过。
  - `. .\scripts\use-go-env.ps1; go test ./internal/update -count=1`：通过。
  - `. .\scripts\use-go-env.ps1; go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`：通过，说明此前这条取消竞态 finding 在当前工作区主路径上已不再成立。

- 前端 TypeScript / 静态构建 / no-browser：
  - `. .\scripts\use-node-env.ps1; node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`：通过。
  - `. .\scripts\use-node-env.ps1; node scripts/build-web-static.mjs`：通过，并把 `web/out` 同步到了 `static/out`。
  - 通过的 repo-local no-browser / logic 脚本：
    - `verify-locale-consistency.mjs`
    - `verify-setting-info-logic.mjs`
    - `verify-home-layout.mjs`
    - `verify-channel-create-flow.mjs`
    - `verify-channel-presentation.mjs`
    - `verify-group-create-flow.mjs`
    - `verify-llm-price-boundary.mjs`
    - `verify-help-hint-accessible.mjs`
    - `verify-dynamic-routing-help.mjs`
    - `verify-circuit-breaker-help.mjs`
    - `verify-model-probe-help.mjs`
    - `verify-ai-config-profile-summary.mjs`
    - `verify-ai-automation-learning-focus.mjs`
    - `verify-ccswitch-flow.mjs`
    - `verify-route-target-copy.mjs`
    - `verify-backup-logic.mjs`
    - `verify-backup-component.cjs`

### 未验证项

- `. .\scripts\use-go-env.ps1; go test ./internal/server/handlers -count=1`：本轮整包执行仍会夹杂当前宿主 `net.Listen(tcp4)` / WinSock provider 错误，无法把所有失败直接定性为仓库回归。
- `. .\scripts\use-go-env.ps1; go test ./internal/relay/... -count=1`：同样仍受当前 Windows 宿主 `socket: The requested service provider could not be loaded or initialized` 影响，属于环境阻塞下的未完全验证项。
- `. .\scripts\use-node-env.ps1; node -r scripts/vitest-no-spawn.cjs web/node_modules/vitest/vitest.mjs run ... --config web/vitest.config.ts`：仍在 `esbuild` 加载 `vitest.config.ts` 阶段报 `spawn EPERM`，因此前端组件测试本轮无法在当前宿主补跑。
- `scripts/smoke-linux-backend.sh` 与 `scripts/smoke-docker-compose.sh`：本轮未执行。当前宿主可用 `bash.exe`，但没有可直接调用的 `docker`，且用户要求本轮主线程尽快完成完整审计；因此 Linux/Docker 端到端 smoke 仍需在具备环境的宿主或 CI 上补跑。

## 4. Comparison Notes

- 当前工作区 vs `HEAD`：
  - `git diff --shortstat HEAD` 显示当前工作区相对 `HEAD` 有 `137 files changed, 19364 insertions(+), 4000 deletions(-)`。
  - 同时还混有大量构建缓存、临时日志、`node_modules`、`.next`、`.gocache`、`.gomodcache` 等未跟踪产物；这会明显放大“工作区 vs HEAD”的噪音。

- 当前分支 vs 稳定基线：
  - `git rev-list --left-right --count origin/dev...HEAD` 结果是 `0 22`，即当前分支相对 `origin/dev` 为 `22 ahead / 0 behind`。
  - `git diff --stat origin/dev...HEAD` 显示 `72 files changed, 6001 insertions(+), 1741 deletions(-)`，说明 `HEAD` 自身已经不是小修小补，而是较大规模扩展分支。

- 代码实现 vs README / docs / 任务说明：
  - README、validation/release workflow、静态构建脚本、动态路由学习主线，现在整体已经比前几轮更一致。
  - 仍存在两类偏差：
    - AI 自动化需求文档对“结构化建议/历史任务”承诺强于代码现实。
    - `config_health` / `config_health_check` 命名没有完全统一。

- 代码实现 vs 测试 / 构建 / 验证覆盖：
  - 好的一面：`go build`、`tsc`、`build-web-static` 和绝大多数 no-browser 脚本都已经能在当前仓库 + repo-local 环境脚本下跑通。
  - 风险面：
    - `relay` / 部分 `handlers` 网络型 Go 测试仍受宿主 WinSock provider 阻塞。
    - 仓库已有的 Vitest 组件测试没有进入正式 CI gate。
    - `internal/update` 虽然有测试且本轮本地通过，但也没有被 CI gate 覆盖。

## 5. Top Next Actions

- 需要优先处理的前三项：
  1. 给 AI Profile 定义真正可消费的结构化 schema，并让 `grouping / channel_recognition / price_recognition / model_classification / config_health_check` 至少有一条真实 consumer 或 validator；如果短期不做，就收缩 docs/UI 语义，不要继续把原始总结包装成“完整方案”。
  2. 补齐任务历史闭环：新增 task list/history API 和前端历史任务区；同时明确 AI task 是否需要跨进程持续，如果需要，就把 config snapshot / 恢复语义持久化；如果不需要，就在启动时清理陈旧的 `pending/running` 任务并写明这是一次性任务系统。
  3. 把现有 Vitest 和 `internal/update` 纳入 CI gate，并在正常 Linux/CI 宿主补跑 `internal/relay/...` 与全量 `handlers` 包，消除当前 Windows 网络栈带来的验证盲区。

- 建议下一步动作：
  - 先做一个产品/实现边界决策：AI 自动化当前到底是“执行配置助手”，还是“真正生成结构化运维方案的任务中心”。这个决策会直接决定是补 consumer，还是收缩文档承诺。
  - 随后做一个小而硬的收口批次：`task history + stale task policy + active profile status normalization + CI gate 增补`，这四项能明显提升这条主线的可信度。
  - 在 CI 或 Linux 宿主上补跑 `smoke-linux-backend.sh`、`smoke-docker-compose.sh`、Vitest 和 `relay/handlers` 全量测试，作为下一次“可发布”判断前置条件。

## 中文摘要

1. 本次触发时间：2026-04-28 03:40:56 +08:00。
2. 做了哪些检查、运行了哪些命令：读取了 automation memory、`git` 状态、分支/提交/差异、README 与 AI 自动化主线文档；运行了 repo-local `go build ./...`、`go test ./internal/op ./internal/server/middleware ./internal/task ./internal/update`、`go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20`；运行了 `tsc --noEmit`、`build-web-static.mjs` 和一整组 no-browser 脚本；尝试运行全量 `handlers` / `relay` 包测试与前端 Vitest，并记录了宿主级阻塞原因。
3. 修改了哪些文件：新增 [2026-04-28-034056-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-28-034056-octopus-repo-complete-audit.md:1)、[2026-04-28-034056-octopus-repo-complete-audit.html](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-28-034056-octopus-repo-complete-audit.html:1)，并更新 [memory.md](/C:/Users/李昊桐/.codex/automations/octopus-repo/memory.md:1)。本次无业务代码变更；另外刷新了被忽略的 `static/out` 构建产物。
4. 发现了什么问题：本轮最重要的真实问题不是编译或取消竞态，而是 AI Profile 仍缺少结构化内容闭环、任务历史未实现、AI task snapshot/运行态只在内存里、Profile `active` 状态会漂移，以及现有 Vitest / `internal/update` 测试没有进入正式 CI gate；另有一个 `config_health` 命名一致性低优先级问题。
5. 本次结果是成功、跳过还是失败：成功。本轮完整审计已经完成并落盘；部分未验证项属于宿主环境限制或 CI 才更适合验证的链路。
6. 是否需要我手动介入：需要。最需要你手动拍板的是两件事：
   - AI 自动化主线是否真的要落到“结构化方案中心”，还是收缩成“AI 配置/建议助手”；
   - AI task 是否必须支持跨进程可恢复/可追溯历史。如果答案是“要”，那就需要把持久化与历史 API 提到更高优先级。
