# 2026-04-23 Phase G 动态路由未归档摘要消息护栏收口

## 1. 任务信息

- 任务名称：动态路由摘要未归档消息中文主显示护栏收口
- 日期：2026-04-23
- 当前阶段：Phase G 截图优先 UI 主线 / 设置页中文化一致性补强
- 对应 milestone：P0 设置页动态路由中文化与帮助提示一致性

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`11.2` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-dynamic-routing-progressive-help-closure.md`
  - `docs/worklog/2026-04-23-phase-g-dynamic-routing-summary-message-localization-closure.md`
- 本次任务目标：
  - 收口 `DynamicRouting` 摘要区对未归档后台消息的中文主显示，避免中文界面继续把英文或内部原句直接顶到主消息行。
  - 保留原始后台消息作为次级详情，兼顾排障可读性与中文主显示一致性。
  - 同步补齐四语 locale、静态脚本、测试和状态文档，避免代码与文档口径再次漂移。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/DynamicRouting.test.tsx`
  - `scripts/verify-dynamic-routing-help.mjs`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/en.json`
  - `web/public/locale/ja.json`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、已有动态路由 worklog、automation memory、当前组件/测试/脚本源码
- 若未使用部分本地资源或上下文，原因：本轮只处理设置页动态路由单模块护栏，不需要扩展到浏览器 smoke、后端实现或备份页逻辑
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：N/A
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮是同一模块的紧耦合 no-browser 小闭环

## 3. 本次硬规则

- 只处理 `DynamicRouting` 摘要消息的中文主显示护栏，不改后端摘要接口语义。
- 中文界面主消息行不能继续直接裸露未归档英文后台原句。
- 未归档消息的原始内容仍要以次级详情保留下来，不能为了中文化直接丢失排障线索。
- 本轮必须留下静态脚本和类型检查级别的可重复验证，不把未跑的浏览器验证记成完成。

## 4. 本次禁止事项

- 不扩散到备份页、模型价格页、`CC Switch` 或 channel/group create dialog。
- 不回滚工作区中已有的其他脏改动。
- 不把宿主上的 `vitest spawn EPERM` 记成代码逻辑失败。

## 5. 本次验收条件

- `DynamicRouting` 对未归档后台摘要消息改为“本地化主提示 + 原始详情”双层展示。
- 四语 locale 补齐 `unmapped` 与 `messageDetailLabel` 文案。
- `scripts/verify-dynamic-routing-help.mjs`、`scripts/verify-locale-consistency.mjs`、`tsc --noEmit -p web/tsconfig.json` 通过。
- 状态文档不再继续把动态路由卡片描述为“仍有整段英文说明未清理”。

## 6. 本次回滚点

- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/components/modules/setting/DynamicRouting.test.tsx`
- `scripts/verify-dynamic-routing-help.mjs`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `web/public/locale/ja.json`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 UI 展示护栏与 locale，再同步测试和静态脚本
- 受影响后端模块：无
- 受影响前端模块：设置页动态路由卡片
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅调整未归档摘要消息的展示方式，已知 message key 与摘要结构保持兼容

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态与已有动态路由 worklog，确认本轮继续留在 `Phase G` 同一截图池。
2. 审查 `DynamicRouting.tsx` 当前逻辑，确认已知 message key 已本地化，但未知消息仍直接落到中文主消息行。
3. 将未知摘要消息改成统一中文主提示，并把原始后台原句降级到详情行；同步补四语 locale、测试与静态脚本。
4. 更新 `CURRENT_STATUS_AND_PLAN` 与 `FRONTEND_UI_MAINLINE_STATUS`，同步收掉动态路由“仍有整段英文说明”的旧口径。
5. 运行直接相关验证，记录 `vitest spawn EPERM` 宿主阻塞与下一轮入口。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-dynamic-routing-help.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 专项验证：
  - `node web/node_modules/vitest/vitest.mjs run web/src/components/modules/setting/DynamicRouting.test.tsx --config web/vitest.config.ts`
  - 结果仍受宿主 `spawn EPERM` 限制，失败发生在 `vite/esbuild` 启动阶段，不属于本轮代码逻辑回归

## 10. 风险与兼容性

- 新风险：低；本轮仅增加未归档消息护栏与详情展示
- 兼容性风险：低；已知 key 仍维持原本地化映射，未归档消息的原始原句仍可见
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（`tsc --noEmit`）
- 测试是否通过：通过（静态脚本与 locale 一致性）；Vitest 因宿主 `spawn EPERM` 未通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、动态路由已有 worklog、automation memory、组件/测试/脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账中的需求 `08`、`10`、`42`、`49` 要求中文界面不能继续泄漏英文主文案，动态路由属于点名问题区。
  - 现有动态路由 worklog 说明“已知消息 key 本地化”已完成，因此本轮最值得补的是未归档消息护栏，而不是重做整个卡片。
  - automation memory 明确当前仍在 `Phase G screenshot-first` 池，且浏览器/CDP 阻塞未解，因此优先做 no-browser 小闭环。
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` 证据链与 `vitest` 子进程仍有既有阻塞；本轮优先完成源码与静态验证闭环
- 待验证页面清单：
  - 设置页动态路由卡片真实浏览器 hover/focus 与 `375px` 布局
  - 同池的 `help-hint` hover/focus 浏览器级证据
  - 备份页剩余英文文案清理后的 no-browser / browser 证据
- 若未使用子 agent，原因：遵循本轮“不创建子 agent”的明确要求，且任务范围小、上下文强耦合
- worklog 是否更新：是
- 遗留项：
  - `DynamicRouting.test.tsx` 已补新断言，但本机 `vitest` 仍受 `spawn EPERM` 限制，暂无新的代码级失败
  - 备份页中文主文案残留仍在同一截图池高位，适合作为下一轮 no-browser 相邻任务
  - 动态路由卡片浏览器级 hover/focus / `375px` 证据仍未补齐
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一截图优先池，优先收口备份页剩余英文主文案与帮助提示一致性；若想继续设置页同主题，也可以补 `help-hint` hover/focus 的静态断言强化
- 同主线候选顺序：
  1. 备份页英文主文案残留 no-browser 收口
  2. `help-hint` hover/focus 静态断言补强
  3. 动态路由/模型探测卡片浏览器级 `375px` 与 hover/focus 证据
  4. channel/group create dialog 浏览器级证据
