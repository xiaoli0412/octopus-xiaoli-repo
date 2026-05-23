# 前端手工验收清单

> 目的：为 Phase G 最终收口提供仓库内固定的桌面端与 `375px` 移动端人工验收入口。
>
> 口径：本清单只记录需要人工确认的复杂交互与视觉可用性；不替代自动 smoke、Go/TypeScript 测试或 Docker/Linux 验证。

---

## 1. 使用前提

- 已完成最小运行链路：后端可启动、前端壳可访问、基础 smoke 不报启动级错误
- 本轮若只完成代码闭环但未实际执行以下检查，结果必须记为“未验收”，不能口头宣称“全部完成”

建议先执行：

- Windows 本地 smoke：`powershell -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`
- Linux smoke：`./scripts/smoke-linux-backend.sh`
- Docker compose smoke：`./scripts/smoke-docker-compose.sh`

---

## 2. 检查视口

- 桌面端：常规浏览器宽屏
- 移动端：`375px`

所有页面都需要至少检查一次这两个视口。

---

## 3. 首页 Home

- 页面可正常加载，不出现白屏或启动级错误
- 页面顺序保持 `Total -> Activity -> StatsChart -> Rank -> TokenBreakdown`
- `by_provider`、`by_channel`、`by_model` 三类 token breakdown 都能渲染
- `StatsChart` 可见 token 用量与渠道用量摘要
- `Rank` 可见渠道维度 token / 用量信息
- 成功/失败摘要、estimated official/gateway price、estimated probe cost、breaker 摘要、recent probe 摘要都能正常显示
- 空态、加载态、窄屏排版不破坏主要卡片布局

通过标准：

- 首页主卡片可见且无阻塞性交互错误
- `375px` 下无明显横向溢出

---

## 4. 渠道页 Channel

- 渠道卡片可展开，key 子卡片信息完整可见
- key 搜索、折叠、单 key 测试入口可正常使用
- `stats/overview` 为单栏信息流，不再出现分裂双栏摘要
- `AI 动态` 单栏运维区块可见，且不破坏概括页对齐
- 渠道表单测试不再因默认零值导致统一报 `channel is disabled`
- `classified / pooled` 与 key 路由策略标识可见
- 长列表和移动端下没有明显错位或操作遮挡

通过标准：

- 关键卡片操作可完成
- `375px` 下按钮、输入框、折叠区域仍可操作

---

## 5. 分组页 Group

- 左栏 `provider -> key -> model` 结构可读
- 桌面端列表为 3 栏，卡片密度仍可读，没有明显稀疏空洞
- 新分组模式 `AI 动态` 可见、可选、可保存
- 模式切换、拖拽排序、权重编辑不出现阻塞性错误
- 分组编辑器在桌面和 `375px` 下都能保持基本可理解性

通过标准：

- 结构层级可辨认
- 编辑器不出现无法提交或无法操作的严重问题

---

## 6. 备份页 Backup

- 导出默认是全量快照，默认包含明文凭据；显式关闭 secrets 后才是脱敏语义
- 长说明文案已压缩，不再占据主交互区
- dry-run、`replace / merge / map / skip` 模式切换都可触发正确界面状态
- `map` 模式下结构化映射编辑、建议映射、missing/unused mapping 风险提示可见
- dry-run 结果中的 preview token 绑定说明、Apply Same Import 风险确认、回滚快照预览可见
- rollback/import preview 可见 prune impact、route diff、mapping 或 alias 变化、skipped targets、risk summaries

通过标准：

- 备份/导入主流程可解释且无阻塞性渲染错误
- `375px` 下导入表单、风险卡片、回滚预览仍可阅读和点击

---

## 7. 日志页 Log

- SSE 实时流与历史列表可加载
- JSON / JSONL 导出入口、时间范围校验、下载行为可正常触发
- attempts 展示、超大 JSON 弹层、无限滚动没有明显崩溃或布局破坏

通过标准：

- 日志页可用
- 导出与大对象展示不会导致页面进入不可恢复状态

---

## 8. 设置页 Setting

- 保持原版双栏瀑布流，不出现大块占位空洞
- 语言切换入口可见，并支持 `zh-Hans / zh-Hant / ja / en`
- `API Docs / CC Switch` 入口可见，且可打开完整交互
- `Model Probe` 搜索可用，展开区域改为卡内滚动，不再整页拉长
- `动态路由`、`模型价格`、`模型探测策略` 的说明文案已压缩为短句，不再出现聊天式长说明
- `AI Automation` 入口与来源切换卡保持轻入口定位，不承担重控制台职责

通过标准：

- 入口存在且交互链不断
- `375px` 下设置卡片仍可滚动阅读和操作

---

## 9. AI 自动化页 AI Automation

- 顶层路由仍为 `ai`，页面进入后首屏结构是 `单 AI 主链 + 右侧状态栏 + 页内二级视图工作区`
- 首屏主链不再出现大段说明卡、宣传卡、折叠空白或长介绍文本
- 主链包含：任务类型、输入区、执行模型/来源、运行按钮、当前结果
- 当前结果默认是摘要视图，并可切换到 `对比 / 原始`
- 右侧状态栏可见：当前执行模型、来源模式、当前激活 Profile、学习状态、最近快照、最近任务
- 页内二级视图可切换到：`运行结果 / AI Profile / 多 AI 调度 / 模板与工具 / 历史与学习`
- `AI Profile` 区域可查看结构化摘要、状态、置信度、更新时间，并可手动激活
- 结果生成 Profile 后，可在结果区一键切换到该 Profile，但不会自动改动当前运行来源
- 多 AI 调度区可访问、可执行，但不抢占首屏主链焦点
- 学习区、模板区、工具区、历史区、原始 JSON、样本列表都在固定容器内滚动，不会把整页无限撑高
- 中文界面无英文硬编码、无 key 泄漏、无乱码、无明显错行或边界穿透
- `375px` 下主链、状态栏和二级视图切换仍可操作，不出现横向溢出
- 版本显示与前后端构建注入口径统一为 `1.19.1`

通过标准：

- 页面结构、语言和交互链符合重建后的 MVP 工作台口径
- 运行结果、Profile、多 AI、学习、模板、工具、历史都可触达且可操作

---

## 10. 记录方式

每次人工验收后，至少记录以下信息到当轮 worklog 或交付说明：

- 执行日期
- 执行环境
- 已检查页面
- 是否覆盖桌面端与 `375px`
- 通过项
- 阻塞项
- 尚未执行的外部 gate

---

## 11. 外部 Gate 说明

以下项目在真实跑通前，一律只能记为“待外部 gate”，不能记为“全部完成”：

- Windows smoke
- Linux / WSL smoke
- Docker compose smoke
- GitHub Actions `validation.yaml` 观察到绿色
- 浏览器桌面端与 `375px` 手工 checklist

补充口径：Docker Hub 官方安装方案已废弃。本轮若涉及安装、引导或发布说明，人工验收时必须确认页面、文档和脚本都没有再把 Docker Hub 作为默认或官方推荐来源。
