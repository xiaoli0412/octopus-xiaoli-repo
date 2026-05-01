# 2026-04-22 Phase F Backup Remaining Tooling Shared-Source Closure

## 1. 任务信息
- 任务名称：Backup 页尾部 remaining migration tooling 分组共享化
- 日期：2026-04-22
- 当前阶段：Phase F / backup-import-rollback frontend closure
- 对应 milestone：Milestone 6 validation and deployment

## 2. 开工前输入
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-f-backup-verification-esm-warning-closure.md`
- 本次任务目标：把 Backup 页底部 remaining migration tooling 的分组结构也抽成共享逻辑，并同步逻辑测试和 Node 验证脚本
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、`docs/worklog` 里的 Phase F 历史记录、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/backup-logic.ts`、`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`、`scripts/verify-backup-component.cjs`
- 本次使用的本地 resources / skills / 记忆上下文：复用了 automation memory 中“继续在 Phase F 做最小同主线闭环”的结论；复用了 canonical plan 与 workflow 对 Phase F 只能做可验证小闭环的约束；复用了前端主线状态文档里 Backup 页仍处于部分完成、仍可继续做同页收口的判断；复用了 `using-superpowers` 和 `brainstorming`
- 若未使用部分本地资源或上下文，原因：未继续展开浏览器手工 smoke，因为本轮目标是把共享逻辑与脚本验证先收口
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行

## 3. 本次硬规则
- 必须围绕 Phase F backup/import/rollback 主线推进
- 必须产生真实代码增量和验证增量
- 不改后端契约，不扩散到无关页面

## 4. 本次禁止事项
- 不做无关清理
- 不改后端导入语义
- 不把 pending 说明写成已完成能力

## 5. 本次验收条件
- Backup 页底部 remaining migration tooling 的分组结构只保留一个共享来源
- Backup 页、逻辑测试、Node 验证脚本都引用同一份数据
- `node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`、`tsc --noEmit` 通过

## 6. 本次回滚点
- 回退 `web/src/components/modules/setting/backup-logic.ts`
- 回退 `web/src/components/modules/setting/Backup.tsx`
- 回退 `web/src/components/modules/setting/backup-logic.test.ts`
- 回退 `scripts/verify-backup-logic.mjs`

## 7. 实现范围
- 先改数据语义还是先改 UI：先改共享逻辑，再让 UI 只消费共享结果
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`
- 受影响逻辑模块：`web/src/components/modules/setting/backup-logic.ts`
- 受影响测试：`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`、`scripts/verify-backup-component.cjs`
- 是否影响旧数据：否
- 是否影响旧行为：否，只是把页脚分组来源抽离出来

## 8. 实施步骤
1. 把 remaining migration tooling 分组结构抽到 `backup-logic.ts`
2. 让 `Backup.tsx` 从共享函数拿页脚清单，再直接渲染
3. 在 `backup-logic.test.ts` 和 `scripts/verify-backup-logic.mjs` 里补同一条共享来源断言
4. 跑 backup 逻辑、组件验证和 web typecheck
5. 记录结果并更新 automation memory

## 9. 测试与验证
- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`node scripts/verify-backup-component.cjs`
- 专项验证：`node scripts/verify-backup-logic.mjs`

## 10. 风险与兼容性
- 新风险：低，仅是共享来源抽离
- 兼容性风险：低，不影响导入、回滚、快照数据结构
- 是否阻塞下一任务：否

## 11. 收工记录
- 构建是否通过：已通过 `tsc --noEmit`
- 测试是否通过：已通过 `node scripts/verify-backup-component.cjs` 和 `node scripts/verify-backup-logic.mjs`
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、workflow、前端主线状态文档、Phase F worklog chain、Backup 相关源码与验证脚本、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 继续把本轮锁定在 Phase F 的最小闭环；canonical plan 和 workflow 限定了本轮只能做可验证的小收口；前端主线状态文档确认 Backup 页仍然是部分完成，适合继续做同页收口；验证脚本给出了稳定的共享逻辑与验收入口
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先收口共享逻辑与可重复验证，未切到浏览器自动化
- 待验证页面清单：Backup 页
- 若未使用子 agent，原因：用户明确要求主线程执行
- worklog 是否更新：是
- 遗留项：仍缺浏览器级手工 smoke
- 下一任务前置条件是否满足：是，当前修改已验证通过，下一轮可以继续找同一 Backup 页里更小的行为缺口，或者切到 browser/manual smoke 证据
