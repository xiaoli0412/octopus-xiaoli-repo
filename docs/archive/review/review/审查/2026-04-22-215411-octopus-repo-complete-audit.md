# Octopus Repo Complete Audit

Triggered at: `2026-04-22T21:54:11.1740612+08:00`
Workspace: `D:\GPT-codex\octopus_repo`
Branch: `feat/erguotou`
HEAD: `bfa27ae` (`v0.1.3`)
Comparison baseline: `origin/dev` (`0 behind / 22 ahead`)
Automation ID: `octopus-repo`

## 1. Findings

### Critical

1. `internal/relay/relay.go` 的真实请求主链路仍未消费这轮新增的“同渠道多 key 回退、多轮重试、failover window”语义，导致当前工作区出现真实行为回归，而不是单纯测试超前。

   证据：主链路只在 `internal/relay/relay.go:74` 做 `for iter.Next()`，在 `internal/relay/relay.go:102` 为当前 `GroupItem` 选取一次 `channel.GetChannelKeyForModel(targetModel)`，然后在 `internal/relay/relay.go:146` 执行一次 `ra.attempt()` 后成功就返回、失败就切下一个 group candidate。这里没有任何对 `GetChannelKeyForModelExcept`、`NextEligibleChannelKeyAfter`、`RetryRounds`、`RetryDelayMs`、`FailoverWindowSec` 的运行时消费。结果就是 `go test ./... -count=1` 在 `internal/relay/relay_more_test.go:516`、`internal/relay/relay_more_test.go:671`、`internal/relay/relay_more_test.go:739`、`internal/relay/relay_more_test.go:824`、`internal/relay/relay_more_test.go:876` 这 5 个 handler 级用例上全部失败，直接表现为期望 `200` 的链路返回 `502`、期望“failover window exceeded”的错误信息没有出现、期望“stale route item”被跳过却只记录成“no available key”。

2. 动态竞速 fallback 看起来“已经实现”，但主流程根本没有接进去，属于典型的“实现存在但主流程不可达”。

   证据：`internal/relay/dynamic_runtime.go:69` 提供了 `shouldEscalateToRace`，`internal/relay/dynamic_runtime.go:208` 提供了 `runRaceFallback`，`internal/relay/dynamic_runtime.go:370` 提供了 `waitForRetryDelay`，`internal/relay/dynamic.go` 还会基于 `ResolveRouteTargetPolicy(...)` 计算 race/circuit tuning。但全仓搜索结果表明，这些 helper 只在测试和 helper 层被引用，没有从 `internal/relay/relay.go` 主路径调用。与此同时，分组模型、设置页和动态路由相关代码都已经暴露了这套能力，让功能表面上像是 ready，实际上 live path 完全到不了这一层。

### High

3. 分组高级故障转移配置已经打通到模型、后端更新逻辑和前端表单，但运行时没有兑现这些配置，属于“配置真实存在、行为未兑现”的高风险不一致。

   证据：`internal/model/group.go:19-24`、`internal/model/group.go:44-49` 已定义 `RetryRounds`、`RetryDelayMs`、`FailoverWindowSec`、`RaceAfterFails`、`RaceConcurrency`；`internal/op/group.go:141-191` 也会持久化这些字段；前端 `web/src/components/modules/group/Editor.tsx` 已经有对应状态和交互（如 `28-37`、`819+` 附近）。但 `internal/relay/relay.go` 主路径仍没有消费这些字段，当前只能选一个 key、尝试一次，再切 group item。也就是说，用户现在能够在 UI 中保存这些策略，但网关主链路并不会按这些策略执行。

4. Windows smoke 已经不再错误复用旧二进制，这是进步，但它仍依赖外部注入 repo-local Go cache/temp 环境变量，否则默认执行会因为宿主默认 `go-build` 目录不可写而失败，验证链的可移植性仍不足。

   证据：默认执行 `powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 时失败，错误是 `C:\Users\李昊桐\AppData\Local\go-build\... Access is denied`；脚本中的 `scripts/smoke-win-backend.ps1:105` 开始构建，`scripts/smoke-win-backend.ps1:109` 已经明确拒绝复用旧二进制，这部分逻辑是正确的。随后我在外部显式注入 `GOCACHE`、`GOTMPDIR`、`TEMP`、`TMP` 到仓库内可写目录后，同一脚本完整通过。这说明脚本逻辑本身可用，但它还没有自我修复默认 Windows 权限环境。

5. 前端生产构建在本机仍无法完成最终验证，`next build` 虽然已经完成编译和 TypeScript 检查，但最终仍被 `spawn EPERM` 卡住，发布态前端产物无法在本机最终证明。

   证据：`Push-Location web; node .\node_modules\next\dist\bin\next build` 的结果是“Compiled successfully”并完成 “Running TypeScript ...”，之后仍报 `Error: spawn EPERM`。这意味着当前代码层面已经不再有前端语法/类型阻塞，但本机生产构建收口仍被环境级权限问题挡住。

### Medium

6. 顶层发布文档仍未收口，README/README_zh 依然带有占位符和 TODO，无法直接作为对外发布文档使用。

   证据：`README.md:45`、`README.md:52-54`、`README.md:71-73`、`README.md:88-89` 以及 `README_zh.md:45`、`README_zh.md:52-54`、`README_zh.md:70-72`、`README_zh.md:87-88` 仍包含 `<CURRENT_IMAGE>`、`<CURRENT_REPOSITORY_URL>`、`<CURRENT_ONE_CLICK_INSTALL_SCRIPT>`、`<CURRENT_RELEASES_URL>` 和 TODO 占位文本。

7. 动态路由文档仍明显超前于真实实现，文档里存在 `shadow-ai`、`hybrid`、`metrics-only`、`incident-safe` 等模式，但当前后端真实落地的只是“daily summary scan”。

   证据：`docs/DYNAMIC_ROUTING_REQUIREMENTS.md:218-222`、`docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:233-237`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.md:212-224`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:227-239` 仍描述这些模式；而真实后端只实现了 `internal/task/dynamic.go:13` 的每日摘要任务，其 `Basis` 明写为 `daily_summary_scan_no_runtime_mutation`（`internal/task/dynamic.go:47`），并在健康开关关闭时只返回 `dynamic routing health disabled; summary scan skipped`（`internal/task/dynamic.go:62`），任务注册也只是在 `internal/task/init.go:41` 注册 summary scan。

### Low

8. `web/package.json` 仍带着硬编码的非仓库可移植 `devs` HTTPS 本地路径，开发脚本可移植性不足。

   证据：`web/package.json:8` 仍写死 `/workplace/code/octopus/build/localhost-key.pem` 和 `/workplace/code/octopus/build/localhost.pem`。

## 2. Completion Assessment

总体判断：仓库并不是空壳，主干管理能力和网关最小主流程都是真实可达的；但当前工作区仍不是可发布状态，主要阻塞点已经从“编不过”收敛成“`internal/relay` 真实行为回归 + 生产构建/验证环境收口不足”。

完成度估计：

- 业务主链完成度约 `80%~85%`。
- 当前工作区发布完成度约 `65%~70%`。

已完成：

- 后端启动、`/healthz`、静态资源本地目录/内嵌回退真实可达，`internal/server/server.go:37-41,65-76` 与 Windows smoke 结果相互印证。
- 设置页导出/导入/回滚接口是真实接线，不是占位：`internal/server/handlers/setting.go:32-64` 已注册 `/export`、`/import`、`/import-snapshots`、`/rollback-*` 路由，`include_secrets` 也在 `internal/server/handlers/setting.go:122` 真正解析。
- Route Target Override API 已真实接入：`internal/server/handlers/route_target.go:36-49` 注册了 list/upsert/delete，`internal/op/route_target.go` 也有缓存与 override 解析逻辑。
- 备份页两条自动化验证链真实可运行：`verify-backup-logic.mjs`、`verify-backup-component.cjs` 本轮均通过。
- 动态路由设置页不是空 UI：`web/src/components/modules/setting/DynamicRouting.tsx:232,289,310` 能展示 summary、budget summary 和 advanced panel。
- CI 验证工作流比前几轮更完整，已经包含前端备份验证与 `internal/relay/...` 的 Go 测试：`.github/workflows/validation.yaml:48-61,79`。

部分完成：

- 高级 failover / race 配置链路已经贯通到 model、op、frontend，但 live relay 还没兑现。
- 动态路由 summary surface 已经接线，但 runtime mutation / race escalation 主路径仍未接入。
- 前端编译与类型检查已通过，本机生产 build 最后一跳仍被 `spawn EPERM` 卡住。

未完成：

- README / README_zh 的发布收口。
- 动态路由文档宣称的多模式收口。
- Windows smoke/build 脚本默认自适应 repo-local cache。
- Docker 相关验证在本机不可执行。

疑似半实现：

- `internal/relay/dynamic_runtime.go` 整套 race helper。
- 分组高级故障转移字段在 UI/后端已可配置，但 runtime 未消费。

## 3. Verification Summary

已验证项：

- `go build ./...`
  结果：通过。
  说明：默认 Go cache 目录无权限时会失败；切到 repo-local `GOCACHE/GOTMPDIR` 并复用本机已有 module cache、关闭联网解析后通过。

- `go test ./... -count=1`
  结果：失败。
  说明：失败集中在 `internal/relay`，其余 `internal/op`、`internal/server/handlers`、`internal/server/middleware`、`internal/task`、transformer 子包均通过。

- `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  结果：通过。

- `node .\scripts\verify-backup-logic.mjs`
  结果：通过。

- `node .\scripts\verify-backup-component.cjs`
  结果：通过。

- `node .\scripts\sync-web-static.mjs`
  结果：通过。

- `powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`
  结果：默认环境失败；在显式设置 repo-local `GOCACHE/GOTMPDIR/TEMP/TMP` 后通过。

- `node .\web\node_modules\next\dist\bin\next build`
  结果：编译与 TypeScript 通过，但最终失败于 `spawn EPERM`。

- `pnpm -v`
  结果：本机 `PATH` 中不可用。

- `docker --version`
  结果：本机 `PATH` 中不可用。

本轮最关键的验证结论：

- `go build` 现在已经恢复，不再是当前 workspace 的主要阻塞。
- 当前更真实的阻塞是 `internal/relay` 的 handler 级行为回归。
- 备份/导入主线与静态资源同步主线本轮验证是绿色的。

未验证项：

- Linux smoke：本机没有直接可用的 Linux 运行环境。
- Docker compose smoke：本机 `docker` 不在 `PATH`，无法执行。
- 成功的本机 `next build` 产物：被 `spawn EPERM` 卡住，无法完成最终证明。
- 完整 Vitest 泛化回归：本轮只执行了仓库内明确指定的两条备份验证链，没有补跑全部 Vitest 用例集。

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 当前工作区相对 `HEAD` 有 `91` 个已跟踪文件改动，`12551` 行新增、`2319` 行删除，另有约 `18251` 个未跟踪文件/目录项。
- 未跟踪数量明显被 `.next/`、`node_modules/`、`.tools/`、构建产物等放大；真正需要重点看的仍是 `internal/relay`、`internal/model`、`internal/op`、`internal/server`、`web/src`、`scripts`、`docs`。
- 本轮最严重的行为回归来自当前 workspace 漂移，而不是 `v0.1.3` 已提交基线本身。

当前分支与稳定基线：

- 仓库没有本地 `main` 分支；本轮使用 `origin/dev` 作为稳定对比基线。
- `git rev-list --left-right --count origin/dev...HEAD` 结果为 `0 22`，说明当前分支不落后于 `origin/dev`，并领先 `22` 个提交。
- 当前分支最近可描述 tag 仍是 `v0.1.3`，但真正的审计风险已经转移到未提交工作区增量上。

代码实现与 README / docs / 任务说明：

- 一致：设置导出默认 `include_secrets` 语义、route-target override API、动态路由 summary scan 的真实语义、Windows smoke 主路径、健康检查和静态资源回退。
- 不一致：README 发布说明仍未填充；动态路由文档仍宣称多模式策略；分组高级 failover/race UI 与 runtime 实际能力不一致。

代码实现与测试 / 构建 / 验证覆盖：

- 一致：CI workflow 现在已经把 `build:static`、备份验证脚本、`go build ./...`、`go test ./internal/relay/...` 纳入验证链，这能覆盖本轮最真实的风险。
- 不一致：本机 `next build` 仍然不能完成最终收口；Windows smoke 需要外部注入 repo-local cache 才能稳定通过；Docker smoke 在本机无法执行。

## 5. Top Next Actions

需要优先处理的前三项：

1. 修复 `internal/relay/relay.go` 的 live path，使其真正消费“同渠道多 key fallback / RetryRounds / RetryDelayMs / FailoverWindowSec”语义，或者在未实现前回退这些暴露面，先让 `internal/relay` 的 handler 级测试恢复为绿色。
2. 对动态竞速能力做一次范围决策：如果这一轮要发版，就必须把 `effectiveDynamicRoutingTuning`、`shouldEscalateToRace`、`runRaceFallback` 从 helper 层接进主流程；如果这一轮不发，就应缩减 UI/docs/测试表述，避免继续制造“看起来实现了”的错觉。
3. 固化验证环境：把 repo-local `GOCACHE/GOTMPDIR/TEMP/TMP` 直接写入 Windows smoke/build 入口，另外在可用环境上完成一次成功的 `next build` 与 Docker smoke，消除当前发布前最后两条环境阻塞。

建议下一步动作：

- 先修 `internal/relay`，不要再优先扩新配置面。
- 修完后优先复跑 `go test ./internal/relay -count=1`、`go test ./... -count=1`、`smoke-win-backend.ps1`。
- 如果要准备对外发布，再补 README/README_zh 的真实镜像、仓库地址、release 下载链接。

## Chinese Summary

1. 本次触发时间
   `2026-04-22T21:54:11.1740612+08:00`

2. 做了哪些检查、运行了哪些命令
   先阅读了 automation memory、git 状态、最近审查报告、README/README_zh、validation workflow、关键入口与 relay/backup/route-target/dynamic 代码，再执行了 `go build ./...`、`go test ./... -count=1`、`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\verify-backup-logic.mjs`、`node .\scripts\verify-backup-component.cjs`、`node .\scripts\sync-web-static.mjs`、`powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`、`node .\web\node_modules\next\dist\bin\next build`，并对比了当前工作区与 `HEAD`、当前分支与 `origin/dev`。

3. 修改了哪些文件
   新增 `docs/review/审查/2026-04-22-215411-octopus-repo-complete-audit.md`。

4. 发现了什么问题
   当前最严重的问题不再是“编不过”，而是 `internal/relay` 的真实行为回归：同渠道多 key fallback、多轮重试、failover window、route-target 驱动的竞速 helper 都还没有真正进入 live handler path，导致 `go test ./...` 在 `internal/relay` 上出现多条失败；其次是 `next build` 仍被 `spawn EPERM` 卡住，以及 Windows smoke 默认环境仍依赖外部注入 repo-local Go cache 才能稳定通过。README/动态路由文档也还有发布与实现不一致的问题。

5. 本次结果是成功、跳过还是失败
   成功。审计任务已完成并产出新报告，但结论是“当前工作区仍有高优先级阻塞项，不能直接视为可发布”。

6. 是否需要我手动介入
   需要。需要优先决定“高级 failover/race 是否本轮发版”，并在可执行 `next build` / Docker smoke 的环境上完成最后的发布验证；如果你希望我下一轮直接进入修复，我建议先从 `internal/relay` live path 开始。
