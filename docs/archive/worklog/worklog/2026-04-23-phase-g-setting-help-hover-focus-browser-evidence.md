# Phase G settings help-hint hover/focus browser evidence

- 时间：2026-04-23
- 主线：Phase G screenshot-first UI closure
- 范围：设置页四卡片 `HelpHint` 浏览器级证据收紧，不扩展新业务功能。

## 背景

上一轮已打通 `verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli` 的基础浏览器 smoke，但本轮继续检查时发现新增 hover/focus 逻辑仍存在两个证据风险：一是 Node stderr 出现异常时 wrapper 可能仍因 stdout/退出码路径误判为 passed；二是初版 hover 通过页面内合成 `MouseEvent` 和全局 tooltip 文本扫描判断，无法证明当前 trigger 自己打开了对应提示。

## 本轮变更

- `HelpHint` 保持真实 `button` 语义，并新增共享 `data-help-hint-id`：触发按钮和 tooltip 内容使用同一个稳定 id，供浏览器 smoke 绑定校验。
- `verify-setting-help-browser-smoke.mjs` 的交互检查改为真实浏览器动作：键盘 `Tab` 聚焦目标 help 按钮，`@playwright/cli hover` 真实悬停目标按钮，只接受当前 `data-help-hint-id` 绑定的 tooltip 文本。
- `verify-setting-help-browser-smoke.mjs` 增加分步 smoke 日志和单次 Playwright CLI 调用超时，避免再出现空日志卡死后难以定位。
- `verify-setting-help-browser-smoke.ps1` 增加 stdout 成功标记检查和 error-like stderr 拦截，避免 Node smoke 报错但 wrapper 误报成功。
- `verify-help-hint-accessible.mjs` 同步覆盖 `data-help-hint-id` 契约。

## 验证

- `node --check .\scripts\verify-setting-help-browser-smoke.mjs`
- `node .\scripts\verify-help-hint-accessible.mjs`
- `pnpm --dir web test:locale-consistency`
- `pnpm --dir web build:static`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`

最终浏览器证据：

- `result: setting-help-browser-smoke passed`
- `desktopHelpButtons: 21`
- `interactionChecks: 4`
- `mobileWidth: 375`
- frontend/backend: `http://127.0.0.1:18081`
- artifact: `C:\Users\李昊桐\AppData\Local\Temp\octopus-setting-help-smoke-6fb09c6359cb42d8a8a1fc8cf12dea0b`

## 结论

Settings 四卡片 help-hint 的 CLI self-start 浏览器证据已经从“基础可见”提升为“375px + keyboard focus + real hover + 绑定 tooltip 文本”的可信闭环。CDP 路径仍按宿主级 Edge/CDP bootstrap blocker 单独记录，不再阻塞该 CLI 主证据。

## 下一步

- 继续同一 Phase G screenshot-first 池，优先补 channel create / group create dialogs 的真实浏览器级 `375px`、展开态和帮助提示证据。
- `CC Switch` 浏览器证据仍保留为同池后续项。
- 不再回到旧的合成 hover 或全局 tooltip 文本扫描实现。
