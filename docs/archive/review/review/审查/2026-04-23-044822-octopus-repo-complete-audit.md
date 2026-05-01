# Octopus Repo Complete Audit

触发时间：`2026-04-23T04:48:23+08:00`

## 1. Findings

本轮未复现会直接阻断当前后端主流程的 `Critical` 级问题。`go build ./...`、`go test ./... -count=1` 和 `scripts/smoke-win-backend.ps1` 都通过，说明启动、鉴权、渠道创建、分组创建、API Key 创建和 `/v1/chat/completions` 主链当前是真实可达的。当前最高优先级问题集中在验证资产漂移、文档承诺超前和发布收口不足。

### High

1. `Frontend backup verification` 已接入 CI，但当前脚本与默认 locale/页面实现失配，按当前工作区会直接失败，形成发布阻塞。
   - 证据：`.github/workflows/validation.yaml:50-52` 明确运行 `pnpm run test:backup-logic && pnpm run test:backup-component`。
   - 本地复现：`node .\scripts\verify-backup-logic.mjs` 失败，断言期望 `Rollback has blocking risks`，实际返回 `回滚存在阻断风险`；`node .\scripts\verify-backup-component.cjs` 失败，期望可见文本 `Include plaintext credentials in the snapshot`，实际页面默认显示中文文案。
   - 代码依据：`web/src/stores/setting.ts:14` 默认 locale 是 `zh-Hans`；`web/src/components/modules/setting/backup-logic.ts:391,786,807-810,830-833` 默认按 `zh-Hans` 生成标题和导出开关文案；`web/src/components/modules/setting/Backup.tsx:407,710` 读取 locale 并把 locale 化后的 `toggleLabel` 传给 `SwitchRow`；而 `scripts/verify-backup-logic.mjs:83,200`、`scripts/verify-backup-component.cjs:184` 仍硬编码英文断言。
   - 判断：这不是“备份逻辑没接上”，而是“验证脚本还停留在旧英文口径”。由于它已被 workflow 真正调用，所以它是实际的高优先级阻塞项。

### Medium

2. 动态路由文档仍显著超前于实现，当前 shipped 行为只是“健康开关 + race budget + 每日摘要扫描”，并没有文档描述的 AI/多模式运行面。
   - 证据：`docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:233-237` 仍声明 `shadow-ai`、`hybrid`、`metrics-only`、`strict-mechanism`、`incident-safe`；`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:227-240` 仍以 `hybrid mode`、`metrics-only`、`strict-mechanism` 为实施/发布门槛。
   - 实现依据：`internal/model/setting.go:24-29,51-56` 当前只有 `dynamic_routing_health_enabled` 和五个 `race_*_budget` 设置；`internal/task/dynamic.go:47-62` 的后台任务只是 `daily_summary_scan_no_runtime_mutation` 摘要扫描；`web/src/components/modules/setting/DynamicRouting.tsx:89-198,269` 仅消费上述设置与摘要结果。
   - README 口径反而更接近现实：`README.md:178-180`、`README_zh.md:167-169` 已明确写成 `dynamic summary scan`，不会持久化新阈值也不会重排用户路由。
   - 判断：这属于“文档/规划承诺领先于代码”，不是运行时空指针，但会让完成度评估和对外交付口径失真。

3. 发布 workflow 的工具链版本是浮动的，和已验证的固定基线不一致，发布可复现性偏弱。
   - 证据：`.github/workflows/release.yaml:26,36,41` 使用 `go-version: 'stable'`、`pnpm version: 'latest'`、`node-version: 'latest'`。
   - 对比：`.github/workflows/validation.yaml:19-33` 使用固定 `go 1.24.4`、`pnpm 10`、`node 22`；`README.md:82-84`、`README_zh.md:81-83` 也写死了 Go/Node/pnpm 基线。
   - 判断：当前代码本身可编译，但 release job 仍可能因为上游工具链漂移而在未改代码时失效，属于真实发布风险。

4. 协议转换主路径可用，但边界能力仍有真实 TODO，且本轮没有找到直达这些 TODO 分支的直接验证。
   - 证据：`internal/transformer/inbound/anthropic/messages.go:162` 仍写着 `TODO: support other result types`；`internal/transformer/outbound/gemini/messages.go:319-321` 对 `json_schema` 仍是 TODO；`internal/transformer/model/model.go:619,861` 仍保留 `JSONSchema`/image-generation tool body 的未完成注释。
   - 已验证现实：当前 `go test ./... -count=1` 通过，说明核心路径没有因为这些 TODO 崩掉；但这并不等于这些边角协议场景已闭环。
   - 判断：这是“部分完成”而不是“空实现”，需要文档降级或补上定向测试/实现。

5. CI 的 Go 测试范围仍比仓库实际已有测试面窄，导致一批已存在测试的包不会在 workflow 中跑到。
   - 证据：`.github/workflows/validation.yaml:58` 只跑 `./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task`。
   - 仓库中实际存在但被漏掉的测试包包括：`internal/client/http_test.go`、`internal/conf/config_test.go`、`internal/helper/fetch_test.go`、`internal/model/channel_test.go`、`internal/model/setting_test.go`、`internal/server/auth/auth_test.go`、`internal/transformer/model/embedding_test.go`、`internal/transformer/outbound/antigravity/messages_test.go`、`internal/transformer/outbound/copilot/chat_test.go`。
   - 本地现实：`go test ./... -count=1` 通过，说明这些包今天是绿的；问题在于 workflow 没有持续兜底它们。
   - 判断：这是验证覆盖漂移，不是当前功能错误，但会削弱后续回归发现能力。

### Low

6. 顶层 README / README_zh 的安装与发布信息仍有占位符，无法作为可直接执行的发布文档。
   - 证据：`README.md:45,51-54,71-73,88` 和 `README_zh.md:45,51-54,70-72,87` 仍包含 `<CURRENT_IMAGE>`、`<CURRENT_REPOSITORY_URL>`、`<CURRENT_ONE_CLICK_INSTALL_SCRIPT>`、`<CURRENT_RELEASES_URL>` 以及 TODO 占位文本。
   - 判断：代码成熟度已经明显高于这些文档，当前问题更多是发布收口不足。

7. `web/package.json` 仍保留一个不可移植的本地 HTTPS 开发脚本。
   - 证据：`web/package.json:8` 的 `devs` 脚本硬编码 `/workplace/code/octopus/build/localhost-key.pem` 和 `/workplace/code/octopus/build/localhost.pem`。
   - 判断：这不影响主流程，但会误导其他环境的开发者，也不适合作为仓库默认脚本保留。

8. `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 里仍保留过期结论，和当前代码/环境不一致。
   - 证据：`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:253,394` 仍写 `manifest.contains_secrets` 与真实导出内容不一致；但 `internal/op/backup.go:149` 已按 `dumpContainsSecrets(d)` 写入真实值，`internal/server/handlers/setting_import_test.go:151` 也验证了包含凭证时 `contains_secrets=true`。同文档第 `5.4/7` 段还写“本机 go 不在 PATH、pnpm 未正确装好链接”，但本轮已实际跑通 `go build ./...`、`go test ./...` 和 Node 直调 `tsc`。
   - 判断：这是状态文档滞后，不是代码问题，但会直接误导接手人和后续自动化判断。

## 2. Completion Assessment

### 总体判断

当前仓库已经不是“半接线的骨架”，后端主流程和大量设置/备份/路由增强都是真实现、真接线；但项目整体仍属于“核心能力大多可用，发布与验证收口还不完整”的状态。

### 完成度评估

| 主题 | 评估 | 判断依据 |
| --- | --- | --- |
| 后端启动、鉴权、HTTP 服务、静态资源回退 | 已完成 | `cmd/start.go`、`internal/server/server.go`、`cmd/healthcheck.go` 主链可用，Windows smoke 已真实走通。 |
| Relay 主路径、failover、race fallback、route-target override | 已完成 | `internal/relay/relay.go` 已消费 `shouldEscalateToRace(...)`、`runRaceFallbackWindow(...)`；`go test ./internal/relay` 与全量 `go test ./...` 通过。 |
| 备份导出/导入/回滚后端 | 已完成 | `internal/server/handlers/setting.go` 和 `internal/op/backup.go` 已接入 export/import/rollback/preview，全量 Go 测试通过。 |
| 备份设置页与相关 UI 主流程 | 部分完成 | 页面已接入并含大量隐藏测试钩子，但当前 Node 验证脚本与 locale 口径漂移。 |
| 动态路由 shipped 能力 | 部分完成 | 已有健康开关、预算、摘要扫描；README 已按真实能力降级。 |
| 动态路由 AI / 多模式体系 | 未完成 / 文档先行 | 文档列出的 `shadow-ai`、`hybrid`、`incident-safe` 等并无对应设置与运行时实现。 |
| 协议转换边角场景 | 部分完成 | 核心路径可用，但 Gemini JSON schema、Anthropic 复杂 tool_result 等仍有 TODO。 |
| 发布文档与发布流水线收口 | 部分完成 | README 占位符、release workflow 浮动版本、硬编码开发脚本仍未收尾。 |

### 已完成、部分完成、未完成、疑似空实现

- 已完成：后端启动、JWT/API Key 鉴权、group/channel 创建链路、relay 真实转发、backup/import/rollback 后端、Windows smoke 路径。
- 部分完成：备份前端验证闭环、动态路由设置页、协议转换边界能力、发布自动化收口。
- 未完成：动态路由 AI/多运行模式、顶层发布文档真实值替换。
- 疑似空实现或“仅文档存在”：动态路由文档中的 `shadow-ai` / `hybrid` / `incident-safe` / `strict-mechanism` 目前更像规划术语，不是当前可配置或可运行的 shipped 功能。

## 3. Verification Summary

### 已验证项

- `git status --short --branch`、`git log --oneline -n 12`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`。
- `go build ./...`：通过。
- `go test ./... -count=1`：通过。
- `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`：通过。
- `node .\scripts\verify-locale-consistency.mjs`：通过。
- `node .\scripts\sync-web-static.mjs`：通过。
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`：通过，已验证 `/healthz`、`/`、`/manifest.json`、`/api/v1/user/login`、`/api/v1/channel/create`、`/api/v1/group/create`、`/api/v1/apikey/create`、`/v1/chat/completions`。
- 关键入口复查：`main.go`、`cmd/start.go`、`cmd/healthcheck.go`、`internal/server/server.go`、`internal/relay/relay.go`、`internal/server/handlers/setting.go`、`internal/server/handlers/route_target.go`、`internal/task/dynamic.go`、`web/src/components/modules/setting/DynamicRouting.tsx`、`web/src/components/modules/setting/Backup.tsx`。

### 失败项

- `node .\scripts\verify-backup-logic.mjs`：失败，英文断言与默认 `zh-Hans` 输出不一致。
- `node .\scripts\verify-backup-component.cjs`：失败，脚本按旧英文标签查找元素，和当前中文可见文案失配。
- `Set-Location web; node .\node_modules\next\dist\bin\next build`：源码编译和 TypeScript 阶段通过，但最终在当前 Windows 主机上以 `spawn EPERM` 结束。
- `Vitest` 启动：同样在当前主机上因为 `esbuild`/子进程 `spawn EPERM` 无法启动，`scripts/vitest-no-spawn.cjs` 也不足以绕过该问题。

### 未验证项

- `Docker` 相关验证未运行：当前主机 `docker` 不在 PATH，无法执行 `scripts/smoke-docker-compose.sh`。
- Linux smoke 未在本机执行：当前是 Windows 主机，只阅读了 `scripts/smoke-linux-backend.sh` 与 workflow。
- 远端 `origin/upstream` 现场状态未重新抓取：`git remote show origin` 和 `git remote show upstream` 在当前主机分别受证书/权限限制，只能基于本地已存在的 `origin/dev` ref 做基线比较。
- `next build`/`vitest` 的最终成功性在当前主机上无法完全证明，原因是宿主环境 `spawn EPERM`，不是前端源码编译错误。

## 4. Comparison Notes

### 当前工作区 vs `HEAD`

- 当前工作区是一个非常脏的审计对象。`git diff --stat HEAD` 显示 110 个已跟踪文件差异、`13516` 行新增、`2685` 行删除，另有大量未跟踪测试、脚本、迁移与文档文件。
- 最近未提交改动主要集中在：`internal/relay/*`、`internal/server/handlers/*`、`internal/task/*`、`web/src/components/modules/setting/*`、`web/src/components/modules/group/Editor.tsx`、`web/src/components/modules/channel/Form.tsx`、`README*`、workflow 与脚本。
- 这意味着本轮结论针对的是“当前工作区”，不是纯 `HEAD` 提交状态。

### 当前分支 vs 稳定基线

- 由于当前主机无法正常完成 `git remote show origin/upstream`，本轮稳定基线使用本地已存在的 `origin/dev` ref。
- `git diff --stat origin/dev...HEAD` 显示当前分支相较 `origin/dev` 已引入 provider 模板、Copilot/Antigravity、频道表单扩展、DocModal、CC Switch、README 增强等大块能力；当前工作区又在此基础上叠加了大量未提交变更。

### 代码实现 vs README / docs / 任务说明

- 一致：README 已把动态路由真实口径降级为 `dynamic summary scan`，这和 `internal/task/dynamic.go`、`DynamicRouting.tsx` 是一致的。
- 不一致：`docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md` 与 `docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md` 仍按 AI/多模式目标描述能力，明显领先于现有实现。
- 不一致：`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` 仍保留已过期的 `manifest.contains_secrets` 结论与环境限制说明。
- 不一致：`README.md` / `README_zh.md` 顶层安装和发布说明仍有占位符，和仓库实际可运行程度不匹配。

### 代码实现 vs 测试 / 构建 / 验证覆盖

- 一致：Go 侧主流程、handler、relay、task、backup 后端与 Windows smoke 是实打实跑通的，不是伪实现。
- 不一致：前端 backup 两条验证命令已经纳入 CI，但当前脚本仍按旧英文文案断言；它们现在验证的是“旧文案约定”，不是“当前实现正确性”。
- 不一致：CI 只跑定向 Go 包测试，而仓库已经存在更多包级测试却没有被 workflow 覆盖。
- 未定：`next build` 与 `vitest` 在当前 Windows 主机被 `spawn EPERM` 阻断，无法据此认定前端源码有回归，只能认定“本机验证可移植性仍不稳定”。

## 5. Top Next Actions

1. 先修 `scripts/verify-backup-logic.mjs` 和 `scripts/verify-backup-component.cjs`，让它们与当前 locale 策略和 `Backup.tsx` 的测试钩子对齐，再复跑 `.github/workflows/validation.yaml` 等价命令。
   - 最直接的做法是：脚本显式传 `locale: 'en'` 给 `backup-logic` helper，或把断言改成与默认 `zh-Hans`/`HiddenTestText` 契约一致。

2. 明确动态路由的对外交付口径。
   - 要么把 `docs/DYNAMIC_ROUTING_REQUIREMENTS*.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN*.md` 下调到“摘要扫描 + 预算 + 运行时 race tuning”真实范围；要么继续补实现，把 AI/mode surface 真正接入设置与运行时。

3. 完成发布收口和可复现性收口。
   - 固定 `release.yaml` 的 Go/Node/pnpm 版本。
   - 清掉 README 占位符。
   - 移除 `web/package.json` 的硬编码 `devs` 路径。
   - 在具备 Docker/Linux 能力的主机上补跑 `smoke-linux-backend.sh` 与 `smoke-docker-compose.sh`。

### 需要改的东西 & 要求总结

- 必改：backup 验证脚本与当前 UI/locale 契约统一，否则 CI 仍会因假阴性失败。
- 必改：动态路由子域文档不能继续超前于 shipped 行为，否则完成度评估会持续失真。
- 建议尽快改：release workflow 版本固定、README 占位符、`devs` 硬编码路径。
- 建议后续补：为 transformer TODO 边界补直接测试，或明确降级为“暂不支持”。

---

## 中文摘要

1. 本次触发时间
   - `2026-04-23T04:48:23+08:00`

2. 做了哪些检查、运行了哪些命令
   - 读取并比对了 `git status`、最近提交、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、README/状态文档/动态路由文档/workflow/关键主流程代码。
   - 实际运行了：`go build ./...`、`go test ./... -count=1`、`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\verify-locale-consistency.mjs`、`node .\scripts\sync-web-static.mjs`、`node .\scripts\verify-backup-logic.mjs`、`node .\scripts\verify-backup-component.cjs`、`web next build`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`，并额外尝试了本机 Vitest 启动。

3. 修改了哪些文件
   - 新增审查报告：`docs/review/审查/2026-04-23-044822-octopus-repo-complete-audit.md`
   - 新增可视化报告：`docs/review/审查/2026-04-23-044822-octopus-repo-complete-audit.html`

4. 发现了什么问题
   - 最大问题不是后端主流程坏掉，而是前端 backup 两条 CI 验证脚本已经和当前 locale/UI 实现漂移，当前工作区会直接失败。
   - 动态路由子域文档仍明显超前于实现，真实 shipped 能力只有摘要扫描和预算/健康开关。
   - release workflow 还在用浮动版本，README 仍有发布/安装占位符，发布收口不完整。
   - 协议转换仍有真实 TODO 边界；CI 的 Go 测试范围也窄于仓库现有测试面。

5. 本次结果是成功、跳过还是失败
   - 成功。审计任务已完成，并已落盘 `md + html` 报告；部分验证命令按预期暴露了当前问题和环境限制。

6. 是否需要我手动介入
   - 需要。如果接下来要进入“可合并/可发布”状态，建议优先手动决定两件事：
   - 第一，backup 验证脚本是统一改成当前 `zh-Hans` 口径，还是改成显式英文 locale 运行。
   - 第二，动态路由是按当前 shipped 行为下调文档承诺，还是继续补 AI/mode 实现。
