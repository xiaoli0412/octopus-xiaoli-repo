# 2026-04-25 Phase G Backup History Item Detail Selectors

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup history item detail selector closure

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-25-phase-g-backup-history-and-apply-selector-closure.md`
- `docs/worklog/2026-04-25-phase-g-backup-history-meta-selector-closure.md`
- automation memory: `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- close one more smallest page-level gap inside the snapshot history item row
- add stable selectors for history item name, path, and latest badge without touching backup business behavior
- keep verification on the existing repo-local proof path when the host allows Node to start

## Candidate Tasks Considered

1. add detail selectors to the snapshot history item header row
2. add more selectors around rollback preview details
3. widen into browser-grade backup smoke
4. return to backup helper cleanup

## Chosen Task

- Core task: add explicit history-item detail selectors for snapshot name, snapshot path, and latest badge
- Companion task: update `Backup.test.tsx` and `scripts/verify-backup-component.cjs` to assert the new selectors directly

## 本次硬规则

- 继续沿用当前 `Phase G backup selector-contract` 主线
- 只补备份页快照历史项的页面级 selector，不改备份业务语义
- 保持改动集中在 `Backup.tsx`、组件测试和 repo-local verifier 三处

## 本次禁止事项

- 不改导入/导出/回滚逻辑
- 不改 locale copy 和帮助提示内容
- 不扩散到其他设置页模块
- 不把本轮任务扩大成浏览器级 smoke 修复

## 本次验收条件

- `Backup.tsx` 为快照历史项补齐 `backup-history-item-name`、`backup-history-item-path`、`backup-history-item-latest-badge`
- `Backup.test.tsx` 直接断言三枚 selector 及其内容
- `scripts/verify-backup-component.cjs` 直接断言三枚 selector 及其内容
- 若宿主允许 Node 启动，则继续通过 `tsc --noEmit` 与 `verify-backup-component.cjs`

## 本次回滚点

- 如 selector 命名或断言造成误判，回退本轮新加的 history item detail selectors 与对应断言即可

## 实现范围

- 先改 UI 契约，再改测试与验证脚本
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.tsx`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否

## 实施步骤

1. 盘点当前 `backup-history-item-*` 现有 selector 覆盖，定位快照历史项 header row 的最小缺口
2. 在 `Backup.tsx` 的 history item header row 中补齐名称、路径和 latest badge 的稳定 selector
3. 在 `Backup.test.tsx` 和 `scripts/verify-backup-component.cjs` 中同步改为 selector-first 断言
4. 尝试运行 `tsc --noEmit` 与 `verify-backup-component.cjs`，若宿主 Node 初始化仍失败，则记录为环境阻塞

## 实际完成

- `Backup.tsx`
  - 在快照历史项 meta 区中新增 `backup-history-item-name`
  - 在快照历史项 meta 区中新增 `backup-history-item-path`
  - 在 latest badge 上新增 `backup-history-item-latest-badge`
- `Backup.test.tsx`
  - 把回滚历史测试改为先缓存 `historyItem`
  - 直接断言 `name/path/latest-badge` 三枚 selector 及其内容
  - 修正了同段里 `backup-page` 根锚点断言的缩进
- `scripts/verify-backup-component.cjs`
  - 在 snapshot history 打开路径下新增 `name/path/latest-badge` selector 断言
  - 同步断言 `snapshot-1`、`snapshots/snapshot-1.json` 和 `Latest` 文本内容
  - 修正了同段里 `backup-page` 根锚点断言的缩进

## 测试与验证

- 尝试执行：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 尝试执行：`node scripts/verify-backup-component.cjs`
- 实际结果：两条 Node 路径都在当前宿主/sandbox 中命中同一类 Node 初始化崩溃：`Assertion failed: ncrypto::CSPRNG(nullptr, 0)`
- 替代证明：
  - 通过 `rg` 确认三枚新 selector 已在 `Backup.tsx`、`Backup.test.tsx`、`scripts/verify-backup-component.cjs` 同步落地
  - 通过 targeted file reads 确认新 selector 的 DOM 位置与断言位置一致
  - 通过 `git diff -- <files>` 核对本轮改动仅落在计划内三文件与对应行段

## 风险与兼容性

- 新风险：新的 history item detail selector 会更早暴露快照历史项 header row 的结构回退，但属于正向护栏
- 兼容性风险：低，仅影响组件测试与 repo-local verifier 的稳定性
- 是否阻塞下一任务：否；当前真正阻塞的是宿主级 Node/CSPRNG 初始化问题，不是本轮 selector 变更

## 收工记录

- 构建是否通过：未能完成；受宿主级 Node 初始化崩溃阻塞
- 测试是否通过：未能完成；受宿主级 Node 初始化崩溃阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、当前状态文档、前端主线文档、最近 backup worklog、automation memory、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：确认当前主线仍是 Phase G backup selector-contract；确认最小闭环应继续停留在 `Backup.tsx` 页面级锚点，而不是发散到逻辑层或其他设置页
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用子 agent；原因是用户明确要求主线程串行推进，且本轮写入范围很小
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮不需要浏览器 smoke；repo-local Node 验证已先被宿主级 `ncrypto::CSPRNG(nullptr, 0)` 阻塞
- 待验证页面清单：备份页 browser-grade smoke、回滚预览更细 selector 合同
- worklog 是否更新：是
- 遗留项：
  - 当前宿主里的 Node 命令初始化崩溃仍需单独处理，才可恢复 `tsc --noEmit` 与 `verify-backup-component.cjs` 的直接运行证明
  - backup history item 之后，下一步更适合继续补 rollback preview 区的细粒度 selector，而不是转回 helper-only churn
- 下一任务前置条件是否满足：是；可继续沿同一 backup selector-contract 主线推进
