# 2026-04-23 Phase G Model Search Placeholder Contract Closure

## 1. 任务信息

- 任务名称：模型页搜索入口占位契约收口
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：模型/价格区域中文契约与 no-browser 守护补强

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.1`、`4.1`、`9`、`14` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.1`、`1.2`、`1.3`、`1.4.1` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-model-search-contract-alignment.md`
  - `docs/worklog/2026-04-23-phase-g-model-card-meta-order-closure.md`
- 本次任务目标：
  - 在同一 `Phase G screenshot-first` 主线下继续收口模型/价格区域的搜索契约缺口
  - 让模型页真实搜索入口和当前中文提示完全对齐，不再只有过滤逻辑支持“规范名称”而入口本身没有对应上下文
  - 同步 `verify-llm-price-boundary.mjs` 与 automation memory，给搜索入口契约补回 no-browser 护栏
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-model-search-contract-alignment.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/toolbar/index.tsx`
  - `scripts/verify-llm-price-boundary.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、上一轮模型搜索 worklog、automation memory、当前工具栏与模型页脚本源码
- 若未使用部分本地资源或上下文，原因：本轮只处理模型页搜索入口提示与回归脚本，不扩散到浏览器证据、首页或渠道页
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮是前端单点契约修正，串行收口更稳

## 3. 本次硬规则

- 只处理模型页搜索入口占位与直接相关的 no-browser 验证，不跨到其它页面或后端筛选接口
- 必须让真实入口提示和现有中文/多语言搜索契约一致，不能继续保留“逻辑支持了，但用户入口看不出来”的半闭环
- 必须补回静态断言，避免下轮再次丢失占位提示或改回泛化搜索文案

## 4. 本次禁止事项

- 不顺手重做工具栏整体布局或搜索宽度动画
- 不把模型页浏览器级 `375px / hover / focus` 证据误记为完成
- 不扩大到 channel/group 页搜索占位之外的复杂筛选逻辑

## 5. 本次验收条件

- 工具栏搜索入口按页面使用对应 placeholder，模型页为“搜索模型名称或规范名称”
- `verify-llm-price-boundary.mjs` 通过并覆盖工具栏模型搜索入口占位契约
- `node scripts/verify-locale-consistency.mjs` 通过
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/toolbar/index.tsx`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `web/public/locale/ja.json`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/worklog/2026-04-23-phase-g-model-search-placeholder-contract-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先补真实搜索入口占位，再补脚本断言
- 受影响后端模块：无
- 受影响前端模块：`toolbar` 搜索入口、多语言文案、模型/价格验证脚本
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：是，搜索输入展开后不再是无上下文空输入框，而是按页面显示更具体的搜索范围提示

## 8. 实施步骤

1. 复核主规划、用户上下文总账、当前状态文档、前端主线状态和 automation memory，确认本轮仍属于 `Phase G screenshot-first` 的 no-browser 小闭环。
2. 更新 `web/src/components/modules/toolbar/index.tsx`，为 channel/group/model 三个页面的搜索入口补上独立 placeholder key。
3. 在四种语言 locale 中补齐对应 placeholder 文案，确保模型页明确写出“模型名称或规范名称”。
4. 更新 `scripts/verify-llm-price-boundary.mjs`，把工具栏占位契约加入静态回归护栏。
5. 运行脚本与类型检查，确认本轮闭环成立。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 专项验证：读回 `toolbar/index.tsx` 与 locale，确认模型页 placeholder 已明确承诺“模型名称或规范名称”，且静态脚本已覆盖入口占位契约

## 10. 风险与兼容性

- 新风险：低；只改占位提示和静态断言，不改搜索状态存储或过滤算法
- 兼容性风险：低；未改接口、未改数据结构、未改页面行为链路
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、上一轮模型搜索 worklog、automation memory、工具栏源码与模型页验证脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账中的需求 `42`、`43`、`49` 要求中文主界面和多语言契约都要完整落到真实交互入口，而不仅是内部实现
  - 当前状态文档与前端主线状态都表明模型/价格区仍在同一 screenshot-first 池，继续做 no-browser 契约收口符合优先级
  - 上一轮模型搜索 worklog 证明搜索逻辑已经支持 `canonical_name`，本轮最值得补的是用户可见入口契约
  - automation memory 继续确认宿主 `Edge/CDP` 阻塞未解，因此本轮继续选择 no-browser 可验证的小闭环
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级 `375px / hover / focus` 证据，本轮继续以 no-browser 验证收口
- 待验证页面清单：模型页浏览器级 `375px` 布局、模型页普通/紧凑切换、设置页帮助提示 hover/focus 证据
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮为前端单点契约修正
- worklog 是否更新：是
- 遗留项：
  - 模型/价格区域浏览器级 `375px` 证据仍未补齐
  - 若继续停留在 no-browser 路径，下一刀应优先选择同池内新的具体入口契约或布局残留，而不是重复扫描旧点
  - toolbar 搜索框当前宽度仍偏保守，若后续影响可读性，可再做同池的小范围布局收紧
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一 screenshot-first 池，优先尝试模型页或设置页的浏览器级 `375px / hover / focus` 证据；若宿主浏览器链路仍阻塞，则转向同池的下一处具体 no-browser 契约或布局缺口。
- 同主线候选顺序：
  1. 模型页浏览器级 `普通 / 紧凑` 与 `375px` 证据
  2. 设置页 help-hint 浏览器级 `hover / focus / 375px` 证据
  3. channel/group create dialog 浏览器级 `375px` / help-hint 证据
  4. 模型页 toolbar 搜索框宽度与小屏提示可读性收紧
