# 2026-04-21 Phase F Replace-Prune Helper Dedup Closure

## 1. 任务信息

- 任务名称：备份页 replace-prune 计数重复逻辑收口
- 日期：2026-04-21
- 当前阶段：Phase F / Milestone 6 validation closure
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入

- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5.4 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` Phase F
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-export-presentation-closure.md`
- 本次任务目标：去掉 `Backup.tsx` 中和共享逻辑重复的 replace-prune 计数函数，收敛到 `backup-logic.ts` 的单一来源
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、最近 Phase F worklog 链、`Backup.tsx`、`backup-logic.ts`、`scripts/verify-backup-logic.mjs`、`scripts/verify-backup-component.cjs`
- 本次使用的本地 resources / skills / 记忆上下文：读取了 `using-superpowers` 与 `brainstorming` 技能说明；复用了 automation memory 中“继续在 Phase F 上做可验证的小闭环”的结论；复用了最近 worklog 中关于 backup 页行为验证和文案收口的记录
- 若未使用部分本地资源或上下文，原因：没有继续展开更早的文档型任务，因为本轮已经可以直接从现有 helper 和 worklog 里定位到一个可收口的小重复逻辑
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围非常小，只涉及一个前端文件的重复逻辑清理

## 3. 本次硬规则

- 必须围绕 Phase F 备份 / 导入 / 回滚主线推进
- 必须留下真实代码增量和可执行验证
- 不扩大到无关 UI 清理或无关重构

## 4. 本次禁止事项

- 不改后端契约
- 不改导入 / 回滚语义
- 不回退用户已有改动
- 不把这轮写成纯记录不落代码

## 5. 本次验收条件

- `Backup.tsx` 不再维护一份独立的 replace-prune 计数实现
- 共享逻辑仍能正确支撑 import / rollback 的 replace-prune 统计
- web TypeScript 校验通过
- 备份页组件级验证通过

## 6. 本次回滚点

- 恢复 `web/src/components/modules/setting/Backup.tsx` 中被删除的本地 replace-prune 计数函数

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享逻辑依赖，再收回 UI 中重复函数
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否，行为仍由共享统计字段提供

## 8. 实施步骤

1. 核对 `Backup.tsx` 和 `backup-logic.ts` 里的 replace-prune 计数来源
2. 删除 `Backup.tsx` 中重复的本地计数函数
3. 跑 TypeScript、backup logic 验证和组件验证

## 9. 测试与验证

- 构建命令：无额外构建
- 测试命令：`node scripts/verify-backup-logic.mjs`，`node scripts/verify-backup-component.cjs`
- 专项验证：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`，`rg -n "countStructuredReplacePrunedItems" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/backup-logic.ts`

## 10. 风险与兼容性

- 新风险：低，只是把重复计数逻辑收回共享 helper 的既有路径
- 兼容性风险：低，统计来源没有变化
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：未单独跑完整构建
- 测试是否通过：通过，`verify-backup-logic.mjs` 和 `verify-backup-component.cjs` 都通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、最近 Phase F worklog、`Backup.tsx`、`backup-logic.ts`、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 提醒继续做可验证的小闭环；canonical plan 和 workflow 限定本轮只做 Phase F；worklog 链表明 backup 页仍在做可收口的细小一致性修正；skills 说明先做上下文盘点再落代码
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮目标是代码收口和命令级验证，不需要浏览器级 smoke
- 待验证页面清单：backup 页行为级 smoke 仍是 Phase F 后续优先项
- 若未使用子 agent，原因：任务太小且用户明确要求主线程执行
- worklog 是否更新：是
- 遗留项：Phase F 还缺浏览器或手工 smoke 证据
- 下一任务前置条件是否满足：是，当前可以继续做 backup 页的下一层行为证据
