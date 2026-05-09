# AI 自动化完整重构移交文档

## 文档目的

这份文档不是给下一位维护者“接着修补当前版本”，而是给即将接手 `AI Automation` 全部业务的 CLI / AI 模型一个明确结论：

- 当前实现只能作为“接口和业务边界参考”，不能作为最终交付基线。
- 前端现状、运行脚本、交互组织、结果消费方式都存在结构性问题。
- 下一位接手者应按“全面重新重构”处理，而不是继续在现有页面上打补丁。

本文件重点回答四件事：

1. 哪些地方最容易踩坑。
2. 哪些地方虽然做过，但当前并不满意。
3. 哪些地方必须推倒重做。
4. 一个最原始、最完整、最硬性的 AI 自动化重构要求。

## 一句话结论

`AI Automation` 当前已经堆了很多“看起来有功能”的前后端碎片，但没有形成一个成熟、稳定、可用的 AI 自动化产品闭环。

必须把它视为：

- 有可复用接口
- 有部分可复用数据模型
- 有部分可复用非破坏式约束
- 但整体前端交互、结果消费、运行验收、脚本链路都不达标

下一位接手者应该以“完全重构”为目标，而不是“局部美化”或“继续沿用当前组织”。

---

## 当前可参考但不要盲继承的范围

### 前端入口

- [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx)
- [web/src/components/modules/ai-automation/workbench-shared.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/workbench-shared.tsx)
- [web/src/components/modules/ai-automation/*.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation)

说明：

- 这些文件只能用于理解“当前页面拆成了哪些块”。
- 不建议直接继承当前 JSX 结构、样式基元、布局组织。
- 当前页面虽然被改成所谓“指挥舱型”，但仍不是一个真正成熟的 AI 自动化工作台，只是本轮过渡实现。

### 前端 API 契约

- [web/src/api/endpoints/ai-automation.ts](/D:/GPT-codex/octopus_repo/web/src/api/endpoints/ai-automation.ts)

说明：

- 这是最值得保留的前端参考之一。
- 下一位接手者应优先从这里理解任务、Profile、动态学习、任务产物的读写接口。
- 但不要把当前前端状态组织方式一并继承。

### 后端 handler 边界

- [internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go)

说明：

- 这里定义了 `/api/v1/ai/*` 和部分动态学习读写入口的 HTTP 边界。
- 下一位接手者可以在不破坏协议的前提下补字段、补只读聚合，但不要轻易把路径设计打散。

### 后端核心执行逻辑

- [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go)
- [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go)
- [internal/model/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go)

说明：

- 这些文件包含现有任务生命周期、任务步骤、Profile 产出、非破坏式工具键、结果保存等核心逻辑。
- 这里有可以复用的业务原则，但当前结构非常重、非常散、隐性语义太多。
- 建议重构时优先保留领域边界，不保留当前实现细节。

---

## 最容易踩坑的地方

### 1. 误把当前页面当成“可继续迭代”的基础

这是最大坑。

当前页面不是稳定基线，原因是：

- UI 风格前后摇摆过多，存在历史方向冲突。
- 页面已经承载了太多中间态决策，视觉和交互都不纯。
- 组件拆分虽然存在，但很多拆分只是展示层拆开，不是真正的领域拆分。
- 页面里仍然混合了配置台、结果台、学习台、资产台、任务台的多重职责。

结论：

- 不要以“保留 80% 再改 20%”的心态接手。
- 要以“保留协议与业务原则，重建产品结构”的方式接手。

### 2. 误把“非破坏式”理解成“什么都不能做”

当前 AI 自动化的业务原则不是“AI 什么都不能碰”，而是：

- 不能直接改底层渠道配置。
- 不能直接改分组表和 `group_items`。
- 不能直接改优先级、路由表、底层运行策略。
- 可以生成 Profile。
- 可以写任务快照。
- 可以在显式用户动作下激活 Profile。
- 可以提供保护动作建议与执行结果。

真正需要保留的是“非破坏式默认边界”，不是把整个系统做成只会展示文本的假工作台。

### 3. `manual / ai_profile` 双来源链很容易被重写坏

重点文件：

- [web/src/components/modules/ai-automation/config-source-logic.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/config-source-logic.ts)
- [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx)
- [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go)

当前来源模式不是单纯的二选一，而是有三层含义：

- 用户请求来源 `requested_config_source_mode`
- 当前实际生效来源 `config_source_mode`
- Profile 请求是否因为无效/缺失/内容问题而发生运行时回退 `source_fallback_reason`

容易踩坑的点：

- 把“请求来源”和“运行来源”混成一个状态。
- 切到 `ai_profile` 时自动覆盖手动配置。
- 回退发生时把原始请求状态清掉，导致用户不知道系统为何回到 manual。

必须保留的原则：

- 用户请求态和运行态分离。
- 回退原因可见。
- manual 配置永远保留，不被 AI Profile 覆盖。

### 4. 结果消费契约非常脆弱

重点文件：

- [web/src/components/modules/ai-automation/result-logic.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/result-logic.ts)

当前结果消费逻辑高度依赖：

- `result_json`
- `result_payload`
- `tool_execution`
- `tool_execution_summary`
- `domain_payload`
- `writes.profile_write.profile_id`

问题：

- 当前前端需要自己从 JSON 深处拼语义。
- 结果结构虽然能用，但不稳、不直观、不够产品化。
- 任何字段名或嵌套调整，都可能导致前端多个区块一起坏掉。

下一位接手者建议：

- 不一定重写协议，但至少要增加稳定的只读聚合字段。
- 前端不要继续深挖原始 JSON 才能知道“有没有 Profile 产物”“有没有保护动作”“结果摘要是什么”。

### 5. 多 AI 当前只是前端 fan-out，不是真正的 orchestration

重点文件：

- [web/src/components/modules/ai-automation/DispatchWorkbench.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/DispatchWorkbench.tsx)
- [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go)

当前所谓“多 AI 调度”本质上是：

- 前端根据 lane 角色多次发任务创建请求。
- 后端没有真正的多任务协调协议。
- 没有真正的 planner/executor/reviewer 共享中间状态编排。

这意味着：

- 当前多 AI 更像“多次提交不同任务”，不是“一个多智能体工作流”。
- 如果下一位接手者误把它包装成成熟多智能体系统，会导致设计与真实能力严重错位。

下一位必须先做一个清晰决定：

- 要么明确继续保留“前端高级 fan-out 模式”，只做产品收敛。
- 要么明确升级为真正的 orchestration，但那已经是新项目级别范围。

### 6. 动态学习视图和设置页联动有 focus-target 约束

重点文件：

- [web/src/components/modules/ai-automation/focus-target.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/focus-target.ts)
- [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx)
- [web/src/components/modules/setting/DynamicRouting.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/DynamicRouting.tsx)

这个点很小，但很容易被重构时不小心打断。

当前约束：

- 设置页动态学习入口可把焦点跳到 AI 自动化页的 learning 区。
- 它依赖 `sessionStorage` 和固定的 `data-testid` / `data-ai-focus-target` 约定。

如果重构时无视这个联动，设置页会失去进入 AI 中心学习区的顺滑入口。

### 7. 浏览 smoke / 运行验收链路本身不稳定

重点文件：

- [scripts/verify-ai-automation-learning-browser-smoke.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-ai-automation-learning-browser-smoke.mjs)

已经确认的坑：

- Windows 下临时 SQLite 目录删除会遇到 `EBUSY` 文件占用。
- 自启动 smoke 依赖前端 `3101`、后端 `18081`，但自启动链路本身不稳定。
- 本地浏览调试还受到 Chrome/Edge、本机 `ws` 模块、CDP 端口等环境因素影响。
- 当前浏览验收链路不能被视作“稳定、可信、一键通过”的最终门禁。

结论：

- 下一位接手者重构产品时，必须顺手重构运行验收链路。
- 否则会一直在“代码看起来能用，但运行态无法稳定证明”这个泥潭里打转。

### 8. 当前仓库里存在真实敏感数据风险

运行态检查时已经观察到本地页面里存在真实样式的 API key 明文展示。

这说明：

- 某些页面/某些测试数据/某些开发环境数据没有被充分脱敏。
- 下一个接手者在调试 AI 自动化时，极有可能误把真实数据当测试数据继续传播。

必须要求：

- 所有交接、测试、截图、demo、录屏、日志，一律脱敏。
- 不允许把真实 key、真实 base_url、真实业务数据当作“方便调试”的默认样本。

---

## 已做过但当前不满意的地方

### 1. UI 已经重做过多轮，但方向始终不稳定

现状：

- 做过偏轻量 SaaS 的版本。
- 做过偏指挥舱型的版本。
- 做过“尽量贴项目原风格”的版本。
- 做过“明确不要贴当前项目风格”的版本。

问题：

- 这些版本切换太频繁，导致页面产物不纯。
- 当前实现虽然是“指挥舱型”，但还没有形成真正稳固的产品语言。
- 视觉上已经变了，产品结构上却还没完全跟上。

不满意结论：

- 当前版本不是终稿。
- 应由下一位接手者重新定义完整的产品级风格与信息层级。

### 2. 组件拆分存在，但职责边界不干净

现状：

- 组件很多，文件也拆了。
- 但拆分大多围绕“显示区域”，不是围绕“产品职责”。

表现为：

- 主页面 `index.tsx` 状态过多。
- 页面仍承担大量 orchestrator 角色。
- 子组件 props 很长，很多只是把中心状态分发出去。

不满意结论：

- 这不是健康的工作台架构。
- 下一位应重新做状态分层，而不是继续在当前 state pile 上迭代。

### 3. 结果面板不够“产品化”

当前结果区能看：

- 摘要
- 对比
- 原始
- 保护动作
- Profile 产物

但不满意点：

- 仍然偏“调试台”，不是真正的结果工作台。
- 对比语义太弱，缺乏操作上下文。
- 原始视图可读性差。
- 用户真正应该下一步做什么，不够明确。

### 4. 配置区仍然混在 AI 页面里，边界感不够清晰

虽然设置页已经轻量化了，但 AI 页面内部仍然保留了一块结果页下方的手动配置区。

不满意点：

- 它不是主链，也不是完整设置台。
- 既存在，又不够完整，容易让用户困惑。
- 下一位需要重新决定“配置”在 AI 中心里的产品定位。

### 5. 多 AI 区虽然收起来了，但价值表达还不清楚

当前做法只是把多 AI 放到二级视图，避免抢首屏。

不满意点：

- 仍然没有回答“为什么我要用多 AI”。
- lane 的配置能力存在，但收益展示弱。
- 高级用户可能觉得浅，普通用户可能觉得复杂。

### 6. 学习区可用，但没有成为真正可操作的学习中枢

当前学习区可以：

- 看是否启用
- 看样本数
- 看最近样本
- 看 top target
- reset

但不满意点：

- 对学习效果的解释太弱。
- 对用户而言，为什么要看、看完能做什么，不足够清晰。
- 它更像一个“后台状态窗口”，不是“可管理的学习面板”。

### 7. 运行脚本和浏览冒烟明显不成熟

这个点已经不是“不满意”，而是“不可接受”。

原因：

- 无法稳定证明前端运行效果。
- 很容易把 UI 问题和脚本问题混在一起。
- 会严重拖累后续任何模型或 CLI 的交接效率。

---

## 必须重做的地方

### 1. 整个前端信息架构必须重做

不是换皮，是重做信息架构。

必须重新明确：

- 首屏到底只服务什么主任务。
- 结果区到底是观察台还是操作台。
- Profile、Dispatch、Assets、History、Learning 五个舱的优先级和用户动线。
- 配置区是否继续作为 AI 页面内部一部分存在。

### 2. 前端状态架构必须重做

当前 `index.tsx` 聚集了太多页面级状态。

下一位应重新设计：

- 哪些是运行态状态。
- 哪些是表单态状态。
- 哪些是视图态状态。
- 哪些应该下沉到模块内部。
- 哪些应该抽成 store / reducer / orchestration hook。

### 3. 结果消费契约必须重做或补聚合层

要求：

- 前端不再依赖深度 JSON 挖掘判断业务语义。
- 至少为结果区提供稳定的摘要层字段。
- Profile 产物、保护动作、工具执行摘要都应有稳定消费入口。

### 4. 运行态 smoke / 浏览验收链必须重做

必须把 AI 自动化的运行态验证做成可交接、可复现、可自动执行的链路。

至少要解决：

- 自启动端口链路稳定性。
- Windows 临时文件占用问题。
- 登录态注入方式。
- 浏览器/CDP 依赖硬编码问题。
- 验收结果和脚本异常要能区分。

### 5. 文案体系必须重做

当前文案来源很多，历史包袱重。

下一位应重新统一：

- 导航语言
- 面板语言
- 结果语言
- 高级区语言
- 风险/保护动作语言

目标是：

- 简洁
- 客观
- 少解释废话
- 不营销
- 不聊天口吻

### 6. 多 AI 能力定位必须重做

必须明确它到底是：

- MVP 高级 fan-out 区
- 还是未来真正的多智能体 orchestration 起点

这个决定不做清楚，后面所有产品和后端设计都会持续摇摆。

---

## 给下一位 AI / CLI 的完整原始重构要求

以下要求视为“最原始版本的完整重构要求”，不是建议，是必须遵守的总纲。

### 产品定位

`AI Automation` 必须被重构成一个独立、成熟、可持续扩展的 AI 工作台，而不是设置页副产物，也不是任务试玩页。

产品目标：

- 服务真实运营 / 管理 / AI 辅助配置分析场景。
- 强调 AI 主链执行、结果观察、Profile 产出、运行边界和学习反馈。
- 明确是一个 MVP，但必须是“能用的 MVP”，不是“概念型页面”。

### 保留不变的约束

- 顶层入口仍为 `ai`。
- 不新增新的顶层导航。
- 不改现有 `/api/v1/ai/*` 路径主干。
- 不改现有 `/api/v1/dynamic-routing/learning*` 主干。
- 不让 AI 直接改渠道、分组、`group_items`、`priority`、核心路由配置。
- 允许生成 Profile、快照、显式激活 Profile。
- manual / ai_profile 双来源机制必须保留。

### 必须达到的前端结构目标

- 单页内工作台，不是长滚动拼贴页。
- 首屏主任务链必须非常清晰。
- 结果区必须默认摘要优先。
- 二级区必须有稳定的固定高度滚动容器。
- 高级功能必须下沉，不抢首屏。
- 页面必须兼顾桌面端和窄屏移动端，不允许横向溢出。

### 必须达到的体验目标

- 用户一进来就知道“现在该做什么”。
- 用户跑完任务就知道“结果是什么、能不能用、下一步是什么”。
- 用户生成 Profile 后能明确感知它和 manual 配置的关系。
- 用户进入学习区后能明白它不是抽象分数板，而是可理解、可管理的运行反馈区。
- 高级用户能用多 AI 和工具开关，普通用户不会被这些东西淹没。

### 必须达到的工程目标

- 组件职责清晰。
- 页面状态分层清晰。
- 结果消费契约稳定。
- 测试覆盖主链行为。
- 运行态 smoke 可复现。
- 文案四语完整。
- 不出现 key path 泄漏、英文硬编码回退、乱码、混语。

### 必须达到的验收目标

- `tsc` 通过。
- AI 自动化定向测试通过。
- learning focus 静态守门通过。
- 静态导出通过。
- 至少一条隔离运行态 smoke 能稳定证明页面真实可用。
- 手动验收时，首屏、结果区、Profile、Dispatch、Assets、History、Learning 都能被实际点击和验证。

---

## 建议下一位接手者的执行顺序

1. 先读接口和领域模型，不要先读当前 UI。
2. 先定义新的产品结构，再写 UI，不要沿用现有 JSX 继续拼。
3. 先决定多 AI 的定位，再决定 Dispatch 区怎么设计。
4. 先补结果聚合契约，再做结果台。
5. 先修运行态 smoke，再做最后视觉迭代。

---

## 不建议继续做的事情

- 不要继续在当前页面上“再调一点样式”。
- 不要继续给当前 `index.tsx` 叠更多状态。
- 不要把当前多 AI fan-out 包装成完整多智能体系统。
- 不要把运行态脚本问题误判成页面没问题。
- 不要把真实 key / 真实环境数据带进后续交接和截图。

---

## 参考文件

- 前端入口：[web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx)
- API 契约：[web/src/api/endpoints/ai-automation.ts](/D:/GPT-codex/octopus_repo/web/src/api/endpoints/ai-automation.ts)
- 来源切换逻辑：[web/src/components/modules/ai-automation/config-source-logic.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/config-source-logic.ts)
- 结果消费逻辑：[web/src/components/modules/ai-automation/result-logic.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/result-logic.ts)
- 设置页来源入口：[web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx)
- focus-target 联动：[web/src/components/modules/ai-automation/focus-target.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/focus-target.ts)
- 后端 handler：[internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go)
- 后端配置与任务入口：[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go)
- 后端执行器：[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go)
- 模型定义：[internal/model/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go)
- 当前 UI 主线文档：[docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md)
- 当前学习 smoke 脚本：[scripts/verify-ai-automation-learning-browser-smoke.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-ai-automation-learning-browser-smoke.mjs)

---

## 最后说明

如果下一位接手者只能记住一句话，请记住这句：

`不要继续修当前 AI 自动化页面，请保留协议和业务边界，但把整个 AI 自动化当成一个需要重新定义产品结构、重新定义结果契约、重新定义运行验收的独立系统来重做。`
