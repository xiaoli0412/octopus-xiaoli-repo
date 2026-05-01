# 2026-04-24 Phase G Channel Key Route-Target Visibility Closure

## 1. 任务信息

- 任务名称：渠道详情 key 级 route-target 可见性收口
- 日期：`2026-04-24`
- 当前阶段：`Phase G screenshot-first UI closure`
- 对应 milestone：`Phase G / 9.1.1 渠道 key 级观测字段补强`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1`、`9.1.1`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.2`、`1.3`、`1.4`、`11.2`、`11.4` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-24-phase-g-channel-page-browser-smoke-closure.md`
  - `docs/worklog/2026-04-23-phase-g-channel-key-readiness-summary.md`
- 本次已盘点本地资源：canonical plan、`CURRENT_STATUS_AND_PLAN.zh-CN.md`、`FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`ENV_READY_AND_NEXT_PLAN.zh-CN.md`、automation memory、`CardContent.tsx`、`channel.ts`、`verify-channel-presentation.mjs`
- 本次使用了哪些本地 resources / skills / 记忆上下文：上述文档、脚本与 automation memory
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、强耦合于单个组件和验证脚本

## 3. 本次微计划

- 当前主线：`Phase G screenshot-first UI closure / channel key observability`
- 当前阶段：渠道总页 selector/no-browser 合同已收口后，继续补 `9.1.1` key 级信息
- 本轮核心任务：把现有 `route-target override` 数据真正挂到单 key 详情里，补齐 key 级计费/探测可见性
- 本轮配套任务：同步 `keyFilter` 搜索维度、no-browser 验证脚本和状态文档
- 预期验证方式：`runtime-win status`、`verify-channel-presentation`、`verify-channel-create-flow`、前端 `tsc --noEmit`
- 完成判定标准：单 key 展开区可展示该 key 的 route-target override 摘要，搜索可命中对应摘要字段，验证脚本通过，状态文档和下一轮入口已写回

## 4. 本次硬规则

- 只停留在 `Phase G screenshot-first` 渠道主线，不扩散到后端 schema 或其他页面
- 只使用现有后端已提供的 `route-target override` 数据，不伪造 quota、熔断实时值或 priority/weight 字段
- 页面事实与状态文档必须同步，不能继续沿用“本机 browser pass 已闭环”的过度表述

## 5. 本次禁止事项

- 不改后端接口契约
- 不擅自增加新的 key 级后端字段
- 不把宿主 Edge/CDP 阻塞写成页面能力已通过

## 6. 本次回滚点

- 回滚 `web/src/components/modules/channel/CardContent.tsx`
- 回滚 `scripts/verify-channel-presentation.mjs`
- 回滚四语 locale 中新增的 route-target key 级文案
- 回滚本次状态文档和 worklog 同步

## 7. 实施范围

- 先改数据语义还是先改 UI：先补渠道详情 key 展开区 UI，再补验证脚本和文档同步
- 受影响前端模块：
  - `web/src/components/modules/channel/CardContent.tsx`
  - `web/public/locale/{zh-Hans,zh-Hant,en,ja}.json`
  - `scripts/verify-channel-presentation.mjs`
- 受影响文档：
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- 是否影响旧数据：否
- 是否影响旧接口或旧行为：否，属于已有详情数据的可见性补强与搜索维度补强

## 8. 实际完成项

1. 在 `CardContent.tsx` 中新增 `routeTargetOverridesByKeyId` 聚合，把已有 `route-target override` 行按 `channel_key_id` 归组。
2. 单 key 折叠头新增“路由覆盖条数”摘要，展开区新增 key 专属 route-target 摘要块，直接显示该 key 的模型、计费模式、探测策略、间隔与并发信息。
3. `keyFilter` 现在除原有密钥、备注、来源、模型、状态码外，也能命中 route-target override 的模型名、计费模式、探测策略与数字字段。
4. 四语 locale 补齐 `routeTargetOverridesCount`、`routeTargetKeyEmpty`、`routeTargetKeyPreviewMore` 文案，避免 key 级 route-target 信息再次回退成英文或缺 key。
5. `verify-channel-presentation.mjs` 增加对 `routeTargetOverridesByKeyId`、新的 key 级 test id、key 级 route-target 空态/预览文案的 no-browser 合同断言。
6. 状态文档口径同步收紧：渠道页当前应视为 selector/no-browser 合同已收口，但宿主机 browser-grade pass 仍受 Edge/CDP bootstrap 阻塞。

## 9. 验证与结果

- 已执行命令：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `node scripts/verify-channel-presentation.mjs`
  - `node scripts/verify-channel-create-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 结果：以上命令均通过

## 10. 风险、阻塞与兼容性

- 本轮没有引入新的接口或数据兼容性风险，改动仅消费现有 `route-target override` 查询结果。
- `9.1.1` 中的 quota / breaker / priority / weight 等字段仍未完成，本轮没有伪造这些字段；后续需要等后端字段稳定后再继续补强。
- 宿主机上的 `channel-page` 真实 browser smoke 仍受 Edge/CDP bootstrap 阻塞；本轮只收口了 selector/no-browser 合同和 key 级字段可见性。

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（`verify-channel-presentation`、`verify-channel-create-flow`）
- 本次使用了哪些本地资源 / skills / 记忆上下文，以及它们分别提供了什么结论：
  - canonical plan：确认 `9.1.1` 目标是 key 级展示粒度，而不是只停留在渠道级摘要
  - 前端主线状态与当前状态文档：确认本轮应停留在 `Phase G` 渠道线，并优先做一个不依赖后端新字段的小闭环
  - 最近 worklog 与 automation memory：确认页面级 selector 合同已补过，本轮更适合继续推进 key 级 route-target 可见性，并且需要修正文档中过度表述
- 本次是否使用了子 agent；若使用，分别负责了什么范围、产出了什么结论：未使用
- 手工 smoke 状态 / 阻塞原因 / 缺少的环境：未做手工点击；真实 browser-grade 验证仍受宿主 Edge/CDP bootstrap 阻塞
- 是否补了 worklog：是
- 是否还有遗留项：
  - 后续仍需在可用宿主上补 `channel-page` browser-grade pass
  - 渠道创建/编辑弹窗更细的 hover/focus 行为仍待浏览器级验收
  - `9.1.1` 中 quota / breaker / priority / weight 等更深字段仍待后端契约稳定后继续推进
- 是否满足进入下一任务的前置条件：满足；下一轮可继续留在同一 `Phase G` 渠道线，优先做创建/编辑弹窗 hover/focus 细节，或继续做 `9.1.1` 的中文化/可见性小闭环
