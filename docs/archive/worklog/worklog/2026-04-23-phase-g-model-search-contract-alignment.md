# 2026-04-23 Phase G Model Search Contract Alignment

## 1. 任务信息

- 任务名称：模型页搜索契约对齐收口
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：模型/价格区域布局与中文语义收口后的搜索契约对齐

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`9` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-model-layout-runtime-closure.md`
  - `docs/worklog/2026-04-23-phase-g-model-card-meta-order-closure.md`
- 本次任务目标：
  - 在同一 `Phase G screenshot-first` 主线下继续收敛模型/价格区域的可见契约偏差
  - 修复模型页搜索提示写着“模型名称或规范名称”，但代码实际只按模型名过滤的问题
  - 同步 `verify-llm-price-boundary.mjs`、状态文档和 automation memory，给新契约加 no-browser 回归护栏
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-model-layout-runtime-closure.md`
  - `docs/worklog/2026-04-23-phase-g-model-card-meta-order-closure.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/model/index.tsx`
  - `scripts/verify-llm-price-boundary.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、两份最近模型 worklog、automation memory、模型列表实现与现有 no-browser 验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只处理模型页搜索契约，不扩散到浏览器证据、后端过滤接口或其他页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮为模型模块单点契约修正，串行处理更稳妥

## 3. 本次硬规则

- 只处理模型页搜索契约与直接相关的 no-browser 验证，不扩散到后端接口或其他筛选维度
- 必须让实现行为与现有中文提示一致，不能保留“文案说一套、代码做一套”的状态
- 必须留下可复跑的静态断言，防止后续回退

## 4. 本次禁止事项

- 不顺手改模型同步、价格计算或探测策略逻辑
- 不把浏览器级 `375px` / hover / focus 证据误记为完成
- 不扩大到首页、渠道页或分组页搜索逻辑

## 5. 本次验收条件

- 模型页搜索同时命中 `name` 与 `canonical_name`
- `verify-llm-price-boundary.mjs` 通过并覆盖新的搜索契约断言
- `node scripts/verify-locale-consistency.mjs` 通过
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/model/index.tsx`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/worklog/2026-04-23-phase-g-model-search-contract-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改前端过滤逻辑，再补 no-browser 断言与状态同步
- 受影响后端模块：无
- 受影响前端模块：模型列表过滤逻辑、模型/价格验证脚本、状态文档
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：是，模型页搜索从“仅模型名命中”收紧为“模型名 + 规范名称”双命中；这与现有输入提示保持一致

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态、最近模型 worklog 和 automation memory，确认模型/价格主线当前允许继续用 no-browser 小闭环收敛可见契约偏差。
2. 更新 `web/src/components/modules/model/index.tsx`，让搜索词同时匹配 `name` 与 `canonical_name`。
3. 更新 `scripts/verify-llm-price-boundary.mjs`，新增对搜索契约实现的静态断言。
4. 运行 `verify-llm-price-boundary`、`verify-locale-consistency` 和 `tsc`，确认本轮闭环成立。
5. 同步前端主线状态文档、当前状态文档、worklog 与 automation memory。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 专项验证：读回 `web/src/components/modules/model/index.tsx` 与 `scripts/verify-llm-price-boundary.mjs`，确认搜索词会同时命中模型名称与规范名称，且静态断言已经覆盖

## 10. 风险与兼容性

- 新风险：低；只调整前端本地过滤条件和静态断言
- 兼容性风险：低；未改接口、未改数据结构、未改后端排序与筛选语义
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、两份最近模型 worklog、automation memory、模型页源码与 no-browser 验证脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账中的需求 `49` 和 `4.12.4` 明确要求模型/价格区既要保持中文语义，也要避免视觉与交互契约失真
  - 前端主线状态与当前状态文档都显示模型/价格区仍处于同一 screenshot-first 池，允许继续做 no-browser 小闭环
  - 最近两份模型 worklog 已经收口布局切换与 meta 顺序，所以本轮最值得推进的是把搜索行为补齐到现有文案承诺
  - automation memory 继续确认宿主 `Edge/CDP` 阻塞未解，本轮维持 no-browser 收口是最稳妥路径
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级 `375px / hover / focus` 证据，本轮继续以 no-browser 验证收口
- 待验证页面清单：模型页浏览器级 `375px` 布局、模型页普通/紧凑切换、设置页相关卡片 hover/focus 证据
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务为模型页单点契约修正
- worklog 是否更新：是
- 遗留项：
  - 模型/价格区域浏览器级 `375px` 证据仍未补齐
  - 若后续发现新的中文界面英文主显示泄漏，仍需在同池继续做 no-browser 小闭环
  - 更细的模型筛选维度仍未补齐
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一 screenshot-first 池，优先尝试模型页或设置页的浏览器级 `375px / hover / focus` 证据；若宿主浏览器链路仍阻塞，则转向同池的下一处具体 no-browser 契约偏差。
- 同主线候选顺序：
  1. 模型页浏览器级 `普通 / 紧凑` 与 `375px` 证据
  2. 设置页 help-hint 浏览器级 `hover / focus / 375px` 证据
  3. 模型页更细的筛选维度或新的中文主显示泄漏清理
  4. channel/group create dialog 浏览器级 `375px` / help-hint 证据
