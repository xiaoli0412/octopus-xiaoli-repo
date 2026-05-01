# 2026-04-25 Phase G Backup Compatibility Details Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page panel-level selector closure and verification-path re-baselining

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- close the next smallest double-coverage gap around `backup-compatibility-details`
- keep the write scope inside backup page verification only
- verify with the existing repo-local path first, and if blocked, record the host-level failure precisely instead of widening scope

## 本次硬规则

- 继续沿用当前 Phase G backup selector-contract 主线
- 只补测试与 repo-local 验证脚本中的 selector 契约，不改备份业务语义
- 若既有验证路径失效，必须先记录真实阻塞原因，再更新下一轮入口

## 本次禁止事项

- 不改导入、导出、回滚逻辑
- 不改 locale copy 与帮助文案
- 不扩到浏览器级 backup smoke 修复
- 不触碰非备份页面文件

## 本次验收条件

- `Backup.test.tsx` 在 incremental、map、replace 三条导入路径都显式断言 `backup-compatibility-details`
- `scripts/verify-backup-component.cjs` 在同三条路径都显式断言 `backup-compatibility-details`
- 清理脚本里同区域重复的 remaining-migration 断言噪音
- 至少尝试运行当前约定的 repo-local 验证路径，并把结果如实落盘

## 本次回滚点

- 若新增 selector 断言误报，撤销本轮新增断言与重复断言清理即可

## 实现范围

- 先改测试契约，再改 repo-local 验证脚本
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/Backup.test.tsx`
- 受影响脚本：`scripts/verify-backup-component.cjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否

## 实施步骤

1. 对照 `Backup.tsx`、`Backup.test.tsx` 和 `verify-backup-component.cjs` 盘点 `backup-*` selector 覆盖差异。
2. 在组件测试的 incremental 与 replace 路径补上 `backup-compatibility-details`，同时确保 map 路径继续显式覆盖该 selector。
3. 在 `verify-backup-component.cjs` 的 incremental、map、replace 路径补上同一 selector 断言，并删除 rollback 路径里重复的 `backup-remaining-migration-trigger` 断言与重复 click。
4. 尝试运行 `tsc --noEmit` 与 `verify-backup-component.cjs`，若失败则继续定位到宿主级别并记录。

## 测试与验证

- 构建命令：无
- 测试命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 专项验证：`node scripts/verify-backup-component.cjs`
- 额外排查：`where.exe node`、`Get-Command node`、`Start-Process D:\gol1\node.exe ...`、`D:\gol1\node.exe -e ...`

## 验证结果

- 代码层结果：本轮新增 selector 断言已落地，相关行位于：
  - `web/src/components/modules/setting/Backup.test.tsx:363`
  - `web/src/components/modules/setting/Backup.test.tsx:366`
  - `web/src/components/modules/setting/Backup.test.tsx:490`
  - `web/src/components/modules/setting/Backup.test.tsx:493`
  - `web/src/components/modules/setting/Backup.test.tsx:539`
  - `web/src/components/modules/setting/Backup.test.tsx:542`
  - `scripts/verify-backup-component.cjs:276`
  - `scripts/verify-backup-component.cjs:384`
  - `scripts/verify-backup-component.cjs:387`
  - `scripts/verify-backup-component.cjs:411`
  - `scripts/verify-backup-component.cjs:414`
- `tsc --noEmit`：未通过，失败点为本机 Node 初始化阶段的宿主级断言 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`
- `verify-backup-component.cjs`：未通过，失败点与上面一致，同样在 Node 初始化阶段崩溃，尚未进入脚本逻辑
- 额外排查结论：
  - `node -v` 可输出版本
  - 但 `node -e "console.log('script-ok')"` 也会触发同一 `CSPRNG` 断言
  - 说明当前阻塞不是 `tsc`、Vitest 或 backup verifier 特有，而是本机 Node 执行脚本时的统一宿主故障

## 风险与兼容性

- 新风险：新增 selector 断言会更早暴露兼容性详情区的结构回退，但属于低风险正向护栏
- 兼容性风险：低，仅影响测试与 repo-local 验证层
- 当前阻塞：是。当前宿主上的 Node 脚本执行统一被 `CSPRNG` 断言阻塞，导致原本可用的 repo-local 验证路径失效
- 是否阻塞下一任务：部分阻塞。继续做同类前端 selector-contract 小补丁仍可推进，但凡依赖 Node 执行的校验都会受影响

## 收工记录

- 构建是否通过：本轮未跑 build
- 测试是否通过：未完全通过；已完成尝试并定位到宿主级 Node 运行时阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、前端主线文档、automation memory、当日 backup worklogs、`Backup.tsx`、`Backup.test.tsx`、`verify-backup-component.cjs`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：确认当前主线仍是 Phase G backup selector-contract 收口；上一轮遗留的最小缺口是 `backup-compatibility-details`；同时发现旧 memory 中记载的稳定验证路径已不再成立
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用子 agent，原因是用户明确要求主线程串行推进，且本轮写作用域很小
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮不需要运行态或浏览器环境；当前更高优先级阻塞是 Node 脚本执行故障
- 待验证页面清单：backup page browser-grade evidence、Node 恢复后的 `tsc --noEmit`、`verify-backup-component.cjs`
- worklog 是否更新：是
- 遗留项：
  - 当前机器上的 Node 只要执行脚本就会在初始化阶段触发 `ncrypto::CSPRNG(nullptr, 0)`
  - backup 页是否还有剩余 panel/root selector 未被双重覆盖，需在 Node 验证恢复后继续盘点
  - browser-grade backup evidence 仍未补齐
- 下一任务前置条件是否满足：部分满足；若继续推进纯静态 selector-contract 小任务可直接接续，若要恢复 repo-local proof path，则需先解决本机 Node 运行时故障
