# 2026-04-23 Deep Audit README Placeholder And Import Multipart File Cap

## 1. 任务信息

- 任务名称：README 发布占位符清理与 setting/import multipart 文件上限收口
- 日期：2026-04-23
- 当前阶段：Phase A / release-readiness and trust-boundary deep audit
- 对应 milestone：small-fix release-readiness + import resource-boundary closure

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：安全默认、发布可控、可回归验证相关章节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 中“高风险优先、小幅修复、定向验证、worklog/memory 回写”
- 上一个相关 worklog：`docs/worklog/2026-04-23-deep-audit-providers-pinned-commit-and-source-observability.md`
- 本次任务目标：
  - 清理 README / README_zh 中残留的发布与安装占位符
  - 复审 `internal/server/handlers/setting.go` 的 multipart 导入链路，并补一层单文件读取上限
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`README.md`、`README_zh.md`、`internal/server/handlers/setting.go`、`internal/server/handlers/setting_import_test.go`、`docker-compose.yml`、`.github/workflows/release.yaml`
- 本次使用的本地 resources / skills / 记忆上下文：现有 README/workflow 给出了真实仓库 URL、GHCR 镜像地址和 release 页面；现有 setting/import 测试给出了 multipart 体积边界的低风险补点位置
- 若未使用部分本地资源或上下文，原因：未展开浏览器/前端主线，因为本轮聚焦 release 文档与后端导入边界
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：不适用
- 若未使用子 agent，原因：任务规模小且依赖当前主线程上下文，直接在主线程完成更稳

## 3. 本次硬规则

- 不做大范围 README 重写，只清理占位符并对齐当前仓库真实路径
- 不改 `/api/v1/setting/import` 的核心语义，只增加单文件读取上限保护
- 所有改动必须有定向验证

## 4. 本次禁止事项

- 不重构 import 流程
- 不改前端 UI 或浏览器验收链
- 不覆盖已有 unrelated dirty workspace 变更

## 5. 本次验收条件

- `README.md` / `README_zh.md` 不再包含 `<CURRENT_...>` 或 TODO 占位符
- multipart 导入路径在外层 request body 限制之外，还对单个上传文件读取做独立上限控制
- 存在对应回归测试
- `go test ./internal/server/handlers -count=1` 通过

## 6. 本次回滚点

- 回滚 `README_zh.md`
- 回滚 `internal/server/handlers/setting.go`
- 回滚 `internal/server/handlers/setting_import_test.go`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 release 文档与后端 import 资源边界
- 受影响后端模块：`internal/server/handlers/setting.go`、`internal/server/handlers/setting_import_test.go`
- 受影响前端模块：无
- 受影响接口：`POST /api/v1/setting/import`
- 是否影响旧数据：否
- 是否影响旧行为：仅在 multipart 单文件超过限制时更早返回 `413 import payload too large`，其余行为不变

## 8. 实施步骤

1. 复核 README / README_zh 的真实占位符残留情况，并与 release workflow / compose / 仓库 URL 对齐。
2. 发现英文 README 在当前工作区已被其他同主题改动清理，本轮实际补齐的是中文 README 的残余占位符。
3. 在 `setting.go` 的 multipart 分支中，把 `fh.Open()` 后的读取改为 `io.LimitReader(..., maxDBImportFileBytes+1)`，并在超限时直接返回 `413`。
4. 新增 `TestImportDBRejectsOversizedMultipartFilePart` 验证“总请求体尚可，但单文件读取超限”场景。

## 9. 测试与验证

- 构建命令：未执行全量 `go build ./...`
- 测试命令：`go test ./internal/server/handlers -count=1`
- 专项验证：
  - `gofmt -w internal/server/handlers/setting.go internal/server/handlers/setting_import_test.go`
  - `Select-String -Path README.md,README_zh.md -Pattern '<CURRENT_|TODO' -CaseSensitive`

## 10. 风险与兼容性

- 新风险：无新增高风险
- 兼容性风险：multipart 上传在极端大文件场景下会更早被 `413` 拒绝，属于期望的收紧
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：未执行全量构建；handlers 定向测试通过
- 测试是否通过：通过，`go test ./internal/server/handlers -count=1` 通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、README/README_zh、release workflow、compose、setting/import 现有测试
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - README / workflow / compose 提供了当前真实仓库 URL、GHCR 镜像和 releases 地址
  - 现有 import 测试说明已有 request-body 超限覆盖，但缺单文件读取超限覆盖
  - automation memory 给出了“README 占位符 + import resource review”是当前优先顺序
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：不适用
- 手工 smoke 状态：未执行浏览器或 Docker smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮改动只涉及文档与 handlers 小修，无需扩大到浏览器/Docker；当前主机对 Docker 也不友好
- 待验证页面清单：无新增页面待验证
- 若未使用子 agent，原因：任务范围小，主线程直接完成更快且更稳
- worklog 是否更新：是
- 遗留项：`setting/import` 仍值得再看 multipart metadata / preview token 以外的资源放大边界；README 之外的 release-ready drift 仍有动态路由文档口径问题
- 下一任务前置条件是否满足：是
