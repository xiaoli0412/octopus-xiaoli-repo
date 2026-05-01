# 2026-04-22 Phase G LLM Price Boundary Closure

## 1. 任务信息

- 任务名称：设置页价格维护卡片与模型探测职责边界收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：截图问题池 / 设置页价格维护与探测归位收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、1.4 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-model-probe-progressive-guidance-closure.md`
- 本次任务目标：
  - 把设置页价格维护卡片收口成“价格维护专属区”，明确不再承载探测或检测入口
  - 补齐价格维护与模型探测之间的帮助提示、默认路径说明与职责边界文案
  - 补一条 no-browser 验证脚本，固定本轮边界断言
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/setting/LLMPrice.tsx`
  - `web/src/components/modules/setting/ModelProbe.tsx`
  - `scripts/verify-model-probe-help.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细执行工作流、前端主线状态、上轮 automation memory、设置页相邻模块现有实现
- 若未使用部分本地资源或上下文，原因：本轮只聚焦设置页价格维护卡片的小范围边界收口，不扩展到备份、首页或后端接口
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、耦合高，适合直接串行收口

## 3. 本次硬规则

- 只处理设置页价格维护卡片与同页探测职责边界，不扩散到其它主题
- 价格页只保留价格和计费语义，不重新混入探测和检测入口
- 必须留下可复跑的 no-browser 验证证据

## 4. 本次禁止事项

- 不修改后端接口、模型持久化结构或探测运行时语义
- 不把模型页、首页或备份页拉进本轮闭环
- 不在未跑浏览器 smoke 的情况下声称页面已全量验收

## 5. 本次验收条件

- 设置页价格维护卡片出现默认路径说明、职责摘要卡片和探测迁移提示
- 三语 locale 补齐新增边界文案与帮助提示
- `node scripts/verify-llm-price-boundary.mjs` 通过
- `node scripts/verify-model-probe-help.mjs` 通过
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/LLMPrice.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先核对价格维护与模型探测的职责边界，再改 UI 与文案
- 受影响后端模块：无
- 受影响前端模块：设置页价格维护卡片与 locale
- 受影响接口：无新增接口；继续复用模型价格同步与设置项接口
- 是否影响旧数据：否
- 是否影响旧行为：仅增强设置页价格维护卡片的结构与帮助说明，不改变价格同步行为

## 8. 实施步骤

1. 复核设置页价格维护卡片、模型探测卡片和用户上下文总账，确认当前主线要求是“价格维护与探测彻底分离”。
2. 重写 `LLMPrice.tsx`，加入默认路径说明、职责摘要卡片、探测迁移提示和帮助提示。
3. 补齐三语 locale、添加 `verify-llm-price-boundary.mjs`，并同步前端主线状态文档。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-model-probe-help.mjs`
- 专项验证：读回 `LLMPrice.tsx`、三语 `llmPrice` locale 片段和前端主线状态文档，确认职责边界与帮助提示已经落地

## 10. 风险与兼容性

- 新风险：低；本轮只增强设置页价格维护解释层与静态校验
- 兼容性风险：低；未改接口和持久化数据
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细执行工作流、前端主线状态、automation memory、设置页价格维护与模型探测现有实现
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账明确要求“探测与检测设置必须从价格区域迁出，归位到设置页”
  - 前端主线状态文档已把该要求列为当前前端优先池，但价格卡片之前仍缺少显式边界说明
  - 上轮 model probe worklog 已把模型探测卡片收口成渐进式结构，本轮应与其形成成对闭环
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先以 no-browser 验证和 typecheck 收口，浏览器级页面回归仍需在同主线下继续安排
- 待验证页面清单：设置页价格维护卡片桌面布局、375px 布局、帮助提示 hover/focus、与模型探测卡片相邻阅读顺序
- 若未使用子 agent，原因：用户要求主线程执行，且本轮为同页小闭环
- worklog 是否更新：是
- 遗留项：
  - 价格维护卡片仍缺浏览器级 smoke 证据
  - 设置页相邻模块的浏览器级帮助提示回归仍未整合成一轮统一 smoke
  - 价格同步后的更细粒度反馈和异常状态展示仍可继续增强
- 下一任务前置条件是否满足：满足
