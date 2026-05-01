# Findings

## Critical

### 1. 当前工作区未达到 release-clean，CI 目标 Go 测试命令在 `internal/op` 稳定失败

- 直接复现 CI 同款命令 `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`，结果稳定失败于 `internal/op`；同一轮中 [internal/server/handlers](D:/GPT-codex/octopus_repo/internal/server/handlers), [internal/server/middleware](D:/GPT-codex/octopus_repo/internal/server/middleware), [internal/relay](D:/GPT-codex/octopus_repo/internal/relay), [internal/task](D:/GPT-codex/octopus_repo/internal/task) 通过。
- 根因是测试契约与当前初始化路径脱节，而不是宿主权限问题。`setupOpTestDB()` 在初始化测试库后立刻调用 `InitCache()`，见 [internal/op/op_test.go](D:/GPT-codex/octopus_repo/internal/op/op_test.go:13)；`InitCache()` 又会先执行 `UserInit()` 和 `settingRefreshCache()`，见 [internal/op/cache.go](D:/GPT-codex/octopus_repo/internal/op/cache.go:9)。
- `UserInit()` 会创建 bootstrap admin，见 [internal/op/user.go](D:/GPT-codex/octopus_repo/internal/op/user.go:71)；`settingRefreshCache()` 会补齐默认 setting 并生成 `auth_token_secret`，见 [internal/op/setting.go](D:/GPT-codex/octopus_repo/internal/op/setting.go:124)。
- 这直接打破了多条测试前提：
- [internal/op/backup_test.go](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:21) 的 `TestDBExportAllIncludesSecretsByDefault` 期望只导出手工创建的用户，但实际多出 bootstrap admin。
- [internal/op/backup_test.go](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:84) 与 [internal/op/backup_test.go](D:/GPT-codex/octopus_repo/internal/op/backup_test.go:450) 等测试手工插入已有 key 的 setting 时，撞上 `UNIQUE constraint failed: settings.key`。
- [internal/op/user_test.go](D:/GPT-codex/octopus_repo/internal/op/user_test.go:5) 与 [internal/op/user_test.go](D:/GPT-codex/octopus_repo/internal/op/user_test.go:21) 试图验证 bootstrap 凭据，但环境变量覆盖发生在 `InitCache()` 之后，断言已经失效。
- 影响：当前工作区最关键的 Go 回归链没有通过，不能视为可发布状态。

### 2. `scripts/verify-go-env.ps1` 会给出假绿结果，Go 校验结论不可靠

- 脚本在执行 `go test ./...` 和 `go build ./...` 之后没有检查 `$LASTEXITCODE`，见 [scripts/verify-go-env.ps1](D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:110) 和 [scripts/verify-go-env.ps1](D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:115)。
- 本次实测 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild` 输出中明确出现 `FAIL`，但脚本整体退出码仍然是 `0`。
- 这不是理论风险，而是已发生的误报：脚本表面成功，实际 `internal/op` 失败。
- 影响：任何依赖该脚本的本地审计、人工回归或自动化记录，都可能把真实失败误记成 “Go 已通过”。

## High

### 3. AI 任务取消存在竞态，任务可标记为 `canceled`，但步骤仍停留在 `running`

- `AITaskCancel()` 只更新任务主记录的 `status / finished_at / progress`，没有同步步骤状态，也没有与正在运行的异步执行协调，见 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:357)。
- 压测复现命令 `go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=50` 本次稳定失败，断言位置在 [internal/server/handlers/ai_automation_test.go](D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation_test.go:239)。
- 复现过程中还能看到执行器输出 `sql: database is closed` 与 `no AI model candidates available`，说明取消与后台执行之间存在真实竞态窗口，位置可追到 [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:89)。
- 影响：用户会看到 “任务已取消，但步骤还在运行” 的不一致状态；相关测试也会出现波动性失败。

### 4. `manual / ai_profile` 看起来已实现切换，但主流程没有真实 runtime consumer

- 设置页会写入 `config_source_mode`，见 [web/src/components/modules/setting/AIAutomationSource.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:24)。
- 后端配置层会暴露 `ConfigSourceMode` 和 `ActiveAIProfileID`，见 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:18)。
- 执行器只把这两个字段记入 AI 任务上下文 payload，并没有据此解析真实运行配置来源，见 [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:358)。
- 前端 AI 自动化页面真正用于执行的仍是 `effectiveBaseURL / effectiveAPIKey / effectiveChannelType / effectiveUseLocalDefault / effectiveTaskModel` 这套手工拼出来的有效配置，见 [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:430), [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:506), [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:528), [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:864)。
- 现有测试只证明 setting 被切换，不证明切换后运行时行为真的改变，见 [internal/op/ai_automation_test.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:293) 与 [internal/server/handlers/ai_automation_test.go](D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation_test.go:177)。
- 影响：这是典型的“看起来做了管理面，但主流程未接入”；与 [docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md:129) 的双轨切换承诺不一致。

### 5. AI 任务 `config_snapshot` 只保存在进程内 `sync.Map`，任务不可持久恢复

- 任务创建时如果带 `ConfigSnapshot`，只会调用 `storeAITaskRuntimeConfig()` 放到内存 map，见 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:272) 和 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:291)。
- 执行器通过 `sync.Map` 读取快照，并在 goroutine 结束时删除，见 [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:37), [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:79), [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:277), [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:292)。
- 模型层 [internal/model/ai_automation.go](D:/GPT-codex/octopus_repo/internal/model/ai_automation.go:119) 的 `AITask` 并没有持久化配置快照字段。
- 影响：任务重启后无法可靠复盘或恢复原始运行配置，这与“可审计、可回看、可重试”的任务系统预期不一致。

## Medium

### 6. README / Dockerfile / CI 的构建口径仍未完全收口

- README 仍引导 `cd web && pnpm install && pnpm run build && cd ..`，见 [README.md](D:/GPT-codex/octopus_repo/README.md:88) 与 [README_zh.md](D:/GPT-codex/octopus_repo/README_zh.md:87)。
- 当前 CI / release 已统一使用 `pnpm run build:static`，见 [.github/workflows/validation.yaml](D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:48) 与 [.github/workflows/release.yaml](D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:46)。
- 仓库构建脚本也已切到 `build:static`，见 [scripts/build.sh](D:/GPT-codex/octopus_repo/scripts/build.sh:258)；但 Debian Dockerfile 仍直接跑 `pnpm run build`，见 [scripts/dockerfiles/Dockerfile.debian](D:/GPT-codex/octopus_repo/scripts/dockerfiles/Dockerfile.debian:7)。
- 本次 `D:\gol1\node.exe .\scripts\build-web-static.mjs` 通过，但原生 `next build` 仍因宿主 `spawn EPERM` 失败。
- 影响：文档、Docker 构建和仓库实际可行路径不一致，容易让开发者走到错误路径上。

### 7. 状态文档与当前代码实现明显漂移

- [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49) 仍写着 “当前状态：文档规划已立项，代码实现未开始”。
- 实际仓库已经存在 AI 自动化中心的迁移、handler、op、executor、页面与测试，见 [internal/db/migrate/012.go](D:/GPT-codex/octopus_repo/internal/db/migrate/012.go:13), [internal/server/handlers/ai_automation.go](D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:19), [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:15), [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:1), [web/src/components/modules/ai-automation/index.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:1)。
- 影响：后续接手人、审计流程和自动化判断会被错误状态说明误导。

### 8. 协议转换边界仍有真实 TODO，当前验证无法证明这些能力已完成

- Anthropic `tool_result` 其他结果类型仍是 TODO，见 [internal/transformer/inbound/anthropic/messages.go](D:/GPT-codex/octopus_repo/internal/transformer/inbound/anthropic/messages.go:162)。
- 统一模型层仍有 `Schema` 和 image-generation request body 的 TODO，见 [internal/transformer/model/model.go](D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:618) 与 [internal/transformer/model/model.go](D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:861)。
- Gemini `json_schema` 转换仍未完整实现，逻辑分支见 [internal/transformer/outbound/gemini/messages.go](D:/GPT-codex/octopus_repo/internal/transformer/outbound/gemini/messages.go:342)。
- 影响：chat / responses / embeddings 主链路已经有正向证据，但高级结构化输出和复杂工具场景仍只能视为部分完成。

### 9. `verify-go-env.ps1 -GoFmt` 扫到了 `.tools\go`，格式检查结果失真

- 脚本直接递归整个仓库根目录下所有 `*.go` 文件，见 [scripts/verify-go-env.ps1](D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:87)。
- 这会把 vendored toolchain 一并纳入格式检查范围；先前同链路已经定位到 `.tools\go\go\src\cmd\compile\internal\syntax\testdata\issue20789.go` 这类工具链测试样本。
- 影响：当前 `-GoFmt` 结果不能直接代表仓库 Go 源码格式状态。

## Low

### 10. `git diff --check` 仍有前端文件 EOF 空行问题

- 本次 `git diff --check` 报告 [web/src/components/modules/channel/Form.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Form.tsx:2219), [web/src/components/modules/group/Editor.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/group/Editor.tsx:1081), [web/src/components/modules/setting/Backup.tsx](D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx:1267) 存在 `new blank line at EOF`。
- 影响较小，但会持续污染 diff hygiene 与 review 噪音。

### 11. 前端静态构建仍有 `baseline-browser-mapping` 过期告警

- `build-web-static.mjs` 本次成功，但多次输出 `baseline-browser-mapping` 数据过旧告警。
- 影响不阻塞当前构建，但说明前端构建依赖仍有维护债务。

# Completion Assessment

## 已完成

- 后端主链路是真接线，不是空壳。`cmd/start.go -> db.InitDB -> op.InitCache -> server.Start -> task.Init/Run` 可以启动，并且 [scripts/smoke-win-backend.ps1](D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1) 本次实测跑通了 `/healthz`、登录、渠道创建、分组创建、API Key 创建和 `/v1/chat/completions`。
- 前端静态导出链路可用。`tsc --noEmit` 和 [scripts/build-web-static.mjs](D:/GPT-codex/octopus_repo/scripts/build-web-static.mjs:1) 本次都通过，并成功同步 `web/out -> static/out`。
- 动态路由不是伪实现。设置页摘要、学习状态、runtime wiring 与 no-browser 校验脚本都是真实存在的。
- 备份 / 导入 / 回滚主界面和接口骨架是真实实现，不是占位；相关 no-browser 校验脚本本次通过。
- AI 自动化中心不是文档壳：已有模型、迁移、handler、executor、任务、profile、页面和测试。
- `ai_automation_enabled` 不是空实现，它已经真实门控模型抓取、任务创建和执行，见 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:83), [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:92), [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:275), [internal/op/ai_automation_executor.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:149)。

## 部分完成

- AI 自动化双轨配置只完成了“设置层 + 管理层 + 任务上下文层”，没有完成“切换后真实驱动 runtime 配置来源”的闭环。
- AI 任务系统可以跑通一轮，但取消一致性、配置快照持久化、重启恢复还没收口。
- 协议转换主路径可用，但结构化输出、复杂 `tool_result`、image generation tool body 等边角能力仍是半实现。
- 构建链在包装脚本下可用，但 README / Dockerfile / 原生 `next build` 的口径还没有统一到同一现实路径。

## 未完成

- `manual / ai_profile` 的真实运行时配置切换。
- AI 任务配置快照持久化与重启恢复。
- 当前工作区的 CI 目标 Go 测试收口。
- Docker / compose 本机运行态验证。
- Vitest 与浏览器级验证在当前宿主上的完整闭环。

## 疑似空实现或表层实现

- `config_source_mode` / `active_ai_profile_id` 当前最接近“表层实现”：UI 有、setting 有、任务上下文有，但没有看到改变真实执行配置来源的消费点。

## 总体完成度评估

- 以当前工作区来看，仓库整体完成度大约在 `78%` 左右。
- 核心网关、管理面、备份与动态路由基础能力已经达到“真实可达、可构建、可做 smoke 验证”的阶段。
- 当前主要缺口不再是“能不能跑起来”，而是“高级能力是否真正接入主流程、是否可持续验证、是否可发布”。

# Verification Summary

## 已验证项

- `git status --short --branch`
- `git branch -vv --all`
- `git log --oneline --decorate -n 12`
- `git diff --stat HEAD`
- `git diff --stat origin/dev...HEAD`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`：通过
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`：通过
- `D:\gol1\node.exe .\scripts\build-web-static.mjs`：通过
- `D:\gol1\node.exe .\scripts\verify-backup-component.cjs`：通过
- `D:\gol1\node.exe .\scripts\verify-locale-consistency.mjs`：通过
- `D:\gol1\node.exe --experimental-strip-types .\scripts\verify-backup-logic.mjs`：通过
- `D:\gol1\node.exe .\scripts\verify-dynamic-routing-help.mjs`：通过
- `D:\gol1\node.exe .\scripts\verify-route-target-copy.mjs`：通过
- `D:\gol1\node.exe .\scripts\verify-home-layout.mjs`：通过
- repo-local `go build ./...`：通过

## 已验证失败项

- `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`：失败，核心失败包为 `internal/op`
- `go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=50`：失败，稳定复现取消竞态
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoFmt`：失败，且扫描范围包含 `.tools\go`
- `git diff --check`：失败，原因是 3 个前端文件 EOF 空行
- `Set-Location web; D:\gol1\node.exe .\node_modules\next\dist\bin\next build`：失败，宿主 `spawn EPERM`
- `Set-Location web; D:\gol1\node.exe .\node_modules\vitest\vitest.mjs run`：失败，宿主 `spawn EPERM`
- `docker --version`：失败，当前主机无 `docker`

## 需要特别说明的验证项

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild` 这次虽然整体返回成功，但输出包含真实 `FAIL`，因此不能再作为 Go 全绿证据使用。
- 原生 `next build` 失败，但仓库包装器 `build-web-static.mjs` 成功，说明当前更可靠的本机构建路径是仓库自带包装器而不是直接 `next build`。

## 未验证项

- Docker compose runtime smoke
- 浏览器驱动的前端行为级测试
- Linux 真机 / CI 环境之外的跨宿主行为一致性
- `manual / ai_profile` 切换对真实运行配置的生效证明

# Comparison Notes

## 当前工作区 vs `HEAD`

- 当前分支为 `feat/erguotou`，`HEAD` 为 `bfa27ae`，同时带有 tag `v0.1.3`。
- 当前工作区相对 `HEAD` 有 `133` 个已跟踪文件改动，合计约 `17406` 行新增、`3443` 行删除，另有大量未跟踪文件。
- 本次审计结论针对的是“当前工作区”，不是只针对 `HEAD` 提交本身。

## 当前分支 vs 最近稳定基线

- 相对 `origin/dev`，当前已提交分支差异约为 `72` 个文件，`6001` 行新增、`1741` 行删除。
- 这说明当前分支已经明显偏离上游开发基线；审计必须同时看 committed 差异和未提交工作区差异，不能只看 tag。

## 代码实现 vs README / docs / 任务说明

- README 与中文 README 的构建说明落后于当前 `build:static` 主线。
- [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49) 对 AI 自动化中心的状态描述已经失真。
- AI 自动化需求文档要求设置页支持 `manual -> ai_profile -> manual` 切换并保持手动配置完整保留，但当前代码只能证明 setting 被切换，不能证明 runtime 真正切换。
- 动态路由相关说明与当前实现的贴合度比前几轮更高，本次没有发现“动态路由只是展示面”的假实现问题。

## 代码实现 vs 测试 / 构建 / 验证覆盖

- 主链路已被正向证据支撑：`go build`、Windows smoke、TypeScript 和静态导出都通过。
- 但覆盖与可信度没有完全一致：CI 目标 Go 测试在当前工作区失败，而本地 Go 校验脚本又会假绿。
- 前端浏览器级 / Vitest / Docker 证据仍被宿主条件阻断，因此当前验证面更偏向 no-browser 与 smoke，而不是完整 release 验收。

# Top Next Actions

## 需要优先处理的前三项

1. 修复 Go 验证链可信度。
- 修改 [scripts/verify-go-env.ps1](D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:87)，让 `go test` / `go build` 正确传递退出码，并让 `-GoFmt` 排除 `.tools` 等 vendored 目录。

2. 收口 `internal/op` 的 bootstrap/default-setting 测试契约。
- 明确 [internal/op/cache.go](D:/GPT-codex/octopus_repo/internal/op/cache.go:9), [internal/op/user.go](D:/GPT-codex/octopus_repo/internal/op/user.go:71), [internal/op/setting.go](D:/GPT-codex/octopus_repo/internal/op/setting.go:124) 的默认初始化语义是否就是新基线；然后统一更新测试或初始化流程，并立刻复跑 CI 同款 Go 测试命令。

3. 收口 AI 自动化的两个主风险。
- 让 [internal/op/ai_automation.go](D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:357) 的取消语义同步到步骤与执行协程。
- 把 `config_source_mode / active_ai_profile_id` 真正接入 runtime consumer，或者下调 UI / 文档承诺。
- 持久化 `config_snapshot`，避免任务在进程重启后失去关键上下文。

## 建议下一步动作

- 同步更新 README、中文 README 和 Debian Dockerfile，让构建路径统一到 `build:static` 现实主线。
- 更新 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49)，去掉 “代码实现未开始” 这类陈旧表述。
- 如果要补完发布证据，需要在具备 Docker 且不会触发 `spawn EPERM` 的宿主或 CI 上补跑 Docker / Vitest / 浏览器级验证。

## 中文摘要

1. 本次触发时间
- `2026-04-25T05:39:01+08:00`

2. 做了哪些检查、运行了哪些命令
- 检查了仓库结构、Git 状态、分支、最近提交、automation memory、既有审查产物、README/docs、CI workflow、AI 自动化、动态路由、备份与核心后端入口。
- 运行了 `git status --short --branch`、`git branch -vv --all`、`git log --oneline --decorate -n 12`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、`git diff --check`。
- 运行了 `verify-go-env.ps1 -GoTest -GoBuild`、repo-local `go build ./...`、repo-local CI 同款 `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`、`go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=50`。
- 运行了 `smoke-win-backend.ps1`、`tsc --noEmit`、`build-web-static.mjs`、`next build`、`vitest run`、`verify-backup-component.cjs`、`verify-locale-consistency.mjs`、`verify-backup-logic.mjs`、`verify-dynamic-routing-help.mjs`、`verify-route-target-copy.mjs`、`verify-home-layout.mjs`、`verify-go-env.ps1 -GoFmt`、`docker --version`。

3. 修改了哪些文件
- 新增 [2026-04-25-053901-octopus-repo-complete-audit.md](D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-25-053901-octopus-repo-complete-audit.md:1)
- 新增同名 HTML 报告
- 本次无业务代码变更

4. 发现了什么问题
- 当前工作区的 CI 目标 Go 测试命令不通过，核心阻塞在 `internal/op` 的 bootstrap/default-setting 契约漂移。
- `verify-go-env.ps1` 会把失败的 `go test` 误报为成功，存在假绿风险。
- AI 自动化取消任务存在步骤状态竞态；`manual / ai_profile` 仍停留在设置层，没有真实 runtime 接线；`config_snapshot` 只保存在进程内。
- README / Dockerfile / CI 的构建口径不一致；状态文档与当前实现不一致；协议转换边界仍有 TODO；本机 Vitest / Docker / 浏览器级验证仍受环境限制。

5. 本次结果是成功、跳过还是失败
- 成功：审计已完整完成并落盘。
- 但结论是“当前工作区未通过关键验证，不能视为 release-clean”。

6. 是否需要我手动介入
- 需要。
- 优先介入点是：确定 `internal/op` 测试应跟随哪套初始化契约；决定 `manual / ai_profile` 是真 runtime 功能还是先降级为设置占位；在可用 Docker / 非 `spawn EPERM` 宿主上补足剩余验证证据。
