# Import Task - Full Backup Migration And Group Cleanup

> 目的：继续按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.7`、`10.4`、`11.5.1`、`11.5.2`、`11.5.3`、`11.5.4` 推进 Phase F 主线，把“全量备份 / 导入后尽量保持原生体验 / 分组残留清理 / relay 运行时兜底”做成真实可验证的闭环。

---

## 1. 任务信息

- 任务名称：Full Backup Migration And Group Cleanup
- 日期：2026-04-17
- 当前阶段：Phase F
- 对应 milestone：里程碑 6 备份 / 导入 / 迁移适配收口

## 2. 开工前输入

- 对应 canonical 章节：
  - 第 `9.7` 节项目级备份 / 导入 / 迁移适配 UI
  - 第 `10.4` 节备份 / 导入 / 迁移适配性能要求
  - 第 `11.5.1` 节备份功能要求
  - 第 `11.5.2` 节导入功能要求
  - 第 `11.5.3` 节迁移适配目标
  - 第 `11.5.4` 节导入后模拟路由验证 / 迁移差异预览
- 对应 workflow 章节：
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 中 Phase F 对应章节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-17-ui-group-home-followup.md`
- 本次任务目标：
  - 默认导出改为面向个人自用迁移的全量快照
  - 导入流程继续保留 `dry-run + incremental/skip`，但语义与全量迁移一致
  - 修复模型删除、渠道模型变更后 `group_items` 残留导致 relay 失败的问题
  - 为历史残留脏路由补运行时兜底，避免继续把失败放大到上游
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/worklog/README.zh-CN.md`
  - `internal/model/backup.go`
  - `internal/model/channel.go`
  - `internal/op/backup.go`
  - `internal/op/channel.go`
  - `internal/op/llm.go`
  - `internal/op/group.go`
  - `internal/server/handlers/setting.go`
  - `internal/server/handlers/channel.go`
  - `internal/task/sync.go`
  - `internal/relay/relay.go`
  - `web/src/components/modules/setting/Backup.tsx`
  - `internal/op/backup_test.go`
  - `internal/op/channel_test.go`
  - `internal/relay/relay_more_test.go`
- 本次使用的本地 resources / skills / 记忆上下文：
  - 直接复用当前线程 handoff，总结此前已完成的 `dry-run / import mode / compatibility report` 主线进展，避免重复阅读和误判
  - 直接复用仓库内 MD、已有 worklog、现有测试基建与 `go test / pnpm build` 命令作为本轮上下文基线
  - 复用本地 `codex-run-as-apply-patch` Windows 补丁落盘方案，降低命令行直接编辑错误率
- 若未使用部分本地资源或上下文，原因：
  - 未使用外部资料，因为本轮需求冲突与技术边界都可由仓库内 MD、代码与当前线程上下文完整确定
- 本次是否启用子 agent 与分工边界：
  - 是
  - 子 agent 只读负责：
    - canonical plan 与最新口头要求对齐
    - `group_items` 残留入口盘点
    - worklog 目录结构与记录边界建议
  - 主线程负责所有代码改动、冲突消解、测试验证与最终 worklog 回写
- 本次子 agent 使用模型：
  - `gpt-5.4`
- 子 agent 使用与阻塞说明：
  - 先复用已存在 agent 线程做只读核查，原因是环境一度触发线程上限
  - 后续关闭旧 agent 后重新开 `gpt-5.4` explorer 做 worklog 边界建议
  - 未使用其他模型作为回退

## 3. 本次硬规则

- 主线必须服从 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`，但若与用户最新口头要求冲突，以用户最新要求为准，并在 worklog 中显式记录
- 只推进 Phase F 相关主线，不扩散到无关功能
- 先做真实可交付闭环，不伪装已完成 `replace / map / rollback` 这类未落地能力
- 对 dirty worktree 只做最小必要改动，不回退不相关用户改动
- 子 agent 默认使用 `gpt-5.4`

## 4. 本次禁止事项

- 不把全量迁移主线混写进旧 UI 或旧 backend task worklog
- 不伪造“已完成自动健康检查 / 自动回滚 / mapping editor”
- 不用危险命令清理工作区
- 不为了记录整理去改动项目功能代码之外的无关文件

## 5. 本次验收条件

- 默认导出快照包含渠道密钥、客户端 API Key 与迁移所需项目配置
- `manifest.contains_secrets` 与导出内容一致，checksum 计算与 manifest 内容不打架
- `ChannelUpdate` 在模型列表收缩时会删除失效 `group_items`
- `LLMDelete / LLMBatchDelete` 删除模型后不会留下分组残留项
- 导入时会跳过与渠道声明模型不一致的无效 `group_items`
- relay 遇到历史残留 route item 时会跳过，而不是继续向上游发失败请求
- `go test ./internal/op ./internal/relay/...` 通过
- `pnpm build`（`web/`）通过

## 6. 本次回滚点

- 备份 / 导入语义改动可按文件维度回退：
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `web/src/components/modules/setting/Backup.tsx`
- 分组残留清理与运行时兜底可按文件维度回退：
  - `internal/op/channel.go`
  - `internal/op/llm.go`
  - `internal/model/channel.go`
  - `internal/relay/relay.go`
  - `internal/relay/relay_more_test.go`

## 7. 实现范围

- 先改数据语义还是先改 UI：先把后端真实行为收口，再同步前端语义
- 受影响后端模块：
  - `internal/op/backup.go`
  - `internal/op/channel.go`
  - `internal/op/llm.go`
  - `internal/model/channel.go`
  - `internal/relay/relay.go`
- 受影响前端模块：
  - `web/src/components/modules/setting/Backup.tsx`
- 受影响测试：
  - `internal/op/backup_test.go`
  - `internal/relay/relay_more_test.go`
- 是否影响旧数据：
  - 是，导入与运行时会更积极地过滤或删除失效 `group_items`
- 是否影响旧接口或旧 UI：
  - 是，设置页备份/导入文案与默认导出语义已发生变化

## 8. 实施步骤

1. 重新对齐 canonical plan 与最新口头要求，确认 Phase F 本轮只做“全量 snapshot -> 可理解导入 -> group 清理 -> relay 兜底”主链。
2. 把默认导出调整为全量迁移快照，并修正 `manifest.contains_secrets` / `checksum` 一致性。
3. 完成 `ChannelUpdate` 模型收缩后的 `group_items` 清理。
4. 完成 `LLMDelete / LLMBatchDelete` 删除模型后的 `group_items` 清理与测试。
5. 在导入 remap 阶段跳过“渠道未声明该模型”的无效 route item。
6. 在 relay 运行时补 `channel.SupportsModel(item.ModelName)` 兜底，避免历史残留 route item 继续打上游。
7. 修正设置页 `Backup` 文案，使其与“个人自用全量迁移快照”语义一致。
8. 运行 `go test` 与 `pnpm build` 收口，并回写 worklog。

## 9. 已完成项

- 已完成 `DBExportAll` 默认全量导出方向调整：
  - 不再默认执行脱敏导出
  - `manifest.contains_secrets` 改为根据真实内容计算
  - `checksum` 计算顺序已修正，避免与 manifest 实际内容不一致
- 已完成设置页 `Backup` 的主语义修正：
  - 导出文案改成面向新服务器直接恢复的 full-project snapshot
  - 导入 apply 文案改成“若快照中存在凭据，则会导入凭据”，不再保留错误的“不会恢复 secrets”描述
- 已完成 `ChannelUpdate` 在模型列表收缩时删除失效 `group_items`
- 已完成 `LLMDelete / LLMBatchDelete` 删除模型后同步清理引用它们的 `group_items`
- 已完成导入阶段对无效 route item 的静态过滤：
  - 若导入时解析到的目标 channel 不声明该模型，则跳过该 `group_item`
  - warning 会明确指出跳过原因
- 已完成 relay 运行时兜底：
  - 若历史 `group_item` 指向的模型不在 channel 的 `model/custom_model` 声明中，直接 skip，不再继续走 key / upstream 调用
  - race fallback 也同步补了这一层 skip 保护
- 已补充并通过如下测试：
  - 默认导出包含 secrets 的测试
  - 渠道模型收缩清理 `group_items` 的测试
  - 模型删除 / 批量删除清理 `group_items` 的测试
  - 导入跳过无效 `group_items` 的测试
  - relay 跳过历史残留 route item 的测试
- 已完成导入后静态校验报告（post-import validation）第一版：
  - 导入 apply 后会静态扫描当前 group / channel / key 状态
  - 返回 `post_import_validation`，显式给出：
    - `degraded_groups`
    - `empty_groups`
    - `disabled_channels`
    - `channels_without_keys`
    - `stale_items_removed`
    - `route_warnings`
  - 当发现“缺失 channel / 渠道未声明该模型”这类可安全清理的历史脏 route item 时，会在导入后自动清理并计入 `rows_affected.cleaned_group_items`
- 已修正全量导入时的零值字段保真问题：
  - 新导入 channel 的 `enabled=false` 等字段现在会按快照原样落库，不再被数据库默认值错误覆盖
  - 同类保真处理同步补到了新导入的 `channel_keys` 与 `api_keys`
- 已完成设置页 `Backup` 对 post-import validation 的最小接线：
  - apply 结果区新增 post-import validation 摘要卡片与列表
  - 文案显式区分“导入前 compatibility”与“导入后静态校验”
  - 底部未完成功能提示已改为“自动连通性检查尚未接入，当前仅为静态校验”
- 已完成导入后自动健康检查基础版：
  - 基于现有 `channel/test-models` 真实探测链路，抽出 `internal/op/healthcheck.go` 作为后端可复用实现
  - 导入 apply 后会对当前保留下来的 group route candidates 做一轮受控健康检查，并把结果挂到 `post_import_validation.health_check`
  - 当前健康检查规则：
    - 先做 Base URL 连通性检查
    - 再做最小 LLM 请求探测
    - `2xx` 视为通过
    - `429` 视为可达且通过（`rate_limited=true`）
    - disabled / 无 key / 未声明模型 / 无 adapter 等场景会以 `skipped` 结果显式返回
  - 设置页 `Backup` apply 结果区已新增 `Post-import health check` 摘要与明细展示
- 已完成导入后健康检查的目标收窄与保守并发控制：
  - `buildPostImportHealthCheck` 不再扫描当前数据库里的全部 group items
  - 健康检查 target 改为基于“本次导入成功保留下来的 route-target”构造，并在导入后按真实仍存在的 `group_items` 再过滤一轮
  - 若 compatibility 已标出 `affected_groups`，健康检查会优先收敛到这些受影响分组，避免把历史无关分组一起纳入导入后探测
  - `RunImportHealthChecks` 已从串行执行改为小并发 worker 池，默认上限保持保守，贴近 `10.4` 的“分批、错峰、避免压垮上游”要求
  - 并发执行下保留稳定去重与排序，确保 `post_import_validation.health_check.checks` 输出顺序不抖动
- 已补充与通过两条关键回归测试：
  - `TestDBImportIncrementalApplyHealthCheckTargetsOnlyImportedRoutes`
  - `TestRunImportHealthChecksHonorsConcurrencyLimit`
- 已完成 `11.5.4` 的结构化候选链差异预览第一版：
  - `compatibility` 新增 `route_preview_diffs`
  - 粒度先收敛到 `(group, model)`
  - 每条 diff 包含导入前候选链、预计导入后候选链、removed / added candidates、`fallback_changed`、`skip_reasons`
  - 当前候选项先保留最小字段：`channel_name / model / priority / weight / enabled / declared / has_key / reason`
  - 前端 `Backup` dry-run 结果区已新增最小展示：`Route Preview Diff` 摘要卡片 + `Route preview diffs` 列表

## 10. 测试与验证

- 构建命令：
  - `pnpm build`（`web/`）
- 测试命令：
  - `go test ./internal/op`
  - `go test ./internal/op ./internal/relay/...`
- 实际结果：
  - `go test ./internal/op`：通过
  - `pnpm build`：通过
  - `go test ./internal/op ./internal/relay/...`：通过
  - 在补齐 post-import validation 与导入零值字段保真后，再次执行 `go test ./internal/op ./internal/relay/...`：通过
  - 在前端补齐 post-import validation 展示后，再次执行 `pnpm build`：通过
  - 在补齐 post-import health check 基础版后，再次执行 `go test ./internal/op ./internal/server/handlers/... ./internal/relay/...`：通过
  - 在前端补齐 health check 展示后，再次执行 `pnpm build`：通过
  - 在补齐“目标收窄 + 小并发 worker 池”后，再次执行 `go test ./internal/op ./internal/server/handlers/... ./internal/relay/...`：通过
  - 在补齐本轮后端改动后，再次执行 `pnpm build`：通过
  - 在补齐 `route_preview_diffs` 结构化差异预览后，再次执行 `go test ./internal/op ./internal/server/handlers/... ./internal/relay/...`：通过
  - 在前端接入 `route_preview_diffs` 最小展示后，再次执行 `pnpm build`：通过

## 11. 风险与兼容性

- 需求覆盖风险：
  - canonical plan 在 `11.5.1` 原本强调“默认导出不泄露明文敏感信息”
  - 当前实现已按用户最新口头要求改成“默认全量迁移快照，包含凭据”，这里属于需求覆盖 canonical，已明确按用户最新要求执行
- 兼容性风险：
  - 旧的脱敏快照仍会被识别并给出 `snapshot does not include plaintext credentials` warning
  - 但当前 UI 和默认导出已不再走安全默认值语义
- 数据风险：
  - 导入与渠道/模型变更过程现在会更积极地删除失效 `group_items`，这是为避免 relay 失败而做的兼容性清理
- 运行时风险：
  - `group_items` 仍然是系统当前主通路之一，`internal/op/group.go` 仍保留完整 CRUD 与缓存依赖，后续若要继续贴主干，仍需进一步收敛 `group_items` 语义

## 12. 手工 smoke 状态

- 手工 smoke：未做完整页面联调
- 阻塞原因：当前以主线代码收口与自动化测试为主，未启动完整前后端联调环境
- 待验证页面：
  - `Setting > Backup`
  - 导入 dry-run 结果区
  - 导入 apply 结果区
  - 渠道模型变更后的 group 调用链路

## 13. 子 agent 与本地资源记录

- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 当前线程 handoff：明确此前已完成 `incremental/skip + compatibility report + UI mode selector`，避免重复劳动
  - `AGENTS.md`：明确必须优先使用本地资源、子 agent 默认 `gpt-5.4`、且要把阻塞原因写回文档
  - `docs/worklog/README.zh-CN.md`：明确本轮应新开 `import-task-*` worklog，而不是把 Phase F 主线混写进旧文件
  - `internal/task/sync.go` 与 `internal/op/llm.go`：证明自动同步与模型删除路径已具备部分 `group_items` 清理逻辑，可继续沿该方向收口
  - `internal/server/handlers/channel.go`：证明现有仓库已有 channel/model 连通性测试入口，后续“导入后验证”可以在此基础上推进
- 本次使用的子 agent：
  - 只读 agent `Nash / Kant / Linnaeus`：用于 canonical 对齐、`group_items` 残留入口盘点、前端文案核查
  - 只读 agent `Carson`（`gpt-5.4`）：用于 worklog 目录边界与命名建议
  - 只读 agent `Erdos / Hypatia`（`gpt-5.4`）：用于 `10.4 / 11.5.2 / 11.5.3 / 11.5.4` 约束复核与“目标收窄 / 并发限制”测试缺口梳理
- 子 agent 分工、负责范围与产出摘要：
  - `Nash`：对齐 `11.5` 与最新用户口头要求，确认本轮应先做最小真实闭环，而不是伪装完成完整迁移平台
  - `Kant`：盘点 `group_items` 残留入口，确认 `channel 删除 / llm 删除 / sync 自动清理` 已补，但 `group.go` 仍是主残留面
  - `Linnaeus`：核查 `Backup.tsx` 与测试文案边界
  - `Carson`：建议新增独立 worklog 文件 `2026-04-17-import-task-full-backup-migration-and-group-cleanup.md`
- 若未使用子 agent，原因：
  - 不适用
  - 本轮一度触发 agent thread limit，上限释放后继续按 `gpt-5.4` 重开只读 agent；阻塞期间主线程按 AGENTS.md 要求改为串行继续推进，不丢上下文

## 14. 遗留项

- 仍未完成 canonical `11.5.2` 中的 `replace / merge / map / rollback`
- 仍未完成“导入后自动健康检查 / 自动路由验证 / 自动价格规则验证 / 自动模型别名验证”的完整闭环
- 当前已完成“导入后静态校验 + 可安全清理项自动清理报告 + 导入后自动健康检查基础版”
- 当前健康检查仍是基础版，尚未完成：
  - 导入前后候选链差异对比
  - 自动路由验证 / 自动价格规则验证 / 自动模型别名验证
  - 更细粒度的 route-target 级策略（例如 probe policy / billing mode / fallback 差异）仍未接入
- 当前 `route_preview_diffs` 仍是第一版，尚未完成：
  - 交互式映射确认 / 修正
  - alias 映射变化的独立结构化展示
  - route-target `(channel,key,model)` 级精细 diff
  - 价格规则 / billing mode / probe policy 级差异预览
- `internal/op/group.go` 仍保留完整 `GroupItem` CRUD / 缓存主路径，后续如要更贴近主干，需要进一步收敛数据模型与缓存语义

## 15. 下一任务前置条件

- 是否满足进入下一任务的前置条件：
  - 基本满足
- 建议下一步：
  1. 继续推进导入后验证：批量 channel/model 连通性测试
  2. 为导入 apply 增加“受影响 group / 清理结果 / 不可用 route-target”摘要
  3. 继续收口 `group_items` 的历史残留清理与 `group.go` 主路径收敛
