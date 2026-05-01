# 2026-04-29 Phase H AI learning same-expectation freshest note closure

## 1. 任务信息

- 任务名称：Phase H6 AI learning same-expectation freshest note closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-strategy-specific-stable-replay-and-json-new-blocker-closure.md`
- 本次任务目标：收紧 `fallback_replay_ready` 场景下的 freshest-copy / recommended-action 语义，避免把“更鲜但 opposite-expectation 或不同 strategy 的副本”误导成更适合当前 invocation 的主证据
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、上一轮 H6 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostics
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON
- 若未使用部分本地资源或上下文，原因：本轮不涉及产品 UI、后端 schema、运行态 external 复跑或新的业务接口，只做 verifier/diagnostic 语义收口
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 verifier / stable replay / host-level blocker handoff 主线，不扩散到产品 UI、后端 schema、宿主修复或新增 external 证据采集
- 不改 stable copy 选择顺序，只修正 freshest-note 与 action 文案对“当前 invocation 最优证据”的表达
- 任何新提示都必须保持 strategy-aware 与 require-cdp-aware，不能重新退化成只看“最近写入时间”

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不新增伪造 stable artifact，不触发新的 external/self-start 运行态验证
- 不扩大到新的 CDP wrapper 架构改造或宿主网络修复

## 5. 本次验收条件

- `check-only -RequireExternalCdpPreflight` 在 selected same-expectation replay 比 fresher opposite-expectation copy 更适合当前 invocation 时，会显式说明 fresher copy 只是 comparison-only
- `stable freshest copy note` 不再只说“selected preview is older than the freshest parseable copy”，而是补足“为何仍保留 selected same-expectation replay”
- `decision summary` 与 `recommended action` 在同场景下同步输出同一语义，不再让 handoff 误判为应改用 fresher opposite-expectation 副本
- `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-same-expectation-freshest-note-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI；先改 verifier 文案语义，再同步 no-browser 守护与阶段文档
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser 守护 `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅收紧 learning smoke `check-only` 的解释层与 handoff 行为

## 8. 实施步骤

1. 复核 Phase H6 最新状态，确认本轮不再追 host-level live rerun，而是优先修正 strategy-aware replay 的解释层歧义。
2. 阅读 `verify-ai-automation-learning-browser-smoke.ps1` 中 freshest-note、decision summary 与 recommended action 逻辑，定位“只按 fresher/older 描述，未解释 fit”的分支。
3. 新增 selected-preference helper，把“比 selected 更新的副本”按 same-expectation / opposite-expectation / alternate-strategy 分类，并给 freshest-note、freshness note、decision summary、recommended action 追加 comparison-only 提示。
4. 更新 `verify-ai-automation-learning-focus.mjs` 守住新文案契约，并同步刷新状态文档与 worklog。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`powershell -NoProfile -Command "& { $tokens=$null; $errors=$null; [void][System.Management.Automation.Language.Parser]::ParseFile('D:\GPT-codex\octopus_repo\scripts\verify-ai-automation-learning-browser-smoke.ps1',[ref]$tokens,[ref]$errors); if ($errors -and $errors.Count -gt 0) { $errors | ForEach-Object { $_.ToString() }; exit 1 } }"`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-same-expectation-freshest-note-closure.md`

## 10. 风险与兼容性

- 新风险：低；只新增策略适配解释，不改变 stable copy 的真实选择结果
- 兼容性风险：低；旧输出仍保留 freshness / coverage / action class 主结构，只是在特定 fallback 场景追加更明确的 preference 说明
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（两次 `check-only`、PowerShell parser、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostics
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 H6 已完成 strategy-specific stable replay 与 `json-new` blocker 结论，因此本轮最值得推进的是减少 handoff 误导，而不是继续采集新外部证据；repo-local stable diagnostics 直接证明 `requireCdp=true` 场景下 selected same-expectation replay 可能比 opposite-expectation copy 更旧，因此需要补“fit 优先于 freshness”的解释层
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未新增运行态 smoke；沿用 repo-local stable diagnostics 做 `check-only` 回放验证
- 手工 smoke 阻塞原因 / 缺少的环境：真实 `vitest/esbuild spawn EPERM` 仍阻塞 jsdom 入口；host-level live external/browser 证据仍受当前宿主 page bootstrap blocker 限制，但不阻塞本轮 verifier 语义收口
- 待验证页面清单：无新增产品页；下一轮若继续 H6，更适合换宿主或走真正不同执行路径
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - `attached-session` strategy-specific stable copy 仍缺失，当前只是在 handoff 上明确“same-expectation fallback 优先于 fresher opposite-expectation comparison copy”
  - `json-new` 与 `attached-session` 都已在本机形成 host-level blocker 结论，继续推进 H6 时应优先换宿主或换真正不同执行路径
- 下一任务前置条件是否满足：满足；verifier 解释层已闭环，下一轮可直接承接 host handoff 或切到同主线相邻任务
