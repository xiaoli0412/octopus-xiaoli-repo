# Findings

## Critical

### 1. `internal/op` 的启动契约与测试前提仍然冲突，`go test ./...` 真实失败

- 证据：`InitCache()` 先调用 `UserInit()` 再调用 `settingRefreshCache()`，见 [internal/op/cache.go](/D:/GPT-codex/octopus_repo/internal/op/cache.go#L9)、[internal/op/user.go](/D:/GPT-codex/octopus_repo/internal/op/user.go#L71)、[internal/op/setting.go](/D:/GPT-codex/octopus_repo/internal/op/setting.go#L124)。
- 证据：本次 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild` 中，`go test ./...` 在 `internal/op` 稳定失败，报出 bootstrap admin 和 `settings.key` 唯一约束冲突。
- 影响：当前工作区不是 release-clean，核心 Go 验证链没有通过。

## High

### 2. `scripts/verify-go-env.ps1` 仍会给出假绿结果，验证结论不可信

- 证据：本次脚本整体退出码为 `0`，但内部 `go test ./...` 已明确失败。
- 证据：脚本直接执行 `go test` / `go build`，没有在命令后显式检查外部进程退出码。
- 影响：任何依赖该脚本的本地审计或自动化记录，都可能把真实失败误记为成功。

### 3. `config_source_mode / active_ai_profile_id` 仍更像管理面元数据，而不是运行时选择器

- 证据：设置页只写 [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx#L20) 到 [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx#L31) 这些值。
- 证据：后端只在管理接口里读写它们，执行器也只是把它们塞进 AI 上下文 payload，见 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go#L20) 和 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go#L369)。
- 证据：仓库级检索只命中 setting/op/executor/tests/docs，没有看到 relay、channel/group/model 的真实 runtime consumer。
- 影响：UI 看起来能切换，但主流程未证明真的换了运行时配置来源。

## Medium

### 4. `ConfigSnapshot` 仍是进程内快照，任务重启后无法恢复原始运行配置

- 证据：任务创建时只会把快照放进内存缓存，见 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go#L289) 到 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go#L308)。
- 证据：执行器从 `sync.Map` 读取并应用快照，见 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go#L156) 到 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go#L302)。
- 证据：`AITask` 模型本身没有持久化的配置快照字段，见 [internal/model/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go#L119)。
- 影响：重启、复盘、重试都拿不到当时的 endpoint / key / model 组合，审计和恢复能力不足。

### 5. README、Dockerfile 和工作流的构建口径仍未完全统一

- 证据：README 仍写 `cd web && pnpm install && pnpm run build && cd ..`，见 [README.md](/D:/GPT-codex/octopus_repo/README.md#L88) 与 [README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md#L87)。
- 证据：`scripts/dockerfiles/Dockerfile.debian` 仍运行 `pnpm run build`，见 [scripts/dockerfiles/Dockerfile.debian](/D:/GPT-codex/octopus_repo/scripts/dockerfiles/Dockerfile.debian#L7)。
- 证据：当前 release / validation workflow 以及仓库构建脚本已经改成 `pnpm run build:static`，见 [.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml#L46)、[.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml#L48)、[scripts/build.sh](/D:/GPT-codex/octopus_repo/scripts/build.sh#L258)。
- 影响：开发者容易走到过时的构建路径，文档与现实不一致。

### 6. 状态文档明显过期

- 证据：[docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md#L49) 仍写“代码实现未开始”。
- 影响：会误导后续接手、审计和自动化判断。

## Low

### 7. `git diff --check` 仍有三个 EOF 空行问题

- 证据：`web/src/components/modules/channel/Form.tsx`、`web/src/components/modules/group/Editor.tsx`、`web/src/components/modules/setting/Backup.tsx` 都报告 `new blank line at EOF`。
- 影响：只影响 diff hygiene，但会持续增加 review 噪音。

# Completion Assessment

## 已完成

- 后端主流程已真实可达：`smoke-win-backend.ps1` 跑通了 `/healthz`、登录、渠道创建、分组创建、API Key 创建和 `/v1/chat/completions`。
- 前端静态导出链路可用：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 和 `D:\gol1\node.exe .\scripts\build-web-static.mjs` 都通过。
- 动态路由的 `dynamic summary scan` 口径已经收口为摘要/观测，不再是“自动改阈值”的伪实现。
- 备份/导入/回滚主界面和后端链路是真实实现，不是纯占位。

## 部分完成

- AI 自动化中心本身是真实接线的，但 `config_source_mode / active_ai_profile_id` 还没有被证明会改变真实 runtime source。
- `ConfigSnapshot` 只进程内暂存，功能能跑，但重启恢复与复盘能力不足。
- 构建口径已经向 `build:static` 收口，但 README 和 Debian Dockerfile 仍在用旧指令。

## 未完成

- `internal/op` 的测试契约收口，当前 `go test ./...` 仍失败。
- `manual / ai_profile` 的真实运行时消费闭环。
- 任务快照持久化与重启恢复。
- 文档、Dockerfile 和工作流的构建口径统一。

## 疑似表层实现

- `config_source_mode / active_ai_profile_id` 目前最接近表层实现：UI 有、setting 有、任务上下文有，但没看到真正的 runtime consumer。

## 总体判断

- 这个仓库已经不是“骨架工程”，但也还没进入 release-clean。
- 当前最主要的问题不是“功能完全没做”，而是“关键契约和验证链还没收口”。

# Verification Summary

## 已执行并通过

- `git status --short --branch`
- `git branch -vv --all`
- `git show -s --format=fuller HEAD`
- `git diff --stat HEAD`
- `git diff --stat origin/dev...HEAD`
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- `D:\gol1\node.exe .\scripts\build-web-static.mjs`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`
- 仓库检索：`rg` 针对 `ConfigSourceMode`、`ActiveAIProfileID`、`ConfigSnapshot`、`dynamic_routing_learning_enabled`、构建口径与状态文档

## 已执行并失败

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild`
- `git diff --check`

## 未复跑 / 未验证

- 本轮没有重新跑 `next build`、Vitest 和 Docker 相关命令。

# Comparison Notes

## 当前工作区 vs `HEAD`

- 当前分支是 `feat/erguotou`，`HEAD` 为 `bfa27ae`。
- 工作区相对 `HEAD` 有大规模 tracked 修改和大量未跟踪文件，审计对象应视为“当前工作区”，不能只看提交点。

## 当前分支 vs 最近稳定基线

- 相对 `origin/dev`，当前分支已经明显偏离上游基线，且当前工作区还叠加了大量未提交改动。

## 代码实现 vs README / docs

- `DYNAMIC_ROUTING_REQUIREMENTS` 和代码对 `dynamic summary scan` 的口径基本一致。
- `CURRENT_STATUS_AND_PLAN.zh-CN.md` 明显过期。
- README / README_zh / Debian Dockerfile 仍没有完全统一到 `build:static` 现实路径。
- `USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md` 仍要求 `manual -> ai_profile -> manual` 的真实切换，但当前实现只证明了设置层写入。

## 代码实现 vs 测试 / 验证覆盖

- `build-web-static` 和 Windows backend smoke 给出了正向证据。
- `go test ./...` 仍被 `internal/op` 拖失败。
- `verify-go-env.ps1` 的可信度不足，因为它会对真实失败给出假绿结果。

# Top Next Actions

1. 先定 `manual / ai_profile` 的真实产品语义。如果它要成为 runtime selector，就把它接到真正的运行时 consumer；如果只是管理元数据，就把 UI 和文档承诺降级。
2. 收口 `internal/op` 的 bootstrap / default-setting 契约，统一 `UserInit()`、`settingRefreshCache()` 和测试前提，让 `go test ./...` 回到稳定绿色。
3. 修复 `scripts/verify-go-env.ps1` 的退出码传播，必要时顺手收紧 `-GoFmt` 扫描范围，避免假绿。
4. 继续补 `ConfigSnapshot` 持久化，这是审计与重试能力的关键缺口。
5. 同步 README、README_zh、Dockerfile 和工作流，把构建口径统一到 `build:static` 真实路径。

## 中文摘要

1. 本次触发时间
- `2026-04-25T09:50:28+08:00`

2. 做了哪些检查、运行了哪些命令
- 检查了仓库结构、`git status`、分支与 `HEAD`、最近提交、审查目录、worklog、README/docs、关键实现文件。
- 运行了 `git status --short --branch`、`git branch -vv --all`、`git show -s --format=fuller HEAD`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、`git diff --check`。
- 运行了 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild`、`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`D:\gol1\node.exe .\scripts\build-web-static.mjs`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`。
- 运行了多组 `rg` 搜索，用于核对 `ConfigSourceMode`、`ActiveAIProfileID`、`ConfigSnapshot`、`dynamic_routing_learning_enabled`、构建口径和状态文档。

3. 修改了哪些文件；如果没有修改，明确写“本次无变更”
- 新增 [2026-04-25-095028-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-25-095028-octopus-repo-complete-audit.md)
- 新增 [2026-04-25-095028-octopus-repo-complete-audit.html](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-25-095028-octopus-repo-complete-audit.html)

4. 发现了什么问题
- `internal/op` 的测试契约失败、`verify-go-env.ps1` 假绿、`config_source_mode / active_ai_profile_id` 没有 runtime consumer、`ConfigSnapshot` 未持久化、README / Dockerfile / workflow 口径漂移、状态文档过期、三个 EOF 空行。

5. 本次结果是成功、跳过还是失败
- 审计成功完成并落盘，但仓库本体仍未达到 release-clean。

6. 是否需要我手动介入
- 需要。
- 优先介入点是：`internal/op` 的测试契约、`manual / ai_profile` 的真实 runtime 语义、以及 `verify-go-env.ps1` 的可信度修复。
