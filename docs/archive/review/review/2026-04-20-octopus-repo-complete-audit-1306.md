# Octopus Repo 完整审计

触发时间：2026-04-20T13:06:01+08:00

审计范围：当前工作区 `D:\GPT-codex\octopus_repo`

本次优先复用的本地资源：`AGENTS.md`、`README.md`、`README_zh.md`、`docs/review/2026-04-20-octopus-repo-complete-audit-0216.md`、`docs/worklog/*`、`.github/workflows/validation.yaml`、`scripts/*`、当前工作区 git 状态与最近提交。

## 1. Findings

### Critical

- `web/src/api/endpoints/setting.ts:441-449` 会在每次前端导出时无条件附带 `include_secrets=false|true`，而 `web/src/components/modules/setting/Backup.tsx:604-606,1532-1535` 又把导出开关默认设为 `false` 并直接传给 `useExportDB()`。这与后端 `internal/server/handlers/setting.go:108-118` 和 README 声明的“省略 `include_secrets` 时默认导出可直接恢复的全量快照”发生了真实偏差。结果是管理 UI 的默认导出行为并不会走后端的默认全量分支，而是稳定地产出脱敏快照；如果操作者按文档理解默认导出可直接恢复，这条备份链会在真正恢复时暴露风险。这不是文案小问题，而是备份主路径语义已经分叉。

### High

- `scripts/smoke-win-backend.ps1:14-16` 存在脚本拼接错误：`use-go-env.ps1` 的点号导入和 `$exePath = ...` 写在同一行，形成 `. (Join-Path $scriptDir 'use-go-env.ps1')$exePath = ...`。这会让仓库自己提供的 Windows 本地 smoke 验证脚本在进入真实 smoke 流程前就处于损坏状态。README 明确给了 Windows 脚本链路，但当前仓库无法用这条路径证明自己。
- `.github/workflows/validation.yaml:46-55` 当前 Go 侧只跑 `go build ./...` 和 `go test ./internal/op -count=1`。这和当前工作区新增风险最密集的 `internal/server/handlers`、`internal/relay`、`internal/task` 并不匹配。仓库现在已经有大量这几个目录下的新增测试和新增逻辑，但 CI 仍不会把它们当成必须通过的发布门槛，关键路径回归有明显漏检窗口。

### Medium

- `scripts/use-go-env.ps1:39-43` 在配置环境后会立刻执行 `go.exe version`。在本机 Go 可执行文件返回 `Access is denied` 的情况下，这会让 `scripts/verify-go-env.ps1`、`scripts/smoke-win-backend.ps1` 这类仓库自带脚本全部在预热阶段直接失败，无法继续提供更细粒度诊断。本轮无法确认这是不是仓库唯一原因，但可以确认它放大了环境阻塞，降低了脚本自诊断价值。
- 前端生产构建仍未完成闭环验证。实际运行 `D:\gol1\node.exe .\node_modules\next\dist\bin\next build` 时，Next.js 16 已完成编译和 TypeScript 阶段，但最终以 `spawn EPERM` 结束。结合 `web/next.config.ts:3-6` 使用 `output: 'export'` 可判断当前源码至少已走到静态导出收尾阶段；不过由于构建没有成功落盘，本轮仍不能把前端发布产物标记为已验证完成。
- `web/src/components/modules/setting/Backup.tsx:2168-2172` 在页面内直接声明 `Still not wired to backend`，其中明确列出 `Full conflict-handling wizard`、`Interactive rollback history browser`、`Route preview rich diff tooling` 仍未接线。这说明备份/导入大功能已不再是空壳，但仍有用户可见的半成品区域，完成度不能按“全部闭环”计算。

### Low

- 当前工作区相对 `HEAD` 仍是大体量脏状态。`git diff --stat HEAD` 显示 71 个已跟踪文件变更，另有大量未跟踪测试、脚本、迁移和文档文件。它不会直接证明逻辑错误，但会显著抬高回归定位成本，也会让“本轮问题来自哪一批改动”更难隔离。

## 2. Completion Assessment

整体判断：

- 按功能落地度看，当前工作区约为 `75% - 80%`。主启动链、基础管理 API、导入/导出/回滚主链、动态路由设置项、路由预算消耗、日志导出、统计增强，都已经不是 README 式空实现。
- 按“可发布并可被证明”看，更接近 `60% - 65%`。主要短板不是“完全没写”，而是导出语义偏差、Windows smoke 自损坏、Go/CI 验证链不足、前端生产构建未闭环。

已完成：

- 主流程真实可达：`main.go` -> `cmd/start.go` -> `db.InitDB` -> `op.InitCache` -> `server.Start` -> `task.Init` -> `task.RUN`。
- 动态路由设置不是空壳：`web/src/components/modules/setting/index.tsx:20-30` 挂载了 `SettingDynamicRouting`，而 `internal/relay/budget.go:51-92`、`internal/relay/dynamic.go:19-37` 确实消费了这些设置。
- 导入/导出/回滚相关 handler 已真实注册并接入 `router.RegisterAll(...)`，不是游离函数。
- 新增 `healthcheck` CLI 已接入 Cobra 根命令。

部分完成：

- 备份/导入能力已经有真实后端和较完整前端，但默认导出语义与文档和后端默认分支不一致。
- Windows / Linux / Docker smoke 链路都有脚本和文档，但只有 Linux/Docker 脚本文本被审计，Windows 脚本当前还存在语法级损坏。
- 前端构建已通过编译和类型检查阶段，但导出落盘未成功。

未完成：

- 完整 Go 验证链 `go build ./...`、`go test ./...`、Windows smoke、Linux smoke、Docker smoke 本轮都未能在当前环境完成闭环。
- 备份页面内自述的冲突向导、丰富 diff、细粒度 rollback 浏览器仍未接线。

疑似空实现：

- 本轮没有发现“界面已做但后端完全未消费”的强证据，尤其动态路由相关旧疑点本轮已排除。
- 仍可判定为“半成品而非空实现”的区域是 `web/src/components/modules/setting/Backup.tsx` 内明确标注未接线的高级交互功能。

## 3. Verification Summary

已验证项：

- `git status --short --branch`
- `git branch -a --verbose --no-abbrev`
- `git log --oneline -n 12`
- `git rev-list --left-right --count origin/dev...HEAD`，结果为 `0 behind / 22 ahead`
- `git diff --stat HEAD`
- `git diff --stat origin/dev...HEAD`
- 核心入口与接线文件：`cmd/start.go`、`internal/server/server.go`、`internal/server/router/router.go`、`internal/server/handlers/setting.go`、`internal/task/init.go`、`internal/task/task.go`、`internal/task/dynamic.go`、`internal/relay/budget.go`、`internal/relay/dynamic.go`
- 文档一致性核对：`README.md`、`README_zh.md`、`docs/review/2026-04-20-octopus-repo-complete-audit-0216.md`
- 前端类型检查：`D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit`，通过
- 前端生产构建尝试：`D:\gol1\node.exe .\node_modules\next\dist\bin\next build`，编译和 TS 阶段通过，最终 `spawn EPERM`
- Go 脚本链路尝试：`powershell.exe -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild`，被 `go.exe` 的 `Access is denied` 阻断

未验证项：

- `go build ./...`
- `go test ./...`
- `go test ./internal/op -count=1` 的当前工作区重跑结果
- `scripts/smoke-win-backend.ps1` 的真实 smoke 结果
- `scripts/smoke-linux-backend.sh`
- `scripts/smoke-docker-compose.sh`

无法运行的具体原因：

- 本机 `C:\Users\李昊桐\AppData\Local\Programs\go\bin\go.exe` 可定位到，但执行返回 `Access is denied`
- 当前会话默认 PowerShell 执行策略禁止直接运行 `.ps1`，需显式 `ExecutionPolicy Bypass`
- 当前环境没有可用 `docker`
- `next build` 在当前环境最终被 `spawn EPERM` 卡住

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 当前工作区不是小修，而是大批量未提交集合。已跟踪改动 71 个文件，未跟踪文件覆盖测试、迁移、脚本、文档和临时补丁。
- 最近工作区增量主要集中在备份/导入/回滚、动态路由、路由预算、统计视图、日志导出和前端设置页增强。

当前分支与稳定基线：

- 当前分支是 `feat/erguotou`。
- 相对 `origin/dev`，当前分支是 `22 ahead / 0 behind`。
- 说明这不是“落后上游导致的问题”，而是本分支新增功能与验证链收口不足的问题。

代码实现与 README / docs / 任务说明一致性：

- 一致的部分：动态路由仅有 daily summary scan、不回写阈值；健康检查 CLI 已真实存在；validation workflow 和 smoke 脚本文件确实已入仓。
- 不一致的部分：README 声称省略 `include_secrets` 时默认导出可恢复快照，但当前 UI 默认不会省略该参数，而是主动发送 `false`；README/脚本宣称的 Windows smoke 路径当前自身有语法损坏。

代码实现与测试 / 构建 / 验证覆盖一致性：

- 一致的部分：前端至少能通过 `tsc --noEmit`，说明类型层面没有明显断裂。
- 不一致的部分：CI 只覆盖 `./internal/op`，和当前新增最密集的 handler / relay / task 区域不匹配；Windows smoke 文档存在，但对应脚本本身损坏；前端 build 文档存在，但本机无法闭环证明构建成功。

## 5. Top Next Actions

需要优先处理的前三项：

1. 修复前端导出参数语义，让“默认导出可恢复快照”的行为在 UI 和后端保持一致；最小改法是未勾选明文凭证时不要强制发送 `include_secrets=false`，或者同步下调文档承诺。
2. 修复 `scripts/smoke-win-backend.ps1` 的语法损坏，并把它纳入一条可实际执行的 Windows 验证链，避免 README 继续引用坏脚本。
3. 扩大 `.github/workflows/validation.yaml` 的 Go 覆盖面，至少补上 `internal/server/handlers`、`internal/relay`、`internal/task`，否则本分支新增主风险面不会被 CI 兜住。

建议下一步动作：

- 先修备份导出语义偏差，这是最接近真实业务后果的问题。
- 再修 Windows smoke 脚本和 CI 覆盖，把“仓库声称可验证”变成“仓库真的能验证”。
- 之后在可执行 Go 与正常 spawn 子进程的环境里补跑完整 Go/前端/compose 验证链，重新确认发布状态。

---

## 中文摘要

1. 本次触发时间

- `2026-04-20T13:06:01+08:00`

2. 做了哪些检查、运行了哪些命令

- 检查了仓库结构、`git status`、最近提交、`origin/dev` 对比、现有审查记录和核心 README/docs
- 核查了主启动链、设置导出链、动态路由设置接线、预算消耗、Windows/Linux/Docker smoke 脚本、CI workflow
- 实际运行了 `D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit`
- 实际运行了 `D:\gol1\node.exe .\node_modules\next\dist\bin\next build`
- 实际运行了 `powershell.exe -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild`

3. 修改了哪些文件；如果没有修改，明确写“本次无变更”

- 新增：`docs/review/2026-04-20-octopus-repo-complete-audit-1306.md`
- 新增：`docs/review/2026-04-20-octopus-repo-complete-audit-1306.html`
- 更新：`C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`

4. 发现了什么问题

- 发现前端默认导出行为与 README/后端默认导出契约不一致，默认 UI 导出会稳定生成脱敏快照
- 发现 `scripts/smoke-win-backend.ps1` 存在语法拼接错误，Windows smoke 路径当前不可用
- 发现 CI 仍只覆盖 `./internal/op`，没有兜住 handler / relay / task 风险面
- 确认前端类型检查可过，但 `next build` 仍在当前环境报 `spawn EPERM`
- 确认 Go 工具链在当前机器上执行即返回 `Access is denied`

5. 本次结果是成功、跳过还是失败

- `成功`

6. 是否需要我手动介入

- `需要`
- 需要优先决定是否修正备份导出默认语义，并提供可执行 Go / Docker / 正常 spawn 子进程的环境，以补完完整验证链

Runtime this run: about 17 minutes.