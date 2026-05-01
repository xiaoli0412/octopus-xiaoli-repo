import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';

export interface AIAutomationConfigValues {
    base_url: string;
    api_key?: string;
    channel_type: string;
    model: string;
    use_local_default: boolean;
}

export interface AIAutomationProfileRef {
    id: number;
    name: string;
    domain: string;
    version: number;
    status: string;
    confidence: number;
    explanation?: string;
    updated_at: string;
}

export interface AIAutomationConfig {
    enabled: boolean;
    base_url: string;
    api_key?: string;
    channel_type: string;
    model: string;
    use_local_default: boolean;
    default_selection_policy: string;
    requested_config_source_mode: 'manual' | 'ai_profile';
    config_source_mode: 'manual' | 'ai_profile';
    requested_active_ai_profile_id: number;
    active_ai_profile_id: number;
    requested_active_ai_profile?: AIAutomationProfileRef;
    active_ai_profile?: AIAutomationProfileRef;
    source_fallback_reason?: string;
    dynamic_routing_learning_enabled: boolean;
    manual_config?: AIAutomationConfigValues;
    effective_config?: AIAutomationConfigValues;
}

export interface AIAutomationConfigUpdateRequest {
    enabled?: boolean;
    base_url?: string;
    api_key?: string;
    channel_type?: string;
    model?: string;
    use_local_default?: boolean;
}

export interface AIModelCandidate {
    name: string;
    source: string;
    channel_id?: number;
    channel_name?: string;
    available: boolean;
    free_likely: boolean;
    success_rate: number;
    avg_latency_ms: number;
    recommended: boolean;
    reason: string;
}

export interface AIModelsFetchResult {
    source: string;
    candidates: AIModelCandidate[];
    selected_name: string;
    policy: string;
}

export interface AIPromptTemplate {
    id: number;
    name: string;
    source: 'builtin' | 'custom';
    task_type: string;
    domain: string;
    prompt: string;
    work_requirement: string;
    enabled: boolean;
}

export interface CreateAIPromptTemplateRequest {
    name: string;
    task_type: string;
    domain?: string;
    prompt: string;
    work_requirement?: string;
    enabled?: boolean;
}

export interface AITaskStep {
    id: number;
    task_id: number;
    step_key: string;
    name: string;
    status: string;
    message: string;
    input_json?: string;
    output_json?: string;
    checkpoint_state?: string;
    retry_count?: number;
    sort_order: number;
}

export interface AITask {
    id: number;
    type: string;
    input_text: string;
    context_scope: string;
    status: string;
    progress: number;
    error_message: string;
    result_profile_id?: number;
    result_summary: string;
    result_json?: string;
    config_snapshot_json?: string;
    context_payload_json?: string;
    prompt_text?: string;
    selected_model?: string;
    model_reason?: string;
    resume_token?: string;
    resume_state?: string;
    executor_version?: string;
    last_heartbeat_at?: string;
    attempt_count?: number;
    created_at: string;
    updated_at: string;
    steps?: AITaskStep[];
}

export interface AITaskListParams {
    page?: number;
    page_size?: number;
    status?: string;
    type?: string;
    profile_domain?: string;
    keyword?: string;
    created_from?: string;
    created_to?: string;
}

export interface AITaskListResult {
    items: AITask[];
    total: number;
    page: number;
    page_size: number;
}

export interface AIAutomationTaskConfigSnapshot {
    base_url?: string;
    api_key?: string;
    channel_type?: string;
    model?: string;
    use_local_default?: boolean;
    tool_keys?: string[];
}

export interface CreateAITaskRequest {
    type: string;
    input_text: string;
    context_scope?: string;
    prompt_template_ids?: number[];
    custom_prompt?: string;
    config_snapshot?: AIAutomationTaskConfigSnapshot;
}


export interface AITaskArtifacts {
    task_id: number;
    config_snapshot_json?: string;
    config_snapshot?: AIAutomationTaskConfigSnapshot;
    context_payload_json?: string;
    context_payload?: unknown;
    result_json?: string;
    result_payload?: unknown;
    prompt_text?: string;
    selected_model?: string;
    model_reason?: string;
    resume_state?: string;
    steps?: AITaskStep[];
}
export interface AIProfile {
    id: number;
    domain: string;
    name: string;
    version: number;
    status: string;
    confidence: number;
    explanation: string;
    source_task_id?: number;
    migration_status?: string;
    migration_error?: string;
    domain_payload_type?: string;
    domain_payload?: Record<string, unknown>;
    created_at: string;
    updated_at: string;
    versions?: AIProfileVersion[];
}

export interface AIProfileVersion {
    id: number;
    profile_id: number;
    version: number;
    content_json: string;
    explanation: string;
    created_at: string;
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

function compactQueryParams(params: AITaskListParams): Record<string, string | number | boolean> {
    return Object.fromEntries(
        Object.entries(params).filter(([, value]) => value !== undefined && value !== null && value !== '')
    ) as Record<string, string | number | boolean>;
}

export function useAIAutomationConfig() {
    return useQuery({
        queryKey: ['ai-automation', 'config'],
        queryFn: () => apiClient.get<AIAutomationConfig>('/api/v1/ai/config'),
    });
}

export function useUpdateAIAutomationConfig() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: AIAutomationConfigUpdateRequest) => apiClient.post<AIAutomationConfig>('/api/v1/ai/config', data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ai-automation', 'config'] }),
    });
}

export function useFetchAIModels() {
    return useMutation({
        mutationFn: (data: { base_url?: string; api_key?: string; channel_type?: string; use_local_default?: boolean }) => apiClient.post<AIModelsFetchResult>('/api/v1/ai/models/fetch', data),
    });
}

export function useAIPromptTemplates() {
    return useQuery({
        queryKey: ['ai-automation', 'prompt-templates'],
        queryFn: () => apiClient.get<AIPromptTemplate[]>('/api/v1/ai/prompt-templates'),
    });
}

export function useCreateAIPromptTemplate() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateAIPromptTemplateRequest) => apiClient.post<AIPromptTemplate>('/api/v1/ai/prompt-templates', data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ai-automation', 'prompt-templates'] }),
    });
}

export function useCreateAITask() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateAITaskRequest) => apiClient.post<AITask>('/api/v1/ai/tasks', data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ai-automation'] }),
    });
}

export function useAITask(id?: number) {
    return useQuery({
        queryKey: ['ai-automation', 'task', id],
        queryFn: () => apiClient.get<AITask>(`/api/v1/ai/tasks/${id}`),
        enabled: !!id,
        refetchInterval: (query) => {
            const task = query.state.data;
            if (!task) return 1500;
            return task.status === 'pending' || task.status === 'running' || task.status === 'recoverable' ? 1500 : false;
        },
    });
}


export function useAITaskArtifacts(id?: number) {
    return useQuery({
        queryKey: ['ai-automation', 'task-artifacts', id],
        queryFn: () => apiClient.get<AITaskArtifacts>(`/api/v1/ai/tasks/${id}/artifacts`),
        enabled: !!id,
    });
}
export function useAITasks(params: AITaskListParams) {
    return useQuery({
        queryKey: ['ai-automation', 'tasks', params],
        queryFn: () => apiClient.get<AITaskListResult>('/api/v1/ai/tasks', compactQueryParams(params)),
        refetchInterval: (query) => {
            const hasRunning = query.state.data?.items?.some((task) => task.status === 'pending' || task.status === 'running' || task.status === 'recoverable');
            return hasRunning ? 3000 : false;
        },
    });
}


export function useRetryAITask() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => apiClient.post<AITask>(`/api/v1/ai/tasks/${id}/retry`, {}),
        onSuccess: (task) => {
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'task', task.id] });
            queryClient.invalidateQueries({ queryKey: ['ai-automation'] });
        },
    });
}
export function useCancelAITask() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => apiClient.post<AITask>(`/api/v1/ai/tasks/${id}/cancel`, {}),
        onSuccess: (_task, id) => {
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'task', id] });
            queryClient.invalidateQueries({ queryKey: ['ai-automation'] });
        },
    });
}

export function useAIProfiles() {
    return useQuery({
        queryKey: ['ai-automation', 'profiles'],
        queryFn: () => apiClient.get<AIProfile[]>('/api/v1/ai/profiles'),
    });
}

export function useAIProfile(id?: number) {
    return useQuery({
        queryKey: ['ai-automation', 'profile', id],
        queryFn: () => apiClient.get<AIProfile>(`/api/v1/ai/profiles/${id}`),
        enabled: !!id,
    });
}

export function useActivateAIProfile() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => apiClient.post<AIProfile>(`/api/v1/ai/profiles/${id}/activate`, {}),
        onSuccess: (profile) => {
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'profiles'] });
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'config'] });
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'profile', profile.id] });
            queryClient.invalidateQueries({ queryKey: ['settings', 'list'] });
        },
    });
}

export function useDynamicRouteLearning() {
    return useQuery({
        queryKey: ['ai-automation', 'dynamic-route-learning'],
        queryFn: () => apiClient.get<DynamicRouteLearningListResult>('/api/v1/dynamic-routing/learning'),
        refetchInterval: 30000,
    });
}

export function useResetDynamicRouteLearning() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: () => apiClient.post<null>('/api/v1/dynamic-routing/learning/reset', {}),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'dynamic-route-learning'] });
            queryClient.invalidateQueries({ queryKey: ['ai-automation', 'config'] });
            queryClient.invalidateQueries({ queryKey: ['settings', 'list'] });
        },
    });
}



