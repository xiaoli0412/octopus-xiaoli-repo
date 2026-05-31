import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const editorPath = path.join(repoRoot, 'web/src/components/modules/group/Editor.tsx');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const cardSource = fs.readFileSync(path.join(repoRoot, 'web/src/components/modules/group/Card.tsx'), 'utf8');
const createSource = fs.readFileSync(path.join(repoRoot, 'web/src/components/modules/group/Create.tsx'), 'utf8');

assert.match(editorSource, /import \{ useTranslations \} from 'next-intl';/);
assert.doesNotMatch(editorSource, /useLocale/);
assert.match(editorSource, /export function GroupEditor\(/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-form`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-advanced-strategy-section`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-advanced-strategy-item`\}/);
assert.match(editorSource, /data-testid="group-advanced-strategy-trigger"/);

assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-picker-section`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-filter`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-auto-add`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-empty-state`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-selected-section`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-selected-empty-state`\}/);

assert.match(editorSource, /const filteredModelChannels = useMemo\(\(\) => \{/);
assert.match(editorSource, /const channels = useMemo\(\(\) => \{/);
assert.match(editorSource, /const emptyTitle = hasAnyModelChannels \? t\('form\.noFilteredModels'\) : t\('form\.noAvailableModelsTitle'\);/);
assert.match(editorSource, /<div className="grid max-h-\[24rem\] grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-3">/);
assert.match(editorSource, /<div className="max-h-56 space-y-2 overflow-y-auto pr-1">/);
assert.match(editorSource, /<ModelPickerSection[\s\S]*?modelChannels=\{modelChannels\}/);
assert.match(editorSource, /<SortSection[\s\S]*?idPrefix=\{idPrefix\}/);
assert.match(editorSource, /<div className="grid h-full min-h-0 grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-\[minmax\(0,1\.05fr\)_minmax\(0,0\.95fr\)\]">/);
assert.match(editorSource, /<div className="mt-3 flex flex-wrap items-center gap-2 text-\[11px\] text-muted-foreground">/);

assert.match(editorSource, /name: mc\.channel_name \|\| t\('channelFallbackName', \{ id: mc\.channel_id \}\)/);
assert.match(cardSource, /channel_name: channelNameByKey\.get\(modelChannelKey\(item\.channel_id, item\.model_name\)\) \?\? t\('channelFallbackName', \{ id: item\.channel_id \}\)/);
assert.match(cardSource, /idPrefix=\{`edit-group-\$\{group\.id \?\? 'unknown'\}`\}/);
assert.match(createSource, /data-testid="group-create-dialog"/);
assert.match(createSource, /idPrefix="new-group"/);

assert.doesNotMatch(editorSource, /GroupEditorCopy/);
assert.doesNotMatch(editorSource, /const copy = useMemo/);
assert.doesNotMatch(editorSource, /flowTitle/);
assert.doesNotMatch(editorSource, /flowHint/);
assert.doesNotMatch(editorSource, /flowDesc/);
assert.doesNotMatch(editorSource, /stepLabel/);
assert.doesNotMatch(editorSource, /flowSteps/);
assert.doesNotMatch(editorSource, /HelpHint/);
assert.doesNotMatch(editorSource, /advancedStrategyHint/);
assert.doesNotMatch(editorSource, /noAvailableModelsHint/);
assert.doesNotMatch(editorSource, /noFilteredModelsHint/);
assert.doesNotMatch(editorSource, /itemsEmptyHint/);
assert.doesNotMatch(editorSource, /xl:sticky xl:top-0/);
assert.doesNotMatch(editorSource, /<div className="grid max-h-56 grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-4">/);
assert.doesNotMatch(cardSource, /Channel \$\{item\.channel_id\}/);

for (const localeName of ['zh-Hans', 'zh-Hant', 'en', 'ja']) {
    const localeSource = fs.readFileSync(path.join(repoRoot, `web/public/locale/${localeName}.json`), 'utf8');
    assert.match(localeSource, /"modelPickerTitle":/);
    assert.match(localeSource, /"itemsEmptyTitle":/);
    assert.match(localeSource, /"retryRounds":/);
    assert.match(localeSource, /"matchRegexInvalidHint":/);
    assert.match(localeSource, /"channelFallbackName":/);
}

console.log('group-create-flow verification passed');
