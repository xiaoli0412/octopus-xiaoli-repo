# 2026-04-23 Phase G 备份页帮助提示入口收口

## 1. 任务信息

- 任务名称：Phase G 备份页帮助提示入口与解析阻塞收口
- 日期：`2026-04-23`
- 当前阶段：`Phase G` 截图优先 UI 主线 / 设置页备份 no-browser 收口
- 对应 milestone：备份页帮助提示入口可见化与验证同步

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-backup-provider-copy-closure.md`
  - `docs/worklog/2026-04-23-deep-audit-backup-verification-locale-contract.md`
- 本次任务目标：
  - 继续留在 `Phase G screenshot-first` 同一主线，收口备份页关键区域的 `HelpHint` 可见入口。
  - 修复 `Backup.tsx` 中误留下的损坏帮助文本块，恢复组件解析与验证链路。
  - 同步脚本、测试、状态文档和 worklog，给下一轮留下明确入口。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-backup-provider-copy-closure.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/setting/Backup.tsx`
  - `web/src/components/modules/setting/Backup.test.tsx`
  - `scripts/verify-backup-component.cjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 backup worklog、automation memory、当前 backup 组件/测试/脚本源码
- 若未使用部分本地资源或上下文，原因：本轮只处理设置页备份单模块 no-browser 小闭环，不扩展到浏览器 smoke、后端导入逻辑或其他设置卡片
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`N/A`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：主线程执行
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务范围小、上下文强耦合

## 3. 本次硬规则

- 只处理备份页帮助提示入口与解析阻塞，不扩散到其他设置模块。
- 英文隐藏测试锚点继续保留，不回退现有 no-browser 验证契约。
- 本轮必须形成“代码变更 + 直接验证 + 状态记录 + 下一轮入口”闭环。

## 4. 本次禁止事项

- 不改后端导入导出接口语义。
- 不回滚工作区中已有的其他脏改动。
- 不把宿主浏览器/CDP 阻塞当作本轮失败原因。

## 5. 本次验收条件

- `Backup.tsx` 恢复可解析，且只保留干净的 `HELP_TEXT`。
- 备份页关键区域 `HelpHint` 按钮数与脚本/测试断言对齐。
- `node .\scripts\verify-backup-component.cjs` 通过。
- `node .\scripts\verify-backup-logic.mjs` 通过。
- `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过。

## 6. 本次回滚点

- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先修 `Backup.tsx` 的帮助提示 UI 与解析阻塞，再同步验证与文档
- 受影响后端模块：无
- 受影响前端模块：设置页备份卡片
- 受影响接口：无
- 是否影响旧数据：`否`
- 是否影响旧行为：仅增强帮助提示入口与恢复解析，不改变导入导出主流程语义

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态、automation memory 和最近 backup worklog，确认本轮继续留在 `Phase G screenshot-first` / backup no-browser 池。
2. 审查 `Backup.tsx`，定位误插入的损坏 `BROKEN_HELP_TEXT` 块，并清理到只保留干净的 `HELP_TEXT`。
3. 保持已接入的 `HelpHint` 结构不变，确认 export/import/replace-prune/rollback/advanced pending 区域都保留问号入口。
4. 运行 `verify-backup-component`、`verify-backup-logic` 与 `tsc --noEmit`，确认 no-browser 验证恢复通过。
5. 同步状态文档、worklog 与 automation memory，固定下一轮候选顺序。

## 9. 测试与验证

- 构建命令：`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：
  - `node .\scripts\verify-backup-component.cjs`
  - `node .\scripts\verify-backup-logic.mjs`
- 专项验证：
  - 复核 `Backup.tsx` 帮助提示按钮计数基线：默认 `8`、`map` 模式 `9`、历史回滚展开 `9`

## 10. 风险与兼容性

- 新风险：低；本轮只处理备份页前端帮助提示与解析阻塞
- 兼容性风险：低；主流程与接口契约未变，仅补充问号帮助入口并修复错误残留
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过`
- 测试是否通过：`通过`
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 backup worklog、automation memory、backup 组件/测试/脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与工作流确认本轮必须继续沿 `Phase G screenshot-first` 同一主线推进，不能扩散。
  - 当前状态文档与前端主线状态确认备份页仍属于设置页中文化与帮助提示优先池。
  - automation memory 提醒上轮已完成 provider-copy no-browser 收口，本轮最值得推进的是 help-hint 入口与解析恢复。
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级证据，本轮继续以 no-browser 闭环为准
- 待验证页面清单：
  - 设置页备份卡片真实浏览器 hover/focus 与 `375px` 布局
  - 同池的 `help-hint` hover/focus 浏览器级证据
  - channel/group create dialog 浏览器级证据
- 若未使用子 agent，原因：遵循本轮“不创建子 agent”的明确要求，且任务范围小、上下文强耦合
- worklog 是否更新：`是`
- 遗留项：
  - 浏览器级 help-hint hover/focus 证据仍未恢复
  - 宿主 `Edge/CDP` 与 `vitest spawn EPERM` 既有阻塞仍在
- 下一任务前置条件是否满足：`满足`

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一截图优先池，优先补 `help-hint hover/focus` 的 no-browser 静态断言强化或浏览器级证据；若浏览器链路仍阻塞，则转到同池的 channel/group create dialog 证据收口。
- 同主线候选顺序：
  1. `help-hint` hover/focus 静态断言补强
  2. 备份页真实浏览器 hover/focus 与 `375px` 证据
  3. channel create dialog 浏览器级证据
  4. group create dialog 浏览器级证据
