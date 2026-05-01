# 2026-04-22 Phase F Import Mode Change Pending Apply Reset Verification

## 1. 任务信息

- 任务名称 Backup 导入模式切换后 pending apply 失效验证补强
- 日期 2026-04-22
- 当前阶段 Phase F / backup-import-rollback frontend closure
- 对应 milestone Milestone 6 validation and deployment

## 2. 开工前输入

- 对应 canonical 章节 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节 `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog `docs/worklog/2026-04-22-phase-f-backup-empty-scope-verification-fix.md`
- 本次任务目标 为 Backup 页补齐“导入模式切换会使挂起的 Apply This Dry-Run 绑定失效”这条行为证据，并保持 no-spawn 验证链一致
- 本次已盘点本地资源 automation memory、canonical plan、workflow、前端主线状态文档、`Backup.tsx`、`Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`scripts/verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文 延续 automation memory 里“继续在 Phase F 做最小同主线闭环”的结论；复用既有 Backup 组件验证脚本作为主验收入口；复用 worklog 链条中此前关于 pending apply 绑定和文件重选失效的上下文
- 若未使用部分本地资源或上下文 本轮没有继续展开浏览器手工 smoke，因为当前最小可闭环缺口是同页状态失效分支验证
- 本次是否启用子 agent 与分工边界 否
- 本次子 agent 使用模型 无
- 若未使用子 agent，原因 用户要求主线程执行且禁止创建子 agent

## 3. 本次硬规则

- 只在 Phase F backup/import/rollback 主线推进
- 每轮必须产生真实代码增量与可执行验证
- 不扩散到后端契约或其他页面

## 4. 本次禁止事项

- 不改导入/回滚后端语义
- 不做无关 UI 清理
- 不把 no-spawn 验证误报为浏览器级 e2e

## 5. 本次验收条件

- Vitest 草稿中新增导入模式切换导致 pending apply 失效并需重新 dry-run 的断言
- no-spawn 组件验证脚本覆盖同一行为路径
- `verify-backup-component`、`verify-backup-logic`、`tsc --noEmit` 全部通过

## 6. 本次回滚点

- 回退 `web/src/components/modules/setting/Backup.test.tsx`
- 回退 `scripts/verify-backup-component.cjs`

## 7. 实现范围

- 先改数据语义还是先改 UI 不改产品语义，只补验证覆盖
- 受影响后端模块 无
- 受影响前端模块 `web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口 无
- 是否影响旧数据 否
- 是否影响旧行为 否，仅增加测试与脚本断言

## 8. 实施步骤

1. 复查 `Backup.tsx` 中 `onImportModeChange -> resetImportFormState` 的状态清理链，确认目标行为是既有语义
2. 在 `Backup.test.tsx` 增加导入模式切换分支：先生成 pending apply，再切换到 `Map`，断言 pending apply 消失并重新 dry-run 产出 map 模式绑定
3. 在 `scripts/verify-backup-component.cjs` 同步新增同路径验证，并接入主执行入口
4. 运行 no-spawn 组件验证、逻辑验证与 web typecheck

## 9. 测试与验证

- 构建命令 `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令 `node scripts/verify-backup-component.cjs`
- 专项验证 `node scripts/verify-backup-logic.mjs`

## 10. 风险与兼容性

- 新风险 低；新增断言会对结果面板和 import mode 交互结构更敏感
- 兼容性风险 低；本轮没有产品逻辑和接口语义变更
- 是否阻塞下一任务 否

## 11. 收工记录

- 构建是否通过 是，`tsc --noEmit` 通过
- 测试是否通过 是，`verify-backup-component` 与 `verify-backup-logic` 通过
- 本次使用了哪些本地资源 / skills / 记忆上下文 automation memory、canonical plan、workflow、前端主线状态文档、Backup 组件与验证脚本、Phase F worklog 链
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论 memory 确认继续 Phase F 小闭环；canonical/workflow 约束本轮不得扩散；主线状态文档确认 Backup 页仍有可继续验证的边界项；既有脚本提供可重复验收入口
- 本次使用了哪些子 agent 及其结论 无
- 子 agent 分工、负责范围与产出摘要 无
- 手工 smoke 状态 未执行
- 手工 smoke 阻塞原因 / 缺少的环境 当前线程仍无浏览器自动化入口，且本机 Vitest 启动仍受 esbuild spawn EPERM 影响
- 待验证页面清单 Backup 页浏览器级手工 smoke
- 若未使用子 agent，原因 用户要求主线程执行且禁止创建子 agent
- worklog 是否更新 是
- 遗留项 浏览器级 Backup smoke 证据仍缺失
- 下一任务前置条件是否满足 是，当前 same-mainline 验证链仍可继续增量收口
- 记录时间 2026-04-22T02:35:00+08:00
