# 2026-04-23 Phase G HelpHint 多语言 aria-label 与稳定选择器收口

## 1. 任务信息

- 任务名称：Phase G `HelpHint` 多语言 aria-label 与稳定选择器 no-browser 收口
- 日期：`2026-04-23`
- 当前阶段：`Phase G` 截图优先 UI 主线 / 设置页帮助提示静态护栏补强
- 对应 milestone：设置页帮助提示一致性与无障碍回归护栏补强

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6`、`9.6.1`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`9.3` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-backup-help-hint-closure.md`
  - `docs/worklog/2026-04-23-phase-g-ui-help-trigger-backup-test-sync-and-cache-values.md`
- 本次任务目标：
  - 继续留在 `Phase G screenshot-first` 同一主线，补强 `HelpHint` 的 no-browser 护栏。
  - 把帮助按钮默认 `aria-label` 从硬编码中文改成按当前语言读取 locale，避免英文/繁中/日文界面在无障碍层串语言。
  - 给帮助按钮补稳定 `data-slot`，让静态脚本与浏览器 smoke 不再绑死中文文案。
  - 同步备份页本地验证脚本与测试，避免帮助提示计数断言继续依赖中文按钮名。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-backup-help-hint-closure.md`
  - `docs/worklog/2026-04-23-phase-g-ui-help-trigger-backup-test-sync-and-cache-values.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/common/HelpHint.tsx`
  - `scripts/verify-help-hint-accessible.mjs`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `scripts/verify-backup-component.cjs`
  - `web/src/components/modules/setting/Backup.test.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 help-hint / backup worklog、automation memory、当前组件/脚本/测试源码
- 若未使用部分本地资源或上下文，原因：本轮只处理 `HelpHint` 与同池验证入口，不扩展到浏览器级 `375px` 或其他模块 UI
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`N/A`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：主线程执行
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮是单组文件的小闭环

## 3. 本次硬规则

- 只处理 `HelpHint` 多语言 aria-label、稳定选择器与同池 no-browser 验证脚本，不扩散到其他截图主题。
- 不能回退 `HelpHint` 的按钮语义和键盘可达性。
- 本轮必须形成“代码改动 + 直接验证 + worklog/状态同步 + 下一轮入口”闭环。

## 4. 本次禁止事项

- 不改业务接口语义。
- 不回滚工作区已有其他脏改动。
- 不把未执行的真实浏览器 hover/focus 证据记成完成。

## 5. 本次验收条件

- `HelpHint.tsx` 使用 `next-intl` 读取 `common.helpHint.ariaLabel`，不再硬编码中文默认标签。
- `HelpHint` 触发按钮带稳定 `data-slot="help-hint-trigger"`，tooltip 内容带稳定 `data-slot="help-hint-content"`。
- `scripts/verify-help-hint-accessible.mjs` 通过，并锁定 locale key / selector 契约。
- `scripts/verify-backup-component.cjs` 通过，并改为兼容新的帮助提示选择器。
- `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过。

## 6. 本次回滚点

- `web/src/components/common/HelpHint.tsx`
- `scripts/verify-help-hint-accessible.mjs`
- `scripts/verify-setting-help-browser-smoke.mjs`
- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `scripts/verify-backup-component.cjs`
- `scripts/verify-backup-component.next-intl-mock.cjs`
- `web/src/components/modules/setting/Backup.test.tsx`
- `web/public/locale/*.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 `HelpHint` 组件语义与 selector，再同步验证脚本与 locale
- 受影响后端模块：无
- 受影响前端模块：公共 `HelpHint`、设置页备份验证入口
- 受影响接口：无
- 是否影响旧数据：`否`
- 是否影响旧行为：仅补多语言 aria-label 与稳定选择器，不改用户可见主流程

## 8. 实施步骤

1. 复盘主规划、前端主线状态、最近 help-hint / backup worklog 与 automation memory，确认本轮继续留在 `Phase G screenshot-first` / help-hint 静态补强池。
2. 更新 `HelpHint.tsx`：接入 `next-intl`，把默认帮助按钮 aria-label 改成 locale 读取，并补 `data-slot` 标记与图标 `aria-hidden`。
3. 更新 `scripts/verify-help-hint-accessible.mjs`，把静态契约从“固定中文标签”改成“locale key + 稳定 selector”。
4. 更新设置页浏览器 smoke 脚本与备份页 no-browser 脚本/测试，让帮助提示选择器不再依赖中文按钮名，并同步 next-intl mock。
5. 运行静态验证、备份组件验证和前端 typecheck，确认本轮闭环。

## 9. 测试与验证

- 构建命令：`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：
  - `node .\scripts\verify-help-hint-accessible.mjs`
  - `node .\scripts\verify-backup-component.cjs`
- 专项验证：
  - 设置页帮助提示浏览器 smoke 脚本改用稳定 `data-slot` 选择器，不再绑死中文 `aria-label`
  - 备份页 no-browser 验证脚本和测试改为兼容新的 `HelpHint` 选择器与 mock 结果

## 10. 风险与兼容性

- 新风险：低；只影响帮助提示按钮的默认无障碍标签和测试选择器
- 兼容性风险：低；用户可见主文案与业务行为未变，新增的是更稳定的 locale / selector 契约
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过`
- 测试是否通过：`通过`
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 help-hint / backup worklog、automation memory、组件/脚本/测试源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与用户上下文要求帮助提示必须与真实实现一致，且多语言界面不能在更深层级串语言。
  - 前端主线状态与 automation memory 都把 `help-hint hover/focus` 静态补强列为当前同主线高位候选。
  - 最近 backup/help-hint worklog 说明上轮已经补了入口与无障碍 warning，本轮最值得继续补的是 locale / selector 契约。
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级 hover/focus 证据，本轮继续以 no-browser 静态护栏为准
- 待验证页面清单：
  - 设置页 `LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe` 真实浏览器 hover/focus 与 `375px`
  - 备份页真实浏览器 hover/focus 与 `375px`
  - channel/group create dialog 浏览器级 `375px` / hover / focus
- 若未使用子 agent，原因：遵循本轮“不创建子 agent”的明确要求，且任务范围小、上下文强耦合
- worklog 是否更新：`是`
- 遗留项：
  - 浏览器级 `help-hint hover/focus` 证据仍未恢复
  - 宿主 `Edge/CDP` 与 `vitest spawn EPERM` 既有阻塞仍在
- 下一任务前置条件是否满足：`满足`

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一截图优先池，优先补设置页或创建弹窗的真实浏览器 `hover / focus / 375px` 证据；若浏览器链路仍阻塞，则继续做同池 no-browser 小闭环。
- 同主线候选顺序：
  1. 设置页 help-hint 浏览器级 `hover / focus / 375px` 证据
  2. channel create dialog 浏览器级 `375px` / help-hint 证据
  3. group create dialog 浏览器级 `375px` / help-hint 证据
  4. 模型/价格区剩余英文主显示与布局收紧
