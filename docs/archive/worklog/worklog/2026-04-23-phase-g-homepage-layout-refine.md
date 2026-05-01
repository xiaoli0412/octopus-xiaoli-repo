# 2026-04-23 Phase G Homepage Layout Refine

## 1. 任务信息

- 任务名称：首页桌面端错位、占地与中文主显示返工
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：首页截图级结构返工与 no-browser 回归保护

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.5`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `11.2`、`11.3`、`11.4` 节
- 本次任务目标：
  - 收口首页统计卡片桌面端比例和首屏占地
  - 把 Token 明细改成默认摘要 + 按需展开运行摘要
  - 清理首页中文路径中的 `Token / Gateway / Tokens` 等主显示英文
  - 补 homepage no-browser 结构验证脚本
- 本次已盘点本地资源：
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - 现有 locale consistency / channel / model 验证脚本
  - 首页现有组件：`home/index.tsx`、`home/total.tsx`、`home/token-breakdown.tsx`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：首页返工与 locale/验证脚本高度耦合，主线程串行更稳妥

## 3. 本次硬规则

- 不改掉原项目深绿圆角风格
- 不把运行时价格、熔断、探测首屏全展开
- 中文首页主显示不再夹杂无必要英文
- 返工后必须补自动化回归保护，避免后续再回退

## 4. 实现范围

- 受影响前端模块：
  - `web/src/components/modules/home/index.tsx`
  - `web/src/components/modules/home/total.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
- 受影响 locale：
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/en.json`
  - `web/public/locale/ja.json`
- 受影响脚本：
  - `scripts/verify-locale-consistency.mjs`
  - `scripts/verify-home-layout.mjs`（新增）

## 5. 具体改动

1. 首页总览卡片重排
   - 把首页顶部改成“请求主卡 + 总量/输入/输出三张摘要卡”。
   - 请求主卡保留请求次数、耗时、成功、失败、成功率，但改成更稳定的桌面端列数。
   - 右侧三张摘要卡在不同断点下切换为更稳妥的单列/双列，减少压缩和错位。

2. Token 明细区收口
   - 默认首页只展示总量、输入、输出、估算网关价格四张摘要。
   - 三组 provider/channel/model 列表默认只展示前 3 项，通过按钮展开更多。
   - 估算价格、熔断状态、最近探测统一放入“运行摘要”折叠区，默认不占首屏高度。

3. 首页整体布局调整
   - 活动热力图从左列下方移到主网格之后，避免桌面端首屏被过高模块挤压。
   - 图表与排行榜继续保留在右侧，整体保持原有风格但收紧比例。

4. 首页中文主显示清理
   - `Token 明细` 改为 `令牌明细` / `令牌明細` / `トークン明細`。
   - `Gateway 价格` 改为 `网关价格` / `網關價格` / `ゲートウェイ価格`。
   - `输入 Tokens / 输出 Tokens` 改为对应语言下的统一主显示。

5. 自动化保护
   - 新增 `scripts/verify-home-layout.mjs`。
   - 校验首页默认结构、折叠态、locale key 与活动热力图区位，避免后续回退成大块堆叠布局。
   - `verify-locale-consistency.mjs` 扩展到把 `home` 模块纳入中文/英文/日文泄漏检查。

## 6. 测试与验证

- `pnpm exec tsc --noEmit`
- `node scripts/verify-home-layout.mjs`
- `node scripts/verify-locale-consistency.mjs`
- `node scripts/verify-channel-create-flow.mjs`
- `node scripts/verify-llm-price-boundary.mjs`
- `pnpm run build:static`

结果：以上命令均通过。

## 7. 风险与遗留项

- 当前首页结构返工已完成，但仍缺真实浏览器截图证据。
- 375px、hover/focus 与更细的错位观察仍需在下一轮补齐。
- 中文首页已清掉本轮新增主显示英文，但全仓中文路径仍需继续扫尾。

## 8. 下一轮建议

1. 用真实浏览器补首页桌面端与 375px 截图证据。
2. 继续清理其它中文路径下的英文主显示，尤其设置页深层和备份页细节态。
3. 回到渠道 key 级观测字段与状态提示补强。
