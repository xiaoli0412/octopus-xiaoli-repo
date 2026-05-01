# Octopus 完整审计报告

- 仓库路径 `D:\GPT-codex\octopus_repo`
- 分支 `feat/erguotou`
- 基线 `origin/dev`
- 生成时间 `2026-04-22 00:33:48 +08:00`

## 1. Findings

### High

route-target 级 override 没有真正回到并发竞速门控里。`shouldEscalateToRace` 只看 `allowsRacingByModel(requestModel)`，而 `allowsRacingByModel` 调的是 `ResolveRouteTargetPolicy(nil, dbmodel.ChannelKey{}, modelName)`。这会丢掉真实的 `(channel, key, model)` 上下文，导致具体 route-target 的 `per_request` / `per_token` / `probe_policy` 覆盖无法影响竞速。证据在 [internal/relay/relay.go](/D:/GPT-codex/octopus_repo/internal/relay/relay.go#L248)、[internal/relay/relay.go](/D:/GPT-codex/octopus_repo/internal/relay/relay.go#L536)、[internal/op/route_target.go](/D:/GPT-codex/octopus_repo/internal/op/route_target.go#L222)。

### Medium

协议转换层还留着真实 TODO 分支，而且没有看到对应的直接测试。`internal/transformer/model/model.go` 里有 `audio`、`prediction`、`Schema`、`image_generation` 相关 TODO，`internal/transformer/outbound/gemini/messages.go` 里还在 `json_schema` 上留了 TODO，`internal/transformer/inbound/anthropic/messages.go` 也只覆盖了单一 `tool_result` 形态。主链路可用，但这些高级格式仍是降级而不是完全翻译。证据在 [internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go#L142)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go#L208)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go#L618)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go#L861)、[internal/transformer/outbound/gemini/messages.go](/D:/GPT-codex/octopus_repo/internal/transformer/outbound/gemini/messages.go#L321)、[internal/transformer/inbound/anthropic/messages.go](/D:/GPT-codex/octopus_repo/internal/transformer/inbound/anthropic/messages.go#L162)。

### Low

`web/package.json` 里的 `devs` 脚本硬编码了证书路径，换机器或换工作区后就不具备可移植性。见 [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json#L8)。

README 还有占位链接没收口。见 [README.md](/D:/GPT-codex/octopus_repo/README.md#L51)、[README.md](/D:/GPT-codex/octopus_repo/README.md#L71)、[README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md#L51)、[README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md#L70)。

## 2. Completion Assessment

整体完成度我评成 **约 85%**。

已完成的部分很扎实。启动链、健康检查、静态资源、渠道与分组管理、route-target 覆盖、统计、日志导出、备份导入回滚、动态摘要扫描，这些都是真接线，不是空骨架。`cmd/start.go` 已经把任务调度挂到启动链上，`internal/task/init.go` 也确实注册了每日动态摘要扫描。

部分完成的主要是两块。第一块是协议转换边界。第二块是高级动态路由门控，尤其是 route-target 级 override 还没完全回到主竞速判断里。

尚未完成的是完整验证闭环和文档收口。当前机器没有可直接用的 `go`、`pnpm`、`docker` 命令，所以 Go build 和 Docker smoke 没法在这里复现。

疑似空实现的地方主要还是 transformer 里的 TODO 分支。它们不是整个模块都空，但确实是留着的真实缺口。

### 状态速览

| 类别 | 结论 |
|---|---|
| 已完成 | 启动链、健康检查、静态资源、渠道 / 分组 / 统计 / 日志 / 备份主流程 |
| 部分完成 | 高级动态路由、协议转换边界、首页统计扩展 |
| 未完成 | Go / Docker 全链路验证、部分协议转换边界测试 |
| 疑似空实现 | transformer 里的 `TODO` 分支 |

## 3. Verification Summary

已验证项

- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- `D:\gol1\node.exe --experimental-strip-types .\scripts\verify-backup-logic.mjs`
- `D:\gol1\node.exe .\scripts\verify-backup-component.cjs`
- `D:\gol1\node.exe .\web\node_modules\next\dist\bin\next build` 在 `web/` 目录里成功编译到生产阶段，但最后死在 `spawn EPERM`

未验证项

- `go build ./...`
- `go test ./...`
- `docker compose up -d --build`
- Linux backend smoke
- 更完整的协议转换边界测试

环境限制

- `go` 不在 PATH
- `pnpm` 不在 PATH
- `docker` 不在 PATH

## 4. Comparison Notes

- 当前分支是 `feat/erguotou`，HEAD 指向 `bfa27ae`，并跟踪 `origin/feat/erguotou`
- 相比 `origin/dev`，HEAD 额外包含一大批重构和功能扩展，`git diff --stat origin/dev..HEAD` 显示 72 个文件变化，约 `6001` 行新增、`1741` 行删除
- 当前工作区本身是脏的，`git status` 里有很多既存修改和未跟踪文件，我没有回滚这些内容
- 实现和文档在备份、route-target、动态摘要扫描、统计页这几块基本一致
- 实现和文档不一致的地方主要是协议转换边界、README 占位符、以及那个机器绑定的 HTTPS 开发脚本路径

## 5. Top Next Actions

1. 把 route-target 的真实上下文接回竞速门控，补一个覆盖 explicit override 的回归测试。
2. 收口 transformer 里的 TODO 分支，再补直接测试。
3. 去掉 `web/package.json` 里硬编码的 `devs` 路径。
4. 补一次可复现的 Go / Docker 验证链。

## 中文摘要

1. 本次触发时间 `2026-04-22 00:33:48 +08:00`
2. 做了仓库结构、git 状态、最近提交、README / docs、关键后端和前端入口、验证脚本、路由 / 备份 / 统计 / 协议转换链路检查；运行了 `git status`、`git log`、`git diff --stat`、`tsc --noEmit`、`verify-backup-logic.mjs`、`verify-backup-component.cjs`、`next build` 等命令
3. 本次修改了 [`docs/review/审查/2026-04-22-003348-octopus-repo-complete-audit.md`](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-22-003348-octopus-repo-complete-audit.md) 和 [`docs/review/审查/2026-04-22-003348-octopus-repo-complete-audit.html`](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-22-003348-octopus-repo-complete-audit.html)，业务代码本次无改动
4. 发现了 route-target 覆盖未接回竞速门控、协议转换 TODO 分支未收口、`devs` 脚本硬编码路径、README 占位内容未清理
5. 本次结果成功
6. 需要你手动介入，优先处理 route-target 竞速门控和 transformer TODO 分支
