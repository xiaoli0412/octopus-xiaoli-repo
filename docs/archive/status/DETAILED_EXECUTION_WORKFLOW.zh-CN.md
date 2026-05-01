# Octopus 详细执行计划与工作流

> 更新时间：2026-04-23
>
> 本文档是对 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md) 的执行化拆解。
>
> 用途不是替代 canonical plan，而是把 canonical plan 变成可直接施工、可直接验收、可直接回滚的工作流。

> 执行硬规则：后续任何实现、补丁、回归修复、UI 调整、路由策略修改、接口扩展与验收动作，都必须先对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)。
> 若当前文档与主规划冲突，以主规划为准；若实现需要偏离主规划，必须先更新主规划，再改代码，再补验证记录。

---
## 1. 总原则

### 1.0 跨自动化统一继承规则

以下规则必须被 octopus、octopus-2、octopus-repo 三条自动化链路同时继承，不能只在单条自动化中生效：

- 每一轮都必须先读取：
  - [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)
  - [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md)
  - [ENV_READY_AND_NEXT_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md)
  - 当前阶段 worklog 与最近一轮自动化 memory
- 开工前必须填写 Master plan aligned before coding (yes/no):，未对齐不得进入编码。
- 施工主体统一为 Codex；文档与记忆中禁止继续使用 OpenCode 作为当前执行口径。
- 当 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 与旧阶段计划冲突时，以用户上下文总账和当前优先级为准。
- 当前主线优先级固定为：
  - P0：工作流/计划/memory 统一与跨自动化继承稳定
  - P1：图片问题池（版本卡片、首页卡片、创建分组、创建渠道、Route Target Overrides、CC Switch、中文化与帮助提示、熔断设置、最新截图里的首页错位/多 Key/模型价格/备份英文/动态路由英文/性能卡顿问题）
  - P1.1：帮助信息与介绍信息减字收缩，作为关键整改点优先处理；默认只保留问号提示和一句主说明，正文不得再堆砌大段硬文本
  - P1.2：本地/内网兼容性收口，禁止把渠道模型拉取或 providers 预设链路误限为 `https-only`；绝对 `http` 与 `https` 都必须允许
  - P2：备份导入导出兼容与跨项目迁移
  - P3：其余主线任务
- 每轮收工必须同步更新 worklog 与对应 automation memory，保证下一轮可直接接续。
- 当前本机默认运行策略为“项目默认停驻，按需 self-start / external / healthcheck”，除非任务明确需要持续人工联调，不应让 `octopus_repo` 前后端长期常驻。
- Windows 本机运行态管理统一使用 `scripts/runtime-win.ps1`；任何精确停服动作都必须限制在 `D:\GPT-codex\octopus_repo` 相关进程，不得扩大到其他项目。

### 1.0.1 主阶段映射说明（2026-04-22）

- canonical 阶段映射仍保留 Phase A-G 作为长期框架。
- `Phase H` 新增为 AI 自动化中心与配置 Profile 双轨主线，覆盖 `AI 自动化中心 + AI Profile + 动态路由 AI 学习`。
- 当前执行阶段视作 Phase G 下的图片问题优先返工窗口，并并行维护 Phase F 的备份兼容收口。
- 若短期优先级与 Phase A-G 顺序存在冲突，按用户上下文优先级执行，但必须在 worklog 记录“为什么临时切换”。

### 1.1 唯一真相源

- 每次开始任何开发任务前，先阅读：
  - [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)
  - [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
  - 对应阶段的 worklog
  - 上一次任务留下的验收结果和未完成项
- 如果代码实现与 MD 冲突，以 MD 为准。
- 如果必须偏离 MD，先改 MD，再改代码，再补验证记录。
- 当前桌面环境统一按 `Codex` 口径执行；文档中所有流程必须能被当前 `Codex` 主线程、自动化链路和按目录切分的子 agent 直接落实。

### 1.1.1 配套执行资产

为了让本工作流真正可落地，仓库内固定配套以下文件：

- [worklog/README.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/worklog/README.zh-CN.md)
- [worklog/WORKLOG_TEMPLATE.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/worklog/WORKLOG_TEMPLATE.zh-CN.md)
- [worklog/2026-04-15-phase-a-stability-closure.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-15-phase-a-stability-closure.md)

规则：

- `README` 负责说明怎么用
- `TEMPLATE` 负责约束格式
- 当前活跃阶段必须有一份正式 worklog

### 1.1.2 本地资源优先与子 Agent 协作

这是执行工作流的硬规则，不是可选建议：

- 每次开工前，先盘点本轮可复用的本地资源，包括仓库内 MD、worklog、脚本、测试命令、当前线程上下文、本地 skills 和记忆上下文。
- 能从本地资源确认的事实，先以本地资源为准；只有本地资源不足时，才补充新的探索与试验。
- 发现任务可以拆成相互独立的子任务时，优先使用子 agent 并行推进，而不是把所有问题串行堆在主线程。
- 子 agent 启动前必须先定义分工，明确其职责边界、文件范围或问题范围，并约束交付物，避免互相覆盖。
- 若当前环境允许并可正常调度，子 agent 默认统一使用 `gpt-5.4`；若因权限、网关或环境原因无法使用该模型，必须先记录阻塞，再回退到本地资源优先的串行方案或经记录的替代模型。
- 主线程必须承担汇总责任，主动同步各子 agent 结论，统一口径，确保互通有无后再进入主线决策或落盘。
- 如果子 agent 因权限、模型或环境原因不可用，必须记录阻塞，并立即回退到本地资源优先的串行方案继续执行。
- 如果采用按目录切分的分目录 agent，必须先写清其负责目录、问题边界和禁止越界范围，确保它们在 `Codex` 当前环境下可直接执行且不会互相覆盖。
### 1.2 每次开工前固定动作

每次正式动手前，必须按下面顺序执行：

1. 盘点本轮可直接复用的本地资源，包括仓库内 MD、当前阶段 worklog、已有脚本、测试命令、当前线程上下文、本地 skills 和记忆上下文。
2. 读取 canonical MD 对应章节。
3. 读取 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 并核对当前总账编号、优先级和新增要求。
4. 读取当前阶段 worklog。
5. 在 worklog 中填写 `Master plan aligned before coding (yes/no):`。
6. 明确本次任务属于哪个 Phase、哪个 Milestone。
7. 抄出本次任务的硬规则、禁止事项、验收条件。
8. 判断本次任务是否存在可并行的独立子任务；如果存在，先拆分任务，明确每个子 agent 的职责边界、文件范围或问题范围，再进入执行。
9. 检查当前代码状态是否已经满足进入该任务的前置条件。
10. 只在前置条件满足后开始编码。

补充要求（跨自动化）：

- 在正式编码前，必须明确“当前主线、当前阶段、当前小闭环任务、完成判定标准、验证命令”。
- 每轮只选 1 个主任务，必要时最多搭配 1 个配套子任务。
- 若主任务阻塞，必须切换到同主线相邻任务，不得跨主题发散。
- 若本轮不需要真实运行态，先执行 `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`，确认项目保持停驻状态。
- 若本轮需要浏览器 smoke 或运行态验证，优先使用仓库现有 `self-start` 或 `external` wrapper；只有在包装器不适用时，才允许手工短时启动服务。

### 1.3 每次收工前固定动作

每次任务结束前，必须补齐：

1. 本次改动影响了哪些硬规则。
2. 本次是否引入了新的兼容性风险。
3. 本次是否影响旧数据、旧 UI、旧接口。
4. 本次是否有可回滚点。
5. 本次构建/测试结果。
6. 本次使用了哪些本地资源、skills、记忆上下文，以及它们分别提供了什么结论。
7. 本次是否使用了子 agent；若使用，分别负责了什么范围、产出了什么结论，主线程最终采纳了哪些内容。
8. 本次 worklog 更新。
9. 若本次涉及截图问题，必须在文档或 worklog 中补齐“现象、根因、风险、修复动作、自动化验收点”五段分析，不能只写“已修复”。

补充要求（跨自动化）：

- 必须在对应 automation memory 写入本轮结论、阻塞与下一轮候选顺序。
- 必须记录“本轮结果：成功 / 跳过 / 失败”与“是否需要人工介入”。
- 若本轮曾临时启动 `octopus_repo` 本地服务，收工前必须确认已执行精确停服，并记录停服方式与剩余端口状态。

### 1.4 截图问题执行模板

当前阶段所有截图问题必须按统一模板拆分，方便自动化与分目录 agent 执行：

1. 问题编号与模块名
2. 截图现象（用户可见层）
3. 根因假设（设计、状态、文案、交互、契约）
4. 用户风险（理解风险、误操作风险、性能风险）
5. 修复动作（UI、文案、逻辑、帮助提示、状态联动）
6. 自动化验收点（脚本断言、组件断言、逻辑断言、手工 smoke 点）

硬规则：

- 没有第 6 条自动化验收点的问题，不得标记为“已完成”。
- 涉及中文化的问题，验收点必须包含“无 i18n key 泄漏、无无必要英文主文案”。
- 涉及设置页复杂配置的问题，验收点必须包含“默认简洁、显式展开、帮助提示可见”。
- 涉及帮助提示、介绍文案或异常说明的问题，验收点必须包含“硬文本显著减字、默认只保留一句主说明、详细内容不再首屏铺开”。
- 涉及桌面端截图错位的问题，验收点必须包含“模块不出框、标题/数值不裁切、普通布局可读”。
- 涉及多 Key 的问题，验收点必须包含“同渠道内多 key 折叠展开成立、真实 key 输入位明确、每个 key 可看见独立模型范围”。
- 涉及模型/价格区域的问题，验收点必须包含“普通布局与紧凑布局两档可切换，且中文界面无无必要英文枚举”。

### 1.4.1 最新截图问题执行附加规则（2026-04-23）

针对当前这批最新截图，自动化和主线程后续执行时必须额外遵守以下规则：

1. 首页问题不能只修文案，必须同时修布局结构和占地比例。
2. 渠道问题不能只修多语言，必须同时修“多 key 同渠道管理”的交互逻辑。
3. 分组问题不能只修翻译 key，必须同时压缩默认视图、收起高级字段。
4. 模型/价格问题不能只修英文枚举，必须同时补普通/紧凑布局切换。
5. 备份与动态路由问题不能只修局部按钮，必须清理整段英文主文案。
6. 性能优化必须和这些 UI 返工同批推进，不能留到最后集中兜底。
7. 帮助提示与介绍信息必须同步减字收缩，不能一边收布局一边继续增加大段硬文本。

---

## 2. 全局执行顺序

整个项目按以下顺序推进，顺序固定，不允许随意跳阶段：

1. `Phase A`：工程稳定性收口
2. `Phase B`：里程碑 1 可用性核心收口
3. `Phase C`：里程碑 2 可观测性与排障收口
4. `Phase D`：里程碑 3 策略与性能收口
5. `Phase E`：里程碑 5 价格归一化与展示收口
6. `Phase F`：里程碑 6 备份/导入/迁移适配收口
7. `Phase G`：UI、移动端、部署与最终验收
8. `Phase H`：AI 自动化中心与配置 Profile 双轨主线

说明：

- 这里的 `Phase A-H` 是当前实际施工顺序。
- 它与 canonical plan 中的 `Phase 0-7` 是“映射关系”，不是替代关系。
- 当前仓库已经不处于“从零设计”状态，而处于“已有半成品，需要按优先级收口”的状态。

---

## 3. Phase A：工程稳定性收口

### 3.1 目标

- 把当前仓库从“能改”提升到“稳定可持续开发”。
- 避免后续所有阶段都被编译错误、脏状态、历史半成品拖累。

### 3.2 开工前必须先看

- canonical plan 第 12、14、16 节
- 当前环境文档：
  - [ENV_READY_AND_NEXT_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md)
- 当前状态文档：
  - [CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md)

### 3.3 本阶段要做什么

- 清理当前分支里会阻塞开发的编译错误、类型错误、结构错误。
- 固定最小冒烟验证清单。
- 明确哪些文件仍属于“历史半成品区”。
- 建立后续每个阶段都要重复执行的基础验证命令。

### 3.4 固定执行顺序

1. 跑后端构建与测试。
2. 跑前端构建。
3. 定位当前阻塞项。
4. 修复阻塞项。
5. 重新构建。
6. 记录阻塞项和修复结果。

### 3.5 本阶段产出

- 稳定可重复执行的基础命令集：
  - `go test ./...`
  - `go build ./...`
  - `pnpm build`
- 一份基础冒烟清单。
- 一份“当前高风险半成品清单”。

### 3.6 验收标准

- 后端构建通过。
- 前端构建通过。
- 核心页面无明显阻塞。
- 当前阶段 worklog 完整。

---

## 4. Phase B：里程碑 1 可用性核心收口

### 4.1 目标

- 完成多 key、key 模型绑定、group 路由、key 回退、多轮重试的核心闭环。
- 达到“系统能可靠路由，不因单点失败破坏无关模型”的标准。

### 4.2 开工前必须先看

- canonical plan：
  - 第 4 节目标架构
  - 第 5 节数据模型设计
  - 第 6 节路由与调度策略
  - 第 14 节验收标准
  - 第 15.1 节实现禁止事项
- 相关 worklog：
  - `2026-04-14-backend-task-01-channel-key-modes.md`
  - `2026-04-14-backend-task-02-advanced-failover-runtime.md`

### 4.3 本阶段拆分任务

#### 任务 B1：Channel / ChannelKey 模型彻底收口

- 确保 `key_management_mode`、`key_routing_policy`、`allowed_models`、`source_type` 的语义一致。
- 明确旧数据迁移默认值。
- 确保 `pooled` 与 `classified` 都可用且互不覆盖。

#### 任务 B2：严格顺序 / 严格权重 / 随机 策略收口

- 严格顺序必须“每次从第 1 个开始”。
- 严格权重必须是确定性权重序列，不允许 weighted-random。
- 随机必须完全随机，不带优先级、成本、历史成功率偏置。

#### 任务 B3：key 级回退与 group 多轮重试收口

- 失败 key 只能影响当前模型候选链，不能扩散。
- 同模型下 key fallback 必须按用户顺序。
- group 多轮重试必须只在响应未实际提交前生效。

#### 任务 B4：phase-1 failover 补到里程碑 1 标准

- 完善 `failover_window_sec`
- 完善 `race_after_fails`
- 完善 `race_concurrency`
- 让 attempts 记录与运行时行为一致

### 4.4 本阶段固定执行顺序

1. 先看 MD 硬规则和禁止事项。
2. 先修数据语义，再修运行时，再补前端配置，再补验证。
3. 每完成一个子任务，立刻补对应测试。
4. 不把多个策略混在一次提交里做。

### 4.5 本阶段验收标准

- 同一渠道可显式管理多个 key。
- 每个 key 可绑定模型列表。
- 同模型下严格按用户顺序尝试。
- 严格权重不是加权随机。
- 单个失败 key 不影响无关模型。
- 响应已写出后不再继续重试。

---

## 5. Phase C：里程碑 2 可观测性与排障收口

### 5.1 目标

- 让日志、统计、首页具备真正的排障能力。
- 让后续复杂策略调试不再依赖“猜”。

### 5.2 开工前必须先看

- canonical plan：
  - 第 5.3 节 Stats 实体
  - 第 5.4 节 Relay 日志实体
  - 第 9.4 节日志页
  - 第 9.5 节首页
  - 第 13 节里程碑 2

### 5.3 本阶段拆分任务

#### 任务 C1：日志导出收口

- 导出格式稳定为 `JSON` / `JSONL`
- attempts 链信息完整
- 导出字段适合 AI 排障
- request/response/error 字段边界清晰

#### 任务 C2：stats 维度统一

- `StatsChannel`
- `StatsModel`
- `StatsAPIKey`
- `StatsTotal`
- probe cost 与正常业务成本分开统计

#### 任务 C3：首页指标收口

- 已有：
  - total token
  - by channel
  - by model
- 待补：
  - by provider
  - official/gateway 价格扩展位
  - probe cost
  - breaker/probe 摘要

### 5.4 本阶段固定执行顺序

1. 先对齐后端返回结构。
2. 再对齐前端类型。
3. 再做首页与日志 UI。
4. 最后做导出验证和字段一致性核对。

### 5.5 本阶段验收标准

- 日志可直接导出给 AI 分析。
- attempts 链可完整还原一次故障转移路径。
- 首页指标与日志明细能相互印证。

---

## 6. Phase D：里程碑 3 策略与性能收口

### 6.1 目标

- 完成探测、熔断、半开恢复、动态调整、并发预算的真正闭环。
- 让 paid/free/public 的默认策略矩阵真正生效。

### 6.2 开工前必须先看

- canonical plan：
  - 第 6.4 节动态路由评分
  - 第 7 节检测与恢复策略
  - 第 10 节性能与并发方案
  - 第 13 节里程碑 3
  - 第 14 节验收标准第 5、6、13、14 条

### 6.3 本阶段拆分任务

#### 任务 D1：billing-aware probe policy 收口

- paid/metered 默认保守
- free/public 可更积极
- probe 成本独立统计

#### 任务 D2：熔断与半开恢复收口

- 失败阈值
- 半开恢复
- 成功恢复计数
- 恢复后重新纳入候选链

#### 任务 D3：动态调整收口

- 只允许调阈值
- 不允许改用户显式优先级
- 不允许价格驱动排序

#### 任务 D4：并发预算与取消机制收口

- 全局预算
- group 预算
- channel 预算
- key 预算
- probe 预算

### 6.4 本阶段验收标准

- 付费 route-target 默认保守探测。
- free/public route-target 可配置更积极探测。
- 动态调整不覆盖用户 priority。
- 并发竞速能及时取消多余请求。

---

## 7. Phase E：价格归一化与展示收口

### 7.1 目标

- 完成价格标准化、别名映射、官方价/网关价双视图。
- 明确“价格只展示，不驱动路由”。

### 7.2 开工前必须先看

- canonical plan：
  - 第 8 节价格归一化与展示
  - 第 13 节里程碑 5
  - 第 14 节验收标准第 4 条

### 7.3 本阶段拆分任务

#### 任务 E1：模型归一化命中链收口

- canonical name
- alias
- manual mapping
- fallback 命中顺序

#### 任务 E2：价格数据语义统一

- official 价格
- gateway 价格
- cache read/write
- billing_mode 对展示的影响

#### 任务 E3：模型页与统计页展示收口

- 双视图展示
- 搜索与筛选
- 命中结果可解释

### 7.4 本阶段验收标准

- 模糊匹配和别名映射稳定可预测。
- 官方价 / gateway 价可同时展示。
- 价格不影响路由优先级。

---

## 8. Phase F：备份 / 导入 / 迁移适配收口

### 8.1 目标

- 把当前“基础导出导入能力”补成真正可安全使用的迁移系统。

### 8.2 开工前必须先看

- canonical plan：
  - 第 10.4 节性能要求
  - 第 11 节迁移与回滚方案
  - 第 11.5 节项目级备份/导入/迁移适配方案
  - 第 14 节验收标准第 11、12 条

### 8.3 本阶段拆分任务

#### 任务 F1：导出安全语义收口

- 默认不泄露明文 secrets
- `manifest.contains_secrets` 与真实导出内容一致
- snapshot/version/checksum 完整

#### 任务 F2：dry-run 与差异分析收口

- schema 差异
- provider/model/key 缺失
- alias 冲突
- route 冲突

#### 任务 F3：导入冲突模式收口

- replace
- merge
- skip
- map

#### 任务 F4：导入后验证与回滚收口

- 健康检查
- 路由验证
- 价格规则验证
- 模型别名验证
- 一键回滚

### 8.4 本阶段验收标准

- 可生成版本化 snapshot。
- 默认不导出明文敏感信息。
- 导入支持 dry-run、差异预览、映射修正和回滚。
- 可完成一次完整的“备份 -> 导入 -> 校验 -> 回滚”演练。

---

## 9. Phase G：UI、移动端、部署与最终验收

### 9.1 目标

- 完成最终用户可见形态。
- 完成移动端可用性与部署验证。

### 9.2 开工前必须先看

- canonical plan：
  - 第 9 节 UI 方案
  - 第 13 节里程碑 4、5、6
  - 第 14 节全部验收标准

### 9.3 本阶段拆分任务

#### 任务 G1：渠道页最终形态

- key 子卡片
- pooled/classified 可视化
- 展开/折叠
- 搜索/筛选

#### 任务 G2：分组页最终形态

- `provider -> key -> model` 左栏分段
- 编辑器可理解、可操作

#### 任务 G3：模型页、首页、日志页最终形态

- 模型搜索/筛选
- 首页指标完整
- 日志导出与排障体验完整

#### 任务 G4：移动端与部署验证

- 375px 可用性检查
- Go 构建
- Web 构建
- Docker build
- compose up
- 流式接口回归

### 9.4 本阶段验收标准

- 手机端保持可用。
- OpenAI-compatible 行为保持完整。
- Docker 与 compose 验证通过。
- 第 14 节 14 条验收标准全部通过。

## 9.5 Phase H：AI 自动化中心与配置 Profile 双轨主线

### 9.5.1 目标

- 新增顶层 `AI 自动化` 栏目。
- 建立 AI 配置、自然语言任务、提示词模板、进度条和 AI Profile 保存体系。
- 在设置页提供 `manual / ai_profile` 双轨切换。
- 保证 AI 生成结果不覆盖用户手动配置。
- 把动态路由本地 AI 学习作为独立专线接入设置和 relay 推荐评分。

### 9.5.2 开工前必须先看

- [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md)
- [AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md)
- [DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md)
- [DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md)
- 本文档第 1.1.2 节的本地资源优先与子 agent 协作规则。

### 9.5.3 本阶段拆分任务

#### 任务 H1：文档与主线并入

- 新增中英文需求与实施计划。
- 更新 canonical plan、用户需求总账、当前状态、前端主线状态、环境 next plan、动态路由文档和 worklog。

#### 任务 H2：后端模型与设置

- 新增 AI config、AI task、AI task step、prompt template、AI Profile、AI Profile version。
- 新增 DynamicRouteLearningState。
- 新增 `ai_automation_*`、`config_source_mode`、`active_ai_profile_id`、`dynamic_routing_learning_enabled` 设置项。

#### 任务 H3：后端 API

- 新增 AI 配置、模型发现、提示词、任务、Profile 和动态学习 API。
- 激活 AI Profile 只能修改来源设置，不能覆盖手动配置表。

#### 任务 H4：前端 AI 自动化页面

- 新增顶层路由与导航。
- 新增模型配置、自然语言输入、提示词、进度条、结果预览和历史任务。

#### 任务 H5：设置页双轨切换

- 新增 `手动配置 / AI 生成方案` 切换。
- 展示当前 Profile、风险提示和回退入口。

#### 任务 H6：动态路由 AI 学习

- relay 完成后记录 `(channel, key, model)` 学习状态。
- scoring 读取学习状态并受 `dynamic_routing_learning_enabled` 控制。
- 学习不写回 `group_items`，不改 priority。

### 9.5.4 本阶段验收标准

- `AI 自动化` 顶层栏目存在。
- AI Profile 可保存、预览、激活和回退。
- `manual -> ai_profile -> manual` 切换后手动配置完整保留。
- AI Profile 无效时自动回退手动配置。
- 动态路由学习开关关闭时，学习数据不参与排序。
- 动态路由学习不永久改写用户配置。
- 文档、代码、测试和 worklog 同步更新。

---

## 10. 每次任务的标准模板

以后每次真正开工，都按下面模板执行：

### 10.1 开工前

- 本次任务名称：
- 对应 canonical 章节：
- 对应 milestone：
- 本次已盘点本地资源：
- 本次是否启用子 agent 与分工边界：
- 本次子 agent 使用模型：
- 本次硬规则：
- 本次禁止事项：
- 本次验收条件：
- 本次回滚点：

### 10.2 实施中

- 先改数据语义还是先改 UI：
- 受影响接口：
- 受影响页面：
- 是否影响旧数据：
- 是否影响旧行为：

### 10.3 收工前

- 构建是否通过：
- 测试是否通过：
- 本次使用了哪些本地资源 / skills / 记忆上下文：
- 本次使用了哪些子 agent 及其结论：
- 手工 smoke 状态 / 阻塞原因 / 缺少的环境：
- 是否补了 worklog：
- 是否还有遗留项：
- 是否满足进入下一任务的前置条件：

---

## 11. 当前建议的真实施工顺序

按当前仓库状态，后续最合理的真实执行顺序是：

1. 先做 `Phase A`
2. 然后收口 `Phase B`
3. 再做 `Phase C`
4. 再推进 `Phase D`
5. 再做 `Phase E`
6. 再做 `Phase F`
7. 最后做 `Phase G`
8. 进入 `Phase H` 的 AI 自动化中心与动态路由 AI 学习主线

原因：

- 当前最大的风险不是“功能没开始”，而是“很多能力只做到半套”。
- 所以必须先把稳定性和里程碑 1 收口，再继续更复杂的探测、价格和迁移系统。
- AI 自动化中心是新增主线，不替代 A-G；实际启动时必须先完成 H1 文档对齐，再按 H2-H9 逐步实施。

### 11.1 当前窗口优先顺序（覆盖所有自动化）

为对齐最新用户上下文，自动化链路当前执行顺序临时收敛为：

1. 先保证跨自动化 workflow/plan/memory 统一
2. 然后优先修图片问题池中的 UI/交互/文案与帮助提示
3. 再推进备份导入导出兼容收口
4. 最后回到其余 canonical 阶段任务

### 11.2 当前窗口的 UI/交互强制顺序（2026-04-23）

为方便自动化和分目录 agent 直接执行，当前这轮截图驱动返工的顺序固定为：

1. 首页统计卡片与渠道卡片布局收口
2. 渠道创建/编辑弹窗的多 Key 交互重做
3. 分组创建/编辑弹窗中文化与高级区收口
4. 模型/价格区域普通布局与紧凑布局双模式收口
5. 备份页与动态路由页英文残留清理
6. 渠道筛选增强（提供商 / 模型名 / key 信息）
7. 性能收口（默认折叠、布局收缩、减少首屏压力）

每一步都必须同时更新：

- 对应代码实现
- [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 的问题编号与状态
- 当前阶段 worklog
- 必要的自动化脚本或验证记录

说明：

- 这不是推翻 Phase A-G，而是当前窗口的优先执行层。
- 窗口任务关闭后，继续按 Phase A-G 主顺序推进。

### 11.3 当前窗口的前后端里程碑实现要求（2026-04-23）

为避免后续自动化只知道“Phase / Milestone 名称”，但不知道当前窗口到底做到哪一步，现把本轮统一实现要求补充如下：

1. 基础构建里程碑
   - 后端必须保持 `go build ./...` 通过。
   - 前端必须保持 `pnpm exec tsc --noEmit` 通过。
   - 静态构建必须保持 `pnpm run build:static` 通过，并同步到 `static/out`。

2. 前端主交互里程碑
   - 渠道：多 Key 折叠交互、中文主显示、搜索增强必须保持通过。
   - 分组：`modeBadge` 泄漏不得回归。
   - 模型/价格：普通 / 紧凑双布局必须保持真实生效，不允许回退成“只切入口不切卡片”。
   - 首页：必须保持“默认摘要 + 按需展开运行摘要”的层级结构，不允许回退成价格、熔断、探测、排行首屏全展开的大块堆叠布局。
   - 首页：首页主区域必须保持“请求主卡 + 三张摘要卡 + 右侧图表排行 + 下方活动热力图”的比例关系，不允许重新出现桌面端首屏错位和占地失衡。

3. 备份兼容里程碑
   - 备份逻辑层的 locale 输出必须保持可用。
   - 备份主流程必须继续通过 no-browser 验证链。
   - 后端 `internal/op` 与 `internal/server/handlers` 相关测试必须保持通过。

4. 自动化里程碑
   - 当前窗口至少保持以下脚本可直接调用：
     - `verify-channel-create-flow.mjs`
     - `verify-channel-presentation.mjs`
     - `verify-group-create-flow.mjs`
     - `verify-home-layout.mjs`
     - `verify-llm-price-boundary.mjs`
     - `pnpm --dir web run test:screenshot-no-browser`
     - `verify-ccswitch-flow.mjs`
   - 若某条浏览器级 smoke 因环境阻塞失败，必须记录阻塞并回退到 no-browser 对应链路，不得空缺记录。

### 11.4 当前窗口补充验收

#### 11.4.1 当前窗口渠道展示验收补充（2026-04-23）

由于当前工作流文档尾部曾存在旧字符污染，后续自动化在执行“渠道页展示与多语言一致性”时，统一按以下补充命令执行，不再依赖旧尾段拷贝：

1. `pnpm exec tsc --noEmit`
2. `pnpm --dir web run test:screenshot-no-browser`
3. `.\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/op/... ./internal/server/handlers/...`
4. `.\scripts\use-go-env.ps1; & $env:GOEXE build ./...`

#### 11.4.2 当前窗口运行态管理补充（2026-04-24）

为避免本机长期残留旧服务干扰当前窗口验证，Windows 本机统一追加以下运行态入口：

1. `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
2. `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
3. `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop`
4. 需要真实在线探活时，再执行 `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action healthcheck`

规则：

- 默认先看 `status`，确认本机没有残留的 `octopus_repo` 常驻进程。
- 真实运行态验证优先走现有 `self-start` / `external` smoke 脚本，避免手工长期挂服务。
- 一轮验证结束后，如果本轮没有明确留下人工联调任务，必须执行 `stop` 并在 worklog 记录结果。

#### 11.4.3 当前窗口统一验收命令

每次收口前，至少执行：

1. `pnpm exec tsc --noEmit`
2. `pnpm --dir web run test:screenshot-no-browser`
3. `.\\scripts\\use-go-env.ps1; & $env:GOEXE test ./internal/op/... ./internal/server/handlers/...`
4. `.\\scripts\\use-go-env.ps1; & $env:GOEXE build ./...`

### 11.5 Phase H 验收命令补充

Phase H 实施后，除当前窗口统一验收命令外，还必须补充：

1. 文档搜索：确认 `AI 自动化`、`AI Profile`、`dynamic_routing_learning_enabled`、`不覆盖用户配置` 出现在对应主线文档中。
2. 后端测试：`go test ./internal/model ./internal/op ./internal/server/handlers ./internal/relay/...`
3. 前端类型检查：`pnpm --dir web exec tsc --noEmit`
4. AI 自动化页面组件测试：后续实现时新增对应脚本或 vitest 测试。
5. 设置页双轨切换测试：必须覆盖 `manual -> ai_profile -> manual` 且原配置不丢失。
6. 动态路由学习测试：必须覆盖学习开关关闭后不参与排序，以及学习推荐不写回 `group_items`。

