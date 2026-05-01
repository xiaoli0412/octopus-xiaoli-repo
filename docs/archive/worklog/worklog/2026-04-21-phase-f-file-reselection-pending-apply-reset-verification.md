# 2026-04-21 Phase F File Reselection Pending Apply Reset Verification

## 1. 任务信息

- 任务名称 Backup 文件重选后 pending apply 失效验证
- 日期 2026-04-21
- 当前阶段 Phase F / Milestone 6 validation closure
- 对应 milestone 里程碑 6 验证与部署

## 2. 开工前输入

- 对应 canonical 章节 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 11.5 节、第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节 `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 8 节 Phase F、第 10 节任务模板
- 上一个相关 worklog `docs/worklog/2026-04-21-phase-f-latest-rollback-component-verification.md`
- 本次任务目标 补一条 backup 验证证据，证明 dry-run 绑定不会跨文件复用，用户一旦重新选择文件，pending apply 应立即失效并要求重新预览
- 本次已盘点本地资源 automation memory、canonical plan、当前状态文档、前端主线状态文档、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/Backup.test.tsx`、`scripts/verify-backup-component.cjs`、`scripts/verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文 复用 automation memory 中继续补高层 Backup 行为证据的下一步结论；复用前端主线状态文档里设置页备份导入仍需继续向主流程验证收口的判断；复用已有 Backup 组件和逻辑验证脚本作为扩展入口
- 若未使用部分本地资源或上下文 原因是本轮不再重复展开更早的 copy alignment worklog，因为当前任务只聚焦文件重选失效这一条行为验证
- 本次是否启用子 agent 与分工边界 否
- 本次子 agent 使用模型 无
- 若未使用子 agent 原因 用户明确要求主线程执行，且本轮任务集中在单页面验证链

## 3. 本次硬规则

- 必须继续围绕 Phase F backup / import / rollback 主线推进
- 必须留下真实代码增量和可执行验证，不得只写总结
- 不改后端导入 / 回滚语义，只补行为级验证证据

## 4. 本次禁止事项

- 不扩散到无关 UI 清理或其他页面
- 不把单进程脚本验证误报成浏览器 e2e
- 不把本轮范围扩大成新的导入功能实现

## 5. 本次验收条件

- 单进程 Backup 组件验证脚本能够直接断言文件重选后，旧的 dry-run 绑定会被清空并要求重新预览
- 同一条验证链能断言重新预览时使用新文件，不复用旧文件的 pending apply
- `node scripts/verify-backup-component.cjs`、`node scripts/verify-backup-logic.mjs`、`tsc --noEmit` 通过

## 6. 本次回滚点

- 回退 `scripts/verify-backup-component.cjs`
- 回退 `web/src/components/modules/setting/Backup.test.tsx`

## 7. 实现范围

- 先改数据语义还是先改 UI 不改产品语义，先补验证断言，再同步 Vitest 草稿
- 受影响后端模块 无
- 受影响前端模块 `web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口 无真实接口契约变更；仅补验证覆盖
- 是否影响旧数据 否
- 是否影响旧行为 否，新增的是验证覆盖范围

## 8. 实施步骤

1. 阅读 `Backup.tsx` 里 `onPickFile`、`resetImportFormState`、`executeImport` 的状态清理路径，确认文件重选会清空 pending apply
2. 在 `scripts/verify-backup-component.cjs` 中新增文件重选验证，断言旧的 Apply This Dry-Run 会消失，新文件需要重新 dry-run 才能继续 apply
3. 在 `web/src/components/modules/setting/Backup.test.tsx` 中同步增加相同行为草稿，确保正式测试恢复后仍有同样的断言
4. 运行 Backup 组件验证、逻辑验证和 `tsc --noEmit`，确认没有新的断言漂移或类型问题

## 9. 测试与验证

- 构建命令 无额外构建
- 测试命令
  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 专项验证
  - 验证第一轮 dry-run 后，切换到新文件会清掉 pending apply
  - 验证新文件重新 dry-run 后，使用的是新文件而不是旧文件的预览 token

## 10. 风险与兼容性

- 新风险 单进程验证脚本对 Backup 页结构更敏感，后续若调整文件输入或结果面板文案，需要同步更新断言
- 兼容性风险 低；没有改动产品逻辑或接口，只增加验证覆盖
- 是否阻塞下一任务 否

## 11. 收工记录

- 构建是否通过 `tsc --noEmit` 通过
- 测试是否通过 通过；`verify-backup-component.cjs`、`verify-backup-logic.mjs` 均成功
- 本次使用了哪些本地资源 / skills / 记忆上下文 automation memory、canonical plan、详细执行工作流、前端主线状态文档、`Backup.tsx`、`Backup.test.tsx`、验证脚本和当前线程上下文
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论 memory 指向继续补高层 Backup 行为证据；canonical plan 和 workflow 限定本轮必须服务于 Phase F；前端主线状态文档确认设置页主流程仍需持续验证收口；`Backup.tsx` 暴露出 native file input 需要主动清空，`Backup.test.tsx` 提供了可重复的 draft 验证入口，验证脚本提供了最小可执行收口
- 本次使用了哪些子 agent 及其结论 无
- 子 agent 分工、负责范围与产出摘要 无
- 手工 smoke 状态 未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境 本机 Vitest / Vite / browser 自动化仍受 Windows `spawn EPERM` 限制，故本轮继续采用仓库内单进程组件验证
- 待验证页面清单 backup 页真实文件上传与浏览器交互、日志页实时流、首页窄屏细节
- 若未使用子 agent，原因 用户明确要求不创建子 agent，且本轮任务聚焦于单页面验证链
- worklog 是否更新 是
- 遗留项 当前仍未形成真正浏览器级的 Backup 导入 / 回滚自动化；下一轮更适合继续补更高层的 backup/import/rollback 行为证据，或者在同主线下继续强化结果面板和风险信号的行为覆盖
- 下一任务前置条件是否满足 是，文件重选失效和同文件重选都已经有稳定的单进程验证入口
- 记录时间 2026-04-21T15:21:15+08:00

## 12. 本轮补充记录

- 记录时间 2026-04-21T16:03:11+08:00
- 本轮实际完成的改动
  - 在 [Backup.test.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.test.tsx) 里加严同文件重选断言，显式检查旧的 `Apply This Dry-Run` 会消失、原生 file input 会被清空、然后必须重新 dry-run 才能再次应用同一个文件
  - 在 [scripts/verify-backup-component.cjs](/D:/GPT-codex/octopus_repo/scripts/verify-backup-component.cjs) 里同步加入同文件重选验证，覆盖 dry-run -> 清空 -> 重选同文件 -> 再次 dry-run 的完整路径
- 已执行的验证
  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 当前仍存在的阻塞或风险
  - 浏览器级手工 smoke 仍未执行
  - `backup-logic.ts` 仍保留 `MODULE_TYPELESS_PACKAGE_JSON` 警告，但这不影响本轮行为验证结果
- 下一轮最适合继续推进的具体事项
  - 继续补 Phase F backup/import/rollback 的更高层行为证据，优先浏览器级 smoke 或更丰富的 rollback/import 交互
