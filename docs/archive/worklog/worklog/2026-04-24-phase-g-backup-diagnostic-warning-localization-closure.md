# 2026-04-24 Phase G Backup Diagnostic Warning Localization Closure

## 1. 任务信息

- 任务名称：备份导入诊断 warning 中文化收口
- 日期：2026-04-24
- 当前阶段：Phase G 截图优先 UI 收口
- 对应 milestone：Phase G settings / backup no-browser contract closure

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9 节前端收口与中文界面一致性要求
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3 节
- 上一个相关 worklog：`docs/worklog/2026-04-24-phase-g-settings-no-browser-green-restoration.md`
- 本次任务目标：收口备份导入详情区仍会在中文界面直出的英文 warning / skip reason，形成代码 + no-browser 验证 + 记录的小闭环
- 本次已盘点本地资源：主计划、当前状态、执行工作流、前端主线状态、automation memory、backup 相关脚本与测试
- 本次使用的本地 resources / skills / 记忆上下文：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`、`web/src/components/modules/setting/backup-logic.ts`、`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`
- 若未使用部分本地资源或上下文，原因：本轮聚焦同池最小闭环，不扩展到浏览器 smoke 或其他页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：N/A
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：主线程执行
- 若使用分目录 agent，负责目录与禁止越界范围：N/A
- 若未使用子 agent，原因：遵循本轮“不要创建子 agent”的明确要求，且任务范围集中

## 3. 本次硬规则

- 只处理 Phase G 同池内真实存在的中文泄漏，不扩散到其他主题
- 改动必须局限在 backup no-browser 链路可直接验证的文件内
- 保持英文 locale 行为不变，只收口非英文 UI 的已知 warning 直出

## 4. 本次禁止事项

- 不改后端导入契约
- 不重做 Backup 页面结构
- 不扩大到 route context / field / impact 全量本地化

## 5. 本次验收条件

- `node scripts/verify-backup-logic.mjs` 通过
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过
- 中文测试预期已与 warning 本地化后结果对齐

## 6. 本次回滚点

- `web/src/components/modules/setting/backup-logic.ts`
- `web/src/components/modules/setting/backup-logic.test.ts`
- `scripts/verify-backup-logic.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 shared helper 输出语义，再对齐 no-browser 验证
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/backup-logic.ts`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅影响中文 / 繁中 / 日文下的备份诊断文案显示，英文不变

## 8. 实施步骤

1. 复查 Phase G 文档、automation memory 与 backup 相关验证脚本，定位仍会直出的英文 warning 点。
2. 在 `backup-logic.ts` 中新增最小诊断词映射，仅收口 `current model not found`、`policy drift`、`missing candidate` 三类 warning / skip reason。
3. 对齐 `backup-logic.test.ts` 与 `verify-backup-logic.mjs` 的中文预期，并重新执行 no-browser 验证。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：`node scripts/verify-backup-logic.mjs`
- 专项验证：
  - `node web/node_modules/vitest/vitest.mjs run web/src/components/modules/setting/backup-logic.test.ts`
  - 结果：宿主机 `spawn EPERM`，属于已知环境阻塞，非本轮改动引入

## 10. 风险与兼容性

- 新风险：低；仅新增最小诊断词映射
- 兼容性风险：低；英文 locale 输出保持原样，其他 locale 仅收口已知英文泄漏
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：`verify-backup-logic.mjs` 通过；Vitest 仍受宿主 `spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：主计划、当前状态、工作流、前端主线状态、automation memory、backup helper / test / script
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主计划与前端主线状态：确认本轮应继续留在 Phase G 截图优先与中文收口池内
  - automation memory：确认上一轮已完成 channel page 闭环，本轮更适合挑一个剩余中文泄漏小点闭环
  - backup helper / test / script：确认 `current model not found` 等英文 warning 会进入中文诊断详情
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：N/A
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮仅需 no-browser 验证
- 待验证页面清单：Backup 详情区剩余 `routing / fallback_model / impact` 等内部词的中文化仍待下一轮继续收口
- 若未使用子 agent，原因：任务范围小且用户要求不创建子 agent
- worklog 是否更新：是
- 遗留项：本轮只收口了 warning / skip reason；上下文名、字段名、impact/billing/probe 等内部值仍可能在中文详情里保留英文
- 下一任务前置条件是否满足：满足