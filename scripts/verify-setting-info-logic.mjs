import assert from 'node:assert/strict';
import fs from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const modulePath = path.join(repoRoot, 'web/src/components/modules/setting/info-logic.ts');

const moduleSource = fs.readFileSync(modulePath, 'utf8');
const transformedSource = stripTypeScriptTypes(moduleSource, {
	mode: 'transform',
	sourceUrl: modulePath,
});
const inlineModuleURL = `data:text/javascript;base64,${Buffer.from(transformedSource, 'utf8').toString('base64')}`;

const {
	formatVersionDisplay,
	formatReleaseTagDisplay,
	getCacheMismatchPresentation,
	isDevelopmentVersion,
	isSameReleaseVersion,
	isUnknownVersion,
} = await import(inlineModuleURL);

assert.equal(isUnknownVersion('unknown'), true);
assert.equal(isUnknownVersion(' UNKNOWN '), true);
assert.equal(isUnknownVersion('dev'), false);
assert.equal(isDevelopmentVersion('dev'), true);
assert.equal(isDevelopmentVersion(' DEVELOPMENT '), true);

assert.equal(formatVersionDisplay('', 'VERSION_UNAVAILABLE'), 'VERSION_UNAVAILABLE');
assert.equal(formatVersionDisplay('unknown', 'FRONTEND_VERSION_UNKNOWN'), 'FRONTEND_VERSION_UNKNOWN');
assert.equal(formatVersionDisplay('dev', 'VERSION_UNAVAILABLE', 'DEVELOPMENT_VERSION'), 'DEVELOPMENT_VERSION');
assert.equal(formatVersionDisplay('v1.2.3', 'VERSION_UNAVAILABLE'), 'v1.2.3');
assert.equal(formatReleaseTagDisplay('v1.12-beta'), '1.12(beta)');
assert.equal(formatReleaseTagDisplay('v1.12.0'), '1.12.0');
assert.equal(isSameReleaseVersion('1.12(beta)', 'v1.12-beta'), true);
assert.equal(isSameReleaseVersion('v1.12.0', '1.12.0'), true);
assert.equal(isSameReleaseVersion('1.12(beta)', '1.12'), false);

assert.deepEqual(
	getCacheMismatchPresentation('unknown', 'dev', {
		versionUnavailable: 'VERSION_UNAVAILABLE',
		frontendVersionUnknown: 'FRONTEND_VERSION_UNKNOWN',
		frontendVersionUnknownHint: 'FORCE_REFRESH_FIRST',
		developmentVersion: 'DEVELOPMENT_VERSION',
	}),
	{
		frontendLabel: 'FRONTEND_VERSION_UNKNOWN',
		backendLabel: 'DEVELOPMENT_VERSION',
		note: 'FORCE_REFRESH_FIRST',
	},
);

assert.deepEqual(
	getCacheMismatchPresentation('v1.0.0', '', {
		versionUnavailable: 'Unavailable',
		frontendVersionUnknown: 'Frontend version not detected',
		frontendVersionUnknownHint: 'Force refresh first.',
		developmentVersion: 'Development build',
	}),
	{
		frontendLabel: 'v1.0.0',
		backendLabel: 'Unavailable',
		note: '',
	},
);

console.log('setting-info-logic verification passed');
