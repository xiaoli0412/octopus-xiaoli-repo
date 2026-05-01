# Octopus 完整审计报告

- 仓库：`D:\GPT-codex\octopus_repo`
- 触发时间：`2026-04-25T22:34:22+08:00`
- 审计基线：当前工作区、`HEAD`（`bfa27ae` / `v0.1.3`）、`origin/dev`
- 产物位置：`docs/review/审查/`
- 说明：本次未修改产品代码，仅新增审计报告文件；宿主环境 `apply_patch` 不可用，因此使用 PowerShell 落盘报告。

## 1. Findings

### Critical

#### 1. 当前工作区把备份页主组件删掉了，但设置页主路径和测试仍在真实导入它

证据：

- [web/src/components/modules/setting/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/index.tsx:19) 仍保留 `const LazySettingBackup = dynamic(() => import('./Backup')...)`。
- [web/src/components/modules/setting/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/index.tsx:98) 仍在设置页真实渲染 `<LazySettingBackup />`。
- [web/src/components/modules/setting/Backup.test.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.test.tsx:7) 仍直接 `import { SettingBackup } from './Backup'`。
- `git diff -- web/src/components/modules/setting/index.tsx web/src/components/modules/setting/Backup.tsx` 明确显示：相对 `HEAD`，当前工作区删除了 `web/src/components/modules/setting/Backup.tsx`；而 `git show HEAD:web/src/components/modules/setting/Backup.tsx` 可证明 `HEAD` 里该文件原本存在。
- `Test-Path web/src/components/modules/setting/Backup.tsx` 结果为 `False`。

判断依据：这不是“功能未完成”，而是主路径 import 直接指向不存在文件。只要前端构建、类型检查或相关测试真正执行，这条链路就会失败。它是当前工作区最直接的发布阻塞项。

### High

#### 2. `config_source_mode / active_ai_profile_id` 只写入设置和 UI，没有接入真实运行时配置来源选择

证据：

- [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:431) 的 `AIProfileActivate()` 只做两件事：写入 `config_source_mode=ai_profile` 与 `active_ai_profile_id=<id>`。
- [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:438) 仅把这两个字段放进任务上下文 `Runtime` payload。
- [web/src/components/modules/setting/AIAutomationSource.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:20) 把这套开关直接暴露给用户。
- 对比之下，[internal/relay/dynamic_mode.go](D:/GPT-codex/octopus_repo/internal/relay/dynamic_mode.go:301) 确实会读取 `dynamic_routing_learning_enabled` 并影响候选评分，说明“真实接线”在仓库内是有先例的。

判断依据：这是典型“看起来实现了，但主流程未接入”。当前 `manual / ai_profile` 切换更像设置层元数据，不足以证明业务配置来源真的发生切换。

#### 3. AI 自动化主线没有通过自己的关键行为测试，远端模型发现和异步任务执行都不稳定

证据：

- `go test ./internal/op -run 'TestAIAutomationFetchModelsPrefersRemoteFreeCandidate|TestAIAutomationRunTaskSuccess|TestBackupExportImportAndRollbackLatest' -count=1` 中，[internal/op/ai_automation_test.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:60) 断言 `remote_model_discovery`，实际得到 `local_model_cache`。
- `go test ./internal/op ./internal/server/middleware ./internal/relay/balancer -count=1` 中，[internal/op/ai_automation_test.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:174)、[274](D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:274)、[393](D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:393) 多次出现任务最终仍为 `running` / `pending`。
- [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:317) 在建任务后立即 `AITaskStartAsync(task.ID)`。
- [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:125) 通过 goroutine 异步执行；测试输出还出现 `failed to read response body: context deadline exceeded`，说明执行链真实不稳，而不是单纯断言写错。

判断依据：AI 自动化已经不只是“骨架未写”，而是核心路径存在真实行为偏差；现有测试已经把问题打出来了。

#### 4. 备份/导入/回滚测试基线被自动初始化管理员与默认设置污染，合同测试大面积失真

证据：

- [internal/op/op_test.go](D:/GPT-codex/octopus_repo/internal/op/op_test.go:13) 的测试初始化统一调用 `InitCache()`。
- [internal/op/cache.go](D:/GPT-codex/octopus_repo/internal/op/cache.go:9) 中 `InitCache()` 会先执行 `UserInit()`。
- [internal/op/user.go](D:/GPT-codex/octopus_repo/internal/op/user.go:86) 的 `UserInit()` 会在空库时自动创建管理员。
- 同一轮 `go test ./internal/op ...` 中，[internal/op/backup_test.go](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:54) 期望只导出 `backup-admin`，实际导出出现额外 `admin`；[87](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:87)、[472](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:472)、[776](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:776)、[1084](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:1084) 等多处出现 `settings.key` 唯一约束冲突；[2024](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:2024) 与 [2223](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:2223) 也直接体现了用户数/回滚基线被污染。

判断依据：当前备份导入合同测试不是零星红点，而是测试前置状态与业务假设已经脱节，导致大量断言同时失效。

### Medium

#### 5. 文档状态与当前实现明显失配，尤其是 AI 自动化与备份主线描述已经过时

证据：

- [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49) 仍写着“AI 自动化中心主线状态：代码实现未开始”。
- 但当前工作区已经存在：
  - [internal/server/handlers/ai_automation.go](D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:16) 的 AI 自动化路由
  - [web/src/route/config.tsx](D:/GPT-codex/octopus_repo/web/src/route/config.tsx:14) 的顶层 `AI Automation` 路由入口
  - [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:1) 的页面实现
- 同一状态文档在 [32](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:32)、[33](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:33)、[544](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:544) 仍把 `Backup.tsx` 描述成“细节清理与浏览器证据待补”的进行中状态；而当前工作区该文件已经被删除。

判断依据：文档不只是“有点落后”，而是已经会误导接手人对真实实现状态和风险优先级的判断。

#### 6. 验证链在当前宿主上只能部分执行，导致前端和完整构建仍有实质性未验证区间

证据：

- `node scripts/verify-backup-component.cjs` 启动即崩溃，报错 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`。
- `pnpm --version` 失败，系统未安装 `pnpm`。
- `corepack --version` 失败，`corepack.cmd` 无法正常启动。
- `docker --version` 失败，当前宿主没有 Docker。
- `go build ./...` 失败，不是代码编译报错，而是无法解析 `proxy.golang.org`，导致缺失模块拉取被网络/DNS 阻断。

判断依据：这不是仓库单点 bug，但它直接阻塞了 README 声称的验证链复现，也让部分高风险前端问题无法在本机得到二次证明。

## 2. Completion Assessment

### 总体判断

- `HEAD`（`v0.1.3`）这一已提交基线更接近“可发布分支”。
- 当前工作区在这个基线上叠加了大体量未提交改动，主线集中在 `AI 自动化 + 动态路由学习 + 备份导入升级 + 大量测试/脚本补齐`，但完成度明显低于文档和测试表面呈现出来的状态。
- 以“当前仓库工作区”作为交付对象来看，当前不应视为可发布状态。

### 已完成

- 核心网关、渠道/分组/模型等既有主线在已提交基线中完成度较高。
- `AI 自动化` 顶层入口、后端 handler、数据模型、前端页面骨架已经落地，不再是纯文档规划。
- 动态路由学习的运行时读取链路已真实接入评分逻辑。
- `internal/server/middleware` 与 `internal/relay/balancer` 当前可通过 `go test`，说明部分底层模块仍保持稳定。

### 部分完成

- AI 自动化中心：页面、接口、任务模型、Profile 结构均已出现，但关键行为未稳定。
- 备份/导入/回滚：后端合同、前端测试、脚本和文档都很多，但当前主路径已回归，测试基线也不可靠。
- 动态路由学习：主逻辑已接线，但其与 AI 自动化/配置来源的整体闭环仍不完整。
- 验证体系：测试、脚本、workflow 很多，但关键链路未形成稳定“绿灯”。

### 未完成或疑似空实现

- `config_source_mode / active_ai_profile_id` 的真实运行时消费。
- AI 自动化任务执行链的稳定闭环。
- 当前工作区中的备份页前端主路径。
- 可复现的全量构建/前端验证/容器烟测闭环。

## 3. Verification Summary

### 已验证项

- 已完成仓库结构、git 状态、最近提交、分支关系与 `origin/dev` 基线核对。
- 已完成当前工作区与 `HEAD` 的重点差异核对，确认 `Backup.tsx` 在工作区中被删除。
- 已完成 README / 状态文档 / 当前实现之间的一致性核对。
- 已运行并复现 `internal/op` 的关键失败用例。
- 已运行并确认 `internal/server/middleware`、`internal/relay/balancer` 的测试通过。
- 已确认 AI 自动化与动态路由关键接线位置的源码证据。

### 未验证项

- 前端 `tsc` / `vitest` / `build:static` / `verify-backup-component.cjs` 未能在本机完成，因为 Node/pnpm/corepack 环境异常。
- 全仓 `go build ./...` 未能完成，因为当前宿主无法解析 `proxy.golang.org`。
- Docker / compose 烟测未能执行，因为当前宿主没有 Docker。
- 浏览器级 smoke 和 README 中列出的 Linux 脚本链路，本次未在该宿主上完成复现。

### 本次实际执行的关键命令

- `git status --short --branch`
- `git branch --all --verbose --no-abbrev`
- `git log --oneline --decorate -n 12`
- `git merge-base HEAD origin/dev`
- `git diff --stat origin/dev...HEAD`
- `git log --oneline origin/dev..HEAD`
- `git diff -- web/src/components/modules/setting/index.tsx web/src/components/modules/setting/Backup.tsx`
- `git show HEAD:web/src/components/modules/setting/Backup.tsx`
- 多轮 `rg -n ...` 与 `Get-Content` 源码/文档核对
- `go test ./internal/op -run 'TestAIAutomationFetchModelsPrefersRemoteFreeCandidate|TestAIAutomationRunTaskSuccess|TestBackupExportImportAndRollbackLatest' -count=1`
- `go test ./internal/op ./internal/server/middleware ./internal/relay/balancer -count=1`
- `go build ./...`
- `node scripts/verify-backup-component.cjs`
- `pnpm --version`
- `corepack --version`
- `docker --version`

## 4. Comparison Notes

### 当前工作区 vs `HEAD`

- 当前工作区非常脏，且变化规模远超普通补丁。
- 相对 `HEAD` 最危险的差异是：`Backup.tsx` 被删除，但设置页与测试还保留真实依赖。
- 当前工作区新增了大量 AI 自动化、动态路由、备份测试、脚本和文档，显示实现推进很快，但闭环程度不足。

### 当前分支 vs 稳定基线

- 当前分支 `feat/erguotou` 的 `HEAD` 与 `origin/feat/erguotou` 相同，且正好是 tag `v0.1.3`。
- `HEAD` 相对 `origin/dev` 的已提交差异总体是连贯的，包括 Copilot、Antigravity、providers、文档与 UI 增强。
- 真正的高风险主要来自 `v0.1.3` 之上的未提交工作区，而不是当前已提交分支基线本身。

### 代码实现 vs README / docs / 任务说明

- README 声称存在较完整的验证链和备份导入契约；当前工作区里，备份主路径已坏，`internal/op` 关键合同测试也不通过。
- 状态文档仍把 AI 自动化写成“未开始”，与真实代码状态不符。
- 备份 worklog / 状态描述仍默认 `Backup.tsx` 在持续收口，但当前工作区已经没有该文件。

### 代码实现 vs 测试 / 构建 / 验证覆盖

- 测试数量明显增加，但关键测试失败恰好暴露出主线还没闭环，而不是“测试缺失”。
- 存在通过的底层测试，但不足以证明 AI 自动化和备份主流程正确。
- 前端与容器验证链在当前宿主上无法完整执行，导致高风险 UI/构建问题缺少最终实证闭环。

## 5. Top Next Actions

### 需要优先处理的前三项

1. 恢复 `web/src/components/modules/setting/Backup.tsx`，或立即把设置页与测试中对 `./Backup` 的主路径引用下线，先消除前端硬性断路。
2. 把 `config_source_mode / active_ai_profile_id` 接入真实运行时配置来源，或者明确降级为“仅记录来源偏好”的 UI 状态，并补一条集成测试证明其真实效果。
3. 先修复 `internal/op` 的两类红灯：
   - AI 自动化的远端模型发现 / 异步任务完成链
   - 备份导入测试的基线污染（bootstrap admin / 默认 setting 初始化）

### 建议下一步动作

- 先以 `v0.1.3` 为稳定参照，对当前工作区做一次“可发布最小集”收敛，避免在坏掉的备份主路径之上继续堆功能。
- 在一台可用的宿主上补跑前端 `tsc` / `build:static` / `vitest` 与 Docker smoke，优先确认备份页、AI 自动化页和设置页是否还能完整构建。
- 修正 `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 与相关 worklog 的状态描述，避免后续接手人被旧状态误导。

---

## 中文摘要

1. 本次触发时间
- `2026-04-25T22:34:22+08:00`

2. 做了哪些检查、运行了哪些命令
- 检查了仓库结构、`git status`、分支、最近提交、`HEAD` 与 `origin/dev` 的差异。
- 核对了 README、状态文档、AI 自动化与备份相关源码、测试与路由接线。
- 运行了 `go test ./internal/op -run 'TestAIAutomationFetchModelsPrefersRemoteFreeCandidate|TestAIAutomationRunTaskSuccess|TestBackupExportImportAndRollbackLatest' -count=1`。
- 运行了 `go test ./internal/op ./internal/server/middleware ./internal/relay/balancer -count=1`。
- 尝试运行了 `go build ./...`、`node scripts/verify-backup-component.cjs`、`pnpm --version`、`corepack --version`、`docker --version`。

3. 修改了哪些文件
- 新增 [2026-04-25-223422-octopus-repo-complete-audit.md](D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-25-223422-octopus-repo-complete-audit.md)
- 新增 [2026-04-25-223422-octopus-repo-complete-audit.html](D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-25-223422-octopus-repo-complete-audit.html)
- 除上述报告文件外，本次无产品代码变更。

4. 发现了什么问题
- 当前工作区删除了 `Backup.tsx`，但设置页和测试仍导入它，属于前端主路径硬故障。
- `config_source_mode / active_ai_profile_id` 只有设置层写入，没有看到真实运行时接线。
- AI 自动化关键测试失败，远端模型发现和异步任务执行不稳定。
- 备份/导入/回滚测试被 bootstrap admin 与默认设置初始化污染，合同测试大面积失真。
- 状态文档与当前实现不同步，容易误导接手判断。
- 当前宿主缺少 pnpm/docker，corepack 异常，Node 存在启动断言错误，Go 构建还受 DNS 阻断。

5. 本次结果
- `成功`

6. 是否需要我手动介入
- `需要`。至少需要优先决定备份页是恢复实现还是暂时下线；如果要完成完整验证，还需要提供可用的 Node/pnpm/Docker/Go 依赖网络环境。
