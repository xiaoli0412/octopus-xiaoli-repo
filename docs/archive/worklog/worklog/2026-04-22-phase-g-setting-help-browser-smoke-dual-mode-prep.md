# 2026-04-22 Phase G Setting Help Browser Smoke Dual-Mode Prep

## 1. 任务信息

- 任务名称：设置页帮助提示浏览器 smoke 双模式入口收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：截图问题池 / 设置页渐进帮助与浏览器验收入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、1.4 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-circuit-breaker-progressive-help-closure.md`
  - `docs/worklog/2026-04-22-phase-g-dynamic-routing-progressive-help-closure.md`
- 本次任务目标：
  - 把设置页帮助提示浏览器 smoke 脚本收口成 `check-only / external / self-start` 双模式入口
  - 修正自启前端时缺少 `NEXT_PUBLIC_API_BASE_URL` 的脚本缺口
  - 给 `ModelProbe` 补稳定 DOM 锚点，并把设置页三块卡片统一纳入同一 smoke 入口
  - 同步前端主线状态、worklog 和 automation memory，让下一轮可直接继续浏览器验收
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/src/components/modules/setting/ModelProbe.tsx`
  - `scripts/verify-dynamic-routing-help.mjs`
  - `scripts/verify-circuit-breaker-help.mjs`
  - `scripts/verify-model-probe-help.mjs`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `web/src/api/client.ts`
  - `web/src/api/endpoints/user.ts`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、手工验收清单、上轮 automation memory、设置页现有组件与验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只收口设置页帮助提示的浏览器入口，不扩散到备份、渠道或分组页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程串行推进，且本轮范围小、耦合高

## 3. 本次硬规则

- 只处理设置页 `DynamicRouting / CircuitBreaker / ModelProbe` 的帮助提示浏览器验收入口，不扩散到无关页面
- 不把未实际跑通的浏览器级 smoke 写成已通过
- 保持 patch-based 收口，只留下可复跑、可接续的验证入口

## 4. 本次禁止事项

- 不改后端设置、熔断、动态路由或模型探测的真实运行时语义
- 不伪造外部服务、浏览器会话或桌面 / `375px` 验收结果
- 不回退当前工作区里与本轮无关的已有修改

## 5. 本次验收条件

- `scripts/verify-setting-help-browser-smoke.mjs` 支持 `--check-only`、`--external`、`--self-start`
- 自启前端会显式注入 `NEXT_PUBLIC_API_BASE_URL`
- `ModelProbe` 有稳定的 `data-testid`，并进入同一设置页 smoke 入口
- `node scripts/verify-model-probe-help.mjs` 通过
- `node scripts/verify-dynamic-routing-help.mjs` 通过
- `node scripts/verify-circuit-breaker-help.mjs` 通过
- `node scripts/verify-setting-help-browser-smoke.mjs --check-only` 通过
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/ModelProbe.tsx`
- `scripts/verify-model-probe-help.mjs`
- `scripts/verify-setting-help-browser-smoke.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-dual-mode-prep.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先核对脚本与设置页状态，再改浏览器 smoke 入口和 `ModelProbe` 锚点
- 受影响后端模块：无
- 受影响前端模块：设置页 `ModelProbe` 卡片
- 受影响接口：无新增接口；复用 `/api/v1/user/login` 和现有设置页接口
- 是否影响旧数据：否
- 是否影响旧行为：否；只增强验证入口和 DOM 锚点，不改变页面业务语义

## 8. 实施步骤

1. 复核主规划、前端主线状态、手工 checklist、automation memory 与设置页现有脚本，确认当前主线仍是 Phase G 设置页帮助提示浏览器收口。
2. 重写 `scripts/verify-setting-help-browser-smoke.mjs`，加入 `check-only / external / self-start` 模式、显式 Playwright session、自启前端 API 基址注入和 `ModelProbe` 检查。
3. 给 `ModelProbe` 补稳定 `data-testid`，并同步无浏览器验证脚本、前端主线状态文档和本轮记录。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：
  - `node scripts/verify-model-probe-help.mjs`
  - `node scripts/verify-dynamic-routing-help.mjs`
  - `node scripts/verify-circuit-breaker-help.mjs`
  - `node scripts/verify-setting-help-browser-smoke.mjs --check-only`
- 专项验证：
  - 复核设置页浏览器 smoke 脚本现在会同时检查 `DynamicRouting / CircuitBreaker / ModelProbe`
  - 复核 `ModelProbe` 的 DOM 锚点和 no-browser 断言已同步
  - 复核前端主线状态文档已把模型探测和双模式 smoke 入口写回当前状态

## 10. 风险与兼容性

- 新风险：低；本轮只增强脚本入口、锚点和文档
- 兼容性风险：低；未改接口、持久化或业务流程
- 是否阻塞下一任务：不阻塞；但真实浏览器 smoke 仍依赖可用的前后端服务或允许起进程的环境

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过（`check-only` 与 no-browser 验证通过；真实浏览器 smoke 仍未在本轮跑通）
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、手工验收清单、automation memory、设置页组件、前端 API 客户端、现有 no-browser 验证脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划和工作流明确本轮应继续留在 Phase G 设置页同主线，不应跳去无关优化
  - 前端主线状态和手工 checklist 明确设置页复杂交互仍缺浏览器级证据，且桌面 / `375px` 都要覆盖
  - automation memory 指向的下一优先项仍是设置页帮助提示浏览器 smoke，先于其它页面
  - `web/src/api/client.ts` 说明自启前端必须显式注入 `NEXT_PUBLIC_API_BASE_URL`，否则会把脚本失败和环境失败混在一起
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：当前线程没有现成前后端服务；本机 host policy 对起子进程链仍有不确定性，因此本轮先把双模式脚本和 `check-only` 入口收口
- 待验证页面清单：设置页 `DynamicRouting`、`CircuitBreaker`、`ModelProbe` 的桌面布局、`375px` 布局、帮助提示 hover/focus
- 若未使用子 agent，原因：用户要求主线程执行，且本轮是强耦合的小闭环
- worklog 是否更新：是
- 遗留项：
  - 真实浏览器级 smoke 仍未在本轮拿到通过证据
  - 目标环境若继续限制起进程，下一轮需要优先复用外部已启动服务走 `--external`
  - `Vitest` 在当前主机仍受 `spawn EPERM` 影响，不适合拿来替代这条浏览器 smoke 链
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：直接在可用环境里执行 `node scripts/verify-setting-help-browser-smoke.mjs --external`，或在允许起进程的环境里执行默认自启模式，补齐桌面与 `375px` 的真实浏览器证据。
- 同主线候选顺序：
  1. 设置页 `DynamicRouting / CircuitBreaker / ModelProbe` 浏览器级 smoke 实跑
  2. 把真实 smoke 结果写回 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md` 与 worklog
  3. 仅在设置页 smoke 证据补齐后，再继续同主线下一个图片问题池页面的浏览器验收