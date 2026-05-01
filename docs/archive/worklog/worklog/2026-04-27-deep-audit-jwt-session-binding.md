# 2026-04-27 深度审查：管理 JWT 会话绑定收口

- Task: deep audit and low-risk fix for management JWT session invalidation
- Canonical refs:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- Milestone / Phase: Phase A security boundary hardening
- Master plan aligned before coding (yes/no): yes

## 本轮直接复用的本地资源

- `$CODEX_HOME/automations/octopus/memory.md`：继承上一轮的首次登录强制改密审查结果，避免重复审查同一问题。
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`：确认当前自动化要优先做安全边界、低风险小修和可验证收口。
- `docs/worklog/2026-04-22-auth-and-container-hardening-audit.md`：沿用容器/鉴权深审主线，继续补管理鉴权边界。
- `internal/server/auth/auth.go`、`internal/server/handlers/user.go`、`internal/op/user.go` 及对应测试：作为本轮认证链路事实依据。

## Findings 与处理

1. `internal/server/auth/auth.go`
   - 问题：管理 JWT 仅校验签名、算法和 issuer，没有绑定当前管理员会话状态。
   - 影响：管理员改密码、改用户名，或首次强制改密完成后，已签发的旧管理 token 在过期前仍可继续访问管理接口，导致认证边界收口不完整。
   - 证据：`GenerateJWTToken`/`VerifyJWTToken` 之前不携带也不校验与当前管理员状态相关的 claims；`internal/server/handlers/user.go` 与 `internal/op/user.go` 中的改用户名、改密码、强制改密都不会轮换独立会话版本，因此旧 token 会一直有效到 `exp`。
   - 处理：已修复。JWT 现在额外绑定 `subject=username` 和基于 `username + password hash + must_change_password + auth_token_secret` 计算的会话指纹；当前管理员状态发生变化后，旧 token 会立即失效。

## 本轮改动文件

- `internal/server/auth/auth.go`
- `internal/server/auth/auth_test.go`
- `internal/server/handlers/user_test.go`

## 验证

- 已执行：`gofmt -w internal/server/auth/auth.go internal/server/auth/auth_test.go internal/server/handlers/user_test.go`
- 已执行：`git diff --check -- internal/server/auth/auth.go internal/server/auth/auth_test.go internal/server/handlers/user_test.go`
- 未完成：`go test ./internal/server/auth ./internal/server/handlers ./internal/server/middleware -count=1`
  - 原因：当前环境网络受限，Go 依赖下载在 `proxy.golang.org` DNS 解析阶段失败，属于宿主环境阻塞，不是这次改动触发的编译/测试失败。

## 剩余风险与下一轮建议

- `internal/server/handlers/providers.go` 的 providers 远端刷新虽然已固定到 pinned commit，但仍依赖 GitHub raw 获取，下一轮适合继续审查供应链稳定性与 fallback 策略。
- 容器与 compose 主线仍值得继续深审最小权限与只读运行面，例如进一步验证 Docker 运行期是否可安全启用 `read_only`/`tmpfs` 等约束。
- 当前无法在本机完成 Go 包测试下载链验证；若后续本地 module cache 恢复，可优先复跑认证/handlers/middleware 相关测试确认补丁完整通过。
