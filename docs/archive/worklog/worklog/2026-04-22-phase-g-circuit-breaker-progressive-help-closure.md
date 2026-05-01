# 2026-04-22 Phase G Circuit Breaker Progressive Help Closure

## 1. 任务信息

- 任务名称：设置页熔断模块渐进帮助与状态摘要收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：截图问题池 / 设置页熔断说明与高级展开收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、1.4 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-ccswitch-progressive-help-closure.md`
- 本次任务目标：
  - 把设置页熔断模块从“三个数字输入框”收口成“默认简洁 + 高级展开 + 帮助提示 + 状态摘要”
  - 复用已有 runtime 熔断摘要字段，不新增后端接口
  - 补一条 no-browser 验证脚本，固定本轮结构和文案断言
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
  - `web/src/api/endpoints/stats.ts`
  - `internal/relay/balancer/circuit.go`
  - `internal/relay/balancer/circuit_test.go`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、前端主线状态、上轮 automation memory、设置页与首页现有实现
- 若未使用部分本地资源或上下文，原因：本轮是同主线下的小范围 UI/文案闭环，不需要扩展到备份和其它页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行且本轮范围小、耦合高

## 3. 本次硬规则

- 只处理设置页熔断模块，不扩散到无关页面
- 文案必须与当前真实熔断恢复逻辑一致
- 必须留下可复跑的静态验证结果

## 4. 本次禁止事项

- 不修改后端熔断接口或持久化结构
- 不把首页 runtime 摘要误写成历史报表
- 不在未做浏览器 smoke 的情况下宣称页面已全量验收

## 5. 本次验收条件

- 设置页熔断模块具备默认说明、推荐提示、状态摘要和高级展开
- 三语 locale 补齐新增帮助提示与恢复说明
- `node scripts/verify-circuit-breaker-help.mjs` 通过
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/CircuitBreaker.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `scripts/verify-circuit-breaker-help.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先核对运行时熔断语义，再改 UI 与文案
- 受影响后端模块：无
- 受影响前端模块：设置页熔断模块与 locale
- 受影响接口：无新增接口；复用 `/api/v1/stats/token-breakdown`
- 是否影响旧数据：否
- 是否影响旧行为：仅增强设置页展示结构和帮助说明，不改变后端熔断判定逻辑

## 8. 实施步骤

1. 核对设置页现状、首页熔断摘要和后端熔断恢复逻辑，确认可复用 runtime 摘要字段。
2. 重写 `CircuitBreaker.tsx`，加入推荐提示、摘要卡片、恢复路径说明和高级展开层。
3. 补齐三语 locale、添加 `verify-circuit-breaker-help.mjs`，并同步前端主线状态文档。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`node scripts/verify-circuit-breaker-help.mjs`
- 专项验证：读回 `CircuitBreaker.tsx`、三语 `circuitBreaker` locale 片段与前端主线状态文档，确认新增结构与文案落地

## 10. 风险与兼容性

- 新风险：低；本轮只增强前端解释层与静态校验
- 兼容性风险：低；未改接口和持久化数据
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、前端主线状态、automation memory、设置页与首页代码、后端熔断逻辑与测试
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账明确本轮必须命中熔断“默认简洁 + 高级自定义 + 帮助提示”要求
  - 首页现有 token breakdown 已提供可复用的熔断摘要字段，无需新增接口
  - 后端熔断逻辑说明当前真实恢复路径是 `Open -> HalfOpen -> Closed` / `HalfOpen -> Open`
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先以 no-browser 验证和 typecheck 收口，浏览器级页面回归仍需在同主线下继续安排
- 待验证页面清单：设置页熔断模块桌面布局、375px 布局、帮助提示 hover/focus
- 若未使用子 agent，原因：用户要求主线程执行，且本轮为同页小闭环
- worklog 是否更新：是
- 遗留项：
  - 浏览器级 smoke 仍未完成
  - 更细粒度的手动恢复入口和失败分类自定义仍未进入 UI
  - CC Switch 页的浏览器 smoke 也仍待补
- 下一任务前置条件是否满足：满足
