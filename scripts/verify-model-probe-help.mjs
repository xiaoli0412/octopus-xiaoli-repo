import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const componentPath = path.join(repoRoot, 'web/src/components/modules/setting/ModelProbe.tsx');
const source = fs.readFileSync(componentPath, 'utf8');

assert.match(source, /data-testid="setting-model-probe-card"/);
assert.match(source, /t\('defaultPathTitle'\)/);
assert.match(source, /t\('defaultPathDesc'\)/);
assert.match(source, /t\('defaultPathHint'\)/);
assert.match(source, /data-testid="setting-model-probe-default-path"/);
assert.match(source, /data-testid="setting-model-probe-search"/);
assert.match(source, /data-testid="setting-model-probe-toggle"/);
assert.match(source, /data-testid="setting-model-probe-scroll-region"/);
assert.match(source, /data-testid="setting-model-probe-model-list"/);
assert.match(source, /data-testid="setting-model-probe-collapsed-state"/);
assert.match(source, /data-testid="setting-model-probe-empty-state"/);
assert.match(source, /t\('summaryCards\.modelCountLabel'\)/);
assert.match(source, /t\('summaryCards\.defaultPolicyLabel'\)/);
assert.match(source, /t\('summaryCards\.intervalLabel'\)/);
assert.match(source, /t\('summaryCards\.overrideLabel'\)/);
assert.match(source, /t\('guidanceTitle'\)/);
assert.match(source, /t\('guidanceHint'\)/);
assert.match(source, /t\('advancedPanelTitle'\)/);
assert.match(source, /t\('advancedPanelDesc'\)/);
assert.match(source, /t\('summaryIntervalValue'/);
assert.match(source, /summaryModels\.length === 0 \? '-'/);
assert.match(source, /policySummary\./);
assert.match(source, /DEFAULT_VISIBLE_MODEL_COUNT = 12/);
assert.match(source, /const shouldFetchModels = showModelList \|\| hasKeyword;/);
assert.match(source, /const \{ data: models \} = useModelList\(\{ enabled: shouldFetchModels \}\);/);
assert.match(source, /shouldRenderModelRows \? t\('collapseModels'\) : t\('showModels'\)/);
assert.match(source, /t\('collapsedState'\)/);
assert.match(source, /showMoreModels/);
assert.match(source, /visibleModels\.slice\(0, visibleCount\)/);
assert.match(source, /setVisibleCount\(DEFAULT_VISIBLE_MODEL_COUNT\)/);
assert.match(source, /overscroll-contain/);

const zhHans = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hans.json'), 'utf8'));
const zhHant = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hant.json'), 'utf8'));
const en = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/en.json'), 'utf8'));
const ja = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/ja.json'), 'utf8'));

for (const locale of [zhHans, zhHant, en, ja]) {
  const modelProbe = locale.setting.modelProbe;
  assert.ok(modelProbe.defaultPathTitle);
  assert.ok(modelProbe.defaultPathDesc);
  assert.ok(modelProbe.defaultPathHint);
  assert.ok(modelProbe.defaultBadge);
  assert.ok(modelProbe.summaryPolicyHint);
  assert.ok(modelProbe.summaryCards.modelCountLabel);
  assert.ok(modelProbe.summaryCards.defaultPolicyLabel);
  assert.ok(modelProbe.summaryCards.intervalLabel);
  assert.ok(modelProbe.summaryCards.overrideLabel);
  assert.ok(modelProbe.summaryCards.overrideValue);
  assert.ok(modelProbe.guidanceTitle);
  assert.ok(modelProbe.guidanceHint);
  assert.ok(modelProbe.advancedPanelTitle);
  assert.ok(modelProbe.advancedPanelDesc);
  assert.ok(modelProbe.collapseModels);
  assert.ok(modelProbe.showModels);
  assert.ok(modelProbe.collapsedState);
  assert.ok(modelProbe.showMoreModels);
  assert.ok(modelProbe.policySummary.passive_only);
}

assert.doesNotMatch(zhHans.setting.modelProbe.defaultPathTitle, /Default/i);
assert.doesNotMatch(zhHant.setting.modelProbe.advancedPanelTitle, /Advanced/i);
assert.equal(en.setting.modelProbe.summaryCards.overrideValue, 'Channel override wins');
assert.equal(zhHans.setting.modelProbe.description, '这里集中管理模型探测策略。');
assert.equal(en.setting.modelProbe.description, 'Manage model-level probe defaults here.');

console.log('model-probe-help verification passed');
