# 2026-04-30 Phase G Group Create Browser And Advanced Strategy Recovery

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / group create dialog browser evidence`
- Current stage: `group create 弹窗浏览器 smoke 与高级策略结构收口`

## 本轮上下文与本地资源

- canonical / 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 需求清单：`docs/PLAN.md`
- 主规划：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 当前状态：`docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-22-phase-g-group-create-default-path-tightening.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作约束核对，不进入设计门禁；本轮属于既有 canonical plan 下的增量重构与验证收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求“不要创建子agent，靠主线程解决”。

## 本轮候选任务

1. 复现 `group-create` browser smoke 当前缺口，确认是页面契约还是共享 smoke 基建问题。
2. 若 browser smoke 链路异常，先做最小 smoke 基建修复，保证本页不再误选 Codex 内置 Node。
3. 若 no-browser 合同与页面实现冲突，优先让代码回到当前 Phase G 文档口径，而不是放宽门禁。

## 本轮计划

- 本轮核心任务：恢复 `group-create` 弹窗的高级策略与 flow/testid 结构，并修复 browser smoke Node 解析漂移。
- 本轮配套任务：回归 `group-create-flow`、`tsc`、`check-only` 与真实 `self-start` browser smoke。
- 预期验证方式：
  - `node .\scripts\verify-group-create-flow.mjs`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- 完成判定标准：
  - `group-create-flow` 重新通过。
  - `group-create` `self-start` browser smoke 通过。
  - 高级策略字段重新接回创建/编辑提交链，不再只留 API 字段和 locale 而 UI 丢失。

## 本轮硬规则

- 只在 `Phase G` 主线内推进，不扩散到无关 UI 改造。
- 写入范围控制在 group 模块、对应 smoke 脚本与必要记录。
- 不回退用户已有改动，不清理无关脏区。

## 本轮禁止事项

- 不改 group 后端语义与数据库结构。
- 不顺手重做其它页面的 smoke。
- 不通过放宽 `verify-group-create-flow.mjs` 来掩盖 UI 回退。

## 本轮回滚点

- `scripts/verify-group-create-browser-smoke.ps1`
- `scripts/verify-group-create-browser-smoke-cdp.ps1`
- `web/src/components/modules/group/Editor.tsx`
- `web/src/components/modules/group/Create.tsx`
- `web/src/components/modules/group/Card.tsx`

## 执行与结果

1. 先跑 `group-create` `check-only`，发现该链路会误选 `C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe`，而同池的 `channel-create` 已经规避到外部 Node。
2. 对 `scripts/verify-group-create-browser-smoke.ps1` 和 `scripts/verify-group-create-browser-smoke-cdp.ps1` 做最小修复：复用 `channel-create` 已验证过的 Node 选择策略，排除 Codex 内置 Node，并补齐 `NODEEXE / NODE_BIN / LOCALAPPDATA Programs\nodejs` 等候选源。
3. 修复后重跑 `group-create` `self-start` browser smoke，确认浏览器级链路通过，说明本轮不是页面 DOM 回归，而是 smoke 入口解析漂移。
4. 随后 `verify-group-create-flow.mjs` 暴露了更关键的一致性问题：当前 [`GroupEditor.tsx`](D:/GPT-codex/octopus_repo/web/src/components/modules/group/Editor.tsx) 已经丢失高级策略折叠区、flow card、稳定 `data-testid` 与部分 fallback 文案，而这些结构仍被 Phase G 文档、locale 和 group API 字段承诺。
5. 因此没有放宽 verifier，而是把 `GroupEditor` 拉回当前主线口径：
  - 恢复 `flow` 摘要卡和 `naming / mode / models` 三步引导。
  - 恢复 `advanced-strategy` 折叠区，并重新接回 `retry_rounds / retry_delay_ms / failover_window_sec / race_after_fails / race_concurrency`。
  - 恢复 `group-create-dialog`、`new-group-*` 等稳定 `data-testid`。
  - 恢复模型筛选、空态提示、右侧选择摘要与加权态提示。
  - 让创建/编辑链路重新提交上述高级策略字段，并统一 `channelFallbackName` 文案回退。
6. 回归后，`group-create-flow`、`tsc`、`check-only` 与真实 `self-start` browser smoke 全部重新通过。

## 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- passed `node .\scripts\verify-group-create-flow.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- passed `git diff --check -- scripts/verify-group-create-browser-smoke.ps1 scripts/verify-group-create-browser-smoke-cdp.ps1 web/src/components/modules/group/Editor.tsx web/src/components/modules/group/Create.tsx web/src/components/modules/group/Card.tsx scripts/verify-group-create-flow.mjs`

## 本轮变更文件

- `scripts/verify-group-create-browser-smoke.ps1`
- `scripts/verify-group-create-browser-smoke-cdp.ps1`
- `web/src/components/modules/group/Editor.tsx`
- `web/src/components/modules/group/Create.tsx`
- `web/src/components/modules/group/Card.tsx`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-browser-and-advanced-strategy-recovery.md`

## 未完成 / 风险 / 阻塞

- `git diff --check` 仅剩 LF/CRLF 警告，不构成当前主线阻塞。
- 本轮没有变更 `web/out` / `static/out` 产物；如果下一轮继续改 group 页面视觉或布局源码，仍需按 `UI_MAINLINE_TASK_2026-04-30` 要求补静态产物同步。
- 当前 group 主线已恢复到文档口径，但其它相邻页面仍可能存在类似“API/locale 仍在，UI 结构回退”的风险，下一轮宜继续走相邻页面级 smoke 而不是重新扫全仓。

## 下一轮候选任务顺序

1. 延续 `Phase G`，转向相邻页面级 smoke 缺口，优先 `channel-create` 更细弹窗交互或首页/设置页同池细节证据。
2. 若继续 group 主线，优先补 `375px / hover / focus` 之外的更细弹窗手感问题，而不是再回头修字段链。
3. 若其它页面再次出现 smoke 漂移，先复用本轮 Node 解析与 verifier 对齐的经验排查，不要直接放宽门禁。
