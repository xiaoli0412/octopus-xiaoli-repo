# Octopus AI 自动化中心实施计划

> 当前范围（2026-04-23 主线规划版）：本文档是 `AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习` 的中文实施计划。需求口径见 [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md)。
>
> 执行硬规则：任何实现必须先对齐 canonical plan、用户需求总账、详细工作流和动态路由文档。第一阶段默认不覆盖用户手动配置，不触发静默迁移。

---

## 1. 实施目标

本主线要分阶段把 AI 自动化能力接入 Octopus：

1. 先完成文档与主线同步。
2. 再落地 AI 自动化的后端模型、设置和任务框架。
3. 再增加顶层 `AI 自动化` 页面和设置页来源切换。
4. 再实现动态路由本地 AI 学习。
5. 最后逐步扩展智能分组、渠道识别、价格识别和审计回滚。

关键不变量：

- AI Profile 与用户手动配置双轨并存。
- AI 生成结果不能静默覆盖 `channels`、`groups`、`group_items`、`llm_infos` 或 `route_target_overrides`。
- 动态路由 AI 学习只影响运行时推荐，不永久改写用户配置。

---

## 2. 阶段划分

### Phase H1：文档与主线并入

目标：把该主线写入项目所有关键 MD。

工作项：

- 新增中英文需求文档。
- 新增中英文实施计划。
- 新增 worklog。
- 更新 canonical plan。
- 更新用户需求总账并追加需求 54-64。
- 更新详细工作流并新增 Phase H。
- 更新当前状态、前端主线状态、环境与下一步计划。
- 更新动态路由中英文需求与实施计划，移除“在线学习是非目标”的旧口径。

验收：

- 关键文档可搜索到 `AI 自动化`、`AI Profile`、`dynamic_routing_learning_enabled`、`不覆盖用户配置`。

### Phase H2：后端数据模型

目标：定义 AI 自动化持久化结构。

新增模型：

- `AIAutomationConfig`
- `AITask`
- `AITaskStep`
- `AIPromptTemplate`
- `AIProfile`
- `AIProfileVersion`
- `DynamicRouteLearningState`

新增设置项：

- `ai_automation_enabled`
- `ai_automation_base_url`
- `ai_automation_model`
- `ai_automation_use_local_default`
- `config_source_mode`
- `active_ai_profile_id`
- `dynamic_routing_learning_enabled`

验收：

- migration 可重复执行。
- 默认值安全。
- 旧数据不受影响。

### Phase H3：后端 API 与任务框架

目标：提供 AI 配置、任务、模板、Profile 和动态学习状态接口。

新增 API：

- `GET /api/v1/ai/config`
- `POST /api/v1/ai/config`
- `POST /api/v1/ai/models/fetch`
- `GET /api/v1/ai/prompt-templates`
- `POST /api/v1/ai/prompt-templates`
- `POST /api/v1/ai/tasks`
- `GET /api/v1/ai/tasks/:id`
- `GET /api/v1/ai/tasks`
- `POST /api/v1/ai/tasks/:id/cancel`
- `GET /api/v1/ai/profiles`
- `GET /api/v1/ai/profiles/:id`
- `POST /api/v1/ai/profiles/:id/activate`
- `GET /api/v1/dynamic-routing/learning`
- `POST /api/v1/dynamic-routing/learning/reset`

要求：

- AI 任务先支持短任务同步或轻量异步状态更新。
- 任务步骤必须能驱动前端进度条。
- Profile 激活必须只修改来源设置，不覆盖手动配置表。

### Phase H4：前端顶层栏目

目标：新增 `AI 自动化` 顶层入口和页面骨架。

页面模块：

- AI 模型状态卡片。
- AI endpoint / API Key / 模型配置区。
- 自动获取模型列表。
- 任务类型卡片。
- 自然语言输入框。
- 内置提示词与自定义提示词区。
- 任务进度条。
- 结果预览和 AI Profile 保存区。
- 历史任务列表。

验收：

- 导航可进入页面。
- 页面保持原项目深绿、圆角、渐进展开风格。
- 中文界面不泄漏 raw enum 或 i18n key。

### Phase H5：AI Profile 双轨切换

目标：设置页提供配置来源切换。

要求：

- 支持 `manual` 与 `ai_profile`。
- 选择 `ai_profile` 时必须选择一个 Profile。
- Profile 无效时自动回退 `manual`。
- UI 明确说明 AI 方案不会覆盖手动配置。

测试：

- `manual -> ai_profile -> manual` 切换后，用户原配置仍完整存在。
- 无效 Profile 不影响运行时可用性。

### Phase H6：动态路由 AI 学习

目标：实现本地、可解释、按 route-target 粒度学习的动态路由专线。

数据粒度：

- `(channel_id, channel_key_id, model_name)`

学习字段：

- `sample_count`
- `success_ewma`
- `failure_ewma`
- `latency_ewma_ms`
- `fallback_ewma`
- `race_win_ewma`
- `score`
- `confidence`
- `last_status_code`
- `last_outcome`
- `last_outcome_at`
- `updated_at`

接入点：

- relay attempt 成功后记录成功、延迟和 token / cost 参考。
- relay attempt 失败后记录失败、状态码和错误分类。
- race fallback winner 记录 race winner。
- 推荐评分读取学习状态并参与 `hybrid` 评分。

硬规则：

- 关闭 `dynamic_routing_learning_enabled` 后不参与评分。
- 不修改用户分组和 priority。
- `strict-mechanism` 保持确定性机制路径。
- `shadow-ai` / `metrics-only` 可以记录和审计，但不改变执行顺序。

### Phase H7：AI 生成分组与渠道识别 MVP

目标：第一批 AI Profile 生成能力。

范围：

- 智能分组建议。
- 渠道类型识别。
- 模型归类建议。
- 配置健康检查。

要求：

- 只生成 Profile。
- 只做预览、解释、保存、切换。
- 不静默应用到手动配置。

### Phase H8：价格识别与智能分组增强

目标：扩展 AI 自动化任务。

范围：

- 价格与计费模式识别。
- source type 建议。
- canonical name 建议。
- 更完整的分组策略模板。

要求：

- 价格识别只生成建议，不影响业务路由排序。
- 写入手动价格规则必须另走 diff / confirm / rollback 流程。

### Phase H9：审计、diff 与回滚增强

目标：为未来“选择性应用 AI 建议到手动配置”预留安全边界。

范围：

- Profile diff。
- 选择性应用。
- 审计日志。
- 回滚点。
- 风险确认。

第一阶段不默认开放该能力。

---

## 3. 文件落点

后端预期文件：

- `internal/model/ai_automation.go`
- `internal/op/ai_automation.go`
- `internal/server/handlers/ai_automation.go`
- `internal/model/dynamic_route_learning.go`
- `internal/op/dynamic_route_learning.go`
- `internal/db/migrate/012.go`
- `internal/relay/dynamic_learning.go`

前端预期文件：

- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/api/endpoints/ai-automation.ts`
- `web/src/components/modules/setting/ConfigSource.tsx`
- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/route/config.tsx`
- `web/src/components/modules/navbar/nav-store.ts`
- `web/public/locale/*.json`

文档文件：

- `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`
- `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.md`
- `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`
- `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.md`

Closure implementation addendum:

- Keep `AIProfile` and `AIProfileVersion` as audit and legacy compatibility shells.
- Add domain typed result tables: `AIGroupingProfile`, `AIChannelRecognitionProfile`, `AIPriceRecognitionProfile`, `AIModelClassificationProfile`, and `AIConfigHealthProfile`.
- Use `config_health_check` as the canonical domain string; `config_health` is legacy compatibility only.
- `AIProfileGet` detail assembles `domain_payload_type`, `domain_payload`, and `migration_status` from typed tables first, then falls back to legacy `content_json` only for old records.
- Persist AI task config/context/prompt/model/checkpoint fields in DB; in-memory maps may only hold current-process cancel handles.
- Recover tasks only from safe checkpoints: `ready_to_collect_context`, `ready_to_select_model`, `ready_to_call_ai`, `ready_to_parse`, `ready_to_generate_profile`, and `ready_to_save_result`.
- `GET /api/v1/ai/tasks` is the history MVP endpoint and must support pagination, status/type/domain filters, keyword search, and created time range filters.

Closure acceptance addendum:

- typed payload is the main profile consumption contract; legacy `content_json` is audit/fallback only.
- History task UI lists saved tasks with pagination, filters, keyword search, and detail reuse while keeping live task progress intact.
- Multi-AI split modes currently execute as frontend lane fan-out over multiple independent `/api/v1/ai/tasks` calls; backend orchestrator graph/state remains a later milestone.
- Restart recovery uses safe DB checkpoints instead of process memory as the only source of truth.
- CI gate includes Vitest and `go test ./internal/update -count=1` in both validation and release workflows.
- Backup advanced migration remains explicitly partial and must not be reported as closed by this AI automation milestone.

---

## 4. 验证计划

文档阶段：

- 搜索 `AI 自动化`、`AI Profile`、`dynamic_routing_learning_enabled`、`不覆盖用户配置`。
- 确认中英文 AI 自动化文档成对存在。
- 确认动态路由文档不再把在线学习写成非目标。

后端阶段：

- `go test ./internal/model ./internal/op ./internal/server/handlers ./internal/relay/... ./internal/update -count=1`
- migration 重复执行测试。
- API handler 测试。
- Profile 激活不覆盖原表测试。
- 动态学习开关与 scoring 测试。

前端阶段：

- `pnpm --dir web exec tsc --noEmit`
- `pnpm --dir web run test`
- locale 一致性检查。
- AI 自动化页面组件测试。
- 设置页配置来源切换测试。
- 动态路由学习开关测试。

端到端阶段：

- 创建 AI 任务并生成 Profile。
- 激活 Profile 后切回手动配置。
- 确认手动配置未被覆盖。
- relay 动态路由学习记录与推荐评分可解释。

---

## 5. 风险与回退

- 风险：AI Profile 语义被误解为覆盖原配置。回退：UI 和 API 均强制 preview / activate 双轨，不写原表。
- 风险：默认免费模型选择被误用于业务路由。回退：文档和测试明确它只用于 AI 自动化执行模型。
- 风险：动态学习改变用户优先级。回退：只在推荐评分层读取，不写 group item，不改 priority。
- 风险：任务中心范围过大。回退：Phase H7 之前只做任务框架和动态路由学习主线。

---

## 6. 第一阶段完成判定

第一阶段完成时必须满足：

1. 文档主线全部同步。
2. AI 自动化栏目存在。
3. AI 配置、本机默认和模型发现可用。
4. 自然语言任务和进度条可用。
5. AI Profile 可保存、预览、激活和回退。
6. 设置页可手动切换手动配置和 AI Profile。
7. 动态路由 AI 学习状态可记录、查询、清空，并受开关控制。
8. 所有关键测试和 worklog 已补齐。
