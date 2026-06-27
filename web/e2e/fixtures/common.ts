const ChannelType = {
    OpenAIChat: 0,
    OpenAIResponse: 1,
    Anthropic: 2,
    Gemini: 3,
    Volcengine: 4,
    OpenAIEmbedding: 5,
    GithubCopilot: 6,
    Antigravity: 7,
    Zen: 8,
} as const;

export const mockUserStatus = { must_change_password: false };

export const mockSettingsList = [
    { key: 'proxy_url', value: '' },
    { key: 'api_base_url', value: 'http://127.0.0.1:1088' },
    { key: 'api_alternate_base_urls', value: '' },
    { key: 'trusted_proxy_cidrs', value: '' },
    { key: 'ops_ip_display_mode', value: 'masked' },
    { key: 'stats_save_interval', value: '60' },
    { key: 'cors_allow_origins', value: '*' },
];

export const mockPublicAccess = {
    primary_base_url: 'http://127.0.0.1:1088',
    alternate_base_urls: [],
    current_base_url: 'http://127.0.0.1:1088',
    trusted_proxy_cidrs: [],
    ops_ip_display_mode: 'masked',
    current_client_ip: '127.0.0.1',
    current_client_label: 'local',
};

export const mockChannelList = [
    {
        id: 1,
        name: 'OpenAI 官方',
        type: ChannelType.OpenAIChat,
        enabled: true,
        key_management_mode: 'pooled',
        key_routing_policy: 'round_robin',
        base_urls: [{ url: 'https://api.openai.com', delay: 120 }],
        custom_header: [],
        keys: [
            {
                id: 1,
                channel_id: 1,
                enabled: true,
                channel_key: 'sk-octopus-mock-key-1',
                source_type: 'paid/metered',
                status_code: 200,
                last_use_time_stamp: 1750996800,
                total_cost: 12.34,
                remark: '主 Key',
                allowed_models: 'gpt-4o,gpt-4o-mini',
                request_capabilities: 'chat',
            },
        ],
        model: 'gpt-4o,gpt-4o-mini',
        custom_model: '',
        proxy: false,
        auto_sync: false,
        auto_group: 0,
        param_override: '',
        channel_proxy: '',
        match_regex: '',
        upstream_site_id: 0,
        upstream_source: '',
        stats: {
            channel_id: 1,
            input_token: 1200000,
            output_token: 340000,
            input_cost: 3.45,
            output_cost: 5.12,
            wait_time: 245000,
            request_success: 1234,
            request_failed: 12,
        },
    },
    {
        id: 2,
        name: 'Anthropic 官方',
        type: ChannelType.Anthropic,
        enabled: true,
        key_management_mode: 'classified',
        key_routing_policy: 'fill_priority',
        base_urls: [{ url: 'https://api.anthropic.com', delay: 180 }],
        custom_header: [],
        keys: [
            {
                id: 2,
                channel_id: 2,
                enabled: true,
                channel_key: 'sk-ant-mock-key-2',
                source_type: 'paid/metered',
                status_code: 200,
                last_use_time_stamp: 1750996500,
                total_cost: 8.76,
                remark: 'Claude Key',
                allowed_models: 'claude-3-5-sonnet-20241022',
                request_capabilities: 'chat',
            },
        ],
        model: 'claude-3-5-sonnet-20241022',
        custom_model: '',
        proxy: false,
        auto_sync: false,
        auto_group: 0,
        param_override: '',
        channel_proxy: '',
        match_regex: '',
        upstream_site_id: 0,
        upstream_source: '',
        stats: {
            channel_id: 2,
            input_token: 560000,
            output_token: 180000,
            input_cost: 1.68,
            output_cost: 2.72,
            wait_time: 132000,
            request_success: 567,
            request_failed: 3,
        },
    },
];

export const mockStatsTotal = {
    id: 1,
    input_token: 1760000,
    output_token: 520000,
    input_cost: 5.13,
    output_cost: 7.84,
    wait_time: 377000,
    request_success: 1801,
    request_failed: 15,
};

export const mockStatsDaily = Array.from({ length: 30 }, (_, i) => {
    const date = new Date('2026-06-27T00:00:00Z');
    date.setDate(date.getDate() - (29 - i));
    const dateStr = date.toISOString().slice(0, 10).replace(/-/g, '');
    return {
        date: dateStr,
        input_token: 50000 + i * 1000,
        output_token: 15000 + i * 300,
        input_cost: 0.15 + i * 0.005,
        output_cost: 0.25 + i * 0.008,
        wait_time: 12000 + i * 100,
        request_success: 50 + i,
        request_failed: i % 5 === 0 ? 1 : 0,
    };
});

export const mockStatsHourly = Array.from({ length: 24 }, (_, i) => ({
    hour: i,
    date: '20260627',
    input_token: 2000 + i * 100,
    output_token: 500 + i * 30,
    input_cost: 0.01 + i * 0.0005,
    output_cost: 0.02 + i * 0.0008,
    wait_time: 500 + i * 20,
    request_success: 5 + i,
    request_failed: i % 7 === 0 ? 1 : 0,
}));

export const mockStatsTokenBreakdown = {
    total_input_token: 1760000,
    total_output_token: 520000,
    total_token: 2280000,
    estimated_official_input_cost: 5.13,
    estimated_official_output_cost: 7.84,
    estimated_official_total_cost: 12.97,
    estimated_gateway_input_cost: 4.85,
    estimated_gateway_output_cost: 7.21,
    estimated_gateway_total_cost: 12.06,
    estimated_price_basis: 'gateway',
    estimated_probe_input_cost: 0,
    estimated_probe_output_cost: 0,
    estimated_probe_total_cost: 0,
    recent_probe_count: 0,
    recent_probe_success_count: 0,
    recent_probe_failed_count: 0,
    recent_probe_last_at: 0,
    recent_probe_last_status: '',
    recent_probe_last_channel: '',
    recent_probe_last_model: '',
    recent_probe_last_message: '',
    probe_summary_basis: '',
    circuit_tracked_count: 0,
    circuit_open_count: 0,
    circuit_half_open_count: 0,
    circuit_closed_count: 0,
    circuit_max_remaining_cooldown_sec: 0,
    circuit_summary_basis: '',
    by_channel: [
        { key: 'channel:1', label: 'OpenAI 官方', input_token: 1200000, output_token: 340000, total_token: 1540000 },
        { key: 'channel:2', label: 'Anthropic 官方', input_token: 560000, output_token: 180000, total_token: 740000 },
    ],
    by_model: [
        { key: 'model:gpt-4o', label: 'gpt-4o', input_token: 900000, output_token: 200000, total_token: 1100000 },
        { key: 'model:claude-3-5-sonnet', label: 'claude-3-5-sonnet', input_token: 560000, output_token: 180000, total_token: 740000 },
        { key: 'model:gpt-4o-mini', label: 'gpt-4o-mini', input_token: 300000, output_token: 140000, total_token: 440000 },
    ],
    by_api_key: [],
    by_channel_key: [],
};

export const mockUpstreamSiteList = [
    {
        id: 1,
        name: 'NewAPI 演示站',
        provider_type: 'newapi',
        base_url: 'https://newapi.example.com',
        api_base_url: 'https://newapi.example.com/v1',
        auth_mode: 'token',
        enabled: true,
        auto_refresh: true,
        refresh_interval_secs: 43200,
        auto_checkin: false,
        checkin_interval_secs: 86400,
        last_checkin_at: '',
        checkin_log: '',
        sync_to_channel: true,
        linked_channel_id: 1,
        last_refresh_at: '2026-06-27T10:00:00Z',
        last_refresh_status: 'success',
        last_refresh_message: '',
        model_count: 42,
        key_count: 3,
        group_count: 5,
        price_count: 18,
        subscription_count: 0,
        balance_available: true,
        balance_used: 12.34,
        balance_remain: 87.66,
        balance_unlimited: false,
        balance_alert_threshold: 10,
        last_balance_check_at: '2026-06-27T10:00:00Z',
        last_balance_value: 87.66,
        auto_create_key: false,
        key_quota_limit: 0,
        key_expire_days: 0,
        auto_sync_group: true,
        auto_sync_price: true,
        source_label: 'newapi-demo',
    },
];

export const mockUpstreamSiteDetail = {
    site: mockUpstreamSiteList[0],
    credentials: [
        { id: 1, upstream_site_id: 1, credential_type: 'token', auth_mode: 'token', display_name: '管理令牌', masked_value: 'sk-****-demo', user_id: '', importable: true, last_validated_at: '2026-06-27T10:00:00Z' },
    ],
    keys: [
        { id: 1, upstream_site_id: 1, name: '主 Key', masked_key: 'sk-****-key1', allowed_models: 'gpt-4o,gpt-4o-mini', request_capabilities: 'chat', groups: 'default', status: 'active', quota: 100000, quota_used: 1234, expires_at: '', importable: true, source_type: 'upstream', channel_key_id: 1 },
        { id: 2, upstream_site_id: 1, name: '副 Key', masked_key: 'sk-****-key2', allowed_models: '*', request_capabilities: 'chat', groups: 'default', status: 'active', quota: 0, quota_used: 0, expires_at: '', importable: true, source_type: 'upstream', channel_key_id: 2 },
    ],
    groups: [
        { id: 1, upstream_site_id: 1, external_id: 'default', name: '默认分组', description: '默认', platform: 'openai', status: 'active', rate_multiplier: 1, models: 'gpt-4o,gpt-4o-mini', request_capabilities: 'chat', source: 'upstream' },
        { id: 2, upstream_site_id: 1, external_id: 'vip', name: 'VIP 分组', description: 'VIP', platform: 'openai', status: 'active', rate_multiplier: 1.2, models: 'claude-3-5-sonnet', request_capabilities: 'chat', source: 'upstream' },
    ],
    prices: [
        { id: 1, upstream_site_id: 1, channel_id: 1, model_name: 'gpt-4o', canonical_name: 'gpt-4o', input: 2.5, output: 10, cache_read: 1.25, cache_write: 0, official_input: 5, official_output: 15, official_cache_read: 1.25, official_cache_write: 0, price_source: 'newapi', price_matched_key: 'gpt-4o', source_label: 'newapi-demo', cache_policy: 'supported', cache_reason: '', cache_supported: true, updated_at: '2026-06-27T10:00:00Z' },
        { id: 2, upstream_site_id: 1, channel_id: 1, model_name: 'gpt-4o-mini', canonical_name: 'gpt-4o-mini', input: 0.15, output: 0.6, cache_read: 0.075, cache_write: 0, official_input: 0.3, official_output: 1.2, official_cache_read: 0.075, official_cache_write: 0, price_source: 'newapi', price_matched_key: 'gpt-4o-mini', source_label: 'newapi-demo', cache_policy: 'supported', cache_reason: '', cache_supported: true, updated_at: '2026-06-27T10:00:00Z' },
    ],
    subscriptions: [],
    linked_channel: { id: 1, name: 'OpenAI 官方', type: 0, enabled: true, key_count: 1 },
};

export const mockAPIKeyList = [
    {
        id: 1,
        name: '开发测试 Key',
        api_key: 'sk-octopus-dev-********',
        enabled: true,
        expire_at: 0,
        max_cost: 100,
        supported_models: 'gpt-4o,gpt-4o-mini',
        rate_limit_rpm: 120,
        rate_limit_tpm: 100000,
        rate_limit_daily: 10000,
    },
    {
        id: 2,
        name: '只读 Key',
        api_key: 'sk-octopus-read-********',
        enabled: false,
        expire_at: 0,
        max_cost: 0,
        supported_models: '',
        rate_limit_rpm: 0,
        rate_limit_tpm: 0,
        rate_limit_daily: 0,
    },
];

export const mockCapabilityInventory = {
    serviceable_models: [
        { name: 'gpt-4o', enabled: true, channel_id: 1, channel_name: 'OpenAI 官方', key_count: 1, request_capabilities: ['chat'], inventory_source: 'channel' },
        { name: 'gpt-4o-mini', enabled: true, channel_id: 1, channel_name: 'OpenAI 官方', key_count: 1, request_capabilities: ['chat'], inventory_source: 'channel' },
    ],
    selectable_models: [
        { name: 'gpt-4o', channel_count: 1, enabled_channel_count: 1, key_count: 1, request_capabilities: ['chat'], inventory_source: 'channel' },
        { name: 'gpt-4o-mini', channel_count: 1, enabled_channel_count: 1, key_count: 1, request_capabilities: ['chat'], inventory_source: 'channel' },
        { name: 'claude-3-5-sonnet-20241022', channel_count: 1, enabled_channel_count: 1, key_count: 1, request_capabilities: ['chat'], inventory_source: 'channel' },
    ],
    routable_models: [
        { name: 'gpt-4o', group_id: 1, group_name: '默认分组', channel_count: 1, enabled_channel_count: 1, key_count: 1, request_capabilities: ['chat'], inventory_source: 'group' },
        { name: 'gpt-4o-mini', group_id: 1, group_name: '默认分组', channel_count: 1, enabled_channel_count: 1, key_count: 1, request_capabilities: ['chat'], inventory_source: 'group' },
    ],
};

export const mockUpdateLatest = {
    tag_name: 'v1.22.0',
    published_at: '2026-06-27T00:00:00Z',
    body: '## v1.22.0\n- UI refinement\n- Visual regression tests',
    message: 'v1.22.0 released',
};

export const mockUpdateNowVersion = 'v1.21.0';

export const mockUpdateStatus = {
    version: 'v1.21.0',
    self_update_supported: false,
    self_update_unsupported_reason: 'self-update is not supported in docker container',
};

export const mockAIGovernanceOverview = {
    enabled: false,
    execution_source: {
        mode: 'manual',
        base_url: '',
        channel_type: '',
        model: '',
        use_local_default: true,
        label: '本地默认',
    },
    runtime_policy: {
        strategy: 'deterministic',
        dispatch_mode: 'priority',
        max_parallel_runs: 1,
        double_review_enabled: false,
        fallback_to_deterministic: true,
        degraded_to_deterministic: true,
        label: '确定性策略',
    },
    managed_group_name: '',
    learning: {
        enabled: false,
        sample_count: 0,
    },
};

export const mockStrategyProfiles = [];
