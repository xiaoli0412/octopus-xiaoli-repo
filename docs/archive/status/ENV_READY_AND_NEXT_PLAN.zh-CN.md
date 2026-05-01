> 执行硬规则：后续任何实现、补丁、回归修复、UI 调整、路由策略修改、接口扩展与验收动作，都必须先对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)。
> 若当前文档与主规划冲突，以主规划为准；若实现需要偏离主规划，必须先更新主规划，再改代码，再补验证记录。
> 当前环境说明与下一步计划还必须同步对齐 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 中的当前优先级、自动化链路要求和 `Codex` 执行口径。

# 环境前提完成报告与下一步计划（2026-04-22）

## 0. 当前会话补充（2026-04-22）

虽然本文件记录的是环境前提，但当前会话新增了必须同步的执行前置条件：

- 环境准备完成不等于可以直接编码，必须先确认主规划和用户上下文总账已对齐。
- 开工前需要在 worklog 填写 `Master plan aligned before coding (yes/no):`。
- 如果任务准备交给自动化 agent 或分目录 agent，必须先定义负责目录和越界限制。
- 后续执行入口统一按 `Codex` 口径，不再引用错误的 `OpenCode` 施工描述。
- 以上前置条件不仅适用于当前线程，也适用于 `octopus`、`octopus-2`、`octopus-repo` 三条自动化链路。
- 当前本机默认运行策略统一为“项目默认停驻，按需启动，验证结束即回收”；如需查看或回收 `octopus_repo` 本地运行态，统一使用 `scripts/runtime-win.ps1`。

## 0.1 当前统一 next plan（应用到所有自动化）

当前不是重新发散路线图，而是收敛到所有自动化共享的近端顺序：

1. `P0`：统一 workflow、plan、memory 的继承规则，确保三条自动化都先读同一套主文档与当前优先级。
2. `P1`：优先处理图片问题池中的 UI、文案、交互、帮助提示与熔断设置问题。
3. `P2`：推进备份导入导出兼容、旧快照样本支持与跨项目迁移适配。
4. `P3`：继续推进其余 canonical 主线任务。
5. `P4`：进入 `AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习` 主线；该主线需先完成 H1 文档对齐，再实施后端模型、API、前端栏目和动态路由学习。

本轮之后，所有自动化的 memory 和 worklog 都应沿用这个顺序，而不是继续停留在旧的单线 Phase 判断。

---

## 1. 已完成的环境前提任务

### 1.1 工具链可用性

- `git`: 已可用
- `node`: `v24.14.1`
- `npm`: `11.11.0`
- `pnpm`: `10.18.3`（已激活并可直接调用）
- `go`: `go1.26.2 windows/amd64`（已安装并可直接调用）
- `gofmt`: 已可直接调用

### 1.2 国内网络可用性配置

为避免依赖下载超时，已完成以下配置：

- Go:
  - `GOPROXY=https://goproxy.cn,direct`
  - `GOSUMDB=sum.golang.google.cn`
- pnpm/npm:
  - `registry=https://registry.npmmirror.com`

### 1.3 项目依赖安装

- 后端依赖：`go mod download` 已完成
- 前端依赖：`pnpm install`（`web/`）已完成

### 1.4 可执行验证

- 后端：
  - `go test ./...` 通过
  - `go build ./...` 通过
  - `go run main.go --help` 可正常输出命令帮助
- 前端：
  - `pnpm build`（`web/`）通过

## 2. 为保证可编译而修复的阻塞项

在环境验证阶段发现并修复了几处会直接阻塞编译/构建的问题：

- [log.ts](/D:/GPT-codex/octopus_repo/web/src/api/endpoints/log.ts)  
  修复 `exportLogs` 参数类型与 `apiClient.raw` 不兼容问题。
- [token-breakdown.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/home/token-breakdown.tsx)  
  修复 `formatCount` 返回结构读取错误（`raw/formatted` 字段）。
- [ItemOverlays.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/model/ItemOverlays.tsx)  
  统一 `EditValues` 类型并对齐 `BillingMode/ProbePolicy`。
- [Item.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/model/Item.tsx)  
  `useState` 显式使用共享 `EditValues`，消除 `onChange` 类型冲突。

## 3. 当前剩余注意事项

- 前端构建有 `baseline-browser-mapping` 过期提醒，属于非阻塞告警，不影响当前可构建与可开发。
- 仓库当前为脏工作区（历史改动较多），后续建议分批收敛，避免一次性大规模混改。
- `docker` / `docker compose` 当前不可用；不影响代码开发与本地构建，但会影响容器化部署验证。
- 当前本机不应让 `octopus_repo` 前后端长期常驻；浏览器 smoke 优先走仓库现有 `self-start / external` wrapper，收工后再回到停驻状态。

## 3.1 Windows 本机运行态入口

- 状态检查：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 精确停服：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop`
- 运行态健康检查：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action healthcheck`
- 入口检查：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
- 低权限主机上，`status` 会在 `CIM` 被拒绝时自动回退到 `Get-Process + Get-NetTCPConnection`；如果当前会话里仍拿不到 owning process，则继续尝试 `netstat -ano -p tcp`，并在输出中标注当前扫描模式。新口径下，`status` 通常能先给出端口 owner PID / 进程名提示，再由主线程判断这些监听是否属于外部程序；`stop` 仍只针对已归因到 `octopus_repo` 的进程，不会因为端口提示扩大停服范围。

## 4. 下一步执行计划（可直接开工）

### 4.0 当前阶段判断

- 主线：LLM Gateway 重构 + 图片问题优先返工主线
- 当前阶段：跨自动化执行口径统一完成后，进入图片问题池优先修复窗口
- 当前小闭环目标：先让所有自动化共享同一套 workflow/plan/memory，再按图片问题顺序持续推进代码修复

## Phase P0：跨自动化 workflow/plan/memory 统一（优先级 P0）

- 目标：让三条自动化都先读同一套文档、使用同一优先级、写回同一类 memory/worklog 结构。
- 任务：
  - 更新共享 workflow 文档，明确所有自动化都要先读的文件和开工前置条件。
  - 更新当前 next plan 与相关状态文档，使图片问题池先于旧阶段泛化计划。
  - 更新三个 automation memory，并留下统一的下一轮入口。
- 验收：
  - 三条 automation memory 都出现同一套当前优先顺序与前置规则。
  - workflow、plan、worklog 三类载体读回一致。

## Phase P1：图片问题池优先修复（优先级 P1）

- 目标：先收敛最新截图已暴露的问题，减少 UI 主线返工和后续混乱。
- 任务：
  - 优先处理版本信息卡片、首页统计卡片、创建分组弹窗、创建渠道弹窗、`Route Target Overrides`、`CC Switch`。
  - 同步清理中文 i18n key 泄漏、中英混杂、帮助提示缺失与熔断设置解释不足问题。
  - 把探测与检测入口从价格区域归位到设置页。
- 验收：
  - 对应问题都有“现象、根因、风险、修复动作、自动化验收点”。
  - 图片问题池优先项形成至少一个已验证闭环。
  - `CC Switch` 已在本机形成一条可复跑的浏览器级闭环：默认使用 `scripts/verify-ccswitch-browser-smoke.ps1 -Mode self-start`，其 CDP bootstrap 顺序固定为 `runtime-page-lifecycle`。
  - 渠道总页也已在本机形成一条可复跑的浏览器级闭环：默认使用 `scripts/verify-channel-page-browser-smoke.ps1 -Mode self-start`，其 `channel-page` 场景会验证组合筛选、详情弹层、key 展开与 `375px` 页面可读性。

## Phase P2：备份/导入与迁移兼容收口（优先级 P2）

- 目标：在图片问题池之后，继续收口旧快照兼容、dry-run、apply、rollback 和跨项目导入适配。
- 任务：
  - 用原始备份样本继续验证导入兼容链路。
  - 收口导出 manifest 与真实 secrets 语义一致性。
  - 继续补足映射、重绑、跳过与导入后验证闭环。
- 验收：
  - 原始样本至少能 dry-run 通过并给出可执行结果。
  - 备份页验证链与最小运行态验证保持绿色。

## Phase A：稳定性收口（当前降为并行关注项）

- 目标：把“可编译”升级为“可持续迭代”。
- 任务：
  - 跑一轮关键页面手工回归（渠道、分组、模型、日志、备份）。
  - 补充最小化冒烟清单（后端 API + 前端核心页面）。
  - 修复回归阻塞项并保持 `go build` + `pnpm build` 持续通过。
- 验收：
  - 冒烟清单全部通过。
  - 两端构建连续通过 2 次。

## Phase B：里程碑 1 未闭环项（优先级 P0）

- 目标：完成渠道/分组故障转移关键闭环。
- 任务：
  - Key 管理 UI 从“平铺字段”收敛到“清晰分层结构”。
  - 高级故障转移运行时从 phase-1 语义补到计划中的最终策略行为。
  - 对应接口参数补齐校验与错误提示。
- 验收：
  - 文档中的里程碑 1 验收条目可逐条打钩。

## Phase C：里程碑 2 可观测性闭环（优先级 P1）

- 目标：首页/日志/统计形成完整闭环。
- 任务：
  - 首页补齐 provider/source/probe/breaker 等缺失视图。
  - 日志导出、token 分解、过滤条件做一致性回归。
  - 指标字段和后端返回结构做一次统一核对。
- 验收：
  - 首页关键指标与日志明细可相互校验。

## Phase D：备份/导入与价格体系（优先级 P1）

- 目标：把“骨架能力”补成“可安全使用”。
- 任务：
  - 备份/导入补齐 replace/merge/skip/map、回滚、冲突报告与导入后校验。
  - 价格体系补齐 alias/manual mapping/fallback 细节与 UI 编辑闭环。
- 验收：
  - 使用真实样本可完成一次“备份 -> 导入 -> 校验 -> 回滚”演练。

## 5. 当前候选任务顺序

1. 继续从图片问题池中选择一个最小闭环 UI 问题优先修复并验证，优先处理剩余中文界面英文泄漏、渠道 `9.1.1` 更细 key 级字段，或 page-level 之后的更细浏览器交互缺口。
2. 同步把最新页面级闭环写回状态文档、worklog 与 automation memory，避免计划和代码事实再次漂移。
3. 继续推进备份导入导出兼容和样本验证。
4. 在以上两条主线稳定后，再回到 canonical 的 Phase A/B/C 收口任务。
5. 进入 Phase H：AI 自动化中心与动态路由 AI 学习。第一步必须确认新增 AI 自动化中英文文档、需求 54-64、动态路由学习口径和 worklog 已对齐。

## 6. Phase H 环境与验证补充

Phase H 开工前需要确认：

- 文档入口存在：`docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_REQUIREMENTS.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.md`。
- 主线文档已包含 `AI 自动化`、`AI Profile`、`dynamic_routing_learning_enabled`、`不覆盖用户配置`。
- 动态路由文档不再把本地在线学习列为非目标。

Phase H 实施后的建议验证命令：

1. `go test ./internal/model ./internal/op ./internal/server/handlers ./internal/relay/...`
2. `pnpm --dir web exec tsc --noEmit`
3. `node scripts/verify-locale-consistency.mjs`
4. 后续新增 AI 自动化页面组件测试。
5. 后续新增设置页 `manual / ai_profile` 切换测试。
6. 后续新增动态路由学习开关与不写回 priority 测试。
