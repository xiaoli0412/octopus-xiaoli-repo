# 2026-04-27 Phase G Backup History Rollback Metadata Attribute Closure

## 1. 任务信息

- 任务名称：backup history / rollback metadata attribute contract closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup page selector-contract tightening
- 对应 milestone：Phase G 当前窗口 backup selector-contract 收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 备份导入主线 / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-26-phase-g-backup-node-entry-and-selector-contract-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-history-size-selector-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-rollback-preview-summary-selector-closure.md`
- 本次任务目标：把 backup 历史卡片和 rollback preview 中仍依赖格式化文本识别的元数据收口成更稳定的属性合同，便于后续 browser-grade 证据复用
- 本次已盘点本地资源：AGENTS.md、canonical plan、状态文档、workflow、用户上下文总账、前端主线状态、上一轮 memory、最近 backup worklogs、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`、`verify-backup-logic.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与文件；session 开头读取了 `using-superpowers` 与 `brainstorming` 本地 skill 文件；automation memory 用于承接上一轮 Node/backup 验证恢复结论
- 若未使用部分本地资源或上下文，原因：本轮不需要额外扩展到运行态脚本、browser smoke 或后端导入实现，因为目标是 backup 页面细粒度 contract 收口
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求本轮不要创建子 agent，且本任务是与当前上下文强耦合的单点页面合同修补

## 3. 本次硬规则

- 继续停留在 Phase G backup selector-contract 主线
- 只做最小 field-level contract 收口，不改备份业务逻辑
- 让组件测试与 repo-local verifier 优先依赖 selector / attribute，而不是更脆弱的可见文本格式
- 验证优先使用 `./scripts/use-node-env.ps1` 恢复的 Node 路径

## 4. 本次禁止事项

- 不扩大到其他设置页或 AI 自动化主线
- 不修改导入/回滚语义
- 不处理与本轮无关的全局 `tsc` 语法错误来源文件

## 5. 本次验收条件

- `Backup.tsx` 中新增可机器读取的 history / rollback metadata 属性合同
- `Backup.test.tsx` 与 `scripts/verify-backup-component.cjs` 同步断言这些属性
- `node scripts/verify-backup-component.cjs` 通过
- `node scripts/verify-backup-logic.mjs` 通过

## 6. 本次回滚点

- 仅回退 `Backup.tsx` / `Backup.test.tsx` / `scripts/verify-backup-component.cjs` 的 metadata attribute contract 改动即可

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 UI selector / attribute contract，再改验证
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否，visible copy 保持不变

## 8. 实施步骤

1. 盘点近期 backup worklog，确认下一最小缺口仍是 history-state / rollback-preview 元数据仍偏文本依赖
2. 在 `Backup.tsx` 为 history path / size / imported_at / latest badge，以及 rollback preview name / meta cells 补充 `data-*` 属性合同
3. 同步 `Backup.test.tsx` 与 `scripts/verify-backup-component.cjs`，把关键断言切到 selector + attribute 路径并重跑 focused verification

## 9. 测试与验证

- 构建命令：
  - `./scripts/use-node-env.ps1; node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `./scripts/use-node-env.ps1; node scripts/verify-backup-component.cjs`
  - `./scripts/use-node-env.ps1; node scripts/verify-backup-logic.mjs`
- 专项验证：
  - `git diff --check -- web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`

## 10. 风险与兼容性

- 新风险：`MetaGridCell` / `SummaryCell` 现在支持显式 `rawValue`，后续若其它调用点错误传入 `undefined` 需要注意属性是否应该回退显示值
- 兼容性风险：低；只新增 `data-*` 属性，不改可见行为
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：部分通过；backup 相关 Node 验证通过，但全量 `tsc` 仍被仓库内既有 `web/src/components/modules/ai-automation/index.tsx` JSX 语法错误阻塞，非本轮引入
- 测试是否通过：通过；`verify-backup-component.cjs` 与 `verify-backup-logic.mjs` 均已恢复为绿
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、用户上下文总账、前端主线状态、上一轮 memory、最近 backup worklogs、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`、`verify-backup-logic.mjs`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow / 用户总账：确认当前仍应优先停留在 Phase G backup 主线，不得扩散
  - 最近 backup worklogs：确认可继续收口 history-state / rollback-preview 的 field-level selector gap
  - automation memory：确认 Windows Node 路径已恢复，应沿用 `./scripts/use-node-env.ps1`
  - `Backup.tsx` / tests / scripts：确认 history path/imported_at/latest 与 rollback meta 仍主要靠文本内容识别，可收口为属性合同
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行 browser-grade smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮目标是 no-browser selector-contract 收口；browser-grade backup evidence 仍是下一小步
- 待验证页面清单：backup 页 browser-grade 证据、help-hint / accordion 真实交互、`375px` 页面可读性
- 若未使用子 agent，原因：用户明确禁止，且本轮任务为主线程强耦合单点修补
- worklog 是否更新：是
- 遗留项：
  - `web/src/components/modules/ai-automation/index.tsx` 仍有既有 JSX 语法错误，导致全量 `tsc --noEmit` 不绿
  - `Backup.tsx` 仍有仓库既有 LF/CRLF warning
  - browser-grade backup-page evidence 仍未关闭
- 下一任务前置条件是否满足：满足；可继续同一主线的 browser evidence 或 line-ending hygiene 小闭环
