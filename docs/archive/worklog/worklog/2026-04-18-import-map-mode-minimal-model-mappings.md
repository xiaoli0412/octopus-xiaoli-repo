# Import Map Mode Minimal Model Mappings

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.2 / 11.5.3 / 11.5.4`，把当前只停留在文档和前端提示层的 `map` 主线推进成“后端可消费、前端可提交、测试可证明”的最小可用闭环。

## 1. 任务信息

- 任务名称：`map` 模式最小模型映射闭环
- 日期：2026-04-18
- 当前阶段：backup/import 主线继续推进
- 对应 milestone：`11.5.2` 冲突处理模式 `replace / merge / skip / map`，`11.5.4` apply 前允许用户确认或修正映射

## 2. 开工前输入

- 对应 canonical 章节：`11.5.2 / 11.5.3 / 11.5.4 / 14.12`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-18-import-identity-normalization-tests-and-validation.md`
  - `docs/worklog/2026-04-18-route-preview-diff-granularity.md`
- 本次任务目标：
  - 给导入链路增加最小可用的 `map` 能力
  - 先支持 `snapshot model -> current model` 的人工映射
  - 让 dry-run preview 与 apply 都真正吃到 `model_mappings`
  - 给前端导入页增加最小输入，不再只停在“Still not wired”提示
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - 既有 import worklog
  - `internal/model/backup.go`
  - `internal/server/handlers/setting.go`
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `web/src/api/endpoints/setting.ts`
  - `web/src/components/modules/setting/Backup.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：
  - 当前线程 handoff
  - 已有 dry-run / route preview / post-import validation / identity normalization 测试
  - 已有 Backup UI 与 API 结构
- 本次是否启用子 agent 与分工边界：是
- 本次子 agent 使用模型：`gpt-5.4`

## 3. 本次硬规则

- 继续以 backup/import 主线为唯一核心，不扩散到无关模块
- 本地资源优先，先复用现有 preview/import 流程，不重造并行实现
- 子 agent 默认统一使用 `gpt-5.4`
- 子 agent 仅做只读分析，主线程负责最终代码落地与验证
- 代码区与记录区分离，过程记录独立写入 `docs/worklog/`

## 4. 本次验收条件

- 后端支持 `mode=map`
- 后端支持接收 `model_mappings` 并在导入前重写模型引用
- dry-run preview 能体现映射后的 route candidate 变化
- apply 后导入结果能体现映射后的 route item
- 前端导入页可以提交 `model_mappings`
- `go test ./internal/op`
- `go test ./internal/relay/...`
- `pnpm build`

## 5. 实施步骤

1. 给后端模型定义增加 `DBImportModeMap` 与 `DBImportOptions`
2. 给导入 handler 增加 `model_mappings` 读取与 JSON 解析
3. 在 `internal/op/backup.go` 中新增 `DBImportIncrementalWithOptions`
4. 在导入前 clone dump，并统一应用模型映射预处理
5. 补回归测试，验证 `map` 对 route preview 和实际导入的影响
6. 给前端 `setting.ts` 增加 `map` 模式与 `modelMappings` 提交
7. 给 `Backup.tsx` 增加最小多行映射输入与解析校验
8. 运行 Go 与前端构建验证

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/model/backup.go internal/server/handlers/setting.go internal/op/backup.go internal/op/backup_test.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 测试命令：`go test ./internal/server/handlers/...`
  - 结果：当前路径无测试文件，非失败
- 测试命令：`go test ./internal/relay/...`
  - 结果：通过
- 前端验证：`pnpm build`
  - 结果：通过
- 新增专项测试：
  - `TestDBImportIncrementalMapModeAppliesModelMappingsToRoutePreviewAndImport`

## 7. 收工记录

- 本轮后端新增：
  - `DBImportModeMap`
  - `DBImportOptions`
  - `DBImportIncrementalWithOptions`
  - 导入前 clone dump 与 `model_mappings` 预处理
- 本轮后端最小闭环已覆盖：
  - `channels.model / custom_model`
  - `channel_keys.allowed_models`
  - `api_keys.supported_models`
  - `group_items.model_name`
  - `llm_infos.name`
- 本轮前端新增：
  - `setting.ts` 支持 `map` 与 `modelMappings`
  - `Backup.tsx` 在 `map` 模式下展示多行模型映射输入
  - 提交前做 `snapshot-model=current-model` 的轻量解析校验
- 本轮文案同步：
  - `map` 不再完全停留在“未接后端”状态
  - 仍保留未完成项：交互式 mapping editor / wizard、rollback、partial restore

## 8. 子 agent 与本地资源记录

- 本次子 agent 默认统一使用 `gpt-5.4`
- 子 agent 分工：
  - 子 agent A：只读分析 `map` 模式最小后端接入面，定位 preview/apply 必须接入的函数
  - 子 agent B：只读分析前端最小 UI 接线方案，限定 `setting.ts` 与 `Backup.tsx`
- 子 agent 职责边界：仅限只读分析，不修改实现文件
- 主线程职责：汇总子 agent 结论，决定最小实现切口，并完成最终代码、测试、构建与记录落地
- 本轮采纳的关键结论：
  - `model_mappings` 应作为导入请求选项，而不是普通 Setting
  - 最小闭环应优先支持模型映射，不先做复杂的交互式 wizard

## 9. 风险与遗留项

- 当前完成的是 `map` 的最小模型映射闭环，不等于 `11.5.2 / 11.5.4` 全量完成
- 仍未完成：
  - 交互式 mapping editor / conflict wizard
  - rollback to previous snapshot
  - partial restore / selective import
  - 更细粒度的 channel/group/provider 映射
- 当前 `map` 模式虽已可用，但还不是完整迁移向导；它更像是“显式模型映射 + 现有 dry-run/import 流程”的第一版骨架

