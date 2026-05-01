# 2026-04-24 Phase G Backup Context Prefix Localization Closure

## 1. 任务信息

- 任务名称：备份详情上下文前缀中文本地化收口
- 日期：2026-04-24
- 当前阶段：Phase G 截图优先 UI 收口
- 对应 milestone：Phase G settings / backup no-browser contract closure

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9 节前端收口与中文界面一致性要求
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3 节
- 上一个相关 worklog：`docs/worklog/2026-04-24-phase-g-backup-detail-metadata-localization-closure.md`
- 本次任务目标：继续收口备份导入详情里会直出的 `group:`、`channel:`、`api_key:`、`api_key_model:` 等英文上下文前缀，让中文主显示不再暴露内部 token 形态
- 本次已盘点本地资源：主计划、当前状态、执行工作流、前端主线状态、automation memory、最新 backup worklog、`backup-logic.ts`、`backup-logic.test.ts`、`verify-backup-logic.mjs`、`internal/op/backup_extra.go`、`internal/op/backup_test.go`
- 本次使用的本地 resources / skills / 记忆上下文：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/worklog/README.zh-CN.md`、`docs/worklog/2026-04-24-phase-g-backup-detail-metadata-localization-closure.md`、`C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`、`web/src/components/modules/setting/backup-logic.ts`、`web/src/components/modules/setting/backup-logic.test.ts`、`scripts/verify-backup-logic.mjs`、`internal/op/backup_extra.go`、`internal/op/backup_test.go`
- 若未使用部分本地资源或上下文，原因：本轮聚焦 backup 同池最小闭环，不扩展到浏览器 smoke、后端接口改动或其他页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：N/A
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：主线程执行
- 若使用分目录 agent，负责目录与禁止越界范围：N/A
- 若未使用子 agent，原因：遵循本轮“不要创建子 agent”的明确要求，且任务范围集中在同一批前端 helper / 断言文件

## 3. 本次硬规则

- 只处理 Phase G backup no-browser 同池中的真实中文泄漏，不扩散到其他主题
- 改动仅限 `backup-logic` 输出层、相关 no-browser 断言和必要记录文件
- 继续保持英文 locale 行为不变，只收口非英文 locale 中的上下文前缀泄漏

## 4. 本次禁止事项

- 不改后端导入结构或兼容性报告 schema
- 不重做 Backup 组件布局
- 不扩大到 route preview 全量渲染器或 post-import warnings 全部结构化改造

## 5. 本次验收条件

- `backup-logic.ts` 能识别并本地化 `group:`、`channel:`、`channel_key:`、`group_route:`、`api_key:`、`api_key_model:` 这类上下文前缀
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过
- `node scripts/verify-backup-logic.mjs` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/backup-logic.ts`
- `web/src/components/modules/setting/backup-logic.test.ts`
- `scripts/verify-backup-logic.mjs`
- `docs/worklog/2026-04-24-phase-g-backup-context-prefix-localization-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 shared helper 输出语义，再对齐 no-browser 断言
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/backup-logic.ts`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅影响非英文 locale 的备份详情上下文展示文本，英文输出保持原样

## 8. 实施步骤

1. 复核 `internal/op/backup_extra.go` 与 `backup_test.go` 中真实会出现的上下文前缀模式，确认本轮最小覆盖集。
2. 在 `backup-logic.ts` 增加前缀型上下文本地化逻辑，并补充相应的 zh-Hans 断言。
3. 运行 `tsc` 与 `verify-backup-logic.mjs`，同步更新 worklog / memory / 状态说明。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：`node scripts/verify-backup-logic.mjs`
- 专项验证：Vitest 仍未运行；本轮以现有 no-browser 脚本为主验证入口

## 10. 风险与兼容性

- 新风险：低；仅新增上下文文本格式化逻辑
- 兼容性风险：低；英文 locale 继续保留原 token 形式，中文 / 繁中 / 日文仅做显示层本地化
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：`node scripts/verify-backup-logic.mjs` 通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主计划、当前状态、工作流、前端主线状态、automation memory、backup helper / test / script、后端 backup 真实输出代码与测试
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主计划、当前状态、前端主线状态：确认本轮应继续留在 Phase G backup 同池，小范围收口中文界面英文泄漏
  - automation memory 与上一轮 worklog：确认 `warning / skip reason / scope badge / detail metadata` 已收口，下一步应落在剩余上下文前缀直出
  - `internal/op/backup_extra.go` 与 `internal/op/backup_test.go`：确认真实会输出 `channel:...`、`group:...`、`group_route:...`、`api_key:...`、`api_key_model:...`、`channels.model`、`group_items.model_name` 等模式
  - `backup-logic.ts`、`backup-logic.test.ts`、`verify-backup-logic.mjs`：提供了本轮最小可验证切口和 no-browser 验收链路
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：N/A
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮仅需 no-browser 验证
- 待验证页面清单：Backup 详情区后续仍可继续复查 `post-import warnings` 原始英文句子与 route preview 候选细项
- 若未使用子 agent，原因：任务范围小且用户要求不创建子 agent
- worklog 是否更新：是
- 遗留项：`backup-logic.ts` 中为规避当前宿主编码问题而保留的旧坏块注释仍待清理；不影响当前编译与 no-browser 验证，但下一轮应优先删除以保持文件整洁
- 下一任务前置条件是否满足：满足
