# 2026-04-24 Phase G Runtime Default Stop Policy Closure

## 1. 任务信息

- 任务名称：Windows 本机运行态默认停驻策略与精确停服入口收口
- 日期：2026-04-24
- 当前阶段：Phase G 截图优先 UI 主线的运行态支撑层
- 对应 milestone：自动化链路可持续接管与本机环境不常驻治理

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 13、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、11.4 节
- 上一个相关 worklog：`docs/worklog/2026-04-16-phase-g-runtime-entrypoints.md`
- 本次任务目标：在不恢复常驻服务的前提下，补齐 `octopus_repo` 专属运行态管理脚本，并把“默认停驻、按需启动、验证结束回收”的口径同步写入主文档与环境文档。
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/worklog/README.zh-CN.md`
  - `docs/worklog/WORKLOG_TEMPLATE.zh-CN.md`
  - `docs/worklog/2026-04-16-phase-g-runtime-entrypoints.md`
  - `scripts/dev-win.ps1`
  - `scripts/smoke-win-backend.ps1`
  - `cmd/healthcheck.go`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档、现有 Windows 开发脚本、healthcheck 入口、当前线程上下文与上一轮已完成的“精确停掉本项目后台进程”操作记录。
- 若未使用部分本地资源或上下文，原因：本轮只补本机运行态治理，不涉及前端组件结构与后端业务逻辑。
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：不适用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：是，由 Codex 主线程直接落地
- 若使用分目录 agent，负责目录与禁止越界范围：不适用
- 若未使用子 agent，原因：任务范围小、上下文耦合强、直接由主线程串行完成更稳。

## 3. 本次硬规则

- 只能管理 `D:\GPT-codex\octopus_repo` 相关运行态进程，不得误伤其他项目。
- 不重新把项目拉回常驻运行状态。
- 文档、脚本和当前用户口径必须同步，不允许只加脚本不写入口规则。

## 4. 本次禁止事项

- 不新增无关业务功能。
- 不改动现有浏览器 smoke 的 self-start / external 主逻辑。
- 不使用模糊停进程方式清理整机 Node/Go 进程。

## 5. 本次验收条件

- `scripts/runtime-win.ps1` 可执行 `check-only / status / stop`。
- 相关主文档已写入“默认停驻、按需启动、验证结束回收”的统一口径。
- worklog 已记录本轮结果和后续自动化使用方法。

## 6. 本次回滚点

- `scripts/runtime-win.ps1`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改 UI，不改业务数据语义，优先补运行态治理脚本与文档口径。
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：仅复用 `go run main.go healthcheck`
- 是否影响旧数据：否
- 是否影响旧行为：仅统一本机运行态管理方式，不改变业务接口行为。

## 8. 实施步骤

1. 新增 `scripts/runtime-win.ps1`，支持 `status / stop / healthcheck / check-only`。
2. 把默认停驻与按需 self-start / external 的策略写入用户总账、详细工作流、环境文档和当前状态文档。
3. 运行脚本验证并记录本轮结果。

## 9. 测试与验证

- 构建命令：无新增构建
- 测试命令：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop`
- 专项验证：确认脚本输出只指向 `D:\GPT-codex\octopus_repo` 运行态与自动化入口。

## 10. 风险与兼容性

- 新风险：若未来本机运行端口调整，`runtime-win.ps1` 默认端口列表可能需要补充。
- 兼容性风险：脚本依赖 Windows PowerShell 与 `Get-NetTCPConnection` / `Get-CimInstance`，不适用于 Linux。
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：本轮无新增构建
- 测试是否通过：`check-only / status / stop` 已通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主工作流文档、用户总账文档、环境文档、当前状态文档、既有 Windows 脚本、healthcheck 入口、当前线程上下文
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户总账与工作流确认了“三个自动化链路必须可调用”，且当前用户明确要求先不要让项目继续在本机后台常驻。
  - 既有 `dev-win.ps1` 和 smoke wrapper 说明仓库已经有按需 self-start 思路，但缺少一个统一的“状态检查 / 精确停服”入口。
  - `cmd/healthcheck.go` 提供了无需常驻脚本改造的探活入口，可直接复用。
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：不适用
- 手工 smoke 状态：未启动浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮目标是停驻治理，不需要恢复运行态页面联调
- 待验证页面清单：后续如需真实页面验证，继续沿用仓库现有 `self-start / external` smoke wrapper
- 若未使用子 agent，原因：任务范围明确且较小，主线程串行更稳
- worklog 是否更新：是
- 遗留项：如后续需要，可再给 Linux 侧补一份同语义的 runtime status/stop 包装脚本
- 下一任务前置条件是否满足：是
