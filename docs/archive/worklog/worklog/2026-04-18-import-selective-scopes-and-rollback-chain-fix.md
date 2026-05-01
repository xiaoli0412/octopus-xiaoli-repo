# Import Selective Scopes And Rollback Chain Fix

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.3 / 11.5.4`，把 selective import 从“后端半接入”推进到“前后端可用 + 有回归测试”，并修复 rollback 流程里会污染 latest snapshot 的快照链 bug。

## 1. 任务信息

- 任务名称：selective import 最小闭环 + rollback 快照链修复
- 日期：2026-04-18
- 当前阶段：backup/import 主线持续推进
- 对应 milestone：`11.5.3` 选择性导入 / 部分恢复，`11.5.3` 一键回滚到上一 snapshot，`11.5.4` 导入后迁移差异预览与安全应用

## 2. 开工前输入

- 对应 canonical 章节：`11.5.2 / 11.5.3 / 11.5.4 / 验收 / 里程碑`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-18-import-map-mode-minimal-model-mappings.md`
  - `docs/worklog/2026-04-18-import-rollback-latest-snapshot-minimal-closure.md`
  - `docs/worklog/2026-04-18-import-identity-normalization-tests-and-validation.md`
- 本次任务目标：
  - 修复 rollback 调用导入时再次自动保存 pre-import snapshot，导致 `latest-import-snapshot` 被错误覆盖的问题
  - 给 selective import 补后端回归测试，证明未选中的 domain 不会被覆盖
  - 把 selective import 前端接入到现有 Backup 页面
  - 修正 rollback 前端返回字段命名与后端 JSON 不一致的问题
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `internal/model/backup.go`
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `internal/server/handlers/setting.go`
  - `web/src/api/endpoints/setting.ts`
  - `web/src/components/modules/setting/Backup.tsx`
  - 已有 2026-04-18 worklog
- 本次使用的本地 resources / skills / 记忆上下文：
  - 当前线程 handoff 摘要
  - repo-local skills：`brainstorming`、`dispatching-parallel-agents`、`using-superpowers`
  - 已存在的导入/回滚/映射实现与测试
- 本次是否启用子 agent 与分工边界：是
- 本次子 agent 使用模型：`gpt-5.4`

## 3. 本次硬规则

- 继续只推进 backup/import 主线，不偏离到无关模块
- 本地资源优先，先复用已有 import/preview/rollback 结构
- 子 agent 仅做只读分析，避免与主线程改同一文件
- 辅助记录与项目代码分离，worklog 独立写入 `docs/worklog/`

## 4. 本次验收条件

- rollback 后不会改写 `latest-import-snapshot` 指向错误快照
- selective import 在后端有回归测试覆盖“仅导入 settings 时其他域保持不变”
- 前端可以提交 `import_scopes`
- 前端 rollback 按钮能正确读取后端返回的 snapshot 名称
- `go test ./internal/op`
- `pnpm build`

## 5. 实施步骤

1. 在 `DBImportOptions` 增加内部选项 `SkipPreImportSnapshot`
2. 修改 rollback 流程，回滚导入时跳过自动 pre-import snapshot 保存
3. 补 rollback 快照链回归测试
4. 补 selective import 作用域回归测试
5. 给前端导入 API 增加 `DBImportScopes` 与 `import_scopes` multipart 提交
6. 在 Backup 页面加入 selective import 开关与 6 个 scope 开关
7. 修正 rollback 返回字段 camel/Pascal 与后端 snake_case 不一致的问题
8. 运行 Go 测试与前端构建验证

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/model/backup.go internal/op/backup.go internal/op/backup_test.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 前端构建：`pnpm build`
  - 结果：通过
- 新增专项测试：
  - `TestDBRollbackLatestImportSnapshotRestoresPreviousState`
    - 补充断言 rollback 前后的 latest snapshot metadata 不被污染
  - `TestDBImportIncrementalSelectiveImportScopesOnlyApplyChosenDomains`
    - 验证仅启用 `settings` scope 时，routing/models/api_keys/stats/logs 都保持现状

## 7. 收工记录

- 本次后端新增/修正：
  - `DBImportOptions.SkipPreImportSnapshot`
  - rollback 导入路径跳过自动 snapshot 保存
  - selective import 作用域行为回归测试
- 本次前端新增/修正：
  - `setting.ts` 支持 `DBImportScopes` 与 `import_scopes`
  - Backup 页面支持 selective import 开关与 scope 选择
  - map 输入框仅在模型相关 scope 有意义时展示
  - rollback toast 使用 `snapshot_name / snapshot_path`
- 本次子 agent：
  - `019d9d59-8a18-7e43-9cd9-e8c2ac8d6639`，模型 `gpt-5.4`
  - 负责范围：只读分析 `setting.ts` 与 `Backup.tsx` 的 selective import 最小接入方案
  - 结论摘要：selective import 应保持为独立 scope 维度，不能混成新的 `mode`；最小提交方案是在 multipart 中新增 `import_scopes`
- 当前对 MD 的推进结论：
  - `11.5.2`：`dry-run / diff / merge / replace / skip / map / post-import validation / rollback-latest` 已有可用闭环
  - `11.5.3`：selective import 已形成最小闭环；rollback latest snapshot 语义修正完成；更细粒度 channel/group/provider 映射仍未完成
  - `11.5.4`：route preview diff、alias preview、policy diff、post-import health check 已有基础闭环；交互式确认/修正映射仍未完成
- 遗留项：
  - selective import 目前是 domain 级布尔范围，不是更细粒度 partial restore
  - rollback 仍只有 latest snapshot，暂无历史浏览、预览与多快照选择
  - 前端尚无交互式 conflict wizard / mapping editor / rollback preview
