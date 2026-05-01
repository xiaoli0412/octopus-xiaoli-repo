# 2026-04-25 Phase G Backup History Preview Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: history-preview selector-contract tightening and rollback-preview evidence prep

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- recent same-day backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- inspect the remaining `backup-*` anchors and rank the next 3-5 bounded candidate tasks
- choose one history/rollback-related selector group as the core task
- add only the smallest stable anchors needed for the component test and repo-local verifier to share the same history-preview contract
- verify with selector scans and minimal host-level Node probes, then write the blocker and next entry point back

## Candidate Tasks Considered

1. Restore a script-executable Node path and rerun `tsc --noEmit` plus `verify-backup-component.cjs`
2. Tighten the backup history item / rollback-preview selector chain with one more dual-covered closure
3. Move directly to browser-grade backup evidence
4. Expand into backup business logic or copy cleanup

## Chosen Task

- Core task: tighten the history snapshot interaction chain by adding stable selectors for the history item action row and the rollback preview summary structure
- Companion task: re-baseline the current host-level Node status so the next round can decide quickly between execution recovery and another static selector closure

## 本次硬规则

- 继续沿用当前 Phase G backup selector-contract 主线
- 只动 backup 页面级 selector 合同，不改导入、导出、回滚业务语义
- 若 Node 执行链仍不可用，必须把阻塞记为宿主层问题，而不是误记为 backup 页面回归

## 本次禁止事项

- 不改 `Backup.tsx` 导入/导出/回滚业务逻辑
- 不改 locale copy 与帮助文案
- 不扩到其他页面
- 不把本轮问题扩写成全局环境重构

## 本次验收条件

- `Backup.tsx` 为 history/rollback 链新增稳定 selector：`backup-history-item-actions`、`backup-rollback-preview-overview`、`backup-rollback-preview-meta-grid`
- `Backup.test.tsx` 和 `verify-backup-component.cjs` 都显式断言上述三个 selector
- 基于 `Backup.tsx` 真实 `data-testid="backup-*"` 的精确盘点，不出现新的单侧 selector 缺口
- 宿主级 Node 阻塞保留命令级记录，能说明 `-v` 与脚本执行的区别

## 本次回滚点

- 若新增 selector 断言造成误报，只需回退 `Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs` 中本轮新增的 history/rollback selector 即可

## 实现范围

- 受影响前端页面：`web/src/components/modules/setting/Backup.tsx`
- 受影响组件测试：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响 repo-local 验证脚本：`scripts/verify-backup-component.cjs`
- 新增记录文件：本 worklog 与 automation memory
- 是否影响旧数据：否
- 是否影响旧行为：否

## 实际完成

- 在 `Backup.tsx` 的快照历史项操作区新增 `backup-history-item-actions`，让“预览 / 回滚”按钮行形成稳定的页面级锚点。
- 在 `Backup.tsx` 的回滚预览区新增 `backup-rollback-preview-overview` 与 `backup-rollback-preview-meta-grid`，把风险摘要和元信息网格从“只靠文字断言”收紧为可复用 selector 合同。
- 在 `Backup.test.tsx` 中同步补上上述三个 selector 的显式断言，确保组件测试直接覆盖这条 history -> preview 链路。
- 在 `scripts/verify-backup-component.cjs` 中同步补上相同三处断言，让 repo-local verifier 与组件测试共用同一组页面级合同。
- 重新扫描 `Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs` 里的 `backup-*` 选择器，确认本轮新增后仍无新的单侧缺口。
- 继续复核宿主级 Node 状态，确认当前仍是“`node -v` 可返回版本，但任何脚本执行都会触发 `ncrypto::CSPRNG(nullptr, 0)` 断言”的宿主层阻塞。

## 测试与验证

- 文档与主线核对：
  - `Get-Content -Path "docs/LLM-Gateway-Refactor-Plan.zh-CN.md" -TotalCount 260`
  - `Get-Content -Path "docs/CURRENT_STATUS_AND_PLAN.zh-CN.md" -TotalCount 260`
  - `Get-Content -Path "docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md" -TotalCount 260`
  - `Get-Content -Path "docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md" -TotalCount 260`
  - `Get-Content -Path "docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md" -TotalCount 260`
- selector 盘点：
  - `rg -n --no-heading 'data-testid="backup-' web/src/components/modules/setting/Backup.tsx`
  - `Select-String -Path 'web/src/components/modules/setting/Backup.tsx' -Pattern 'data-testid="(backup-[^"]+)"' ...`
  - `Select-String -Path 'web/src/components/modules/setting/Backup.test.tsx' -Pattern 'backup-[a-z0-9-]+' -AllMatches ...`
  - `Select-String -Path 'scripts/verify-backup-component.cjs' -Pattern 'backup-[a-z0-9-]+' -AllMatches ...`
  - `rg -n --no-heading 'backup-history-item-actions|backup-rollback-preview-overview|backup-rollback-preview-meta-grid' web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
- Node / 宿主状态复核：
  - `$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; node -v`
  - `$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; node -e "console.log('script-ok')"`
  - `$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; & 'C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe' -v`
  - `$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; & 'C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe' -e "console.log('script-ok')"`
  - `if (Test-Path 'D:\gol1\node.exe') { ... & 'D:\gol1\node.exe' -e "console.log('script-ok')" }`

## 验证结果

- 代码层结果：
  - `Backup.tsx` 已新增 `backup-history-item-actions`、`backup-rollback-preview-overview`、`backup-rollback-preview-meta-grid`
  - `Backup.test.tsx` 与 `verify-backup-component.cjs` 已同步显式断言上述三个 selector
  - 基于 `Backup.tsx` 真实 `data-testid="backup-*"` 的精确扫描结果，当前 `Backup.tsx + Backup.test.tsx + verify-backup-component.cjs` 之间没有新的单侧 selector 缺口
- 宿主级阻塞：
  - `node -v` 与 `C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe -v` 均可返回版本号
  - 但 `node -e "console.log('script-ok')"`、`C:\Users\李昊桐\AppData\Local\OpenAI\Codex\bin\node.exe -e ...` 以及 `D:\gol1\node.exe -e ...` 仍全部触发 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`
  - 因此本轮依旧无法恢复 `tsc --noEmit` 或 `verify-backup-component.cjs` 的真实执行，阻塞仍是宿主层而非 backup 页面层

## 风险与兼容性

- 新风险：history/rollback 区结构一旦发生回退，会更早被组件测试与 repo-local verifier 暴露；属于低风险正向护栏
- 兼容性风险：低，仅影响 backup 页面验证层
- 当前阻塞：是，宿主级 Node 在脚本初始化阶段触发 `CSPRNG` 断言
- 是否阻塞下一任务：部分阻塞。继续推进静态 selector-contract 小闭环仍可做，但所有依赖实际 Node 脚本执行的 repo-local JS 验证继续受限

## 收工记录

- 构建是否通过：本轮未跑 build，受宿主级 Node 脚本执行阻塞限制
- 测试是否通过：静态 selector-contract 扫描通过；Node 脚本执行验证未恢复
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、用户上下文总账、automation memory、当日 backup worklogs、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`
- 本次使用这些本地资源得到的结论：当前仍应停留在 Phase G backup selector-contract 主线；pending-apply 之后最适合继续收口的是 history/rollback 页面级锚点；Node 执行阻塞仍属于宿主层
- 是否使用子 agent：否，用户明确要求主线程串行推进
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因：当前优先阻塞仍是宿主级 Node 脚本执行故障

## 遗留项

- `tsc --noEmit` 与 `verify-backup-component.cjs` 仍无法在本机实际执行，阻塞点仍为 `CSPRNG` 初始化失败
- `runtime-win.ps1` 的嵌套 PowerShell 稳定性仍未重新验证
- backup 页 browser-grade 证据仍未补齐

## 下一轮建议

1. 先尝试恢复一条真正可执行脚本的 Node 路径，优先目标是让 `node -e`、`tsc --noEmit`、`verify-backup-component.cjs` 重新可跑。
2. 若 Node 仍不可恢复，继续停留在 Phase G backup selector-contract 主线，优先寻找下一个最小的 page-level 锚点或 browser-smoke 预备 selector，而不是切回泛化分析。
3. 一旦 repo-local JS 验证恢复，立即重跑 `tsc --noEmit` 与 `verify-backup-component.cjs`，随后转向 backup 页 browser-grade 证据，重点复用 `backup-page`、`backup-history-panel`、`backup-rollback-preview-panel`、`backup-compatibility-panel`、`backup-remaining-migration-panel`。
