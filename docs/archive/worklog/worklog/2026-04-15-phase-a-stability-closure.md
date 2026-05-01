# Phase A - 工程稳定性收口

> 当前活跃阶段 worklog

---

## 1. 任务信息

- 任务名称：Phase A 工程稳定性收口
- 日期：2026-04-15
- 当前阶段：Phase A
- 对应 milestone：为 Phase B 进入里程碑 1 施工提供前置条件

## 1.1 当前阶段状态摘要

- Phase A 状态：未完成
- 当前状态：自动化稳定性基本收口，但仍阻塞中
- 唯一出阶段阻塞：核心页面真实手工 smoke 未完成
- 阻塞原因：当前线程没有附着的 app terminal / browser session，不能形成真实页面打开记录
- 下一步：先补真实页面手工冒烟记录，再判断是否进入 Phase B

## 2. 开工前输入

- 对应 canonical 章节：
  - 第 12 节工作流计划
  - 第 14 节验收标准
  - 第 16 节实施规则
- 对应 workflow 章节：
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 第 3 节
- 上一个相关 worklog：
  - `2026-04-14-backend-task-01-channel-key-modes.md`
  - `2026-04-14-backend-task-02-advanced-failover-runtime.md`
- 本次任务目标：
  - 保持仓库处于可持续编译、可持续测试、可持续推进状态
  - 固定后续阶段都会重复执行的基础验证链路
  - 列出当前高风险半成品区域
  - 先补齐 MD 规范欠账，再继续沿主线补服务器运行与 Windows 本机试跑链路
- 本次已盘点本地资源：
  - 仓库内 MD：`AGENTS.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、当前阶段与相关 worklog
  - 本地脚本与命令：`scripts/phase-a-check.ps1`、`go test ./...`、`go build ./...`、`pnpm build`
  - 当前线程上下文与既有结论：此前已完成的 relay / op / balancer 测试护栏、Phase A 阻塞点结论
  - 本地 skills：`using-superpowers`、`dispatching-parallel-agents`
- 本次是否启用子 agent 与分工边界：
  - 是
  - 子 agent 负责主规则文档中的“本地资源优先与子 Agent 协作”规则落点梳理与写入
  - 主线程负责 worklog 规范、活跃阶段记录、结果复核与冲突清理
- 本次子 agent 使用模型：
  - 规则目标：统一使用 `gpt-5.4`
  - 本轮实际情况：历史尝试中存在因模型权限或网关异常导致的失败，后续规则已补齐为优先使用 `gpt-5.4`，不可用时记录阻塞并回退主线程串行推进

## 3. 本次硬规则

- 后端构建必须通过
- 前端构建必须通过
- 当前阶段 worklog 必须完整
- 未满足稳定性前置条件时，不进入后续 Phase 主施工
- 本轮必须优先复用本地资源与既有上下文，不重复发明流程
- 遇到可并行的独立子任务时，优先使用子 agent 明确分工并回收结论

## 4. 本次禁止事项

- 不允许跳过基础构建验证直接继续叠加复杂功能
- 不允许在未记录遗留项的情况下切到下一个 Phase
- 不允许把编译错误、类型错误、明显结构错误留给后续阶段顺带修
- 不允许在未核对本地资源的情况下重复探索已确认过的结论

## 5. 本次验收条件

- `go test ./...` 可重复执行
- `go build ./...` 可重复执行
- `pnpm build` 可重复执行
- 核心页面没有明显阻塞型编译/类型问题
- 已形成基础冒烟清单
- 已形成当前高风险半成品清单
- 已把本地资源优先与子 agent 协作规则写入主文档、工作流和 worklog 规范

## 6. 本次回滚点

- 以“小步修复 + 每轮构建验证”为主
- 若某次修复引入新阻塞，回退到上一个构建通过点
- 文档层保留阶段记录，避免重复踩坑

## 7. 实现范围

- 先改数据语义还是先改 UI：先处理阻塞构建的问题，再补流程与验证
- 受影响后端模块：
  - `internal/conf/config.go`
  - `internal/server/server.go`
  - `internal/server/middleware/static.go`
- 受影响前端模块：
  - 无直接业务代码改动
- 受影响接口：
  - 无新增业务 API
- 是否影响旧数据：
  - 原则上不主动改旧数据语义
- 是否影响旧行为：
  - 仅允许修复阻塞性问题，不允许顺带重构业务行为

## 8. 基础冒烟清单

1. 后端基础命令：
   - `go test ./...`
   - `go build ./...`
   - `go run main.go --help`
2. 前端基础命令：
   - 在 `web/` 目录执行 `pnpm build`
3. 推荐统一脚本：
   - `powershell -ExecutionPolicy Bypass -File .\scripts\phase-a-check.ps1`
4. 页面冒烟清单：
   - [2026-04-15-phase-a-smoke-checklist.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-15-phase-a-smoke-checklist.md)
5. 核心关注页面：
   - 渠道页
   - 分组页
   - 模型页
   - 日志页
   - 设置页中的备份/导入区域

## 9. 当前高风险半成品清单

- advanced failover runtime 仍是 phase-1 语义
- key 管理 UI 仍偏平铺字段，不是最终子卡片形态
- 备份/导入还没有完整的 snapshot + rollback 闭环
- 首页指标只完成了部分 token breakdown
- 价格归一化只完成了基础字段与部分骨架

详细审计见：

- [2026-04-15-phase-a-risk-audit.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-15-phase-a-risk-audit.md)

## 10. 测试与验证

- 构建命令：
  - `go test ./...`
  - `go build ./...`
  - 在 `web/` 目录执行 `pnpm build`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\phase-a-check.ps1`
- 专项验证：
  - 后续每次进入新阶段前，先重复执行上面三条
  - 第一批后端路由测试护栏见：
    - [2026-04-15-backend-task-routing-test-guardrails.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-15-backend-task-routing-test-guardrails.md)
  - `internal/op/channel.go`、`internal/op/group.go`、`internal/relay/balancer/circuit.go`、`internal/relay/balancer/session.go`、`internal/relay/relay.go` 与 `runRaceFallback` 的关键测试护栏已补齐；本轮又补上 `attempts` 并发编号连续性、`race_concurrency` 上限边界、race fallback 成功后入站响应转换失败时不能误记成功的运行时修复，以及 race 提前返回、最高优先级竞速候选成功时快速收敛、probe 失败 attempts 记录、取消探测不污染 failed 语义的边界护栏
  - 2026-04-15 本轮复核再次确认：go test ./... 通过，powershell -ExecutionPolicy Bypass -File .\scripts\phase-a-check.ps1 通过
  - 2026-04-16 主线继续推进：修复 race fallback 成功结果在入站响应转换失败时误判为成功的问题，并新增对应测试；`go test ./internal/relay ./internal/relay/balancer` 通过，`go test ./...` 通过`r`n  - 2026-04-16 继续补强 race fallback 边界：新增“最高优先级竞速候选成功时快速返回”与“race probe 真实失败会落 attempts”测试，并完成对应运行时收敛；`go test ./internal/relay ./internal/relay/balancer` 通过，`go test ./...` 通过`r`n  - 2026-04-16 再次收紧运行时：当前 race 池最高优先级候选一旦成功即立刻返回，不再被更慢的低优先级 probe 拖住；`go test ./internal/relay ./internal/relay/balancer` 通过，`go test ./...` 通过`r`n  - 2026-04-16 继续贴近主干语义：被提前取消的低优先级 race probe 不再污染 failed 语义，而是按调度性放弃处理；`go test ./internal/relay ./internal/relay/balancer` 通过，`go test ./...` 通过
- 手工冒烟状态：
  - 页面清单已固定，但本线程当前没有附着的 app terminal / browser session
  - `read_thread_terminal` 返回无 attached app terminal session
  - 因此尚不能在本线程内形成真实页面打开结果，不应伪造“已通过”的手工冒烟记录
  - 当前仅完成自动化验证与运行链路补强，不等同于页面级 smoke 完成

## 11. 风险与兼容性

- 新风险：
  - 仓库当前仍是脏工作区，后续实施时要避免误改非当前任务区域
  - 子 agent 曾出现模型权限不可用的情况，主线程已接管剩余文档改动并记录该阻塞
- 兼容性风险：
  - 若在 Phase A 中顺带改动业务语义，容易掩盖真正的稳定性问题
- 是否阻塞下一任务：
  - 是，自动化稳定性已基本收口，但手工页面冒烟记录未完成前，不进入 Phase B 主施工

## 12. 收工记录

- 构建是否通过：
  - `go test ./...`：通过
  - `go build ./...`：通过
  - `go run main.go --help`：通过
  - `pnpm build`：通过（需在 `web/` 目录执行）
  - `scripts/phase-a-check.ps1`：通过
  - 2026-04-15 本轮复核：`go test ./...` 再次通过，`scripts/phase-a-check.ps1` 再次通过
- 测试是否通过：
  - 当前自动化验证链路通过
  - 本轮新增 relay 运行时修复对应测试通过
  - 本轮新增 race fallback 提前返回与 probe 失败记录对应测试通过
  - 本轮新增“最高优先级竞速候选成功时立即返回”的运行时收敛验证通过
  - 本轮新增“取消的低优先级 race probe 不污染 failed 语义”的细节验证通过
  - 本轮新增静态资源运行链路测试通过：`go test ./internal/server/...`
  - 本轮新增本机运行链路验证通过：`powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1 -CheckOnly`
  - 本轮新增本机运行链路验证通过：`powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1 -CheckOnly`
  - 本轮完成 Windows 构建实测：`powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1`
- 本次使用了哪些本地资源 / skills / 记忆上下文：
  - 使用了 `AGENTS.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/worklog/README.zh-CN.md`、`docs/worklog/WORKLOG_TEMPLATE.zh-CN.md`、当前活跃阶段与相关 worklog 作为主上下文
  - 使用了本地 skills `using-superpowers`、`dispatching-parallel-agents` 约束本轮先看本地资源、再并行拆分独立文档任务
  - 使用了此前已验证的 Phase A 自动化构建与测试结论，避免重复推翻已确认事实
  - 使用了 `README.md`、`README_zh.md`、`scripts/build.sh`、`scripts/phase-a-check.ps1`、`static/static.go`、`internal/server/server.go`、`internal/server/middleware/static.go` 作为运行链路与部署方式的本地事实来源
- 本次使用了哪些子 agent 及其结论：
  - 使用过多名子 agent 负责主规则文档落点与规则写入，最终有效结论一致：应把“本地资源优先与子 Agent 协作”写入 `AGENTS.md` 与 `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - 个别子 agent 因模型权限不可用而失败，主线程已接管剩余 worklog 规范与活跃记录写入，并保留有效结论、丢弃失败尝试
  - 本轮使用 `gpt-5.4` 子 agent 辅助复核 MD 缺口与 Windows 脚本方向；主线程吸收其有效结论后，独立完成规则落盘、Go 侧运行链路实现和本机脚本验证
- 手工 smoke 状态：
  - 未完成
- 手工 smoke 阻塞原因 / 缺少的环境：
  - 缺少附着的 app terminal / browser session
  - 当前线程无法在不伪造结果的前提下完成页面级人工打开与交互检查
- 待验证页面清单：
  - Home
  - Channel
  - Group
  - Model
  - Log
  - Setting
- worklog 是否更新：是
- 遗留项：
  - 需要按页面冒烟清单补一轮真实手工检查记录
  - 当前线程没有附着页面会话，无法在不伪造结果的前提下完成 Home、Channel、Group、Model、Log、Setting 六个页面的人工冒烟
  - `internal/op/channel.go`、`internal/op/group.go`、`internal/relay/balancer/circuit.go`、`internal/relay/balancer/session.go`、`internal/relay/relay.go` 与 `runRaceFallback` 的关键测试护栏已补齐；本轮又补上 `attempts` 并发编号连续性、`race_concurrency` 上限边界、race fallback 成功后入站响应转换失败时不能误记成功的运行时修复，以及 race 提前返回、最高优先级竞速候选成功时快速收敛、probe 失败 attempts 记录、取消探测不污染 failed 语义的边界护栏；当前自动化测试侧剩余项已主要收敛到更复杂的 attempts 顺序与并发边界细节
- 下一任务前置条件是否满足：
  - 当前已完成基础构建链路验证，并补上第一批路由、熔断/会话、业务层与关键 relay 运行时测试护栏，也已补齐本地资源与子 agent 协作规则文档
  - 但尚未完成核心页面手工冒烟记录，因此不满足无缺口进入 Phase B 的前置条件




