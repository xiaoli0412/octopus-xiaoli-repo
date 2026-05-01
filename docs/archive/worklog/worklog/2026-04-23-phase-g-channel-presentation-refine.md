# 2026-04-23 Phase G Channel Presentation Refine

## 1. 任务信息

- 任务名称：渠道页多 Key 展示、四语统一与自动化回归护栏收口
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：渠道页 screenshot-first 收口与 no-browser 回归保护

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1`、`9.1.1`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.4`、`11.2`、`11.3`、`11.4` 节
- 本次任务目标：
  - 收口渠道卡片与详情区的多 Key 摘要展示
  - 补齐 `zh-Hans / zh-Hant / en / ja` 四语的渠道页新增字段
  - 清理渠道页及同批可见路径里的 `Key / Token` 残留词
  - 新增 no-browser 渠道展示验证脚本，避免后续回退
- 本次已盘点本地资源：
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - 现有 `verify-locale-consistency.mjs`、`verify-channel-create-flow.mjs`
  - 现有渠道模块：`Card.tsx`、`CardContent.tsx`、`Form.tsx`、`key-label.ts`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：本轮改动集中在同一组渠道组件与四语 locale，串行更稳妥

## 3. 本次硬规则

- 不改掉原项目深绿圆角风格
- 不把多 Key 高级项默认全摊开
- 中文、繁中、日文路径都要保持同语种统一
- 修完展示必须补自动化回归保护，避免再次回退

## 4. 实现范围

- 受影响前端模块：
  - `web/src/components/modules/channel/Card.tsx`
  - `web/src/components/modules/channel/CardContent.tsx`
  - `web/src/components/modules/channel/Form.tsx`
  - `web/src/components/modules/channel/key-label.ts`
- 受影响 locale：
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/en.json`
  - `web/public/locale/ja.json`
- 受影响脚本：
  - `scripts/verify-locale-consistency.mjs`
  - `scripts/verify-channel-presentation.mjs`（新增）

## 5. 具体改动

1. 渠道卡片摘要收口
   - 渠道卡片顶部副标题改为“渠道类型 + 密钥数量”。
   - 卡片 badge 区统一展示 key 模式、路由策略、密钥数量，不再出现裸 `Key:` 语义。

2. 渠道详情区 Key 摘要重排
   - Key 列表顶部新增摘要线，显示总数、启用数和当前筛选命中数。
   - 每个 Key 的折叠头现在统一展示：备注或回退标签、启用状态、状态 badge、来源类型、密钥预览、成本、允许模型数量。
   - 展开区补齐“启用状态、状态文本、来源类型、允许模型、最近使用/尚未使用”。

3. 四语 locale 收口
   - `zh-Hant` 补齐 `keyCount`、`keyCountBadge`、`keyFallbackLabel`、`keySummaryLine`、`statusBadge.*`、`enabledState`、`maskedKey`、`neverUsed` 等新增字段。
   - `en` 补齐渠道页新增字段，确保新的 Key 摘要、状态 badge 和回退标签有完整英文文案。
   - `ja` 补齐新增字段，并把渠道页与同批会直接显示的 `Token / Key` 旧表达收口为统一日文语义。
   - `zh-Hans` 同步把渠道表单中的 `intervals` 表达改成中文“秒”语义，避免中文路径漏英文缩写。

4. locale 一致性脚本升级
   - `verify-locale-consistency.mjs` 重写并纳入 `channel.card` 模块检查。
   - 新增渠道页关键字段断言：`keyCountBadge`、`maskedKey`、`statusBadge.notChecked`、`metrics.totalToken` 等。
   - 新增中英混排拦截规则，避免 `zh-Hant / ja` 回退到 `Key / Token` 旧词。

5. 新增渠道展示回归脚本
   - 新增 `scripts/verify-channel-presentation.mjs`。
   - 校验渠道卡片不再回退成裸 `Key:`。
   - 校验渠道详情区确实使用 `keySummaryLine / labels.maskedKey / labels.enabledState / labels.neverUsed / statusBadge.*`。
   - 校验 `key-label.ts` 和 `Form.tsx` 都使用了 `fallbackLabel`。

## 6. 测试与验证

- `pnpm exec tsc --noEmit`
- `node scripts/verify-locale-consistency.mjs`
- `node scripts/verify-channel-create-flow.mjs`
- `node scripts/verify-channel-presentation.mjs`

结果：以上命令均通过。

## 7. 风险与遗留项

- 本轮完成的是 no-browser 结构与四语一致性收口，真实浏览器层的桌面端比例、375px、hover/focus 证据仍未补齐。
- 渠道页筛选维度还未扩展到“提供商 / 模型家族名 / 更强 key 模糊搜索”的最终版本。
- 渠道页与首页之外的其它深层页面，仍可能残留少量语言不统一词，需要继续按 screenshot-first 池扫尾。

## 8. 下一轮建议

1. 用真实浏览器补渠道页桌面端和 375px 的截图证据。
2. 继续推进用户反复强调的“同渠道内多 Key + 独立模型集合 + 渐进展开”交互细化。
3. 顺着本轮 locale 护栏继续清理分组、价格、备份等深层页面的语言残留。

## 9. 补充收口

- 同日继续把渠道列表筛选从“单搜索词 + 启用状态”收口为“启用状态 + 提供商 + 模型关键词 + key 信息”的组合筛选。
- 同日继续给多 Key 管理区补上“总数 / 已填写真实密钥 / 已启用 / 待补充”的摘要线，减少用户在折叠态下的判断成本。
- 本次补充改动已再次通过前端类型检查、渠道展示验证、渠道创建流程验证、首页布局验证和多语言一致性验证。
