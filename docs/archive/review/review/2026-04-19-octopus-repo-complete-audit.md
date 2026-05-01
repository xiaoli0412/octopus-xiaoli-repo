# Octopus Repo 完整审计

触发时间：2026-04-19T09:17:16+08:00

审计范围：当前工作区 `D:\GPT-codex\octopus_repo`

本次优先复用了以下本地资源：`AGENTS.md`、`README.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/worklog/*`、`.github/workflows/validation.yaml`、现有脚本目录 `scripts/`、仓库内新增测试文件清单、当前线程上下文。

本次未使用子 agent。原因：当前会话未获得显式委派授权，且本轮审计需要在同一证据链上连续比对工作区、文档、路由、任务调度与验证脚本。

## 1. Findings

### Critical

- `internal/server/handlers/setting.go:105` 引用了未定义符号 `includeSecrets`，当前工作区内 `rg -n "\bincludeSecrets\b" -S` 仅命中这一处，没有任何定义或解析逻辑。这会直接导致后端编译失败，`/api/v1/setting/export` 所在 handlers 包无法通过构建，属于发布阻塞项。相关证据：`internal/server/handlers/setting.go:101-106`，`internal/op/backup.go:65` 已将 `DBExportAll` 签名改为四参，但 handler 侧未补齐入参来源。

### High

- `stats_save_interval` 设置对运行中的定时任务不生效，只写库不热更新。前端系统设置页明确暴露了该字段并在失焦时保存：`web/src/components/modules/setting/System.tsx:17-19`、`web/src/components/modules/setting/System.tsx:124-128`。但后端 `setSetting` 只对 `model_info_update_interval` 和 `sync_llm_interval` 调用了 `task.Update(...)`：`internal/server/handlers/setting.go:82-97`。实际 `stats_save` 任务只在启动时注册一次：`internal/task/init.go:52-59`。这意味着用户修改后 UI 显示成功，但统计持久化周期直到重启前都不会改变。

- `DynamicRoutingAdjustmentTask` 名称和文档语义是“动态调节”，实现却只是每日扫描并写入一份内存摘要，没有改动任何路由、阈值、缓存或持久化配置，属于疑似空实现/伪实现。注册点在 `internal/task/init.go:40-41`，实现主体 `internal/task/dynamic.go:44-113` 仅统计 channel/group/key 数量并调用 `setDynamicRoutingAdjustmentSummary(...)`。`Basis` 甚至显式写成 `daily_runtime_scan_dynamic_thresholds_only`：`internal/task/dynamic.go:44-48`。这与 README/worklog 中“动态健康调节”“策略收口”的表述不一致。

- 动态路由与竞速预算相关设置已经进入后端模型默认值，但前端设置常量和设置页没有接线，用户无法通过管理台配置这些关键开关。后端定义了 `dynamic_routing_health_enabled`、`race_global_budget`、`race_group_budget`、`race_channel_budget`、`race_key_budget`、`race_probe_budget`：`internal/model/setting.go:12-25`、`internal/model/setting.go:37-51`。前端 `SettingKey` 只暴露到 `api_base_url`：`web/src/api/endpoints/setting.ts:14-26`；设置页只渲染 `System/Log/LLMSync/CircuitBreaker/Backup` 等模块：`web/src/components/modules/setting/index.tsx:19-28`。结果是“代码支持、默认值存在、文档强调”，但用户侧不可配置。

### Medium

- `healthcheck` CLI 默认会把配置中的监听地址直接当作客户端请求地址使用：`cmd/healthcheck.go:21-24`。默认配置里 `server.host` 是 `0.0.0.0`：`internal/conf/config.go:106-109`。监听地址 `0.0.0.0` 适合服务端绑定，不适合作为客户端探活目标；在默认配置下，`octopus healthcheck` 很可能对着 `http://0.0.0.0:8080/healthz` 探测，造成“服务正常但命令误报失败”的运维误导。

- 前端设置面板对熔断设置的轮询刷新会覆盖用户未保存的编辑，行为与 `SettingSystem` 的保护逻辑不一致。`SettingSystem` 通过 `initialized` 和“后续轮询只更新 ref 不更新 state”避免覆盖本地草稿：`web/src/components/modules/setting/System.tsx:25-49`。但 `SettingCircuitBreaker` 每次 `settings` 刷新都 `queueMicrotask(() => setThreshold/setCooldown/setMaxCooldown)`：`web/src/components/modules/setting/CircuitBreaker.tsx:22-36`。由于 `useSettingList` 每 30 秒轮询一次，用户输入中的未保存值可能被后台轮询重置。

### Low

- 验证链与文档宣称之间仍有缺口。README 和 CI 工作流宣称覆盖了 `go test ./internal/op`、前端类型检查、Linux smoke、Docker compose smoke：`README.md:158-165`、`.github/workflows/validation.yaml:41-65`。但当前显式的 Go 单测步骤只覆盖 `internal/op`，而本次发现的 `internal/server/handlers/setting.go` 编译阻塞说明“非 `internal/op` 的后端包”依赖 smoke 阶段才能兜底。验证链并非无缺口，只是把问题延后到更重的步骤暴露。

## 2. Completion Assessment

整体判断：

- 按“功能面完成度”看，当前工作区大约在 `75%` 左右。核心服务启动链路、基础 Relay 主流程、导入/回滚后端主链、路由目标覆盖、首页 token breakdown 与 probe/circuit 摘要都已经真实接线，不是纯文档状态。
- 按“可发布完成度”看，当前工作区更接近 `55%`。原因不是功能完全没做，而是存在明确编译阻塞、配置生效不完整、动态调节存在疑似空实现，以及本地无法完成完整构建验收。

已完成：

- 启动链路、配置加载、静态资源本地/嵌入回退、`/healthz` 暴露已经真实接线。
- 导入/回滚主链路已经不仅是 UI 骨架，后端存在 dry-run、preview token、snapshot 列表、回滚接口与 `op` 层实现。
- Route target override 已从 model/op/handler 接到前端 channel 页面，不属于悬空 API。
- 首页 token breakdown 已接上统计接口，并补充了 probe/circuit 摘要信息。

部分完成：

- 动态路由/高级 failover 有真实策略代码，但“动态 adjustment task”本身并未完成到可称为 runtime adjustment 的程度。
- 设置体系后端字段扩展明显快于前端配置入口，造成“可用默认值存在，但不可运营配置”。
- 验证链在仓库层面有脚本和 CI 草案，但当前环境下无法闭环，且代码侧存在可以直接阻塞编译的问题。

未完成或疑似空实现：

- `DynamicRoutingAdjustmentTask` 更像状态扫描，不像动态调节。
- 动态路由预算/健康开关的前端运营入口缺失。
- 当前工作区尚不能证明通过完整构建、类型检查、Go 测试与 smoke。

## 3. Verification Summary

已验证项：

- 仓库结构、当前分支、工作区状态、最近提交、`origin/dev` 基线差异。
- 核心文档与 worklog：`README.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/worklog/README.zh-CN.md` 及近期 worklog。
- 启动主链：`main.go` -> `cmd/start.go` -> `db.InitDB/op.InitCache/server.Start/task.Init`。
- 配置加载与静态资源回退：`internal/conf/config.go`、`internal/server/server.go`。
- Relay 动态调优路径：`internal/relay/relay.go`、`internal/relay/dynamic.go`、`internal/op/route_target.go`。
- 备份/导入/回滚路径：`internal/server/handlers/setting.go`、`internal/op/backup.go`、`internal/model/backup.go`。
- 统计展示路径：`internal/server/handlers/stats.go`、`web/src/api/endpoints/stats.ts`、`web/src/components/modules/home/token-breakdown.tsx`。
- Route target API 与前端调用链：`internal/server/handlers/route_target.go`、`web/src/api/endpoints/channel.ts`。
- 任务调度与热更新接线：`internal/task/task.go`、`internal/task/init.go`、`internal/server/handlers/setting.go`。
- CI/验证声明：`.github/workflows/validation.yaml`。

本次执行过的关键命令：

- `Get-ChildItem -Force`
- `git status --short --branch`
- `git log --oneline -n 12`
- `git branch -a --verbose --no-abbrev`
- `git diff --stat HEAD`
- `git diff --stat origin/dev...HEAD`
- `git diff --name-only HEAD`
- `git diff --name-only origin/dev...HEAD`
- `git diff --unified=40 HEAD -- ...`
- 多轮 `rg -n ...` 与 `Get-Content ...`
- `node -v`
- `corepack pnpm -v`
- `corepack pnpm exec tsc --noEmit`
- `go version`
- `docker --version`

未验证项：

- `go build ./...`、`go test ./...`、`go test ./internal/op -count=1` 未能在本机完成。原因：当前环境 `go` 不在 PATH；探测到的 `C:\Users\李昊桐\AppData\Local\Programs\go\bin\go.exe` 运行时报 `Access is denied`。
- 前端 `pnpm exec tsc --noEmit` 未能真正执行类型检查。原因：当前环境虽然可以通过 `corepack` 拿到 pnpm 版本，但 `tsc` 命令不可执行，返回 `'tsc' is not recognized as an internal or external command`。
- Docker/compose smoke 未执行。原因：当前环境 `docker` 不在 PATH。

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 当前工作区不是小修，而是大规模未提交状态。`git diff --stat HEAD` 显示 `65 files changed, 8234 insertions(+), 1071 deletions(-)`，同时还有大量未跟踪测试、迁移、脚本与文档文件。
- 这意味着本次审计对象必须是“工作区中的待提交整体”，不能只按 `HEAD` 做 review。

当前分支与主分支/稳定基线：

- 当前分支是 `feat/erguotou`，远端存在 `origin/dev` 作为最近稳定基线。
- `origin/dev...HEAD` 显示该分支已经包含 Copilot、Antigravity、providers、DocModal、README/截图等一批已提交能力；当前工作区又在此基础上继续推进 Phase D/F/G 的动态路由、备份导入、统计展示和脚本化验证。

代码实现与 README / docs / 任务说明一致性：

- 一致的部分：导入/回滚、token breakdown、route target override、静态资源回退、validation workflow 都能在代码中找到真实接线。
- 不一致的部分：README/worklog 对“动态调节”“Milestone 3 runtime”表述强于代码实际；当前 `DynamicRoutingAdjustmentTask` 并没有做真正的调节动作。
- 另一个不一致点是设置可运营性：后端已有动态路由预算与健康开关，但前端设置台没有对等入口。

代码实现与测试/构建/验证覆盖一致性：

- 仓库中新增了较多测试文件，尤其覆盖 `internal/op`、`internal/server/handlers`、`internal/relay`、`internal/task`。这说明作者有补测试意识。
- 但测试存在并不等于当前工作区已可发布。本次发现的 `includeSecrets` 未定义问题位于 handler 层，如果只看 `go test ./internal/op` 无法兜住。
- CI 工作流在文档层面是完整的，但本地环境无法复跑，且前端类型检查在当前环境中没有真正启动。

## 5. Top Next Actions

优先处理前三项：

1. 先修复后端编译阻塞：为 `exportDB` 明确定义 `include_secrets` 入参来源，或回退这次签名改动，随后立即执行 `go build ./...` 与最小 smoke。
2. 修复“配置保存成功但运行时不生效”的问题：至少补上 `stats_save_interval` 的 `task.Update(...)`，并明确哪些设置支持热更新、哪些必须重启。
3. 对动态路由收口做一个明确决策：要么把 `DynamicRoutingAdjustmentTask` 实现成真正的运行时调节；要么改名为 summary scan，并同步收缩 README/worklog 说法与测试命名，避免继续制造“已实现”错觉。

需要修改的东西与要求总结：

- 修复 `internal/server/handlers/setting.go` 的 `includeSecrets` 未定义问题，并补相应 handler/编译级测试。
- 让 `stats_save_interval`、以及所有宣称可在线调整的 setting，具备真实热更新行为；做不到的要在 UI 明示“需重启后生效”。
- 补齐动态路由健康开关与竞速预算的前端设置入口，或者在 README/docs 中下调宣称范围。
- 重新定义 `DynamicRoutingAdjustmentTask` 的职责与验收标准，避免 summary-only 代码继续以 runtime adjustment 名义存在。
- 在可执行环境中补跑完整验证链：`go build ./...`、`go test ./internal/op -count=1`、前端 `pnpm install --frozen-lockfile && pnpm exec tsc --noEmit`、`scripts/smoke-linux-backend.sh`、`scripts/smoke-docker-compose.sh`。

---

## 中文摘要

1. 本次触发时间

- `2026-04-19T09:17:16+08:00`

2. 做了哪些检查、运行了哪些命令

- 检查了仓库结构、`git status`、最近提交、`origin/dev` 基线差异、README/docs/worklog。
- 沿着启动链、配置链、Relay 主流程、动态路由、备份导入回滚、统计展示、前端设置页、CI 工作流做了源码级核查。
- 实际执行了 `git status --short --branch`、`git log --oneline -n 12`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、多轮 `rg -n`、多轮 `Get-Content`、`node -v`、`corepack pnpm -v`、`corepack pnpm exec tsc --noEmit`、`go version`、`docker --version`。

3. 修改了哪些文件

- 新增：`docs/review/2026-04-19-octopus-repo-complete-audit.md`

4. 发现了什么问题

- 发现一个明确的后端编译阻塞：`internal/server/handlers/setting.go` 使用了未定义的 `includeSecrets`。
- 发现一个明确的配置生效问题：`stats_save_interval` 在 UI 可保存，但运行中的定时任务不会更新。
- 发现一个疑似空实现：`DynamicRoutingAdjustmentTask` 只有扫描摘要，没有实际调节动作。
- 发现后端已支持的动态路由预算/健康设置没有前端入口。
- 发现 `healthcheck` 默认目标地址采用 `0.0.0.0`，存在误报风险。

5. 本次结果是成功、跳过还是失败

- `成功`。审计已完成并落盘，但构建/测试验证因环境限制未能闭环。

6. 是否需要我手动介入

- `需要`。
- 需要先提供可运行的 Go / pnpm / Docker 环境，或允许在可用环境中继续执行构建与 smoke。
- 在代码层面，需要优先人工确认 `include_secrets` 的预期产品语义，以及动态路由 adjustment 是要真正实现还是要缩减宣称。
