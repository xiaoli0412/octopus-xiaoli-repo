# 2026-04-21 Phase F Remaining Migration Tooling Copy Alignment

## 1. 任务信息
- 任务名称：Backup 页脚剩余能力清单抽到共享逻辑
- 日期：2026-04-21
- 当前阶段：Phase F / Milestone 6 validation closure
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-rollback-domain-copy-alignment.md`
- 本次任务目标：把 Backup 页脚的 remaining migration tooling 文案抽到共享逻辑源，避免页面、测试和脚本各自维护三份重复清单
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/backup-logic.ts`、`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`、`scripts/verify-backup-component.cjs`
- 本次使用的本地 resources / skills / 记忆上下文：复用 automation memory 中“继续在 Phase F 做可验证的小闭环”的结论；复用前端主线状态文档中“Backup 页仍可继续做同页小闭环但不应扩散”的判断；复用 existing backup verification scripts 作为稳定验证入口
- 若未使用部分本地资源或上下文，原因：未继续展开更早的 Phase F copy alignment worklog，因为本轮目标已经收束到同一份清单的单一来源
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮是单页单逻辑源的小闭环

## 3. 本次硬规则
- 必须围绕 Phase F backup / import / rollback 主线推进
- 必须留下真实代码增量和可执行验证，不得只写总结
- 不改后端契约，不扩散到无关页面

## 4. 本次禁止事项
- 不扩散到其他页面或无关 UI 清理
- 不改后端导入语义
- 不把共享文案源扩大成新功能设计

## 5. 本次验收条件
- Backup 页脚剩余能力清单只保留一个共享来源
- Backup 页、逻辑测试、Node 验证脚本都引用同一份数据
- `node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`、`tsc --noEmit` 通过

## 6. 本次回滚点
- 回退 `web/src/components/modules/setting/backup-logic.ts`
- 回退 `web/src/components/modules/setting/Backup.tsx`
- 回退 `web/src/components/modules/setting/backup-logic.test.ts`
- 回退 `scripts/verify-backup-logic.mjs`

## 7. 实现范围
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`
- 受影响逻辑模块：`web/src/components/modules/setting/backup-logic.ts`
- 受影响测试：`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`、`scripts/verify-backup-component.cjs`
- 是否影响旧数据：否
- 是否影响旧行为：否，只是把共享文案源抽离出来

## 8. 实施步骤
1. 把 remaining migration tooling 文案抽到 `backup-logic.ts`
2. 让 `Backup.tsx` 从共享函数拿页脚清单，再按分组切片展示
3. 在 `backup-logic.test.ts` 和 `scripts/verify-backup-logic.mjs` 里补同一条共享来源断言
4. 跑 backup 逻辑、组件验证和 web typecheck
5. 记录结果并更新 automation memory

## 9. 测试与验证
- 实际执行命令：`node scripts/verify-backup-logic.mjs`
- 实际执行命令：`node scripts/verify-backup-component.cjs`
- 实际执行命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 结果：全部通过

## 10. 风险与兼容性
- 风险低，仅是共享来源抽离
- 不影响导入、回滚、快照数据结构
- `MODULE_TYPELESS_PACKAGE_JSON` 仍会出现，但属于既有非阻塞警告

## 11. 收工记录
- 构建是否通过：`tsc --noEmit` 通过
- 测试是否通过：`node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs` 均通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、backup 相关逻辑与验证脚本、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 继续把本轮锁定在 Phase F 的小闭环；canonical plan 和 workflow 限定本轮必须服务于 Phase F；前端主线状态文档确认 Backup 页可以继续做同页收口；backup 逻辑与验证脚本提供了稳定的共享来源和验证入口
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮优先收口共享逻辑与可重复验证，未切到浏览器自动化
- 待验证页面清单：Backup 页真实文件上传与浏览器交互、日志页实时流、首页窄屏细节
- 遗留项：当前仍未形成真正浏览器级的 Backup 导入/回滚自动化；下一轮更适合继续补同页更高层的行为证据，或者切到浏览器级 smoke
- 下一任务前置条件是否满足：是，backup 页脚的共享来源已经稳定下来，可继续向更高层行为验证推进