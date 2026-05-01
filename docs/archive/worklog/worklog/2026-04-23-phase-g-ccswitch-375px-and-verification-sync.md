# 2026-04-23 Phase G CC Switch 375px 与验证口径同步收口

## 1. 任务信息

- 任务名称：`CC Switch` 移动端 375px 布局收紧与静态验证同步
- 日期：`2026-04-23`
- 当前阶段：`Phase G` 截图优先 UI 主线
- 对应 milestone：`P0-B` 同主线小闭环补强

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-ccswitch-progressive-unlock-closure.md`
  - `docs/worklog/2026-04-23-phase-g-setting-help-attached-session-timeout-order-contrast.md`
- 本次任务目标：
  - 收紧 `CC Switch` 在 `375px` 场景最容易挤压的弹窗头部、tab 与客户端选择按钮布局。
  - 让静态验证脚本对齐当前真实门控，补上 `hasGroupOptions` 与 import blocked hint 的断言，避免“实现已变但验证还停在旧逻辑”。
  - 保持当前 `CC Switch` 渐进展开语义不变，不扩散到其他截图主题。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-ccswitch-progressive-unlock-closure.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/navbar/DocModal.tsx`
  - `scripts/verify-ccswitch-flow.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：canonical plan、用户上下文总账、前端主线状态、上一轮 `CC Switch` worklog、automation memory、当前 `DocModal` 与静态脚本源码
- 若未使用部分本地资源或上下文，原因：本轮只收口 `CC Switch` 单切片，不需要扩展到浏览器 smoke 脚本或其他页面组件
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`N/A`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：`否`
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务是同一模块的紧耦合小闭环

## 3. 本次硬规则

- 只处理 `CC Switch` 的移动端布局收紧与静态验证同步，不更改深链协议和后端接口。
- 帮助提示与 blocked hint 必须与当前真实代码状态一致，不沿用旧脚本口径。
- 本轮必须保留可重复执行的无浏览器验证链，不把纯样式收紧写成已验收截图问题。

## 4. 本次禁止事项

- 不回滚仓库中已有的无关脏改动。
- 不把本轮扩散到 channel/group create dialog 或设置页帮助提示浏览器 smoke。
- 不把未执行的真实浏览器 `375px` 截图验证伪装成已完成。

## 5. 本次验收条件

- `DocModal` 的头部、tab 与 `CC Switch` 客户端按钮在源码层具备更适合 `375px` 的响应式约束。
- `verify-ccswitch-flow.mjs` 对齐当前真实 `hasGroupOptions` / blocked hint / import hint 逻辑。
- `node scripts/verify-ccswitch-flow.mjs`、`node scripts/verify-locale-consistency.mjs`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过。

## 6. 本次回滚点

- `web/src/components/modules/navbar/DocModal.tsx`
- `scripts/verify-ccswitch-flow.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先确认静态脚本与真实实现漂移点，再一起收口 UI 与脚本
- 受影响后端模块：`无`
- 受影响前端模块：`DocModal / CC Switch`
- 受影响接口：`无`
- 是否影响旧数据：`否`
- 是否影响旧行为：仅收紧移动端布局和静态校验，不改变 `CC Switch` 深链语义

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态与上一轮 `CC Switch` worklog，确认本轮继续留在 `Phase G` 同一截图池。
2. 复盘 `DocModal.tsx` 与 `verify-ccswitch-flow.mjs`，识别脚本仍停留在旧 `selectedApiKey !== ''` 门控，而真实实现已新增 `ccswitchHasGroupOptions` 与 import blocked hint 分支。
3. 调整 `DocModal` 的头部 padding、tab 容器横向滚动、内容区 padding、客户端按钮与导入按钮尺寸与换行策略，使 `375px` 下更不容易横向挤爆。
4. 更新静态脚本，补 `hasGroupOptions`、blocked hint、import blocked hint 以及移动端类名断言。
5. 运行本轮相关无浏览器验证与 TypeScript 检查，并记录下一轮入口。

## 9. 测试与验证

- 测试命令：
  - `node .\scripts\verify-ccswitch-flow.mjs`
  - `node .\scripts\verify-locale-consistency.mjs`
  - `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 专项验证：
  - `verify-ccswitch-flow.mjs` 现在要求 `ccswitchHasGroupOptions`、`ccswitchModelBlockedHint`、`ccswitchImportBlockedHint` 存在并被渲染。
  - `DocModal` 头部、tabs、客户端按钮、导入按钮的移动端类名已固定到脚本断言中。

## 10. 风险与兼容性

- 新风险：低；本轮仅调整弹窗排版与静态脚本
- 兼容性风险：低；未修改后端 API、深链参数或数据结构
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过（TypeScript noEmit）`
- 测试是否通过：`通过`
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、前端主线状态、上一轮 `CC Switch` worklog、automation memory、`DocModal` 与静态脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账与前端主线状态明确指出 `CC Switch / 375px / help-hint` 仍在 Phase G screenshot-first 候选池高位。
  - 上一轮 `CC Switch` worklog 已说明交互流程型重做完成，这轮更适合补移动端收紧与验证口径同步，而不是重新改交互语义。
  - automation memory 提醒浏览器证据仍受宿主 `Edge/CDP` 阻塞，因此本轮优先选择可闭环的 no-browser 小切片。
- 本次使用了哪些子 agent 及其结论：`未使用`
- 手工 smoke 状态：`未执行真实浏览器 375px smoke`
- 手工 smoke 阻塞原因 / 缺少的环境：浏览器级证据仍受当前宿主 `Edge/CDP` bootstrap 阻塞；本轮只完成无浏览器闭环与移动端源码约束收紧
- 待验证页面清单：
  - `DocModal` 的 `CC Switch` 标签真实 `375px` 截图
  - `CC Switch` 的 hover/focus 行为
  - channel/group create dialog 的同池浏览器级证据
- 若未使用子 agent，原因：用户要求主线程执行，且本轮任务范围小、上下文强耦合
- worklog 是否更新：`是`
- 遗留项：
  - `CC Switch` 浏览器级 `375px` 和 hover/focus 证据仍未补齐
  - 同一截图池中的 channel/group create dialog 浏览器级验证仍待推进
- 下一任务前置条件是否满足：`满足`

## 12. 下一轮建议

- 下一轮最适合继续推进：在同一 `Phase G` 主线下优先补 `CC Switch` 真实浏览器 `375px / hover / focus` 证据；若宿主浏览器阻塞仍未解开，则切换到 channel/group create dialog 的同池可闭环项
- 同主线候选顺序：
  1. `CC Switch` 浏览器级 `375px` / hover / focus 证据
  2. channel create dialog 浏览器级证据
  3. group create dialog 浏览器级证据
  4. help-hint hover/focus 浏览器级断言补强
