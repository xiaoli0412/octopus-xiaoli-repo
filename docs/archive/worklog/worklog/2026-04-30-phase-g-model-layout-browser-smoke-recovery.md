# 2026-04-30 Phase G Model Layout Browser Smoke Recovery

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / model layout browser evidence`
- Current stage: `模型页运行态 browser smoke 收口`

## 本轮上下文与本地资源

- canonical / 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 需求清单：`docs/PLAN.md`
- 主规划：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 当前状态：`docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- 相邻 worklog：`docs/archive/worklog/worklog/2026-04-24-phase-g-model-layout-browser-smoke-closure.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 canonical plan 下的增量重构与验证收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求“不要创建子agent，靠主线程解决”。

## 本轮候选任务

1. 重跑模型页 `self-start` browser smoke，定位当前真实断点。
2. 若断点在共享 CDP wrapper / smoke 逻辑，做最小修复并回归模型页。
3. 复核模型页 no-browser / type / check-only 合同仍为绿色。
4. 若主任务提前闭环，再考虑相邻 Phase G 页面级验证缺口。

## 本轮计划

- 本轮核心任务：修复阻塞模型页 browser smoke 的共享 CDP smoke 失败判定。
- 本轮配套任务：回归模型页 `no-browser + tsc + check-only + self-start smoke`。
- 预期验证方式：
  - `node .\scripts\verify-llm-price-boundary.mjs`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- 完成判定标准：模型页 `self-start` smoke 通过，或至少输出稳定且可行动的 bootstrap blocker 分类，不再停留在泛化 websocket 断链报错。

## 本轮硬规则

- 只在 `Phase G` 主线内推进，不扩散到无关 UI 改造。
- 写入范围控制在模型页相关共享 smoke 脚本与必要记录。
- 不回退用户已有改动，不清理无关脏区。

## 本轮禁止事项

- 不做模型页视觉重设计。
- 不做共享 wrapper 的大范围重写。
- 不顺手处理 Phase H 或备份链路问题。

## 本轮回滚点

- 仅回滚 `scripts/verify-channel-create-browser-smoke-cdp.mjs` 本轮补丁即可撤销行为变化。

## 执行与结果

1. 先跑模型页直接相关验证，确认 `verify-llm-price-boundary.mjs`、`tsc --noEmit` 与 `check-only` 均通过，断点仅剩 `self-start` browser smoke。
2. 重跑 `self-start` 后，确认失败并不在模型页 DOM 契约，而是在共享 `scripts/verify-channel-create-browser-smoke-cdp.mjs` 中：CDP bootstrap/probe 已经给出页面会话不可用证据，但主流程仍继续执行 `setCdpViewport`，把根因冲淡成泛化的 `CDP websocket is not connected`。
3. 对共享 Node smoke 做最小修复：新增 bootstrap-unavailable 识别函数，统一处理 `timed out / websocket closed / not connected / session closed / target closed` 这类 page bootstrap 不可用信号；在 `bootstrap` 阶段和 `probe` 阶段都提前中止并生成 `CdpPageBootstrapUnavailableError`，避免继续执行后续 viewport / 页面操作。
4. 修复后重跑模型页 `self-start` smoke，确认模型页 browser smoke 重新通过。

## 验证

- passed `node .\scripts\verify-llm-price-boundary.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only`
- observed expected pre-fix failure from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- passed post-fix `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- passed `git diff --check -- scripts\verify-channel-create-browser-smoke-cdp.mjs docs\archive\worklog\worklog\2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`

## 本轮变更文件

- `scripts/verify-channel-create-browser-smoke-cdp.mjs`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`

## 未完成 / 风险 / 阻塞

- 本轮没有改模型页静态 UI 代码，因此未触发 `build:static` / `web/out -> static/out` 同步动作；后续若再改页面源码，仍需按 `UI_MAINLINE_TASK_2026-04-30` 的要求同步运行态产物。
- 共享 CDP smoke 当前 host 已对模型页恢复可用，但其它页面若继续依赖同一 bootstrap 逻辑，仍应优先复用这次的失败分类口径，避免再次退化成泛化 websocket 错误。

## 下一轮候选任务顺序

1. 延续 `Phase G`，转向下一个页面级 browser smoke 缺口，优先首页或同一截图优先池中的相邻页面。
2. 若模型页后续再出现 smoke 波动，先复用本轮共享 bootstrap 判定口径排查，不要重新从 DOM 契约开始怀疑。
3. 只有当页面源码再次变化时，才补 `build:static` 与运行态静态产物同步验证。
