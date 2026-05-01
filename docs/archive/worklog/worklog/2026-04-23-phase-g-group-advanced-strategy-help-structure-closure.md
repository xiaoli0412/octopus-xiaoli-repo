# 2026-04-23 Phase G 分组高级策略帮助提示结构收口

## 1. 任务信息

- 任务名称：分组创建高级策略折叠头帮助提示结构收口
- 日期：2026-04-23
- 当前阶段：Phase G 截图优先 UI 主线
- 对应 milestone：group create no-browser 结构一致性补强

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.2`、`9.3`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`9.3` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-group-create-localized-fallback-closure.md`
  - `docs/worklog/2026-04-23-phase-g-help-hint-locale-and-selector-closure.md`
- 本次任务目标：
  - 继续留在 `Phase G screenshot-first` 同一主线内，为 group create/edit 再收一刀可验证的小闭环。
  - 把高级策略折叠头中的 `HelpHint` 从原始 `AccordionPrimitive.Trigger` 按钮树内移出，改用公共 `AccordionTrigger` 的 `addon` 槽位。
  - 同步 `verify-group-create-flow.mjs`，锁定这条结构契约，防止后续回退到嵌套交互结构。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-group-create-localized-fallback-closure.md`
  - `docs/worklog/2026-04-23-phase-g-help-hint-locale-and-selector-closure.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/ui/accordion.tsx`
  - `scripts/verify-group-create-flow.mjs`
  - `scripts/verify-help-hint-accessible.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、同模块强耦合，直接收口更稳。

## 3. 本次硬规则

- 只处理 group create/edit 的高级策略折叠头帮助提示结构，不扩散到新的页面或后端主题。
- 不把未执行的真实浏览器 `375px / hover / focus` 证据记成完成。
- 本轮必须留下“代码改动 + 静态验证 + 状态同步 + 下一轮入口”的闭环。

## 4. 本次禁止事项

- 不回滚工作区已有无关改动。
- 不改 group 提交语义、后端接口或模型选择结构。
- 不把帮助提示降级成不可聚焦或不可访问的纯视觉图标。

## 5. 本次验收条件

- `GroupEditor` 高级策略折叠头改用公共 `AccordionTrigger`，帮助提示走 `addon` 槽位，不再嵌在原始 trigger 按钮树里。
- `scripts/verify-group-create-flow.mjs` 补上对应结构断言，并禁止该入口回退为原始 `AccordionPrimitive.Trigger` + 内嵌 `HelpHint`。
- `node scripts/verify-group-create-flow.mjs`、`node scripts/verify-help-hint-accessible.mjs`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 全部通过。

## 6. 本次回滚点

- `web/src/components/modules/group/Editor.tsx`
- `scripts/verify-group-create-flow.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/worklog/2026-04-23-phase-g-group-advanced-strategy-help-structure-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 group create 的折叠头结构，再补脚本与状态记录
- 受影响后端模块：无
- 受影响前端模块：`GroupEditor` 高级策略折叠头、group create 静态验证脚本
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅改善折叠头帮助提示的 DOM 结构与回归护栏，不改变表单字段、保存逻辑或分组选择体验

## 8. 实施步骤

1. 复核主规划、前端主线状态、group create 上一轮 worklog 与 automation memory，确认本轮继续停留在 `Phase G screenshot-first` / group create 同池收口。
2. 检查 `web/src/components/ui/accordion.tsx` 的公共 `AccordionTrigger` 能力，确认已有 `addon` 槽位可复用。
3. 更新 `web/src/components/modules/group/Editor.tsx`，把高级策略折叠头从原始 `AccordionPrimitive.Trigger` 改为公共 `AccordionTrigger`，并把 `HelpHint` 放到 `addon`。
4. 更新 `scripts/verify-group-create-flow.mjs`，补上导入、结构与禁止回退断言。
5. 运行静态验证与 TypeScript 检查，确认本轮闭环。
6. 同步前端主线状态、当前状态文档与本轮 worklog。

## 9. 测试与验证

- `node scripts/verify-group-create-flow.mjs`
- `node scripts/verify-help-hint-accessible.mjs`
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`

## 10. 风险与兼容性

- 新风险：低；本轮只改 group create 高级策略折叠头的 DOM 结构与静态断言
- 兼容性风险：低；未修改 group 提交结构、排序逻辑或模型选择逻辑
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（TypeScript noEmit）
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、group create 上一轮 worklog、automation memory、`GroupEditor` / `AccordionTrigger` / 静态验证脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 前端主线状态与 automation memory 都确认浏览器级证据仍受宿主 CDP 阻塞，因此本轮应继续选择同池可验证的 no-browser 小闭环。
  - group create 上一轮 worklog 已经收掉文案和回退兜底，因此这轮最值得继续补的是折叠头帮助提示结构风险，而不是重开文案主题。
  - 公共 `AccordionTrigger` 已有 `addon` 槽位，可直接复用，不需要再造新交互容器。
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级 `375px / hover / focus` 证据，本轮继续以 no-browser 验证收口
- 待验证页面清单：
  - group create dialog 浏览器级 `375px` 与 help-hint hover/focus
  - channel create dialog 浏览器级 `375px` 与 help-hint hover/focus
  - 设置页 help-hint 浏览器级 `375px` 与 hover/focus
- worklog 是否更新：是
- 遗留项：
  - 分组创建/编辑弹窗浏览器级 `375px` 证据仍未补齐
  - 同一 screenshot-first 池中的 help-hint hover/focus 浏览器证据仍待推进
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一 screenshot-first 池，优先尝试 `group create` 或 `channel create` 的浏览器级 `375px / hover / focus` 证据；若宿主浏览器链路仍阻塞，则继续做同池 help-hint / create-dialog 的 no-browser 结构补强。
- 同主线候选顺序：
  1. group create dialog 浏览器级 `375px` / hover / focus 证据
  2. channel create dialog 浏览器级 `375px` / hover / focus 证据
  3. 设置页 help-hint 浏览器级 `375px` / hover / focus 证据
  4. 同池剩余中文界面英文主显示或折叠交互结构收口
