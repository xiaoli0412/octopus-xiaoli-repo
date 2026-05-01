# 2026-04-23 Phase G Model Layout Runtime Closure

## 1. 任务信息

- 任务名称：模型页普通 / 紧凑布局真实生效收口
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：模型/价格区域普通布局与紧凑布局双模式收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、1.4、9 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-llm-price-boundary-closure.md`
- 本次任务目标：
  - 修复模型页工具栏布局切换只停留在入口状态、未真正作用到卡片层的问题
  - 补齐模型/价格区域的 no-browser 回归断言，避免再次出现“有切换入口但卡片不变化”的假闭环
  - 同步前端状态文档、当前状态文档和 automation memory
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/model/index.tsx`
  - `web/src/components/modules/model/Item.tsx`
  - `web/src/components/modules/toolbar/index.tsx`
  - `scripts/verify-llm-price-boundary.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、automation memory、模型模块现有实现与验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只处理模型页布局真实生效的小闭环，不扩散到后端、浏览器 smoke 或其它页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、写入点高度集中，串行收口更稳妥

## 3. 本次硬规则

- 只处理模型/价格区域普通布局与紧凑布局真实生效的问题，不扩散到其他主线
- 必须保留当前工具栏入口语义，只补真实渲染链与回归断言
- 必须留下可复跑的 no-browser 验证证据

## 4. 本次禁止事项

- 不修改后端接口、模型数据结构或价格同步语义
- 不把浏览器级 `375px` 证据误记为已经完成
- 不顺手扩散处理其他模型/价格英文文案问题

## 5. 本次验收条件

- 模型页工具栏中的“普通 / 紧凑”切换真实作用到 `ModelItem`
- `scripts/verify-llm-price-boundary.mjs` 覆盖模型布局回归断言并通过
- `node scripts/verify-locale-consistency.mjs` 通过
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/model/index.tsx`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/worklog/2026-04-23-phase-g-model-layout-runtime-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先核对布局状态传递链，再改渲染与验证
- 受影响后端模块：无
- 受影响前端模块：模型页渲染链、模型/价格 no-browser 验证脚本、状态文档
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只修正布局切换的实际生效范围，不改变模型数据和工具栏交互入口

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态和 automation memory，确认当前模型/价格主线要求是“普通布局与紧凑布局都要真实成立”。
2. 更新 `web/src/components/modules/model/index.tsx`，把当前 `layout` 真实传给 `ModelItem`。
3. 重写 `scripts/verify-llm-price-boundary.mjs`，补齐模型布局断言并规避宿主编码问题。
4. 同步前端主线状态文档、当前状态文档与本轮 worklog。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 专项验证：读回 `web/src/components/modules/model/index.tsx` 与 `scripts/verify-llm-price-boundary.mjs`，确认布局状态传递与 no-browser 回归断言均已落地

## 10. 风险与兼容性

- 新风险：低；只修正前端布局状态传递与静态断言
- 兼容性风险：低；未改接口、未改后端、未改数据结构
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、automation memory、模型模块现有实现与验证脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账把“模型/价格区域普通布局与紧凑布局双模式”列为需求 49 和截图优先池中的当前缺口
  - 前端主线状态文档明确模型/价格区仍缺普通/紧凑双模式收口
  - 当前 automation memory 指向同一 Phase G screenshot-first 池，允许在浏览器证据阻塞时继续推进 no-browser 小闭环
  - 现有模型页代码显示工具栏已持有 `layout`，但 `ModelItem` 未消费该状态，属于可直接闭环的真实断点
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先以 no-browser 验证和 typecheck 收口；浏览器级 `375px` 与 hover/focus 证据仍受宿主 Edge/CDP 阻塞影响
- 待验证页面清单：模型页普通布局、模型页紧凑布局、模型/价格区域英文主显示清理、浏览器级 `375px` 布局
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮为模型页单点收口任务
- worklog 是否更新：是
- 遗留项：
  - 模型/价格区域英文主显示与普通布局信息顺序仍需继续收口
  - 浏览器级 `375px` 证据仍未补齐
  - 更细的模型/价格区中文化与布局压缩还需后续同池任务继续推进
- 下一任务前置条件是否满足：满足
