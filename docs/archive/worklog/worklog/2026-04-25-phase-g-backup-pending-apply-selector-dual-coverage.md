# 2026-04-25 Phase G Backup Pending Apply Selector Dual Coverage

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: pending-apply selector dual coverage and host-level verification path re-baselining

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- same-day backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- first probe whether Node-based repo-local verification can be restored on this host
- then close the smallest remaining selector gap that still lacks dual coverage
- verify with exact selector scans and host-level command attempts, and write the blocker back if repo-local execution is still broken

## Candidate Tasks Considered

1. Restore Node and rerun `tsc --noEmit` plus `verify-backup-component.cjs`
2. Close the next backup selector gap by dual-covering `backup-pending-apply-ready` and `backup-pending-apply-meta-grid`
3. Expand immediately to browser-grade backup smoke
4. Touch backup business logic or copy

## Chosen Task

- Core task: dual-cover the pending-apply selector group without touching backup business logic
- Companion task: classify the current host-level Node failure precisely enough that the next round can either fix it or route around it quickly

## 本次硬规则

- 继续沿用当前 Phase G backup selector-contract 主线
- 只改 backup 验证层，不改导入、导出、回滚业务语义
- 若 Node 验证路径仍失效，必须把阻塞缩到宿主层而不是误记为 backup 回归

## 本次禁止事项

- 不改 `Backup.tsx` 业务逻辑
- 不改 locale copy 与帮助文案
- 不扩到其它页面
- 不把本轮问题发散成全局 Node/PowerShell 环境改造

## 本次验收条件

- `Backup.test.tsx` 对 pending-apply 流程显式断言 `backup-pending-apply-ready` 与 `backup-pending-apply-meta-grid`
- `verify-backup-component.cjs` 在 dry-run / map / replace 三条路径显式断言 `backup-pending-apply-ready` 与 `backup-pending-apply-meta-grid`
- 基于 `Backup.tsx` 中真实 `data-testid="backup-*"` 的精确盘点结果，不再出现“仅在单侧覆盖”的 backup selector
- 宿主级 Node/PowerShell 阻塞要有可复用的命令级记录

## 本次回滚点

- 若新增断言误报，回退 `Backup.test.tsx` 与 `verify-backup-component.cjs` 中本轮新增的 pending-apply selector 断言即可

## 实现范围

- 受影响前端验证文件：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响 repo-local 验证脚本：`scripts/verify-backup-component.cjs`
- 新增记录文件：本 worklog 与 automation memory
- 是否影响旧数据：否
- 是否影响旧行为：否

## 实际完成

- 在 `Backup.test.tsx` 的 incremental、map、replace 三条 pending-apply 路径中补上了 `backup-pending-apply-meta-grid` 断言，和既有的 `backup-pending-apply-ready` 一起锁住同一组 pending-apply 结构。
- 在 `verify-backup-component.cjs` 中把三条对应路径从“只等 `backup-pending-apply-panel` 出现”收紧成“先等 `backup-pending-apply-ready`，再显式断言 `backup-pending-apply-panel`、`backup-pending-apply-ready`、`backup-pending-apply-meta-grid`”。
- 用精确的 `data-testid="backup-*"` 扫描重新盘点 `Backup.tsx`、`Backup.test.tsx` 与 `verify-backup-component.cjs` 的覆盖差异，结果已经收口到零缺口。
- 额外排查了三套 Node 可执行入口与运行态脚本入口，确认当前 repo-local JS 验证阻塞仍是宿主级问题，不是 backup 页面逻辑回退。

## 测试与验证

- 文档与主线核对：
  - `Get-Content -Raw docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `Get-Content -Raw docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `Get-Content -Raw docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `Get-Content -Raw docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `Get-Content -Raw docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `Get-Content -Raw docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- 目标盘点：
  - `rg -n "backup-[a-z0-9-]+" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
  - `rg -n "backup-pending-apply-ready|backup-pending-apply-meta-grid" web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
  - exact `data-testid="backup-*"` coverage scan against `Backup.tsx`
- Node / repo-local 验证恢复尝试：
  - `where.exe node`
  - `Get-Command node`
  - `node -v`
  - `node -e "console.log('script-ok')"`
  - `Start-Process D:\gol1\node.exe -ArgumentList @('-e','console.log("script-ok")')`
  - `Start-Process D:\gol1\node.exe -ArgumentList @('web/node_modules/typescript/bin/tsc','--noEmit','-p','web/tsconfig.json')`
  - `Start-Process D:\gol1\node.exe -ArgumentList @('scripts/verify-backup-component.cjs')`
  - `C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe -e "console.log('script-ok')"`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`

## 验证结果

- 代码层结果：
  - `Backup.test.tsx` 现已在三条 pending-apply 路径显式断言 `backup-pending-apply-meta-grid`
  - `verify-backup-component.cjs` 现已在 dry-run / map / replace 三条路径显式断言 `backup-pending-apply-ready` 与 `backup-pending-apply-meta-grid`
  - 基于 `Backup.tsx` 真正 `data-testid="backup-*"` 的精确扫描结果，当前 backup selector-contract 在 `Backup.test.tsx + verify-backup-component.cjs` 之间已无单侧缺口
- 宿主级阻塞：
  - `D:\gol1\node.exe`、`C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe` 只要执行脚本就会在初始化阶段触发 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`
  - `C:\Program Files\WindowsApps\...\node.exe` 在当前环境中无法正常启动，报“找不到指定的模块”
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status` 当前触发宿主级 `8009001d`，说明本机嵌套 PowerShell 入口也不稳定
- 因此：
  - 本轮无法恢复 `tsc --noEmit` 与 `verify-backup-component.cjs` 的真实执行
  - 但已经把 backup 页下一个最小 selector-contract 缺口补成了代码级闭环，并把运行时阻塞准确缩到宿主层

## 风险与兼容性

- 新风险：pending-apply 区块一旦结构回退，会被组件测试与 repo-local 脚本更早暴露；属于低风险正向护栏
- 兼容性风险：低，仅影响 backup 验证层
- 当前阻塞：是，宿主级 Node CSPRNG 初始化失败 + 嵌套 PowerShell `8009001d`
- 是否阻塞下一任务：部分阻塞。继续做纯静态 selector-contract 小补丁仍可推进；所有依赖 Node 执行的 repo-local JS 验证都会受影响

## 收工记录

- 构建是否通过：本轮未跑 build，受宿主级 Node 阻塞限制
- 测试是否通过：静态 selector-contract 盘点通过；Node/PowerShell 运行态验证未恢复
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、前端主线文档、用户上下文总账、automation memory、当日 backup worklogs、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`
- 本次使用这些本地资源得到的结论：确认当前仍应停留在 Phase G backup selector-contract 主线；上一轮的最小缺口已经从 compatibility details 收缩到 pending-apply selector 组；Node/PowerShell 问题属于宿主层，而不是 backup 页面逻辑回退
- 是否使用子 agent：否，用户明确要求主线程串行推进
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因：当前更高优先级阻塞是宿主级 Node / PowerShell 运行时故障

## 遗留项

- `tsc --noEmit` 与 `verify-backup-component.cjs` 仍无法在本机实际执行，阻塞点统一为 `CSPRNG` 初始化失败
- `runtime-win.ps1` 当前也受宿主级 `8009001d` 影响，运行态入口未恢复
- backup 页 browser-grade 证据仍未补齐

## 下一轮建议

1. 先从宿主层恢复一个可执行脚本的 Node 入口，至少让 `node -e`、`tsc --noEmit`、`verify-backup-component.cjs` 重新可跑。
2. 如果 Node 仍不可恢复，继续沿 Phase G backup selector-contract 主线推进下一个纯静态 panel/root selector 小闭环，避免空转。
3. 一旦 repo-local JS 验证恢复，优先重跑 `tsc --noEmit` 和 `verify-backup-component.cjs`，然后再转向 backup 页 browser-grade 证据，复用现有锚点：`backup-page`、`backup-pending-apply-panel`、`backup-history-panel`、`backup-compatibility-panel`、`backup-remaining-migration-panel`。
