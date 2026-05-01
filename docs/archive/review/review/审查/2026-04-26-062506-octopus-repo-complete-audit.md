# Octopus 完整审计报告

审计时间：2026-04-26 06:25:06 +08:00
仓库：`D:\GPT-codex\octopus_repo`

## 1. Findings

### High - 备份/导入验证当前是红的，核心测试套件被初始化策略和唯一约束冲突打断
`internal/op/op_test.go:13-29` 的 `setupOpTestDB` 会先 `InitDB()` 再 `InitCache()`，而 `InitCache()` 会自动创建默认设置和 bootstrap admin。`internal/op/backup_test.go` 里不少用例仍按“空库/无默认记录”写断言，于是出现一批稳定的失败：`TestDBExportAllExcludesInternalAuthTokenSecret`、`TestDBImportIncrementalDryRunBuildsStructuredCompatibility`、`TestDBImportIncrementalSkipModePreservesExistingRows`、`TestDBImportIncrementalReplaceModePrunesSettingsMissingFromSnapshot`、`TestDBRollbackLatestImportSnapshotRestoresPreviousState` 等。`go test ./...` 的输出里直接能看到 `settings.key` 唯一约束冲突、`rows_summary[users]` 断言偏差、以及回滚前置的用户名校验失败。

这意味着备份/导入链路虽然代码量很大，但当前测试层面不能证明它已经稳定可用，发布门禁也因此不可信。

### Medium - 粘性会话回归测试本身失效，不能再证明迭代器的 sticky 顺序
`internal/relay/balancer/iterator.go:37-54` 的 sticky 重排逻辑是在构造和 `Reset()` 时把 sticky 候选挪到首位；但 `internal/relay/balancer/iterator_test.go:54-80` 在 `it.Reset()` 后直接断言 `it.IsSticky()`，而 `IsSticky()` 依赖 `Iterator.index >= 0`。此时 `index` 仍然是 `-1`，所以测试会失败，即使 sticky 候选已经正确在首位。

结果是这条主路由的会话亲和性验证失去有效性，属于“实现可能还对，但回归测试已经不能证明它对”的状态。

### Low - 动态路由学习相关文档有轻微口径漂移
代码里 `internal/relay/dynamic_mode.go`、`internal/op/dynamic_route_learning.go`、`web/src/api/endpoints/ai-automation.ts` 已经实际消费 `dynamic_routing_learning_enabled`，但 `docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.md:79,156,190` 和中文版仍把它描述成“下一阶段”目标。这个漂移不会阻断运行，但会让文档读者误判当前能力边界。

## 2. Completion Assessment

整体完成度偏高，主线能力大体已接线：管理端、relay、动态路由、AI 自动化、备份导入导出、providers 和前端设置页都能在代码里找到闭环。真正拉低完成度的不是功能缺块，而是验证链路不健康：备份/导入相关测试大面积失败，sticky 迭代器测试失效，前端类型检查在当前环境里也没跑通。

我会把当前状态评估为：功能实现大多已完成，验证完整度仍然偏部分完成。

## 3. Verification Summary

已验证项：
- 读了 `git status --short --branch`、`git log --oneline -n 12`、`git diff --stat HEAD..origin/dev`，确认当前分支是 `feat/erguotou`，并且相对 `HEAD`、`origin/dev` 都有大规模变更。
- 检查了 `README.md`、`README_zh.md`、`CHANGELOG.md`，确认文档已经写入动态路由、AI 自动化、备份导入导出、providers、CC Switch、API Base URL 等主线说明。
- 复核了 `cmd/start.go`、`internal/server/server.go`、`internal/relay/relay.go`、`internal/server/handlers/channel.go`、`internal/server/handlers/setting.go`、`internal/server/handlers/providers.go`，主链路是接通的。
- 复核了 `internal/model/setting.go`、`internal/op/setting.go`、`web/src/api/endpoints/setting.ts`、`web/src/components/modules/setting/System.tsx`、`web/src/components/modules/setting/DynamicRouting.tsx`、`web/src/components/modules/navbar/DocModal.tsx`，前后端设置与文档入口大体一致。
- 复核了 `internal/relay/balancer/iterator.go`、`internal/relay/balancer/iterator_test.go`、`internal/relay/balancer/balancer.go`，sticky 逻辑实现存在，但测试断言方式失效。

未验证项：
- `go build ./...` 没有完整跑通，原因是当前环境的 Go 构建缺少可用模块缓存，且外网 `proxy.golang.org` 访问失败。
- `pnpm exec tsc --noEmit` 和 `pnpm test:settings-no-browser` 没有跑通，原因是当前可用的 `node.exe` 是 Codex 包装器，执行 pnpm 脚本时在 `node::InitializeOncePerProcessInternal` 触发 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`。
- 由于上面两个环境阻塞，前端 TS/无浏览器测试只能做代码抽查，不能做完整命令级验证。
- `go test ./...` 通过了大量包级测试，但 `internal/op` 和 `internal/relay/balancer` 有红例，说明当前验证还不能算通过。

## 4. Comparison Notes

当前工作区相对 `HEAD` 的改动非常大，既有后端中继、统计、备份、AI 自动化，也有前端设置/文档/国际化和大量新增测试。`origin/dev` 则更像上游稳定基线，和当前分支相比，已经吸收了 Copilot、Antigravity、providers、DocModal、动态路由、AI 自动化等扩展能力。

代码与文档方面，大方向基本一致：README 已经把动态路由口径收敛成 `dynamic summary scan`，也明确写了 `api_base_url`、备份导入/导出和手工验收清单。剩下的偏差主要是文档陈述滞后于代码，比如 `dynamic_routing_learning_enabled` 在文档里还偏“下一阶段”，但代码里已经开始真正使用。

代码与测试方面，当前最大的不一致是“实现看起来完整，但关键验证没有完全跟上”。备份导入/导出测试大量失败，说明这条发布最敏感路径的可验证性仍然不够；粘性迭代器测试也因为断言方式失效而失真。

## 5. Top Next Actions

1. 先修 `internal/op/backup_test.go` 的测试夹具与断言，明确区分“初始化默认 admin/默认 settings 后的状态”和“业务数据状态”，把 `settings.key` 唯一约束冲突清掉。
2. 修正 `internal/relay/balancer/iterator_test.go` 的 sticky 断言，让它在 `CandidateAt(0)` 或 `Next()` 之后验证，而不是在 `Reset()` 后直接调用 `IsSticky()`。
3. 给前端验证补一个可独立运行的 Node/pnpm 入口，或者把 CI / 本地脚本改成不依赖当前 Codex 包装器的 `node.exe`，否则 TS 校验和无浏览器测试仍然无法闭环。

## 中文摘要

- 触发时间：2026-04-26 06:25:06 +08:00
- 检查内容：仓库结构、`git status`、`git log`、`git diff`、README/CHANGELOG、后端主链路、设置/动态路由/providers、备份导入导出、sticky 迭代器、前端设置页与无浏览器验证入口。
- 运行命令：`git status --short --branch`、`git log --oneline --decorate -n 12`、`git diff --stat HEAD..origin/dev`、`go test ./...`、`go build ./...`、`pnpm exec tsc --noEmit` 尝试、`pnpm test:settings-no-browser` 尝试，以及多组 `rg` / `Get-Content` 抽查命令。
- 修改文件：本次新增 `docs/review/审查/2026-04-26-062506-octopus-repo-complete-audit.md` 和 `docs/review/审查/2026-04-26-062506-octopus-repo-complete-audit.html`。
- 发现问题：备份/导入测试套件红、sticky 迭代器测试失效、文档与代码在 `dynamic_routing_learning_enabled` 上有轻微口径漂移、前端验证受本地 Node 包装器断言阻塞。
- 结果：部分成功。代码审计完成，验证未完全闭环。
- 是否需要手动介入：需要，优先介入备份测试夹具和前端 Node/pnpm 验证环境。
