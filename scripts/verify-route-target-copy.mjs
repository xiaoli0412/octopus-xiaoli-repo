import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const formPath = path.join(repoRoot, 'web/src/components/modules/channel/Form.tsx');
const zhHansPath = path.join(repoRoot, 'web/public/locale/zh-Hans.json');
const zhHantPath = path.join(repoRoot, 'web/public/locale/zh-Hant.json');

const formSource = fs.readFileSync(formPath, 'utf8');
const zhHans = JSON.parse(fs.readFileSync(zhHansPath, 'utf8'));
const zhHant = JSON.parse(fs.readFileSync(zhHantPath, 'utf8'));

const formLocale = zhHans.channel.form;
const formLocaleHant = zhHant.channel.form;

assert.match(
	formSource,
	/<Accordion type="single" collapsible className="w-full border rounded-xl bg-muted\/20">[\s\S]*?value="route-target-overrides"/,
);
assert.match(formSource, /disabled=\{!channelId\}/);
assert.doesNotMatch(formSource, /Save the channel first to manage persisted route-target overrides\./);

assert.equal(formLocale.routeTargetTitle, '\u9ad8\u7ea7\u8def\u7531\u8986\u76d6');
assert.match(formLocale.routeTargetSaveFirst, /\u8bf7\u5148\u4fdd\u5b58\u6e20\u9053/);
assert.match(formLocale.routeTargetSaveFirst, /\u5355\u72ec\u8bbe\u7f6e\u8ba1\u8d39\u3001\u63a2\u6d4b\u4e0e\u6062\u590d\u7b56\u7565/);
assert.match(formLocale.routeTargetHint, /\u9ad8\u7ea7\u80fd\u529b/);
assert.equal(formLocale.routeTargetExisting, '\u5df2\u4fdd\u5b58\u89c4\u5219');

assert.equal(formLocaleHant.routeTargetTitle, '\u9ad8\u7d1a\u8def\u7531\u8986\u5beb');
assert.match(formLocaleHant.routeTargetSaveFirst, /\u8acb\u5148\u5132\u5b58\u4f9b\u61c9\u6e90/);
assert.match(formLocaleHant.routeTargetSaveFirst, /\u55ae\u7368\u8a2d\u5b9a\u8a08\u8cbb\u3001\u63a2\u6e2c\u8207\u6062\u5fa9\u7b56\u7565/);
assert.match(formLocaleHant.routeTargetHint, /\u9ad8\u7d1a\u80fd\u529b/);
assert.equal(formLocaleHant.routeTargetExisting, '\u5df2\u5132\u5b58\u898f\u5247');

console.log('route-target-copy verification passed');
