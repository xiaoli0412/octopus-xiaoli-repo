import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

function read(relativePath) {
  return fs.readFileSync(path.join(repoRoot, relativePath), 'utf8');
}

const routeConfig = read('web/src/route/config.tsx');
const appContainer = read('web/src/components/app.tsx');
const homeIndex = read('web/src/components/modules/home/index.tsx');
const channelIndex = read('web/src/components/modules/channel/index.tsx');
const groupIndex = read('web/src/components/modules/group/index.tsx');
const modelIndex = read('web/src/components/modules/model/index.tsx');
const settingIndex = read('web/src/components/modules/setting/index.tsx');
const channelDetail = read('web/src/components/modules/channel/CardContent.tsx');

assert.match(routeConfig, /\{ id: 'home'/);
assert.match(routeConfig, /\{ id: 'channel'/);
assert.match(routeConfig, /\{ id: 'group'/);
assert.match(routeConfig, /\{ id: 'model'/);
assert.match(routeConfig, /\{ id: 'ai'/);
assert.match(routeConfig, /\{ id: 'log'/);
assert.match(routeConfig, /\{ id: 'setting'/);

assert.match(appContainer, /max-w-6xl/);
assert.match(appContainer, /md:grid-cols-\[auto_1fr\]/);

assert.match(homeIndex, /<Total \/>/);
assert.match(homeIndex, /<Activity \/>/);
assert.match(homeIndex, /<StatsChart \/>/);
assert.match(homeIndex, /<Rank \/>/);
assert.doesNotMatch(homeIndex, /TokenBreakdown/);
assert.doesNotMatch(homeIndex, /home-main-grid/);

assert.match(channelIndex, /columns=\{\{ default: 1, md: 2, lg: 2 \}\}/);
assert.match(groupIndex, /columns=\{\{ default: 1, md: 2, lg: 2 \}\}/);
assert.match(modelIndex, /lg: layout === 'list' \? 1 : 2/);

assert.match(settingIndex, /columns-1 gap-4 pb-24 md:columns-2 md:pb-4/);
assert.doesNotMatch(settingIndex, /DeferredSection/);
assert.doesNotMatch(settingIndex, /IntersectionObserver/);

assert.doesNotMatch(channelDetail, /xl:grid-cols-\[minmax\(0,0\.82fr\)_minmax\(0,1\.18fr\)\]/);
assert.match(channelDetail, /data-testid=\{`channel-route-target-summary-\$\{channel\.id\}`\}/);
assert.match(channelDetail, /data-testid=\{`channel-key-accordion-\$\{channel\.id\}`\}/);
assert.match(channelDetail, /keyFilterPlaceholder/);
assert.match(channelDetail, /handleTestKey/);

console.log('ui-baseline-layout verification passed');
