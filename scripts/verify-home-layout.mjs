import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const homeIndexPath = path.join(repoRoot, 'web/src/components/modules/home/index.tsx');
const activityPath = path.join(repoRoot, 'web/src/components/modules/home/activity.tsx');
const chartPath = path.join(repoRoot, 'web/src/components/modules/home/chart.tsx');
const rankPath = path.join(repoRoot, 'web/src/components/modules/home/rank.tsx');
const totalPath = path.join(repoRoot, 'web/src/components/modules/home/total.tsx');
const tokenBreakdownPath = path.join(repoRoot, 'web/src/components/modules/home/token-breakdown.tsx');

const homeIndexSource = fs.readFileSync(homeIndexPath, 'utf8');
const activitySource = fs.readFileSync(activityPath, 'utf8');
const chartSource = fs.readFileSync(chartPath, 'utf8');
const rankSource = fs.readFileSync(rankPath, 'utf8');
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

assert.match(rankSource, /data-testid="home-rank-section"/);
assert.match(rankSource, /data-testid="home-rank-list"/);
assert.match(rankSource, /data-testid=\{`home-rank-card-\$\{rank\}`\}/);
assert.match(rankSource, /sm:grid-cols-2 xl:grid-cols-3/);
assert.match(rankSource, /max-h-\[23rem\] overflow-y-auto pr-1/);
assert.match(rankSource, /TabsList className="w-full justify-start bg-background\/60 sm:w-fit"/);

assert.match(activitySource, /data-testid="home-activity-section"/);
assert.match(activitySource, /data-testid="home-activity-grid"/);

assert.match(chartSource, /data-testid="home-stats-chart-section"/);
assert.match(chartSource, /data-testid="home-stats-chart"/);

assert.match(totalSource, /data-testid="home-total-section"/);
assert.match(totalSource, /data-testid=\{`home-total-summary-card-\$\{index\}`\}/);
assert.match(totalSource, /md:grid-cols-2 lg:grid-cols-4/);
assert.match(totalSource, /writing-mode:vertical-lr/);

assert.match(tokenBreakdownSource, /const \[showRuntimeDetails, setShowRuntimeDetails\] = useState\(false\);/);
assert.match(tokenBreakdownSource, /const \[showMoreLists, setShowMoreLists\] = useState\(false\);/);
assert.match(tokenBreakdownSource, /data-testid="home-breakdown-section"/);
assert.match(tokenBreakdownSource, /data-testid="home-breakdown-lists"/);
assert.match(tokenBreakdownSource, /data-testid="home-runtime-price-card"/);
assert.match(tokenBreakdownSource, /data-testid="home-runtime-circuit-card"/);
assert.match(tokenBreakdownSource, /data-testid="home-runtime-probe-card"/);
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
