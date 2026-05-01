# 2026-04-25 Phase G Backup Apply Confirm Switch Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: backup page selector contract and browser-evidence prep

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup mainline
- add the smallest remaining page-level selector contract for `backup-apply-confirm-switch`
- keep the write scope narrow and avoid touching backup logic
- verify with repo-local typecheck and backup verification

## 本次硬规则

- 继续沿用当前 Phase G backup selector-contract 主线
- 本轮只补 `backup-apply-confirm-switch` 的 selector 契约，不扩到更大的备份逻辑
- 保持页面级 selector-first 验证风格

## 本次禁止事项

- 不改备份导入/导出业务语义
- 不扩展到浏览器级 smoke 修复
- 不触碰非备份页面代码

## 本次验收条件

- `Backup.test.tsx` 对 `backup-apply-confirm-switch` 有稳定断言
- `scripts/verify-backup-component.cjs` 对同一 selector 有稳定断言
- `tsc --noEmit` 与 backup verification 通过

## 本次回滚点

- 如断言引入误判，撤销本轮新增的 selector 断言即可

## 实现范围

- 先改测试契约，再改验证脚本
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否

## 实施步骤

1. 在 backup 页面测试中补充 `backup-apply-confirm-switch` 的 root-level selector 断言。
2. 在 `scripts/verify-backup-component.cjs` 的 dry-run/apply path 中补充同一 selector 断言。
3. 运行 typecheck 和备份验证，确认 selector 契约不回退。

## 测试与验证

- 构建命令：无
- 测试命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 专项验证：`node scripts/verify-backup-component.cjs`

## 风险与兼容性

- 新风险：新增 selector 断言可能暴露页面结构漂移
- 兼容性风险：低，属于测试/验证契约
- 是否阻塞下一任务：否

## 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、LLM-Gateway-Refactor-Plan.zh-CN.md、CURRENT_STATUS_AND_PLAN.zh-CN.md、DETAILED_EXECUTION_WORKFLOW.zh-CN.md、当前 worklog、automation memory、`Backup.test.tsx`、`scripts/verify-backup-component.cjs`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：确认当前主线仍是 Phase G backup selector-contract 收口，且 `backup-apply-confirm-switch` 仍缺少 repo-local 双重固化，需要补齐测试与验证脚本。
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用子 agent，原因是本轮任务很小且写作用域可控