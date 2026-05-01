import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const source = fs.readFileSync(path.join(repoRoot, 'web/src/components/common/HelpHint.tsx'), 'utf8');
const browserSmokeSource = fs.readFileSync(path.join(repoRoot, 'scripts/verify-setting-help-browser-smoke.mjs'), 'utf8');
const browserSmokeCdpSource = fs.readFileSync(path.join(repoRoot, 'scripts/verify-setting-help-browser-smoke-cdp.mjs'), 'utf8');

assert.match(source, /type HelpHintProps = \{/);
assert.match(source, /ariaLabel\?: string;/);
assert.match(source, /import \{ useTranslations \} from 'next-intl';/);
assert.match(source, /const t = useTranslations\('common\.helpHint'\);/);
assert.match(source, /const resolvedAriaLabel = ariaLabel \?\? t\('ariaLabel'\);/);
assert.match(source, /aria-label=\{resolvedAriaLabel\}/);
assert.match(source, /data-slot="help-hint-trigger"/);
assert.match(source, /data-help-hint-trigger="true"/);
assert.match(source, /const hintId = useId\(\);/);
assert.match(source, /data-help-hint-id=\{hintId\}/);
assert.match(source, /data-slot="help-hint-content"/);
assert.match(source, /id=\{hintId\}/);
assert.match(source, /<button/);
assert.match(source, /focus-visible:ring-2/);
assert.match(source, /focus-visible:ring-offset-2/);
assert.match(source, /<HelpCircle aria-hidden="true" className="size-full" \/>/);
assert.doesNotMatch(source, /<span[^>]*role="button"/);
assert.match(browserSmokeSource, /const helpButtonSelector = process\.env\.OCTOPUS_UI_SMOKE_HELP_SELECTOR \|\| 'button\[data-help-hint-trigger="true"\]';/);
assert.match(browserSmokeCdpSource, /const helpButtonSelector = process\.env\.OCTOPUS_UI_SMOKE_HELP_SELECTOR \|\| 'button\[data-help-hint-trigger="true"\]';/);

console.log('help-hint-accessible verification passed');
