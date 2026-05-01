import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const componentPath = path.join(repoRoot, 'web/src/components/modules/setting/CircuitBreaker.tsx');
const source = fs.readFileSync(componentPath, 'utf8');

assert.match(source, /data-testid="setting-circuit-breaker-card"/);
assert.match(source, /useStatsTokenBreakdown\(/);
assert.match(source, /circuitBreaker\.defaultPathTitle/);
assert.match(source, /circuitBreaker\.defaultPathDesc/);
assert.match(source, /circuitBreaker\.recommendation\.currentValues/);
assert.match(source, /data-testid="setting-circuit-breaker-recovery-trigger"/);
assert.match(source, /AccordionItem value="circuit-breaker-recovery"/);
assert.match(source, /recoverySteps\.map/);
assert.match(source, /circuitBreaker\.recoveryStep1Title/);
assert.match(source, /circuitBreaker\.recoveryStep2Title/);
assert.match(source, /circuitBreaker\.recoveryStep3Title/);
assert.match(source, /AccordionItem value="circuit-breaker-advanced"/);
assert.match(source, /data-testid="setting-circuit-breaker-advanced-trigger"/);
assert.match(source, /circuitBreaker\.advancedTitle/);
assert.match(source, /circuitBreaker\.advancedDesc/);
assert.match(source, /addon=\{<HelpHint className="mt-1 size-3\.5">\{t\('circuitBreaker\.advancedHint'\)\}<\/HelpHint>\}/);
assert.match(source, /aria-label=\{field\.label\}/);
assert.match(source, /tokenBreakdown\?\.circuit_half_open_count/);
assert.match(source, /formatSeconds\(tokenBreakdown\?\.circuit_max_remaining_cooldown_sec \?\? 0, t\)/);
assert.match(source, /<HelpHint className="size-3\.5">\{card\.helper\}<\/HelpHint>/);

const zhHans = fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hans.json'), 'utf8');
const zhHant = fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hant.json'), 'utf8');
const en = fs.readFileSync(path.join(repoRoot, 'web/public/locale/en.json'), 'utf8');
const ja = fs.readFileSync(path.join(repoRoot, 'web/public/locale/ja.json'), 'utf8');

for (const locale of [zhHans, zhHant, en, ja]) {
    assert.match(locale, /"defaultPathTitle":/);
    assert.match(locale, /"defaultPathDesc":/);
    assert.match(locale, /"advancedTitle":/);
    assert.match(locale, /"advancedDesc":/);
    assert.match(locale, /"recoveryStep3Desc":/);
    assert.match(locale, /"trackedLabel":/);
    assert.match(locale, /"cooldownValue":/);
    assert.match(locale, /"currentValues":/);
}

assert.doesNotMatch(zhHans, /"advancedTitle": "Advanced/);
assert.doesNotMatch(zhHant, /"advancedTitle": "Advanced/);

console.log('circuit-breaker-help verification passed');
