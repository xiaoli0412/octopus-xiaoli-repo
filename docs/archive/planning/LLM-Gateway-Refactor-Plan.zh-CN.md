# LLM 网关重构计划（中文版）

> 仓库：`erguotou520/octopus`
>
> 分支：`feat/erguotou`
>
> 状态：当前实施阶段的唯一 canonical 计划文档（中文审阅版）。
>
> 规则：后续任何实现改动都必须与本文档保持一致；如果实现需要偏离，必须先更新本文档，再改代码。

---

## 1. 范围与指导原则

### 1.1 核心目标

1. **第一优先级：LLM 可用性**
   - 任何一个 key、渠道或模型失败，都不能拖垮不相关的模型流量。
   - 系统必须保住 OpenAI-compatible 行为和流式输出语义。

2. **第二优先级：性能**
   - 减少锁竞争、过度重试、不必要的探测流量和路由选择开销。
   - 在并发负载下保持吞吐能力。

3. **第三优先级：配置简单但能力丰富**
   - 用户侧配置保持简单直观。
   - 高级配置要存在，但必须是可选的、低摩擦的。

### 1.2 不可妥协约束

- 不允许把价格作为路由优先级信号。
- 不允许因为某个 provider 更便宜就自动降权。
- 路由必须尊重用户显式设置：`priority`、`sort_order`、`weight`、`enabled`、`health`、`model compatibility`。
- 所有检测、熔断、重试和并发策略都必须考虑成本。
- 必须保住 OpenAI-compatible 接口，尤其是 `/v1/chat/completions`、`/v1/models`、`/v1/responses` 和流式输出。
- 渠道模型拉取、`/v1/models` 兼容链路与 providers 预设不得误限为 `https-only`；用户自管上游的绝对 `http` 与 `https` 地址都必须允许，仅拒绝缺 host 或非 `http/https` 协议。
- 改动必须尽量低侵入、可回退、兼容旧数据。
- 当前施工环境统一按 `Codex` 口径执行；文档里的自动化、子 agent 和分目录 agent 协作描述，必须能在 `Codex` 当前环境下直接落实。
- 每次正式编码前，必须先完成主规划对齐，并在 worklog 中显式记录 `Master plan aligned before coding (yes/no):`。

---

## 2. 仓库现状扫描

### 2.1 已确认结构

- `cmd/` —— CLI 入口（`root.go`、`start.go`）
- `internal/conf/` —— 运行时配置与环境变量覆盖
- `internal/db/` —— GORM 初始化与迁移流程
- `internal/model/` —— DB 实体与请求/响应模型
- `internal/op/` —— 业务操作与缓存
- `internal/server/` —— Gin 服务、handlers、middleware、router
- `internal/relay/` —— 请求路由与上游转发核心
- `internal/relay/balancer/` —— 策略选择与候选遍历
- `internal/transformer/` —— 入站 / 出站协议适配器
- `web/src/` —— Next.js 前端源码
- `web/public/locale/` —— 国际化文案
- `docs/` —— 项目文档
- `static/` —— 嵌入式前端产物目标目录

### 2.2 现有基线能力

- 多渠道聚合
- 单渠道多 key 支持
- 请求路由、负载均衡、故障转移、模型兼容过滤
- 模型 / 价格同步
- 统计追踪
- Relay 日志 + SSE 流
- GitHub Copilot / Antigravity OAuth 流程
- CC Switch 深链集成

### 2.3 当前主要限制

- ChannelKey 还没有被清晰建模成“可独立管理的、按模型边界划分的 key 池”，对最终用户而言还不够简单。
- Group 路由已经存在，但新的要求需要更明确地控制严格顺序、严格权重、故障转移和多轮行为。
- 价格是一个被追踪的领域，但它不能影响路由优先级。
- 详细导出和 AI 友好的诊断能力还不完整。
- 首页的 token 使用需要更清晰地按 provider/channel 与 model 拆分。

---

## 3. 当前架构总结

### 3.1 后端流程

`main.go -> cmd/start.go -> conf -> db -> op -> server -> relay -> transformer -> upstream`

### 3.2 前端流程

`web/src/app -> app shell -> route config -> modules -> api endpoints -> backend`

### 3.3 两套认证模式

- 管理端 JWT 模式：`/api/v1/*`
- API Key 网关模式：`/v1/*`

### 3.4 核心领域对象

- `Channel`
- `ChannelKey`
- `Group`
- `GroupItem`
- `LLMInfo`
- `APIKey`
- `RelayLog`
- `StatsTotal / StatsDaily / StatsHourly`
- `StatsModel / StatsChannel / StatsAPIKey`

---

## 4. 目标架构

### 4.1 目标系统应具备的能力

- 将单个 channel 视为多个 key 的容器。
- 允许每个 key 维护一份简单、易编辑的允许模型列表。
- 允许同一 channel 下，不同 key 负责不同模型子集。
- 允许同一模型下的多个 key 独立轮询与故障转移。
- 让 group 路由保持可预测、可显式控制。
- 增加成本感知的探测和健康管理。
- 保持 UI 对手机端友好且简单直观。

### 4.1.1 同一渠道内多 key（硬规则）

这不是“后端底层支持多 key”就算完成，而是必须满足以下用户可见结构：

- **同一个渠道卡片 / 同一个渠道界面中显式管理多个 key**
- 不允许继续靠“重复创建多个同渠道”来表达多个 key
- 每个 key 都必须是独立可见、独立配置、独立观测的对象

每个 key 最终必须能独立展示并管理：

- 模型列表
- 状态
- 检测
- 熔断
- 额度 / 剩余额度
- 计费方式
- 探测策略
- 优先级 / 权重（如适用）

也就是说，目标 UI 不是“一个渠道里藏着多个 key 字符串”，而是“一个渠道里有多个 key 子卡片 / 子项”。

### 4.1.2 渠道内 Key 管理模式

同一供应商 / 同一渠道下，必须同时支持两种 key 管理模式，并且两种模式都保留：

- `classified`
- `pooled`

#### 模式 A：`classified`（key 分类模式）

适用于：

- 同一供应商下，不同 key 的模型权限不同
- 用户希望把不同模型集明确绑定到不同 key

行为要求：

- 先按模型过滤
- 再在对应 key 子集里做路由
- 每个 key 独立展示、独立检测、独立熔断、独立备注、独立策略

#### 模式 B：`pooled`（统一 key 池模式）

适用于：

- 同一供应商下，多个 key 的模型集合完全相同
- 用户不想做复杂 key 分类，而是希望多个 key 作为统一资源池处理

行为要求：

- 多个 key 共享同一模型集合
- 系统将这些 key 当作统一 key pool 处理
- 保留原来“一次性添加多个 key”的体验

#### 硬规则

1. 同一渠道必须允许用户显式切换模式
2. 不能靠系统自动猜模式
3. 旧逻辑迁移时必须兼容
4. 如果旧渠道原来就是统一多 key 轮询，应优先迁移为 `pooled`
5. 新建渠道时，用户必须可显式选择模式

#### 数据模型要求

必须新增并保留一个明确字段，例如：

- `key_management_mode`

允许值至少包括：

- `classified`
- `pooled`

#### 向后兼容要求

- 原来的 key 管理逻辑不能删除
- 原来的备注字段不能删除
- 原来“一次性添加多个 key”的 pooled 方式不能删除
- 这次不是新逻辑替代旧逻辑，而是：
  - **保留旧 pooled 逻辑**
  - **新增更强的 classified 模式**

### 4.2 绝对不能发生的事情

- 不允许仅因为价格便宜就自动改写路由。
- 不允许对付费 key 乱发探测流量。
- 不允许隐藏式重排覆盖用户显式优先级。
- 不允许一个 key 的失败影响无关模型。

---

## 5. 数据模型设计

### 5.1 ChannelKey

为每个 key 增加模型绑定字段：

- `allowed_models` —— 逗号分隔的模型名列表
- 为空表示“支持所有模型”，用于兼容旧数据

后续建议支持的字段：

- `source_type`
- `routing_priority`

### 5.1.3 Key 模式与 route-target 继承

在实现 `classified / pooled` 双模式时，必须明确：

- `key_management_mode` 是 channel 级字段
- key 保留自己的：
  - 认证
  - 限额
  - 来源类型
  - 备注
  - 基础健康状态
- route-target `(channel, key, model)` 继续保留：
  - `billing_mode`
  - `probe_policy`
  - `fallback_policy`
  - `circuit / recovery policy`
  - 成本感知行为

其中：

- `classified` 模式下，key-model 绑定是显式的、小范围的
- `pooled` 模式下，多个 key 共享同一模型集合，route-target 在运行时从 key pool 中分配

#### 5.1.1 source_type / billing_mode / probe_policy 约束

这里做一个重要更正：

- `billing_mode`
- `probe_policy`
- `probe_interval`
- `probe_concurrency_limit`

**不能简单绑定在 key 上**。

正确建模应当是：

- key 主要负责：
  - 认证
  - 来源信息
  - 限额
  - 可用性
  - 基础健康状态

- model / route-target 主要负责：
  - `billing_mode`
  - `probe_policy`
  - 探测频率
  - 并发探测策略
  - 故障转移策略
  - 成本感知行为

最终决策粒度必须是：

- `(channel, key, model)`

也就是 route-target 粒度，而不是仅按 key 粒度。

推荐优先级规则写死为：

- `route-target explicit override > model default > channel/key inheritance`

也就是说：

- key 可以提供默认值
- model 可以覆盖默认值
- 实际 route-target 可以最终覆写

这些约束不是未来建议，而是后续实施必须落地的策略基础：

- `source_type`
  - `public/free`
  - `paid/metered`
  - `private/internal`
  - `unknown`

- `billing_mode`（route-target / model 级）
  - `per_request`
  - `per_token`
  - `per_quota`
  - `flat`
  - `free`
  - `unknown`

- `probe_policy`（route-target / model 级）
  - `passive_only`
  - `sparse_single`
  - `sequential`
  - `concurrent`

这些字段不仅影响探测策略，也影响实际请求时的故障转移并发策略。

### 5.1.2 Route-target 视角的数据模型

为了满足计费差异和探测差异，计划中必须显式引入“route-target”概念：

- 一个 route-target = `(channel, key, model)`

这个 route-target 需要承载或派生以下策略信息：

- `billing_mode`
- `probe_policy`
- `probe_interval`
- `probe_concurrency_limit`
- `fallback_policy`
- `circuit_policy`
- `recovery_policy`
- `cost_behavior`

这样才能正确表达：

- 同一个 key 下，不同模型计费方式不同
- 同一个 key 下，不同模型故障转移策略不同
- 同一个渠道下，不同 key 和不同模型组合的策略不同

### 5.2 Group

增加 group 级别的重试配置：

- `retry_rounds`
- `retry_delay_ms`

后续可能扩展的路由元数据：

- 策略调优提示
- 健康预算
- group 级并发预算

### 5.3 Stats 实体

保留 / 扩展：

- `StatsChannel`
- `StatsModel`
- `StatsAPIKey`
- `StatsTotal`

Token breakdown 必须能按 channel 和 model 派生出来。

此外还必须支持以下独立统计维度：

- 正常业务成本
- probe 成本
- official price 统计
- gateway price 统计

其中 probe cost 必须独立统计，不允许混入正常业务成本。

### 5.4 Relay 日志实体

导出型、适合喂 AI 的日志必须包含：

- 请求模型
- 实际模型
- 渠道名
- key 尝试链
- 每次尝试的状态
- 延迟
- token 数
- 成本
- 请求 / 响应内容
- 错误信息

---

## 6. 路由与调度策略

### 6.1 Key 级路由

路由器必须：

1. 找到允许当前请求模型的 keys。
2. 排除禁用 / 冷却中 / 熔断中的 keys。
3. 只在该模型允许的 keys 集合里轮转。
4. 若一个 key 失败，切到同模型下一个可用 key。
5. 不能把失败扩散到无关模型。

#### 6.1.1 key 级回退（硬规则）

以下情况都必须触发 key 级回退：

- 认证失败
- key 已禁用
- key 过期
- key 额度耗尽
- key 对当前模型无权限
- key 熔断中
- 连续失败次数过多
- 最近失败率过高
- 网络不可达 / 超时 / 5xx

回退顺序必须按用户设置的优先顺序执行，不允许随机绕序。

#### 6.1.2 pooled key routing rules

`pooled` 模式下，多个 key 共享同一模型集合，系统把这些 key 当作统一资源池处理，而不是按 key-model 白名单切碎成多个子树。

必须支持的策略：

- 严格顺序轮询
- fill_priority / 填充优先
- 优先级顺序尝试
- 额度归零自动回退
- 同模型下多个 key 的 fallback

典型场景：

##### 场景 A：英伟达 / Ollama

- key1, key2, key3 ...
- 模型集合相同
- 用户选择 `pooled`
- 对外表现为同一模型集下多个 key 池化轮询

##### 场景 B：OpenAI-compatible 商家多个账号

- 多个账号都有 GPT5.x 模型集合
- 每个账号额度独立
- 一个账号额度归零后自动回退另一个
- 用户可以选择：
  - 轮询
  - fill_priority
  - 优先级顺序

### 6.2 Group 级策略模式

必须保留且清晰可预测的策略：

- **Round Robin** —— 每次从头开始的严格顺序遍历
- **Random** —— 候选项之间完全随机
- **Weighted** —— 确定性的加权序列，不是“加权随机”
- **Failover** —— 先主后备，按优先级优先

#### 6.2.1 轮询机制（硬规则）

这里的“轮询”**不是经典 round-robin**。

必须明确：

- 每次请求都从第 1 个开始
- 禁止全局游标
- 禁止记忆上次位置
- 禁止从上次成功位置继续

实际语义是：

- 严格顺序尝试
- 每次请求从列表头开始
- 列表顺序就是实际执行顺序

#### 6.2.2 随机机制（硬规则）

随机必须是完全随机：

- 所有候选等概率
- 不带权重偏置
- 不带优先级偏置
- 不带历史成功率偏置
- 不带成本偏置

#### 6.2.3 严格权重分配（硬规则）

严格权重分配**不是加权随机**。

必须实现为：

- 确定性的严格权重序列
- 周期结束后重新从第 1 个开始
- 不允许抽奖式随机

示例：

- A = 5, B = 3, C = 2
- 一个周期内严格表现为：`A A A A A B B B C C`
- 下一轮重新从 `A` 开始

用户看到的顺序必须能与实际路由行为对应。

#### 6.2.4 fill_priority / 填充优先

这是正式策略，不是口头概念。

语义定义：

- 优先把流量灌满当前主 key / 主目标
- 当前主 key 只有在以下情况才切到下一个：
  - 额度归零
  - 熔断
  - 不可用
  - 权限不符

这与“轮询”的区别必须明显：

- 轮询：按顺序平衡分配
- fill_priority：优先吃完当前主目标，再切下一个

适用场景：

- 多个账号模型相同，但希望先吃完一个账号额度
- 某个 key 永远作为主力，其它 key 作为补位 / 备用

要求：

- `fill_priority` 可用于 `pooled`
- 在 `classified` 模式下，如果“同模型命中多个 key”，也允许复用该策略

### 6.3 多轮重试

重试必须是有边界且安全的：

- 只有在响应还没有真正提交给客户端时才允许重试。
- 必须尊重请求上下文取消和 deadline。
- 重试轮数和延迟未来可以做成自适应，但先要可配置。
- 候选历史必须在日志 / attempts 中可观察。

#### 6.3.1 故障转移（硬规则）

故障转移不是简单 retry，也不是简单 first-wins。

必须满足：

- `360 秒总窗口` 是硬规则
- 连续两次失败后进入并发竞速阶段
- 并发竞速不是无上限并发风暴，必须受并发预算约束
- 不是简单“谁先返回谁赢”

最终输出裁决规则：

1. 收集并发成功结果
2. 按用户设置的模型 / 渠道优先级排序
3. 优先级高者优先返回
4. 若优先级相同，则先返回者优先
5. 一旦确定最终输出，立即取消其余请求

必须同时具备：

- 取消机制
- 并发预算
- 总超时机制
- 360 秒窗口结束即终止

### 6.4 Relay 运行时调优与每日动态摘要扫描

当前主线不再宣称存在一条“后台自动持久化动态调参”流水线。

真实口径收敛为两部分：

- relay 请求路径内允许做即时运行时调优，作用于 `(channel, key, model)` 维度
- 后台每天运行一次 `dynamic summary scan`，只生成摘要与观测结果，不直接持久化改写用户配置

即时运行时调优可参考的输入包括：

- 最近成功率
- 最近失败率
- 最近延迟
- 超时频率
- 半开恢复状态
- 当前并发预算与探测预算

硬规则：

- relay 运行时调优不能覆盖用户的显式优先级
- 不允许通过任何评分机制重排用户 `priority / sort_order / 显式顺序`

#### 6.4.1 Relay 运行时调优允许与禁止项

relay 运行时调优不能改写用户 priority / sort_order / 显式顺序。

relay 运行时调优**只能**影响当前请求或当前运行窗口内的阈值类行为，例如：

- retry 阈值
- probe interval
- circuit threshold
- recovery threshold
- 并发预算

明确禁止：

- 通过运行时评分悄悄重排用户显式设置顺序
- 因为价格便宜或贵就自动改写优先级
- 在后台任务中静默持久化覆盖用户设置

#### 6.4.2 Relay 运行时调优不作用于价格优先级

必须再次写死：

- 价格不能影响路由优先级
- 便宜 provider 不能因为便宜被自动降权
- 官方价 / gateway 价都不能参与排序决策

#### 6.4.3 每日动态摘要扫描运行周期

必须写死以下时间规则：

- `dynamic summary scan` 任务每天运行一次
- 该任务只生成摘要与观测结果，不直接持久化改写用户设置
- 更新失败时保留上一个成功摘要
- 摘要数据只允许用于帮助理解运行态趋势，以及辅助 relay 内即时阈值调优
- 摘要数据不得用于直接改写用户的 `priority / sort_order`

再次明确：

> 每日动态摘要扫描只负责观测与摘要；真实调优仅发生在 relay 请求路径内，且不允许通过任何评分机制改变用户显式排序。

---

## 7. 检测与恢复策略

### 7.1 Key / channel / model 检测

检测必须是成本感知、类型感知的：

- 付费 / 按量 key 应该更保守地探测
- 免费 / 公共 / 低成本 key 可更积极地探测
- 探测优先用轻量检查和缓存状态
- 避免高频探测风暴

这里也做重要更正：

- 探测策略不能只按 key 生效
- 探测策略必须按 model / route-target 生效
- 当同一个 key 下不同模型计费方式不同，策略必须按 `(channel, key, model)` 决定

#### 7.1.1 计费感知探测规则（硬规则）

必须写死以下默认策略：

- `billing_mode = per_request`（route-target 级）
  - 默认 `passive_only` 或极长周期探测
  - 严禁高频探测
  - 严禁一有请求就顺便探测
  - 尽量不要用真实推理请求探测
  - 尽量避免一次业务请求中多路并发命中多个付费 provider

- `billing_mode = per_token`（route-target 级）
  - 默认 `sparse_single`
  - 探测必须轻量、低频
  - 必须低 token
  - 必须低并发
  - 能用 `/v1/models` 就优先 `/v1/models`
  - 若必须调用推理接口，必须是最短、最少 token 的轻量请求

- `source_type = free/public`
  - 可更积极探测
  - 可更积极进行多轮并发竞速
  - 但仍必须受预算和取消机制约束

#### 7.1.2 paid / free 站点差异不仅影响探测，也影响实际请求故障转移

必须明确：

- `paid / metered` 站点：
  - 默认顺序尝试
  - 默认单轮尝试
  - 尽量避免一次业务请求中并发命中多个付费 provider

- `free / public / 公益` 站点：
  - 可配置允许多轮并发竞速
  - 可更积极进行故障转移并发

这个差异必须基于：

- `source_type`
- `billing_mode`
- `probe_policy`

而不是基于价格自动推断。

#### 7.1.3 探测成本独立统计

必须满足：

- 探测成本单独统计
- 探测成本不能混入真实业务成本
- 探测触发必须节流
- 不能在每个用户请求旁路顺手探测一次

并且：

- 首页和统计页最终必须能看见 probe cost
- probe cost 必须作为显式可观测指标暴露给用户

### 7.2 模型同步过滤与 allowlist 机制

必须明确禁止把项目做成 OpenRouter 式超级模型池。

#### 明确不要的行为

- 不要默认拉取几百、几千、甚至 2000+ 模型并全部暴露
- 不要 provider 一同步就把全部模型直接塞到本地
- 不要把 `/v1/models` 返回的大 catalog 全量渲染 / 全量路由
- 不要让某个 provider 的巨大模型列表直接拖垮服务器和 UI

#### 必须支持的同步机制

- allowlist
- filtered sync
- keyword filter
- 只同步用户选中的模型
- 分批拉取
- 惰性加载

#### 与 key 管理模式的关系

- `classified` 模式下应优先采用“小范围明确绑定”
- 不允许把整个 provider 的全量模型直接暴露给所有 key

示例：

- key1：DeepSeek V3.1 / V3.2
- key2：GPT5.2 / GPT5.3

目标是这种**精确、可控、小范围**的 key-model 绑定，而不是 provider 全量模型自动开放。

#### 7.1.4 route-target 默认探测与故障转移策略矩阵

| route-target 类型 | 默认探测策略 | 默认故障转移策略 | 默认并发竞速 |
|---|---|---|---|
| `per_request + paid/metered` | `passive_only` / 极长周期 | 顺序尝试 | 默认禁止 |
| `per_token + paid/metered` | `sparse_single` | 顺序尝试 | 默认禁止或极低预算 |
| `free/public` | 可更积极探测 | 可更积极 failover | 可配置允许 |
| `unknown` | 保守默认 | 顺序尝试 | 默认禁止 |

#### 7.1.5 paid / free route-target 默认并发策略

对 `paid / metered route-target`：

- 默认**禁止**多路并发竞速
- 默认采用顺序尝试
- 默认单轮尝试优先
- 只有用户显式开启时，才允许进入 hedged / racing
- 即便开启，也必须使用更严格的并发预算与取消机制

对 `free / public / 公益 route-target`：

- 可配置允许更积极的多轮并发竞速
- 可更积极地进行恢复验证和故障转移扩散
- 但仍必须受预算控制

### 7.2 熔断器

熔断行为必须支持：

- model 级隔离
- key 级隔离
- channel 级隔离
- 半开恢复
- 安静期后自动重测

### 7.3 恢复

恢复必须是渐进式的：

- 短期失败不能永久惩罚一条路由
- 成功探测应逐步恢复信任
- 恢复过程不能制造突发流量尖峰

---

## 8. 价格归一化与展示

### 8.1 价格规则

价格必须按模型身份归一化，不能只靠完全一致的字符串匹配。

需要支持：

- 完整归一化精确匹配
- 家族匹配
- 后缀 / 版本匹配
- 别名映射
- 手工映射

#### 8.1.1 价格系统必须补齐的细分步骤

必须明确写入施工范围：

- 标准清洗（normalize）
- 家族识别
- 版本提取
- tier / variant 提取
- alias / manual mapping
- fallback 命中顺序
- cache

推荐 fallback 命中顺序：

1. 完整归一化 key 精确匹配
2. 家族 + 版本匹配
3. 家族 + 类别匹配
4. 家族模糊匹配
5. 默认规则
6. 手工 alias / mapping 修正

#### 8.1.2 标准清洗细则

价格归一化必须显式包含以下处理步骤：

- 标准清洗（大小写、分隔符、空格、常见噪音字符）
- 家族识别（如 gpt / deepseek / glm / claude 等）
- 版本提取（如 4.5 / v3.2 / 0324 等）
- tier / variant 提取（如 pro / mini / turbo / flash / ultra）
- alias / manual mapping
- fallback 匹配
- cache 命中与缓存回填

### 8.2 价格展示

必须同时保留两种价格视图：

1. **官方 provider 价格** —— 用于参考和展示
2. **Gateway 价格** —— 用于内部记账 / 实际网关经济性

两者都必须展示，但两者都不能影响路由顺序。

### 8.3 价格不能驱动路由

这是硬规则。

- 价格只用于展示、记账和对比
- 价格不能成为路由优先级信号

---

## 9. UI 方案

### 9.1 渠道页

目标：

- 一个渠道卡片内可以包含多个 key
- 每个 key 能显示自己的允许模型列表
- 每个 key 都能单独搜索、测试、折叠、编辑
- 保持现有视觉风格和手机端可读性

#### 9.1.1 渠道页必须最终达到的 key 展示粒度

每个 key 必须独立显示：

- key 名称 / 掩码 key
- 模型列表
- 状态
- 检测按钮
- 熔断状态
- 额度 / 剩余额度
- 计费方式
- 探测策略
- 优先级 / 权重（如适用）

搜索能力最终必须支持：

- 渠道名
- key 名
- 模型名
- 模型家族名
- 状态
- 熔断状态
- 失败率
- 最近成功时间

### 9.2 分组页

目标：

- 保留当前 group 页面风格
- 清晰暴露策略选择
- 以简单表单暴露重试轮数 / 延迟
- 列表顺序要与真实执行顺序一致

并且必须最终支持：

- 轮询
- 随机
- 严格权重
- 故障转移

同时支持用户明确排序或拖拽，且“看到的顺序 = 实际执行顺序”。

#### 9.2.1 分组左栏 provider -> key 分段展示

分组页左半栏“添加模型”区域必须增强成分层展示：

##### `classified` 模式

```text
Provider/Channel
  ├─ K1
  │   ├─ model A
  │   ├─ model B
  ├─ K2
  │   ├─ model C
  │   ├─ model D
```

##### `pooled` 模式

```text
Provider/Channel
  ├─ 共享模型集
  │   ├─ model A
  │   ├─ model B
  ├─ Key Pool
  │   ├─ K1
  │   ├─ K2
  │   ├─ K3
```

必须满足：

- 用户一眼能看出当前 channel 是 `pooled` 还是 `classified`
- 不要把所有 key 和模型平铺成难以理解的大列表
- 允许折叠 / 展开
- 允许搜索 / 筛选
- 大量 key 时可考虑虚拟列表
- pooled / classified 必须有可视标识
- 优先考虑手机端可用性
- 不破坏原有分组操作逻辑

### 9.3 模型添加 / 同步 UI

目标：

- 保留原有的 search-and-add 流程
- 额外增加一个本地关键词筛选框
- 支持快速筛选像 “DeepSeek” 这种大列表关键词

### 9.4 日志页

目标：

- 保留现有日志流和列表行为
- 增加导出能力
- 保持手机端可用

### 9.5 首页

目标：

- 展示总 token 使用
- 展示按渠道 / provider 的使用拆分
- 展示按模型的使用拆分
- 手机端布局要紧凑且清晰

#### 9.5.1 首页最终必须展示的字段清单

首页最终必须展示：

1. 总 token 使用
2. 按 channel 的 token 使用
3. 按 provider 的 token 使用
4. 按 model 的 token 使用
5. 官方价格统计
6. gateway 价格统计
7. probe cost
8. 成功率 / 失败率
9. 熔断状态摘要
10. 最近探测状态摘要

并且必须明确：

- 首页展示的是“统计视图”，不是路由排序依据
- 官方价和 gateway 价都只用于参考和记账，不用于路由

### 9.6 设置 / CC Switch / 正则辅助

目标：

- 扩展 CC Switch 支持更多 CLI 工具
- 提升正则精度和匹配可信度
- 保持高级选项“可发现但不吵”

#### 9.6.1 当前用户上下文优先补充要求

本阶段必须并入以下用户优先要求，并以其作为设置页和辅助配置 UI 的当前返工基线：

- 探测与检测能力必须从价格区完全剥离，统一归到设置页。
- `CC Switch` 必须按流程型交互重做，优先采用“先选工具 -> 再选 API Key -> 再选主模型 -> 再展开高级映射”的同页渐进展开方式。
- 设置页中的复杂配置项必须配套“圈内问号 + 悬停帮助提示”，且帮助内容基于当前真实实现重写。
- 帮助提示与正文介绍都必须减字收缩，默认只保留一句主说明；同一信息若已进入 `HelpHint`，正文不得再重复铺开成长段硬文本。
- 熔断设置必须提供默认推荐值与高级自定义展开层，不能只暴露一个简单总开关。
- 中文界面必须以全中文为主，不能把英文枚举值和内部状态词直接裸露给用户。

### 9.7 项目级备份 / 导入 / 迁移适配 UI

必须新增项目级备份与导入能力，而不是仅导出少量配置。

这里的“项目”主要指：

- 网关配置
- 渠道 / key / 模型绑定
- 分组 / 策略
- 价格规则
- 别名规则
- relay 运行时调优与动态摘要扫描配置
- UI 配置
- 统计配置（默认可选）
- 迁移映射规则

推荐交互能力：

- 备份导出入口
- dry-run 导入预检
- 差异分析报告
- 冲突处理向导
- 导入映射表编辑
- 一键回滚到上一个 snapshot
- 部分恢复（按 channel / group / price rule）

### 9.8 AI 自动化中心与 AI Profile 双轨配置

必须新增顶层 `AI 自动化` 栏目，与首页、渠道、分组、模型/价格、日志、设置并列。

该栏目是所有 AI 辅助运维任务的统一入口，包括：

- 智能分组
- 渠道识别
- 价格识别
- 模型归类
- 配置健康检查
- 动态路由说明整理
- 后续更多 AI 自动化任务

AI 自动化中心必须支持：

- 自定义 AI `base_url`、API Key、渠道类型和模型。
- 未自定义时默认使用本机 Octopus OpenAI-compatible 地址，优先从当前服务地址推导，兜底为 `http://127.0.0.1:8080/v1`。
- 自动获取模型列表。
- 默认推荐免费、近期成功率高、延迟较低的模型。
- 自然语言任务输入。
- 内置提示词模板。
- 用户自定义提示词和手动工作要求。
- 任务进度条，至少覆盖收集上下文、选择模型、调用 AI、解析输出、生成方案和保存结果。
- AI 生成方案的预览、保存、历史回看和版本化。

AI 生成的分组、渠道识别、价格识别和模型归类结果必须保存为独立 `AI Profile`，不能直接覆盖用户手动配置。

硬规则：

- 用户手动配置与 AI 生成 Profile 必须同时保留。
- 设置页必须提供 `手动配置 / AI 生成方案` 的手动切换。
- 切换 AI Profile 只改变读取来源，不删除、不覆盖、不重排原有用户配置。
- AI Profile 缺失、损坏、未覆盖目标模型或未通过校验时，必须自动回退到手动配置。
- `channels`、`groups`、`group_items`、`llm_infos`、`route_target_overrides` 不能被 AI 任务静默覆盖。
- “免费优先、成功率高、延迟低”的默认模型选择只用于 AI 自动化任务执行模型，不得影响业务请求路由排序。

动态路由是例外主线：

- 动态路由不纳入普通 AI Profile 覆盖体系。
- 动态路由保留本地机制与 AI 学习两条线。
- `dynamic_routing_learning_enabled` 控制本地 AI 学习是否参与推荐。
- AI 学习只影响允许模式下的当前请求推荐排序。
- AI 学习不得写回 `group_items`，不得覆盖用户 priority，不得永久改写渠道、分组或 key 配置。

配套需求与实施计划见：

- [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md)
- [AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md)

---

## 10. 性能与并发方案

### 10.1 路由热路径

减少：

- 每个请求都做不必要排序
- 全局随机锁竞争
- 候选列表重复扫描（能缓存就缓存）

### 10.2 网络

确保：

- 连接池能被有效复用
- 流式响应保持透明
- 高并发流量不因共享锁而阻塞

### 10.2.1 并发预算（硬规则）

最终必须支持以下预算层级：

- 全局并发预算
- group 并发预算
- channel 并发预算
- key 并发预算
- probe 并发预算

并发竞速一旦确定最终输出，其他请求必须尽快取消，避免白白消耗额度和连接。

### 10.3 Stats 与日志

避免 stats 和日志成为热路径瓶颈：

- 必要时批量写入
- 导出与正常流量隔离
- 避免无谓的内存增长

---

## 10.4 备份 / 导入 / 迁移适配性能要求

- 备份包生成不能阻塞主请求热路径
- 导入 dry-run 必须优先做静态校验，再做有成本的健康检查
- 导入后验证必须分批、错峰进行，避免一次性压垮上游
- 大型导出 / 导入必须支持分阶段 / 分片处理

---

## 11. 迁移与回滚方案

### 11.1 数据库迁移

先为以下字段添加迁移：

- `channel_keys.allowed_models`
- `groups.retry_rounds`
- `groups.retry_delay_ms`

后续可能增加：

- source type
- billing mode
- probe policy 字段
- routing priority 字段

### 11.2 向后兼容

- 没有 `allowed_models` 的旧 channel 必须仍能工作
- 没有 retry 配置的旧 group 必须安全兜底
- 尽量保持旧路由兼容

### 11.3 回滚

回滚应该可以通过以下方式完成：

- 回退应用代码
- 保留旧列不动
- 默认值语义保持兼容

### 11.4 Feature flag / fallback 策略

为了降低重构风险，实施时必须按阶段考虑 feature flag / fallback：

- `ENABLE_KEY_MODEL_BINDING`
- `ENABLE_GROUP_MULTI_ROUND_RETRY`
- `ENABLE_DYNAMIC_ROUTING_HEALTH`
- `ENABLE_LOG_EXPORT`
- `ENABLE_TOKEN_BREAKDOWN`
- `ENABLE_PRICE_NORMALIZATION_V2`
- `ENABLE_PROJECT_BACKUP_EXPORT`
- `ENABLE_PROJECT_IMPORT_DRYRUN`
- `ENABLE_PROJECT_IMPORT_APPLY`

每个阶段必须至少有一个可回退点：

- 数据迁移可保留旧语义
- 新路由逻辑可切回旧逻辑
- 新 UI 可在高级选项中灰度展示

### 11.5 项目级备份 / 导入 / 迁移适配方案

#### 11.5.1 备份功能要求

必须提供版本化项目快照导出能力，建议格式：

- `zip`
- `json + manifest`
- `yaml + manifest`

备份内容至少包括：

- channels
- keys
- model bindings
- group 路由
- routing order / priority / weight
- price rules
- alias / normalization rules
- dynamic tuning settings
- feature flags
- UI 配置
- stats 配置（默认可选）
- 导入映射规则

备份元数据必须包括：

- 版本号
- schema 版本
- 导出时间
- checksum
- 导出来源环境标识
- 可选的加密导出模式

安全要求：

- 默认导出全量迁移快照，包含跨环境恢复所需的明文凭据
- 如需脱敏导出，必须显式关闭 secrets，并在 manifest 中如实标记 `contains_secrets=false`

#### 11.5.2 导入功能要求

导入必须支持：

1. dry-run 预检
2. 差异分析
3. 冲突处理模式：replace / merge / skip / map
4. 自动适配
5. 导入后验证
6. 回滚

差异分析至少覆盖：

- 缺失 provider
- 缺失模型
- alias 冲突
- route 冲突
- schema 版本不匹配
- key 不可用
- base_url 不一致

自动适配至少覆盖：

- 模型别名重映射
- provider / base_url 差异适配
- 旧版本 schema migration
- 缺失 key 的占位重绑定
- 导入后自动修正排序和引用关系

导入后验证必须包括：

- 自动健康检查
- 自动路由验证
- 自动价格规则验证
- 自动模型别名验证

#### 11.5.3 迁移适配目标

目标不是“导进去就算了”，而是：

> 导入后系统能尽量自动适配到新环境，最大化保持原有行为一致。

因此必须考虑：

- 同模型不同名称
- 同 provider 不同 base_url
- 某些 key 在新环境失效
- 某些模型在新环境不存在
- 某些路由规则需要映射到新名称

推荐增强项：

- 导入映射表
- 迁移预览页
- 冲突解决向导
- 兼容性报告
- 一键回滚到上一个 snapshot
- 选择性导入 / 部分恢复

#### 11.5.4 导入后模拟路由验证 / 迁移差异预览

为了保证“能导入”不等于“行为偏差”，导入流程必须支持模拟路由验证与差异预览：

- 对常用 group / model 做 dry-run 路由模拟
- 输出导入前后候选链差异
- 标出模型别名映射变化
- 标出 fallback 链变化
- 标出导入后失效目标
- 标出哪些 route-target 会因缺失 provider / key / model 被跳过
- 导入应用前允许用户确认或修正映射

目标是：

> 导入之后尽量保持原有行为一致，而不是只把配置“导进去”。

必须继续保留并增强：

- snapshot
- schema version
- checksum
- dry-run
- replace / merge / skip / map
- 导入后健康检查
- 导入后价格规则验证
- 导入后模型别名验证
- 回滚到上一个 snapshot

并且新增要求：

- 对常用 group / model 做 dry-run 路由模拟
- 输出导入前后候选链差异
- 标出 alias 映射变化
- 标出 fallback 链变化
- 标出失效 route-target
- 标出缺失 provider / key / model
- 在真正 apply 前允许用户确认或修正映射

---

## 12. 工作流计划（可施工版）

本节是当前 `Codex` 桌面环境下的施工用 workflow plan，优先级高于纯设计性描述。

### Phase 0：扫描仓库现状

目标：确认现有目录结构、后端 API、前端页面、数据库模型、relay/balancer/middleware/pricing/stats/log 代码落点。

主要文件范围：

- `internal/**`
- `web/src/**`
- `docs/**`
- `README.md`

产出：

- 现状扫描结论
- canonical plan 更新

验证：

- 计划文档与仓库现状一致

回滚：

- 文档级回滚即可

### Phase 1：更新 canonical plan

目标：把全部口述要求、长 MD 要求、工程增强项统一写进 canonical plan。

产出：

- `docs/LLM-Gateway-Refactor-Plan.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `本轮文档修订摘要`
  - 新增了哪些硬规则
  - 修正了哪些术语
  - 哪些地方是为了防止后续实现偏差

验证：

- 需求来源映射完整
- 硬规则全部明确写死

### Phase 1.5：修订后摘要输出

在进入大规模实施前，必须先向审阅方输出本轮计划修订摘要，至少包括：

1. 哪些地方从 key 级修正成了 route-target 级
2. paid / free 默认并发策略如何定义
3. relay 运行时调优与每日动态摘要扫描时间规则如何定义
4. 首页最终展示字段有哪些
5. 备份 / 导入 / 迁移适配新增了什么
6. 新增了哪些禁止事项

### Phase 2：先做数据模型 / migration

目标：先把 schema 和兼容层打牢，再改调度内核。

主要文件：

- `internal/model/*.go`
- `internal/db/migrate/*.go`

产出：

- 多 key / route-target / 备份导入所需字段

验证：

- 旧数据兼容性检查
- migration 可执行性检查

回滚：

- 保留旧列语义
- feature flag 关闭新路径

### Phase 3：路由与调度内核

目标：实现严格顺序、严格权重、随机、故障转移、key 回退、多轮重试。

主要文件：

- `internal/relay/relay.go`
- `internal/relay/balancer/*.go`
- `internal/op/channel.go`
- `internal/op/group.go`

验证：

- 渠道多 key 测试
- key 模型白名单测试
- key 回退测试
- 严格顺序轮询测试
- 完全随机测试
- 严格权重顺序测试
- 故障转移测试

回滚：

- 切回旧路由逻辑 flag

### Phase 4：探测 / 熔断 / relay 运行时调优

目标：实现 billing-aware probe policy、半开恢复、relay 运行时阈值调优，以及每日动态摘要扫描。

主要文件：

- `internal/relay/balancer/*`
- `internal/op/stats.go`
- `internal/server/handlers/channel.go`
- `internal/server/handlers/group.go`

验证：

- paid key 探测节流测试
- free/public key 并发探测测试
- 熔断测试
- 半开恢复测试
- 5 次成功恢复测试
- relay 运行时调优不覆盖 priority 测试

回滚：

- 关闭 relay 运行时调优 / 动态摘要扫描相关 feature flag

### Phase 5：价格归一化

目标：实现模糊匹配、官方价展示、gateway 价格展示、别名管理。

主要文件：

- `internal/op/llm.go`
- `internal/model/llm.go`
- `internal/server/handlers/model.go`
- `web/src/components/modules/model/*`

验证：

- 模糊匹配测试
- 官方价 / gateway 价双展示测试
- 别名映射测试

回滚：

- 切回旧 price matching 逻辑

### Phase 6：备份 / 导入 / 迁移适配

目标：实现项目级快照导出、dry-run 导入、冲突处理、导入后验证与回滚。

主要文件：

- `internal/server/handlers/backup.go`（或新增导入/导出 handlers）
- `internal/op/backup.go`
- `internal/model/*`
- `web/src/components/modules/setting/*`（或新增 backup/import 页面）

验证：

- dry-run 预检测试
- 差异分析测试
- merge / replace / skip / map 测试
- 导入后健康检查测试
- 一键回滚测试

回滚：

- 恢复上一个 snapshot

### Phase 7：UI / 首页 / 统计 / 导出 / 测试

目标：完成渠道页、分组页、价格页、首页、日志导出、模型筛选等用户可见增强。

主要文件：

- `web/src/components/modules/**`
- `web/src/api/endpoints/**`
- `web/public/locale/**`

验证：

- 手机端可用性检查
- 日志导出测试
- 首页 token breakdown 测试
- 价格页搜索筛选测试
- 帮助提示完整性和中文化检查
- `CC Switch` 渐进展开流程检查
- 探测归位到设置页后的信息架构检查

回滚：

- 新 UI 灰度 flag 关闭

### Phase H：AI 自动化中心与配置 Profile 双轨主线

目标：实现 `AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习`，让用户用自然语言生成分组、渠道识别、价格识别和模型归类建议，同时保证 AI 结果不覆盖手动配置。

主要文件：

- `internal/model/ai_automation.go`
- `internal/op/ai_automation.go`
- `internal/server/handlers/ai_automation.go`
- `internal/model/dynamic_route_learning.go`
- `internal/op/dynamic_route_learning.go`
- `internal/relay/dynamic_learning.go`
- `web/src/components/modules/ai-automation/*`
- `web/src/api/endpoints/ai-automation.ts`
- `web/src/components/modules/setting/*`
- `web/src/route/config.tsx`

阶段拆分：

1. H1：文档与主线并入。
2. H2：后端 AI 配置、任务、提示词、Profile、动态学习状态模型。
3. H3：AI 配置、模型发现、任务、Profile 和动态学习 API。
4. H4：新增顶层 `AI 自动化` 页面。
5. H5：设置页新增 `manual / ai_profile` 双轨切换。
6. H6：动态路由本地 AI 学习接入 relay 推荐评分。
7. H7：AI 生成分组与渠道识别 MVP。
8. H8：AI 价格识别和智能分组扩展。
9. H9：diff、审计、选择性应用和回滚增强。

验证：

- AI Profile 激活不覆盖用户手动配置。
- `manual -> ai_profile -> manual` 后原配置完整保留。
- AI Profile 无效时自动回退手动配置。
- `dynamic_routing_learning_enabled=false` 时学习数据不参与排序。
- 动态路由 AI 学习不写回 `group_items`，不改 priority。

回滚：

- 切回 `config_source_mode=manual`。
- 关闭 `dynamic_routing_learning_enabled`。
- 保留 AI Profile 历史数据但不参与运行时读取。

---

## 12. 风险清单

### 高风险

- 改路由逻辑时，悄悄破坏模型兼容性
- 探测策略对付费 key 造成额外成本
- relay 运行时调优覆盖用户显式路由偏好

### 中风险

- 手机端布局变化
- 日志导出 payload 过大
- cache 和 DB 临时不同步时的统计一致性

### 低风险

- 翻译文案新增
- 搜索 / 筛选输入框
- 非侵入式展示增强

---

## 13. 里程碑

### 里程碑 1 —— 可用性核心

- key 级模型绑定
- 同模型 key 轮询
- 安全 fallback
- 多轮重试

涉及文件（第一阶段主改动范围）：

- `internal/model/channel.go`
- `internal/op/channel.go`
- `internal/model/group.go`
- `internal/op/group.go`
- `internal/relay/relay.go`
- `internal/relay/balancer/balancer.go`
- `internal/relay/balancer/iterator.go`
- `internal/db/migrate/*.go`
- `web/src/api/endpoints/channel.ts`
- `web/src/api/endpoints/group.ts`
- `web/src/components/modules/channel/*`
- `web/src/components/modules/group/*`

验证方法：

- key 级模型白名单路由测试
- 同模型多 key 轮询测试
- 一个 key 失败不影响其他模型测试
- 写出响应后不再重试测试

阶段验收点：

- 同一渠道内可显式管理多个 key
- 每个 key 可绑定模型列表
- 每次请求从第 1 个开始严格顺序尝试
- 严格权重不是加权随机

### 里程碑 2 —— 观测与排障

- 详细日志导出
- AI 友好的 attempts 导出
- 更丰富的 stats 与 token breakdown

涉及文件：

- `internal/model/log.go`
- `internal/op/log.go`
- `internal/server/handlers/log.go`
- `internal/op/stats.go`
- `internal/server/handlers/stats.go`
- `web/src/api/endpoints/log.ts`
- `web/src/api/endpoints/stats.ts`
- `web/src/components/modules/log/*`
- `web/src/components/modules/home/*`

验证方法：

- JSON / JSONL 导出测试
- attempts 链完整性测试
- 首页 channel/model token breakdown 展示测试

阶段验收点：

- 可直接导出日志给 AI 排障
- 首页可看到官方 / gateway 统计扩展位（后续接价格）

### 里程碑 3 —— 策略与性能

- relay 运行时调优
- 每日动态摘要扫描
- 成本感知探测
- 减少锁竞争

涉及文件：

- `internal/relay/relay.go`
- `internal/relay/balancer/*`
- `internal/op/stats.go`
- `internal/model/channel.go`
- `internal/model/group.go`
- `internal/server/handlers/channel.go`
- `internal/server/handlers/group.go`

验证方法：

- paid / free / public 策略差异测试
- relay 运行时调优不覆盖用户 priority 测试
- 并发预算与取消机制测试

阶段验收点：

- relay 运行时调优只调阈值，不改用户顺序
- 每日动态摘要扫描只产出摘要，不持久化改写用户配置
- 付费 provider 不会被高频并发探测拖死

### 里程碑 4 —— UI 与易用性

- 渠道卡片增强
- 分组编辑器增强
- 模型搜索 / 筛选
- 手机端优化

验证方法：

- 375px 宽度移动端可用性检查
- 渠道卡片多 key 展开/折叠检查
- 搜索 / 筛选有效性检查

### 里程碑 5 —— 价格与 CC Switch 扩展

- 标准化价格展示
- 官方价 / 网关价双视图
- 支持更多 CLI 的 CC Switch

涉及文件：

- `internal/op/llm.go`
- `internal/model/llm.go`
- `internal/server/handlers/model.go`
- `web/src/components/modules/model/*`
- `web/src/components/modules/setting/*`
- `docs/CC_SWITCH.md`（如存在）

验证方法：

- 模型名模糊归一化测试
- 官方价 / gateway 价双展示测试
- CLI deep link 生成测试

### 里程碑 6 —— 验证与部署

- 测试
- Docker 验证
- 回滚验证

验证方法：

- go test / 前端 build / Docker build / compose up
- 流式接口回归测试
- 低配机器资源观察

### 里程碑 7 —— AI 自动化中心与动态路由 AI 学习

- 顶层 `AI 自动化` 栏目
- AI endpoint / model 配置
- 自动模型发现
- 自然语言任务输入
- 内置提示词与用户自定义提示词
- 任务进度条
- AI Profile 保存、预览、激活和回退
- 设置页 `manual / ai_profile` 双轨切换
- 动态路由 `dynamic_routing_learning_enabled` 开关
- `(channel, key, model)` 粒度本地学习状态

涉及文件：

- `internal/model/ai_automation.go`
- `internal/op/ai_automation.go`
- `internal/server/handlers/ai_automation.go`
- `internal/model/dynamic_route_learning.go`
- `internal/op/dynamic_route_learning.go`
- `internal/relay/dynamic_learning.go`
- `web/src/components/modules/ai-automation/*`
- `web/src/api/endpoints/ai-automation.ts`
- `web/src/components/modules/setting/*`

验证方法：

- AI Profile 不覆盖用户手动配置测试。
- 配置来源切换回退测试。
- AI Profile 无效 fallback 测试。
- 动态路由学习开关测试。
- 动态路由学习不写回 priority / group items 测试。

阶段验收点：

- 用户能在独立栏目中通过自然语言生成 AI 建议。
- AI 建议以 Profile 形式保存并可切换。
- 手动配置与 AI Profile 同时保留。
- 动态路由 AI 学习只做运行时推荐，不永久改写用户配置。

---

## 14. 验收标准

只有满足以下条件，重构才算通过：

1. 单个失败 key 不会破坏无关模型。
2. 一个 channel 可以在一个 UI 卡片内管理多个 key。
3. 路由保持显式、可控。
4. 价格不会影响路由优先级。
5. 付费 / 按次 / 按量的 route-target 探测必须保守。
6. free / public route-target 可以在预算内更积极地探测。
7. 日志可以导出成适合 AI 分析的格式。
8. 首页可以按 channel 和 model 展示 token 使用。
9. 手机端保持可用。
10. OpenAI-compatible 行为保持完整。
11. 项目级备份导出可生成版本化快照，默认支持直接恢复的新环境全量快照；若显式关闭 secrets，则必须在 manifest 中准确标记为脱敏快照。
12. 导入支持 dry-run、差异预览、映射修正和回滚。
13. paid / metered route-target 默认顺序尝试，默认禁止多路并发竞速。
14. free / public route-target 可配置更积极的多轮并发竞速。
15. 设置页中的探测、熔断、`CC Switch`、备份和多 key 复杂配置都具备与真实实现一致的全中文帮助提示。
16. 所有当前施工流程、自动化入口和子 agent 协作描述都能在 `Codex` 当前环境下直接执行，不依赖错误的 `OpenCode` 口径。
17. 顶层 `AI 自动化` 栏目可作为所有 AI 自动化任务的统一入口。
18. AI 自动化支持本机默认 endpoint、自定义 endpoint、自动模型发现、自然语言任务、提示词模板和进度条。
19. AI 生成分组、渠道识别、价格识别和模型归类结果必须保存为 AI Profile，并且不能覆盖用户手动配置。
20. 设置页必须支持 `manual / ai_profile` 双轨切换，且切换后用户原配置完整保留。
21. 动态路由 AI 学习必须由 `dynamic_routing_learning_enabled` 控制，只影响运行时推荐，不写回 `group_items`，不覆盖用户 priority。
22. providers 预设、渠道模型拉取与模型全部列表链路必须接受绝对 `http` 或 `https` `base_url`，不得强制 `https-only`。

---

## 15. 需求来源映射

| 来源 | 类型 | 说明 |
|---|---|---|
| 之前口述要求 | 用户需求 | key/model/channel 路由、日志导出、模型筛选、token breakdown、CC Switch、正则、手机端、Docker |
| 后续口述要求 | 用户需求 | 优先可用性、性能优先、配置简单、relay 运行时调优 + 每日动态摘要扫描 |
| 最终长 MD 提示 | 用户需求 | canonical plan、严格路由规则、成本感知探测、官方价 / 网关价、部署计划 |
| 用户更正要求 | 用户需求 | 按次 / 按量属于 model / route-target 级，而不是简单 key 级 |
| 用户新增要求 | 用户需求 | 项目级备份 / 导入 / 迁移适配 / 回滚 / 差异预览 |
| 用户新增要求（图片与上下文收口） | 用户需求 | 图片点名文档一致性、帮助提示体系、熔断自定义增强、`Codex` 施工口径、分目录 agent 可执行边界 |
| 用户新增要求（AI 自动化中心） | 用户需求 | 顶层 AI 自动化栏目、AI Profile 双轨配置、手动/AI 方案切换、动态路由本地 AI 学习 |
| README / 仓库证据 | 仓库现状 | 当前架构、功能、多 key 渠道、stats、relay、日志流、UI 结构 |
| 我的工程增强 | 增强项 | 迁移布局、回滚、观测、relay 运行时调优骨架、测试计划 |

---

## 15.1 实现禁止事项（路由与故障转移）

### 严格轮询禁止项

- 禁止全局轮询游标
- 禁止记忆上次成功位置
- 禁止从上次位置继续
- 禁止后台自动平衡顺序

### 严格权重禁止项

- 禁止 weighted random
- 禁止概率抽签
- 禁止平滑近似分配替代严格序列
- 禁止因成功率、价格等自动改写周期顺序

### 故障转移禁止项

- 禁止把故障转移退化成普通 retry
- 禁止在 paid 场景默认多路并发
- 禁止无取消机制的并发竞速
- 禁止无 360 秒总窗口控制的无限循环

---

## 16. 实施规则

本文档是当前工作流的唯一真相源。

如果实现必须偏离：

1. 先更新本文档。
2. 再实现改动。
3. 最后验证 diff 与文档一致。
