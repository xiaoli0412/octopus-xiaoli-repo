# 2026-04-21 Phase F Backup Redacted Export Toggle Verification

## 1. 任务信息
- 任务名称：Backup 导出 plaintext credentials 切换验证
- 日期：2026-04-21
- 当前阶段：Phase F / Milestone 6 validation closure
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-export-presentation-closure.md`
- 本次任务目标：补一条可验证的 Backup 导出 redacted 切换证据，让 UI、组件脚本和前端类型检查都能覆盖 `include_secrets=false` 路径
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`scripts/verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：复用 automation memory 中“继续在 Phase F 做可验证的小闭环”的结论；复用 canonical plan 与 workflow 对 Phase F 收口的边界；复用备份页现有导出切换/验证脚本作为稳定入口
- 若未使用部分本地资源或上下文，原因：未继续展开更早的 Phase F 文案型 worklog，因为本轮只补备份导出切换的一条小验证链
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行

## 3. 本次硬规则
- 必须围绕 Phase F backup/import/rollback 主线推进
- 必须留下真实代码增量和可执行验证，不得只写总结
- 不改后端契约，不扩散到无关页面

## 4. 本次禁止事项
- 不扩散到其他页面或无关 UI 清理
- 不改后端导入语义
- 不把导出切换写成新的产品功能设计

## 5. 本次验收条件
- Backup 页能稳定切到 redacted export 状态
- 组件测试和 node 验证脚本都能覆盖 `include_secrets=false` 路径
- 前端 `tsc --noEmit` 通过

## 6. 本次回滚点
- 回退 `web/src/components/modules/setting/Backup.test.tsx`
- 回退 `scripts/verify-backup-component.cjs`

## 7. 实现范围
- 先改数据语义还是先改 UI：先补验证，再微调测试辅助函数
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否，只是补充 export toggle 的验证覆盖

## 8. 实施步骤
1. 给 Backup 前端测试补一条 redacted export 场景
2. 给 node 版 backup verification script 补同一条断言
3. 修正测试辅助函数的 switch 定位方式，避免误命中 export 区域里的多 switch
4. 跑组件验证和前端 typecheck
5. 记录结果并更新 automation memory

## 9. 测试与验证
- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`node scripts/verify-backup-component.cjs`
- 专项验证：`node scripts/verify-backup-logic.mjs`

## 10. 风险与兼容性
- 新风险：测试辅助函数如果定位太宽，会误命中 export 区域里的多 switch
- 兼容性风险：低，只影响验证层
- 是否阻塞下一任务：否

## 11. 收工记录
- 构建是否通过：待验证
- 测试是否通过：待验证
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、backup 相关源码与验证脚本、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 继续把本轮锁定在 Phase F 的小闭环；canonical plan 与 workflow 约束本轮必须服务于备份/导入/回滚收口；前端主线状态文档确认 Backup 页仍可继续做同页验证补强
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮优先补 source-level validation，未切到浏览器自动化
- 待验证页面清单：Backup 页导出切换、日志页实时流、首页窄屏细节
- 若未使用子 agent，原因：用户明确要求主线程执行
- worklog 是否更新：是
- 遗留项：还缺浏览器级手工 smoke
- 下一任务前置条件是否满足：待验证
