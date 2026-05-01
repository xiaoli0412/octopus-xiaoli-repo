# 2026-04-21 Phase F Empty Selective Rollback Fallback

## 1. 任务信息

- 任务名称：空 selective rollback scope 回退为 full restore
- 日期：2026-04-21
- 当前阶段：Phase F / Milestone 6 validation closure
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入

- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5.4 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-backup-history-refresh-lock.md`
- 本次任务目标：收紧 Backup 页的 selective rollback 边界，当 selectiveImport 打开但没有任何 active scopes 时，回滚行为回退为 full snapshot restore，而不是向 API 发送空 scope 对象
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、最近 Phase F worklog 链、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`scripts/verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：复用了 automation memory 中“继续在 Phase F 上补可验证的小闭环”的结论；复用了前端主线状态文档中对备份 / 导入 / 回滚仍需继续收口的判断；复用了现有 Backup 组件和 no-spawn 验证脚本作为最小闭环入口
- 若未使用部分本地资源或上下文，原因：没有继续展开更早的后端导入语义 worklog，因为本轮已经可以直接在现有 Backup 页面和脚本验证链里落一个小闭环
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务很小，只涉及同一页面内的一个行为边界

## 3. 本次硬规则

- 必须围绕 Phase F 备份 / 导入 / 回滚主线推进
- 必须留下真实代码增量和可执行验证
- 不能扩散到无关 UI 清理或后端大改

## 4. 本次禁止事项

- 不扩大到新的备份功能设计
- 不修改后端导入契约
- 不回退用户已有改动
- 不把这轮做成纯记录不落代码

## 5. 本次验收条件

- selectiveImport 开启且没有 active scopes 时，rollback 预览与回滚都按 full restore 处理
- no-spawn 验证脚本可断言该 fallback
- Vitest 草稿同步同一行为
- `node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`、`tsc --noEmit` 通过

## 6. 本次回滚点

- 恢复 `web/src/components/modules/setting/Backup.tsx` 中的 selective rollback fallback
- 恢复 `scripts/verify-backup-component.cjs` 中对应的验证
- 恢复 `web/src/components/modules/setting/Backup.test.tsx` 中对应的 Vitest 草稿

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 UI fallback，再用验证脚本和 Vitest 草稿证明回退语义一致
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口：无契约变化；仅避免 UI 发送空 `import_scopes`
- 是否影响旧数据：否
- 是否影响旧行为：否，empty selective scope 之前没有清晰语义，这次显式回退为 full restore

## 8. 实施步骤

1. 调整 `Backup.tsx` 中 `rollbackImportScopes` 的推导，要求 selective import 至少存在一个 active scope 才发送 scope override
2. 在 `scripts/verify-backup-component.cjs` 中补充 empty selective scope fallback 验证
3. 在 `web/src/components/modules/setting/Backup.test.tsx` 中同步补同一条行为草稿
4. 运行 Backup 组件验证、逻辑验证和前端 typecheck

## 9. 测试与验证

- 构建命令：无额外构建
- 测试命令：`node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`
- 专项验证：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 10. 风险与兼容性

- 新风险：单进程验证脚本对 Backup 页结构仍然敏感，后续若调整按钮文案或回滚区布局，需要同步更新断言
- 兼容性风险：低；没有改动后端契约或数据结构，只调整了 selective rollback 的 fallback 规则
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：未单独跑完整构建
- 测试是否通过：通过；`verify-backup-component.cjs`、`verify-backup-logic.mjs`、`tsc --noEmit` 都通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、最近 Phase F worklog、`Backup.tsx`、`Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 提醒继续补可验证的小闭环；canonical plan 和 workflow 限定本轮仍在 Phase F；前端主线状态文档确认备份 / 导入 / 回滚还差一层收口；现有验证脚本提供了最小可执行闭环
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮优先补可重复执行的单进程行为验证，浏览器手工 smoke 仍未接通
- 待验证页面清单：backup 页真实文件上传与浏览器交互、日志页实时流、首页窄屏细节
- worklog 是否更新：是
- 遗留项：当前仍未形成真正浏览器级的 Backup 导入 / 回滚自动化；下一轮更适合继续补同页面内的更高层行为证据，或者在同主线下转去 browser smoke
- 下一任务前置条件是否满足：是
- 记录时间：2026-04-21T20:04:12.539+08:00

## 12. 本轮补充记录

- 记录时间：2026-04-21T20:04:12.539+08:00
- 本轮实际完成的改动
  - 在 [Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx) 里把 rollback selective scope 的派生规则收紧为“selectiveImport 开启且至少有一个 active scope”才发送 scope override
  - 在 [scripts/verify-backup-component.cjs](/D:/GPT-codex/octopus_repo/scripts/verify-backup-component.cjs) 里补了 empty selective rollback fallback 验证，确认 preview / rollback 都会回退为 full restore
  - 在 [web/src/components/modules/setting/Backup.test.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.test.tsx) 里同步补了相同行为草稿
  - 更新了 [docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md) 的 Phase F 状态说明
- 已执行的验证
  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 当前仍存在的阻塞或风险
  - 浏览器级手工 smoke 仍未执行
  - `backup-logic.ts` 的 `MODULE_TYPELESS_PACKAGE_JSON` 警告仍然非阻塞
- 下一轮最适合继续推进的具体事项
  - 继续留在 Phase F 的同一 Backup 页上，只做下一层可验证的导入 / 回滚行为证据；如果没有，就转向 browser smoke 证据


## 13. 本轮补充记录

- 记录时间: 2026-04-21T20:38:14.0294284+08:00
- 本轮实际检查
  - 复查了 Backup.tsx、Backup.test.tsx、scripts/verify-backup-component.cjs、ackup-logic.ts
  - 重新运行了 
ode scripts/verify-backup-component.cjs
  - 重新运行了 
ode scripts/verify-backup-logic.mjs
  - 重新运行了 D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json
- 本轮结果
  - 本次无变更
  - 先前怀疑的 rollback toggle 锁定问题在当前工作区里已是绿色
  - 当前仍缺浏览器级手工 smoke
- 下一步
  - 继续留在 Phase F / Backup 页，找一个真正的新行为证据点，或转去 browser smoke 证据