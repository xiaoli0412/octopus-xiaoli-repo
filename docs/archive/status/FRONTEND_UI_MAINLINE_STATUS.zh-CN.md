# 前端主线落地说明（对齐 LLM-Gateway-Refactor-Plan 第 9 节）

> 范围：本文只覆盖 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9` 节相关前端工作，用于持续记录 UI 主线落地状态、完成度与下一步优先级。
>
> 边界：本文不替代 canonical MD，不替代后端设计文档，也不替代 worklog；它只负责把“前端已经做了什么、还差什么、下一刀改哪里”单独说清楚。

> 执行硬规则：后续任何实现、补丁、回归修复、UI 调整、路由策略修改、接口扩展与验收动作，都必须先对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)。
> 若当前文档与主规划冲突，以主规划为准；若实现需要偏离主规划，必须先更新主规划，再改代码，再补验证记录。
> 当前前端主线状态还必须同步对齐 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 中的图片优先问题池、帮助提示要求、全中文要求、`CC Switch` 重做要求和 `Codex` 执行口径。

---

## 0. 当前会话前端优先池（2026-04-23）

在现有前端主线状态之外，当前会话新增以下强制优先返工项：

- 图片暴露的版本卡片、首页卡片、创建渠道弹窗、创建分组弹窗、`Route Target Overrides`、`CC Switch` 交互问题优先修复。
- 中文界面 i18n key 泄漏和中英混杂问题必须优先清理。
- 探测与检测设置必须从价格区域迁出，归位到设置页。
- 最近一轮把主壳层、首页和分组页收得过窄，已偏离用户要的“更铺屏、更接近原版”的占比，必须优先回退并保持防溢出。
- 设置页被改成单栏顺序流，已偏离原版双栏瀑布流；必须恢复双栏，并控制卡片留白和阅读节奏。
- 左侧 dock/切换栏需要回到桌面端左侧中线附近的稳定停靠位，不得再出现偏左上或失衡的停靠感。
- 全局字体回退、字号与行高必须统一修复，避免中文界面继续出现“字体怪”“过小过挤”和文本漂移。
- 增加“圈内问号 + 悬停帮助提示”体系，并保证帮助内容与当前真实实现一致。
- 帮助提示与介绍信息减字收缩升级为关键整改点：默认只保留问号提示和一句主说明，减少硬文本暴露，不再让帮助与介绍抢占主交互区。
- 熔断设置从基础开关升级为“默认简洁 + 高级自定义展开”的可解释配置面板。
- 首页统计卡片、渠道卡片和 Token 明细区的桌面端错位、挤压、占地过大问题进入 P0 UI 返工池。
- 渠道创建/编辑弹窗中的多 Key 交互必须重做成同渠道内折叠结构，并明确区分 key 值、开关、备注和模型范围。
- 分组创建/编辑弹窗中 `group.form.*`、`modeBadge.*` 等原始 key 泄漏进入 P0 缺陷池。
- 模型/价格区域必须补普通布局与紧凑布局两档，并清理 `Official / Canonical / unknown / passive_only` 等英文主显示。
- `2026-04-23` 已补上模型页布局状态真正下传到卡片层的 no-browser 收口：工具栏里的“普通 / 紧凑”切换不再只停留在入口状态，`ModelItem` 会跟随 `layout` 切换普通卡片和紧凑卡片，`verify-llm-price-boundary.mjs` 也已补了对应回归断言。
- 同日继续收紧了模型卡片的信息顺序：普通布局下的 `规范名称 / 计费模式 / 官方价格` 已集中到统一的中文 meta 信息带，避免信息分别散落在标题下方和卡片尾部；`verify-llm-price-boundary.mjs` 也新增了对该信息带与顺序的 no-browser 断言。
- 同日继续把模型页搜索行为对齐到当前中文提示：搜索词现在会同时命中模型名称和规范名称，不再出现输入框写着“搜索模型名称或规范名称”但代码只按 `name` 过滤的契约偏差；`verify-llm-price-boundary.mjs` 已补上对应 no-browser 断言。
- 备份卡片中的整段英文说明与“剩余迁移能力”区域的中文主文案混杂已完成 no-browser 收口；本轮进一步把导出快照 scope badge 与备份诊断 warning / skip reason 收口到 locale 输出，避免中文界面仍露出英文 badge。
- 性能优化并入本轮前端主线，重点减少首屏与弹窗默认展开的 DOM 压力。
- 所有正文说明块、默认帮助段落和异常态解释都要同步做减字收缩；如果同一信息已经进入 `HelpHint`，正文不得再重复摊开长段解释。
- 新增顶层 `AI 自动化` 栏目规划：后续前端需要在主导航中新增独立入口，用于 AI 渠道/模型配置、自然语言任务、提示词模板、进度条、AI Profile 结果预览和历史任务，不得把这些能力散落在价格页或设置页。
- 设置页需要新增配置来源切换：`手动配置 / AI 生成方案`，并清楚说明 AI Profile 不覆盖用户手动配置。
- 动态路由设置页需要新增 `dynamic_routing_learning_enabled` 开关和学习状态入口，动态路由 AI 学习不通过普通 AI Profile 覆盖。

本文件后续更新必须同步引用 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 对应需求编号，避免只写完成度不写验收口径。

### 0.1 AI 自动化顶层栏目规划

后续前端实现 `AI 自动化` 栏目时必须满足：

- 与 `home/channel/group/model/log/setting` 同级。
- 页面保持现有深绿、圆角、渐进展开风格。
- 首屏展示当前 AI 模型状态、AI endpoint 配置、任务类型卡片和自然语言输入。
- 模型列表支持自动获取，并展示来源、可用性、免费/付费倾向、成功率和延迟。
- 任务执行必须有进度条，不能只显示加载中。
- AI Profile 结果区必须展示自然语言总结、结构化内容、置信度、风险提示和保存状态。
- 中文界面不得泄漏 raw enum、内部 key 或无解释英文主文案。

前端相关需求编号：需求 54-64。

---

## 1. 对齐基线

- canonical 来源：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1` 到 `9.7` 节
- 最近参考记录：
  - `docs/worklog/2026-04-17-ui-group-home-followup.md`
- 当前覆盖页面：
  - 渠道页
  - 分组页
  - 日志页
  - 首页
  - 模型添加 / 同步相关 UI

本文不覆盖的事项：

- 后端路由语义与数据模型设计
- 未暴露到前端的后端统计字段设计
- 更底层的 relay / route-target 运行时策略实现细节

## 2. 渠道页状态

对齐章节：`9.1 / 9.1.1`

### 已完成

- 已满足“一个渠道内显式管理多个 key”的主线方向，渠道详情区不再只是平铺 key 字符串。
- 渠道卡片已显示 `classified / pooled` 与 key 路由策略 badge。
- 渠道详情区已新增 routing 信息区，显式展示 key 模式与 key 策略说明。
- key 列表已升级为独立 key 卡片，可展示：
  - masked key
  - status code
  - total cost
  - `source_type`
  - `remark`
  - `allowed_models`
  - last used time
- key 列表已支持：
  - 本地搜索 / 筛选
  - 折叠 / 展开
  - 单 key 测试
  - 单 key 直达编辑
- 渠道表单已支持 `focusKeyId`，可从详情区直接跳到对应 key 编辑位。
- `2026-04-23` 已继续收口渠道卡片与详情区的多 Key 展示：
  - 渠道卡片副标题改为“渠道类型 + 密钥数量”，卡片 badge 区补齐密钥数量，不再回退成裸 `Key:` 文案。
  - 渠道详情区 key 列表顶部新增“总数 / 启用数 / 当前筛中数”摘要线。
  - 单 key 折叠头统一展示备注或回退标签、启用状态、状态 badge、来源、密钥预览、成本与允许模型数；展开区补齐“尚未使用”等兜底文案。
  - `verify-channel-presentation.mjs` 已新增并纳入 no-browser 回归护栏，防止上述结构和文案再次回退。
- `2026-04-23` 已继续收口渠道创建/编辑弹窗中的多 Key 输入引导：
  - 折叠头新增“待填写真实密钥 / 真实密钥已填写”状态 badge，不再只靠用户自己猜测当前 key 是否已经具备可用凭据。
  - 展开区改成“第 1 步：填写真实 API 密钥”与“第 2 步：按需补充进阶设置”的两段式结构，把真实 key 输入位从来源分类、备注和模型范围里明确抬出来。
  - `verify-channel-create-flow.mjs` 已同步升级，补上对新状态 badge、两步式标题和主状态提示的 no-browser 断言。
- 同日已补齐渠道页四语 locale 缺口，并把同批直接可见路径中的 `Key / Token` 残留词进一步收口：
  - `zh-Hant`、`ja` 不再在渠道页主显示中混入 `Key / Token`。
  - `en` 补齐渠道页新字段，避免新 UI 落地后英文语言包断 key。
  - `verify-locale-consistency.mjs` 已重写并把 `channel.card` 也纳入一致性检查。
- `2026-04-24` 已继续补齐渠道页浏览器级主证据：
  - 渠道页根节点、卡片、详情弹层、route target 摘要、key 过滤框、key accordion 与单 key 测试结果区现已补齐稳定 `data-testid` 锚点，browser smoke 不再依赖脆弱的文案匹配。
  - 已新增 `scripts/verify-channel-page-browser-smoke.ps1`，并复用宿主已验证通过的 Edge CDP 包装链路，通过 `channel-page` 场景在本机准备最小渠道数据、route target override 与双 key 背景后，验证工具栏中的“提供商 + 模型关键词 + key 信息”组合筛选、详情弹层打开、key 行展开，以及 `375px` 宽度下无明显横向溢出。
  - `scripts/verify-channel-presentation.mjs` 也已补上对应 selector 合同护栏，防止页面级 smoke 依赖的 channel page 锚点后续无声回退。
  - 同日继续补上了单 key 详情里的 route-target override 可见性：展开单 key 后会直接显示该 key 命中的计费 / 探测摘要行与覆盖条数，`keyFilter` 也可命中模型名、计费模式、探测策略和间隔/并发数字，不再只有渠道级 route target 摘要而缺 key 级入口。
- `2026-04-28` 已继续收口渠道页的“模型列表获取”与本机运行默认地址：
  - `web/src/components/modules/channel/model-fetch.test.tsx` 已在 repo-local `vitest` 下稳定通过，确认 `分类 K` 模式会按“当前 key + 当前 base URL + 当前代理 / 请求头配置”请求 `/api/v1/channel/fetch-model`，不再回退成拿第一把可用 key 兜底。
  - `web/src/components/modules/navbar/DocModal.tsx` 与四语 locale 里的文档示例默认地址已统一切到 `http://127.0.0.1:8080`，避免这台 Windows 宿主的 `localhost` 解析异常继续干扰文档示例、手工验证和 no-browser 测试链。
  - 渠道页与设置 / AI 自动化相关的简中、繁中文案又完成一轮主显示英文清理：`Profile / Hybrid / Shadow AI` 等高频直出词已继续向中文方案档案、混合模式、影子模式收口，并同步加严了 `verify-locale-consistency.mjs` 的断言，降低后续回退风险。
- `2026-04-28` 最新 UI 恢复轮又继续压缩渠道创建/编辑弹窗里的多 Key 折叠手感：
  - 多 Key 顶部摘要从大块统计卡改成单行摘要与小徽标，降低默认 DOM 和视觉占用。
  - 每个 key 的折叠头只保留一个展开手势和 `ChevronDown` 状态提示，不再在同一行放重复展开按钮。
  - 未展开时只暴露备注、启用状态、真实密钥是否已填写、已选模型数等判断信息；真实 key 输入、按 key 请求模型、允许模型列表和高级项都留在展开层。
  - 该轮与 `scripts/verify-channel-create-flow.mjs` 的新断言对齐，防止后续又把多 Key 区回退成首屏硬铺开。

### 当前判断

- `9.1` 主线已基本完成。
- 渠道页当前已经具备“同一 channel 内显式管理多个 key”的可见结构与基本操作能力。

### 仍待增强

- 更轻量的 key 快速编辑仍可继续优化，但不影响 `9.1` 主线闭环。
- `9.1.1` 中关于熔断状态、额度 / 剩余额度、计费方式、探测策略、优先级 / 权重等字段，前端仍未全部补齐；其中部分依赖后端字段。
- 搜索能力中“模型家族名 / 失败率 / 最近成功时间 / 熔断状态”等维度仍未完全覆盖；当前已先收口到提供商、模型关键词和 key 信息三级组合筛选。
- 渠道总页的 selector 合同、详情弹层结构与 no-browser 守护已闭环；当前宿主机上的真实 browser-grade pass 仍受 Edge/CDP bootstrap 阻塞，剩余浏览器缺口主要收敛到创建/编辑弹窗更细的 hover/focus 细节，以及后续在可用宿主上补齐 `channel-page` 的真实 pass 证据。

- 当前渠道页的工具栏筛选已继续收口为“启用状态 + 提供商 + 模型关键词 + key 信息”的组合筛选；后续浏览器验收应优先验证这四路筛选组合是否符合真实使用习惯。
- 当前渠道创建/编辑弹窗的多 Key 区已新增“总数 / 已填写真实密钥 / 已启用 / 待补充”的摘要线，后续浏览器验收应确认该摘要在桌面端和窄屏下都不挤压、不误导。
- 当前剩余的 P0 渠道页缺口已更明确收敛为“浏览器真实手感”而不是链路断开：重点不再是模型请求发不出来，而是继续补桌面端占地、折叠动画手感、375px 触达性，以及更深层弹窗里的中文主显示收尾。

## 3. 分组页状态

对齐章节：`9.2 / 9.2.1 / 9.3`

### 已完成

- 分组页保留了原有 group 编辑风格，没有破坏主流程。
- 分组策略与重试相关参数已在 UI 中显式暴露：
  - `retry_rounds`
  - `retry_delay_ms`
  - `failover_window_sec`
  - `race_after_fails`
  - `race_concurrency`
- 右侧成员列表继续支持排序 / 拖拽，保持“看到的顺序 = 实际执行顺序”的方向。
- 左栏“添加模型”区域已从单层平铺升级为 `channel -> section -> model` 分层结构。
- 已支持本地搜索 / 筛选。
- 已显式显示 `pooled / classified` 模式 badge。
- `2026-04-23` 已补齐分组创建/编辑路径的本地化兜底收口：
  - 正则错误不再把 JS 原始英文 message 直接回显到界面，而是统一显示本地化主提示与简短修正说明。
  - 分组卡片成员列表在缺渠道名时不再回退成硬编码 `Channel {id}`，而是统一走 locale 文案。
  - `verify-group-create-flow.mjs` 已补上上述两处兜底断言，防止后续回退。
- 同日继续收口了分组创建高级策略折叠头的帮助提示结构：`HelpHint` 已从原始 `AccordionPrimitive.Trigger` 按钮树内移出，改为复用公共 `AccordionTrigger` 的 `addon` 槽位，避免在该入口继续累积嵌套交互结构风险；`verify-group-create-flow.mjs` 也已补上对应静态断言。

### `9.2.1` 结构表达状态

#### classified

- 已按 `keys[].allowed_models` 做 key 分段展示。
- 已补 `unassigned` fallback，避免静默丢失模型。
- 已增强 key section 的可理解性，可稳定展示 `remark / source_type / allowed_models` 相关语义。

#### pooled

- 已补齐 `Shared Model Set + Key Pool` 结构表达。
- 用户现在可以在左栏直接看到 pooled 不是普通模型列表，而是“共享模型集 + 多 key 池”。

### 当前判断

- `9.2` 主线已大体完成。
- `9.2.1` 的核心结构表达已完成，且没有破坏 group 提交语义。
- `9.3` 要求的本地关键词筛选框已满足。

### 仍待增强

- 大量 key 场景下的虚拟列表仍未做。
- 三层嵌套结构在手机端虽然可用，但仍有继续打磨空间。
- group 保存结构目前仍是 channel/model 粒度；若未来要做 key/model 精确绑定，需要单独的后端与交互设计，不应由前端擅自越界改动。
- 浏览器级 `375px` 证据与 hover/focus 行为仍待在同一 screenshot-first 池中补齐；当前只完成了 no-browser 文案与兜底闭环。
- `2026-04-30` 已继续收口创建分组弹窗的同池缺口：当前 `group-create` browser smoke 不再误选 Codex 内置 Node，`scripts/verify-group-create-browser-smoke.ps1` 与 `...cdp.ps1` 已对齐到和 `channel-create` 相同的外部 Node 选择策略；同时 [`GroupEditor.tsx`](D:/GPT-codex/octopus_repo/web/src/components/modules/group/Editor.tsx) 已恢复 `flow` 摘要卡、`advanced-strategy` 折叠区、稳定 `new-group-*` / `edit-group-*` testid、模型筛选空态，以及 `retry_rounds / retry_delay_ms / failover_window_sec / race_after_fails / race_concurrency` 的创建/编辑提交链。对应 `verify-group-create-flow.mjs`、`tsc --noEmit`、`check-only` 与真实 `self-start` browser smoke 已重新通过，因此当前 group 主线不再停留在“高级策略字段只存在于 API/locale，UI 丢失”的回退状态。
- `2026-04-30` 同一条 `group-create` 验证子线又继续把顶层 CLI PowerShell wrapper 从自复制实现收口到共享 `verify-channel-create-browser-smoke.ps1`：现在 `group-create` 的 `Driver=cli` 路径会和 `backup / ccswitch` 共用同一套 loopback 预检、稳定日志读取与 `spawn EPERM` host-blocker 分类，而不再单独维护一份旧版 CLI 包装逻辑。直接结果是 `check-only -Driver cli` 继续通过，且本机 `self-start -Driver cli` 现在会像其它共享 CLI 页面一样明确报出 Playwright CLI 子进程 `spawn EPERM`，说明当前 `group-create` 剩余的 CLI browser gap 也已收敛为宿主执行环境问题，而不是分组创建页面自身回归。

## 4. 日志页与首页状态

对齐章节：`9.4 / 9.5 / 9.5.1`

### 日志页

#### 已完成

- 现有日志流、历史分页、SSE 实时推送、滚动加载更多行为已保留。
- 已有导出能力，且前端现在已显式暴露：
  - `json / jsonl` 格式
  - 开始 / 结束时间
  - 导出上限
  - 最近 `24h / 7d` 快捷范围
- 已补前端轻校验：非法时间范围会在提交前提示并禁用确认导出。
- 导出面板已兼顾手机端宽度与底部按钮堆叠。

#### 当前判断

- `9.4` 主线已完成。
- 日志页现在不只是“能导出”，而是已经具备更接近真实使用场景的导出交互。

#### 仍待增强

- 后续仍可继续增加“最近 30 天”等快捷范围，但这属于体验增强，不是主线硬缺口。

### 首页

#### 已完成字段

- `总 token 使用`
- `按 channel 的 token 使用`
- `按 provider 的 token 使用`
- `按 model 的 token 使用`
- `成功率 / 失败率摘要`
- `估算官方价格 / 估算 gateway 价格摘要`
- `probe cost`
- `熔断状态摘要`
- `最近探测状态摘要`

#### 已完成实现

- `Total` 卡片已展示 token / request / cost 等基础统计。
- `2026-04-23` 已把首页总览区重排为“请求主卡 + 总量/输入/输出三张摘要卡”的稳定比例，不再把所有指标挤在同一层横向列里。
- 首页总览已补：
  - 成功请求
  - 失败请求
  - 成功率
- `Token Breakdown` 已支持：
  - `by_provider`
  - `by_channel`
  - `by_model`
- `2026-04-23` 已把 `Token Breakdown` 收口为“默认摘要 + 按需展开运行摘要”结构：
  - 默认首页只展示总量、输入、输出、估算网关价格和三组 Top 列表
  - 估算价格、熔断状态、最近探测摘要统一移入同页折叠区
  - 排行列表默认展示前 3 项，可继续展开更多，不再首屏一次性全部摊开
- `2026-04-23` 已把首页统计区右侧的活动热力图下移到主网格之后，减少桌面端首屏错位和纵向挤压。
- 首页已新增基于 `by_model token + 模型价格表` 的估算价摘要：
  - `estimated official price`
  - `estimated gateway price`
  - `estimated delta`
- 首页已新增 recent runtime probe 摘要：
  - recent probe count
  - recent probe success / failed
  - last probe status / channel / model / message
- 首页已新增 breaker runtime 摘要：
  - tracked breakers
  - open breakers
  - half-open breakers
  - max remaining cooldown
- 首页已新增 estimated probe cost 展示位。
- 首页中文主显示已继续收口：`Token 明细 / Gateway 价格 / Tokens` 等首页主标签已改成中文表达，保留专名时也包裹在中文语境中。
- 布局仍保持移动端友好，未破坏现有首页风格。
- 已新增 `scripts/verify-home-layout.mjs`，对首页结构、默认折叠态和 locale key 做 no-browser 回归保护。
- `2026-04-24` 已继续补齐首页浏览器级主证据：
  - 首页总览、主网格、令牌明细、图表、排行和活动热力图现已补齐稳定 `data-testid` 锚点，便于 browser smoke 直接验证真实结构，而不是依赖文案猜 DOM。
  - 已新增 `scripts/verify-home-layout-browser-smoke.ps1`，并复用宿主已验证通过的 Edge CDP 包装器链路，通过 `home-layout` 场景在本机拉起最小运行数据后验证桌面端布局、默认折叠的运行摘要、展开后的三张 runtime 卡片，以及 `375px` 宽度下无大幅横向溢出。
  - 因此首页当前不再停留在“只有 no-browser 结构闭环”，而是已经具备真实浏览器级布局证据。

#### 未完成字段（对应 `9.5.1`）

- 严格意义上的“完整账单级 probe cost”仍未完成
- 当前熔断/探测摘要仍是 runtime recent-window 观测视图，不是持久化历史报表

补充说明：

- 当前首页价格相关能力是“估算视图”，不是完整账单统计。
- 之所以先标为 `estimated`，是因为当前链路还未纳入完整 cache usage 口径，也没有完整的 provider/channel 级官方价聚合。
- 当前首页 probe/circuit 区域也是“运行时观测摘要”，主要服务于可观测性，不应被理解为正式记账报表。

#### 当前判断

- 首页核心统计主线已部分完成。
- 目前已完成的是“总量 + provider/channel/model + 成功率/失败率”这一层。
- 未完成的多数项目已经越过纯前端能力边界，更依赖后端补数或新增聚合字段。
- `2026-04-28` 已按最新口径把主壳层、首页和分组页从“过窄/过密”方向回退到更接近原版的铺屏占比，并恢复了设置页双栏瀑布流与桌面端左侧中线导航停靠；对应 `tsc --noEmit` 与 `pnpm run test:screenshot-no-browser` 已重新通过。
- `2026-04-28` 同一轮又把首页主网格和主壳层验收基线同步更新：`AppContainer` 保持宽内容区但不再无节制撑满，桌面端 dock 锚定在左侧中线附近；首页主网格改为主内容列更宽、右栏更克制的比例，`scripts/verify-home-layout.mjs` 已同步到当前源码比例，避免旧 no-browser 脚本继续按过时栅格误判。
- `2026-04-28` 同一轮已补最新源码下的首页真实浏览器证据：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120 -CdpCommandTimeoutMs 15000` 通过，覆盖临时运行态、Edge CDP、桌面端布局、运行摘要默认折叠和 `375px` 无明显横向溢出。验证结束后已确认服务端口未被长期占用。

## 5. 前端主线完成度与后续优先级

### 完成度清单

- 渠道页：`已完成`
  - 差距说明：字段展示仍有深化空间，但 `9.1` 主线已闭环；`channel page` 当前也已补齐页面级浏览器证据，剩余缺口主要收敛到 `9.1.1` 里的更细 key 级可观测字段与创建/编辑弹窗细节交互。
- 分组页：`已完成`
  - 差距说明：移动端和超大列表体验仍可继续优化，但 `9.2 / 9.2.1 / 9.3` 核心目标已满足。
- 日志页：`已完成`
  - 差距说明：后续更多是体验增强，不是主线缺口。
- 首页：`已部分完成`
  - 差距说明：已补 token/provider/channel/model、成功率/失败率、estimated 官方/网关价格、estimated probe cost、熔断摘要、最近探测摘要，并完成桌面端比例、默认展开层级和浏览器级 `375px` 证据收口；当前剩余缺口主要是完整账单语义与历史持久化口径，已超出当前前端闭环范围。
- 设置页价格维护：`已部分完成`
  - 差距说明：已把价格维护卡片收口为“价格与同步节奏专属区”，显式说明探测/检测入口已迁到设置页模型探测策略，并补了 no-browser 边界验证；但浏览器级 smoke 与更细的同步反馈仍未闭环。
- 模型页 / 价格卡片布局：`已部分完成`
  - 差距说明：工具栏中的“普通 / 紧凑”布局现在已经真实作用到模型卡片层，普通与紧凑两档不再是只改容器不改内容；`verify-llm-price-boundary.mjs` 已补回归断言，防止后续再次出现“切换入口存在但卡片不变化”的假闭环。本轮又把普通布局中的 `规范名称 / 计费模式 / 官方价格` 收紧到统一 meta 信息带，并把搜索契约补齐到“模型名称 + 规范名称”双命中，避免输入提示与实际过滤行为不一致。当前剩余缺口主要收敛为浏览器级 `375px` 证据，以及是否还存在新的中文界面英文主显示泄漏。
- 设置页熔断：`已部分完成`
  - 差距说明：已继续收口为“摘要优先 + 短默认路径 + 恢复步骤按需展开 + 高级策略延后展开”的第二轮默认压缩路径：摘要卡片去掉首屏重复 helper 文本，改成数值 + 问号帮助提示；恢复流程改成独立折叠区，默认只保留一句主说明，三步细节仅在排障时展开；高级熔断策略标题下也只保留一行短说明。`scripts/verify-circuit-breaker-help.mjs`、`tsc --noEmit` 与 locale consistency 已同步通过。当前剩余缺口主要是浏览器级真实证据，以及是否还需要把同类 summary-first 压缩继续推广到同池其它设置卡片之外的更深层细节路径。
- 设置页模型探测策略：`已部分完成`
  - 差距说明：当前已重新对齐到“默认路径摘要 + 搜索入口 + 手动展开 + 卡内滚动 + 分批显示模型行”的默认路径：`defaultPath` 摘要、帮助提示、搜索框、折叠开关、卡内滚动区与空态/折叠态锚点都已恢复到统一契约，展开态默认先渲染前 `12` 个模型，继续展开时再按批次追加，避免设置页在本地环境下一次性铺开大量模型 Accordion；`scripts/verify-model-probe-help.mjs`、等价 `settings-no-browser` Node 守护链与 `build:static` / `web/out -> static/out` 同步已通过。当前剩余缺口主要是浏览器级真实证据，以及宿主 `pnpm` 自调用仍会因本机 `pnpm config rc EPERM` 与 PATH 解析问题阻塞统一脚本入口，但这不再是 `ModelProbe` 代码契约 blocker。
- 设置页动态路由学习轻入口：`已部分完成`
  - 差距说明：本轮已把设置页动态路由卡片继续收口成“轻量学习开关 + 三个摘要点 + 跳转 AI 自动化中心查看详情”的主路径，明确 settings 只负责 `dynamic_routing_learning_enabled` 的手动控制与边界说明，不再让学习样本明细、重置与复盘重新挤回设置页；`DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 以及 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口主要是浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM` 仍未解除，导致新增组件测试只能维持静态回归与代码审阅层验证。
  - `2026-04-28` 同主线又继续把 `AI 自动化` 学习区本身的状态边界补齐：学习区现在会显式展示“当前开关状态 / 当前样本数 / 当前是否参与推荐评分”，并根据“已启用但无样本 / 已关闭但保留样本 / 已关闭且无样本”给出不同说明；无样本时重置按钮会禁用并显示轻提示，避免出现可点但无效果的假动作。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、四语 locale 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
  - `2026-04-28` 同一 Phase H6 主线继续补齐学习区“首屏一眼可读”的摘要：学习区摘要卡现在除了状态、样本数和运行时影响外，还会直接显示最近采样时间与当前最高分对象；样本卡内部也补了最近采样时间，减少用户从 settings 跳到 `AI 自动化` 后还要先读整排样本卡片的成本。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、四语 locale 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍收敛为浏览器级证据与宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 settings 学习卡补成真正的轻量摘要闭环：卡片现已直接显示真实样本数、最近采样时间、当前最高分对象，并提供就地 `reset learning state` 入口；当没有样本时，重置按钮会禁用并显示短提示，从而和文档里的“查询入口 + 清空入口”口径对齐，同时仍保持详细样本与复盘留在 `AI 自动化` 页。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把两个学习入口的动作反馈收口到一致口径：顶层 `AI 自动化` 学习区的 reset 按钮现已补上与 settings 同步的成功/失败 toast，避免 settings 有动作反馈而顶层页面静默清空；对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 settings 学习卡的“当前是否参与推荐评分”收口到首屏摘要层：现在卡片会直接显示 runtime scoring 是否启用，因此用户即使不跳转 `AI 自动化` 页，也能先判断 learning 当前是“正在参与推荐”还是“已暂停参与评分”。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-dynamic-routing-help.mjs`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 settings 学习卡与顶层 `AI 自动化` 学习区的运行时数据来源对齐：settings 卡的 runtime scoring 摘要现在优先读取 `useDynamicRouteLearning().data.enabled`，不再只盯着本地 setting；同时 learning 开关保存成功后会主动 `learning.refetch()`，`reset learning` 也会额外失效 `ai/config` 与 `settings/list`，从而减少两个 consumer 在短刷新窗口里的状态不一致。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、`web/src/api/endpoints/ai-automation.ts` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把顶层 `AI 自动化` 学习开关自身的回退链收口到与 settings 一致：学习开关现在会优先采用草稿态，再看 `useDynamicRouteLearning().data.enabled`，最后回退 `ai/config.dynamic_routing_learning_enabled`，从而避免 learning 查询短空窗里把开关和运行时摘要闪成关闭；保存失败时也会清掉草稿并显示专用失败 toast。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、`scripts/verify-ai-automation-learning-focus.mjs` 与四语 locale 已同步更新，并通过 `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把顶层学习摘要与 settings 的 highest-score 语义完全对齐：`AI 自动化` 学习区里的“当前最高分对象”现在不再直接拿第一条样本，而是和 settings 一样按最高 `score` 选取；最近采样时间仍单独按最新 `last_sample_at` 计算，因此在样本顺序与评分不一致时两个 consumer 不会再给出不同结论。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍收敛为浏览器级 `375px / hover / focus` 证据、宿主 `vitest/esbuild spawn EPERM`，以及本机 loopback 初始化阻塞导致的 browser smoke 不可用。
  - `2026-04-28` 同一 Phase H6 主线继续把两个 consumer 的 learning 摘要推导抽成共享 helper：`AI 自动化` 学习区与 settings 学习卡现在共用 `learning-summary.ts` 中的 `latest sample / top target / has samples` 推导与采样时间格式化逻辑，不再各自保留一套 `reduce` 链。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning 展示状态派生也并入共享 helper：`learning-summary.ts` 现在除了 `latest sample / top target / has samples` 之外，还会统一产出 `sampleCount / canReset / runtimeKey / noticeState`，因此 `AI 自动化` 学习区与 settings 学习卡不再分别维护“样本数显示 / reset 按钮禁用 / enabled/disabled notice 分支”这组条件判断。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning 摘要展示骨架抽成共享面板：新增 `LearningSummaryPanel.tsx` 后，`AI 自动化` 学习区与 settings 学习卡已共用摘要卡栅格和 notice 区结构，只保留各自字段集合与动作区差异，减少后续 JSX 漂移。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`LearningSummaryPanel.tsx`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning footer 动作条抽成共享 action bar：`AI 自动化` 学习区与 settings 学习卡不再各自维护 reset/open 按钮区和禁用提示，而是统一通过 `LearningSummaryPanel.tsx` 中的共享 action bar 传入动作配置与提示文本；现有 locale、toast 语义和 settings 的 test id 保持不变。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`LearningSummaryPanel.tsx`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把两个 consumer 剩余的 learning 摘要 item 组装和 notice 选择并入共享 helper：`learning-summary.ts` 现已新增 `resolveLearningNoticeValue`、`formatLearningTopTarget` 与 `buildLearningSummaryItems`，因此 `AI 自动化` 学习区与 settings 学习卡不再分别维护 top target fallback、notice 文案分支和 summary items 列表；settings 只在共享 base items 上额外拼接 `scope / details` 两个差异字段。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning 摘要的 view-model 组装抽成共享入口：`learning-summary.ts` 现已新增 `buildLearningSummaryViewModel`，由它统一产出 `summary / display / latestSampleLabel / topTargetLabel / notice / items`，因此 `AI 自动化` 学习区与 settings 学习卡不再分别手写 latest sample、top target、notice 和 summary items 的整条组装链；同时本轮也明确不继续硬抽 sample card，保留“settings 轻摘要、AI 自动化页展示样本列表”的既有信息分层。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 settings 学习卡消费共享摘要项的方式从“按数组顺序取值”收口到“按语义 key 取值”：`learning-summary.ts` 现已新增 `indexLearningSummaryItems`，settings 卡会显式读取 `status / samples / runtime / latest-sample / top-target` 五个共享项，再额外拼接自身的 `scope / details` 差异字段，从而避免后续 shared items 新增或换序时静默打乱 settings 摘要结构。对应 `setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 settings 学习卡的摘要 section 组装也并入共享 helper：`learning-summary.ts` 现已新增 `buildLearningSummarySections`，统一把 `status / samples / runtime / latest-sample / top-target` 这些共享摘要项按 `primary / secondary` 两组返回，因此 `DynamicRouting.tsx` 不再自己维护 item 索引、缺项 guard 和 base section 组装，只保留 `scope / details` 两个 settings 专属字段。对应 `setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把默认摘要分组定义也收口到共享常量：`learning-summary.ts` 新增 `DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES` 后，顶层 `AI 自动化` 学习区也改为通过 `buildLearningSummarySections` 输出 `primary / secondary` 两组摘要栅格，不再直接把 `items` 整包喂给单一 grid；settings 学习卡则在同一套默认 section entry 上插入 `scope / details` 差异字段，减少两处残留的重复 key 列表与布局漂移风险。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`scripts/verify-ai-automation-learning-focus.mjs` 与 `scripts/verify-dynamic-routing-help.mjs` 已同步更新，并通过 `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM`。
  - `2026-04-28` 同一 Phase H6 主线继续把 `AI 自动化` 学习区 browser smoke 正式接入共享 CDP 骨架：learning 的 PowerShell 入口不再只走 `playwright-cli`，而是默认复用仓库现有的 `self-start/external + cdp` 包装链；共享 `verify-channel-create-browser-smoke-cdp.mjs` 已新增 `ai-learning` 场景，会在运行前种最小动态学习样本，并验证 learning 页的加载、preset、reset、switch 与 `375px` 宽度。与此同时，共享 CDP wrapper 也已补上与 CLI wrapper 一致的稳定 Node 解析，并把 Windows loopback service-provider 失败从“端口不可用”改成准确的 host blocker 说明。对应 `scripts/verify-channel-create-browser-smoke-cdp.mjs`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check`；当前真实 `self-start` 仍受宿主 loopback 初始化阻塞，但 smoke 入口与错误分类已收口到可继承状态。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 host-friendly external 复跑方式固化为 wrapper 级快捷入口：`verify-ai-automation-learning-browser-smoke.ps1` 新增 `-UseHostFriendlyExternalDefaults` 后，会自动套用 `9233 + relaxed + workspace-fixed + external bootstrap + 30000ms command timeout + 70s node timeout` 这一组已知更适合宿主对照的参数；同时该快捷入口不再默认打开本地后端/前端自启动，只有显式加 `-SelfStartServices`（或兼容别名 `-SelfStartLocalServices`）时才会恢复旧的 local service bootstrap 语义。这样在 loopback service-provider 已损坏的宿主上，learning smoke 不会再先死在本地端口探测阶段，而是能更快进入 external CDP 预检和更高层的失败分类；对应 `verify-ai-automation-learning-focus.mjs` 已补静态守护，当前 `check-only` 与真实 `external` 复跑都能体现这组最新入口语义。
  - `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 external 失败再向前收口成结构化 preflight：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已新增 backend / frontend / CDP 三段 reachability 预检，并把结果写入 `external-preflight-diagnostic.json`，因此 `verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 在当前宿主上已不再只报 `Timed out waiting for ...`，而是先明确把 `backend healthcheck` 归类为 `host_networking_blocker`。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`tsc --noEmit` 与 `git diff --check` 已同步通过；这让下一轮在健康宿主上更容易直接区分服务可达性问题和浏览器/CDP 问题。
  - `2026-04-28` 同一 Phase H6 主线继续把 external preflight 的失败表达从“首个失败点”收口成聚合诊断：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已新增 skipped entry、failed check 聚合、统一 hints 与 `schemaVersion = 2` 的 diagnostic JSON，因此当前宿主上的 `verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 会一次性报告 `backend + frontend` 的 `host_networking_blocker`，并把当前未要求的 CDP 项明确显示为 `skipped`。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`tsc --noEmit` 与 `git diff --check` 已同步通过；这让下一轮在健康宿主上能更直接地区分“服务未暴露”与“浏览器/CDP bootstrap”两类问题。
- `2026-04-28` 同一 Phase H6 主线继续把 aggregated external preflight 的诊断结果直接接到 learning wrapper 入口：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已把 `skippedChecks / primaryBlockingCheck / summaryLines / hints / checkDetails` 固化进诊断 JSON，而 `verify-ai-automation-learning-browser-smoke.ps1` 会在 external 失败时自动读取 `Diagnostic:` 路径并先打印一段稳定的 `Latest external preflight diagnostic` 摘要。因此当前宿主上的 learning smoke 已不再只是抛一长段失败消息，而是会先明确给出 `overallClassification=preflight_failed`、`failedChecks=backend, frontend`、`skippedChecks=cdp`、`primaryBlockingCheck=backend` 和 artifact 路径。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`runtime-win status`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按环境阻塞预期失败；这让下一轮沿同一入口继续排查时，不需要再手工翻临时 JSON。
- `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 external 失败从“有聚合诊断但还要自己拼复跑命令”收口到命令级 next-step 输出：`verify-ai-automation-learning-browser-smoke.ps1` 现在会在 `Latest external preflight diagnostic` 后继续打印 `External preflight next steps`，直接给出标准 `-Mode external -UseHostFriendlyExternalDefaults` 复跑命令，以及需要本机 local service 对照时的 `-SelfStartServices` 变体；`verify-ai-automation-learning-focus.mjs` 已同步补静态守护。当前宿主上的 real external 仍按预期卡在 `backend + frontend -> host_networking_blocker`，但下一轮已不需要再从 hints 或 JSON 手动推导入口命令。
- `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 external 诊断副本固定到 repo-local 路径：wrapper 在 external 失败时除了打印临时 artifact 路径外，还会把诊断 JSON 复制到 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`，并在 `check-only`/失败摘要里直接打印这条稳定副本路径。这样后续同主线复跑时，不需要先找回上一次的临时目录，也能直接从仓库内稳定位置继承上一轮诊断结果；`verify-ai-automation-learning-focus.mjs` 已同步守住这一输出契约。
- `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 `check-only` 入口补成真正可消费的 repo-local 诊断回放：wrapper 现在在 `check-only` 模式下若发现 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json` 已存在，会直接打印完整的 `Latest external preflight diagnostic` 摘要，而不是只给路径；若稳定副本缺失或损坏，也会明确提示下一步该先跑 external 种子还是直接检查 JSON。这样下一轮即使不再重跑 external，也能沿同一个 repo-local artifact 继续判断是 `host_networking_blocker` 还是后续 CDP/browser 问题；同时历史 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁残留已清理，减少工作区噪音。
- `2026-04-28` 同一 Phase H6 主线继续把 stable diagnostic 的 repo-local 回放补成“可判断新旧”的入口：当前 learning smoke wrapper 在 `check-only` 模式下除了回放 `failed/skipped checks`、summary lines、hints 与 next steps 之外，还会明确打印 `diagnostic source / checked at / diagnostic age`，并补一句“这只是最近一次 external 失败的保存结果，不是 live probe”的提示。这样同一条 H6 诊断链在不重跑 external 的情况下也更不容易误判旧 artifact；对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步对齐。
- `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 external CDP 预检入口显式化：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已新增 `RequireExternalCdpPreflight`，AI learning wrapper 也会透传该开关，并在 `check-only`/stable diagnostic 回放里直接打印 `External preflight CDP requirement` 与新的 `-RequireExternalCdpPreflight` next step。这样在健康宿主上如果需要第一时间拿到同时覆盖 backend/frontend/CDP 的 fresh diagnostic，就不必再依赖 `BootstrapExternalCdpSession` 的隐式分支语义；本轮通过 `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、以及两次 `check-only`（含显式 `-RequireExternalCdpPreflight`）验证了这条入口契约。当前 repo-local stable copy 仍是历史 `requireCdp=false` 诊断，因此真实 `requireCdp=true` 的 stable artifact 还需要在服务可达环境上补一次 external 复跑。
- `2026-04-28` 同一 Phase H6 主线继续把 stable diagnostic replay 提示补成“可比较本次命令预期”的口径：当前 `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 除了回放 stable copy 当时的 `External preflight CDP requirement` 外，还会额外显示 `External preflight CDP expectation for this invocation`；当本次命令显式带了 `-RequireExternalCdpPreflight`、但 stable copy 仍是旧的 `requireCdp=false` 结果时，会直接打印 `External preflight CDP mismatch note`，提示下一步应重跑 external 来刷新 stable copy，而不是继续把 `cdp skipped` 误解成当前 wrapper 没有要求 CDP。为了让这条前端主线的 `tsc --noEmit` 验证链恢复，本轮也顺手修复了两个已存在的语法断点：`web/src/components/modules/ai-automation/index.tsx` 里 profile 预览卡多余的 JSX 关闭结构，以及 `web/src/api/endpoints/ai-automation.ts` 中 `task artifacts/retry` 两条 endpoint 的模板字符串缺失反引号。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前前端侧剩余重点仍然是健康宿主上的真实 external/browser 证据，而不是 stable replay 或类型链回归。
- `2026-04-28` 同一 Phase H6 主线继续把 learning smoke 的 repo-local stable diagnostic 从“单一总副本”收口成“按 external CDP 预期分桶”的回放结构：wrapper 现在会在 external 失败时除了维护旧的 `latest-external-preflight-diagnostic.json` 外，还按 `requireCdp` 额外落一份 `latest-external-preflight-diagnostic-optional-cdp.json` 或 `latest-external-preflight-diagnostic-require-cdp.json`；`check-only` 则优先选择和本次命令 `requireCdp` 预期一致的副本，并通过 `External preflight stable copy note` 明确说明是命中匹配副本，还是因为缺少匹配副本而回退到最近可用副本。与此同时，`check-only` 在输出 stable preview 后会直接结束，不再继续误入共享 wrapper 去创建临时目录，因此这条 Phase H6 no-browser 诊断链现在能稳定承担“接手时快速判断当前 external 证据是否匹配本次命令”的职责。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke 的 stable replay 可见性补成“命中的副本之外，其它变体也一眼可见”的状态：wrapper 现在会在 `check-only` 输出中直接列出 `matching / alternate / legacy stable diagnostic copy status`，逐条说明每个 repo-local 诊断副本当前是 `missing`、`present but could not be parsed`，还是 `present and parsed (recorded with requireCdp=true|false)`，并标出哪一份是 `selected for preview`。这样前端这条 H6 no-browser 诊断链在不重跑 external 的情况下，也能直接判断当前是否仍缺 `requireCdp=true` 变体文件，而不需要再手动翻 `build/verify-ai-automation-learning/` 目录。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的后续动作收口成“当前调用专属 refresh profile”：wrapper 现在会在 `check-only` 回放里额外打印 `External preflight preferred refresh command`，并让 `next steps` 优先复用与本次调用一致的 external 命令，而不是只给一组固定模板。因此如果这次回放来自 `-RequireExternalCdpPreflight` 调用，输出会直接建议同样带 `-RequireExternalCdpPreflight` 的 external 复跑；若后续有人在 check-only 上显式带了自定义 URL、CDP bootstrap 或 timeout 参数，这些 profile 也会原样进入推荐命令，降低人工重拼命令造成的 H6 诊断漂移。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍是健康宿主上的真实 external/browser 证据，而不是 replay 后还要自己猜下一条命令。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的 freshness 判断从“只显示 age”收口成“直接显示 fresh/stale 结论”：wrapper 现在会在 `matching / alternate / legacy stable diagnostic copy status` 行中追加 `fresh against 24h threshold` 或 `stale against Nh threshold`，默认用 24 小时阈值判断 repo-local stable diagnostic 是否仍值得复用，也允许通过 `-StableDiagnosticFreshnessThresholdHours` 临时调小阈值来验证 stale 分支。这样这条前端 H6 no-browser 诊断链在真实 external/browser 仍受宿主阻塞时，也能先快速回答“当前副本够不够新，值不值得去健康宿主刷新”，而不是只给 age 让接手人自己换算。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的 freshness 对比再收口成“单条 note 可直接指导下一步”的口径：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight stable freshness note`，直接说明当前 selected preview 是否仍 fresh、是否已经 stale，以及当 `matching` 变体缺失时当前其实只是 `legacy fallback preview`，同时补一句 `alternate requirement-specific copy` 是否与它同样新鲜。这样在前端这条 H6 no-browser 诊断链里，即使不打开 JSON、也不逐条比对 `matching / alternate / legacy` 三行状态，接手人也能直接判断“现在可以先沿用当前 stable copy”还是“已经 stale，应该去健康宿主执行 preferred refresh command”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的副本比较再收口成“当前预览是不是 repo-local 最新可解析证据”的口径：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight stable freshest copy note`，直接说明当前 selected preview 是唯一可比较副本、已经是 freshest，还是只与另一份 `legacy / alternate` 副本同鲜度并列；若未来 selected preview 比其它 parseable copy 更旧，也会直接指出更晚的副本类型。这样这条前端 H6 no-browser 诊断链在不重跑 external 的情况下，也能先回答“当前 replay 代表的到底是不是仓库里最新一份 saved diagnostic”，而不需要再人工比较三行 `age`。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的多条提示再压缩成最终决策层：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight decision summary`，直接说明当前是“可以继续用 repo-local replay 做 blocked-host triage”，还是“当前 invocation 仍缺 requirement-specific fresh evidence，应去健康宿主或先暴露服务后执行 preferred refresh command”。这样前端这条 H6 no-browser 诊断链在不打开 JSON、不逐条读 `freshness note / freshest copy note / mismatch note` 的前提下，也能先给出本轮是否值得立刻刷新 external artifact 的结论。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 replay 解释成本。
- `2026-04-29` 同一 Phase H6 主线继续把 legacy fallback replay 与 alternate requirement-specific 旁证的关系压进最终决策层：当当前 invocation 仍缺 matching `requireCdp=true` 副本、只能回放 legacy fallback 时，`External preflight decision summary` 现在会直接补一句 repo-local 是否已经存在 opposite CDP expectation 下的 alternate requirement-specific copy，以及它与 selected legacy fallback 是同鲜、更新还是更旧。这样这条前端 H6 no-browser 诊断链在只看最终 summary 的前提下，也能直接知道“仓库里是否已经留有 requirement-specific 旁证”，而不需要再倒回 `stable freshness note` 手工读 alternate suffix。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke stable replay 的 repo-local 覆盖关系再收口成单条覆盖摘要：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight stable coverage note`，直接说明当前仓库是否同时拥有可解析的 `requireCdp=true / requireCdp=false` 稳定副本，还是只覆盖了其中一类、另一类仍缺失或不可解析。这样前端这条 H6 no-browser 诊断链在 blocked-host 场景下不需要再人工对照 `matching / alternate / legacy` 三行状态，也能先判断 repo-local 是否已经具备 requirement-specific 双变体证据。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`runtime-win.ps1 -Action status`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 coverage completeness 从“还要读一条额外 note”收口进最终 summary：wrapper 现在会让 `External preflight decision summary` 直接补出 repo-local stable coverage 是否 complete，以及当前 invocation 还缺哪类 `requireCdp` 变体。这样前端这条 H6 no-browser 诊断链在只看最终 summary 的前提下，也能先判断“当前 matching replay 虽然 fresh，但 coverage 仍 incomplete、还缺 parseable `requireCdp=true` variant”，不必再和 `stable coverage note` 交叉阅读。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 future “coverage complete” 的最终措辞也预先钉死：wrapper 现在在 repo-local 同时具备 parseable `requireCdp=true / requireCdp=false` 两类副本时，会让 `External preflight decision summary` 直接切到“Repo-local stable coverage is complete ... so only freshness and live reachability remain relevant now.”，从而把 blocked-host triage 的下一步判断彻底收口成“看 final summary 就知道现在还缺变体，还是已经只剩 freshness/live reachability”。本轮除了默认与 `-RequireExternalCdpPreflight` 两次 learning smoke `check-only`、`verify-ai-automation-learning-focus.mjs` 与 `tsc --noEmit` 之外，还通过 helper 级合成状态验证了这一 complete-coverage 分支，不需要往 repo-local stable 目录写入伪造 artifact。当前剩余缺口仍继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 stable replay 的最终消费入口收口成动作级输出：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight recommended action`，用一条 action-first 句子直接说明当前是“继续沿 matched repo-local replay 做 blocked-host triage”，还是“把 selected legacy fallback copy 只当 fallback-only context，并去健康宿主或先暴露服务后执行 preferred refresh command 补 requirement-specific artifact”。这样前端这条 H6 no-browser 诊断链在只看最终两条摘要时，也能同时拿到“结论”和“下一步动作”，不需要再从 `decision summary + next steps` 手工拼行动建议。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把上述动作建议再压成稳定分类：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight recommended action class`，用少量固定值区分当前 replay 的可用状态。最新 repo-local 结果中，默认 external profile 会输出 `matched_replay_ready`，而 `-RequireExternalCdpPreflight` 且仍缺 matching 副本时会输出 `fallback_only_refresh_required`。这样前端这条 H6 no-browser 诊断链在接手时可以先看分类标签，再决定是否需要通读完整 action 文案或直接去健康宿主补 `requireCdp=true` artifact。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only` 与 `tsc --noEmit` 已同步通过；当前剩余缺口仍继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据。
- `2026-04-29` 同一 Phase H6 主线继续把 learning smoke 的真实 `external + self-start + require-cdp` 失败也并入 repo-local replay 闭环：AI learning wrapper 现在会在共享 CDP smoke wrapper 只抛出 `CDP diagnostic file:` 的情况下，自动桥接生成 `external-preflight-diagnostic.cdp-bridge.json` 并发布到 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic-require-cdp.json`。因此本机最新一轮真实 external 复跑虽然仍按预期卡在 `page_bootstrap_timeout_attached_session`，但 `check-only -RequireExternalCdpPreflight` 已能直接回放 fresh `requireCdp=true` stable artifact，给出 `overallClassification=cdp_smoke_failed`、`failedChecks=cdp`、`backend/frontend reachable`、`coverage complete` 和当前 `attached-session + runtime-page-lifecycle` 的 page bootstrap 超时细节；默认 `check-only` 也已确认 repo-local 同时具备 parseable `requireCdp=false` 与 parseable `requireCdp=true` 两类 stable variant。这样这条前端 H6 no-browser/host-level 诊断链的剩余缺口已从“缺少 requirement-specific artifact”收敛为“是否继续刷新 optional profile 到同一运行层，以及如何处理当前宿主上的 CDP page bootstrap timeout”，不再需要回头解决 stable artifact 缺失。
- `2026-04-29` 同一 Phase H6 主线继续把 optional / require-cdp 两条 invocation profile 的运行层差异也压成直接可消费的摘要：wrapper 现在会在 `check-only` 摘要里额外打印 `External preflight invocation profile alignment note`，直接说明 matched replay 与 opposite replay 是否已经对齐到同一服务可达层。当前最新 repo-local 结果中，默认 `requireCdp=false` profile 仍停在 `preflight_failed; failed checks backend, frontend`，而 opposite `requireCdp=true` profile 已经进入 `cdp_smoke_failed; failed checks cdp`；因此这条前端 H6 no-browser 诊断链现在会直接指向“先刷新 optional external profile，让两条 invocation path 都达到 backend/frontend 已可达的同一层，再比较 attached-session CDP bootstrap timeout”，避免下一轮再手工对照两次 `check-only` 输出后才决定刷新顺序。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过。
- `2026-04-29` 同一 Phase H6 主线已把上述刷新顺序真正执行完，并把 replay 的动作层同步切到对齐后的主阻塞：真实 `external -UseHostFriendlyExternalDefaults -SelfStartServices` 已将默认 optional profile 刷新到 `cdp_smoke_failed; failed checks cdp`，使其与 `requireCdp=true` profile 同时落在 backend/frontend 已可达、只剩 attached-session CDP bootstrap 超时的同一运行层。wrapper 现在会在 `check-only` 摘要里输出新的 `External preflight recommended action class: aligned_cdp_bootstrap_focus`，并让 `decision summary / recommended action / invocation profile alignment note` 一致地指向“继续把 live rerun 聚焦在 attached-session CDP bootstrap 比较，而不是回到 service reachability 或 artifact coverage”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、一次真实 optional external refresh、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前这条前端 H6 no-browser 诊断链的剩余工作已明确收敛为 attached-session `Runtime.enable / Page.enable` timeout 的进一步比较，而不是 invocation profile 再次失配。
- `2026-04-29` 同一 Phase H6 主线继续把“比较 attached-session CDP bootstrap 时到底跑哪条 live rerun 命令”也收口成固定输出：wrapper 现在会在 `aligned_cdp_bootstrap_focus` 场景下额外打印 `External preflight CDP bootstrap comparison command`，直接给出与当前 invocation 同 profile、仅切换 `-CdpBootstrapCommandOrder` 的 external 命令。这样前端这条 H6 no-browser 诊断链在看到 `aligned_cdp_bootstrap_focus` 后，不需要再手工补 `-RequireExternalCdpPreflight`、`-UseHostFriendlyExternalDefaults` 或反向 bootstrap order；默认与 require-cdp 两条 profile 都能直接复制各自的比较命令去验证 `runtime-page-lifecycle` 与 `page-lifecycle-runtime` 在 attached-session 下是否仍同样卡在 `Page.enable / Runtime.enable`。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前这条主线的剩余工作因此更明确收敛为“是否需要实际执行 comparison command 做 host-level 对照”，而不是 replay 结果还要人工翻译成下一条命令。
- `2026-04-29` 同一 Phase H6 主线已继续把这条 comparison command 收口到“默认停驻宿主可直接复现”的层面：本轮先实跑默认 optional profile 的 comparison 命令，确认单独复制 `External preflight CDP bootstrap comparison command` 会先因为 frontend 未启动退回 `preflight_failed / failed checks frontend`；随后再用同 profile 追加 `-SelfStartServices` 的 live rerun，稳定复现到与 repo-local replay 同层级的 `cdp_smoke_failed / failed checks cdp`，并验证 `page-lifecycle-runtime` 顺序下 attached-session bootstrap 仍卡在 `Page.enable -> Page.setLifecycleEventsEnabled -> Runtime.enable`。因此 wrapper 现在会在 `check-only` 输出中额外提供 `External preflight self-start CDP bootstrap comparison command` 与 `External preflight self-start refresh command`，把“当前宿主默认停驻，需要 self-start 才能做同层级 CDP 对照”这一步直接固化成命令，而不是让下一轮再手工推断。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、一次按预期失败的 non-self-start external comparison、一次按预期失败并落到 `cdp_smoke_failed` 的 self-start external comparison、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前这条前端 H6 no-browser/host-level 诊断链的下一步更适合直接复制 self-start comparison command 做 like-for-like 对照，若仍无差异，再决定是否继续放大 `-CdpCommandTimeoutMs`。
- `2026-04-29` 同一 Phase H6 主线已继续把“若 bootstrap order 仍无差异，下一条 timeout 对照命令该怎么拼”也收口成固定出口：本轮在默认 optional 与 `-RequireExternalCdpPreflight` 两条 profile 下真实执行 `self-start + page-lifecycle-runtime` 对照，确认反向 bootstrap order 不会改变 attached-session 的失败等级，仍稳定停在 `page_bootstrap_timeout_attached_session / Page.enable -> Page.setLifecycleEventsEnabled -> Runtime.enable`。因此 wrapper 现在会在 `aligned_cdp_bootstrap_focus` 场景下额外打印 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command`，统一给出 `-CdpCommandTimeoutMs '45000'` 与自动换算后的 `-NodeSmokeTimeoutSeconds '155'` 组合，让下一轮可以直接做最后一轮 bounded timeout 对照，而不是再手工换算参数。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 self-start external comparison、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前这条前端 H6 no-browser/host-level 诊断链的剩余工作因此继续收敛为“是否还要实跑这条更长 timeout 对照命令”，而不是 bootstrap order 或 profile 对齐仍有缺口。
- `2026-04-29` 同一 Phase H6 主线已继续把这条 timeout lane 从“还要不要再试一次”压成“是否值得做最后一次再直接定性”的层级：本轮分别真实执行了 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 45000ms` external rerun，确认两条 invocation profile 都仍停在同一 `page_bootstrap_timeout_attached_session / Runtime.enable -> Page.enable -> Page.setLifecycleEventsEnabled` 层。wrapper 现在会基于这份 fresh `45000ms` matching artifact，自动把 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command` 提升到下一档 `-CdpCommandTimeoutMs '60000' -NodeSmokeTimeoutSeconds '200'`，同时把 `External preflight recommended action class` 切到新的 `attached_session_bootstrap_blocker_candidate`。这样前端这条 H6 no-browser/host-level 诊断链在只看 `check-only` 摘要时，就能直接判断“当前宿主已经具备 attached-session bootstrap blocker 候选证据，只剩一轮 `60000ms` bounded timeout 对照是否还值得执行”，不需要再手工综合 live trace 与旧的 `45000ms` 提示。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 `45000ms` self-start external rerun、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs` 与 `tsc --noEmit` 已同步通过/按预期失败后落盘；当前这条主线的剩余工作因此进一步收敛为“最后一轮 `60000ms` host-level 对照”或“直接记录宿主级 blocker”，而不是 replay 解释层仍不够收口。
- `2026-04-29` 同一 Phase H6 主线已把这条 host-level timeout lane 正式收口为 blocker 结论：本轮分别真实执行了 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 60000ms` external rerun，确认两条 invocation profile 仍稳定停在同一 `page_bootstrap_timeout_attached_session / Runtime.enable -> Page.enable -> Page.setLifecycleEventsEnabled` 层，没有因为更长 timeout 或 profile 差异进入新的运行层。wrapper 现在会基于这份 fresh `60000ms` matching artifact，把 `External preflight recommended action class` 切到新的 `attached_session_bootstrap_blocker_confirmed`，并让 `decision summary / recommended action` 直接说明“当前宿主已确认是 attached-session bootstrap blocker，应停止继续调这条 wrapper 参数线并记录宿主级阻塞”。同时 confirmed 分支下不再继续输出更长 timeout 的 comparison command，避免前端这条 H6 no-browser/host-level 诊断链继续在同一 attached-session timeout lane 上空转。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 `60000ms` self-start external rerun、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前这条主线的下一步因此应切到不同执行路径或不同宿主，而不是继续堆高 timeout。
- `2026-04-29` 同一 Phase H6 主线继续把 confirmed blocker 后的交接出口固化成命令级产物：当前 `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 在 `attached_session_bootstrap_blocker_confirmed` 场景下会额外输出 `External preflight alternate execution path command` 与 `External preflight self-start alternate execution path command`，默认把下一条 live 对照压成“保持当前 invocation profile 与 `60000ms / 200s` 预算，只把 `-CdpPageBootstrapStrategy` 切到 `json-new`”的显式命令。这样前端这条 H6 no-browser/host-level 诊断链在确认 attached-session blocker 后，不再只会给出抽象的“切换执行路径”，而是直接给出可复制的下一条 page-strategy 对照入口，并且不会因为 `check-only` 默认参数退回 `30000ms` 而稀释当前 blocker 证据层级。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前最适合的下一步因此是先跑这条 `json-new` alternate execution path command，再决定是否需要换宿主继续取证。
- `2026-04-29` 同一 Phase H6 主线继续把 `attached-session` invocation 对 `json-new` repo-local 证据的解释层收口到不误导下一轮：当前 strategy-specific `attached-session` stable copy 仍缺失时，`check-only` 现在不会再简单把已存在的 `latest-external-preflight-diagnostic-{optional-cdp|require-cdp}.json` 当成 generic fallback 后继续催促 refresh matching copy，而是会在 `matching-generic / alternate-generic stable diagnostic copy status` 里明确标出 “saved with alternate page-bootstrap strategy 'json-new'”，并把 `recommended action class` 切到 `fallback_replay_ready`。这样这条前端 H6 no-browser 诊断链在看到 `attached-session` strategy-specific 副本缺失、但 `json-new` 同一 `requireCdp` 证据已存在时，会直接建议“沿 `json-new` same-expectation evidence 继续 blocked-host triage，优先执行 `alternate execution path command`，只有在确实需要 `attached-session` saved diagnostic 时才再 refresh”。同时 alternate execution 命令也会继承 repo-local `json-new` 证据里的 `60000ms / 200s` 预算，不再回落到默认 `30000ms / 110s`。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前这条主线的下一步因此更明确收敛为“直接用现成 `json-new` 命令继续取证或切宿主”，而不是回到 strategy-mismatch 场景里的错误 refresh 动作。
- `2026-04-29` 同一 Phase H6 主线已继续把这条 `json-new` lane 收口成 strategy-aware stable replay：learning smoke stable artifact 现在会同时区分 `requireCdp` 与 `page-bootstrap strategy`，落成 `latest-external-preflight-diagnostic-{require-cdp|optional-cdp}-{attached-session|json-new}.json`，并在 strategy-specific 副本缺失时明确退回 `same-expectation fallback copy`，不再把 `json-new` 副本误判成 `attached-session` invocation 的 matching requirement-specific evidence。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local parser、四次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`；当前 `attached-session` invocation 的 remaining gap 已明确变成“缺少 strategy-specific attached-session stable copy”，而不是 verifier 继续混淆错误层级。
- `2026-04-29` 同一 Phase H6 主线还继续把 `json-new` 的 host-level 结论从“只是 alternate path”收口成“新的 confirmed blocker”：本轮真实执行 optional `json-new + self-start + 60000ms` external rerun 后，repo-local 已补齐 `latest-external-preflight-diagnostic-optional-cdp-json-new.json`，因此 `json-new` 下的 optional / require-cdp 双 profile 都能直接被 `check-only` strategy-specific 回放。当前 `check-only -CdpPageBootstrapStrategy 'json-new'` 已会输出 `External preflight recommended action class: page_bootstrap_strategy_blocker_confirmed`，并明确说明两条 invocation profile 都会在本机稳定停在 `page_bootstrap_timeout_json_new`；对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、一条真实 optional `json-new` external rerun、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘。当前这条前端 H6 no-browser/host-level 诊断链因此不再建议继续在本机放大 `json-new` timeout 或顺序参数，而是优先切换宿主或真正不同的执行路径。
- `2026-04-29` 同一 Phase H6 主线又继续把 `fallback_replay_ready` 场景里的 fresh/stale 提示收口到“适配当前 invocation”的语义：当 `attached-session` invocation 只能回放 `same-expectation fallback copy`，但仓库里存在更新一些的 `opposite-expectation` 或不同 page-bootstrap strategy 副本时，wrapper 现在会在 `stable freshest copy note`、`decision summary` 和 `recommended action` 中明确说明这些 fresher 副本只是 comparison-only，不应替换当前已选中的 same-expectation replay。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过默认/`-RequireExternalCdpPreflight` 两次 `check-only`、focus 守护与 `tsc --noEmit`；当前前端这条 H6 no-browser/host-level 诊断链的 handoff 口径因此进一步收敛为“继续用 same-expectation replay 或直接执行 alternate execution path command”，而不是因为 fresher opposite-expectation artifact 再把下一轮带回错误 refresh 分支。
- `2026-04-29` 同一 Phase H6 主线继续把 blocked-host handoff 从“只给出下一条命令”收口成“直接标记当前宿主该停手”：当前 `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 除了 `recommended action / alternate execution path command` 之外，还会额外输出 `External preflight host handoff note`。这条 note 会在 `attached_session_bootstrap_blocker_confirmed`、`page_bootstrap_strategy_blocker_confirmed` 以及当前 `fallback_replay_ready + json-new same-expectation fallback` 场景下，直接说明当前宿主是否已经被 repo-local diagnostics 证明为 host-level blocker。对于当前仓库里已存在的 `json-new` 双 profile stable 副本，它会明确提示“本机对 `json-new` 也稳定停在 `page_bootstrap_timeout_json_new`，因此 alternate execution path command 只值得带到别的宿主上复跑，在本机继续 live rerun 基本不会新增证据”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 `check-only`、PowerShell parser 校验、focus 守护、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前这条前端 H6 no-browser/host-level 诊断链的下一步因此进一步稳定为“换宿主或换真正不同路径”，而不是重新尝试同一台机器上的 `json-new` live 命令。
- `2026-05-01` 同一 Phase G browser smoke reliability 子线继续收口了本机运行态探测的低权限 gap：`scripts/runtime-win.ps1 -Action status` 现在在 `CIM denied` 且 `Get-NetTCPConnection` 不可用或不给 owning process 时，会继续回退到 `netstat -ano -p tcp`，从而在低权限宿主上仍尽量给出监听端口的 PID / 进程名提示。最新 repo-local 验证显示，定向探测 `8588 / 9210` 时已能明确看到 `HiviewService` 与 `QQ`，同时输出里会额外说明这些 `port owner hints` 只用于判断当前宿主应走 `external` 还是 `self-start`，不会被 `stop` 当成 `octopus_repo` 进程误停。因此下一轮若再遇到 `3000 / 8080` 类监听占位，主线程可以先用 `status` 判断其是否明显属于外部程序，再决定 live smoke 入口，而不必继续把“owning process unavailable”当成固定 blocker。
- 设置页四卡片帮助提示浏览器 smoke：`CLI self-start 已完成，CDP 宿主 blocker 保留`
  - 本轮补充：wrapper 现在会在 `external + cdp` 模式下先预检 `CDP /json/version`，因此即使还没先拉起后端和前端，也能稳定复现“外部 CDP 端点缺失”的受控失败；同时报错里的 `remote-debugging-port` 提示会跟随自定义 `CdpUrl` 端口，避免下一轮复现时拿错端口命令。
  - 额外收口：`scripts/verify-setting-help-browser-smoke-cdp.mjs` 现在会输出 `cdp.diagnostic.json`，而 PowerShell wrapper 在 Node 进程先被总超时截断时，也会从 trace tail 兜底生成最小分类摘要。因此当前 `self-start + cdp` 的失败路径已经能直接给出 `page_bootstrap_timeout_preempted / CdpPageBootstrapPendingTimeout / pageMode=json-new`，不再需要先人工翻 trace 才知道卡在哪一层。
  - 本轮继续补了显式 `CDP page bootstrap strategy` 对照能力；`auto / json-new / attached-session` 现在可从 wrapper 透传到 Node smoke。最新对照结果表明，即使跳过 `json/new`、直接走 `attached-session`，当前宿主上的 Edge/CDP 仍会在 `Page.enable` 阶段超时，且 wrapper 已能输出 `page_bootstrap_timeout_attached_session / CdpPageBootstrapPendingTimeout / pageMode=attached-session / pageStrategy=attached-session` 的结构化摘要。
  - 本轮收束：`attached-session timeout + bootstrap order` 对比已结束，`30000ms` 更长命令超时和 `runtime-page-lifecycle / page-lifecycle-runtime` 顺序切换都没有解除本机 Edge/CDP 的 page bootstrap 卡点；该阻塞已正式定性为宿主机级 `attached-session` bootstrap 问题。下一步不再继续调参，直接转向同一 screenshot-first 池的下一项（`CC Switch`、channel/group create dialogs、`375px`、help-hint hover/focus）。
  - 本轮再补一层 no-browser 护栏：`HelpHint` 的默认 aria-label 已改成通过 locale 读取，不再把中文“查看帮助”硬编码到所有语言；同时帮助按钮与 tooltip 内容新增稳定 `data-slot`，设置页浏览器 smoke 与备份页 no-browser 验证入口改为优先依赖稳定 selector，而不是依赖某一种语言的按钮文案。
  - `2026-04-23` 继续把该浏览器 smoke 从“合成 hover / 全局 tooltip 文本”收紧为真实浏览器路径：`HelpHint` 触发器与内容共享稳定 `data-help-hint-id`，CLI smoke 通过键盘 `Tab` 聚焦目标按钮、通过 `@playwright/cli hover` 真实悬停目标按钮，并只接受当前 trigger 绑定的 tooltip 文本。PowerShell wrapper 也会强制检查 Node stdout 成功标记并拦截 error-like stderr，避免再次出现“stderr 有异常但 wrapper 误报 passed”的假阳性。
  - 最新 `cli self-start` 证据已通过：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`，结果为 `setting-help-browser-smoke passed`、`desktopHelpButtons: 21`、`interactionChecks: 4`、`mobileWidth: 375`、frontend/backend 均为 `http://127.0.0.1:18081`。因此 settings 四卡片 help-hint 的 `375px + keyboard focus + real hover` 浏览器证据已关闭；剩余 CDP 项仍按宿主级 Edge/CDP bootstrap blocker 单独记录，不再阻塞 CLI 主证据。
- `CC Switch`：`浏览器级主证据已恢复，CLI 路径仍保留宿主诊断项`
  - 差距说明：已收口同页四步进度、主模型/名称/Claude 进阶映射的渐进解锁、无可用分组时的模型锁定提示，以及导入按钮的分步阻塞提示；`scripts/verify-ccswitch-flow.mjs` 已同步覆盖这些门控与四语 locale key。
  - `2026-04-24` 已继续补齐 `CC Switch` 的浏览器级主证据：`scripts/verify-ccswitch-browser-smoke.ps1` 现在默认把 CDP bootstrap 顺序固定为当前宿主已验证通过的 `runtime-page-lifecycle`，`scripts/verify-channel-create-browser-smoke-cdp.mjs` 也补上了 `CC Switch` 帮助提示的真实 `focus + hover` 断言。当前 `self-start + cdp` 已可在本机验证桌面端结构、帮助提示 tooltip、导入门控、以及 `375px` 宽度下无明显横向溢出。
  - 当前剩余缺口：`@playwright/cli` 仍受宿主 `Node spawn EPERM` 影响，个别自启动回合里 Edge remote debugging 端口也会出现启动波动；这些保留为宿主诊断项，不再作为 `CC Switch` 主浏览器证据的阻塞条件。
- 设置页备份 / 导入：`已部分完成`
  - `2026-04-25` 已补齐 backup locale provider 的类型收口：`locale.tsx` 不再是备份页主线 blocker，当前四语 locale 结构已与现有 AI 自动化 UI 契约对齐，并通过 `tsc --noEmit` 与 `verify-locale-consistency.mjs`。
  - 差距说明：已收口为“项目级快照导出 + dry-run 预检 + `incremental / map / merge / replace / skip` 导入模式 + 结构化兼容性报告 + 回滚 preview”真实主流程，并补齐 manifest、影响范围、结果摘要、风险提示、建议映射与当前未支持能力说明；但这仍只代表备份主链可用，不代表高级迁移 fully done。更强的 conflict wizard、完整 mapping editor、交互式 diff 与更细粒度 partial restore 仍未闭环。
  - 备份页导入结果区的 rows badge 已收口为 `Dry-run rows` / `Applied rows`，并在 verification script 中补了 dry-run、apply、latest rollback refresh 的断言。
  - 备份页的 `Apply This Dry-Run` 结果区现在还会显式展示 captured preview token，并把 dry-run binding 说明同步到 preview token 级别。
  - 备份页的 `Apply This Dry-Run` 元数据标签已继续收紧为 `captured apply mode` 与 `captured model mappings`，并同步到 Vitest 和 no-browser verification script。
  - 备份页底部的 `Advanced migration tooling still pending` 区域已改成三栏分组结构，按 Import tooling / Rollback tooling / Route analysis 收纳剩余缺口，避免长句堆叠。
  - 备份页历史回滚区域已补齐 latest rollback 的 pending 互斥，rollback 进行中时会锁住 refresh / preview / latest rollback 入口，避免并发操作窗口。
  - 备份页历史回滚区域进一步把 `importSnapshots.isFetching` 纳入统一 busy 锁，snapshot history 刷新中同样会锁住 refresh / preview / latest rollback / selected rollback 入口，避免历史列表刷新与回滚操作交错。
  - 备份页现在还会把空的 selective rollback scope 视为 full snapshot restore，避免 selectiveImport 开着但没有任何 active scopes 时提交空 scope 对象。
  - 备份页现在还会在 selective import 开着但没有 active scopes 时直接禁用导入按钮，并提示必须先选 scope，避免把空选择提交到 import 路径。
  - 备份页兼容性详情里残留的 `缺失渠道 / Provider` 已改回 `缺失渠道 / 供应商`，并在 no-browser verification script 中补了中文主显示断言，避免中文界面再次泄漏英文主文案。
  - 备份页底部“高级迁移能力待补”区域中的 `replace/map`、`remap` 已从简体/繁体中文主显示中清走，统一改成“替换导入 / 映射导入”与“快照模型=当前模型（目前模型）”语义，并在 `backup-logic` 与组件级 no-browser 验证中补了显式断言。
  - 备份页导出、导入、替换清理、历史回滚和高级待补区域现已补齐 `HelpHint` 可见入口；对应 `Backup.tsx` 中损坏的历史乱码帮助文本块也已清理，当前 no-browser 回归基线为默认 8 个帮助按钮、`map` 模式 9 个帮助按钮、历史回滚展开 9 个帮助按钮。
  - `HelpHint` 现在会按当前语言返回无障碍标签，备份页 no-browser 验证也已从依赖中文按钮名切到兼容稳定 selector 的路径，避免四语切换后帮助提示验证链再次漂移。
  - `2026-04-24` 已继续把 map mode 的兼容性详细诊断验收口径收紧到最新 UI 契约：默认只显示“详细诊断默认折叠，按需展开查看”，测试与 no-browser script 都先断言默认折叠，再通过兼容 `Show / 展开 N` 的按钮选择器展开详情，随后检查 `Model mapping previews / Missing mapping targets / Unused model mappings` 等 detail list。`scripts/verify-backup-component.cjs`、`scripts/verify-backup-logic.mjs` 与 `tsc --noEmit` 现已同步通过。
  - 受宿主环境影响，`Backup.test.tsx` 对应的 `vitest/vite/esbuild` 入口在本机仍会因 `spawn EPERM` 卡在 config startup；这已被记录为环境 blocker，而不是 Backup 组件回归。
  - `2026-04-30` 又继续把 backup 页 CLI browser smoke 的失败语义收口到共享 wrapper：当前 `verify-backup-browser-smoke.ps1 -Mode self-start` 若在本机命中 Playwright CLI 子进程 `spawn EPERM`，会直接输出明确的 host-level blocker，而不再误报成“channel create smoke 未写出 success marker”。这说明 backup 页面当前剩余 browser gap 仍是宿主执行环境问题，不是 backup 导入/回滚页面契约回退。

### 后续优先级

#### P1：直接影响 MD 闭环的缺口

- 主壳层铺屏占比、设置页双栏瀑布流、左侧 dock 停靠位与全局字体观感仍需继续结合真实浏览器截图做细节微调。
- 最新一轮代码已完成铺屏占比、左侧中线 dock、首页比例和渠道密度的 no-browser 基线更新；下一步必须用真实浏览器补桌面端与 `375px` 的截图级确认，尤其看是否还有字符超限、模块出框、hover/focus 焦点不清和折叠动画手感问题。
- 首页与渠道页的最新源码真实浏览器主证据已通过；下一步的浏览器级重点转到渠道创建/编辑弹窗内部的多 Key 折叠 hover/focus 手感、设置页更深层弹窗、模型页普通/紧凑切换细节，以及其它二级页面的文字统一性排查。
- `2026-04-28` 又继续收口了渠道多 Key 折叠头的交互噪音：公共 `AccordionTrigger` 已支持按场景关闭默认指示器，渠道创建/编辑弹窗中的多 Key 头部现只保留自定义状态箭头，不再叠加 Radix 默认箭头，从而避免折叠头出现重复展开暗示、桌面端视觉发虚的问题。对应 `tsc --noEmit`、`verify-channel-create-flow.mjs` 与 `verify-locale-consistency.mjs` 已通过。
- 同一轮还增强了共享浏览器 smoke 包装器对 Windows 宿主抖动的容错：`scripts/verify-channel-create-browser-smoke-cdp.ps1` 现在会对 loopback 端口探测中的 service-provider 异常做短重试，并把“端口空闲但 PowerShell 当前会话 bind 抖动”与“端口真被占用”区分开来；后端自启动也会在遇到该类瞬时错误时换 fresh config/port 重试。当前这条链路在本机会继续被更底层的 Windows socket provider 初始化故障拦住，因此它被记录为宿主级 blocker，而不是渠道创建 UI 本身的回归。
- `2026-04-30` 又继续收口了同池共享 CDP wrapper 的一个更高优先级缺口：当前 `channel-create` / `group-create` / `settings help` 等 PowerShell smoke wrapper 已统一改用稳定的日志读取函数，并把 stderr 的 error-like 判定从少数内建 `Error` 名扩到自定义 `*Error` / `*Exception`。直接原因是本机 `verify-channel-create-browser-smoke.ps1 -Mode self-start` 先前会在 Node 子进程实际产出 `CdpPageBootstrapUnavailableError`、`cdp.diagnostic.json` 和 trace tail 的情况下仍误报 passed；修复后同一命令会正确失败并明确归类为 `page_bootstrap_timeout_attached_session` / `CdpPageBootstrapUnavailableError`，因此本轮结论是“共享验证链假阳性已修掉，`channel-create` 当前仍是宿主级 CDP bootstrap blocker”，而不再是页面级通过证据。
- `2026-04-30` 同一 Phase G 主线又补齐了另一处验证入口漂移：`verify-home-layout-browser-smoke.ps1`、`verify-model-layout-browser-smoke.ps1`、`verify-channel-page-browser-smoke.ps1` 与 `verify-ccswitch-browser-smoke.ps1` 现在都会把 `-RequireExternalCdpPreflight` 透传给共享 `verify-channel-create-browser-smoke-cdp.ps1`。这样这些 page-level CDP wrapper 的 `check-only` / external 入口就能和 learning / shared wrapper 一样显式声明“首轮 external preflight 是否必须先检查 CDP reachability”，避免同池脚本继续出现“有的页面能拿到结构化 external CDP 预检口径，有的页面只能隐式复用旧默认值”的分裂。
- `2026-04-30` 同一 Phase G 主线随后继续把上一条里剩下的两份复制版 wrapper 也收口到共享 helper：`verify-group-create-browser-smoke-cdp.ps1` 已改成直接 forward 到共享 `verify-channel-create-browser-smoke-cdp.ps1`，`verify-group-create-browser-smoke.ps1` 也同步补齐 `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices` 这些透传参数；`verify-setting-help-browser-smoke.ps1` 则把 `Driver=cdp` 路径前移到共享 wrapper，保留既有 CLI 路径不变。直接结果是 `group-create-cdp` 与 `settings-help` 的 `check-only -RequireExternalCdpPreflight` 现在也会输出统一的 `Explicit external CDP preflight requirement: enabled` / `External mode initial CDP preflight: required` 口径，因此当前这条验证主线的剩余工作已从“入口契约分裂”收敛回“健康宿主上的真实 browser/CDP 证据”和现有 host blocker 分类，而不是继续维护多份分叉 wrapper。
- `2026-04-30` 同一 Phase G 主线又继续补齐了剩余的一处 CLI 入口分叉：`verify-group-create-browser-smoke.ps1` 现在不再保留自复制的 CLI 自启动逻辑，而是像 `verify-backup-browser-smoke.ps1` / `verify-ccswitch-browser-smoke-cli.ps1` 一样，通过环境变量把 `group-create` 专属 `mjs` 脚本、success marker 与 label forward 给共享 `verify-channel-create-browser-smoke.ps1`。这样 `group-create` CLI 路径后续也会自动继承共享 wrapper 的 `spawn EPERM`、loopback 预检和 success-marker/error-like stderr 护栏，减少同池维护漂移。
- `2026-04-30` 同一 Phase G 主线继续把 `settings-help` 的 CLI 入口也收口到共享 `verify-channel-create-browser-smoke.ps1`：当前 `verify-setting-help-browser-smoke.ps1` 已不再维护自复制的 CLI 自启动 / 日志解析 / host-blocker 分类逻辑，而是像 `backup / ccswitch / group-create` 一样，通过环境变量 forward `scripts/verify-setting-help-browser-smoke.mjs`、success marker 与 label。直接结果是 `check-only -Driver cli` 继续通过，且本机 `self-start -Driver cli` 会稳定落到共享 wrapper 的 Playwright CLI `spawn EPERM` host blocker 口径，说明 settings 剩余的 CLI browser gap 也已收敛为宿主执行环境问题，而不是设置页 wrapper 本身的分叉实现。
- `2026-04-30` 同一 Phase G 主线继续把上述 shared-wrapper 收口补成 repo-local 守门：新增 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，并接入 `scripts/run-frontend-verification-suite.mjs screenshot` 与 `web/package.json` 的 `test:browser-smoke-wrapper-alignment`。当前 no-browser 守卫会固定检查 `backup / ccswitch / group-create / settings-help` 的 CLI wrapper 仍是薄 forwarder，`group-create / ccswitch / home-layout / model-layout / channel-page / settings-help` 的 CDP wrapper 仍正确指向共享 `verify-channel-create-browser-smoke-cdp.ps1`，且这些入口没有重新长回 `Wait-HttpOk / Start-LocalOctopusBackend / Assert-NodeSmokeSucceeded` 这类复制版 helper。直接结果是最近几轮已经完成的 wrapper 收口，不再只靠人工盘点维持；只要同池入口再次回退成复制版脚本，`test:screenshot-no-browser` 就会先在 repo-local 失败。
- `2026-04-30` 同一 Phase G 主线继续把 thin forwarder 的用户参数契约也收口到统一口径：`verify-group-create-browser-smoke-cdp.ps1` 与 `verify-setting-help-browser-smoke.ps1` 现在都改为公开 `-Browser` 参数，并保留 `-BrowserPath` 作为兼容 alias，再统一 forward 到共享 `verify-channel-create-browser-smoke-cdp.ps1` / `verify-channel-create-browser-smoke.ps1`。对应 `verify-browser-smoke-wrapper-alignment.mjs` 也新增了对 `[Alias('BrowserPath')]`、`[string]$Browser` 与 `Browser -> BrowserPath` 透传的断言，因此当前 page-level wrapper 家族不再出现“只有个别页面要求 `-BrowserPath`，其它页面要求 `-Browser`”的调用分裂；下一轮可以继续直接沿同一主线补健康宿主上的真实 browser/CDP 证据，而不需要先回忆单页参数差异。
- `2026-04-30` 同一 Phase G 主线随后又把这条参数统一口径补到了剩余 page-level CDP thin forwarder：`verify-home-layout-browser-smoke.ps1`、`verify-model-layout-browser-smoke.ps1`、`verify-channel-page-browser-smoke.ps1` 与 `verify-ccswitch-browser-smoke.ps1` 现在也都公开 `-Browser`，并保留 `[Alias('BrowserPath')]` 兼容旧命令，再统一 forward 到共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 `BrowserPath`。`verify-browser-smoke-wrapper-alignment.mjs` 也同步把这四条入口纳入 alias/透传守门，因此当前同池 page-level CDP wrapper 家族已经统一到一套公开浏览器参数口径；下一轮可以继续把精力留在健康宿主上的真实 browser/CDP 证据，而不是继续排查单页调用差异。
- `2026-04-30` 同一 Phase G 主线继续把这套浏览器参数口径补齐到 CLI thin forwarder：`verify-backup-browser-smoke.ps1`、`verify-ccswitch-browser-smoke-cli.ps1` 与 `verify-group-create-browser-smoke.ps1` 现在也都公开 `-Browser`，并保留 `[Alias('BrowserPath')]` 兼容旧命令，再分别 forward 到共享 `verify-channel-create-browser-smoke.ps1` 的 `Browser` 或 `BrowserPath` 参数。`verify-browser-smoke-wrapper-alignment.mjs` 已同步把这三条 CLI 入口纳入 alias/透传守门，而且本轮还直接用旧 `-BrowserPath` 调用跑通了三条 `check-only`，说明当前同池 browser smoke wrapper 家族的公开浏览器参数口径已经在 CLI 与 CDP 两侧同时收敛；下一轮可以继续留在真实 browser/CDP 证据与宿主 blocker 分类，而不用再处理参数面分裂。
- `2026-04-30` 同一 Phase G 主线又把 wrapper 守门从“检查几个 thin forwarder 还在”补到“检查整套 browser smoke PowerShell 家族没有无声漂移”：`scripts/verify-browser-smoke-wrapper-alignment.mjs` 现在会固定枚举全部 `verify-*-browser-smoke*.ps1` 清单，要求它们只能落在 `thin forwarders / shared roots / specialized root(ai-learning)` 三类已知族谱里；同时除了继续检查 `backup / ccswitch / group-create / settings-help / home / model / channel-page` 这些 thin forwarder 的共享引用、alias 与透传外，还会额外断言共享根包装器 `verify-channel-create-browser-smoke.ps1` 与 `verify-channel-create-browser-smoke-cdp.ps1` 仍保留 Codex bundled Node 过滤、环境变量驱动的 scenario/success-marker 入口，以及 `Browser -> BrowserPath` 的公共契约。`verify-ai-automation-learning-browser-smoke.ps1` 也被显式标记为受控特例：它允许保留 host-handoff 与 stable-diagnostic 逻辑，但必须继续 forward 到共享 CDP wrapper，而不是重新长出另一套独立 browser smoke 栈。直接结果是后续若新增/回退同池 wrapper 而未同步 guard，`test:screenshot-no-browser` 会先失败；当前主线的剩余工作因此继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 wrapper 族谱失控。
- `2026-04-30` 同一 Phase G 主线继续把这套浏览器参数统一口径收口到共享根 wrapper：`verify-channel-create-browser-smoke.ps1` 现已和同池 thin forwarder 一样显式接受 `-BrowserPath` 兼容 alias，而 `verify-channel-create-browser-smoke-cdp.ps1` 也改为公开 `-Browser` 并保留 `[Alias('BrowserPath')]`，内部仍统一映射到底层 `BrowserPath`/Edge bootstrap helper。`verify-browser-smoke-wrapper-alignment.mjs` 已同步把这两条 shared root 的 alias 与 `Browser -> BrowserPath` 映射纳入 guard，本轮也直接用共享根的旧 `-BrowserPath` 和新 `-Browser` 两组 `check-only` 证明兼容与统一同时成立。这样当前 browser smoke PowerShell 家族不再剩“forwarder 已统一、shared root 还保留旧参数面”的分裂；下一轮可以继续把精力留在真实 browser/CDP 证据与宿主 blocker，而不用再处理共享根调用差异。
- `2026-04-30` 同一 Phase G 主线又继续把 shared-root `Driver=cdp` 的高级参数能力也补回统一根入口：`verify-channel-create-browser-smoke.ps1` 现在会把 `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices` 一并 forward 给共享 `verify-channel-create-browser-smoke-cdp.ps1`，不再只接受基础 `Browser`/端口参数。`verify-browser-smoke-wrapper-alignment.mjs` 也同步把这组 shared-root `cdp` 参数面纳入 guard，本轮已用统一根命令直接跑通 `-Driver cdp -RequireExternalCdpPreflight`、`-BootstrapExternalCdpSession -CdpPageBootstrapStrategy auto -CdpBootstrapCommandOrder runtime-page-lifecycle -CdpUrl http://127.0.0.1:9333` 和旧 `-BrowserPath` 三组 `check-only`。这样后续无论从 shared root 还是 page-level forwarder 进入 `cdp` smoke，都不会再出现“forwarder 支持 external-preflight / bootstrap 参数，但 shared root 顶层命令不认”的契约分裂；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 shared-root 参数透传缺口。
- `2026-04-30` 同一 Phase G 主线又继续把这条参数统一口径补到受控特例 `verify-ai-automation-learning-browser-smoke.ps1`：当前 `ai-learning` specialized root 也显式接受 `-Browser` 与兼容 `-BrowserPath`，并继续把浏览器路径映射到共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 `BrowserPath`。与此同时，`verify-browser-smoke-wrapper-alignment.mjs` 也不再只把它当作“能 forward 即可”的特例，而是额外锁定 `[Alias('BrowserPath')]`、`SelfStartServices / SelfStartLocalServices` 合流，以及 host-friendly 默认参数仍会启用 `BootstrapExternalCdpSession`。这样当前 browser smoke PowerShell 家族连 specialized root 也不再残留旧参数面分裂；下一轮可以继续直接聚焦真实 browser/CDP 证据与宿主 blocker，而不是回到 wrapper 契约补洞。
- `2026-04-30` 同一 Phase G 主线继续把 `ai-learning` specialized root 的 stable-diagnostic 预览合同收口成 repo-local guard：`verify-browser-smoke-wrapper-alignment.mjs` 现在会额外检查 `StableDiagnosticFreshnessThresholdHours`、timeout-comparison helper 和 `check-only` 里固定输出的 freshness threshold，因此这条受控特例后续即便仍能 forward 到共享 CDP wrapper，只要 stable-diagnostic 预览合同悄悄缩水，`test:screenshot-no-browser` 也会先失败。当前剩余工作继续留在真实 browser/CDP 证据与宿主 blocker，而不是 specialized root preview 漂移。
- `2026-04-30` 同一 Phase G 主线又继续把 page-level wrapper 家族里已经文档化的默认 CDP bootstrap 口径补成静态守门：`scripts/verify-browser-smoke-wrapper-alignment.mjs` 现在除了继续检查 alias/透传与 shared-wrapper 引用之外，还会固定断言 `verify-ccswitch-browser-smoke.ps1` 仍保留 `attached-session + runtime-page-lifecycle` 这组宿主已验证默认值，并锁定 `home-layout / model-layout / channel-page / group-create-cdp` 仍维持 `attached-session + page-lifecycle-runtime` 的 page-level 默认组合。这样后续如果有人只改了 thin forwarder 的默认 `CdpBootstrapCommandOrder` 或 `CdpPageBootstrapStrategy`，`test:screenshot-no-browser` 会先在 repo-local 失败，而不是等到下一轮再靠状态文档回忆哪一页本来就该走哪条顺序。当前主线的剩余工作因此继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 page-level 默认 bootstrap 契约无声漂移。
- `2026-04-30` 同一 Phase G 主线又继续把共享 CDP Node smoke 的执行入口文案也收口到真实入口：`scripts/verify-channel-create-browser-smoke-cdp.mjs` 的 `printUsage()` 先前仍错误引用 `verify-setting-help-browser-smoke-cdp.mjs`，本轮已改回共享真实入口 `verify-channel-create-browser-smoke-cdp.mjs`，并在 `scripts/verify-ai-automation-learning-focus.mjs` 中补上 usage 文案守门，明确要求共享 `ai-learning` 场景继续依附共享 CDP 脚本，而不是回退成错误的 page-specific 入口提示。对应 `verify-ai-automation-learning-focus.mjs`、`run-frontend-verification-suite.mjs settings` 与 `git diff --check` 已通过；当前主线剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是共享 usage 入口漂移。
- `2026-04-30` 同一 Phase G 主线随后又把 `settings-help` 这条 page-level CDP wrapper 已文档化的默认 bootstrap 组合补成静态守门：`scripts/verify-browser-smoke-wrapper-alignment.mjs` 现在会和 `ccswitch / home-layout / model-layout / channel-page / group-create-cdp` 一样，固定断言 `verify-setting-help-browser-smoke.ps1` 继续维持 `CdpPageBootstrapStrategy='auto'` 与 `CdpBootstrapCommandOrder='page-lifecycle-runtime'`。这样即使后续没有人立刻去跑 live smoke，只要 `settings-help` 的默认 bootstrap 被无声改动，`test:screenshot-no-browser` 就会先在 repo-local 失败，而不会等到下一轮再靠状态文档回忆这一页本来默认走哪条策略。
- `2026-05-01` 同一 Phase G 主线继续把共享 CLI wrapper 的宿主 blocker 口径补成 repo-local 守门：`scripts/verify-channel-create-browser-smoke.ps1` 现在在 Node 解析失败时也会跟随 `OCTOPUS_UI_SMOKE_LABEL` 输出当前页面标签，不再残留 `channel create smoke` 这类误导性硬编码；`scripts/verify-browser-smoke-wrapper-alignment.mjs` 则同步锁定了这条 label-aware Node 失败提示、既有 `spawn EPERM` host-blocker 分类，以及 `octopus-<label>-smoke-*` 临时目录命名。对应 `verify-browser-smoke-wrapper-alignment.mjs`、`verify-backup-browser-smoke.ps1 -Mode check-only`、`run-frontend-verification-suite.mjs screenshot` 与 `git diff --check` 已通过，因此当前 `backup / ccswitch / group-create / settings-help` 这组共享 CLI 入口后续若再回退成 `channel create` 专属报错或错误临时目录名，`test:screenshot-no-browser` 会先在 repo-local 失败；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是共享 CLI wrapper 语义漂移。
- `2026-05-01` 同一 Phase G 主线继续把“支撑层 drift 也要纳入 screenshot-first 守门”补齐到运行态与 AI learning 两侧：`scripts/verify-browser-smoke-wrapper-alignment.mjs` 现在除了继续锁 shared wrapper 家族之外，还会固定校验 `scripts/runtime-win.ps1` 仍保留 `Port scan mode`、loopback readiness、runtime entrypoint 提示和 `Automation entrypoints` 这组低权限 handoff contract；与此同时，活跃状态/workflow 文档里已经修过一次的 canonical/status 入口也被纳入同一守门，避免下一轮再次回指到不存在的根目录文档。本轮还顺手把 `scripts/verify-ai-automation-learning-focus.mjs` 对学习区焦点锚点的旧 `ref={learningSectionRef}` 断言更新到当前 `bindWorkbenchSection('learning')` 绑定实现，因此 `run-frontend-verification-suite.mjs screenshot` 在本机重新恢复全绿。这样当前 screenshot-first no-browser 主链的剩余缺口继续稳定收敛为真实 browser/CDP 证据与宿主 blocker，而不是 repo-local 文档入口或静态守门本身漂移。
- `2026-05-01` 同一 Phase G 主线继续把“被要求开工前优先阅读”的 archive 输入文档也纳入同一条 screenshot-first 守门：`USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、动态路由 planning/requirements、AI 自动化 requirements（中英）以及 `worklog/README.zh-CN.md` 先前仍保留一批指向不存在根目录 `docs/*.md` / `docs/*.zh-CN.md` 的旧入口，本轮已统一改回真实 `docs/archive/planning/`、`docs/archive/requirements/` 与 `docs/archive/status/` 路径，并在 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 新增对应 include/exclude 断言。对应 `run-frontend-verification-suite.mjs screenshot` 继续通过，说明当前 no-browser 接手入口已从“只守活跃状态文档”收口到“活跃状态文档 + 活跃 archive 输入文档”同一层；下一轮可继续直接聚焦真实 browser/CDP 证据与宿主 blocker，而不是 requirements/planning/worklog 入口再漂移。
- 设置页 `9.7`：继续向冲突处理、映射修正、回滚与部分恢复收口
- 渠道页 `9.1.1`：更完整的 key 级可观测字段仍缺一部分
- 中文统一性：继续清理中文路径里仍可能残留的英文主显示，重点关注设置页深层、弹窗和备份页细节态
- 模型/价格页：继续补更细的浏览器级普通/紧凑切换细节与剩余中英混杂排查，避免 page-level 主证据关闭后细部契约再次漂移

### 5.1 当前前端里程碑小结（2026-04-23）

- 里程碑 G1：渠道页主交互已进入“可用闭环 + 继续精修”状态。
  - 已实装同渠道多 Key 折叠交互、筛选增强、中文文案补齐，以及创建/编辑弹窗中的“两步式真实密钥输入引导”。
  - 当前已补齐渠道总页的浏览器级桌面端 / 筛选 / 详情弹层 / `375px` 主证据；后续重点收敛到 key 级观测字段、创建/编辑弹窗更细的 hover/focus 行为，以及桌面端比例精修。
- 里程碑 G2：分组页主交互已进入“主线闭环”状态。
  - 当前重点不再是基本可用，而是移动端与大列表体验打磨。
- 里程碑 G3：模型/价格页已从“伪布局切换”提升为“真实普通/紧凑双布局”。
  - 当前重点转向浏览器级细节验证与更多中文主显示清理。
- 里程碑 G4：首页已进入“结构重排 + 浏览器证据闭环”的状态。
  - 当前已完成首屏层级压缩、运行摘要折叠、Top 列表默认收缩、中文主显示清理，以及浏览器级桌面端 / `375px` 主证据。
  - 下一步重点不再是首页主布局证据，而是其它页面剩余英文清理和同池的浏览器级缺口。
- 里程碑 G4：备份页与帮助提示体系进入“主链逻辑和 no-browser 验证已过，settings 四卡片浏览器级 smoke 已完成，跨页面浏览器证据继续补齐”的状态。高级迁移工具、完整 mapping editor、交互式 diff 与更细粒度 partial restore 仍保持部分完成。

### 5.2 当前前端统一验证命令

- `pnpm run test:locale-consistency`
- `pnpm exec tsc --noEmit`
- `pnpm exec vitest run src/components/modules/channel/model-fetch.test.tsx`
- `pnpm run build:static`
- `pnpm run test:screenshot-no-browser`：统一覆盖 `locale consistency + Home / Channel / Group / Model / CC Switch / Route Target Overrides + settings no-browser` 的 screenshot-first no-browser 守护
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120 -CdpCommandTimeoutMs 15000`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120 -CdpCommandTimeoutMs 15000`

最近一次已通过的轻量验收组合：`pnpm run test:locale-consistency`、`pnpm exec tsc --noEmit`、`pnpm run test:screenshot-no-browser`、`pnpm exec vitest run src/components/modules/channel/model-fetch.test.tsx`。`build:static` 仍是发布/嵌入前必须跑的更重入口，不应被前述轻量组合替代。
最近一次已通过的浏览器主证据：`verify-home-layout-browser-smoke.ps1` 与 `verify-channel-page-browser-smoke.ps1` 均以 `self-start` 模式通过；执行后已通过 `scripts/runtime-win.ps1 stop` 确认本地项目保持停止状态。

#### P2：体验增强但不改语义

- 渠道页 key 级快速编辑进一步轻量化
- 分组页移动端三层嵌套可读性继续打磨
- 日志页导出快捷范围与摘要展示继续增强
- 模型/价格卡片继续补浏览器级普通/紧凑与 `375px` 证据；若发现新的中文界面英文主显示泄漏或新的搜索/筛选契约偏差，再按同池 no-browser 小闭环继续清理

#### P3：依赖后端补数的项目

- 首页价格统计
- 首页完整账单级 probe cost
- 首页持久化熔断 / 探测历史统计
- 渠道页更完整的 key 级观测字段（如额度、计费方式、探测策略）

### 前端不应越界改动的边界

- 不应擅自改变 group 提交结构
- 不应在前端重写路由语义
- 不应通过前端临时聚合伪装后端尚未提供的权威字段，除非文档中明确标注为“前端近似聚合视图”

---

## 6. 维护规则

- 以后每完成一项第 `9` 节前端主线增量，都要同步更新本文。
- 若实现需要偏离 canonical MD，必须先更新 canonical MD，再更新本文，再改代码。
- 若子 agent 参与了前端主线推进，应在相关 worklog 中记录其职责边界、模型和结论，本文只吸收最终结论，不记录中间对话。

---

## 7. 前端验收分层

### 当前自动 smoke 能覆盖什么

- 现有后端 smoke 已经能自动准备一套最小真实数据：
  - 登录
  - 创建渠道
  - 创建分组
  - 创建 API Key
  - 发送一条真实网关请求
- 因此它不仅能验证后端 API 主链路，也能为前端首页、渠道、分组、日志等界面提供最基本的非空数据背景。
- 当前自动 smoke 已覆盖：
  - `/healthz`
  - `/`
  - 当前 `web/src` 导出的静态壳
  - `/manifest.json`
  - 最小管理 API 主链路
  - Docker Compose 运行态启动
- `2026-04-21` 已把验证链路收口成“前端 typecheck + `pnpm run build:static` + backend build + smoke”，CI 会先从当前 `web/src` 生成 `web/out` 并同步到 `static/out`，不再只依赖可能陈旧的仓库静态产物。
- Linux / Windows smoke 现在都会优先选择较新的 `web/out` 或显式传入的 `OCTOPUS_SMOKE_STATIC_DIR`，从而让前端外壳校验更贴近当前源码导出结果。
- `2026-04-20` 已补通 Windows smoke 脚本链；当前环境下可以稳定走通健康检查、前端外壳、登录、渠道/分组/API Key 创建和一条真实网关聊天请求。
- 当前自动 smoke 的已知边界：本机可发现的 `go.exe` 仍返回 `Access is denied`，所以 Windows smoke 现在优先复用可运行的 `build/octopus-smoke.exe` 完成主链路验证；它能证明运行态和前端壳正常，但还不能证明“当前未提交 Go 源码可在这台机器上重新编译”。

### 自动 smoke 与手工 checklist 矩阵

仓库内正式手工验收入口：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`

- 首页：
  - 自动 smoke：服务启动、首页壳可访问、统计接口能被前端消费、不出现启动级 crash
  - 手工 checklist：图表细节、窄屏排版、统计口径、空态与加载态
- 渠道页：
  - 自动 smoke：最小 smoke 数据能生成可展示的 channel 数据源
  - 手工 checklist：卡片展开、key 搜索/折叠、单 key 测试、OAuth 特殊流程、长列表体验
- 分组页：
  - 自动 smoke：最小 smoke 数据能生成可展示的 group 数据源
  - 手工 checklist：拖拽排序、权重编辑、模式切换、移动端三层结构可读性
- 备份页：
  - 自动 smoke：设置页和备份区块能正常挂载
  - 手工 checklist：导出默认全量快照、显式脱敏导出、导入上传、dry-run、replace/merge/map/skip、建议映射、Apply Same Import 确认、回滚快照、导入后校验结果
- 日志页：
  - 自动 smoke：最小 smoke 数据能生成至少一条日志背景
  - 手工 checklist：SSE 实时流、无限滚动、导出下载、超大 JSON 弹层、时间范围校验

### 当前判断

- 这套分层更贴合主 MD 现阶段状态：
  - 自动 smoke 先负责“运行得起来、静态壳没坏、主数据链路没断”
  - 手工 checklist 负责“复杂交互、复杂导入导出、实时流和移动端体验”
- 在浏览器自动化插件或稳定的前端 e2e 基建接入之前，不应把备份页和日志页的复杂分支误报为“已经被全自动完整验收”。

### 2026-04-24 设置页验证链补洞

- 设置页同池的 no-browser 守护现已统一收口到 `web/package.json` 的 `test:settings-no-browser`，覆盖 `Backup / DynamicRouting / CircuitBreaker / ModelProbe / HelpHint / SettingInfo`，不再依赖 workflow 里手写一长串零散脚本名。
- `validation / release` 工作流现在会先通过 `pnpm run test:screenshot-no-browser` 覆盖整个 screenshot-first no-browser 主链，再由该入口顺带跑完整的 settings no-browser 守护，避免 `DynamicRouting`、`CircuitBreaker`、`ModelProbe`、设置版本信息和备份逻辑继续停留在“本地有统一入口，但主验证链只覆盖部分页面”的状态。
- `2026-04-24` 同一主线下又继续收口了统一入口中的两处脚本漂移：`scripts/verify-backup-component.cjs` 现已改为先展开外层“高级迁移能力仍在持续补齐”区，再检查内层“导入工具补强”分组；`scripts/verify-model-probe-help.mjs` 也已同步到当前 `ModelProbe` 的“短主说明 + HelpHint + 摘要卡 + 折叠模型明细”契约，不再错误要求组件继续直出旧版 `defaultPathDesc` 长说明。
- 因此当前 `test:settings-no-browser` 已重新恢复为绿，直接验证结果为 `backup-logic / backup-component / help-hint-accessible / dynamic-routing-help / circuit-breaker-help / model-probe-help / setting-info-logic` 全部通过；同池 no-browser 验证链当前不再遗留脚本级漂移。
- 同轮补充验证显示，前端类型检查可通过 `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 直接跑通；当前更真实的剩余缺口已重新收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `spawn EPERM` / Edge CDP bootstrap 问题，而不再是设置页 no-browser 合同本身。
- `2026-04-24` 同池继续把 screenshot-first 主线的其余 no-browser 护栏统一收口到 `test:screenshot-no-browser`，把 `locale consistency`、`verify-home-layout`、`verify-channel-create-flow`、`verify-channel-presentation`、`verify-group-create-flow`、`verify-llm-price-boundary`、`test:settings-no-browser`、`verify-ccswitch-flow` 和 `verify-route-target-copy` 合并为一个稳定入口，并同步接入 `validation / release`。当前主线不再出现“settings 已统一、但 Home / Channel / Model / Route Target 仍散落在 workflow 命令串里”的验证漂移。
- `2026-04-30` 已继续把这条统一入口真正恢复到当前 Windows 宿主可直接执行的状态：`scripts/use-node-env.ps1` 现在会把 `corepack / pnpm` 缓存和用户配置重定向到仓库内 `.tmp-tooling/`，并显式提供 PowerShell `pnpm` 函数；同时新增 `scripts/run-frontend-verification-suite.mjs` 作为单次 Node 进程的聚合执行器，让 `web/package.json` 的 `test:settings-no-browser` 与 `test:screenshot-no-browser` 不再通过嵌套 `pnpm run ...` 自调用放大 `pnpm config rc EPERM` 风险。统一入口恢复后，本轮又顺手把首页、渠道卡片层与模型页 verifier 的同池 no-browser 漂移收口到当前 `UI_MAINLINE_TASK_2026-04-30` 口径：首页保留单栏例外但补回 `home-page / home-main-grid / home-total-section / home-stats-chart-section / home-rank-section / home-activity-section` 稳定锚点；渠道页补回 `channel-page` 根锚点、卡片层 key 数 / mode / policy badge 与 `channel-card-*` / `channel-detail-dialog-*` 合同；模型/价格 verifier 则从旧的两栏断言更新到当前桌面端三栏。直接验证结果为 `pnpm --dir web run test:settings-no-browser` 与 `pnpm --dir web run test:screenshot-no-browser` 在当前机器上重新跑绿，因此当前主线不再需要回退到“手工逐条 Node 脚本替代统一入口”的临时验证方式。
- `2026-04-30` 同一 settings 池又继续把 `Model Probe` 的 browser smoke 从“卡片存在”收口到“关键交互真实验证”：`scripts/verify-setting-help-browser-smoke.mjs` 与 `...cdp.mjs` 现在会在运行前通过后台 `model/create` API 种入一批探测模型，然后真实断言默认折叠占位、canonical name 搜索、`12` 条分批 `show more`、空搜索态，以及长列表仍留在卡内 scroll region，而不再只检查 `setting-model-probe-card` 是否挂载。与此同时，`scripts/verify-setting-help-browser-smoke.ps1` 也补齐了 settings 专用的 Node 选择优先级和 CLI `spawn EPERM` host blocker 分类：CLI 路径不再误用 Codex bundled Node，也不会再把 Playwright CLI 子进程 `spawn EPERM` 误报成 success-marker 漂移；CDP 路径则继续准确落到既有 `page_bootstrap_timeout_attached_session` 宿主 blocker。当前结论是 `Model Probe` 的 browser-grade 契约已补齐，但本机真实 browser pass 仍分别受 `Edge/CDP bootstrap` 与 `Playwright CLI spawn EPERM` 阻塞，下一轮应优先换宿主或把它视为同主线宿主问题，而不是继续改 `ModelProbe` 页面本身。


- 2026-04-25 已继续收口备份 helper 链中的模型策略 warning 句级本地化：`backup-logic.ts` 现在会把 `billing_mode / probe_policy / probe_interval / probe_concurrency changed from ... to ...` 与 `model:... concurrent probe/race may increase cost` 这类 warning 转成可读的中文/繁中/日文详情文本，`scripts/verify-backup-logic.mjs` 也已补上对应 zh-Hans 与 English 断言；同时 `locale.tsx` 的类型解阻也已完成，因此当前备份主线真正剩余的页面级工作主要是 `Backup.tsx` 细节契约清理与后续浏览器级证据补齐，而不是 locale provider 阻塞；本轮已经为高级迁移折叠区补上 `backup-remaining-migration-panel` 稳定锚点，并把 `backup-page` 根锚点与 `backup-remaining-migration-section-*` 内层锚点一并纳入 `scripts/verify-backup-component.cjs` 的 map/replace 分支断言，同时补齐了 `Backup.test.tsx` 的 dry-run/apply 根锚点断言，下一轮浏览器级证据可以直接复用这些 selector。


