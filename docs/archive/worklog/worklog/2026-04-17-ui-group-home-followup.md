# UI Follow-up - Group Key Pool & Home Provider Breakdown

> 目的：继续按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.2.1 与第 9.5 节推进前端主线，把分组页左栏结构继续贴近 pooled/classified 目标，并补齐首页缺失的 provider token 拆分。

---

## 1. 任务信息

- 任务名称：UI Group/Home Follow-up
- 日期：2026-04-17
- 当前阶段：Phase B / UI 主线收口
- 对应 milestone：里程碑 2 可观测性与可配置性增强

## 2. 开工前输入

- 对应 canonical 章节：
  - 第 9.2 节分组页
  - 第 9.2.1 节分组左栏 provider -> key 分段展示
  - 第 9.5 节首页
  - 第 9.5.1 节首页最终必须展示的字段清单
- 对应 workflow 章节：
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 3、4 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-15-backend-task-routing-test-guardrails.md`
- 本次任务目标：
  - 在分组页左栏为 `pooled` 模式补上显式 `Key Pool` 展示
  - 继续增强 `classified` key section 的可理解性，稳定显示 key 元信息
  - 在首页 `Token Breakdown` 中补齐 `provider` 维度拆分
  - 在首页总览补齐成功/失败/成功率摘要
  - 在日志页补显式导出面板，暴露格式 / 时间范围 / 导出上限
  - 新增前端专属主线 MD，单独沉淀第 9 节 UI 完成度
  - 在首页补 estimated official/gateway price summary
  - 在首页补 estimated probe cost、熔断摘要与最近探测摘要
- 本次已盘点本地资源：
  - canonical MD
  - 当前线程上下文与既有 handoff 总结
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
  - `web/src/api/endpoints/channel.ts`
  - `web/src/api/endpoints/stats.ts`
  - `internal/op/stats.go`
- 本次使用的本地 resources / skills / 记忆上下文：
  - 直接继承已完成的 `classified/pooled` 前端分层实现
  - 直接复用已有 `useChannelList()` 渠道元数据与 `useStatsTokenBreakdown()` 统计接口
  - 继承当前线程里关于 Windows 下 `apply_patch` 受限、需回退到 `codex-run-as-apply-patch` 的经验
- 若未使用部分本地资源或上下文，原因：
  - 未新增后端接口，因为首页 `provider` 维度可先由现有前端数据低风险聚合完成
- 本次是否启用子 agent 与分工边界：
  - 是。主线程负责最终实现与验证；只读 explorer 负责分别核对分组页与首页/日志页主线缺口
- 本次子 agent 使用模型：
  - `gpt-5.4`
- 若未使用子 agent，原因：
  - 不适用

## 3. 本次硬规则

- 严格贴住 canonical MD，不偏离主线扩需求
- 不改 group 提交 payload 与后端 group 语义
- 分组页增强仅限前端结构表达层
- 首页 provider 拆分优先复用现有数据，不为此新开后端接口

## 4. 本次禁止事项

- 不顺带改 group 保存结构为 key/model 级绑定
- 不把 pooled/classified 可视化扩展成新的后端数据模型
- 不为了首页 provider 维度而引入未验证的统计接口改造

## 5. 本次验收条件

- 分组页左栏可显式展示 `pooled` 下的 `Key Pool`
- `classified` key section 可稳定展示 `remark/source_type/allowed_models` 相关语义
- 首页 `Token Breakdown` 可展示 `by_provider`
- `pnpm build` 通过

## 6. 本次回滚点

- 仅涉及前端页面组件与三套 locale 文案
- 如实现不理想，可按文件维度回退：
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
  - `web/public/locale/*.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先复用现有数据，再改 UI 表达
- 受影响后端模块：无业务语义变更；仅读取既有 `internal/op/stats.go` 输出
- 受影响前端模块：
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
  - `web/src/components/modules/home/total.tsx`
  - `web/src/components/modules/log/index.tsx`
  - `web/src/api/endpoints/stats.ts`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
- 受影响后端模块：
  - `internal/op/stats.go`
  - `internal/op/probe.go`
  - `internal/relay/probe.go`
  - `internal/relay/balancer/circuit.go`
  - `internal/server/handlers/stats.go`
- 受影响文档：
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-17-ui-group-home-followup.md`
- 受影响接口：
  - 复用 `/api/v1/channel/list`
  - 复用 `/api/v1/stats/token-breakdown`
- 是否影响旧数据：否
- 是否影响旧行为：否，均为展示层增强

## 8. 实施步骤

1. 重新对齐 canonical MD 第 9 节，并让两个 `gpt-5.4` 子 agent 只读核对分组页与首页/日志页剩余缺口。
2. 在 `group/Editor.tsx` 中增强左栏 section 构造与搜索命中范围：
   - `classified` section 补 `remark/source_type/allowed_models` 语义
   - `pooled` section 补显式 `Key Pool`
3. 在 `home/token-breakdown.tsx` 中基于 `by_channel + useChannelList()` 聚合 `by_provider`，并补三套 locale 文案。
4. 在 `home/total.tsx` 中把请求统计扩展为请求数、成功数、失败数与成功率摘要。
5. 在 `log/index.tsx` 中补日志导出面板，显式暴露 `json/jsonl`、时间范围、导出上限与前置校验。
6. 新增 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`，把 canonical 第 9 节前端主线单独沉淀。
7. 运行 `pnpm build` 做收口验证。
8. 在后端补 recent runtime probe 与 breaker summary 聚合，并回到首页展示 `probe cost / 熔断摘要 / 最近探测摘要`。

## 9. 测试与验证

- 构建命令：`pnpm build`（`web/`）
- 测试命令：`go test ./...`
- 专项验证：
  - `pnpm build`：通过
  - `go test ./...`：通过
  - 首页/分组页/日志页相关 TypeScript 编译：通过
  - recent probe / breaker summary 后端测试：通过

## 10. 风险与兼容性

- 新风险：
  - 首页 provider 维度当前由前端按 channel type 聚合，若未来 provider 概念与 `ChannelType` 脱钩，需要再下沉为后端显式字段
- 兼容性风险：
  - 低；本轮不改接口返回结构与保存结构
- 是否阻塞下一任务：
  - 否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：前端构建通过；未额外跑后端测试
- 本次使用了哪些本地资源 / skills / 记忆上下文：
  - canonical MD
  - 线程 handoff 总结
  - 既有 `group/home/stats/channel` 代码
  - Windows apply_patch 回退经验
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - canonical MD：明确本轮必须贴的主线是 `9.2.1` 与 `9.5`
  - handoff 总结：明确分组页已完成部分与剩余缺口，避免重复劳动
  - 现有代码：证明 `provider` 维度可由前端聚合，不必新开后端接口
  - 现有 stats total 数据：证明成功/失败/成功率可直接由前端总览复用，不必改后端
  - 现有 log export 接口：证明日志页仅靠前端即可补足格式/范围/上限导出交互
- 现有 `StatsTokenBreakdown + LLMInfo`：证明首页 estimated price summary 可通过最小只读聚合补齐，而无需改动持久化统计结构
  - 现有 relay race probe 与 breaker runtime：证明首页 recent probe / circuit summary 可先基于运行时观测补齐，而无需先重做持久化 stats schema
  - apply_patch 回退经验：保证在 Windows 环境下仍可稳定落 patch
- 本次使用了哪些子 agent 及其结论：
  - `Bohr`（`gpt-5.4`）：确认分组页最高价值缺口是 `pooled` 缺少 `Key Pool` 与 `classified` 缺少更稳定的 key 元信息
  - `Planck`（`gpt-5.4`）：确认首页最高价值缺口是缺少 `provider` token 拆分，且日志导出已基本满足 MD
  - `Lorentz`（`gpt-5.4`）：给出前端专属 MD 的建议结构，帮助把第 9 节前端主线独立沉淀
  - `Bacon`（`gpt-5.4`）：对日志导出面板做只读审查，确认“小屏稳健性 + 参数前置保护”是最值得收口的两点
- 子 agent 分工、负责范围与产出摘要：
  - `Bohr` 只读负责 `group/Editor.tsx + MD 9.2/9.2.1` 对照
  - `Planck` 只读负责 `home/log + stats + MD 9.4/9.5` 对照
  - `Lorentz` 只读负责“前端专属 MD”结构设计
  - `Bacon` 只读负责日志页导出面板的移动端/交互风险审查
  - 主线程负责代码实现、冲突消解、构建验证与 worklog 回写
- 手工 smoke 状态：未做完整页面手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：当前以构建通过为主，未启动完整前后端联调环境
- 待验证页面清单：
  - 分组页左栏 pooled/classified 展示
  - 首页 Token Breakdown provider/channel/model 三列布局
  - 日志页导出面板（格式 / 时间范围 / 上限 / 非法时间保护）
  - 首页 estimated official/gateway price summary
  - 首页 estimated probe cost / circuit summary / recent probe summary
- 若未使用子 agent，原因：不适用
- worklog 是否更新：已更新
- 遗留项：
- 当前首页 probe/circuit 仍是 recent runtime 观测摘要，尚不是持久化历史统计
- 当前首页价格相关能力仍是 `estimated` 视图，尚不是完整账单统计
  - 日志页导出后续仍可继续补更多快捷范围与摘要细化
  - 前端专属 MD 后续需要随第 9 节增量持续更新
- 下一任务前置条件是否满足：满足，可继续沿 `9.7` 设置页项目级备份 / 导入 / 迁移适配 UI 收口推进

## 12. 追加记录：设置页 9.7 收口

- 追加日期：2026-04-17
- 对应 canonical 章节：
  - 第 `9.7` 节项目级备份 / 导入 / 迁移适配 UI
  - 第 `11.5.1` 节备份功能要求
  - 第 `11.5.2` 节导入功能要求
  - 第 `11.5.3 / 11.5.4` 节迁移适配与差异预览目标
- 本次追加目标：
  - 把设置页备份区收口为“项目级快照导出 + dry-run 预检 + 增量导入”的真实主流程
  - 修正默认导出明文凭据与 manifest 语义不一致的问题
  - 补清导入结果、兼容性摘要、影响范围、风险边界与当前未支持能力说明
- 本次使用的本地资源：
  - canonical MD 第 `9.7 / 11.5 / Phase 6`
  - `internal/op/backup.go`
  - `internal/server/handlers/setting.go`
  - `web/src/components/modules/setting/Backup.tsx`
  - `web/src/api/endpoints/setting.ts`
  - 现有 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 本次子 agent 使用情况：
  - `Turing`（`gpt-5.4`）只读负责 canonical `9.7 / 11.5 / Phase 6` 对齐，结论是本轮边界应收在 `json + manifest + dry-run + 差异提示`，不能伪装已有映射编辑 / 回滚 / 部分恢复。
  - `Arendt`（`gpt-5.4`）只读负责 `Backup.tsx + setting.ts` 盘点，结论是应优先增强操作语义、manifest 完整度、兼容性总评、已选文件信息与未支持能力说明。
  - 主线程负责后端安全默认值、前端设置页实现、测试验证与文档回写。
- 本次实际代码落点：
  - 后端导出默认改为安全快照：导出时默认去除 `channel_keys.channel_key`、`api_keys.api_key`，并脱敏可能承载凭据的自定义 header / 代理认证信息。
  - 后端导入在 dry-run / apply 阶段都会显式识别空凭据；apply 时会跳过空 `channel key / API key`，并返回 warning，避免把默认脱敏快照导成半残凭据记录。
  - 设置页 `Backup` 重构为更贴近 `9.7` 的真实流程：
    - 项目级快照说明
    - 导出范围摘要与安全默认值提示
    - dry-run / apply 模式摘要条
    - 已选文件信息
    - manifest 完整展示（补 `contains_secrets`）
    - 兼容性总评卡、影响范围、冲突 / 缺失 / skipped / route warning 分区
    - 当前未接后端能力清单，避免误导为已支持回滚 / 映射编辑 / 部分恢复
- 本次验证：
  - 新增后端测试覆盖默认脱敏导出、dry-run 红线提示、apply 跳过空凭据
  - 新增后端测试覆盖 `skip` 模式保留现有行、结构化兼容性报告中的 `base_url / schema / alias / route` 差异
  - `go test ./internal/op`：通过
  - `pnpm build`：通过
- 本次遗留项：
  - `replace / merge / map` 真正的冲突处理模式仍未有后端协议
  - rollback snapshot、部分恢复、导入映射表、导入后自动验证仍未落地
  - 设置页文案当前以现有 locale key 复用为主，后续若继续深挖 `9.7`，建议单独补更精确的 i18n 文案

## 13. 追加记录：11.5.2 继续收口（import mode + structured diff）

- 追加日期：2026-04-17
- 对应 canonical 章节：
  - 第 `11.5.2` 节导入功能要求
  - 第 `11.5.3` 节迁移适配目标
  - 第 `11.5.4` 节导入后模拟路由验证 / 迁移差异预览
- 本次追加目标：
  - 把导入从“固定 incremental”推进到“显式 `incremental / skip` 模式”
  - 把兼容性报告从粗粒度 warning 扩展到结构化 `base_url / schema / alias / route` 差异
  - 让设置页真实暴露 mode 选择和结构化差异展示，但不伪装 `replace / merge / map / rollback`
- 本次使用的本地资源：
  - canonical MD 第 `11.5.2 / 11.5.3 / 11.5.4`
  - 现有 `internal/op/backup.go` 与 `internal/op/backup_test.go`
  - 现有 `web/src/components/modules/setting/Backup.tsx`
  - 前端主线状态文档与当前线程 handoff 总结
- 本次子 agent 使用情况：
  - `Linnaeus`（`gpt-5.4`）只读负责核对 canonical `11.5.2 / 11.5.3 / 11.5.4` 与当前实现差距，结论是本轮最可信的下一步应收在 `incremental / skip` 与结构化差异，不能伪装 replace / rollback。
  - `Newton`（`gpt-5.4`）只读负责设置页 UI 对齐检查，结论是最值得落地的是 mode selector、结构化差异展示，以及保留未实现能力的 disabled/说明状态。
  - 主线程负责后端 import mode、structured diff、前端接线、测试验证与文档回写。
- 本次实际代码落点：
  - 后端导入新增显式 `mode`：支持 `incremental / skip`，默认仍为 `incremental`。
  - 后端 dry-run compatibility 新增结构化字段：
    - `alias_conflicts`
    - `route_conflicts`
    - `base_url_mismatches`
    - `schema_mismatches`
  - 后端 apply 针对 `skip` 模式新增真实语义：
    - `settings / llm_infos / stats` 使用 `DoNothing` 而不是 upsert
    - 已存在 `channel / group` 时，不再继续导入其 `channel_keys / group_items`
    - `channel / group` 导入按业务键名称回查 ID，避免继续依赖快照内旧 ID 直接落库
  - 设置页 `Backup` 新增：
    - `incremental / skip` selector
    - 随 mode 变化的导入语义摘要
    - `base_url / schema / alias / route` 结构化差异分区
    - summary 卡片中新增 `Base URL Diff / Schema Diff`
    - 未实现能力提示改为只保留 `replace / merge / map` 等仍未接线项
- 本次验证：
  - `gofmt -w internal/model/backup.go internal/op/backup.go internal/op/backup_test.go internal/server/handlers/setting.go`
  - `go test ./internal/op`：通过
  - `pnpm build`（`web/`）：通过
  - 非阻塞告警：`baseline-browser-mapping` 数据过旧
- 本次遗留项：
  - `replace / merge / map` 仍未实现
  - rollback snapshot、mapping editor、partial restore、post-import health check 仍未实现
  - 当前 route diff 仍属于“结构化差异提示”，还不是完整的导入前后候选链模拟器
