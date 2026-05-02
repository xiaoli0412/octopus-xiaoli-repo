import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const componentPath = path.join(repoRoot, 'web/src/components/modules/log/index.tsx');
const helperPath = path.join(repoRoot, 'web/src/components/modules/log/range-logic.ts');

const componentSource = fs.readFileSync(componentPath, 'utf8');
const helperSource = fs.readFileSync(helperPath, 'utf8');
const enLocale = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/en.json'), 'utf8'));
const zhHansLocale = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hans.json'), 'utf8'));
const zhHantLocale = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hant.json'), 'utf8'));
const jaLocale = JSON.parse(fs.readFileSync(path.join(repoRoot, 'web/public/locale/ja.json'), 'utf8'));

assert.match(helperSource, /export function hasPartialTimeRange\(startTime: string, endTime: string\)/);
assert.match(helperSource, /return Boolean\(\(startTime && !endTime\) \|\| \(!startTime && endTime\)\);/);

assert.match(componentSource, /import \{ hasPartialTimeRange \} from '\.\/range-logic';/);
assert.match(componentSource, /const partialRange = hasPartialTimeRange\(startTime, endTime\);/);
assert.match(componentSource, /if \(partialRange\) \{/);
assert.match(componentSource, /toast\.error\(t\('list\.partialRange'\)\);/);
assert.match(componentSource, /disabled=\{isExporting \|\| invalidRange \|\| partialRange\}/);
assert.match(componentSource, /\{partialRange && \(/);
assert.match(componentSource, /!partialRange && invalidRange/);

for (const locale of [enLocale, zhHansLocale, zhHantLocale, jaLocale]) {
	assert.ok(locale.log.list.partialRange, 'missing log.list.partialRange locale key');
}

assert.equal(enLocale.log.list.partialRange, 'Start and end time must be set together');
assert.equal(zhHansLocale.log.list.partialRange, '开始时间和结束时间必须同时填写');
assert.equal(zhHantLocale.log.list.partialRange, '開始時間和結束時間必須同時填寫');
assert.equal(jaLocale.log.list.partialRange, '開始時刻と終了時刻は同時に入力してください');

console.log('log-export-range verification passed');
