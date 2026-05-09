# Octopus 当前状态与接手计划

> 生成时间：2026-04-14
>
> 目的：基于现有需求文档与当前未提交代码，判断哪些要求已经完成、哪些还没完成、哪些完成得好、哪些完成得不好，并给出后续施工计划。

> 执行硬规则：后续任何实现、补丁、回归修复、UI 调整、路由策略修改、接口扩展与验收动作，都必须先对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)。
> 若当前文档与主规划冲突，以主规划为准；若实现需要偏离主规划，必须先更新主规划，再改代码，再补验证记录。
> 当前阶段还必须同步对齐 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 中的需求总账、图片优先问题池、`Codex` 执行口径和分目录 agent 边界规则。

---

## 0. 当前会话优先补丁（2026-04-22）

为避免本文件停留在早期阶段判断，当前会话新增以下强制优先事项：

- 先按图片问题池返工 UI 与交互，再做其他扩展主线。
- 探测与检测配置必须归位到设置模块，不得继续混入价格模块。
- 多 key、`CC Switch`、全中文清洗、帮助提示（圈内问号悬停说明）和熔断自定义增强进入同一优先返工池。
- 备份导入导出必须兼容旧快照样本，并兼顾同项目恢复与跨项目迁移。
- 所有执行流程统一按 `Codex` 口径，开工前必须完成主规划对齐并在 worklog 记录。

补充（2026-04-23 最新截图优先项）：

- 首页统计卡片、渠道卡片、Token 明细区仍存在明显错位、出框、占地过大和中英混杂问题，必须继续保持 P0 优先级。
- 最近一轮前端壳层、首页和分组页被收得过窄，设置页也被改成单栏顺序流，已偏离用户要的原版铺屏占比与双栏瀑布流，必须优先回退。
- 左侧 dock/切换栏需要恢复到桌面端左侧中线附近的稳定停靠位，同时统一修复中文界面字体观感、字号间距和文本溢出/漂移。
- 渠道创建/编辑弹窗中的多 Key 管理仍未达到“同渠道、多 key、每 key 带模型、默认折叠展开”的目标，必须继续作为主返工项。
- 分组创建/编辑弹窗中的 raw i18n key 泄漏与本地化兜底已完成 no-browser 收口；同一主线下，advanced strategy 折叠头的 `HelpHint` 也已从 trigger 按钮树内移出并补上静态断言。当前剩余缺口主要是浏览器级 `375px / hover / focus` 证据。
- 模型/价格区域的工具栏“普通 / 紧凑”布局切换已在 no-browser 验证层真正接入卡片渲染链，不再是只有入口状态没有卡片差异；同一主线下，模型卡片普通布局里的 `规范名称 / 计费模式 / 官方价格` 也已收紧到统一中文 meta 信息带，并补上了对应 no-browser 断言。当前剩余缺口主要是浏览器级布局证据，以及是否还有新的中文界面英文主显示残留。
- `2026-04-24` 已继续收口模型页浏览器证据后的静态构建验证链：当前仓库内新增 `scripts/build-web-static.mjs` 作为 Windows 宿主稳定入口，会先执行 `tsc --noEmit`，再以兼容补丁方式完成 Next 静态构建，并把 `web/out` 同步到 `static/out`。因此模型页这条 `普通/紧凑布局 + 375px + browser smoke + build:static` 链路已经闭环，下一轮更适合转向首页浏览器级证据。
- 模型页搜索行为也已和当前中文提示对齐：搜索框现在会同时命中模型名称与规范名称，不再存在“文案承诺可搜规范名称，但代码只按模型名过滤”的偏差；对应 no-browser 断言已同步写回 `verify-llm-price-boundary.mjs`。
- `2026-04-24` 已确认并收口模型列表请求兼容性：`/api/v1/channel/fetch-model` 本身并不强制 `https`，真实阻塞点在 providers 预设 `base_url` 校验。该校验现已放开为仅接受绝对 `http/https` URL，继续拒绝其他协议；若后续仍取模失败，应优先排查上游 `/models` 兼容性、连通性或鉴权。
- `2026-04-28` 已继续把“模型列表请求链”和“本机默认地址”从疑似问题收口为已验证状态：repo-local `vitest run src/components/modules/channel/model-fetch.test.tsx` 已通过，确认 `分类 K` 模式下的模型抓取会绑定当前 key / base URL / 代理 / 请求头配置；同时文档弹框与四语 locale 的默认 API 地址已统一改为 `http://127.0.0.1:8080`，规避当前 Windows 宿主上的 `localhost` 解析波动。
- 备份页中文主文案的剩余英文泄漏与关键帮助提示入口已完成 no-browser 收口；本轮继续收口了备份详情区 warning / skip reason 与导出快照 scope badge 的中文展示，并保留下一轮继续压缩详情字段直出的入口。
- 2026-04-25 已继续收口备份 helper 链中的模型策略 warning 句级本地化，并完成 locale provider 的类型解阻；当前 `backup-logic.ts` 已能把模型策略变更与并发成本提示转换成可读的中文/繁中/日文详情文本，`locale.tsx` 的 AI 自动化 locale 结构问题也已消除，并通过 repo-local `verify-backup-logic.mjs` 与 `verify-locale-consistency.mjs` 稳定验证。当前备份主线的剩余前端工作已收敛为 `Backup.tsx` 页面细节契约清理、后续浏览器级证据补齐，以及高级迁移能力继续保持“部分完成”状态；本轮又为高级迁移折叠区补上了稳定的 `backup-remaining-migration-panel` 锚点，并把 `backup-page` 根锚点与 `backup-remaining-migration-section-*` 内层锚点一并纳入 `scripts/verify-backup-component.cjs` 的 map/replace 分支断言，同时补齐了 `Backup.test.tsx` 的 dry-run/apply 根锚点断言，下一轮可以直接基于这些 selector 继续补 browser-grade 证据。
- 2026-05-03 已沿同一 Phase 6 主线把备份导入与回滚预览再收口一轮：`Backup.tsx` 现在不再只是 line-based `model_mappings` textarea，而是以结构化 mapping rows 作为唯一编辑源，并在提交时继续序列化回原有 line payload；回滚预览区也已补上直接消费 `route_preview_diffs` 的“当前状态 vs 快照状态”对比面板。与此同时，后端 `DBImportScopes.Validate()` 与 handler 测试口径已同步改成“空 `import_scopes` 对象合法但为空”，不再误报 `invalid import_scopes`。对应 `go test ./internal/model`、`go test ./internal/server/handlers`、`vitest Backup.test.tsx + backup-logic.test.ts`、`tsc --noEmit`、`verify-backup-logic.mjs` 与 `verify-backup-component.cjs` 已通过。当前 Phase 6 真正剩余的高优先级事项已收敛为“更细粒度的导入冲突引导”与“更细粒度的 rollback domains 编辑”，而不是继续把已工作的 mapping editor / compare workflow / route diff 维持成待办。
- 2026-05-03 同一 Phase 6 主线又把 rollback domains 编辑真正接进了 backup history 主流程：当前 `Backup.tsx` 已新增独立的 rollback scope editor，可在 `routing / models / api_keys / settings / stats / logs` 六个域上选择性回滚；selected snapshot 的 preview / rollback 会共享同一组 scope 入参，scope 变化会立即使旧 rollback preview 失效，避免继续展示过期 compare 结果；若未选中任何域，则仍按 full snapshot restore 回退，不会向后端发送空 scope 对象。对应 `tsc --noEmit`、`verify-backup-component.cjs` 与 `verify-backup-logic.mjs` 已通过；`Backup.test.tsx` 的 Vitest 入口在当前 Windows 宿主上仍被 `vite/vitest spawn EPERM` 阻塞，因此这条验证暂时继续记作环境 blocker，而不是代码回退。当前 Phase 6 剩余最高优先级事项已进一步收敛为“更细粒度的导入冲突引导”，以及在条件允许时补一轮 browser/runtime 证据。
- 2026-05-03 同一 Phase 6 主线继续把“更细粒度的导入冲突引导”真正收口进备份页：`Backup.tsx` 现已基于既有 `compatibility` payload 补出“下一步处理建议”分组，并把 `credential_rebind_targets`、`invalid_route_targets`、`model_policy_diffs` 等结构化细项直接展开到兼容性详情区。用户现在不再只看到冲突计数，而是能直接按“先处理阻断风险 / 补齐缺失对象 / 准备凭证重绑定 / 修正模型映射 / 复核路由与策略差异”执行下一步。对应 `tsc --noEmit`、`verify-backup-logic.mjs` 与 `verify-backup-component.cjs` 已通过；`backup-logic.test.ts` 与 `Backup.test.tsx` 也已同步更新，但当前 Windows 宿主上的 `Vitest/Vite spawn EPERM` 仍是环境 blocker。当前 Phase 6 真正剩余高优先级事项因此进一步收敛为 `Backup` 页 browser/runtime 证据补齐，以及宿主级 Vitest 阻塞的后续处理，而不是导入冲突引导本身。
- 2026-05-03 同一 Phase 6 主线继续把上述导入冲突引导从“有建议卡”收口到“建议卡与下方细节真正一一对应”：当前 `Backup.tsx` 的兼容性详情区已新增 `route_preview_warnings`、`route_preview_diffs` 与 `alias_preview_mappings` 的结构化展开，用户在看到“复核路由与策略差异”后，可以直接在同一区块查看路由预警、候选链路差异与别名映射预览，而不必只靠 signal list 或去猜具体受影响对象。对应 `verify-backup-logic.mjs`、`verify-backup-component.cjs` 与 `tsc --noEmit` 已再次通过；`backup-logic.test.ts` / `Backup.test.tsx` 断言也已同步补齐，但 `Vitest/Vite spawn EPERM` 仍保持为当前 Windows 宿主的环境 blocker，因此剩余工作继续收敛为 `Backup` 页 browser/runtime 证据补齐和宿主级 Vitest 解阻，而不是兼容性详情消费链本身。
- 2026-05-03 同一 Phase 6 主线继续把 `Backup` 页 browser/runtime 证据入口从“只剩 CLI wrapper”收口到“CLI + CDP 双入口”：当前 `scripts/verify-backup-browser-smoke.ps1` 已公开 `-Driver cdp|cli`，新增的 `scripts/verify-backup-browser-smoke-cdp.ps1` 会像 `group-create / channel-page / settings-help` 一样继续复用共享 `verify-channel-create-browser-smoke-cdp.ps1`，而 `scripts/verify-backup-browser-smoke-cdp.mjs` 则真正承接 Backup 导入 dry-run、Apply Same Import、history、rollback preview 与 `375px` 检查的 Edge CDP 场景。对应 `--check`、`verify-backup-browser-smoke.ps1 -Mode check-only -Driver cdp`、`-BrowserPath` 兼容调用，以及 `verify-browser-smoke-wrapper-alignment.mjs` 已通过，因此当前剩余工作已进一步收敛为“在可用宿主/外部 Edge 会话上补一轮 live CDP 证据”，而不是继续为 Backup 页面补 wrapper 入口本身。
- 2026-05-03 同一 Phase 6 主线继续把 `Backup` 页 live smoke 覆盖面从“只看 preview 面板存在”收口到“真正覆盖 rollback scope editor 与 rollback compare 合同”：当前 `scripts/verify-backup-browser-smoke.mjs` 与 `scripts/verify-backup-browser-smoke-cdp.mjs` 已新增 rollback scope 默认 full-restore 摘要、`routing / stats / logs` 关闭后的 narrowed scope 摘要、scope 变化触发 preview invalidation、再次 preview 后的 `applied_scopes` 回显，以及 rollback route-diff compare 面板/compare row 断言。repo-local `--check`、`check-only -Driver cli|cdp` 与 `verify-backup-component.cjs` 已通过；真实 `self-start -Driver cli` 仍稳定卡在 Playwright CLI `spawn EPERM`，`self-start -Driver cdp` 仍稳定卡在 `http://127.0.0.1:9222/json/version` 启动超时，因此当前真正剩余的高优先级事项已更明确收敛为“换可用宿主或外部 Edge 会话补 live 绿灯证据”，而不是继续补 Backup smoke 断言本身。
- 2026-05-07 同一 Phase 6 主线继续把 `Backup` 导入兼容性详情从“有 guidance 和 signal 计数”收口到“关键兼容性字段都能直接看到具体对象”：当前 `Backup.tsx` 已把 `affected_groups / affected_channels / base_url_mismatches / schema_mismatches / skipped_targets` 接进兼容性详情区，`backup-logic.ts` 也为 `channel_key/api_key empty credential`、`skip mode preserved` 与 `snapshot schema differs` 补上了共享本地化格式化。对应 `tsc --noEmit`、`verify-backup-logic.mjs` 与 `verify-backup-component.cjs` 已通过，因此当前 Phase 6 在 repo-local 可继续推进的细项已进一步收敛为“是否还存在未消费的 import compatibility 字段 / replace-map 引导缺口”，以及宿主条件允许时再补 live browser/CDP 证据，而不是兼容性细项继续半隐藏。
- 动态路由页的主说明与摘要消息中文化已完成 no-browser 收口。
- `HelpHint` 组件的默认无障碍标签与验证选择器已完成 no-browser 收口：帮助按钮不再把中文“查看帮助”硬编码到所有语言，设置页与备份页的回归入口也已切到稳定 selector，降低后续多语言与 smoke 漂移风险。
- 设置页四卡片 `HelpHint` 的浏览器级主证据已从初版合成事件 smoke 收紧并通过：当前 `cli self-start` 路径会验证 `375px`、键盘 `Tab` 聚焦、真实 `hover`、以及当前 trigger 绑定的 tooltip 文本，最新结果为 `desktopHelpButtons: 21`、`interactionChecks: 4`、`mobileWidth: 375`、`setting-help-browser-smoke passed`。剩余 CDP 项继续按宿主级 Edge/CDP bootstrap blocker 单独跟踪，不再阻塞 settings help-hint 主证据闭环。
- `2026-04-24` 已恢复 `CC Switch` 的浏览器级主证据：当前 `scripts/verify-ccswitch-browser-smoke.ps1` 默认走 `attached-session + runtime-page-lifecycle` 的宿主已验证顺序，`self-start + cdp` 可在本机完成 `DocModal` / `CC Switch` 页签、帮助提示 `focus + hover`、导入门控，以及 `375px` 宽度无大幅横向溢出的验证。CLI 路径里的 `Node spawn EPERM` 与个别 Edge 远程调试端口启动波动继续作为宿主诊断项保留，但不再阻塞 `CC Switch` 主证据闭环。
- `2026-04-24` 已继续收口渠道总页的 selector/no-browser 合同：当前 `scripts/verify-channel-page-browser-smoke.ps1`、`scripts/verify-channel-presentation.mjs` 与详情区稳定 test id 已对齐到 `channel-page` 场景，覆盖工具栏中的“提供商 + 模型关键词 + key 信息”组合筛选、详情弹层打开、key 行展开，以及单 key 的 route-target override 摘要入口。页面级真实 browser pass 在本机仍受 Edge/CDP bootstrap 阻塞，因此当前剩余缺口主要收敛到 `9.1.1` 更细字段和弹窗级细节交互，而不是 selector 契约本身。
- 性能问题已被用户在本地环境直接感知，说明 UI 与交互返工必须同步兼顾渲染和布局收缩。
- `2026-04-28` 已完成一轮回退式修复：主壳层恢复到更接近原版的铺屏占比，设置页恢复双栏瀑布流，左侧导航恢复桌面端中线停靠，并统一修复全局中文字体回退；对应 `pnpm exec tsc --noEmit` 与 `pnpm run test:screenshot-no-browser` 已通过。
- `2026-04-28` 同一轮继续把中文统一与 no-browser 验收链重新压实：当前 `pnpm run test:locale-consistency`、`pnpm run test:screenshot-no-browser`、`pnpm exec tsc --noEmit` 与渠道模型抓取单测均已通过；简中、繁中里最容易直出的 `Profile / Hybrid / Shadow AI` 等词又完成一轮主显示收口，并同步加严 `verify-locale-consistency.mjs`，后续若回退到英文主显示会直接被门禁挡住。
- `2026-04-28` 同一轮又把最新 UI 恢复口径写进前端验收链：主壳层保持更宽但不撑爆的内容区，左侧 dock 固定在桌面端左侧中线附近，首页主网格改为主内容更宽、右栏更克制的比例，渠道页列表改为更紧凑的桌面端密度；渠道创建/编辑的多 Key 折叠也从大块说明压缩成“单行摘要 + 状态徽标 + 展开后填写真实密钥和模型范围”的结构，避免再把用户主交互区挤没。
- `2026-04-28` 同一轮把 AI 自动化深层文案也纳入四语一致性门禁：`scripts/verify-locale-consistency.mjs` 现在覆盖 `aiAutomation`，并对简中、繁中、日文里容易泄漏的 `Profile / endpoint / base URL / API Key / group_items / priority / manual / OpenAI-compatible` 等主显示词加了更硬的检查。当前结果是 no-browser 与类型链通过；由于这轮刚调整了主壳层比例，桌面端与 `375px` 真实浏览器截图仍需要按最新源码再补一轮视觉证据。
- `2026-04-28` 同一轮已补最新源码下的首页与渠道页真实浏览器主证据：`verify-home-layout-browser-smoke.ps1 -Mode self-start` 与 `verify-channel-page-browser-smoke.ps1 -Mode self-start` 均通过，验证了临时运行态、Edge CDP、首页桌面/窄屏布局、渠道页组合筛选、详情弹层、key 行展开与 `375px` 无明显横向溢出。验证后已执行 `scripts/runtime-win.ps1 stop`，确认 `3000 / 3001 / 3101 / 8080 / 18081 / 9222` 均无监听，项目未被留在本机运行态。
- `2026-04-28` 同一轮继续向“剩余高优先级交互细节”推进时，又确认了一类独立宿主 blocker：`verify-channel-create-browser-smoke.ps1` / `verify-model-layout-browser-smoke.ps1` 这类 `self-start` 浏览器 smoke 目前会被当前 PowerShell 会话里的 Windows socket service-provider 初始化故障拦住，表现为临时后端在 `listen tcp` 阶段直接报 `The requested service provider could not be loaded or initialized`。本轮已把共享包装器补上 loopback 端口探测重试、端口空闲回退判断和后端自启动重试，因此当前阻塞已明确收敛为宿主网络层问题，而不是渠道创建或模型页 UI 自身回归。
- 当前本机运行策略已收敛为“项目默认停驻，按需启动，验证结束即回收”；Windows 本机统一通过 `scripts/runtime-win.ps1` 做 `status / stop / healthcheck / check-only`，避免旧进程长期占用端口干扰后续自动化与手工验证。
- `scripts/runtime-win.ps1` 已补齐低权限回退链路：当 `Get-CimInstance` 被拒绝时，会优先降级到 `Get-Process + Get-NetTCPConnection`，若当前会话拿不到 `Get-NetTCPConnection` 归属信息，则继续回退到 `netstat -ano -p tcp`，最后才退回纯 `.NET listeners` 探测；因此 `status` 在低权限主机上现在通常仍能给出端口 owner PID / 进程名提示，同时继续保持 `stop` 只针对已归因到 `octopus_repo` 的进程，不会因为端口提示误停其他项目。
- 新增中长期主线：`AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习`。该主线要求新增顶层 `AI 自动化` 栏目，支持自定义 AI 渠道/模型、本机默认 Octopus endpoint、自动模型发现、自然语言任务、提示词模板、进度条和 AI Profile 保存；同时必须保证用户手动配置与 AI 生成 Profile 双轨并存，AI 不覆盖原配置。
- 动态路由在该主线中作为例外专线处理：不走普通 AI Profile 覆盖逻辑，而是在设置页通过 `dynamic_routing_learning_enabled` 手动控制本地 AI 学习是否参与运行时推荐；学习不得写回 `group_items`，不得覆盖用户 priority。

后续状态更新时，必须与 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md) 的总账编号同步。

### 0.1 AI 自动化中心主线状态（2026-04-23）

当前状态：主入口、后端接口、任务执行链、AI Profile 保存/手动激活、设置页来源切换与动态路由学习开关均已接线，不再属于“代码实现未开始”阶段；当前主要缺口已收敛为受保护动作执行语义、UI 体验收口、浏览器级证据与主文档同步。

- `2026-04-28` 已继续把设置页动态路由区域收口到当前主线口径：设置页现在提供 `dynamic_routing_learning_enabled` 的轻量开关、学习作用范围摘要，以及跳转到顶层 `AI 自动化` 页面查看学习样本/复盘/重置的入口；学习样本明细没有重新堆回 settings，保持了“设置页轻量、AI 自动化页承接重任务”的既有分工。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前这条子线的剩余缺口主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 仍导致新增 jsdom 用例无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线又继续把 `AI 自动化` 页面学习区的运行时状态边界补齐：页面现在会明确展示学习是否参与推荐评分、当前样本数，以及针对“无样本 / 已关闭但保留样本”等状态的首屏说明；空样本时 `Reset learning data` 也会禁用并显示轻提示，不再让用户在 settings 跳转后还需要自己判断当前状态。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、四语 locale 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍然主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致新增 jsdom 用例无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把学习区首屏摘要收口到“无需下翻样本卡也能先判断状态”的口径：`AI 自动化` 学习区现在新增最近采样时间和当前最高分对象摘要，样本卡也会直接展示最近采样时间，承接上一轮已完成的 settings -> learning 跳转与状态说明。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、四语 locale 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级证据与宿主 `vitest/esbuild spawn EPERM`。
- `2026-04-28` 同一 Phase H6 子线继续把 settings 学习卡补到文档承诺的最小闭环：设置页现在会直接显示真实样本数、最近采样时间和当前最高分对象，并提供就地 `reset learning state` 轻入口；无样本时该入口会禁用并显示轻提示，因此 settings 已同时具备学习开关、状态摘要、查询入口和清空入口，而详细样本列表仍留在 `AI 自动化` 页。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续收口两个学习入口之间的动作反馈一致性：顶层 `AI 自动化` 页面里的学习重置动作现已补上与 settings 一致的成功/失败 toast，不再是静默清空；对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx` 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 settings 学习卡的运行时影响摘要提到首屏：当前卡片除了状态、样本数、最近采样时间与最高分对象之外，还会直接显示 learning 是否正在参与 runtime scoring，从而在不跳转 `AI 自动化` 页的情况下，也能先判断“当前有没有参与推荐评分”。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、四语 locale 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-dynamic-routing-help.mjs`、`verify-ai-automation-learning-focus.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 settings 学习卡与顶层 `AI 自动化` 学习区的运行时状态来源收口到同一口径：settings 卡现在会优先读取 `dynamic-routing/learning` 返回的 `enabled` 状态来判断是否参与 runtime scoring，并在保存 learning 开关后主动刷新 learning 查询；同时 `reset learning` mutation 也会额外失效 `ai/config` 与 `settings/list`，减少两个 consumer 在刷新窗口里的状态漂移。对应 `DynamicRouting.tsx`、`DynamicRouting.test.tsx`、`web/src/api/endpoints/ai-automation.ts` 与 `verify-dynamic-routing-help.mjs` 已同步更新；当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把顶层 `AI 自动化` 学习开关的状态回退链也收口到与 settings 一致的语义：页面现在会优先采用本地草稿态，其次读取 `dynamic-routing/learning` 的运行时 `enabled`，最后再回退到 `ai/config.dynamic_routing_learning_enabled`，从而减少 learning query 短空窗里的错误关闭态；learning 开关保存失败时也会回退草稿并显示专用失败 toast。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx`、`verify-ai-automation-learning-focus.mjs` 与四语 locale 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`。当前剩余缺口仍主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把顶层 `AI 自动化` 学习摘要里的 top target 选择语义收口到与 settings 一致：学习摘要不再直接取数组首项作为“当前最高分对象”，而是和 settings 一样按最高 `score` 选取，同时最近采样时间继续独立按最新 `last_sample_at` 计算。对应 `ai-automation/index.tsx`、`ai-automation/index.test.tsx` 与 `verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs`；真实 `vitest` 入口仍受宿主 `esbuild spawn EPERM` 阻塞，浏览器级证据仍受本机 loopback 初始化问题阻塞。
- `2026-04-28` 同一 Phase H6 子线继续把 learning 摘要推导本身收口到共享 helper：当前 `AI 自动化` 学习区与 settings 学习卡已共用 `learning-summary.ts` 中的 `latest sample / top target / has samples` 推导与采样时间格式化逻辑，不再分别手写 `reduce` 链，从而减少后续继续出现 consumer 语义漂移的风险。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护与 `verify-locale-consistency.mjs`。当前剩余缺口仍主要收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 learning 展示状态派生也收口到共享 helper：`learning-summary.ts` 现已统一产出 `sampleCount / canReset / runtimeKey / noticeState`，顶层 `AI 自动化` 学习区与 settings 学习卡不再分别维护样本数、reset 按钮可用性和 enabled/disabled notice 分支，从而把 H6 consumer 剩余的重复条件判断继续压缩到同一数据源。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把两个 consumer 的 learning 摘要展示骨架也收口到共享面板：新增 `LearningSummaryPanel.tsx` 承接摘要栅格与 notice 区展示后，`AI 自动化` 学习区和 settings 学习卡不再各自维护一套近似 JSX，只保留各自不同的字段集合与动作区。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`LearningSummaryPanel.tsx`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 learning 摘要 footer 动作条也收口到共享实现：`LearningSummaryPanel.tsx` 现已新增共享 action bar，`AI 自动化` 学习区与 settings 学习卡不再分别维护 reset/open 按钮区与禁用提示，只保留各自动作配置与既有 test id。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`LearningSummaryPanel.tsx`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把两个 consumer 剩余的 learning 摘要 item 组装和 notice 选择也收口到共享 helper：`learning-summary.ts` 现在新增 `resolveLearningNoticeValue`、`formatLearningTopTarget` 与 `buildLearningSummaryItems`，`AI 自动化` 学习区和 settings 学习卡不再分别维护 top target fallback、enabled/disabled notice 文案分支与 summary items 列表组装；settings 只在共享 base items 之上额外拼接 `scope / details` 差异字段。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把两个 consumer 剩余的 learning 摘要 view-model 组装也并入共享入口：`learning-summary.ts` 现已新增 `buildLearningSummaryViewModel`，统一返回 `summary / display / latestSampleLabel / topTargetLabel / notice / items`，因此 `AI 自动化` 学习区与 settings 学习卡不再分别维护 latest sample、top target、notice 和 summary items 的整条组装链；同时也明确保留“settings 轻摘要、AI 自动化页承接样本卡”的边界，不为了抽象而强行共享 sample card。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 settings 学习卡对共享摘要项的依赖从“按数组顺序取值”收口到“按语义 key 取值”：`learning-summary.ts` 新增 `indexLearningSummaryItems` 后，settings 不再依赖 `items[0..4]` 的隐式顺序，而是显式读取 `status / samples / runtime / latest-sample / top-target`，从而减少后续新增摘要项时无声打乱 settings 展示顺序的风险。对应 `setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 settings 学习卡的摘要分组组装也并入共享 helper：`learning-summary.ts` 现已新增 `buildLearningSummarySections`，由它统一解析 `status / samples / runtime / latest-sample / top-target` 这些共享摘要项，并按 `primary / secondary` section 返回结果；因此 `DynamicRouting.tsx` 不再自己维护 item 索引、缺项报错和 base section 组装，只保留 `scope / details` 两个 settings 专属差异字段。对应 `setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口仍主要收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把两个 consumer 的默认摘要分组定义也收口到共享常量：`learning-summary.ts` 现已新增 `DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES`，顶层 `AI 自动化` 学习区改为通过 `buildLearningSummarySections` 组装 `primary / secondary` 两组 grid，不再把 `items` 整包直接塞入单个摘要栅格；settings 学习卡也改为在同一套默认 section entry 上插入 `scope / details` 两个差异字段，从而进一步压缩两处残留的重复 key 列表。对应 `ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、`learning-summary.ts`、`verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs` 已同步更新，并通过 repo-local `tsc --noEmit`、两条 no-browser 守护、`verify-locale-consistency.mjs` 与 `git diff --check`。当前剩余缺口继续收敛为浏览器级 `375px / hover / focus` 证据，以及宿主 `vitest/esbuild spawn EPERM` 导致真实 jsdom 入口仍无法在本机补跑。
- `2026-04-28` 同一 Phase H6 子线继续把 `AI 自动化` 学习区的 browser smoke 入口从 CLI-only 收口到共享 CDP 包装链：`verify-ai-automation-learning-browser-smoke.ps1` 现在默认复用 `verify-channel-create-browser-smoke-cdp.ps1`，共享 `verify-channel-create-browser-smoke-cdp.mjs` 也新增了 `ai-learning` 场景，会主动种最小 learning 样本并验证 learning 页加载、preset、reset、switch 与 `375px` 宽度；同时共享 CDP wrapper 的 Node 解析已对齐到稳定本地 Node，并把 loopback service-provider 失败改成准确的 host networking blocker，不再误报成“找不到空闲端口”。对应 `scripts/verify-channel-create-browser-smoke-cdp.mjs`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check`；真实 `self-start` 仍受当前 Windows 宿主 loopback service-provider 初始化失败阻塞，但阻塞分类已收口准确。
- `2026-04-30` 同一 Phase G 主线继续把 browser smoke 共享根 wrapper 的公开浏览器参数口径彻底收齐：当前 `verify-channel-create-browser-smoke.ps1` 已保留 `[Alias('BrowserPath')]`，`verify-channel-create-browser-smoke-cdp.ps1` 也改为公开 `-Browser` 并兼容旧 `-BrowserPath`，内部仍统一映射到底层 `BrowserPath` 与 Edge bootstrap helper。对应 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 已把 shared-root alias 与 `Browser -> BrowserPath` 映射纳入静态守门，且共享 CLI/CDP 根 wrapper 的旧 `-BrowserPath` 与新 `-Browser` `check-only` 调用都已通过。因此当前同池 browser smoke PowerShell 家族的参数面不再只在 thin forwarder 一侧统一，shared root 也已同步收口；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 wrapper 参数面分裂。
- `2026-04-30` 同一 Phase G 主线继续把 shared-root `Driver=cdp` 的高级参数面也补到统一根入口：当前 `verify-channel-create-browser-smoke.ps1` 已不再只是把 `Browser` 和基础端口转给 `verify-channel-create-browser-smoke-cdp.ps1`，而是同步透传 `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices`。对应 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 也新增了 shared-root `Driver=cdp` 参数面的静态守门，因此统一根命令现在既能兼容旧 `-BrowserPath`，也能直接承接 page-level forwarder 已有的 external-preflight / bootstrap 参数，不再出现“page-level wrapper 可以、shared root 顶层命令却报 unknown parameter”的能力面分裂。当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 shared root `cdp` 参数透传缺口。
- `2026-04-30` 同一 Phase G 主线又继续把这条参数统一口径补到受控特例 `verify-ai-automation-learning-browser-smoke.ps1`：当前 `ai-learning` specialized root 也显式接受 `-Browser` 与兼容 `-BrowserPath`，并继续把浏览器路径映射到共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 `BrowserPath`。`verify-browser-smoke-wrapper-alignment.mjs` 也同步把这条特例的 `[Alias('BrowserPath')]`、`SelfStartServices / SelfStartLocalServices` 合流，以及 host-friendly 默认参数仍会启用 `BootstrapExternalCdpSession` 纳入 guard。这样当前 browser smoke wrapper 家族连 specialized root 也不再残留旧参数面分裂；下一轮可以继续直接聚焦真实 browser/CDP 证据与宿主 blocker，而不是回到 wrapper 契约补洞。
- `2026-04-30` 同一 Phase G 主线继续把 `ai-learning` specialized root 的 stable-diagnostic 合同也补进静态守门：当前 `verify-browser-smoke-wrapper-alignment.mjs` 不再只检查 alias、self-start 合流和 host-friendly bootstrap，还额外锁定 `StableDiagnosticFreshnessThresholdHours`、timeout-comparison helper，以及 `check-only` 预览里固定输出 freshness threshold。这样 specialized root 后续若再次回退成“还能 forward，但 stable-diagnostic 预览合同悄悄缩水”，`test:screenshot-no-browser` 会先在 repo-local 失败；当前主线剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 preview/guard 漂移。
- `2026-04-30` 同一 Phase G 主线又继续把 page-level browser smoke wrapper 的默认 bootstrap 契约也纳入 repo-local guard：`scripts/verify-browser-smoke-wrapper-alignment.mjs` 现在不再只检查 `Browser/BrowserPath` alias、shared-wrapper 引用与参数透传，还会额外锁定 `verify-ccswitch-browser-smoke.ps1` 的默认 `attached-session + runtime-page-lifecycle` 顺序，以及 `group-create-cdp / home-layout / model-layout / channel-page` 这些 page-level wrapper 维持 `attached-session + page-lifecycle-runtime` 的默认值。这样后续即使没有人去动共享 root，只要某个 thin forwarder 的默认 `CdpBootstrapCommandOrder` 或 `CdpPageBootstrapStrategy` 被无声改坏，`test:screenshot-no-browser` 也会先在 repo-local 报错，而不用等到下一轮再人工回忆“哪一页本来默认走哪条 bootstrap 顺序”。当前剩余工作因此继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 wrapper 默认值漂移。
- `2026-04-30` 同一 Phase G 主线又继续收口了一处共享 Node smoke 入口漂移：共享 `scripts/verify-channel-create-browser-smoke-cdp.mjs` 的 `printUsage()` 先前仍错误提示执行 `verify-setting-help-browser-smoke-cdp.mjs`，会把后续接手人带到错误入口；本轮已改回真实共享入口 `verify-channel-create-browser-smoke-cdp.mjs`，并在 `scripts/verify-ai-automation-learning-focus.mjs` 中补上 usage 文案守门，明确要求共享 CDP 场景不再回退到 `setting-help` 专属脚本名。对应 `verify-ai-automation-learning-focus.mjs`、`run-frontend-verification-suite.mjs settings` 与 `git diff --check` 已通过，因此当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是共享 usage/文档入口漂移。
- `2026-04-30` 同一 Phase G 主线随后又把 `settings-help` page-level CDP wrapper 的默认 bootstrap 契约补进 repo-local 守门：`scripts/verify-setting-help-browser-smoke.ps1` 当前约定的 `CdpPageBootstrapStrategy='auto'` 与 `CdpBootstrapCommandOrder='page-lifecycle-runtime'` 现已纳入 `scripts/verify-browser-smoke-wrapper-alignment.mjs`。这样后续如果有人只改了 `settings-help` 这条 thin forwarder 的默认值，而没有同步状态文档或 live 证据，`test:screenshot-no-browser` 会先失败；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 page-level 默认 bootstrap 漂移。
- `2026-05-01` 同一 Phase G 主线继续把共享 CLI wrapper 的宿主 blocker 语义补成静态守门：`scripts/verify-channel-create-browser-smoke.ps1` 现在在 Node 解析失败时不再硬编码 `channel create smoke`，而是跟随当前 `OCTOPUS_UI_SMOKE_LABEL` 输出页面标签；同时 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 已新增对这条 label-aware Node 失败提示、既有 `spawn EPERM` host-blocker 分类，以及 `octopus-<label>-smoke-*` 临时目录命名的断言。这样 `backup / ccswitch / group-create / settings-help` 这组共享 CLI thin forwarder 后续即使再遇到 Node/Playwright 宿主级失败，也不会无声回退成误导性的 `channel create` 专属报错；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是共享 CLI wrapper 契约漂移。
- `2026-05-01` 同一 Phase G 主线又继续收掉了两类 repo-local 漂移：其一，活跃状态/workflow 文档此前仍把 canonical plan 和状态入口指向不存在的 `docs/*.zh-CN.md` 根路径，本轮已统一改回真实 `docs/archive/planning/` 与 `docs/archive/status/` 路径，并把这组入口合同纳入 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 静态守门；其二，上一轮补进 `runtime-win.ps1` 的低权限 `status/check-only` contract 现在也被 guard 锁定，包括 `Port scan mode`、loopback readiness、runtime entrypoint 提示与 `Automation entrypoints` 输出，因此下一轮不需要再靠人工回忆这些支持层约定。顺手还把 `scripts/verify-ai-automation-learning-focus.mjs` 对 `AI 自动化` 学习区锚点的旧断言从 `ref={learningSectionRef}` 更新到当前 `bindWorkbenchSection('learning')` 绑定写法，使 `node .\scripts\run-frontend-verification-suite.mjs screenshot` 在本机重新恢复全绿；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是活跃文档入口或 no-browser guard 自身漂移。
- `2026-05-01` 同一 Phase G 主线继续把“仍被要求开工前先读”的 archive 输入文档入口也收口到真实归档路径：`USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、`DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`、`AI_AUTOMATION_CENTER_REQUIREMENTS(.md/.zh-CN.md)` 与 `worklog/README.zh-CN.md` 先前仍回指不存在的根目录 `docs/*.md` / `docs/*.zh-CN.md`，本轮已统一改回 `docs/archive/planning/`、`docs/archive/requirements/` 与 `docs/archive/status/` 的真实入口，并把这些活跃 archive 输入文档一并纳入 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 守门。对应 `verify-browser-smoke-wrapper-alignment.mjs`、`run-frontend-verification-suite.mjs screenshot` 与定向 `git diff --check` 已通过，因此下一轮接手时不会再从错误的 requirements/planning/worklog 入口起步；当前剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，而不是 archive 输入文档入口漂移。
- `2026-04-28` 同一 Phase H6 子线继续把 learning smoke 的 `external + cdp` 复跑入口收口成 host-friendly 默认组合：`verify-ai-automation-learning-browser-smoke.ps1` 现已新增 `-UseHostFriendlyExternalDefaults`，会显式注入 `CdpUrl=http://127.0.0.1:9233`、`NodeSmokeTimeoutSeconds=70`、`CdpCommandTimeoutMs=30000`、`EdgeLaunchPreset=relaxed`、`EdgeProfileStrategy=workspace-fixed`，并自动打开 `BootstrapExternalCdpSession`；本地后端/前端自启动不再被默认打开，只有显式加 `-SelfStartServices`（或兼容别名 `-SelfStartLocalServices`）时才会走旧的 local service bootstrap。因此在 loopback service-provider 已损坏的 Windows 宿主上，host-friendly external 路径不会再提前死在本地端口探测上，而是优先走外部 CDP 预检与真实入口分类；对应 wrapper 守护已同步写入 `verify-ai-automation-learning-focus.mjs`，并通过 repo-local 守护、check-only、真实 `external` 复跑与 `tsc --noEmit`。
- `2026-04-28` 同一 Phase H6 子线继续把 host-friendly `external + cdp` 的粗粒度等待收口成结构化 external preflight：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已新增 backend / frontend / CDP 三段 reachability 预检，会输出具体分类、提示语与 `external-preflight-diagnostic.json`，因此 `verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 在当前宿主上已不再只报 `Timed out waiting for ...`，而是先明确把 `backend healthcheck` 归类为 `host_networking_blocker`。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`tsc --noEmit` 与 `git diff --check` 已同步通过；下一轮可直接基于该诊断文件继续区分 backend/frontend/CDP 哪一层外部依赖仍不可达。
- `2026-04-28` 同一 Phase H6 子线继续把 external preflight 从“第一处失败即停止”收口成“尽量收集全量 reachability 结果后统一报错”：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现已新增 skipped entry、failed-check aggregation、统一 hint 生成和 `schemaVersion = 2` 的 diagnostic JSON；因此 `verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 在当前宿主上会一次性报告 `backend + frontend` 的 `host_networking_blocker`，并把当前未要求的 CDP 检查明确标记为 `skipped`，而不是继续停在首个 backend 阻塞点。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`tsc --noEmit` 与 `git diff --check` 已同步通过；下一轮可直接沿这份聚合诊断判断是先修服务可达性，还是在健康宿主上继续收集真实 browser 证据。
- `2026-04-28` 同一 Phase H6 子线继续把 aggregated external preflight 真正收口成“wrapper 可直接消费”的失败摘要：共享 `verify-channel-create-browser-smoke-cdp.ps1` 现在会把 `skippedChecks / primaryBlockingCheck / summaryLines / hints / checkDetails` 一并写入 `external-preflight-diagnostic.json`，而 `verify-ai-automation-learning-browser-smoke.ps1` 会在 external 失败时自动解析 `Diagnostic:` 路径并先打印一段稳定的 `Latest external preflight diagnostic` 摘要，因此在当前宿主上已能直接看到 `preflight_failed`、`backend + frontend`、`cdp skipped` 与诊断文件路径，不再需要人工再进临时目录翻 JSON。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`runtime-win status`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按环境阻塞预期失败；下一轮可直接基于 wrapper 打印的摘要决定是先补外部服务可达性，还是在健康宿主上继续追 CDP/browser 证据。
- `2026-04-28` 同一 Phase H6 子线继续把 learning wrapper 的 external 失败从“有诊断摘要但还要自己想下一步”收口成命令级指导：`verify-ai-automation-learning-browser-smoke.ps1` 现在会在 `Latest external preflight diagnostic` 后继续打印 `External preflight next steps`，明确给出标准 external 复跑命令，以及需要本机对照时追加 `-SelfStartServices` 的变体；`verify-ai-automation-learning-focus.mjs` 已同步守住该输出契约，并通过 repo-local `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、learning smoke `check-only` 与一次按环境阻塞预期失败的 `external` 复跑。当前产品侧剩余缺口仍主要是可达服务 / 健康宿主上的真实 browser 证据，而不是 wrapper 可消费性。
- `2026-04-28` 同一 Phase H6 子线继续把 external 诊断入口从“临时目录可消费”收口到“repo-local 可继承”状态：`verify-ai-automation-learning-browser-smoke.ps1` 现在会把 external 失败里解析出的 `external-preflight-diagnostic.json` 额外复制到 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`，并在 `check-only` 与真实失败摘要里同时打印这条稳定副本路径。这样下一轮即使临时目录已变化，也可以优先从 repo 内稳定副本继续对照诊断；对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、真实 `external` 复跑、`tsc --noEmit` 与 `git diff --check` 已同步纳入同一条 H6 验证链。
- `2026-04-28` 同一 Phase H6 子线继续把 `check-only` 入口从“只告诉你稳定副本放哪”收口成“直接回放最近一次稳定诊断摘要”：当前 `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults` 若仓库内已有 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`，会直接打印 `Latest external preflight diagnostic`、`failed/skipped checks`、summary lines、hints 与 next steps；若稳定副本缺失或损坏，也会明确提示应先跑一次 external 种子或直接检查 JSON。这样下一轮即使不重跑 external，也能沿同一 repo-local artifact 继续判断当前仍是 `host_networking_blocker`，还是已进入后续 CDP/browser 问题；本轮也已清理历史 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁残留。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`tsc --noEmit`、`runtime-win status` 与 `git diff --check` 已同步纳入验证链。
- `2026-04-28` 同一 Phase H6 子线继续把 stable diagnostic 回放再收口到“能判断是不是旧结果”的口径：`verify-ai-automation-learning-browser-smoke.ps1` 现在在 `check-only` 回放 repo-local 稳定副本时，会额外打印 `diagnostic source / checked at / diagnostic age`，并明确提示这只是最近一次 external 失败保存下来的诊断，而不是 live probe；因此下一轮接手时不容易把旧副本误当成当前外部环境。对应 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步纳入验证链；当前真实 browser 级缺口仍然是宿主 `host_networking_blocker` 与 `vite/esbuild spawn EPERM`，不是 wrapper 消费能力不足。
- `2026-04-28` 同一 Phase H6 子线继续把 learning smoke 的 external CDP 预检从“隐式依赖 bootstrap helper 是否启用”收口成显式可控入口：共享 `verify-channel-create-browser-smoke-cdp.ps1` 新增 `RequireExternalCdpPreflight` 后，external 第一步的 `RequireCdp` 逻辑已统一由显式参数控制；`verify-ai-automation-learning-browser-smoke.ps1` 也新增同名透传开关，并在 stable diagnostic 回放里直接显示 `External preflight CDP requirement`。这样下一轮若需要在健康宿主上第一时间拿到“backend/frontend + CDP 同时预检”的 fresh diagnostic，就可以直接运行 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`，而不必再依赖隐式分支推断为什么 stable copy 里 `cdp` 是 `skipped`。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`（含显式 `-RequireExternalCdpPreflight`）与 `tsc --noEmit` 已同步通过；当前 stable copy 仍是历史 `requireCdp=false` 副本，因此真实 `requireCdp=true` artifact 仍需在可达服务环境上重跑 external 才能生成。
- `2026-04-28` 同一 Phase H6 子线继续把 stable diagnostic 回放从“只显示当时的 CDP 要求”收口成“还能判断和本次命令是否一致”的入口：当前 `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 会额外打印 `External preflight CDP expectation for this invocation`，并在当前命令要求 `-RequireExternalCdpPreflight`、但 repo-local stable copy 仍来自旧的 `requireCdp=false` external 失败时，直接显示 `External preflight CDP mismatch note`，明确提示应重跑 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight` 刷新 stable copy，而不是继续从 `cdp skipped` 猜测 wrapper 分支。为了让本轮 `tsc --noEmit` 验证链重新恢复，还顺手修复了同一主线里已存在的两个前端语法断点：`web/src/components/modules/ai-automation/index.tsx` 中 profile 预览卡多余的 JSX 关闭片段，以及 `web/src/api/endpoints/ai-automation.ts` 里 `task artifacts/retry` 两条 endpoint 模板字符串缺失反引号的问题。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍是健康宿主上的真实 external/browser 证据，而不是 stable replay 可消费性或前端类型链本身。
- `2026-04-28` 同一 Phase H6 子线继续把 repo-local stable diagnostic 从“只有一个总副本”收口成“按 external CDP 预期分桶回放”的入口：`verify-ai-automation-learning-browser-smoke.ps1` 现在会在 external 失败时除了维护旧的 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json` 外，还按 `requireCdp` 额外写入 `latest-external-preflight-diagnostic-optional-cdp.json` 或 `latest-external-preflight-diagnostic-require-cdp.json`；`check-only` 会优先回放与本次命令匹配的副本，并额外打印 `External preflight stable copy note` 说明当前是“命中匹配副本”还是“缺少匹配副本后回退到最近可用副本”。同一轮还修复了 `check-only` 在回放完 stable copy 后继续误入共享 wrapper、导致去创建临时目录失败的问题，当前该模式会在输出摘要后直接 `exit 0`。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external artifact 采集。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 从“知道命中了哪个副本”收口成“还能直接看到其它 repo-local 变体是否存在”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在除了打印当前选中的 stable copy / legacy copy / selection note 之外，还会明确输出 `matching / alternate / legacy stable diagnostic copy status`，逐条说明副本是 `missing`、`present but could not be parsed`，还是 `present and parsed (recorded with requireCdp=true|false)`，并标记哪一份是 `selected for preview`。这样下一轮不需要再手工进 `build/verify-ai-automation-learning/` 查看文件名，也能直接判断当前是否仍缺 `latest-external-preflight-diagnostic-require-cdp.json`。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍是健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 stable replay 可见性不足。
- `2026-04-29` 同一 Phase H6 子线继续把 stable variant 可见性从“知道有没有文件”收口成“还能直接比较新鲜度”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会在 `matching / alternate / legacy stable diagnostic copy status` 行里直接追加该副本的 `checked at / age` 信息，因此不需要再打开 JSON 或比对文件时间，就能判断当前命中的 repo-local 副本是否过旧、备用副本是否更近，以及 `requireCdp=true` 变体是否仍然缺失。对应 `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 stable variant 新旧判断成本。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的“下一步怎么重跑”从通用建议收口成“当前调用专属命令”：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight preferred refresh command`，并在 `next steps` 里优先复用与本次调用完全对齐的 external 复跑命令。这样当本次命令已经带了 `-RequireExternalCdpPreflight` 时，回放结果会直接给出同样带该开关的 external 命令；若本次调用显式带了自定义 URL、CDP bootstrap 或超时参数，也会沿用同一 profile 生成建议，减少下一轮手工拼命令导致的诊断漂移。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍然是健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 replay 后不知道该重跑哪条命令。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的“看 age 自己判断”收口成“直接给 fresh/stale 结论”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会在 `matching / alternate / legacy stable diagnostic copy status` 行里追加 `fresh against 24h threshold` 或 `stale against Nh threshold`，默认按 24 小时阈值判断 repo-local stable diagnostic 是否足够新，同时允许通过 `-StableDiagnosticFreshnessThresholdHours` 临时收紧阈值做对照验证。这样下一轮在健康宿主复跑 external 之前，可以先直接判断当前副本是否已经“过旧到值得刷新”，不必只看 `age` 再人工换算。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 stable replay freshness 结论不足。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的 freshness 判断从“看多行状态自己拼结论”收口成“单条 note 直接说清楚当前该不该刷新”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight stable freshness note`，把当前选中的 matching/legacy fallback 副本是否仍 fresh、是否已经 stale、以及 alternate requirement-specific 副本是否与它同样新鲜，压缩成一条可直接消费的结论。这样当 `-RequireExternalCdpPreflight` 仍缺匹配副本时，回放结果会直接说明“当前只是 legacy fallback/alternate fallback，而且 stale 时应去健康宿主执行 preferred refresh command”；若当前匹配副本仍 fresh，也会明确说明暂时可以沿用。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 replay 后还要人工比较 fallback 新鲜度。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的 freshness 对比从“知道当前 fresh/stale，但还要自己比哪个副本更新”收口成“当前预览是不是 repo-local 最新可解析副本”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight stable freshest copy note`，直接说明当前 selected preview 是唯一可比较副本、已经是 freshest、还是只是与另一份副本同鲜度并列；若未来 selected preview 比其它 parseable copy 更旧，也会明确指出是哪一类副本更新得更晚。这样下一轮不需要再手工横向比较 `matching / alternate / legacy` 三行里的 age，就能先判断当前 replay 是否已经代表 repo-local 最晚一份证据。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍是健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 repo-local 副本之间的新旧比较成本。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的多条 note 再收口成最终决策摘要：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight decision summary`，直接给出“可以继续沿用 repo-local replay 做 blocked-host 排障”还是“当前 invocation 仍缺 requirement-specific fresh evidence，应去健康宿主或先暴露服务后执行 preferred refresh command”的结论。默认阈值下，匹配当前调用且仍 fresh 的 replay 会明确建议继续用本地 saved diagnostic 做 triage；`-RequireExternalCdpPreflight` 且仍缺匹配副本时，则会明确建议只把当前 fallback replay 当成临时证据，并在服务可达环境上刷新 requirement-specific artifact。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 replay 结论仍需人工拼装。
- `2026-04-29` 同一 Phase H6 子线继续把 legacy fallback replay 的 requirement-specific 旁证关系也并入最终决策层：当 `-RequireExternalCdpPreflight` 仍缺 matching 副本、只能回放 legacy fallback 时，`External preflight decision summary` 现在会直接补出 repo-local 是否已经存在 opposite CDP expectation 下的 alternate requirement-specific copy，以及它与当前 selected legacy fallback 是同鲜、更新还是更旧。这样接手人不再需要回头对照 `stable freshness note` 才知道“虽然当前没有 matching artifact，但仓库里其实已经有一份同样新/更晚的 requirement-specific 旁证”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、默认阈值与 `1h` 阈值下四次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口依旧只剩健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 repo-local replay 旁证关系还要人工拼接。
- `2026-04-29` 同一 Phase H6 子线继续把 repo-local stable replay 的“知道选中了哪份副本”收口成“还能直接看出当前仓库覆盖了哪类 `requireCdp` 变体”的入口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight stable coverage note`，直接说明 repo-local 当前是否同时拥有可解析的 `requireCdp=true / requireCdp=false` 两类稳定副本，还是只覆盖了其中一类、另一类仍缺失或不可解析。这样下一轮不需要再手工综合 `matching / alternate / legacy` 三行状态，光看 coverage note 就能先判断“当前是否已经具备 requirement-specific 双变体证据”，再决定要不要去健康宿主刷新 `requireCdp=true` artifact。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`runtime-win.ps1 -Action status`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 repo-local 变体覆盖关系还要人工判断。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的覆盖结论从“还要再读 `stable coverage note`”收口到最终决策层：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 的 `External preflight decision summary` 现在会直接补出 repo-local stable coverage 是否已经 complete，以及当前 invocation 还缺哪类 `requireCdp` 变体。这样默认 `requireCdp=false` 调用会明确显示“matching replay 仍 fresh，但仓库仍未捕获 parseable `requireCdp=true` variant”；`-RequireExternalCdpPreflight` 调用则会直接显示“当前 invocation coverage 仍 incomplete because no parseable `requireCdp=true` variant has been captured yet”，不需要再把 summary 与 coverage note 交叉阅读后才能确认下一步。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 repo-local coverage 结论还要人工拼装。
- `2026-04-29` 同一 Phase H6 子线继续把 final summary 的“coverage complete”终态语义也预先收口：`External preflight decision summary` 现在在 repo-local 同时具备 parseable `requireCdp=true / requireCdp=false` 两类副本时，会直接切到“Repo-local stable coverage is complete ... so only freshness and live reachability remain relevant now.” 的措辞，而不是仍然沿用缺变体时的 incomplete 口径。本轮除了默认与 `-RequireExternalCdpPreflight` 两次 learning smoke `check-only`、`verify-ai-automation-learning-focus.mjs` 与 `tsc --noEmit` 外，还通过临时加载 helper 分支、喂入双变体 parseable 的合成状态对象，验证了 coverage-complete 分支会按预期产出这句 summary，且不需要在仓库里留下任何伪造 stable artifact。当前剩余缺口仍稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据；一旦该 artifact 补齐，接手人只看 final summary 就能知道当前只剩 freshness/live reachability，而不是再被“variant 缺失”误导。
- `2026-04-29` 同一 Phase H6 子线继续把 stable replay 的“看完 final summary 还要自己提炼下一步动作”收口成动作级出口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight recommended action`。当当前 invocation 已命中 matching 副本且仍 fresh 时，它会直接建议继续用 matched repo-local replay 做 blocked-host triage；当 `-RequireExternalCdpPreflight` 仍缺 matching 副本时，它会明确说明当前 selected legacy fallback copy 只能作为 fallback-only context，并提示去健康宿主或先暴露服务后执行 preferred refresh command 来补 requirement-specific artifact。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口仍只剩健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 blocked-host replay 还要人工再提炼动作。
- `2026-04-29` 同一 Phase H6 子线继续把 blocked-host replay 的动作建议再收口成稳定分类标签：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight recommended action class`。默认 `requireCdp=false` 且命中 fresh matching 副本时会输出 `matched_replay_ready`；`-RequireExternalCdpPreflight` 但仍缺 matching requirement-specific 副本时会输出 `fallback_only_refresh_required`。这样下一轮不必先通读整句自然语言建议，就能先按分类判断“继续沿 matched replay 做 triage”还是“直接去健康宿主补 requirement-specific artifact”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only` 与 `tsc --noEmit` 已同步通过；当前剩余缺口仍稳定收敛为健康宿主上的真实 `requireCdp=true` external/browser 证据，而不是 replay 建议缺少稳定分档。
- `2026-04-29` 同一 Phase H6 子线已把“真实 `external + self-start + require-cdp` 失败后无法沉淀 repo-local requirement-specific 证据”的最后主阻塞收口：`verify-ai-automation-learning-browser-smoke.ps1` 现在会在共享 CDP smoke wrapper 抛出 `CDP diagnostic file:` 但没有新的 `external-preflight-diagnostic.json` 可直接发布时，自动从相邻 external preflight JSON 与异常消息里的 CDP 摘要桥接生成 `external-preflight-diagnostic.cdp-bridge.json`，再继续发布到 repo-local stable copy。这样本机这次真实 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -SelfStartServices` 虽然仍按预期卡在 `page_bootstrap_timeout_attached_session`，但已经成功落盘 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic-require-cdp.json`，后续 sequential `check-only -RequireExternalCdpPreflight` 会直接回放这份 fresh requirement-specific 诊断，并显示 repo-local stable coverage 已 complete、当前真实 primary blocking check 已从“缺少 requireCdp=true artifact”收敛为“backend/frontend 已可达，但 attached-session 下 `Runtime.enable / Page.enable` 仍超时”。同一轮 sequential 默认 `check-only` 也已确认仓库现在同时拥有 parseable `requireCdp=true / requireCdp=false` 两类 stable variant；当前下一步更适合继续在同主线下刷新 optional profile 到同一运行层，或直接聚焦 CDP bootstrap 策略，而不是再回头解决 artifact 缺失。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、真实 external self-start require-cdp 复跑、两次 sequential `check-only`、`runtime-win.ps1 -Action status`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘。
- `2026-04-29` 同一 Phase H6 子线继续把“下一轮先刷新哪个 invocation profile”从手工对照两次 `check-only` 输出，收口成一条直接可消费的对齐提示：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight invocation profile alignment note`，直接比较 matched replay 与 opposite replay 是否已经落到同一服务可达层。当前 repo-local 结果下，默认 `requireCdp=false` replay 会明确提示“它仍停在 `preflight_failed / backend,frontend`，而 opposite `requireCdp=true` replay 已经到 `cdp_smoke_failed / cdp`，因此下一轮应先刷新 `requireCdp=false` external profile，再谈 CDP/bootstrap 对比”；反向查看 `-RequireExternalCdpPreflight` 时，也会直接提示 opposite `requireCdp=false` profile 仍落后于当前运行层。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口因此进一步收敛为“先把 optional profile 刷新到 backend/frontend 已可达的同一层，再决定是否继续比较 attached-session CDP bootstrap 超时策略”。
- `2026-04-29` 同一 Phase H6 子线已完成上述 optional profile 刷新，并把最终 replay 语义收口到“两个 invocation profile 已对齐、下一轮直接聚焦 attached-session CDP bootstrap”这一层：真实 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -SelfStartServices` 虽然仍按预期失败在 `page_bootstrap_timeout_attached_session`，但已成功把 `latest-external-preflight-diagnostic-optional-cdp.json` 刷新到与 `requireCdp=true` 相同的 `cdp_smoke_failed / failed checks cdp` 运行层。基于这份 fresh artifact，`check-only` 现在会把 `External preflight recommended action class` 切到新的 `aligned_cdp_bootstrap_focus`，并让 `decision summary / recommended action / invocation profile alignment note` 统一说明“两条 profile 都已到 backend/frontend reachable，后续 live rerun 只需要继续比较 attached-session CDP bootstrap，而不是回到 service reachability 或 artifact coverage”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、一次真实 optional external refresh、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前剩余缺口因此继续收敛为 attached-session 下 `Runtime.enable / Page.enable` 超时本身，而不是 invocation profile 未对齐。
- `2026-04-29` 同一 Phase H6 子线继续把“已对齐后下一条 live rerun 该怎么拼”也收口成稳定出口：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会在 `aligned_cdp_bootstrap_focus` 场景下额外打印 `External preflight CDP bootstrap comparison command`，直接给出与当前 invocation 保持同一 profile、只切换 `-CdpBootstrapCommandOrder` 的 external 比较命令。默认 profile 现在会输出 `... -Mode external -UseHostFriendlyExternalDefaults -CdpBootstrapCommandOrder 'page-lifecycle-runtime'`，`-RequireExternalCdpPreflight` profile 则会输出同样带 `-RequireExternalCdpPreflight` 的比较命令，因此下一轮不需要再手工从 `decision summary / recommended action / hints` 里拼参数，就能直接拿现成命令去对照 `runtime-page-lifecycle` 与 `page-lifecycle-runtime` 的 attached-session bootstrap 表现。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前剩余缺口继续收敛为是否值得在本机执行这条比较命令以进一步确认 `Runtime.enable / Page.enable` timeout 的宿主特征，而不是 replay 输出还需要人工转译成命令。
- `2026-04-29` 同一 Phase H6 子线已继续把上述 comparison command 收口到“当前默认停驻主机上也能直接复现同一失败层级”的入口：本轮先实跑了默认 optional profile 的两条 live 命令，确认单独复制 `External preflight CDP bootstrap comparison command` 会因为项目默认停驻而先掉回 `preflight_failed / failed checks frontend`，只有追加 `-SelfStartServices` 才会再次落到与 repo-local replay 对齐的 `cdp_smoke_failed / failed checks cdp`，并在 `page-lifecycle-runtime` 顺序下稳定复现 attached-session bootstrap 超时。基于这个结果，`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在在输出原 comparison/refresh command 的同时，还会额外打印 `External preflight self-start CDP bootstrap comparison command` 与 `External preflight self-start refresh command`，让当前这台默认停驻宿主上的下一轮 live rerun 不必再手工补 `-SelfStartServices`。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、一次按预期失败的 non-self-start external comparison、一次按预期失败并落到 `cdp_smoke_failed` 的 self-start external comparison、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前剩余缺口因此进一步收敛为“是否还值得在 `-SelfStartServices` 同层级下继续放大 `-CdpCommandTimeoutMs` 对照”，而不是 replay 输出仍缺可直接执行的 host-ready 命令。
- `2026-04-29` 同一 Phase H6 子线已把上述“是否还值得继续放大 `-CdpCommandTimeoutMs`”也收口成 wrapper 级固定出口：本轮先在默认 optional 与 `-RequireExternalCdpPreflight` 两条 profile 下，真实执行 `self-start + page-lifecycle-runtime` 对照，确认 bootstrap order 切到反向顺序后仍旧稳定停在同一 `page_bootstrap_timeout_attached_session / Page.enable -> Page.setLifecycleEventsEnabled -> Runtime.enable` 层，没有出现新的失败等级。基于这个结果，`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command`，直接给出下一条 bounded rerun 所需的 `-CdpCommandTimeoutMs 45000 + -NodeSmokeTimeoutSeconds 155` 组合，不必再由下一轮手工换算超时参数。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 self-start external comparison、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、`runtime-win.ps1 -Action status` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前剩余缺口因此继续稳定收敛为“是否还要做最后一轮更长 timeout 对照”，而不是 bootstrap order 或 service reachability 仍未分类清楚。
- `2026-04-29` 同一 Phase H6 子线继续把这条 bounded timeout lane 再收紧到“可停手定性”的前一层：本轮分别真实执行了 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 45000ms` live rerun，确认它们都仍稳定停在同一 `page_bootstrap_timeout_attached_session / Runtime.enable -> Page.enable -> Page.setLifecycleEventsEnabled` 层，没有因为 profile 差异而出现新的运行层。基于这份 fresh `45000ms` matching artifact，`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会自动把 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command` 升级到下一档 `-CdpCommandTimeoutMs '60000' -NodeSmokeTimeoutSeconds '200'`，同时把 `External preflight recommended action class` 切到新的 `attached_session_bootstrap_blocker_candidate`，并让 `decision summary / recommended action` 明确说明“当前宿主已经具备 attached-session bootstrap blocker 候选证据，只剩最后一轮 bounded timeout 对照是否还值得执行”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 `45000ms` self-start external rerun、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs` 与 `tsc --noEmit` 已同步通过/按预期失败后落盘；当前剩余缺口因此继续收敛为“是否再跑一轮 `60000ms` host-level 对照，还是直接把该宿主定性为 attached-session bootstrap blocker”，而不是 replay 输出或 profile 对齐仍有不确定性。
- `2026-04-29` 同一 Phase H6 子线已完成上述最后一轮 `60000ms` host-level 对照，并把 replay 语义正式收口到“宿主 blocker 已确认”：本轮分别真实执行了 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 60000ms` external rerun，结果仍稳定停在同一 `page_bootstrap_timeout_attached_session / Runtime.enable -> Page.enable -> Page.setLifecycleEventsEnabled` 层，没有出现新的失败等级、没有跨到新的运行层，也没有出现 profile 差异。基于这份 fresh `60000ms` matching artifact，`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会把 `External preflight recommended action class` 切到新的 `attached_session_bootstrap_blocker_confirmed`，并让 `decision summary / recommended action` 直接说明“当前宿主已确认是 attached-session bootstrap blocker，应停止继续调这条 wrapper 参数线并记录宿主级阻塞”；同时在这个 confirmed 分支下不再继续输出更长 timeout 的 comparison command，避免下一轮重新掉回 `90000ms` 调参空转。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次真实 `60000ms` self-start external rerun、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过/按预期失败后落盘；当前这条主线的剩余工作因此收敛为“保留 blocker 结论并切换到不同执行路径或不同宿主”，而不是继续在同一 attached-session timeout lane 放大参数。
- `2026-04-29` 同一 Phase H6 子线继续把 confirmed blocker 结论从“知道该停手”收口到“知道下一条该试什么”：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在在 `attached_session_bootstrap_blocker_confirmed` 场景下会额外输出 `External preflight alternate execution path command` 与 `External preflight self-start alternate execution path command`，默认把下一条 live rerun 固化为保留当前 invocation profile、保留 `60000ms / 200s` 超时预算、仅切换 `-CdpPageBootstrapStrategy 'json-new'` 的 external 命令。这样接手人不需要再从 blocked-host 结论里手工拼“不同执行路径”的命令，也不会因为 `check-only` 默认参数回落到 `30000ms` 而丢失当前 host-level blocker 的证据层级。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前这条主线的下一步因此更明确收敛为“先用现成 alternate execution path command 比较 `json-new` 路径，若仍无新增运行层，再切宿主”，而不是停在抽象的“换条路径试试”。
- `2026-04-29` 同一 Phase H6 子线继续把 `attached-session` invocation 对 `json-new` repo-local 证据的消费语义收口到 strategy-aware 状态：当前 `build/verify-ai-automation-learning/` 已存在 `latest-external-preflight-diagnostic-{optional-cdp|require-cdp}-json-new.json`，但 `attached-session` 的 `check-only` 回放此前仍把它们统一当成 generic fallback，并继续给出“优先 refresh matching copy”的动作，容易把下一轮带回错误路径。本轮已把 wrapper 调整为在 `matching-generic / alternate-generic` 状态行中显式标注“saved with alternate page-bootstrap strategy 'json-new'”，并让 `decision summary / recommended action / recommended action class` 在 strategy-specific `attached-session` 副本缺失、但同一 `requireCdp` 的 `json-new` fresh 副本已存在时改为 `fallback_replay_ready`：先沿现有 `json-new` 证据做 blocked-host triage，优先执行 `alternate execution path command`，只有在明确需要 `attached-session` strategy-specific saved diagnostic 时才再走 preferred refresh command。同时这条 alternate execution 命令现在也会继承已保存 `json-new` 诊断里的 `60000ms / 200s` 预算，不再退回默认 `30000ms / 110s`。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 sequential `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 已同步通过；当前下一轮更适合直接使用现成 `json-new` alternate execution 命令或切换宿主，而不是在 `attached-session` invocation 下误判为“只剩 refresh matching copy”。
- `2026-04-29` 同一 Phase H6 子线已继续把 learning smoke stable replay 从“只按 requireCdp 选稳定副本”收口到“同时区分 page-bootstrap strategy 与 generic fallback”：当前 wrapper 会把 stable artifact 落成 `latest-external-preflight-diagnostic-{require-cdp|optional-cdp}-{attached-session|json-new}.json`，并在 strategy-specific 副本缺失时明确回退到 `same-expectation fallback copy`，不再把 `json-new` 证据误报成 `attached-session` invocation 的 matching requirement-specific artifact。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过 repo-local parser / `check-only` / `verify-ai-automation-learning-focus.mjs` / `tsc --noEmit` 验证；当前 default `attached-session` invocation 已明确显示“只有 generic fallback、缺少 strategy-specific stable copy”，不会再把错误层级继续混淆。
- `2026-04-29` 同一 Phase H6 子线还继续把 `json-new` live path 的宿主结论正式升级为新的 strategy-specific blocker：本轮真实执行了 optional `json-new + self-start + 60000ms` external rerun，补齐了 `latest-external-preflight-diagnostic-optional-cdp-json-new.json`，从而让 repo-local `json-new` 双 profile 覆盖闭环。当前 `check-only -CdpPageBootstrapStrategy 'json-new'` 已可直接输出 `External preflight recommended action class: page_bootstrap_strategy_blocker_confirmed`，并明确说明 optional / require-cdp 两条 invocation profile 都会在本机稳定停在 `page_bootstrap_timeout_json_new`。因此这条 H6 主线的下一步不再是继续在本机加 timeout 或比较 bootstrap order，而是换宿主或换真正不同的执行路径。
- `2026-04-29` 同一 Phase H6 子线继续把 strategy-aware replay 的“freshest copy”提示收口到不误导 handoff 的语义：当前 `attached-session` invocation 在缺少 strategy-specific stable copy、但已选中 `same-expectation fallback copy` 时，如果 repo-local 里同时存在更晚写入的 `opposite-expectation` 或其他 page-bootstrap strategy 副本，`check-only` 不再只说“selected preview is older than the freshest parseable copy”，而会进一步明确这些 fresher 副本只是比较材料，不能替换当前 invocation 的最佳 same-expectation replay。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-ai-automation-learning-focus.mjs` 已同步更新，并通过默认/`-RequireExternalCdpPreflight` 两次 `check-only`、focus 守护与 `tsc --noEmit`；当前下一轮在 `fallback_replay_ready` 场景下可直接沿现有 same-expectation replay 或 `alternate execution path command` 继续，而不是误把 fresher opposite-expectation copy 当成更优主证据。
- `2026-04-29` 同一 Phase H6 子线继续把 blocked-host 交接结论从“给出 alternate execution path command”收口到“明确当前宿主不该再跑这条 live 命令”：`verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外输出 `External preflight host handoff note`。当 `attached-session` 或某个 page-bootstrap strategy 已经通过 repo-local saved diagnostics 被确认成 host-level blocker 时，这条 note 会直接说明“当前宿主不要再继续重跑同一路径，应换宿主或换真正不同的执行路径”；而在当前 `fallback_replay_ready + json-new same-expectation fallback` 场景下，它还会进一步识别 `matching-generic / alternate-generic` 这对 `json-new` stable 副本，并明确提示“repo-local 已经证明本机对 `json-new` 也稳定停在 `page_bootstrap_timeout_json_new`，因此 alternate execution path command 只值得拿到别的宿主上复跑，在本机继续 live rerun 基本不会新增证据”。对应 `scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、两次 `check-only`、PowerShell parser 校验、focus 守护、`tsc --noEmit` 与 `git diff --check` 已同步通过；下一轮因此可以直接从“换宿主或切真正不同路径”起步，而不是重新判断当前宿主是否还值得试一次 `json-new` live rerun。

已确认口径：

- AI 自动化中心是顶层栏目，不是设置页里的附属按钮。
- AI 生成的分组、渠道识别、价格识别、模型归类等结果必须保存为 AI Profile。
- 手动配置和 AI Profile 必须同时保留，AI 任务不能静默覆盖 `channels`、`groups`、`group_items`、`llm_infos`、`route_target_overrides`。
- 设置页必须提供 `manual / ai_profile` 的显式切换。
- 动态路由 AI 学习是本地可解释学习，只影响运行时推荐，不永久改写用户配置。
- `profile_activate` 与 `snapshot_guard` 已按“受保护动作”接入任务执行链：前者只在任务成功生成新 Profile 后激活该 Profile，后者落独立 AI 任务快照目录，不再污染导入回滚快照链。

主入口文档：

- [AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md)
- [AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md)

---

## 1. 本次核对所依据的文档

本次判断主要依据以下文档：

1. `AGENTS.md`
2. `docs/PLAN.md`
3. `docs/WECHAT_ARTICLE.md`
4. `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
5. `docs/LLM-Gateway-Refactor-Plan.success-cases.zh-CN.md`
6. `docs/worklog/2026-04-14-backend-task-01-channel-key-modes.md`
7. `docs/worklog/2026-04-14-backend-task-02-advanced-failover-runtime.md`

说明：

- `docs/PLAN.md` 更像第一阶段需求清单，主要覆盖“渠道创建体验 + API 文档入口”。
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 是当前唯一 canonical plan，应作为最高优先级判断标准。
- 两份 `worklog` 说明当前已经实施到 canonical plan 的哪一段。

---

## 2. 总体判断

当前仓库不是“从头未做”，而是一个已经做到中段、但明显还没收口的重构分支。

总体完成度判断：

- 第一阶段旧需求：大部分已完成。
- canonical plan Phase 0-2：基本完成。
- canonical plan Phase 3：已进入实现，但只完成了前半段。
- canonical plan Phase 4-7：都有启动痕迹，但多数还是“字段、接口、UI 骨架已接上”，离验收标准还有明显距离。

一句话总结：

> 这个项目现在最像“核心骨架已经搭起来了，但很多能力仍停留在 phase-1/半成品状态，而且缺少一次完整的编译与验收收口”。

---

## 3. 已完成的要求

### 3.1 完成得比较好的部分

#### A. 第一阶段渠道体验增强

对应来源：

- `docs/PLAN.md`
- `docs/WECHAT_ARTICLE.md`
- `README_zh.md`

判断：

- 渠道预设库已经接入。
- GitHub Copilot OAuth Device Flow 已接入。
- Antigravity OAuth Web Flow 已接入。
- 创建渠道前测试模型的 UI 已接入。
- API 基础地址设置、curl 示例、CC Switch 深链能力已经落地。

评价：

- 这部分不是只停在文档里，README、文章、前端入口、接口链路都能互相印证。
- 这是当前仓库里最“完整成型”的一块。

#### B. key 管理模式与基础字段链路

对应来源：

- canonical plan 4.1.2 / 5.1 / 6.1
- `worklog 01`

判断：

- `key_management_mode` 已加入 `Channel`。
- `key_routing_policy` 已加入 `Channel`。
- `allowed_models` 已加入 `ChannelKey`。
- `source_type` 已加入 `ChannelKey`。
- 前后端创建、编辑、更新链路都已经接上。
- migration `003`、`007`、`009` 已存在。

评价：

- 这部分“字段贯通”做得比较完整。
- 至少从数据模型、API 类型、表单、保存逻辑上看，不是只改一半。

#### C. group 高级 failover 参数链路

对应来源：

- canonical plan 6.3 / 7.1.4 / 7.1.5
- `worklog 02`

判断：

- `retry_rounds`
- `retry_delay_ms`
- `failover_window_sec`
- `race_after_fails`
- `race_concurrency`

以上字段已经进入：

- model
- op
- migration
- frontend endpoint type
- group create/edit UI

评价：

- 这一块也属于“配置链路基本打通”。
- 文档与代码方向一致。

#### D. 里程碑 2 的一部分观测能力

对应来源：

- canonical plan 9.4 / 9.5 / milestone 2

判断：

- 日志导出接口已新增。
- stats token breakdown 接口已新增。
- 首页新增了 `TokenBreakdown` 组件。
- 设置页新增了备份/导入区域。

评价：

- 这部分已经进入“用户可见层”，不是只有后端字段。
- 属于可见成果比较明确的一组改动。

---

## 4. 已完成但完成得不好的部分

这部分不是“完全没做”，而是“做了，但离 canonical plan 要求还差不少”。

### 4.1 Key 管理模式只做到了字段和基础路由，没做到最终 UI 形态

canonical plan 要求：

- 同一渠道内显式管理多个 key
- key 应是独立子项/子卡片
- `pooled` / `classified` 必须清晰可见
- 分组页左栏最终要做到 `provider -> key -> model`

当前状态：

- 现在只是把 `key_management_mode`、`key_routing_policy`、`allowed_models`、`source_type` 塞进了现有表单。
- 仍然是偏“平铺输入框”的形态。
- 还没达到 canonical plan 要求的“key 子卡片 / 分层展示 / 可折叠 / 易理解”。

评价：

- 后端准备得比前端更成熟。
- 这块目前是“能配置”，但还不算“好用”。

### 4.2 advanced failover runtime 只完成了 phase-1

canonical plan 要求：

- 360 秒窗口
- 连续失败后进入并发竞速
- 受预算控制
- paid/metered 默认禁止并发竞速
- 不是简单 first-wins
- 应按用户优先级裁决
- 一旦确定输出，其他请求尽快取消

当前状态：

- 已有多轮重试。
- 已有 failover window。
- 已有 `race_after_fails` 与 `race_concurrency`。
- 已有非流式竞速 fallback。
- 已有部分 `source_type / billing_mode / probe_policy` 门控。

但问题是：

- 还不是最终裁决规则。
- 还是 phase-1 winner rule。
- streaming race 还没做。
- route-target 策略矩阵没有完整落地。
- paid/free 的默认策略只是部分接入。

评价：

- 方向对。
- 但从 canonical plan 标准看，只能算“半完成”。

### 4.3 备份/导入已经开始做，但离 canonical plan 差距很大

canonical plan 要求：

- 版本化快照
- checksum
- dry-run
- 差异分析
- replace / merge / skip / map
- 导入后验证
- 回滚
- 默认导出全量迁移快照，包含可直接恢复所需的明文凭据

当前状态：

- 已有 `manifest`
- 已有 checksum
- 已有 dry-run
- 已有兼容性报告
- 已有部分 route preview warning

但明显不足：

- 仍然是 `json` 导出，不是完整 snapshot 体系。
- 没有 replace / merge / skip / map 这类冲突模式。
- 没有真正的回滚。
- 没有导入后健康检查。
- 没有导入映射表编辑。
- 当前导出内容按最新产品决策默认包含 `ChannelKey/APIKey` 等恢复所需凭据；后续收口目标是不再把这一行为描述成“默认脱敏导出”，而不是回退到旧口径。
- `manifest.contains_secrets` 当前写成 `false`，与真实导出内容不一致。

评价：

- 这是“启动了，但还没有达到可放心交付”的典型代表。

### 4.4 首页增强只完成了一部分

canonical plan 要求首页展示：

1. 总 token
2. 按 channel
3. 按 provider
4. 按 model
5. 官方价格统计
6. gateway 价格统计
7. probe cost
8. 成功率/失败率
9. 熔断摘要
10. 最近探测状态摘要

当前状态：

- 总 token 有。
- 按 channel 有。
- 按 model 有。

但仍缺：

- 按 provider 拆分
- 官方价 / gateway 价汇总展示
- probe cost
- breaker 摘要
- probe 状态摘要

评价：

- 首页增强已经开始，但只完成了“token breakdown 基础版”。

### 4.5 价格系统已扩字段，但归一化体系还没真正完成

canonical plan 要求：

- canonical name
- 官方价 / gateway 价双视图
- alias / manual mapping
- fallback 命中顺序
- 家族/版本/variant 识别

当前状态：

- 已新增 `canonical_name`
- 已新增官方价格字段
- 已新增 `billing_mode / probe_policy` 等模型元数据
- 已新增 `internal/llmname/`

但仍缺：

- alias/manual mapping 体系
- 完整 fallback 命中链
- UI 侧双视图体验是否完整仍不足
- “价格不能驱动路由”虽然文档写清了，但还缺验收层面的测试闭环

评价：

- 这块属于“地基在铺，但房子还没盖完”。

---

## 5. 尚未完成的主要要求

### 5.1 严格达到 canonical plan 的最终 UI 目标

未完成内容包括：

- 渠道页 key 子卡片化
- pooled/classified 的清晰可视化
- 分组页左栏 `provider -> key -> model` 分段展示
- 渠道搜索/筛选增强
- 模型页更完整的搜索/筛选
  - `2026-04-23` 已补齐“模型名称 + 规范名称”双命中搜索契约；当前剩余缺口主要收敛为更细的筛选维度与浏览器级移动端验收。
- 手机端针对这些新增能力的专项验收

### 5.2 route-target 级策略没有真正做完

未完成内容包括：

- paid / metered / free / public 默认策略矩阵完整落地
- 并发预算分层
- 导入后模拟路由验证
- relay 运行时调优不覆盖用户优先级的完整验证
- 探测/恢复/半开策略的完整实现

### 5.3 完整备份/导入/迁移适配体系没有做完

未完成内容包括：

- 冲突处理模式
- 回滚到上一个 snapshot
- 映射表编辑
- 导入后验证
- 部分恢复
- 默认全量快照 / 显式脱敏导出契约的最终验收

### 5.4 验收与构建闭环没有做完

canonical plan 里明确要求：

- `go test`
- 前端 build
- Docker build
- compose up
- 流式回归测试

当前状态：

- 仓库里没有看到这轮改动对应的完整验收记录。
- 当前环境下也无法直接复跑：
  - 本机 `go` 不在 PATH
  - 本机 `pnpm` 未正确装好链接

结论：

- 当前分支最大的现实问题，不是“完全没做”，而是“缺一次系统验收收口”。

---

## 6. 当前实现里我认为做得好的点

1. 需求文档和代码方向基本一致，没有完全跑偏。
2. worklog 写得比较清楚，能够说明当前做到哪一阶段。
3. schema/migration/前后端字段链路推进得比较连续。
4. 严格权重序列没有退化成 weighted random，这一点方向是对的。
5. 日志导出、token breakdown、dry-run import 这些增强都属于高价值功能，不是低价值 UI 装饰。

---

## 7. 当前实现里我认为做得不好的点

1. 有明显“写到一半没收尾”的痕迹。
2. 有些改动只完成了字段和表单，但没完成 canonical plan 的最终用户体验。
3. advanced failover 仍明显是 phase-1，实现层自己也承认未完成。
4. 备份导出在安全语义上存在明显问题：当前导出内容与 manifest 标记不一致。
5. 缺少编译、测试、Docker、回归这类收口动作。
6. 当前未提交改动里出现过明显的结构错误，说明这轮代码还没经过完整 build 检查。

---

## 8. 结论版状态表

| 需求主题 | 状态 | 判断 |
|---|---|---|
| 渠道创建体验增强 | 已完成 | 完成度高 |
| 供应商模板 / OAuth / 模型测试 / CC Switch | 已完成 | 完成度高 |
| key 管理模式字段链路 | 已完成 | 完成度高 |
| key 最终 UI 形态 | 未完成 | 仍是基础表单态 |
| key 路由 phase-1 | 已完成 | 但还未达到最终规则 |
| group 多轮重试 / failover 参数链路 | 已完成 | 完成度高 |
| advanced failover runtime 最终版 | 部分完成 | 当前只有 phase-1 |
| paid/free route-target 策略矩阵 | 部分完成 | 只接入了部分门控 |
| 动态探测 / 半开恢复 / relay 运行时调优 + 每日动态摘要扫描 | 部分完成 | 已收口为真实语义，但完整环境验收仍未完成 |
| 价格 canonical/official 字段扩展 | 部分完成 | 地基已铺 |
| 完整价格归一化体系 | 未完成 | 还没闭环 |
| 日志导出 | 已完成 | 方向正确 |
| 首页 token breakdown | 部分完成 | channel/model 有，provider/价格/probe 摘要无 |
| 项目级备份导出 | 部分完成 | 有 manifest + checksum |
| dry-run 导入预检 | 部分完成 | 已有基础版 |
| 冲突模式 / 回滚 / 映射编辑 / 导入后验证 | 未完成 | 还没做完 |
| 手机端最终验收 | 未完成 | 还缺完整验证 |
| go test / web build / docker 验收 | 未完成 | 没有收口证据 |

---

## 9. 我建议的后续实施顺序

### Phase A：先做“工程收口”

目标：

- 先让当前分支重新进入“可编译、可验证”状态。

内容：

1. 修掉当前未提交代码中的结构错误、类型错误、明显并发问题。
2. 补齐最小编译链。
3. 记录一次基础验收结果。

原因：

- 不先收口，后面继续堆功能只会越来越难接手。

### Phase B：完成 milestone 1 的剩余硬规则

目标：

- 把 route/key/group 的 phase-1 做到 milestone 1 验收标准。

内容：

1. 明确 `fill_priority` 与 `priority_order` 的真实差异。
2. 确认每次请求从第 1 个开始的严格顺序语义。
3. 完成 key 失败不扩散到无关模型的验证。
4. 补齐并发竞速路径的取消和记录安全性。

### Phase C：把 milestone 2 做扎实

目标：

- 先完成“可观测与排障”，因为这会直接提高后续调试效率。

内容：

1. 日志导出格式稳定化。
2. attempts 链完整性检查。
3. 首页 token/provider/model 拆分补齐。

### Phase D：再回到价格与备份导入

目标：

- 把 Phase 5 和 Phase 6 从“字段已加”推进到“功能闭环”。

内容：

1. 价格归一化完整链路。
2. 官方价 / gateway 价双视图。
3. 备份导出安全策略。
4. dry-run -> diff -> apply -> verify -> rollback 的真正闭环。

---

## 10. 当前建议的下一份施工文档目标

接下来最适合继续写的，不是再扩一份大而全的 plan，而是写一份更偏执行的施工清单，建议拆成三份：

1. `docs/worklog/2026-04-14-audit-and-stabilization.md`
   - 当前审计发现
   - 当前明显断点
   - 第一批必须收口项

2. `docs/worklog/2026-04-14-milestone-1-gap-closure.md`
   - key/group/failover 剩余硬规则
   - 要补的测试

3. `docs/worklog/2026-04-14-backup-import-gap-closure.md`
   - 备份导出安全问题
   - 导入/映射/回滚闭环

---

## 11. 最终判断

当前项目最准确的状态不是：

补充状态（2026-05-08）：Docker Hub 安装方案已正式废弃并进入文档主线。后续安装与发布口径统一为 GHCR 官方镜像、显式私有 / 代理镜像，以及安装脚本在 GHCR 拉取失败时的源码支撑 Docker 构建；不得再把 Docker Hub 作为默认或官方推荐来源。

- “已经完成”
- 也不是“还没开始”

而是：

> 关键骨架已经做出来了，第一批用户价值功能也已经有成果，但 canonical plan 意义上的最终重构还没有完成，尤其缺少 UI 最终形态、导入回滚闭环、完整策略矩阵和系统验收收口。备份主链可用不等于高级迁移 fully done；高级迁移能力仍需继续补齐。

- 2026-04-25 已继续收口备份 helper 链中的模型策略 warning 句级本地化，并完成 locale provider 的类型解阻；当前 `backup-logic.ts` 已能把模型策略变更与并发成本提示转换成可读的中文/繁中/日文详情文本，`locale.tsx` 的 AI 自动化 locale 结构问题也已消除，并通过 repo-local `verify-backup-logic.mjs` 与 `verify-locale-consistency.mjs` 稳定验证。当前备份主线的剩余前端工作已收敛为 `Backup.tsx` 页面细节契约清理、后续浏览器级证据补齐，以及高级迁移能力继续保持“部分完成”状态；本轮又为高级迁移折叠区补上了稳定的 `backup-remaining-migration-panel` 锚点，并把 `backup-page` 根锚点与 `backup-remaining-migration-section-*` 内层锚点一并纳入 `scripts/verify-backup-component.cjs` 的 map/replace 分支断言，同时补齐了 `Backup.test.tsx` 的 dry-run/apply 根锚点断言，下一轮可以直接基于这些 selector 继续补 browser-grade 证据。


