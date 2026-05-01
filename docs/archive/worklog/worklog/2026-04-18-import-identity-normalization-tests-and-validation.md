# Import Identity Normalization Tests And Validation

> 日期：2026-04-18
>
> 目标：围绕 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.2 / 11.5.3 / 11.5.4`，把本轮已经落地的导入身份归一化能力补成“有测试、有验证、有记录”的主线成果，并同步修正文案与当前实现不一致的说明。

## 1. 任务信息

- 任务名称：导入身份归一化测试与验证收口
- 日期：2026-04-18
- 当前阶段：backup/import 主线补测与验证
- 对应 milestone：备份/导入迁移主线，重点对应 `11.5.2 导入功能要求`、`11.5.3 迁移适配目标`、`11.5.4 导入后模拟路由验证 / 迁移差异预览`

## 2. 开工前输入

- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.2 / 11.5.3 / 11.5.4 / 13 / 14.11 / 14.12`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-17-import-task-full-backup-migration-and-group-cleanup.md`
  - `docs/worklog/2026-04-18-import-mode-merge-replace.md`
  - `docs/worklog/2026-04-18-route-preview-diff-granularity.md`
- 本次任务目标：
  - 为导入身份归一化补上关键回归测试
  - 验证 `channel_key / api_key` 不再依赖 snapshot 数字 ID 身份
  - 验证 `stats_channel / stats_api_key / relay_logs` 跟随本地 ID remap
  - 修正文案与当前实现不一致的导入说明
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/worklog/WORKLOG_TEMPLATE.zh-CN.md`
  - 上述 import 相关 worklog
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `internal/model/backup.go`
  - `web/src/components/modules/setting/Backup.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：
  - 当前线程 handoff 摘要
  - 已存在的 import mode / route preview / post-import validation 测试
  - 本地子 agent 协作规则
- 若未使用部分本地资源或上下文，原因：未展开与本轮测试无关的前端子模块和非 backup/import 主线文件，避免偏题
- 本次是否启用子 agent 与分工边界：是
- 本次子 agent 使用模型：`gpt-5.4`
- 若未使用某个子 agent 结论，原因：只采纳与本轮主线直接相关的只读结论，未把建议自动扩展到其他未开工主线

## 3. 本次硬规则

- 以 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 为主，不偏离 backup/import 主线
- 本地资源优先，不跳过已有 worklog、现有测试和当前线程上下文重造流程
- 子 agent 默认统一使用 `gpt-5.4`
- 子 agent 仅执行只读分析或限定文档范围，职责边界清晰，不与主线程争抢实现文件
- 主线程负责汇总结论、消除冲突、完成最终代码与记录落地
- 代码区与辅助区分离：项目代码改动留在项目文件，过程记录落在 `docs/worklog/`

## 4. 本次验收条件

- `internal/op/backup_test.go` 新增回归测试，覆盖导入身份归一化核心场景
- 新测试通过，且不破坏现有 `internal/op`、`internal/relay` 测试
- 能证明以下行为已经有代码证据：
  - `channel_key` 通过自然键复用本地 ID，而不是沿用 snapshot ID
  - `api_key` 通过 secret 复用本地 ID
  - `stats_channel / stats_api_key` 跟随 remap 到本地 ID
  - `relay_logs` 的 `channel_id / attempts[].channel_id / attempts[].channel_key_id` 跟随 remap
  - 无法解析的 `attempt channel key` 引用会被置 `0` 并产生 warning
- 前端导入说明不再继续错误声称“没有自动连通性检查”

## 5. 实施步骤

1. 重新核对 `backup.go` 当前导入事务流，确认 `prepare/import/remap` 实现已经落位
2. 盘点 `backup_test.go` 现有覆盖面，锁定缺口为“身份归一化 + stats/log remap”
3. 新增两组回归测试：
   - `channel_key + stats_channel + relay_logs`
   - `api_key + stats_api_key`
4. 运行 `gofmt` 与 Go 测试，确认新测试和既有包测试通过
5. 修正 `Backup.tsx` 中与当前实现不一致的导入说明
6. 补写本 worklog，记录本轮结论、验证命令与遗留项

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/op/backup_test.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 测试命令：`go test ./internal/server/handlers/...`
  - 结果：当前路径无测试文件，非失败
- 测试命令：`go test ./internal/relay/...`
  - 结果：通过
- 专项验证：
  - 新增 `channel_key` 身份归一化测试
  - 新增 `api_key` 身份归一化测试
  - 验证 `relay_logs` 中不可解析 `attempt channel key` 会置 `0` 并给出 warning
  - 验证 `stats_channel / stats_api_key` 仅导入可解析并已 remap 的记录

## 7. 收工记录

- 本轮新增/补强的测试：
  - `TestDBImportIncrementalMergeModeReusesExistingChannelKeyIDAndRemapsStatsAndLogs`
  - `TestDBImportIncrementalMergeModeReusesExistingAPIKeyIDAndRemapsStats`
- 本轮新增测试覆盖的行为：
  - 以本地自然键身份复用既有 `channel_key`
  - 以本地 secret 身份复用既有 `api_key`
  - `stats_channel / stats_api_key / relay_logs` 跟随 remap 到本地 ID
  - 无法解析的 relay attempt key 引用归零并告警
- 文案同步：修正 `web/src/components/modules/setting/Backup.tsx`，删除“未提供自动连通性检查”的过期描述，保留仍未完成的 `rollback / partial restore / mapping editor`
- 手工 smoke 状态：未做完整 UI 手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮聚焦后端回归测试与文档收口，未启动完整前端交互链路

## 8. 子 agent 与本地资源记录

- 本次子 agent 默认统一使用 `gpt-5.4`；若后续因权限、网关或环境原因不可用，应记录阻塞原因后回退到主线程串行方案或明确记录的替代模型
- 本次子 agent 分工：
  - 子 agent A：只读核对 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 中 `11.5.2 / 11.5.3 / 11.5.4 / 13 / 14` 对 backup/import 主线的完成度与缺口
  - 子 agent B：只读整理本轮 worklog 结构建议，确保符合 `AGENTS.md` 与 `WORKLOG_TEMPLATE.zh-CN.md`
  - 子 agent C：限定写入 `docs/worklog/2026-04-18-import-identity-normalization-tests-and-validation.md` 的草稿任务
- 各子 agent 职责边界：仅限只读分析或 `docs/worklog` 单文件写入，不直接修改实现代码文件
- 主线程职责：汇总各子 agent 结论，统一测试口径，消除冲突后完成测试代码、验证执行、前端说明修正与最终 worklog 落地
- 本轮实际采纳的子 agent 结论：
  - 已完成项与未完成项判断中，采纳了“`map` / rollback / mapping editor / partial restore 仍未完成”的结论
  - worklog 结构中，采纳了“必须明确记录本地资源优先、子 agent 默认 `gpt-5.4`、职责边界、主线程汇总、验证命令”的建议

## 9. 风险与遗留项

- 当前已完成的是“身份归一化 + stats/log remap”的测试与验证，不代表 `11.5.x` 全部完成
- 仍未完成的主线缺口：
  - `map` 模式后端与交互式映射输入
  - 回滚到上一个 snapshot
  - partial restore / selective import
  - mapping editor 与交互式 diff 确认
- `14.11` 原文要求“默认不泄露明文敏感信息”，而当前导出默认保留 secrets，这与近期个人迁移需求一致，但与计划原文存在偏离，需要后续明确是否更新 MD 或做双模式策略
- 仍需补的验证：
  - 基于真实导出文件的导入 smoke
  - Linux 服务器链路验证
  - Windows 本机导入/导出 UI 实测

