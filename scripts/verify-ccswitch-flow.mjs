import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const docModalPath = path.join(repoRoot, 'web/src/components/modules/navbar/DocModal.tsx');
const docModalSource = fs.readFileSync(docModalPath, 'utf8');

assert.match(docModalSource, /const ccswitchProgress = useMemo\(/);
assert.match(docModalSource, /ccswitchProgressTitle/);
assert.match(docModalSource, /ccswitchProgressHint/);
assert.match(docModalSource, /ccswitchProgressFootnote/);
assert.match(docModalSource, /ccswitchStepDone/);
assert.match(docModalSource, /ccswitchStepPending/);
assert.match(docModalSource, /ccswitchProgress\.map\(\(step, index\) => \(/);
assert.match(docModalSource, /step\.done \? t\('ccswitchStepDone'\) : t\('ccswitchStepPending'\)/);
assert.match(docModalSource, /app === 'claude' \? t\('ccswitchAppClaude'\) : t\('ccswitchAppCodex'\)/);
assert.match(docModalSource, /AccordionItem value="ccswitch-advanced"/);
assert.match(docModalSource, /<HelpHint className="mt-1 size-3\.5">\{t\('ccswitchAdvancedHint'\)\}<\/HelpHint>/);
assert.match(docModalSource, /const ccswitchHasGroupOptions = groupOptions\.length > 0;/);
assert.match(docModalSource, /const ccswitchCanChooseModel = selectedApiKey !== '' && ccswitchHasGroupOptions;/);
assert.match(docModalSource, /const ccswitchCanConfirmName = ccswitchCanChooseModel && ccswitchForm\.model\.trim\(\) !== '';/);
assert.match(docModalSource, /const ccswitchCanOpenAdvanced =[\s\S]*ccswitchForm\.name\.trim\(\) !== '';/);
assert.match(docModalSource, /const ccswitchModelBlockedHint =[\s\S]*t\('ccswitchModelLockedHint'\)[\s\S]*t\('ccswitchModelUnavailableHint'\);/);
assert.match(docModalSource, /const ccswitchImportBlockedHint =[\s\S]*t\('ccswitchImportLockedHintKey'\)[\s\S]*t\('ccswitchImportLockedHintGroup'\)[\s\S]*t\('ccswitchImportLockedHintModel'\)[\s\S]*t\('ccswitchImportLockedHintName'\)[\s\S]*'';/);
assert.match(docModalSource, /className="flex items-center justify-between border-b border-border p-4 shrink-0 sm:p-6"/);
assert.match(docModalSource, /className="flex shrink-0 overflow-x-auto border-b border-border px-4 sm:px-6"/);
assert.match(docModalSource, /whitespace-nowrap px-4 py-3 text-sm font-medium border-b-2 transition-colors -mb-px/);
assert.match(docModalSource, /<Select value=\{selectedApiKey\} onValueChange=\{setSelectedApiKey\}>/);
assert.match(docModalSource, /<Select value=\{ccswitchForm\.model\} onValueChange=\{v => updateCCSwitch\(\{ model: v \}\)\} disabled=\{!ccswitchCanChooseModel\}>/);
assert.match(docModalSource, /\{!ccswitchCanChooseModel \? \([\s\S]*ccswitchModelBlockedHint[\s\S]*\) : null\}/);
assert.match(docModalSource, /\{ccswitchCanConfirmName \? \([\s\S]*ccswitchNameHint[\s\S]*\) : \([\s\S]*ccswitchNameLockedHint[\s\S]*\)\}/);
assert.match(docModalSource, /ccswitchCanOpenAdvanced \? \([\s\S]*AccordionItem value="ccswitch-advanced"[\s\S]*\) : \([\s\S]*ccswitchAdvancedLockedHint[\s\S]*\)/);
assert.match(docModalSource, /\{!ccswitchReady && ccswitchImportBlockedHint \? \([\s\S]*ccswitchImportBlockedHint[\s\S]*\) : null\}/);
assert.match(docModalSource, /'h-auto min-h-11 rounded-xl px-3 py-2 text-center text-xs whitespace-normal capitalize sm:text-sm'/);
assert.match(docModalSource, /className="h-auto min-h-11 w-full gap-2 rounded-xl px-4 py-3 text-sm whitespace-normal"/);

const zhHansLocale = fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hans.json'), 'utf8');
const zhHantLocale = fs.readFileSync(path.join(repoRoot, 'web/public/locale/zh-Hant.json'), 'utf8');
const enLocale = fs.readFileSync(path.join(repoRoot, 'web/public/locale/en.json'), 'utf8');
const jaLocale = fs.readFileSync(path.join(repoRoot, 'web/public/locale/ja.json'), 'utf8');

for (const localeSource of [zhHansLocale, zhHantLocale, enLocale, jaLocale]) {
    assert.match(localeSource, /"ccswitchProgressTitle":/);
    assert.match(localeSource, /"ccswitchProgressHint":/);
    assert.match(localeSource, /"ccswitchProgressFootnote":/);
    assert.match(localeSource, /"ccswitchStepDone":/);
    assert.match(localeSource, /"ccswitchStepPending":/);
    assert.match(localeSource, /"ccswitchAppClaude":/);
    assert.match(localeSource, /"ccswitchAppCodex":/);
    assert.match(localeSource, /"ccswitchModelLockedHint":/);
    assert.match(localeSource, /"ccswitchModelUnavailableHint":/);
    assert.match(localeSource, /"ccswitchNameLockedHint":/);
    assert.match(localeSource, /"ccswitchAdvancedLockedHint":/);
    assert.match(localeSource, /"ccswitchImportLockedHintKey":/);
    assert.match(localeSource, /"ccswitchImportLockedHintGroup":/);
    assert.match(localeSource, /"ccswitchImportLockedHintModel":/);
    assert.match(localeSource, /"ccswitchImportLockedHintName":/);
}

assert.doesNotMatch(zhHansLocale, /"ccswitchNamePlaceholder": "e\.g\./);
assert.doesNotMatch(zhHantLocale, /"ccswitchNamePlaceholder": "e\.g\./);
assert.doesNotMatch(zhHantLocale, /"ccswitchNamePlaceholder": "例：My Claude"/);
assert.doesNotMatch(zhHantLocale, /"ccswitchCliTool": "CLI 工具"/);

console.log('ccswitch-flow verification passed');
