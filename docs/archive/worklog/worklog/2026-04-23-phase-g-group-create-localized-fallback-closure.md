# 2026-04-23 Phase G 分组创建本地化兜底收口

## 1. 任务信息

- 任务名称：分组创建/编辑弹窗本地化兜底与静态验证收口
- 日期：2026-04-23
- 当前阶段：Phase G 截图优先 UI 主线
- 对应 milestone：P0 分组弹窗中文化与一致性补强

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.2`、`9.3`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`11.2` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-group-create-progressive-guidance-closure.md`
  - `docs/worklog/2026-04-23-phase-g-channel-create-multi-key-entry-guidance-closure.md`
- 本次任务目标：
  - 收掉分组创建/编辑路径里仍可能泄漏到中文界面的英文正则错误回显。
  - 收掉分组卡片成员列表缺渠道名时的硬编码 `Channel {id}` 回退，统一走 locale 文案。
  - 把这两个兜底点写进 `verify-group-create-flow.mjs`，避免后续回退。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/modules/group/Card.tsx`
  - `scripts/verify-group-create-flow.mjs`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/en.json`
  - `web/public/locale/ja.json`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、上一轮 channel create/CC Switch worklog、automation memory、当前 group 模块源码与验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只处理同一截图池内的分组模块小闭环，不需要扩展到浏览器 smoke 或后端实现
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：N/A
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、上下文强耦合

## 3. 本次硬规则

- 只处理分组创建/编辑路径的本地化兜底和验证同步，不改后端接口与 group 提交语义。
- 中文界面的错误提示和渠道回退名不能再直接裸露英文技术 message 或硬编码英文前缀。
- 本轮必须留下可重复执行的无浏览器验证，不把文案收口写成未验证完成。

## 4. 本次禁止事项

- 不扩散到首页、模型价格页、备份页或浏览器证据链。
- 不回滚工作区中已有的无关修改。
- 不把尚未执行的真实浏览器 `375px` 验证记成已完成。

## 5. 本次验收条件

- `GroupEditor` 不再把 `Invalid regex` 之类的英文错误消息直接回显到界面，而是使用本地化主提示与帮助说明。
- `GroupCard` 在渠道名缺失时不再回退成硬编码 `Channel {id}`，而是使用 locale 文案。
- `scripts/verify-group-create-flow.mjs`、`scripts/verify-locale-consistency.mjs`、`web` TypeScript 检查通过。

## 6. 本次回滚点

- `web/src/components/modules/group/Editor.tsx`
- `web/src/components/modules/group/Card.tsx`
- `scripts/verify-group-create-flow.mjs`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `web/public/locale/ja.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 UI 兜底文案与回退逻辑，再同步脚本与 locale
- 受影响后端模块：无
- 受影响前端模块：分组创建/编辑弹窗、分组卡片成员展示
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅收口本地化兜底与错误提示表达，不改变 group 保存结构或路由语义

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态与 automation memory，确认本轮继续停留在 `Phase G` 同一截图池。
2. 复盘 `GroupEditor.tsx` 与 `GroupCard.tsx`，定位仍会进入中文路径的英文兜底点：正则错误 message 直出与 `Channel {id}` 回退。
3. 将正则错误回显改为本地化主提示 + 说明文案，将渠道名 fallback 改为 locale 文案。
4. 更新四语 locale，给 `verify-group-create-flow.mjs` 增加这两个兜底点的断言与禁止回退断言。
5. 运行无浏览器验证与 TypeScript 检查，记录下一轮入口。

## 9. 测试与验证

- `node scripts/verify-group-create-flow.mjs`
- `node scripts/verify-locale-consistency.mjs`
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`

## 10. 风险与兼容性

- 新风险：低；本轮仅收口兜底文案与脚本断言
- 兼容性风险：低；未修改后端 API、group 提交结构或拖拽排序逻辑
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（TypeScript noEmit）
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、automation memory、group 模块源码与静态脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账中的需求 `13`、`42`、`43`、`48` 明确要求中文界面不能继续泄漏 raw key 或无必要英文兜底。
  - 前端主线状态说明分组页主结构已完成，本轮更适合做小范围一致性补强，而不是重写交互。
  - automation memory 指向同一 `Phase G` 截图池，且浏览器级证据仍受宿主 CDP 阻塞，因此本轮优先选择可闭环的 no-browser 小切片。
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮优先完成源码与静态验证闭环；浏览器级 `375px / hover / focus` 证据仍受宿主 CDP 阻塞
- 待验证页面清单：分组创建/编辑弹窗真实浏览器 `375px` 布局、group create 同池截图回归
- 若未使用子 agent，原因：遵循本轮“不创建子 agent”的明确要求，且任务是同一模块的紧耦合小闭环
- worklog 是否更新：是
- 遗留项：
  - 分组创建/编辑弹窗浏览器级 `375px` 证据仍未补齐
  - 同一截图池中的 help-hint hover/focus 与 channel/group create dialog 浏览器证据仍待推进
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一截图优先池，优先补 group create dialog 浏览器级 `375px` 证据；若宿主浏览器链路仍阻塞，则切到同池的 help-hint hover/focus 或继续做剩余 no-browser 小闭环。
- 同主线候选顺序：
  1. group create dialog 浏览器级 `375px` / hover / focus 证据
  2. channel create dialog 浏览器级证据
  3. help-hint hover/focus 浏览器级或 no-browser 补强
  4. 同池剩余中文化与布局压缩问题
