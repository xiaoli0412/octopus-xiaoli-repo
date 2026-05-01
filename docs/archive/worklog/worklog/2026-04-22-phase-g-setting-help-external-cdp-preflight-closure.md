# 2026-04-22 Phase G 设置页帮助提示 external CDP 预检前移收口

## 1. 任务信息

- 任务名称：设置页帮助提示 external CDP 缺端点预检前移收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器 smoke 阻塞继续收敛

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-reuse-contract.md`
- 本次任务目标：把 `external + cdp` 的缺端点失败路径前移成独立可验证的小闭环，不再依赖先拉起后端和前端
- 本次已盘点本地资源：canonical plan、用户上下文总账、详细工作流、前端主线状态、最近两条 setting-help smoke worklog、automation memory、当前 wrapper 与 CDP smoke 脚本
- 本次使用的本地 resources / skills / 记忆上下文：同上
- 若未使用部分本地资源或上下文，原因：本轮是单脚本行为收口，不需要再扩散到其他主线文档
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 自动化链路主线程串行执行
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程解决，不创建子 agent

## 3. 本次硬规则

- 只修改设置页帮助提示浏览器 smoke wrapper，不扩散到设置页业务逻辑
- external 模式继续保持“只复用外部会话”的契约
- 必须留下当前环境可直接复跑的最小验证命令

## 4. 本次禁止事项

- 不回退工作区其他已有改动
- 不把缺端点受控失败写成真实浏览器 smoke 已通过
- 不扩散到备份导入、渠道、多 key、分组或 `CC Switch` 主线

## 5. 本次验收条件

- `external + cdp` 在 CDP 缺失时会先于前后端检查直接失败
- 失败提示里的 `remote-debugging-port` 能跟随自定义 `CdpUrl` 端口
- `check-only` 与新的缺端点失败路径都能在当前环境稳定复现

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-preflight-closure.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本行为
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅影响 smoke wrapper 的失败路径与提示文案

## 8. 实施步骤

1. 复核 wrapper 当前 `external + cdp` 分支，确认缺少 CDP 时仍会晚于前后端验证失败，不利于独立复现。
2. 在 `scripts/verify-setting-help-browser-smoke.ps1` 中新增 `Test-CdpEndpointReady`、`Resolve-PortHintFromUrl`、`Assert-ExternalCdpEndpointReady`，并把 external CDP 预检前移到主流程最前面。
3. 补充 `check-only` 提示，复跑 `check-only` 与缺端点失败用例，确认当前环境可稳定拿到受控结果。

## 9. 测试与验证

- 构建命令：未涉及
- 测试命令：
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -CdpUrl 'http://127.0.0.1:9999' -NodeSmokeTimeoutSeconds 15 ; exit $LASTEXITCODE"`
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'external' -Driver 'cdp' -FrontendUrl 'http://127.0.0.1:3101' -BackendUrl 'http://127.0.0.1:18081' -CdpUrl 'http://127.0.0.1:9999' -NodeSmokeTimeoutSeconds 15 ; exit $LASTEXITCODE"`
- 专项验证：
  - `check-only` 通过，并额外显示 external / self-start 的 CDP 契约提示
  - `external + cdp` 在当前环境直接走 CDP 缺失受控失败，不再被前后端依赖抢先阻塞
  - 失败提示中的 `--remote-debugging-port` 已对齐到 `9999`

## 10. 风险与兼容性

- 新风险：低；仅调整验证脚本分支顺序和提示文案
- 兼容性风险：低；未改接口、数据库、业务逻辑
- 是否阻塞下一任务：不阻塞；下一轮可直接用真实外部 Edge CDP 会话继续对照

## 11. 收工记录

- 构建是否通过：未涉及
- 测试是否通过：通过；`check-only` 成功，缺端点失败路径按预期触发
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、用户上下文总账、详细工作流、前端主线状态、最近 setting-help smoke worklog、automation memory、当前 wrapper 与 CDP smoke 脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：主计划确认继续停留在 Phase G 同一主线；前端主线状态与 memory 指向 external CDP 复用契约；最近 worklog 给出不可扩散边界与真实浏览器证据仍未闭环的背景
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未做真实浏览器 smoke；本轮只收口脚本级受控失败路径
- 手工 smoke 阻塞原因 / 缺少的环境：仍缺一个已开启 remote debugging 的外部 Edge 会话用于真实 `external + cdp` 对照
- 待验证页面清单：设置页四卡片帮助提示桌面与 `375px` 真实浏览器视图
- 若未使用子 agent，原因：用户明确要求主线程串行推进
- worklog 是否更新：是
- 遗留项：
  - 仍需拿真实外部 Edge CDP 会话跑完整 wrapper，确认是否能越过 page/runtime bootstrap
  - 若外部会话也卡在同一阶段，需要把结论升级为宿主级 CDP 行为而非 self-start 特有问题
- 下一任务前置条件是否满足：满足

