# 2026-04-22 Phase G Channel Create Layered Guidance Closure

## 1. 任务信息

- 任务名称：创建渠道弹窗分层引导与帮助提示收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：图片问题池 / 创建渠道弹窗主路径引导收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9 节、第 14 节、第 16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.2 节、第 1.3 节、第 1.4 节、第 9 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-route-target-overrides-copy-and-fold-closure.md`
- 本次任务目标：
  - 把创建渠道弹窗的主路径说明补齐到“基础信息 -> 连接与密钥 -> 模型 -> 可选高级设置”的可理解层级。
  - 给关键入口补上帮助提示，减少用户对“先填什么、什么时候展开高级项”的猜测成本。
  - 为模型空态补下一步动作提示，并用无浏览器脚本固定这轮断言。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/channel/Form.tsx`
  - `web/src/components/common/HelpHint.tsx`
  - `scripts/verify-channel-create-flow.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：canonical plan、用户上下文总账、详细工作流、前端主线状态、上一轮 automation memory、本地 `frontend-design` / `brainstorming` / `using-superpowers`
- 若未使用部分本地资源或上下文，原因：本轮是创建渠道同页小闭环，不需要继续展开备份导入或后端主线文件
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求本轮不要创建子 agent，且本任务是单文件附近的前端微闭环

## 3. 本次硬规则

- 只处理创建渠道弹窗主路径引导、帮助提示与空态说明，不扩散到其他截图主题。
- 必须对齐需求 14、19、20、27、42，不允许新增英文主显示文案。
- 必须留下可重复执行的无浏览器验证，不把纯目测改动标记为完成。

## 4. 本次禁止事项

- 不改后端接口、路由语义和持久化结构。
- 不把 `CC Switch`、熔断设置或备份导入问题混进本轮修改。
- 不把未执行的浏览器级 smoke 写成已完成。

## 5. 本次验收条件

- 创建渠道弹窗主路径层次更清楚，关键字段有帮助提示说明。
- 模型空态提供明确下一步动作提示。
- `scripts/verify-channel-create-flow.mjs` 通过，且前端 typecheck 通过。

## 6. 本次回滚点

- `web/src/components/modules/channel/Form.tsx`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `scripts/verify-channel-create-flow.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 UI 引导和帮助提示，再补脚本断言
- 受影响后端模块：无
- 受影响前端模块：创建渠道 / 编辑渠道弹窗
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅增强表单引导、提示和空态文案，不改变保存行为

## 8. 实施步骤

1. 读取 `Form.tsx`、locale 和用户上下文总账中“创建渠道弹窗问题”章节，确认当前缺口集中在主路径解释和帮助提示。
2. 在 `ChannelForm` 中补 `Provider 预设`、渠道名称、渠道类型、连接分段、基础地址、自动分组的帮助提示，并让模型空态带出下一步动作说明。
3. 更新中英繁三套 locale，并增强 `verify-channel-create-flow.mjs`，把新增提示和分段断言固定下来。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`node scripts/verify-channel-create-flow.mjs`
- 专项验证：确认 `Form.tsx` 中连接分段使用 `flowCopy.keySectionTitle`，基础地址改用统一 `HelpHint`，模型空态包含 `modelNoSelectedHint`

## 10. 风险与兼容性

- 新风险：低；本轮只涉及表单结构提示和 locale 文案
- 兼容性风险：低；未修改 API 请求结构
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、用户上下文总账、详细工作流、前端主线状态、上一轮 memory、`frontend-design`、`brainstorming`、`using-superpowers`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账明确这轮必须命中“创建渠道弹窗问题”的分层引导与中文帮助提示。
  - 前端主线状态和上一轮 memory 指向同一主线的最小闭环就是创建渠道引导收口。
  - `HelpHint` 组件已存在，可直接复用而不必新增交互形态。
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先做同页静态收口与 typecheck；浏览器级验证仍待同主线统一安排
- 待验证页面清单：创建渠道弹窗、编辑渠道弹窗、创建渠道后的模型读取与高级设置展开
- 若未使用子 agent，原因：用户要求主线程执行，且本轮任务是单模块微闭环
- worklog 是否更新：是
- 遗留项：创建渠道浏览器级 smoke 尚未完成；`CC Switch` 渐进帮助和中文化仍是同主线下一优先项
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：`CC Switch` 渐进式帮助提示与中文化收口
- 同主线候选顺序：
  1. `CC Switch` 渐进帮助提示与同页分层说明收口
  2. 创建渠道弹窗浏览器级 smoke
  3. 熔断设置的“默认简洁 + 高级展开”说明收口
