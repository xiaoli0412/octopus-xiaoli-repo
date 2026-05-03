import type {
    AIAutomationConfig,
    AIAutomationConfigValues,
    AIAutomationProfileRef,
    AIProfile,
} from '@/api/endpoints/ai-automation';

type ConfigSourceMode = 'manual' | 'ai_profile';

type ProfileLike = Pick<AIAutomationProfileRef, 'id' | 'name'>;

type DraftConfigInput = {
    baseURL?: string;
    apiKey?: string;
    channelType?: string;
    modelName?: string;
    useLocalDefault?: boolean;
};

type ProfileSummaryState = {
    activeProfile?: AIProfile | AIAutomationProfileRef;
    requestedProfile?: AIProfile | AIAutomationProfileRef;
    persistedProfile?: AIProfile;
    resolvedProfile?: AIProfile | AIAutomationProfileRef;
};

type ResolveProfileStateOptions = {
    profiles: AIProfile[];
    requestedActiveProfileID: number;
    resolvedActiveProfileID: number;
    persistedActiveProfileID: number;
    requestedProfileSummary?: AIAutomationProfileRef;
    activeProfileSummary?: AIAutomationProfileRef;
};

export function formatConfigProfileLabel(profile?: ProfileLike) {
    if (!profile) return '';
    return profile.name ? `${profile.name} (#${profile.id})` : `#${profile.id}`;
}

export function resolveProfileSummaryState(options: ResolveProfileStateOptions): ProfileSummaryState {
    const {
        profiles,
        requestedActiveProfileID,
        resolvedActiveProfileID,
        persistedActiveProfileID,
        requestedProfileSummary,
        activeProfileSummary,
    } = options;

    const activeProfile = profiles.find((profile) => profile.id === resolvedActiveProfileID);
    const requestedProfile = profiles.find((profile) => profile.id === requestedActiveProfileID) ?? requestedProfileSummary;
    const persistedProfile = profiles.find((profile) => profile.id === persistedActiveProfileID);
    const resolvedProfile = profiles.find((profile) => profile.id === resolvedActiveProfileID) ?? activeProfileSummary;

    return {
        activeProfile,
        requestedProfile,
        persistedProfile,
        resolvedProfile,
    };
}

export function resolveSelectedProfile(options: ProfileSummaryState & { profiles: AIProfile[]; selectedProfileID?: number }) {
    const {
        profiles,
        selectedProfileID,
        requestedProfile,
        resolvedProfile,
        activeProfile,
        persistedProfile,
    } = options;

    return profiles.find((profile) => profile.id === selectedProfileID)
        ?? requestedProfile
        ?? resolvedProfile
        ?? activeProfile
        ?? persistedProfile
        ?? profiles[0];
}

function fallbackManualConfig(config?: AIAutomationConfig): AIAutomationConfigValues {
    return config?.manual_config ?? {
        base_url: config?.base_url || 'http://127.0.0.1:1088/v1',
        api_key: config?.api_key || '',
        channel_type: config?.channel_type || 'openai-compatible',
        model: config?.model || '',
        use_local_default: config?.use_local_default ?? true,
    };
}

function fallbackEffectiveConfig(config: AIAutomationConfig | undefined, manualConfig: AIAutomationConfigValues): AIAutomationConfigValues {
    return config?.effective_config ?? {
        base_url: config?.base_url || manualConfig.base_url || 'http://127.0.0.1:1088/v1',
        api_key: config?.api_key || manualConfig.api_key || '',
        channel_type: config?.channel_type || manualConfig.channel_type || 'openai-compatible',
        model: config?.model || manualConfig.model || '',
        use_local_default: config?.use_local_default ?? manualConfig.use_local_default,
    };
}

export function resolveConfigSourceRuntime(config?: AIAutomationConfig, draft: DraftConfigInput = {}) {
    const requestedConfigSourceMode: ConfigSourceMode = config?.requested_config_source_mode ?? config?.config_source_mode ?? 'manual';
    const configSourceMode: ConfigSourceMode = config?.config_source_mode ?? 'manual';
    const requestedActiveProfileID = config?.requested_active_ai_profile_id ?? config?.active_ai_profile_id ?? 0;
    const activeProfileID = config?.active_ai_profile_id ?? 0;
    const requestedActiveProfile = config?.requested_active_ai_profile;
    const activeProfile = config?.active_ai_profile;
    const sourceFallbackReason = config?.source_fallback_reason ?? '';
    const runtimeFallbackActive = requestedConfigSourceMode === 'ai_profile' && configSourceMode === 'manual';

    const manualConfig = fallbackManualConfig(config);
    const runtimeConfig = fallbackEffectiveConfig(config, manualConfig);

    const manualDraftBaseURL = draft.baseURL || manualConfig.base_url || 'http://127.0.0.1:1088/v1';
    const manualDraftAPIKey = draft.apiKey || manualConfig.api_key || '';
    const manualDraftChannelType = draft.channelType || manualConfig.channel_type || 'openai-compatible';
    const manualDraftUseLocalDefault = draft.useLocalDefault ?? manualConfig.use_local_default;
    const manualDraftModel = draft.modelName || manualConfig.model || '';

    const profileEffectiveBaseURL = runtimeConfig.base_url || manualDraftBaseURL || 'http://127.0.0.1:1088/v1';
    const profileEffectiveAPIKey = runtimeConfig.api_key || manualDraftAPIKey || '';
    const profileEffectiveChannelType = runtimeConfig.channel_type || manualDraftChannelType || 'openai-compatible';
    const profileEffectiveUseLocalDefault = runtimeConfig.use_local_default ?? manualDraftUseLocalDefault;
    const profileEffectiveModel = runtimeConfig.model || manualDraftModel || '';

    const effectiveBaseURL = configSourceMode === 'ai_profile' ? profileEffectiveBaseURL : manualDraftBaseURL;
    const effectiveAPIKey = configSourceMode === 'ai_profile' ? profileEffectiveAPIKey : manualDraftAPIKey;
    const effectiveChannelType = configSourceMode === 'ai_profile' ? profileEffectiveChannelType : manualDraftChannelType;
    const effectiveUseLocalDefault = configSourceMode === 'ai_profile' ? profileEffectiveUseLocalDefault : manualDraftUseLocalDefault;
    const effectiveModel = configSourceMode === 'ai_profile' ? profileEffectiveModel : manualDraftModel;

    return {
        requestedConfigSourceMode,
        configSourceMode,
        requestedActiveProfileID,
        activeProfileID,
        requestedActiveProfile,
        activeProfile,
        sourceFallbackReason,
        runtimeFallbackActive,
        manualConfig,
        runtimeConfig,
        manualDraftBaseURL,
        manualDraftAPIKey,
        manualDraftChannelType,
        manualDraftUseLocalDefault,
        manualDraftModel,
        profileEffectiveBaseURL,
        profileEffectiveAPIKey,
        profileEffectiveChannelType,
        profileEffectiveUseLocalDefault,
        profileEffectiveModel,
        effectiveBaseURL,
        effectiveAPIKey,
        effectiveChannelType,
        effectiveUseLocalDefault,
        effectiveModel,
    };
}
