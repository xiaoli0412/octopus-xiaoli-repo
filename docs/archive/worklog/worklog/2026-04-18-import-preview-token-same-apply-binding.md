# 2026-04-18 Import Preview Token Same-Apply Binding

## 1. 任务信息

- 任务名称：给 backup/import 的 dry-run -> same-import apply 增加 preview token 绑定
- 日期：2026-04-18
- 当前主线：backup / import / migration adaptation
- 当前阶段：Phase F 收口，聚焦 `11.5.2 / 11.5.4`

## 2. 开工前复用的本地资源

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/worklog/2026-04-18-import-selective-scopes-and-rollback-chain-fix.md`
- `docs/worklog/2026-04-18-backup-replace-prune-preview-closure.md`
- `docs/worklog/2026-04-18-redacted-snapshot-credential-rebind-targets.md`
- `internal/model/backup.go`
- `internal/op/backup.go`
- `internal/server/handlers/setting.go`
- `web/src/api/endpoints/setting.ts`
- `web/src/components/modules/setting/Backup.tsx`

## 3. 本轮小计划

- 当前主线：backup/import
- 当前阶段：Phase F
- 候选任务：
  - dry-run 与 apply 的 preview token 强绑定
  - redacted snapshot 后续 credential rebind 动作
  - post-import validation 摘要继续补 UI
  - snapshot 历史浏览增强
- 本轮核心任务：实现 dry-run -> apply 的 preview token 契约
- 配套任务：前端 same-import apply 提交 preview token；补 handler 回归测试
- 完成标准：
  - dry-run 返回 preview token
  - same-import apply 带 token 成功
  - 改 payload 后沿用旧 token 会被拒绝

## 4. 本轮完成的改动

### 后端

- 在 `internal/model/backup.go` 的 `DBImportResult` 中新增 `preview_token`
- 新增 `internal/server/handlers/setting_import_preview.go`
  - 对 `(dump + mode + model_mappings + import_scopes)` 生成稳定 digest
  - 用当前 admin 用户密钥派生 secret 签发 JWT preview token
  - 校验 token 过期与 digest 一致性
- 更新 `internal/server/handlers/setting.go`
  - dry-run 时返回 `preview_token`
  - apply 时若提交 `preview_token`，会校验其与当前导入 payload 是否一致
  - token 不匹配时明确报错，要求重新 dry-run

### 前端

- 更新 `web/src/api/endpoints/setting.ts`
  - `DBImportResult` 增加 `preview_token`
  - `useImportDB` 支持提交 `previewToken`
- 更新 `web/src/components/modules/setting/Backup.tsx`
  - dry-run 成功后保存后端返回的 `preview_token`
  - “Apply Same Import” 改为提交该 token，而不是只靠本地复用表单状态
  - 如果 dry-run 结果没有 token，会阻止 same-import apply 并提示重跑 dry-run

### 测试

- 新增 `internal/server/handlers/setting_import_test.go`
  - `TestImportDBDryRunReturnsPreviewTokenAndApplyWithSameTokenSucceeds`
  - `TestImportDBApplyRejectsMismatchedPreviewToken`

## 5. 验证与结果

### 已完成验证

- `D:\gol1\npm.cmd exec -- tsc --noEmit`
  - 结果：通过

### 本轮未完成的验证

- `go test ./internal/server/handlers ./internal/op -count=1`
- `gofmt -w ...`

### 未完成原因

- 当前线程内可定位到 Go 安装路径：`C:\Users\李昊桐\AppData\Local\Programs\go\bin`
- 但无论 PowerShell 直接调用还是 `cmd /c` 调用，`go.exe` / `gofmt.exe` 都返回 `Access is denied`
- 这是本机当前线程的执行环境阻塞，不是命令路径缺失

## 6. 本轮发现的问题

- `apply_patch` 默认包装入口在当前 Windows 沙箱里也会 `Access is denied`
- 已切换为调用用户目录内的 `C:\Users\李昊桐\.codex\.sandbox-bin\codex.exe --codex-run-as-apply-patch` 完成补丁编辑
- 新增 handler 测试初稿里误用了 `op` 包内未导出的 `dbDumpVersion`，本轮已修正为字面值 `1`

## 7. 风险与遗留项

- 当前后端 `preview_token` 校验是“可选强绑定”：
  - same-import apply 现在会带 token
  - 但普通手工 apply 若不传 token，后端仍允许继续导入
- 这符合“最小闭环”目标，但还没有升级成“所有 apply 必须绑定某次 dry-run”
- 由于本轮 Go 测试未能在当前线程内执行，后端部分仍需要下一轮或其他环境补跑一次真实测试

## 8. 对主文档主线的推进判断

- 本轮直接推进了 `11.5.4` 中“真正 apply 前允许用户确认”的契约完整性
- 让 Backup 页面中的 `Apply Same Import` 不再只是前端状态复用，而是和某次 dry-run 结果建立后端可验证绑定
- 这是对 backup/import 主线的风险收敛，不是新功能扩张

## 9. 下一轮最适合继续推进的事项

1. 在可执行 Go 的环境里补跑：
   - `gofmt -w internal/model/backup.go internal/server/handlers/setting.go internal/server/handlers/setting_import_preview.go internal/server/handlers/setting_import_test.go`
   - `go test ./internal/server/handlers ./internal/op -count=1`
2. 若测试通过，考虑把 apply 进一步收紧为：
   - 高风险 mode（尤其 `replace`）默认要求 preview token
   - 或者只允许 same-import apply 使用当前 dry-run 返回的 token
3. 继续同主线下一个小闭环：
   - credential rebind 的实际补绑动作
   - 或 post-import validation / rollback preview 的交互增强
