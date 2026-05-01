import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const homeIndexPath = path.join(repoRoot, 'web/src/components/modules/home/index.tsx');
const totalPath = path.join(repoRoot, 'web/src/components/modules/home/total.tsx');
const tokenBreakdownPath = path.join(repoRoot, 'web/src/components/modules/home/token-breakdown.tsx');

const homeIndexSource = fs.readFileSync(homeIndexPath, 'utf8');
const totalSource = fs.readFileSync(totalPath, 'utf8');
const tokenBreakdownSource = fs.readFileSync(tokenBreakdownPath, 'utf8');

assert.match(homeIndexSource, /<Total \/>/);
assert.match(homeIndexSource, /<TokenBreakdown \/>/);
assert.match(homeIndexSource, /<StatsChart \/>/);
assert.match(homeIndexSource, /<Rank \/>/);
assert.match(homeIndexSource, /<Activity \/>/);
assert.match(homeIndexSource, /data-testid="home-page"/);
assert.match(homeIndexSource, /data-testid="home-main-grid"/);
assert.match(homeIndexSource, /data-testid="home-main-grid"[\s\S]*?<Activity \/>[\s\S]*?<StatsChart \/>[\s\S]*?<Rank \/>[\s\S]*?<TokenBreakdown \/>/);
assert.doesNotMatch(homeIndexSource, /home-breakdown-column|home-right-column/);

assert.match(totalSource, /data-testid="home-total-section"/);
assert.match(totalSource, /md:grid-cols-2 lg:grid-cols-4/);
assert.match(totalSource, /writing-mode:vertical-lr/);

assert.match(tokenBreakdownSource, /const \[showRuntimeDetails, setShowRuntimeDetails\] = useState\(false\);/);
assert.match(tokenBreakdownSource, /const \[showMoreLists, setShowMoreLists\] = useState\(false\);/);
assert.match(tokenBreakdownSource, /data-testid="home-breakdown-section"/);
assert.match(tokenBreakdownSource, /data-testid="home-runtime-toggle"/);
assert.match(tokenBreakdownSource, /data-testid="home-runtime-panel"/);
assert.match(tokenBreakdownSource, /summaryHint/);
assert.match(tokenBreakdownSource, /runtimePanelTitle/);
assert.match(tokenBreakdownSource, /runtimePanelHint/);
assert.match(tokenBreakdownSource, /showMoreLists/);
assert.match(tokenBreakdownSource, /hideMoreLists/);
assert.match(tokenBreakdownSource, /topItems/);
assert.match(tokenBreakdownSource, /AnimatePresence initial=\{false\}/);
assert.match(tokenBreakdownSource, /min-\[375px\]:grid-cols-2/);
assert.match(tokenBreakdownSource, /focus-visible:ring-2 focus-visible:ring-primary\/20/);

for (const localeName of ['zh-Hans', 'zh-Hant', 'en', 'ja']) {
    const localeSource = fs.readFileSync(path.join(repoRoot, `web/public/locale/${localeName}.json`), 'utf8');
    assert.match(localeSource, /"summaryHint":/);
    assert.match(localeSource, /"showRuntime":/);
    assert.match(localeSource, /"hideRuntime":/);
    assert.match(localeSource, /"runtimePanelTitle":/);
    assert.match(localeSource, /"runtimePanelHint":/);
    assert.match(localeSource, /"showMoreLists":/);
    assert.match(localeSource, /"hideMoreLists":/);
    assert.match(localeSource, /"topItems":/);
}

console.log('home-layout verification passed');
