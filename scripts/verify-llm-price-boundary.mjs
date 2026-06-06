import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const componentPath = path.join(repoRoot, 'web/src/components/modules/setting/LLMPrice.tsx');
const source = fs.readFileSync(componentPath, 'utf8');
const modelIndexPath = path.join(repoRoot, 'web/src/components/modules/model/index.tsx');
const modelIndexSource = fs.readFileSync(modelIndexPath, 'utf8');
const modelItemPath = path.join(repoRoot, 'web/src/components/modules/model/Item.tsx');
const modelItemSource = fs.readFileSync(modelItemPath, 'utf8');
const toolbarPath = path.join(repoRoot, 'web/src/components/modules/toolbar/index.tsx');
const toolbarSource = fs.readFileSync(toolbarPath, 'utf8');
const toolbarViewOptionsPath = path.join(repoRoot, 'web/src/components/modules/toolbar/view-options-store.ts');
const toolbarViewOptionsSource = fs.readFileSync(toolbarViewOptionsPath, 'utf8');

assert.match(source, /data-testid="setting-llm-price-card"/);
assert.match(source, /t\('llmPrice\.hint'\)/);
assert.match(source, /t\('llmPrice\.defaultPathTitle'\)/);
assert.match(source, /t\('llmPrice\.probeRedirectTitle'\)/);
assert.match(source, /t\('llmPrice\.scopeCards\.probeValue'\)/);
assert.match(source, /t\('llmPrice\.updateInterval\.hint'\)/);
assert.match(source, /t\('llmPrice\.manualUpdate\.hint'\)/);
assert.match(source, /<HelpHint>\{t\('llmPrice\.hint'\)\}<\/HelpHint>/);

assert.match(modelIndexSource, /const modelDensity = useToolbarViewOptionsStore\(\(s\) => s\.modelDensity\);/);
assert.match(modelIndexSource, /const estimateItemHeight = modelDensity === 'compact' \? 126 : 156/);
assert.match(modelIndexSource, /useUpstreamPriceSummaries/);
assert.match(modelIndexSource, /const upstreamByModel = useMemo\(\(\) => \{/);
assert.match(modelIndexSource, /columns=\{\{ default: 1, md: 2, lg: 3, xl: 3, '2xl': 3 \}\}/);
assert.match(modelIndexSource, /gap=\{modelDensity === 'compact' \? 10 : 12\}/);
assert.match(modelIndexSource, /const canonicalName = \(m\.canonical_name \?\? ''\)\.toLowerCase\(\);/);
assert.match(modelIndexSource, /return m\.name\.toLowerCase\(\)\.includes\(term\) \|\| canonicalName\.includes\(term\);/);
assert.match(toolbarSource, /const searchPlaceholderKey = toolbarItem === 'channel'/);
assert.match(toolbarSource, /const searchInputWidthClass = toolbarItem === 'channel'/);
assert.match(toolbarViewOptionsSource, /getLayout: \(item\) => get\(\)\.layouts\[item\] \|\| \(item === 'model' \? 'list' : 'grid'\),/);
assert.match(toolbarViewOptionsSource, /channelDensity: 'normal',/);
assert.match(toolbarViewOptionsSource, /modelDensity: 'compact',/);
assert.match(toolbarSource, /placeholder=\{t\(searchPlaceholderKey\)\}/);
assert.match(toolbarSource, /data-slot="toolbar-search-expanded"/);
assert.match(toolbarSource, /'w-28 sm:w-44 md:w-52'/);
assert.match(toolbarSource, /'w-24 sm:w-36 md:w-40'/);
assert.doesNotMatch(toolbarSource, /className="w-20 bg-transparent text-sm outline-none placeholder:text-muted-foreground"/);
assert.match(toolbarSource, /data-testid=\{`toolbar-view-options-trigger-\$\{toolbarItem\}`\}/);
assert.match(toolbarSource, /data-testid=\{`toolbar-view-options-content-\$\{toolbarItem\}`\}/);
assert.match(toolbarSource, /data-testid=\{`toolbar-layout-grid-\$\{toolbarItem\}`\}/);
assert.match(toolbarSource, /data-testid=\{`toolbar-layout-list-\$\{toolbarItem\}`\}/);
assert.match(toolbarSource, /const layoutSectionLabel = isDensityToolbar \? t\('popover\.density'\) : t\('popover\.layout'\);/);
assert.match(toolbarSource, /const densityButtonMap: Record<'grid' \| 'list', CardDensity>/);
assert.match(modelItemSource, /data-slot=\{isCompact \? 'model-card-meta-compact' : 'model-card-meta'\}/);
assert.match(modelItemSource, /label: tSetting\('canonicalName'\)/);
assert.match(modelItemSource, /label: tSetting\('billingMode'\)/);
assert.match(modelItemSource, /label: tSetting\('officialPricePair'\)/);
assert.match(modelItemSource, /metaItems\.map\(\(item\) => \(/);
assert.match(modelItemSource, /key: 'canonical'[\s\S]*key: 'billing'[\s\S]*key: 'official'/);
assert.match(modelIndexSource, /data-testid="model-page"/);
assert.match(modelIndexSource, /data-layout="grid"/);
assert.match(modelIndexSource, /data-density=\{modelDensity\}/);
assert.match(modelIndexSource, /renderItem=\{\(model, index\) => <ModelItem model=\{model\} upstreamPrice=\{upstreamByModel\.get\(model\.name\.toLowerCase\(\)\)\} density=\{modelDensity\} index=\{index\} \/>\}/);
assert.match(modelItemSource, /data-testid=\{typeof index === 'number' \? `model-card-\$\{index\}` : undefined\}/);
assert.match(modelItemSource, /data-slot="model-card"/);
assert.match(modelItemSource, /data-layout="grid"/);
assert.match(modelItemSource, /data-density=\{density\}/);
assert.match(modelItemSource, /const isCompact = density === 'compact';/);

const zhHans = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hans.json'), 'utf8'));
const zhHant = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hant.json'), 'utf8'));
const en = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/en.json'), 'utf8'));

for (const locale of [zhHans, zhHant, en]) {
  const llmPrice = locale.setting.llmPrice;
  assert.ok(llmPrice.hint);
  assert.ok(llmPrice.defaultPathTitle);
  assert.ok(llmPrice.defaultPathDesc);
  assert.ok(llmPrice.scopeCards.pricingLabel);
  assert.ok(llmPrice.scopeCards.probeValue);
  assert.ok(llmPrice.scopeCards.syncHint);
  assert.ok(llmPrice.probeRedirectTitle);
  assert.ok(llmPrice.probeRedirectDesc);
  assert.ok(llmPrice.updateInterval.hint);
  assert.ok(llmPrice.manualUpdate.hint);
}

assert.equal(en.setting.llmPrice.scopeCards.probeValue, 'Moved to Model Probe Policy');
assert.doesNotMatch(zhHans.setting.llmPrice.hint, /probe/i);
assert.doesNotMatch(zhHant.setting.llmPrice.hint, /probe/i);
assert.match(zhHans.setting.llmPrice.scopeCards.probeValue, /\u6a21\u578b\u63a2\u6d4b\u7b56\u7565/);
assert.match(zhHant.setting.llmPrice.scopeCards.probeValue, /\u6a21\u578b\u63a2\u6e2c\u7b56\u7565/);
assert.match(en.setting.llmPrice.probeRedirectDesc, /Model Probe Policy/);
assert.match(zhHans.toolbar.popover.normal, /^\u666e\u901a$/);
assert.match(zhHans.toolbar.popover.compact, /^\u7d27\u51d1$/);
assert.equal(zhHans.toolbar.searchPlaceholder.model, '搜索模型名称或规范名称');
assert.equal(zhHant.toolbar.searchPlaceholder.model, '搜尋模型名稱或規範名稱');
assert.equal(en.toolbar.searchPlaceholder.model, 'Search by model or canonical name');
assert.equal(zhHans.toolbar.searchPlaceholder.channel, '搜索渠道名称、供应商或密钥信息');
assert.equal(zhHans.toolbar.searchPlaceholder.group, '搜索分组名称或成员信息');

console.log('llm-price-boundary verification passed');
