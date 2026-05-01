import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const editorPath = path.join(repoRoot, 'web/src/components/modules/group/Editor.tsx');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const editorComponentStart = editorSource.indexOf('export function GroupEditor(');
assert.notEqual(editorComponentStart, -1, 'GroupEditor component should exist');
const renderStart = editorSource.indexOf('return (', editorComponentStart);
assert.notEqual(renderStart, -1, 'render block should exist');
const renderSource = editorSource.slice(renderStart);
const advancedStrategySectionMatch = renderSource.match(/<AccordionItem[^>]*data-testid=\{`\$\{idPrefix\}-advanced-strategy-item`\}[^>]*value="advanced-strategy"[^>]*className="border-none">[\s\S]*?<\/AccordionItem>/);
assert.ok(advancedStrategySectionMatch, 'advanced strategy section should exist');
const advancedStrategySection = advancedStrategySectionMatch[0];
const cardSource = fs.readFileSync(path.join(repoRoot, 'web/src/components/modules/group/Card.tsx'), 'utf8');

assert.match(editorSource, /import \{ useTranslations \} from 'next-intl';/);
assert.doesNotMatch(editorSource, /useLocale/);
assert.match(editorSource, /const copy = useMemo<GroupEditorCopy>\(\(\) => \(\{/);
assert.match(editorSource, /flowTitle: t\('form\.flowTitle'\)/);
assert.match(editorSource, /flowHint: t\('form\.flowHint'\)/);
assert.match(editorSource, /flowDesc: t\('form\.flowDesc'\)/);
assert.match(editorSource, /stepLabel:\s*\(index\)\s*=>\s*t\('form\.stepLabel',\s*\{\s*index\s*\}\s*\)/);
assert.match(editorSource, /flowSteps: \{/);
assert.match(editorSource, /naming: t\('form\.flowSteps\.naming'\)/);
assert.match(editorSource, /mode: t\('form\.flowSteps\.mode'\)/);
assert.match(editorSource, /models: t\('form\.flowSteps\.models'\)/);
assert.match(editorSource, /namingSectionTitle: t\('form\.namingSectionTitle'\)/);
assert.match(editorSource, /modeSectionTitle: t\('form\.modeSectionTitle'\)/);
assert.match(editorSource, /advancedStrategyHint: t\('form\.advancedStrategyHint'\)/);
assert.match(editorSource, /modelPickerTitle: t\('form\.modelPickerTitle'\)/);
assert.match(editorSource, /itemsEmptyTitle: t\('form\.itemsEmptyTitle'\)/);
assert.match(editorSource, /retryRounds: t\('form\.retryRounds'\)/);
assert.match(editorSource, /raceConcurrencyHint: t\('form\.raceConcurrencyHint'\)/);

assert.match(renderSource, /copy\.flowTitle/);
assert.match(renderSource, /copy\.flowHint/);
assert.match(renderSource, /copy\.flowSteps\[step as keyof GroupEditorCopy\['flowSteps'\]\]/);
assert.match(renderSource, /data-testid=\{`\$\{idPrefix\}-form`\}/);
assert.match(renderSource, /data-testid=\{`\$\{idPrefix\}-flow-card`\}/);
assert.match(renderSource, /data-testid=\{`\$\{idPrefix\}-naming-section`\}/);
assert.match(renderSource, /data-testid=\{`\$\{idPrefix\}-mode-section`\}/);
assert.match(renderSource, /data-testid=\{`\$\{idPrefix\}-advanced-strategy-section`\}/);
assert.match(renderSource, /<ModelPickerSection[\s\S]*?idPrefix=\{idPrefix\}/);
assert.match(renderSource, /<SortSection[\s\S]*?idPrefix=\{idPrefix\}/);
assert.match(renderSource, /copy\.modeSectionTitle/);
assert.match(renderSource, /copy\.nameHint/);
assert.match(renderSource, /copy\.matchRegexHint/);
assert.match(renderSource, /t\('form\.matchRegexInvalid'\)/);
assert.match(renderSource, /copy\.advancedStrategyHint/);
assert.match(editorSource, /copy\.modelPickerTitle/);
assert.match(editorSource, /copy\.noAvailableModelsTitle/);
assert.match(editorSource, /copy\.noAvailableModelsHint/);
assert.match(editorSource, /copy\.noFilteredModelsHint/);
assert.match(editorSource, /copy\.itemsEmptyTitle/);
assert.match(editorSource, /xl:grid-cols-\[minmax\(0,1fr\)_minmax\(0,1fr\)\] 2xl:grid-cols-\[minmax\(0,1\.02fr\)_minmax\(0,0\.98fr\)\]/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-selected-section`\}[\s\S]*?xl:sticky xl:top-0 xl:max-h-\[calc\(100vh-16rem\)\]/);
assert.match(editorSource, /<div className="grid max-h-56 grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-4">/);
assert.match(editorSource, /<div className="flex flex-wrap items-center gap-2 text-\[11px\] text-muted-foreground">/);

assert.match(advancedStrategySection, /<AccordionItem[^>]*data-testid=\{`\$\{idPrefix\}-advanced-strategy-item`\}[^>]*value="advanced-strategy"[^>]*className="border-none">/);
assert.match(advancedStrategySection, /data-testid="group-advanced-strategy-trigger"/);
assert.match(editorSource, /import \{ Accordion, AccordionContent, AccordionItem, AccordionTrigger \} from '@\/components\/ui\/accordion';/);
assert.match(advancedStrategySection, /<AccordionTrigger[\s\S]*?data-testid="group-advanced-strategy-trigger"/);
assert.match(advancedStrategySection, /addon=\{<HelpHint className="mt-1 size-3\.5">\{copy\.advancedStrategyHint\}<\/HelpHint>\}/);
assert.match(editorSource, /<span className="truncate text-sm font-medium text-foreground">\{copy\.modelPickerTitle\}<\/span>/);
assert.match(editorSource, /<div className="font-medium text-foreground">\{copy\.itemsEmptyTitle\}<\/div>/);
assert.match(editorSource, /<div className="grid max-h-56 grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-4">/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-auto-add`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-filter`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-picker-section`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-selected-section`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-model-empty-state`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-selected-empty-state`\}/);
assert.match(editorSource, /data-testid=\{`\$\{idPrefix\}-advanced-strategy-item`\}/);
assert.match(editorSource, /hasAnyModelChannels \? t\('form\.noFilteredModels'\) : copy\.noAvailableModelsTitle/);
assert.match(editorSource, /hasAnyModelChannels \? copy\.noFilteredModelsHint : copy\.noAvailableModelsHint/);
assert.doesNotMatch(editorSource, /<div className="mt-1 text-xs text-muted-foreground">\{t\('form\.selectionSummaryHint'\)\}<\/div>/);
assert.match(editorSource, /name: mc\.channel_name \|\| t\('channelFallbackName', \{ id: mc\.channel_id \}\)/);
assert.match(cardSource, /channel_name: channelNameByKey\.get\(modelChannelKey\(item\.channel_id, item\.model_name\)\) \?\? t\('channelFallbackName', \{ id: item\.channel_id \}\)/);
assert.match(cardSource, /idPrefix=\{`edit-group-\$\{group\.id \?\? 'unknown'\}`\}/);

const createSource = fs.readFileSync(path.join(repoRoot, 'web/src/components/modules/group/Create.tsx'), 'utf8');
assert.match(createSource, /data-testid="group-create-dialog"/);
assert.match(createSource, /idPrefix="new-group"/);

assert.doesNotMatch(editorSource, /strategySectionTitle/);
assert.doesNotMatch(editorSource, /const strategySummary = useMemo/);
assert.doesNotMatch(editorSource, /function formatDurationLabel/);
assert.doesNotMatch(editorSource, /Invalid regex/);
assert.doesNotMatch(advancedStrategySection, /<AccordionPrimitive\.Trigger[\s\S]*?data-testid="group-advanced-strategy-trigger"/);
assert.doesNotMatch(cardSource, /Channel \$\{item\.channel_id\}/);

for (const localeName of ['zh-Hans', 'zh-Hant', 'en', 'ja']) {
    const localeSource = fs.readFileSync(path.join(repoRoot, `web/public/locale/${localeName}.json`), 'utf8');
    assert.match(localeSource, /"flowTitle":/);
    assert.match(localeSource, /"flowHint":/);
    assert.match(localeSource, /"flowDesc":/);
    assert.match(localeSource, /"stepLabel":/);
    assert.match(localeSource, /"flowSteps": \{/);
    assert.match(localeSource, /"namingSectionTitle":/);
    assert.match(localeSource, /"modeSectionTitle":/);
    assert.match(localeSource, /"modelPickerTitle":/);
    assert.match(localeSource, /"itemsEmptyTitle":/);
    assert.match(localeSource, /"retryRounds":/);
    assert.match(localeSource, /"raceConcurrencyHint":/);
    assert.match(localeSource, /"matchRegexInvalidHint":/);
    assert.match(localeSource, /"channelFallbackName":/);
}

console.log('group-create-flow verification passed');
