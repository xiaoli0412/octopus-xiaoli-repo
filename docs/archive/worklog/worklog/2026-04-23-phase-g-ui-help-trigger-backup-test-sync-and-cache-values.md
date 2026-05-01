# Phase G UI Help Trigger / Backup Test Sync / Cache Values Optimization

## 1. 任务信息

- 任务名称：Phase G 截图主线残留收口 + 低风险后端性能减冗余
- 日期：`2026-04-23`
- 当前阶段：`Phase G`（截图优先 UI 主线），附带一刀低风险后端性能优化
- 对应 milestone：`P0-B` 前端主线残留闭环 + 性能并入主线

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`1.1`、`1.2`、`4.1.1`、`9`、`14`
- 对应 workflow 章节：`1.0`、`1.1`、`1.2`、`1.3`、`1.4.1`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-help-hint-accessibility-closure.md`
  - `docs/worklog/2026-04-22-phase-g-dynamic-routing-progressive-help-closure.md`
  - `docs/worklog/2026-04-17-ui-group-home-followup.md`
- 本次任务目标：
  - 收口前端当前真实残留：`DynamicRouting.test.tsx` 断言漂移、`HelpHint` 在多个 `AccordionTrigger` 内形成嵌套交互 warning、备份页测试仍停留在旧英文口径。
  - 顺手做一刀低风险后端性能优化，减少缓存列表读取时不必要的整表 map 拷贝。
- 本次已盘点本地资源：
  - 五份主线文档
  - 最新 relay 主线 worklog
  - 三条 automation memory
  - 当前 thread summary
  - 仓库内现有 verification 脚本与 `web`/`go` 测试命令
- 本次使用的本地 resources / skills / 记忆上下文：
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-p0a-relay-live-path-dynamic-circuit-threshold-closure.md`
  - `docs/worklog/2026-04-23-p0a-relay-live-race-fallback-handler-closure.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`
  - `C:\Users\李昊桐\.codex\automations\octopus\memory.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 若未使用部分本地资源或上下文，原因：本轮是小闭环残项和低风险优化，不需要再扩展到新的成功案例或外部资料。
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`N/A`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：主线程执行，收工后同步到 automation memory
- 若使用分目录 agent，负责目录与禁止越界范围：`N/A`
- 若未使用子 agent，原因：本轮阻塞点小而紧耦合，涉及同一批前端结构与测试同步，主线程直接收口更快且更稳。

## 3. 本次硬规则

- 不把“测试漂移”误判为“实现问题”，先对齐当前真实 UI / 文案 / 帮助提示口径。
- 不为了去 warning 破坏可访问性；`HelpHint` 继续保留可聚焦按钮语义，只从 DOM 结构上移出 `AccordionTrigger` 按钮树。
- 性能优化只做低风险减冗余，不触碰还在演进中的复杂 relay / route-target 语义。
- 中文主线继续保持简中一致，不回退到旧英文主显示。

## 4. 本次禁止事项

- 不回滚仓库中已有的无关脏改动。
- 不扩大到新的后端主题，不把本轮从 `P0-B` 又拉回大范围 backend 设计讨论。
- 不把 `HelpHint` 降级成纯视觉元素或丢失键盘可达性。

## 5. 本次验收条件

- `DynamicRouting` 组件测试恢复通过。
- `HelpHint` 可访问性校验继续通过，且 `AccordionTrigger` 相关嵌套交互 warning 根因被结构性消除。
- `web` 现有测试全绿，构建通过。
- 后端低风险性能优化完成并通过相关 `go test`。
- 本轮结论写入 worklog 与 automation memory。

## 6. 本次回滚点

- `web/src/components/ui/accordion.tsx` 的 `addon` 结构可单独回滚。
- `web/src/components/modules/channel/Form.tsx` 的 key 行交互布局可单独回滚。
- `internal/utils/cache/cache.go` / `internal/utils/cache/shard.go` 的 `Values()` 读取优化可单独回滚。

## 7. 实现范围

- 先改数据语义还是先改 UI：先收口 UI 结构与测试，再补后端缓存读取减冗余。
- 受影响后端模块：
  - `internal/utils/cache`
  - `internal/op/apikey.go`
  - `internal/op/channel.go`
  - `internal/op/group.go`
  - `internal/op/llm.go`
  - `internal/op/route_target.go`
  - `internal/op/stats.go`
- 受影响前端模块：
  - `web/src/components/ui/accordion.tsx`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/src/components/modules/channel/Form.tsx`
  - `web/src/components/modules/navbar/DocModal.tsx`
  - `web/src/components/modules/setting/DynamicRouting.test.tsx`
  - `web/src/components/modules/setting/backup-logic.test.ts`
  - `web/src/components/modules/setting/Backup.test.tsx`
  - `scripts/verify-help-hint-accessible.mjs`
- 受影响接口：无新增接口，仅缓存读取内部实现更轻。
- 是否影响旧数据：`否`
- 是否影响旧行为：
  - 前端可见行为更合理：折叠头里的帮助问号和多 Key 控件不再嵌套在触发按钮内。
  - 后端外部行为不变，仅减少列表路径冗余分配。

## 8. 实施步骤

1. 复盘五份主线文档、memory 和当前 thread summary，确认这轮属于 `P0-B` 残留收口，不再发散。
2. 识别 `HelpHint` warning 根因不是单文件，而是多个 `AccordionTrigger` 内嵌按钮；给 `AccordionTrigger` 增加外置 `addon` 槽位，并把相关问号迁出触发按钮树。
3. 收口渠道多 Key 折叠头中的 `Switch` / 删除按钮，避免继续在触发按钮内嵌交互控件。
4. 同步修复 `DynamicRouting.test.tsx` 与备份页测试，全部跟当前中文主线和隐藏测试文本对齐。
5. 在缓存层新增 `Values()`，把常用列表路径改成直接取值切片，减少 `GetAll()` 先拷整张 map 再组 slice 的冗余。
6. 跑前端、脚本、后端验证，确认本轮闭环成立。

## 9. 测试与验证

- 构建命令：
  - `pnpm --dir web build`
- 测试命令：
  - `node .\scripts\verify-help-hint-accessible.mjs`
  - `pnpm --dir web test:locale-consistency`
  - `pnpm --dir web test`
  - `$env:GOCACHE='D:\GPT-codex\octopus_repo\.tools\gocache'; $env:GOTMPDIR='D:\GPT-codex\octopus_repo\.tools\gotmp'; $env:TEMP='D:\GPT-codex\octopus_repo\.tools\tmp'; $env:TMP='D:\GPT-codex\octopus_repo\.tools\tmp'; go test ./internal/op ./internal/task ./internal/utils/cache -count=1`
- 专项验证：
  - `DynamicRouting.test.tsx` 恢复通过。
  - 备份页逻辑与组件测试全部恢复通过，中文主线与隐藏英文测试锚点一致。
  - `AccordionTrigger` 结构现支持外置 `addon`，帮助提示不再嵌套于触发按钮内。

## 10. 风险与兼容性

- 新风险：
  - `AccordionTrigger` 结构改造会影响少量折叠头布局，需要后续浏览器 smoke 再看桌面端和 `375px` 细节。
- 兼容性风险：
  - 由于仓库本身已有大量并行脏改，本轮只在目标切片内操作，没有清理其它历史改动。
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过`
- 测试是否通过：`通过`
- 本次使用了哪些本地资源 / skills / 记忆上下文：
  - 五份主线文档
  - 最新 relay worklog
  - 三条 automation memory
  - 当前线程 summary
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账确认了这轮仍是截图优先主线，且性能优化必须并入主线。
  - canonical / workflow 确认了本轮应先收口残留，再补 worklog 与 automation memory。
  - `octopus-2` memory 明确提醒当前前端主线不要再只做 copy，要补浏览器级证据；因此本轮先把 no-browser 残留彻底收绿。
  - relay worklog 证明 backend P0-A 已不再是默认阻塞，允许把主线拉回 `P0-B`。
- 本次使用了哪些子 agent 及其结论：`未使用`
- 子 agent 分工、负责范围与产出摘要：`N/A`
- 手工 smoke 状态：未做浏览器级手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：当前真正的浏览器级证明仍受本机 Edge / CDP page bootstrap 阻塞，需要沿既有 browser-smoke 主线继续推进。
- 待验证页面清单：
  - 设置页帮助提示真实浏览器 hover/focus
  - `CC Switch` 浏览器级证据
  - 渠道创建 / 分组创建弹窗浏览器级证据
  - `375px` 布局与帮助提示交互
- 若未使用子 agent，原因：本轮问题范围小、互相耦合、且需要主线程统一判断现状与测试口径。
- worklog 是否更新：`是`
- 遗留项：
  - 五份主线文档对应内容仍未全部收口，尤其浏览器级证据和动态路由文档与真实交付边界的同步还没完全做完。
  - `Phase G` 下一入口仍应是 browser-grade proof，而不是再次扩散到新的后端主题。
  - 若继续做后端，剩余边界仅建议聚焦 same-channel `(channel,key,model)` race 粒度设计。
- 下一任务前置条件是否满足：`满足`
