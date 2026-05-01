# Octopus 完整审计报告

- 本次触发时间：2026-04-29 13:03:11 +08:00
- 仓库路径：`D:\GPT-codex\octopus_repo`
- 审计分支：`feat/erguotou`
- 审计提交：`bfa27aecbec14329a43550c0d5409aefea257af0`
- 对比基线：`origin/dev`

## 1. Findings

1. `High` Antigravity 被 README 和 UI 文案宣称为“内建默认 OAuth / 开箱即用”，但后端默认实现仍要求显式 `client_id/client_secret`，默认路径会直接失败。  
判断依据：README 在 [README.md](/D:/GPT-codex/octopus_repo/README.md:285) 写着 `works out-of-the-box`，同时又在 [README.md](/D:/GPT-codex/octopus_repo/README.md:294) 和 [README.md](/D:/GPT-codex/octopus_repo/README.md:295) 把 `OCTOPUS_ANTIGRAVITY_CLIENT_ID` / `...CLIENT_SECRET` 标成 required；后端 [antigravity.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/antigravity.go:197) 缺少 `client_id` 时直接返回错误，[antigravity.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/antigravity.go:312) 在 callback 里缺少 `client_id/client_secret` 也会把 session 标成失败。  
影响：这是用户可见主路径失配。Copilot 的“默认可用”是真实现；Antigravity 当前仍是“有 OAuth 流程，但默认配置并不自足”。

2. `Medium` Windows 本地运维脚本 `runtime-win.ps1 -Action healthcheck` 对继承的坏 Go 环境不够稳，可能在真正发起健康检查前就被环境变量打断。  
判断依据：`runtime-win.ps1` 只在 `GOCACHE/GOTMPDIR` 为空时兜底，见 [runtime-win.ps1](/D:/GPT-codex/octopus_repo/scripts/runtime-win.ps1:292) 和 [runtime-win.ps1](/D:/GPT-codex/octopus_repo/scripts/runtime-win.ps1:297)；它随后直接执行 `go run main.go healthcheck`，见 [runtime-win.ps1](/D:/GPT-codex/octopus_repo/scripts/runtime-win.ps1:606)。本轮实测第一次被继承的不可写 `D:\DevCache\go-build` 卡住，第二次又被异常 `GOPROXY` 卡住；而同仓库的 [smoke-win-backend.ps1](/D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1:129) 到 [smoke-win-backend.ps1](/D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1:132) 会先把 `GOCACHE/GOTMPDIR/TEMP/TMP` 规范化到可写目录，并成功跑通最小主链。  
影响：这不会破坏 Octopus 服务器本身，但会让 Windows 本地“查看健康 / 验证运行态”的官方入口出现假失败，增加排障成本。

3. `Medium` 协议转换主链可用，但高级边界仍保留真实 TODO，相关分支缺少直接回归证明。  
判断依据：`internal/transformer/model/model.go` 在 [142](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:142)、[208](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:208)、[618](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:618)、[861](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:861) 仍保留 audio / prediction / JSON schema / image-generation 相关 TODO；[anthropic inbound](/D:/GPT-codex/octopus_repo/internal/transformer/inbound/anthropic/messages.go:162) 还写着 `support other result types`；[gemini outbound](/D:/GPT-codex/octopus_repo/internal/transformer/outbound/gemini/messages.go:344) 仍未完成 JSON schema 转换。虽然本轮 `go test ./... -count=1` 已通过，但没有额外证据证明这些 TODO 分支已被覆盖。  
影响：对主流聊天/嵌入/基本协议互转来说，这更像“边界部分完成”而非主链失效；但若对外继续强调“无缝协议转换”，这些边界应先补测试或明确标注暂不支持。

4. `Low` `.gitignore` 过窄，当前工作区噪音已经明显干扰审计与回归判断。  
判断依据：根目录 `.gitignore` 只有 [1](/D:/GPT-codex/octopus_repo/.gitignore:1) 到 [5](/D:/GPT-codex/octopus_repo/.gitignore:5) 五行；本轮 `git status --short` 共有 `425` 行，包含大量 `.tmp*`、cache、`.next`、`node_modules`、patch、日志和工具输出。  
影响：这不直接破坏运行时，但会持续提高误提交本地产物、误读 diff、漏看真实业务回归的概率。

## 2. Completion Assessment

总体判断：

- 按 Octopus 核心网关主线看，当前完成度约 `88%-90%`。启动链、管理 API、中继 API、构建链、主要协议主路径、Windows 后端最小 smoke、Go 全量测试、`tsc`、静态导出和大量 no-browser 校验都已真实可达。
- 按最近扩展能力看，完成度更接近 `78%-82%`。Copilot / Antigravity / provider 模板 / AI 自动化 / 动态路由学习 / 备份增强都已有主体实现，但部分承诺偏强或边界仍未收口。

已完成：

- `cmd/start.go -> server -> handlers -> relay` 主入口真实可达，`go build ./...` 和 `go test ./... -count=1` 均通过。
- 前端 `tsc --noEmit`、`build-web-static.mjs`、以及 repo-local no-browser 校验链在当前工作区通过。
- AI 自动化任务、历史列表、Profile 持久化、typed payload 回填、激活与恢复链都已接线，不再属于“未实现”。
- 动态路由学习开关、摘要、列表、重置与设置页入口形成了真实 UI/API 闭环。
- 备份/导入/恢复-ready 快照主链是真实现，相关逻辑与 no-browser 验证已通过。

部分完成：

- Antigravity OAuth 流程存在，但“默认 OAuth app 开箱即用”并未真实成立。
- 协议转换高级边界仍留 TODO，主链可用但边角能力未完全闭环。
- Windows 运行态验证脚本对坏宿主 Go 环境的自愈性不足。

未完成和疑似空实现：

- 本轮没有再确认到“UI/路由已声明但主流程完全未调用”的纯空实现。
- 需要收缩表述或补实现的主要是 Antigravity 默认 OAuth 承诺，而不是空壳接口。

## 3. Verification Summary

已验证项：

- Git / 版本上下文：`git status --short --branch`、`git branch -vv`、`git log --oneline --decorate -n 18`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`git diff --stat origin/dev...HEAD`。
- Go 构建和测试：在把 `GOCACHE/GOMODCACHE/GOTMPDIR` 指到仓库可写目录后，`go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1`、`go test ./... -count=1` 全部通过。
- 前端类型和构建：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`、`node scripts/build-web-static.mjs` 通过。
- Windows 后端运行态 smoke：`powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 通过，覆盖 `/`、`/manifest.json`、`/healthz`、登录、建 channel/group/apikey、以及 `/v1/chat/completions` 网关转发。
- no-browser 校验：`verify-locale-consistency`、`verify-home-layout`、`verify-channel-create-flow`、`verify-channel-presentation`、`verify-group-create-flow`、`verify-llm-price-boundary`、`verify-backup-logic`、`verify-backup-component`、`verify-help-hint-accessible`、`verify-dynamic-routing-help`、`verify-circuit-breaker-help`、`verify-model-probe-help`、`verify-setting-info-logic`、`verify-ai-config-profile-summary`、`verify-ai-automation-learning-focus`、`verify-ccswitch-flow`、`verify-route-target-copy` 全部通过。

未验证项：

- `pnpm run test` / Vitest：本轮在 `corepack` 缓存改到仓库内后，仍会在 `vitest.config.ts` 装载阶段被 `esbuild spawn EPERM` 卡住，属于宿主对子进程/worker 池的限制，无法在本机证明真实单测结果。
- Linux backend smoke：`bash --version` 当前就因 Git Bash 信号管道权限问题失败，无法在这台宿主执行 `scripts/smoke-linux-backend.sh`。
- Docker compose smoke：`docker` 命令不可用，无法执行 `scripts/smoke-docker-compose.sh`。
- `runtime-win.ps1 -Action healthcheck`：产品 `go run main.go healthcheck` 本身可运行并能正确返回“当前 8080 没启动”，但脚本入口会被继承环境中的坏 `GOCACHE/GOPROXY` 放大成假失败，因此本轮不把该脚本结果当作产品回归证据。
- 手工前端 checklist：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md` 本轮未做人眼验收。

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 工作区非常脏，`git status --short` 达到 `425` 行；但这次没有再确认到“当前工作区已经打断 Go build / Go test / tsc / static build”的硬回归。
- 相对 `HEAD` 的大量未提交内容里，既有真实业务修改，也有明显宿主/缓存/临时产物噪音；`.gitignore` 过窄是持续增噪源。

当前分支与稳定基线：

- `origin/dev...HEAD = 0 behind / 22 ahead`，说明分支本身不是落后上游的破损分叉。
- 真正的风险来自当前超大未提交工作区，而不是提交分支落后。

代码实现与 README / docs / 任务说明：

- 多 AI split mode 的当前文档已经明说“前端并行拉起多个 lane 任务，不承诺存在服务端 orchestrator”，见 [AI 自动化需求文档](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:68)；代码里也确实只是 `Promise.all(...)` 或串行多次 `createTask.mutateAsync(...)`，见 [AI 控制台](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:1515) 到 [AI 控制台](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:1559)，因此这条不再是正式不一致问题。
- AI Profile 文档写明 typed payload 是主消费合同，见 [AI 自动化需求文档](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:97)；当前代码确实优先读取 typed payload，但运行时真正抽取的仍是执行配置字段，见 [applyActiveAIProfileConfig](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:837) 和 [extractAIAutomationConfigFromPayload](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_profile_typed.go:359)。这说明 typed payload 并非空实现，但“主消费合同”目前更多体现在配置切换层，而非 rich domain payload 的广泛下游消费。
- 备份高级迁移仍属 partial 这一点，本轮未发现新的文档/实现冲突。
- Antigravity 是当前最明显的不一致点：README / locale 继续强调默认 OAuth，而后端实现并未提供对应默认值。

代码实现与测试/构建/验证覆盖：

- Go 主链、前端 `tsc`、静态导出、Windows 后端 smoke 和整组 no-browser 脚本已经与当前实现形成真实闭环。
- GitHub Actions 里已把 `vitest`、`build:static`、定向 Go 测试、Linux smoke、Docker smoke 纳入 gate；本地能覆盖其中大部分，但 Vitest/Linux/Docker 仍受宿主限制未能同机证明。

## 5. Top Next Actions

1. 修正 Antigravity 默认 OAuth 承诺。  
要求：二选一并同步 README、前端文案、后端实现。  
- 方案 A：真正提供默认 `client_id/client_secret` 或等效默认 app 注入，让“开箱即用”变成真。  
- 方案 B：撤回 `built-in default OAuth app / works out-of-the-box` 口径，明确要求管理员先配置环境变量。

2. 提升 Windows 本地运行态脚本的环境收敛能力。  
要求：让 `runtime-win.ps1 -Action healthcheck` 至少像 `smoke-win-backend.ps1` 一样，能把 `GOCACHE/GOTMPDIR/TEMP/TMP` 收敛到仓库可写目录；同时在必要时重置或显式覆盖坏 `GOPROXY`，避免脚本在真正发请求前假失败。

3. 为协议转换高级边界补直接证明。  
要求：优先围绕 `JSON schema`、Anthropic 复杂 `tool_result`、prediction、audio、image-generation` TODO` 分支补测试；如果短期不支持，就把 README/文档口径收窄到“主流链路已覆盖”。

建议下一步动作：

- 清理 `.gitignore`，把 `.tmp*`、cache、日志、`.next`、`.tools` 等明显宿主产物纳入忽略，降低工作区噪音。
- 在允许 `spawn` 的宿主或 CI 上补跑 Vitest；在可用 Linux/Docker 环境上补跑 `smoke-linux-backend.sh` 与 `smoke-docker-compose.sh`。

## 中文摘要

1. 本次触发时间：`2026-04-29 13:03:11 +08:00`
2. 做了哪些检查、运行了哪些命令：读取 automation memory、git 状态/分支/最近提交/与 `origin/dev` 差异，抽查 README、workflow、AI 自动化、动态路由、Antigravity、protocol transformer、Windows 脚本；运行了 `go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1`、`go test ./... -count=1`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`、`node scripts/build-web-static.mjs`、`powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`，以及 17 条 no-browser 校验脚本；同时尝试了 `pnpm run test`、`runtime-win.ps1 -Action healthcheck`、`bash --version`、`docker --version`、`wsl -l -v` 以确认环境限制。
3. 修改了哪些文件：新增 `docs/review/审查/2026-04-29-130311-octopus-repo-complete-audit.md`、`docs/review/审查/2026-04-29-130311-octopus-repo-complete-audit.html`，更新 automation memory；另外本轮曾被 `corepack` 自动写入 `web/package.json` 的 `packageManager` 字段，已当场撤回。除审查产物外，本次无业务变更。
4. 发现了什么问题：最重要的问题是 Antigravity 的“默认 OAuth / 开箱即用”承诺与后端实现不一致；其次是 `runtime-win.ps1` 的 healthcheck 对坏 Go 环境不够稳，以及协议转换高级边界仍保留 TODO 且缺少直接证明；另外 `.gitignore` 过窄持续制造噪音。
5. 本次结果是成功、跳过还是失败：`成功`。完整审计已完成并落盘；当前工作区的 Go 构建/测试、`tsc`、静态导出、Windows 后端 smoke 和大部分 repo-local 自动验证链均已通过。
6. 是否需要我手动介入：`需要`。优先决定 Antigravity 是补默认 OAuth 实现还是收缩对外文案；其次建议安排修补 `runtime-win.ps1` 的环境收敛逻辑，并在合适宿主上补 Vitest / Linux / Docker smoke。
