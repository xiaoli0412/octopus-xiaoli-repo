# Octopus AI 自动化中心需求说明

> 当前范围（2026-04-23 主线规划版）：本文档定义 `AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习` 主线。它是后续实现、审阅、修改和检索的中文需求入口。
>
> 执行硬规则：本文档必须与 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)、[USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 和 [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 保持一致。若实现需要偏离，必须先更新文档，再改代码。

---

## 1. 产品目标

AI 自动化中心要成为 Octopus 内所有 AI 辅助运维任务的统一入口，而不是分散在渠道、分组、模型价格或设置页的零散按钮。

当前主目标：

- 新增顶层 `AI 自动化` 栏目，与首页、渠道、分组、模型/价格、日志、设置并列。
- 支持用户自定义 AI 使用的渠道、模型、`base_url` 与 API Key。
- 未自定义时默认使用本机 Octopus OpenAI-compatible 地址，优先从当前服务地址推导，兜底为 `http://127.0.0.1:8080/v1`。
- 支持自动获取模型列表，并默认推荐免费、近期成功率高、延迟较低的模型。
- 支持自然语言输入、内置提示词模板和用户自定义提示词。
- 让 AI 生成分组、渠道识别、价格识别、模型归类等建议，并保存为独立 AI Profile。
- 在设置页允许用户手动切换 `手动配置` 与指定的 `AI 生成方案`。
- 保证 AI 生成结果不覆盖用户手动配置，只通过显式切换改变读取来源。
- 把动态路由作为例外专线：动态路由保留本地机制与 AI 学习开关，AI 学习只影响运行时推荐，不永久改写分组、渠道或优先级。

---

## 2. 核心原则

### 2.1 双轨并存

- 用户手动配置是主资产，必须始终保留。
- AI 生成配置作为独立 AI Profile 保存，不能静默覆盖原表数据。
- `channels`、`groups`、`group_items`、`llm_infos`、`route_target_overrides` 等手动配置表不能被 AI 任务直接覆盖。
- 切换 AI Profile 只改变运行时或管理界面读取来源，不删除、不覆盖、不重排用户原配置。

### 2.2 显式切换

- 设置页必须提供 `手动配置 / AI 生成方案` 的手动切换。
- 选择 AI 生成方案时，必须指定一个明确的 AI Profile。
- AI Profile 缺失、损坏、未覆盖目标模型或未通过校验时，必须自动回退到手动配置。
- 第一阶段默认只允许保存、预览、切换，不开放静默“一键覆盖手动配置”。

### 2.3 AI 只做可解释建议

- AI 输出必须包含自然语言总结、结构化建议、置信度、风险说明和来源依据。
- AI 生成结果要可回看、可版本化、可比较。
- 未来如果开放“应用到手动配置”，必须先有 diff、确认、审计和回滚机制。

### 2.4 价格不驱动业务路由

- “免费优先、成功率高、延迟低”的默认模型选择只用于 AI 自动化任务自身选择执行模型。
- 该规则不能影响业务请求路由排序。
- 业务路由仍必须尊重用户显式设置和动态路由文档中的硬规则。

---

## 3. AI 自动化中心页面要求

AI 自动化中心至少包含以下区域：

- 当前 AI 模型状态：展示当前使用的 AI endpoint、模型、来源、是否本机默认。
- AI 渠道配置：支持自定义 `base_url`、API Key、渠道类型和模型。
- 模型发现：支持自动获取模型列表，并展示模型来源、可用性、免费/付费倾向、近期成功率和平均延迟。
- 任务类型卡片：至少包含智能分组、渠道识别、价格识别、模型归类、配置健康检查、动态路由说明整理。
- 自然语言输入框：允许用户直接描述需求，例如“根据这些渠道生成分组建议”。
- 提示词模板：提供内置模板，并允许用户追加自定义提示词和工作要求。
- 任务进度条：展示收集上下文、选择模型、调用 AI、解析输出、生成方案、保存结果等阶段。
- 结果区：展示 AI 总结、结构化 Profile、风险提示、可执行项和后续建议。当前多 AI 工作流会以前端并行拉起多个 lane 任务，不承诺存在服务端 orchestrator。
- 历史任务区：展示任务输入、使用模型、生成时间、状态、结果和关联 Profile。

---

## 4. AI Profile 要求

AI Profile 是 AI 生成配置的独立保存单元。

必须支持的 Profile 类型：

- `grouping`：AI 生成的分组方案。
- `channel_recognition`：AI 渠道识别结果。
- `price_recognition`：AI 价格与计费识别结果。
- `model_classification`：AI 模型归类和规范名称建议。
- `config_health`：AI 配置健康检查结果。

每个 AI Profile 必须至少记录：

- 名称、领域、状态、版本。
- 生成任务 ID。
- 原始自然语言输入。
- 使用的 AI endpoint 与模型。
- 结构化内容。
- 置信度。
- 风险说明。
- 创建时间和更新时间。
- 是否当前激活。

AI Profile 的结构化内容必须可以被后续前端预览和后端校验，不允许只保存一段不可解析的自然语言。typed payload 现在是主消费合同；legacy `content_json` 保留为审计与兼容回退层。

---

## 5. 设置页切换要求

设置页需要新增“配置来源”区域。

必须支持：

- 选择 `手动配置`。
- 选择 `AI 生成方案`。
- 当选择 AI 生成方案时，从 AI Profile 列表中指定一个激活方案。
- 展示当前激活来源、方案名称、更新时间、置信度和风险提示。
- 提供回退到手动配置的显式入口。
- 清楚说明 AI 方案不会覆盖手动配置。

禁止行为：

- 禁止切换 AI Profile 时删除用户手动配置。
- 禁止切换 AI Profile 时重写 `channels`、`groups`、`group_items` 等原表。
- 禁止 AI Profile 无效时让运行时进入不可解释状态。

---

## 6. 动态路由 AI 学习专线

动态路由不纳入普通 AI Profile 覆盖体系。

动态路由保留两条线：

- 本地机制：现有 `shadow-ai`、`hybrid`、`metrics-only`、`strict-mechanism`、`incident-safe` 模式，以及 stats / route-target policy / circuit / probe 信号。
- AI 学习：本地在线学习状态，按 `(channel, key, model)` 粒度记录成功率、失败率、延迟、fallback、race winner、score 和 confidence。

设置页动态路由区域必须新增：

- `dynamic_routing_learning_enabled` 开关。
- 学习状态摘要。
- 学习数据查询入口。
- 清空学习状态入口。

硬规则：

- AI 学习只影响允许模式下的运行时推荐排序。
- AI 学习不能写回 `group_items`。
- AI 学习不能覆盖用户 priority。
- AI 学习不能永久改写渠道、分组或 key 配置。
- 关闭学习开关后，学习数据可以保留，但不能参与推荐排序。

---

## 7. 配置项与接口要求

新增设置项：

- `ai_automation_enabled`
- `ai_automation_base_url`
- `ai_automation_model`
- `ai_automation_use_local_default`
- `config_source_mode`
- `active_ai_profile_id`
- `dynamic_routing_learning_enabled`

新增管理 API：

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

Public contract closure notes:

- Canonical profile domain name is `config_health_check`; legacy `config_health` is only a compatibility alias for old records and documentation search.
- `AIProfile` detail must expose `domain_payload_type`, `domain_payload`, and `migration_status`; frontends consume typed domain payload first and only show `AIProfileVersion.content_json` as a legacy snapshot fallback.
- Task history MVP is `GET /api/v1/ai/tasks` with `page`, `page_size`, `status`, `type`, `profile_domain`, `keyword`, `created_from`, and `created_to` filters.
- `AITask` must persist `config_snapshot_json`, `context_payload_json`, `prompt_text`, `selected_model`, `model_reason`, `resume_state`, `last_heartbeat_at`, and `attempt_count` so recoverable work is not memory-only.
- Acceptance closure includes typed payload as the main consumption contract, safe checkpoint recovery, history task list UI/API, Vitest in CI, and `go test ./internal/update -count=1` in CI.
- Backup advanced migration remains partially complete; this AI automation closure must not mark it fully done.

---

## 8. 验收标准

本主线完成的最低条件：

1. 顶层 `AI 自动化` 栏目存在，并能作为统一任务入口。
2. AI 自动化配置支持本机默认和自定义 endpoint / model。
3. 模型列表可自动获取，并能标识默认推荐依据。
4. 自然语言任务、提示词模板、用户自定义提示词和进度条均可用。
5. AI 生成结果以 AI Profile 保存，不覆盖手动配置。
6. 设置页可在手动配置和 AI Profile 之间显式切换。
7. AI Profile 无效时自动回退手动配置。
8. 动态路由 AI 学习有独立开关和学习状态。
9. 动态路由 AI 学习只影响运行时推荐，不永久改写用户配置。
10. 文档、后端接口、前端入口、测试和 worklog 对该主线保持一致。

---

## 9. 明确非目标

第一阶段不做以下内容：

- 不开放 AI 静默覆盖手动配置。
- 不开放无 diff、无确认、无回滚的一键应用。
- 不把普通 AI Profile 作为动态路由的覆盖来源。
- 不让价格影响业务路由排序。
- 不依赖外部大模型服务完成动态路由学习；动态路由学习必须是本地可解释学习。
