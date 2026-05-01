# 2026-04-23 Phase G 设置页帮助提示浏览器 smoke CLI 自启动收口

## 1. 任务信息

- 任务名称：Phase G 设置页帮助提示浏览器 smoke `cli self-start` 收口
- 日期：`2026-04-23`
- 当前阶段：`Phase G` 截图优先 UI 主线 / 浏览器级证据闭环
- 对应 milestone：设置页帮助提示真实浏览器证据恢复

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-23-phase-g-ui-help-trigger-backup-test-sync-and-cache-values.md`
  - `docs/worklog/2026-04-23-phase-g-help-hint-locale-and-selector-closure.md`
- 本次任务目标：
  - 继续推进五份主线文档对应的 `Phase G` 未完成项，不停在 planning。
  - 修复设置页帮助提示浏览器 smoke 的 `cli self-start` 真实阻塞点，恢复浏览器级通过证据。
  - 把这轮排查得到的真实根因写回 worklog / automation memory，避免下轮重复走错假设。
- 本次已盘点本地资源：
  - 五份主线文档
  - 最新 `Phase G` / help-hint / backup worklog
  - 三份 automation memory
  - 当前线程 summary 与失败工件
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `web/src/components/common/HelpHint.tsx`
  - `scripts/verify-help-hint-accessible.mjs`
  - `web/src/components/modules/setting/Backup.test.tsx`
- 本次是否启用子 agent 与分工边界：`否`
- 若未使用子 agent，原因：问题范围紧耦合，且需要主线程统一判断脚本、静态产物、浏览器会话与验证口径。

## 3. 本次硬规则

- 不把 smoke 脚本失败直接等同于页面实现失败，先看真实工件和运行现场。
- 不在未确认前继续沿“wrapper frontend URL 选错”这个假设扩大改动。
- 必须给出浏览器级通过证据或明确的主机级阻塞，而不是只做 no-browser 推断。

## 4. 本次禁止事项

- 不回滚工作区中的无关脏改动。
- 不为了过 smoke 改坏 `HelpHint` 的按钮语义与可访问性。
- 不把手工调试结论写成未经验证的最终事实。

## 5. 本次验收条件

- 找到 `cli self-start` 真实阻塞点。
- 设置页帮助提示 `cli self-start` 浏览器 smoke 成功通过。
- `HelpHint` 真实浏览器选择器与 smoke / 可访问性校验口径一致。
- 前端静态导出与 `web` 测试恢复通过。
- 本轮结论写回 worklog 与 automation memory。

## 6. 实现范围

- 受影响前端 / 脚本模块：
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `scripts/verify-help-hint-accessible.mjs`
  - `web/src/components/common/HelpHint.tsx`
  - `web/src/components/modules/setting/Backup.test.tsx`
- 受影响构建产物：
  - `web/out`
  - `static/out`

## 7. 实施步骤

1. 复核主线文档、worklog、automation memory 与失败工件，确认这轮仍应留在 `Phase G` 浏览器级证据主线。
2. 复查 `verify-setting-help-browser-smoke.ps1` 与后端静态托管逻辑，确认 `frontendBaseUrl = backendBaseUrl` 在 wrapper 自启动路径中并非根因。
3. 直接读取失败工件并手工拆解 `@playwright/cli` 会话，找到真实问题链：
   - `playwright-cli open` 后首条命令存在短暂 session 注册延迟；
   - `verify-setting-help-browser-smoke.mjs` 把 `eval --raw` 结果再次 `JSON.stringify`，导致 snapshot 长期解析漂移；
   - `HelpHint` 真实浏览器 DOM 落到 `tooltip-trigger`，旧 smoke 选择器数不到帮助按钮；
   - 自启动 wrapper 的 Node 子进程成功退出后，PowerShell 仍把 `ExitCode` 读成 `unknown` 并误判失败；
   - 静态构建需重新同步到 `static/out` 才能让新的选择器进入后端托管页面。
4. 用最小改动修复上述链路：
   - `verify-setting-help-browser-smoke.mjs` 改为直接解析 `eval --raw` JSON，并给 `session not open` 增加短重试；
   - `HelpHint` 新增稳定 `data-help-hint-trigger="true"`；
   - smoke / accessible 校验脚本统一切到稳定 selector；
   - `verify-setting-help-browser-smoke.ps1` 改用 `Wait-Process` + 显式 `ExitCode` 读取，避免成功误报失败；
   - 重建 `web/out` 并同步到 `static/out`；
   - 同步 `Backup.test.tsx` 的帮助按钮与 placeholder 测试口径。
5. 运行浏览器 smoke、前端测试与静态校验，确认这轮真正闭环。

## 8. 测试与验证

- `node .\scripts\verify-help-hint-accessible.mjs`
- `pnpm --dir web test:locale-consistency`
- `pnpm --dir web build:static`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cli -NodeSmokeTimeoutSeconds 30`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 120 -KeepArtifacts`
- `pnpm --dir web exec vitest run src/components/modules/setting/Backup.test.tsx`
- `pnpm --dir web test`

## 9. 本次关键结论

- 先前“wrapper 把 frontend URL 指到 backend URL”并不是这条 `cli self-start` 的主要阻塞点，因为该路径本来就是后端直接托管 `static/out`。
- 真正的阻塞链是脚本/工具口径问题，而不是设置页实现缺失：
  - `eval --raw` 结果被 Node smoke 双重编码；
  - `HelpHint` 需要一个稳定而不受 tooltip 包装层影响的浏览器级 selector；
  - PowerShell wrapper 把成功的 Node 进程误判为 `unknown` 失败；
  - 浏览器 smoke 前必须保证最新静态构建已同步到 `static/out`。
- 本轮已经拿到真实浏览器级通过证据：
  - `setting-help-browser-smoke passed`
  - `desktopHelpButtons: 20`
  - `mobileWidth: 375`
  - `Frontend URL: http://127.0.0.1:18082`

## 10. 风险与兼容性

- 新风险：低；本轮主要是 smoke/selector/wrapper 稳定性修复，不改业务流程。
- 兼容性风险：低；`HelpHint` 仍保持真实按钮语义，只是增加稳定属性，便于浏览器和测试脚本对齐。
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过`
- 测试是否通过：`通过`
- 手工 smoke 状态：通过 `cli self-start` 真实浏览器 smoke 拿到设置页帮助提示和 `375px` 证据
- 浏览器级证据摘要：
  - 设置页四张卡片已渲染
  - 帮助按钮计数 `20`
  - 移动端视口宽度 `375`
  - wrapper 自启动链路成功
- 遗留项：
  - `Driver cdp` 这条宿主级 Edge/CDP bootstrap 主线还未在本轮完全恢复；
  - 可继续把同一 smoke 入口扩展到更细的 hover/focus 截图证据，或沿同池推进 `channel/group create dialog` 浏览器级证据。

## 12. 下一轮建议

1. 沿同一 `Phase G` 浏览器主线继续，优先补 `Driver cdp` 或更细的 hover/focus 截图证据，而不是回到抽象 planning。
2. 把已跑通的 CLI 自启动 smoke 结果同步回前端主线状态文档，明确“设置页帮助提示 + 375px”已具备真实浏览器级通过证据。
3. 在同一截图优先池里继续推进 `channel create` / `group create` dialog 的浏览器级证据。
