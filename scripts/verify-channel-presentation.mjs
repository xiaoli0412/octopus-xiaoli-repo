import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

function read(relativePath) {
  return fs.readFileSync(path.join(repoRoot, relativePath), 'utf8');
}

const cardSource = read('web/src/components/modules/channel/Card.tsx');
const cardContentSource = read('web/src/components/modules/channel/CardContent.tsx');
const channelIndexSource = read('web/src/components/modules/channel/index.tsx');
const formSource = read('web/src/components/modules/channel/Form.tsx');
const keyLabelSource = read('web/src/components/modules/channel/key-label.ts');
const toolbarSource = read('web/src/components/modules/toolbar/index.tsx');
const toolbarStoreSource = read('web/src/components/modules/toolbar/view-options-store.ts');

assert.match(cardSource, /const keyCount = channel\.keys\?\.length \?\? 0;/);
assert.match(cardSource, /t\('keyCount', \{ count: keyCount \}\)/);
assert.match(cardSource, /t\('keyCountBadge', \{ count: keyCount \}\)/);
assert.doesNotMatch(cardSource, /Key:\s*\{/);

assert.match(cardContentSource, /t\('keySummaryLine', \{/);
assert.match(cardContentSource, /const keyReadinessSummary = useMemo\(\(\) => \{/);
assert.match(cardContentSource, /ready: keys\.filter\(\(key\) => key\.channel_key\.trim\(\)\)\.length/);
assert.match(cardContentSource, /pending: keys\.filter\(\(key\) => !key\.channel_key\.trim\(\)\)\.length/);
assert.match(cardContentSource, /unchecked: keys\.filter\(\(key\) => !key\.status_code\)\.length/);
assert.match(cardContentSource, /attention: keys\.filter\(\(key\) => key\.status_code > 0 && key\.status_code !== 200\)\.length/);
assert.match(cardContentSource, /t\('keyReadiness\.ready'\)/);
assert.match(cardContentSource, /t\('keyReadiness\.pending'\)/);
assert.match(cardContentSource, /t\('keyReadiness\.unchecked'\)/);
assert.match(cardContentSource, /t\('keyReadiness\.attention'\)/);
assert.match(cardContentSource, /t\('labels\.maskedKey'\)/);
assert.match(cardContentSource, /labels\.enabledState/);
assert.match(cardContentSource, /t\('labels\.neverUsed'\)/);
assert.match(cardContentSource, /t\('labels\.routeTargetOverridesCount'\)/);
assert.match(cardContentSource, /t\('statusBadge\.notChecked'\)/);
assert.match(cardContentSource, /t\('statusBadge\.available'\)/);
assert.match(cardContentSource, /t\('actions\.routeTargetSummary', \{/);
assert.match(cardContentSource, /const routeTargetOverridesByKeyId = useMemo\(\(\) => \{/);
assert.match(cardContentSource, /routeTargetOverridesByKeyId\.get\(key\.id\) \?\? \[\]/);
assert.match(cardContentSource, /formatBillingModeLabel\(row\.billing_mode\)/);
assert.match(cardContentSource, /formatProbePolicyLabel\(row\.probe_policy\)/);
assert.match(cardContentSource, /t\('actions\.routeTargetKeyEmpty'\)/);
assert.match(cardContentSource, /t\('actions\.routeTargetKeyPreviewMore', \{ total: keyRouteTargetOverrides\.length \}\)/);
assert.match(cardContentSource, /getChannelKeyLabel\(key, \{ fallbackLabel: t\('keyFallbackLabel'\) \}\)/);
assert.match(channelIndexSource, /data-testid="channel-page"/);
assert.match(channelIndexSource, /data-layout=\{layout\}/);
assert.match(channelIndexSource, /renderItem=\{\(item\) => <Card channel=\{item\.raw\} stats=\{item\.formatted\} layout=\{layout\} \/>\}/);
assert.match(cardSource, /data-testid=\{`channel-card-trigger-\$\{channel\.id\}`\}/);
assert.match(cardSource, /data-testid=\{`channel-card-\$\{channel\.id\}`\}/);
assert.match(cardSource, /data-channel-name=\{channel\.name\}/);
assert.match(cardSource, /data-testid=\{`channel-card-badges-\$\{channel\.id\}`\}/);
assert.match(cardSource, /data-testid=\{`channel-card-metrics-\$\{channel\.id\}`\}/);
assert.match(cardSource, /data-testid=\{`channel-detail-dialog-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-detail-view-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-routing-section-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-route-target-summary-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-filter-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-accordion-\$\{channel\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-item-\$\{key\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-trigger-\$\{key\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-models-\$\{key\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-route-target-\$\{key\.id\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-route-target-row-\$\{key\.id\}-\$\{row\.model_name\}`\}/);
assert.match(cardContentSource, /data-testid=\{`channel-key-test-results-\$\{key\.id\}`\}/);

assert.match(channelIndexSource, /channelProviderFilter/);
assert.match(channelIndexSource, /channelModelKeyword/);
assert.match(channelIndexSource, /channelKeyKeyword/);
assert.match(channelIndexSource, /getProviderFilterKey/);
assert.match(channelIndexSource, /getProviderFilterKey\(channel\.raw\.type\) === providerFilter/);

assert.match(formSource, /getChannelKeyLabel\(key, \{ fallbackLabel: t\('keyFallbackLabel'\) \}\)/);
assert.match(formSource, /buildChannelKeyLabelMap\(formData\.keys \?\? \[], \{ fallbackLabel: t\('keyFallbackLabel'\) \}\)/);
assert.match(formSource, /t\('routeTargetSummary', \{/);
assert.match(formSource, /tDetail\('keySummaryLine', \{/);
assert.match(keyLabelSource, /fallbackLabel\?: string;/);
assert.match(keyLabelSource, /const fallbackLabel = options\?\.fallbackLabel\?\.trim\(\) \|\| 'Key';/);

assert.match(toolbarStoreSource, /export type ChannelProviderFilter = 'all' \| 'openai' \| 'anthropic' \| 'gemini' \| 'volcengine' \| 'github-copilot' \| 'antigravity' \| 'zen';/);
assert.match(toolbarStoreSource, /channelProviderFilter: 'all',/);
assert.match(toolbarStoreSource, /channelModelKeyword: '',/);
assert.match(toolbarStoreSource, /channelKeyKeyword: '',/);
assert.match(toolbarStoreSource, /clearChannelDetailFilters/);

assert.match(toolbarSource, /CHANNEL_PROVIDER_FILTER_OPTIONS/);
assert.match(toolbarSource, /channelProviderFilterLabelKeys/);
assert.match(toolbarSource, /t\('popover\.filter\.channelProviderTitle'\)/);
assert.match(toolbarSource, /t\('popover\.filter\.channelModelTitle'\)/);
assert.match(toolbarSource, /t\('popover\.filter\.channelKeyTitle'\)/);
assert.match(toolbarSource, /setChannelProviderFilter/);
assert.match(toolbarSource, /setChannelModelKeyword/);
assert.match(toolbarSource, /setChannelKeyKeyword/);
assert.match(toolbarSource, /clearChannelDetailFilters/);
assert.match(toolbarSource, /data-testid=\{toolbarItem === 'channel' \? `toolbar-channel-filter-\$\{option\.value\}` : undefined\}/);
assert.match(toolbarSource, /data-testid=\{`toolbar-channel-provider-\$\{option\.value\}`\}/);
assert.match(toolbarSource, /data-testid="toolbar-channel-model-keyword"/);
assert.match(toolbarSource, /data-testid="toolbar-channel-key-keyword"/);
assert.match(toolbarSource, /data-testid="toolbar-channel-clear-detail-filters"/);

for (const localeName of ['zh-Hans', 'zh-Hant', 'en', 'ja']) {
  const localeSource = read(`web/public/locale/${localeName}.json`);
  assert.match(localeSource, /"keyCount":/);
  assert.match(localeSource, /"keyCountBadge":/);
  assert.match(localeSource, /"keyFallbackLabel":/);
  assert.match(localeSource, /"keySummaryLine":/);
  assert.match(localeSource, /"keyReadiness":/);
  assert.match(localeSource, /"ready":/);
  assert.match(localeSource, /"pending":/);
  assert.match(localeSource, /"unchecked":/);
  assert.match(localeSource, /"attention":/);
  assert.match(localeSource, /"enabledState":/);
  assert.match(localeSource, /"enabledOn":/);
  assert.match(localeSource, /"enabledOff":/);
  assert.match(localeSource, /"maskedKey":/);
  assert.match(localeSource, /"neverUsed":/);
  assert.match(localeSource, /"routeTargetOverridesCount":/);
  assert.match(localeSource, /"statusBadge":/);
  assert.match(localeSource, /"notChecked":/);
  assert.match(localeSource, /"available":/);
  assert.match(localeSource, /"authFailed":/);
  assert.match(localeSource, /"rateLimited":/);
  assert.match(localeSource, /"upstreamError":/);
  assert.match(localeSource, /"requestError":/);
  assert.match(localeSource, /"warning":/);
  assert.match(localeSource, /"routeTargetKeyEmpty":/);
  assert.match(localeSource, /"routeTargetKeyPreviewMore":/);
  assert.match(localeSource, /"channelProviderTitle":/);
  assert.match(localeSource, /"channelProvider":/);
  assert.match(localeSource, /"channelModelTitle":/);
  assert.match(localeSource, /"channelModelPlaceholder":/);
  assert.match(localeSource, /"channelKeyTitle":/);
  assert.match(localeSource, /"channelKeyPlaceholder":/);
  assert.match(localeSource, /"channelClearDetailFilters":/);
  assert.match(localeSource, /"multiKeySummaryLine":/);
}

console.log('channel-presentation verification passed');
