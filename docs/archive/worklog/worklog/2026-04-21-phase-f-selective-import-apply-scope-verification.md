# 2026-04-21 Phase F Selective Import Apply Scope Verification

## 1. 任务信息

- 任务名称：Backup selective import 范围复用验证补强
- 日期：2026-04-21
- 当前阶段：Phase F / Milestone 6 validation closure
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入

- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-backup-component-verification.md`
- 本次任务目标：把 Backup 页的 selective import 行为证据从“只验证 rollback preview/apply”补到“dry-run 捕获 scopes -> Apply Same Import 复用 scopes -> rollback preview/apply 同步带 scope override”的完整闭环
- 本次已盘点本地资源：automation memory、canonical plan、当前状态文档、详细执行工作流、前端主线状态文档、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`scripts/verify-backup-component.setting-mock.cjs`、`scripts/verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：复用 automation memory 中“继续补更高层 Backup 行为证据”的下一步结论；复用前端主线状态文档里“设置页备份导入仍需继续向主流程验证收口”的判断；复用已有 Backup 组件/逻辑验证脚本作为扩展入口
- 若未使用部分本地资源或上下文，原因：本轮不再重复展开更早的 copy alignment worklog，因为当前任务只聚焦 selective import 的行为一致性闭环
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务强耦合于同一页面与同一条验证链

## 3. 本次硬规则

- 必须继续围绕 Phase F backup / import / rollback 主线推进
- 必须留下真实代码增量和可执行验证，不得只写总结
- 不改后端导入/回滚语义，只补行为级验证证据

## 4. 本次禁止事项

- 不扩散到无关 UI 清理或其他页面
- 不把单进程脚本验证误报成浏览器 e2e
- 不把本轮范围扩大成新的导入功能实现

## 5. 本次验收条件

- 单进程 Backup 组件验证脚本能够直接断言 selective import scopes 在 dry-run 和 Apply Same Import 之间被原样复用
- 同一条验证链能断言 rollback preview 和 rollback apply 都带上当前 scope override
- `Backup.test.tsx` 同步补上相同行为断言，作为环境恢复后的正式测试草稿
- `node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`、`tsc --noEmit` 通过

## 6. 本次回滚点

- 回退 `scripts/verify-backup-component.cjs`
- 回退 `scripts/verify-backup-component.setting-mock.cjs`
- 回退 `web/src/components/modules/setting/Backup.test.tsx`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品语义，先补 mock 和验证断言，再同步 Vitest 草稿
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口：无真实接口契约变更；仅扩展本地 mock 对 `importScopes` / rollback payload 的记录
- 是否影响旧数据：否
- 是否影响旧行为：否，新增的是验证覆盖范围

## 8. 实施步骤

1. 阅读 `Backup.tsx` 中 `buildPreparedImportRequest`、`executeImport`、`onApplySameImport`、`onPreviewRollbackSnapshot`、`onRollbackSnapshot` 的 scope 传递路径，确认真实行为是 dry-run/apply/rollback 都复用 `selectiveImport` 生成的 `importScopes`
2. 更新 `scripts/verify-backup-component.setting-mock.cjs`，让 rollback preview 和 rollback apply 都记录完整 payload，而不只记录 snapshot 名称
3. 更新 `scripts/verify-backup-component.cjs`，补 selective import 场景下的 dry-run/apply/rollback scope 断言，并校验捕获的 scope 摘要文案
4. 同步更新 `web/src/components/modules/setting/Backup.test.tsx`，把同一条 selective import 断言补到正式测试草稿中
5. 运行 Backup 相关验证链，确认没有新的断言漂移或类型问题

## 9. 测试与验证

- 构建命令：无额外构建
- 测试命令：
  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 专项验证：
  - 验证 selective import 下关闭 `Routing` 后，dry-run 提交的 `importScopes` 只包含剩余作用域
  - 验证 Apply Same Import 复用同一 `previewToken` 时，也复用 dry-run 捕获的 `importScopes`
  - 验证 rollback preview / rollback apply 都带上同一份 scope override

## 10. 风险与兼容性

- 新风险：单进程验证脚本对 Backup 页结构更敏感，后续若调整 selective import 区块或结果面板文案，需要同步更新断言
- 兼容性风险：低；没有改动产品逻辑或接口，只增加验证覆盖
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：`tsc --noEmit` 通过
- 测试是否通过：通过；`verify-backup-component.cjs`、`verify-backup-logic.mjs` 均成功
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、已有 Backup 组件/逻辑验证脚本、当前线程上下文
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 明确要求继续补高层 Backup 行为证据；canonical plan 和 workflow 限定本轮必须服务于 Phase F；前端主线状态文档确认设置页主流程仍需持续验证收口；已有脚本提供了最小可执行入口，避免重新发明验证链
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮优先补可重复执行的单进程行为验证；本机浏览器自动化链仍受 `spawn EPERM` 环境限制
- 待验证页面清单：Backup 页真实文件上传与浏览器交互、日志页实时流、首页窄屏细节
- 若未使用子 agent，原因：用户明确要求不创建子 agent，且本轮任务集中在单页面验证链
- worklog 是否更新：是
- 遗留项：当前仍未形成真正浏览器级的 Backup 导入/回滚自动化；下一轮更适合继续补“真实文件上传 + dry-run -> apply same import”方向的更高层 smoke 证据，或在同主线下继续强化结果面板和风险信号的行为覆盖
- 下一任务前置条件是否满足：是，selective import 的核心 scope 复用路径已经有稳定的单进程验证入口
