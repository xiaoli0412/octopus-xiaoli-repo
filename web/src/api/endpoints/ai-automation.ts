import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';

export interface AIGovernanceExecutionSource {
	mode: 'manual' | 'ai_profile';
	base_url: string;
	channel_type: string;
	model: string;
	use_local_default: boolean;
	label: string;
}

export interface AIGovernanceLearningSummary {
	enabled: boolean;
	sample_count: number;
	top_target?: string;
	last_sample_at?: number;
	top_score?: number;
}

export interface GovernanceSessionSummary {
	id: number;
	goal: string;
	scope: string;
	expert_preset_id: string;
	status: string;
	current_stage: string;
	operator_summary: string;
	risk_summary: string;
	confidence: number;
	mutation_count: number;
	can_apply: boolean;
	created_at: string;
	updated_at: string;
	applied_at?: string;
}

export interface GovernanceFindingView {
	severity: string;
	title: string;
	detail: string;
}

export interface GovernanceDecisionView {
	title: string;
	summary: string;
}

export interface GovernanceGroupUpsertMutation {
	group_name: string;
	mode: number;
}

export interface GovernanceGroupItemMutation {
	group_name: string;
	channel_id: number;
	model_name: string;
	priority?: number;
	weight?: number;
	channel_key_id?: number;
}

export interface GovernanceGroupItemReorderMutation {
	group_name: string;
	items: GovernanceGroupItemMutation[];
}

export interface GovernanceRouteTargetOverrideMutation {
	channel_id: number;
	channel_key_id: number;
	model_name: string;
	billing_mode?: string;
	probe_policy?: string;
	probe_interval_seconds?: number;
	probe_concurrency_limit?: number;
}

export interface GovernanceStrategyProfileActivateMutation {
	strategy_profile_id: number;
}

export interface GovernanceLLMPriceUpsertMutation {
	name: string;
	canonical_name?: string;
	input: number;
	output: number;
	cache_read: number;
	cache_write: number;
	official_input: number;
	official_output: number;
	official_cache_read: number;
	official_cache_write: number;
	billing_mode?: string;
	probe_policy?: string;
	probe_interval_seconds?: number;
	probe_concurrency_limit?: number;
	source?: string;
}

export interface GovernanceSettingMutation {
	key: string;
	value: string;
}

export interface GovernanceRuntimePolicyView {
	strategy: string;
	dispatch_mode: string;
	max_parallel_runs: number;
	double_review_enabled: boolean;
	fallback_to_deterministic: boolean;
	degraded_to_deterministic: boolean;
	label?: string;
}

export interface GovernanceMutation {
	type: string;
	summary: string;
	group_upsert?: GovernanceGroupUpsertMutation;
	group_item_attach?: GovernanceGroupItemMutation;
	group_item_detach?: GovernanceGroupItemMutation;
	group_item_reorder?: GovernanceGroupItemReorderMutation;
	route_target_override_upsert?: GovernanceRouteTargetOverrideMutation;
	route_target_override_delete?: GovernanceRouteTargetOverrideMutation;
	llm_price_upsert?: GovernanceLLMPriceUpsertMutation;
	dynamic_routing_setting_set?: GovernanceSettingMutation;
	runtime_policy_set?: GovernanceRuntimePolicyView;
	strategy_profile_activate?: GovernanceStrategyProfileActivateMutation;
}

export interface GovernanceDomainPlanView {
	key: string;
	title: string;
	summary: string;
	status: string;
	finding_count: number;
	mutation_count: number;
	findings?: GovernanceFindingView[];
	mutations?: GovernanceMutation[];
}

export interface GovernancePlanView {
	findings: GovernanceFindingView[];
	decisions: GovernanceDecisionView[];
	mutations: GovernanceMutation[];
	domains?: GovernanceDomainPlanView[];
	risk_summary: string;
	confidence: number;
	operator_summary: string;
}

export interface GovernancePreviewImpactCounts {
	groups: number;
	items: number;
	overrides: number;
	profiles: number;
}

export interface GovernancePreviewView {
	headline: string;
	summary_lines: string[];
	impact_counts: GovernancePreviewImpactCounts;
	risk_notes: string[];
	apply_blockers: string[];
	can_apply: boolean;
	mutation_count: number;
	mutations: GovernanceMutation[];
}

export interface GovernanceApplyAuditItem {
	mutation_type: string;
	summary: string;
	status: string;
	message: string;
}

export interface GovernanceApplyAudit {
	summary: string;
	items: GovernanceApplyAuditItem[];
}

export interface GovernanceApplyRunView {
	id: number;
	session_id: number;
	status: string;
	result_summary: string;
	error_message?: string;
	audit: GovernanceApplyAudit;
	created_at: string;
	updated_at: string;
}

export interface GovernanceSnapshotSummary {
	channels: number;
	enabled_channels: number;
	groups: number;
	group_items: number;
	route_target_overrides: number;
	models: number;
	missing_prices?: number;
	active_source_mode: string;
	active_source_label: string;
	highlights: string[];
}

export interface GovernanceRollbackPointView {
	id: number;
	session_id: number;
	apply_run_id?: number;
	snapshot_checksum: string;
	summary: string;
	created_at: string;
	updated_at: string;
}

export interface GovernanceSessionDetail extends GovernanceSessionSummary {
	plan: GovernancePlanView;
	preview: GovernancePreviewView;
	snapshot_checksum: string;
	apply_runs: GovernanceApplyRunView[];
	rollback_points?: GovernanceRollbackPointView[];
	snapshot_summary: GovernanceSnapshotSummary;
}

export interface StrategyProfileSummary {
	id: number;
	name: string;
	summary: string;
	status: string;
	source_session_id?: number;
	activated_at?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface ExpertPresetView {
	id: string;
	name: string;
	description: string;
	review_depth: string;
	create_managed_group: boolean;
	sync_bindings: boolean;
	cleanup_stale: boolean;
}

export interface AIGovernanceOverview {
	enabled: boolean;
	execution_source: AIGovernanceExecutionSource;
	runtime_policy: GovernanceRuntimePolicyView;
	managed_group_name: string;
	learning: AIGovernanceLearningSummary;
	active_strategy_profile?: StrategyProfileSummary;
	recent_session?: GovernanceSessionSummary;
}

export interface GovernanceSessionCreateRequest {
	goal: string;
	expert_preset_id?: string;
}

export interface GovernanceStrategyProfileCreateRequest {
	session_id: number;
	name: string;
}

export interface GovernanceRollbackRequest {
	rollback_point_id?: number;
}

export interface DynamicRouteLearningState {
	id: number;
	channel_id: number;
	channel_key_id: number;
	model_name: string;
	success_count: number;
	failure_count: number;
	fallback_count: number;
	race_winner_count: number;
	latency_ms_ewma: number;
	score: number;
	confidence: number;
	last_sample_at: number;
}

export interface DynamicRouteLearningListResult {
	enabled: boolean;
	states: DynamicRouteLearningState[];
}

export function useAIGovernanceOverview() {
	return useQuery({
		queryKey: ['ai-governance', 'overview'],
		queryFn: () => apiClient.get<AIGovernanceOverview>('/api/v1/ai/overview'),
	});
}

export function useGovernanceSessions() {
	return useQuery({
		queryKey: ['ai-governance', 'sessions'],
		queryFn: () => apiClient.get<GovernanceSessionSummary[]>('/api/v1/ai/sessions'),
	});
}

export function useGovernanceSession(id?: number) {
	return useQuery({
		queryKey: ['ai-governance', 'session', id],
		queryFn: () => apiClient.get<GovernanceSessionDetail>(`/api/v1/ai/sessions/${id}`),
		enabled: typeof id === 'number' && id > 0,
	});
}

export function useCreateGovernanceSession() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: GovernanceSessionCreateRequest) => apiClient.post<GovernanceSessionDetail>('/api/v1/ai/sessions', data),
		onSuccess: (session) => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'sessions'] });
			queryClient.setQueryData(['ai-governance', 'session', session.id], session);
		},
	});
}

export function useReplanGovernanceSession() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (id: number) => apiClient.post<GovernanceSessionDetail>(`/api/v1/ai/sessions/${id}/replan`, {}),
		onSuccess: (session) => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'sessions'] });
			queryClient.setQueryData(['ai-governance', 'session', session.id], session);
		},
	});
}

export function useApplyGovernanceSession() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (id: number) => apiClient.post<GovernanceSessionDetail>(`/api/v1/ai/sessions/${id}/apply`, {}),
		onSuccess: (session) => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'sessions'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'strategy-profiles'] });
			queryClient.setQueryData(['ai-governance', 'session', session.id], session);
		},
	});
}

export function useRollbackGovernanceSession() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ id, rollback_point_id }: { id: number; rollback_point_id?: number }) => apiClient.post<GovernanceSessionDetail>(`/api/v1/ai/sessions/${id}/rollback`, { rollback_point_id }),
		onSuccess: (session) => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'sessions'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'rollback-points'] });
			queryClient.setQueryData(['ai-governance', 'session', session.id], session);
		},
	});
}

export function useGovernanceApplyRuns(id?: number) {
	return useQuery({
		queryKey: ['ai-governance', 'apply-runs', id],
		queryFn: () => apiClient.get<GovernanceApplyRunView[]>(`/api/v1/ai/sessions/${id}/apply-runs`),
		enabled: typeof id === 'number' && id > 0,
	});
}

export function useGovernanceRollbackPoints(sessionID?: number) {
	return useQuery({
		queryKey: ['ai-governance', 'rollback-points', sessionID],
		queryFn: () => apiClient.get<GovernanceRollbackPointView[]>(sessionID ? `/api/v1/ai/rollback-points?session_id=${sessionID}` : '/api/v1/ai/rollback-points'),
	});
}

export function useStrategyProfiles() {
	return useQuery({
		queryKey: ['ai-governance', 'strategy-profiles'],
		queryFn: () => apiClient.get<StrategyProfileSummary[]>('/api/v1/ai/strategy-profiles'),
	});
}

export function useCreateStrategyProfile() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: GovernanceStrategyProfileCreateRequest) => apiClient.post<StrategyProfileSummary>('/api/v1/ai/strategy-profiles', data),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ai-governance', 'strategy-profiles'] }),
	});
}

export function useActivateStrategyProfile() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (id: number) => apiClient.post<StrategyProfileSummary>(`/api/v1/ai/strategy-profiles/${id}/activate`, {}),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'strategy-profiles'] });
		},
	});
}

export function useExpertPresets() {
	return useQuery({
		queryKey: ['ai-governance', 'expert-presets'],
		queryFn: () => apiClient.get<ExpertPresetView[]>('/api/v1/ai/expert-presets'),
	});
}

export function useGovernanceRuntimePolicy() {
	return useQuery({
		queryKey: ['ai-governance', 'runtime-policy'],
		queryFn: () => apiClient.get<GovernanceRuntimePolicyView>('/api/v1/ai/runtime-policy'),
	});
}

export function useUpdateGovernanceRuntimePolicy() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: GovernanceRuntimePolicyView) => apiClient.post<GovernanceRuntimePolicyView>('/api/v1/ai/runtime-policy', data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'runtime-policy'] });
		},
	});
}

export function useAIGovernanceLearningSummary() {
	return useQuery({
		queryKey: ['ai-governance', 'learning-summary'],
		queryFn: () => apiClient.get<AIGovernanceLearningSummary>('/api/v1/ai/learning/summary'),
	});
}

export function useDynamicRouteLearning() {
	return useQuery({
		queryKey: ['dynamic-routing', 'learning'],
		queryFn: () => apiClient.get<DynamicRouteLearningListResult>('/api/v1/dynamic-routing/learning'),
	});
}

export function useResetDynamicRouteLearning() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: () => apiClient.post<null>('/api/v1/dynamic-routing/learning/reset', {}),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['dynamic-routing', 'learning'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'overview'] });
			queryClient.invalidateQueries({ queryKey: ['ai-governance', 'learning-summary'] });
		},
	});
}
