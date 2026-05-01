import assert from 'node:assert/strict';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const require = createRequire(import.meta.url);
const { createJiti } = require(path.join(repoRoot, 'web/node_modules/.pnpm/jiti@2.6.1/node_modules/jiti/lib/jiti.cjs'));

const jiti = createJiti(import.meta.url, {
    moduleCache: false,
    alias: {
        '@': path.join(repoRoot, 'web/src'),
    },
});

const {
    formatConfigProfileLabel,
    resolveConfigSourceRuntime,
    resolveProfileSummaryState,
    resolveSelectedProfile,
} = jiti(path.join(repoRoot, 'web/src/components/modules/ai-automation/config-source-logic.ts'));

const readyProfile = {
    id: 11,
    domain: 'grouping',
    name: 'Routing Plan Alpha',
    version: 2,
    status: 'ready',
    confidence: 0.87,
    explanation: 'Keeps manual config intact while switching runtime reads.',
    updated_at: '2026-04-27T09:00:00Z',
};

const activeProfile = {
    ...readyProfile,
    status: 'active',
};

const manualConfig = {
    base_url: 'http://127.0.0.1:8080/v1',
    api_key: '',
    channel_type: 'openai-compatible',
    model: 'gpt-free',
    use_local_default: true,
};

const effectiveConfig = {
    base_url: 'https://profile.example/v1',
    api_key: 'profile-key',
    channel_type: 'anthropic',
    model: 'profile-model',
    use_local_default: false,
};

assert.equal(formatConfigProfileLabel(undefined), '');
assert.equal(formatConfigProfileLabel({ id: 11, name: 'Routing Plan Alpha' }), 'Routing Plan Alpha (#11)');
assert.equal(formatConfigProfileLabel({ id: 12, name: '' }), '#12');

const profileSummaryState = resolveProfileSummaryState({
    profiles: [],
    requestedActiveProfileID: 11,
    resolvedActiveProfileID: 0,
    persistedActiveProfileID: 11,
    requestedProfileSummary: readyProfile,
    activeProfileSummary: undefined,
});

assert.equal(profileSummaryState.activeProfile, undefined);
assert.deepEqual(profileSummaryState.requestedProfile, readyProfile);
assert.equal(profileSummaryState.persistedProfile, undefined);
assert.equal(profileSummaryState.resolvedProfile, undefined);

const selectedProfile = resolveSelectedProfile({
    profiles: [],
    selectedProfileID: undefined,
    requestedProfile: profileSummaryState.requestedProfile,
    resolvedProfile: profileSummaryState.resolvedProfile,
    activeProfile: profileSummaryState.activeProfile,
    persistedProfile: profileSummaryState.persistedProfile,
});

assert.deepEqual(selectedProfile, readyProfile);

const fallbackRuntime = resolveConfigSourceRuntime({
    enabled: true,
    base_url: manualConfig.base_url,
    api_key: manualConfig.api_key,
    channel_type: manualConfig.channel_type,
    model: manualConfig.model,
    use_local_default: manualConfig.use_local_default,
    default_selection_policy: 'free-success-latency',
    requested_config_source_mode: 'ai_profile',
    config_source_mode: 'manual',
    requested_active_ai_profile_id: 11,
    active_ai_profile_id: 0,
    requested_active_ai_profile: readyProfile,
    active_ai_profile: undefined,
    source_fallback_reason: 'profile_invalid_content',
    dynamic_routing_learning_enabled: true,
    manual_config: manualConfig,
    effective_config: manualConfig,
}, {
    baseURL: manualConfig.base_url,
    apiKey: manualConfig.api_key,
    channelType: manualConfig.channel_type,
    modelName: manualConfig.model,
    useLocalDefault: manualConfig.use_local_default,
});

assert.equal(fallbackRuntime.requestedConfigSourceMode, 'ai_profile');
assert.equal(fallbackRuntime.configSourceMode, 'manual');
assert.equal(fallbackRuntime.runtimeFallbackActive, true);
assert.equal(fallbackRuntime.sourceFallbackReason, 'profile_invalid_content');
assert.equal(fallbackRuntime.requestedActiveProfileID, 11);
assert.equal(fallbackRuntime.activeProfileID, 0);
assert.deepEqual(fallbackRuntime.requestedActiveProfile, readyProfile);
assert.equal(fallbackRuntime.activeProfile, undefined);
assert.equal(fallbackRuntime.effectiveBaseURL, manualConfig.base_url);
assert.equal(fallbackRuntime.effectiveChannelType, manualConfig.channel_type);
assert.equal(fallbackRuntime.effectiveModel, manualConfig.model);

const aiProfileRuntime = resolveConfigSourceRuntime({
    enabled: true,
    base_url: manualConfig.base_url,
    api_key: manualConfig.api_key,
    channel_type: manualConfig.channel_type,
    model: manualConfig.model,
    use_local_default: manualConfig.use_local_default,
    default_selection_policy: 'free-success-latency',
    requested_config_source_mode: 'ai_profile',
    config_source_mode: 'ai_profile',
    requested_active_ai_profile_id: 11,
    active_ai_profile_id: 11,
    requested_active_ai_profile: activeProfile,
    active_ai_profile: activeProfile,
    source_fallback_reason: '',
    dynamic_routing_learning_enabled: true,
    manual_config: manualConfig,
    effective_config: effectiveConfig,
}, {
    baseURL: 'http://127.0.0.1:9090/v1',
    apiKey: '',
    channelType: 'openai-compatible',
    modelName: 'manual-draft-model',
    useLocalDefault: true,
});

assert.equal(aiProfileRuntime.runtimeFallbackActive, false);
assert.equal(aiProfileRuntime.configSourceMode, 'ai_profile');
assert.equal(aiProfileRuntime.effectiveBaseURL, effectiveConfig.base_url);
assert.equal(aiProfileRuntime.effectiveAPIKey, effectiveConfig.api_key);
assert.equal(aiProfileRuntime.effectiveChannelType, effectiveConfig.channel_type);
assert.equal(aiProfileRuntime.effectiveUseLocalDefault, effectiveConfig.use_local_default);
assert.equal(aiProfileRuntime.effectiveModel, effectiveConfig.model);
assert.equal(aiProfileRuntime.manualDraftBaseURL, 'http://127.0.0.1:9090/v1');
assert.equal(aiProfileRuntime.manualDraftModel, 'manual-draft-model');

console.log('ai-config-profile-summary verification passed');
