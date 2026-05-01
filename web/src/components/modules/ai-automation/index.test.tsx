import * as React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { AIAutomation } from './index';
import { AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY } from './focus-target';

const state = vi.hoisted(() => ({
  config: {
    enabled: true,
    base_url: 'http://127.0.0.1:8080/v1',
    api_key: '',
    channel_type: 'openai-compatible',
    model: 'gpt-free',
    use_local_default: true,
    default_selection_policy: 'free-success-latency',
    requested_config_source_mode: 'manual' as 'manual' | 'ai_profile',
    config_source_mode: 'manual' as 'manual' | 'ai_profile',
    requested_active_ai_profile_id: 11,
    active_ai_profile_id: 11,
    requested_active_ai_profile: { id: 11, domain: 'grouping', name: 'AI Profile A', version: 1, status: 'ready', confidence: 0.92, explanation: 'profile-a', updated_at: '2026-04-24T00:00:00Z' },
    active_ai_profile: { id: 11, domain: 'grouping', name: 'AI Profile A', version: 1, status: 'ready', confidence: 0.92, explanation: 'profile-a', updated_at: '2026-04-24T00:00:00Z' },
    source_fallback_reason: '',
    dynamic_routing_learning_enabled: true,
    manual_config: {
      base_url: 'http://127.0.0.1:8080/v1',
      api_key: '',
      channel_type: 'openai-compatible',
      model: 'gpt-free',
      use_local_default: true,
    },
    effective_config: {
      base_url: 'http://127.0.0.1:8080/v1',
      api_key: '',
      channel_type: 'openai-compatible',
      model: 'gpt-free',
      use_local_default: true,
    },
  },
  promptTemplates: [
    { id: 1, name: 'Builtin Group', source: 'builtin', task_type: 'group_suggestion', domain: 'grouping', prompt: 'builtin prompt', work_requirement: '', enabled: true },
    { id: 2, name: 'Builtin NL', source: 'builtin', task_type: 'natural_language', domain: 'general', prompt: 'nl prompt', work_requirement: '', enabled: true },
  ],
  profiles: [
    { id: 11, domain: 'grouping', name: 'AI Profile A', version: 1, status: 'ready', confidence: 0.92, explanation: 'profile-a', created_at: '2026-04-24T00:00:00Z', updated_at: '2026-04-24T00:00:00Z' },
    { id: 12, domain: 'pricing', name: 'AI Profile B', version: 1, status: 'ready', confidence: 0.73, explanation: 'profile-b', created_at: '2026-04-24T00:00:00Z', updated_at: '2026-04-24T00:00:00Z' },
  ],
  profileDetails: {
    11: { id: 11, domain: 'grouping', name: 'AI Profile A', version: 1, status: 'ready', confidence: 0.92, explanation: 'profile-a', migration_status: 'typed_backfilled', domain_payload: { summary: 'group summary' }, versions: [{ id: 101, profile_id: 11, version: 1, content_json: '{"groups":["alpha"]}', explanation: 'v1', created_at: '2026-04-24T00:00:00Z' }], created_at: '2026-04-24T00:00:00Z', updated_at: '2026-04-24T00:00:00Z' },
    12: { id: 12, domain: 'pricing', name: 'AI Profile B', version: 1, status: 'ready', confidence: 0.73, explanation: 'profile-b', versions: [{ id: 102, profile_id: 12, version: 1, content_json: '{"prices":["beta"]}', explanation: 'v1', created_at: '2026-04-24T00:00:00Z' }], created_at: '2026-04-24T00:00:00Z', updated_at: '2026-04-24T00:00:00Z' },
  } as Record<number, unknown>,
  latestTask: undefined as undefined | { id: number; status: string; progress: number; result_summary: string; result_json?: string; error_message: string; steps: Array<{ id: number; name: string; status: string; message: string }>; selected_model?: string; updated_at?: string; type?: string },
  taskHistory: { items: [{ id: 601, type: 'group_suggestion', status: 'succeeded', progress: 100, result_summary: 'history task', error_message: '', created_at: '2026-04-24T00:00:00Z', updated_at: '2026-04-24T00:00:00Z' }], total: 1, page: 1, page_size: 8 },
  learning: {
    enabled: true,
    states: [
      { id: 301, channel_id: 7, channel_key_id: 3, model_name: 'gpt-free', success_count: 9, failure_count: 1, fallback_count: 2, race_winner_count: 6, latency_ms_ewma: 180, score: 0.88, confidence: 0.81, last_sample_at: 1713900000 },
    ],
  },
  modelCandidates: [
    { name: 'gpt-free', source: 'local', channel_name: 'Local Octopus', available: true, free_likely: true, success_rate: 0.97, avg_latency_ms: 220, recommended: true, reason: 'free and healthy' },
  ],
  updateConfigMutateAsync: vi.fn(),
  fetchModelsMutateAsync: vi.fn(),
  createPromptTemplateMutateAsync: vi.fn(),
  createTaskMutateAsync: vi.fn(),
  cancelTaskMutate: vi.fn(),
  retryTaskMutate: vi.fn(),
  activateProfileMutateAsync: vi.fn(),
  resetLearningMutate: vi.fn(),
  setSettingMutate: vi.fn(),
  configRefetch: vi.fn(),
  learningRefetch: vi.fn(),
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
  toastWarning: vi.fn(),
  scrollIntoViewMock: vi.fn(),
}));

vi.mock('next-intl', () => ({
  useTranslations: (namespace?: string) => (key: string, values?: Record<string, unknown>) => {
    if (!values) return `${namespace}.${key}`;
    return `${namespace}.${key}:${JSON.stringify(values)}`;
  },
}));

vi.mock('@/components/common/Toast', () => ({
  toast: {
    success: state.toastSuccess,
    error: state.toastError,
    warning: state.toastWarning,
  },
}));

vi.mock('@/components/common/PageWrapper', () => ({
  PageWrapper: ({ children, className, 'data-testid': testId }: { children: React.ReactNode; className?: string; 'data-testid'?: string }) => <div className={className} data-testid={testId}>{children}</div>,
}));

vi.mock('@/components/ui/button', () => ({
  Button: ({ children, type = 'button', onClick, disabled, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button type={type} onClick={onClick} disabled={disabled} {...props}>{children}</button>,
}));

vi.mock('@/components/ui/input', () => ({
  Input: (props: React.InputHTMLAttributes<HTMLInputElement>) => <input {...props} />,
}));

vi.mock('@/components/ui/badge', () => ({
  Badge: ({ children, ...props }: React.HTMLAttributes<HTMLSpanElement>) => <span {...props}>{children}</span>,
}));

vi.mock('@/components/ui/progress', () => ({
  Progress: ({ value }: { value?: number }) => <div data-testid="progress">{value ?? 0}</div>,
}));

vi.mock('@/components/ui/card', () => ({
  Card: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <section {...props}>{children}</section>,
  CardHeader: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>,
  CardTitle: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>,
  CardDescription: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>,
  CardContent: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>,
}));

vi.mock('@/components/ui/switch', () => ({
  Switch: ({ checked, onCheckedChange, ...props }: { checked?: boolean; onCheckedChange?: (value: boolean) => void } & React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button type="button" role="switch" aria-checked={checked ? 'true' : 'false'} onClick={() => onCheckedChange?.(!checked)} {...props}>switch</button>
  ),
}));

vi.mock('@/components/ui/select', async () => {
  const React = await import('react');
  type SelectItemLikeProps = { children?: React.ReactNode; value?: string };
  const SelectContext = React.createContext<{ items: Array<{ label: string; value: string }>; onValueChange?: (value: string) => void; value?: string }>({ items: [], onValueChange: undefined, value: undefined });

  function collectItems(children: React.ReactNode, items: Array<{ label: string; value: string }> = []) {
    React.Children.forEach(children, (child) => {
      if (!React.isValidElement(child)) return;
      const element = child as React.ReactElement<SelectItemLikeProps>;
      if ((element.type as { __testIsSelectItem?: boolean }).__testIsSelectItem) {
        items.push({ label: typeof element.props.children === 'string' ? element.props.children : String(element.props.value), value: element.props.value ?? '' });
        return;
      }
      if (element.props?.children) collectItems(element.props.children, items);
    });
    return items;
  }

  function Select({ children, onValueChange, value }: { children: React.ReactNode; onValueChange?: (value: string) => void; value?: string }) {
    const items = React.useMemo(() => collectItems(children), [children]);
    const contextValue = React.useMemo(() => ({ items, onValueChange, value }), [items, onValueChange, value]);
    return <SelectContext.Provider value={contextValue}>{children}</SelectContext.Provider>;
  }

  function SelectTrigger({ className, children, id, 'aria-label': ariaLabel }: { className?: string; children?: React.ReactNode; id?: string; 'aria-label'?: string }) {
    const context = React.useContext(SelectContext);
    return (
      <select id={id} aria-label={ariaLabel} className={className} role="combobox" onChange={(event) => context.onValueChange?.(event.target.value)} value={context.value}>
        {children}
        {context.items.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
      </select>
    );
  }

  function SelectValue() { return null; }
  function SelectContent({ children }: { children?: React.ReactNode }) { return <>{children}</>; }
  function SelectItem() { return null; }
  (SelectItem as { __testIsSelectItem?: boolean }).__testIsSelectItem = true;

  return { Select, SelectContent, SelectItem, SelectTrigger, SelectValue };
});

vi.mock('@/api/endpoints/setting', () => ({
  useSetSetting: () => ({ mutate: state.setSettingMutate }),
}));

vi.mock('@/api/endpoints/ai-automation', () => ({
  useAIAutomationConfig: () => ({ data: state.config, refetch: state.configRefetch }),
  useUpdateAIAutomationConfig: () => ({ mutateAsync: state.updateConfigMutateAsync }),
  useFetchAIModels: () => ({ isPending: false, data: { candidates: state.modelCandidates }, mutateAsync: state.fetchModelsMutateAsync }),
  useAIPromptTemplates: () => ({ data: state.promptTemplates }),
  useCreateAIPromptTemplate: () => ({ mutateAsync: state.createPromptTemplateMutateAsync }),
  useCreateAITask: () => ({ isPending: false, data: state.latestTask, mutateAsync: state.createTaskMutateAsync }),
  useAITask: () => ({ data: state.latestTask }),
  useAITaskArtifacts: () => ({ data: state.latestTask?.result_json ? { result_json: state.latestTask.result_json } : undefined }),
  useAITasks: () => ({ data: state.taskHistory, isLoading: false }),
  useCancelAITask: () => ({ isPending: false, mutate: state.cancelTaskMutate }),
  useRetryAITask: () => ({ isPending: false, mutate: state.retryTaskMutate }),
  useAIProfiles: () => ({ data: state.profiles }),
  useAIProfile: (id?: number) => ({ data: id ? state.profileDetails[id] : undefined }),
  useActivateAIProfile: () => ({ isPending: false, mutateAsync: state.activateProfileMutateAsync }),
  useDynamicRouteLearning: () => ({ data: state.learning, refetch: state.learningRefetch }),
  useResetDynamicRouteLearning: () => ({ isPending: false, mutate: state.resetLearningMutate }),
}));

describe('AIAutomation', () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    Object.defineProperty(Element.prototype, 'scrollIntoView', { configurable: true, value: state.scrollIntoViewMock });

    state.latestTask = undefined;
    state.updateConfigMutateAsync.mockReset();
    state.fetchModelsMutateAsync.mockReset();
    state.createPromptTemplateMutateAsync.mockReset();
    state.createTaskMutateAsync.mockReset();
    state.cancelTaskMutate.mockReset();
    state.retryTaskMutate.mockReset();
    state.activateProfileMutateAsync.mockReset();
    state.resetLearningMutate.mockReset();
    state.setSettingMutate.mockReset();
    state.configRefetch.mockReset();
    state.learningRefetch.mockReset();
    state.toastSuccess.mockReset();
    state.toastError.mockReset();
    state.toastWarning.mockReset();
    state.scrollIntoViewMock.mockReset();

	state.config.api_key = '';
	state.config.manual_config.api_key = '';
	state.config.effective_config.api_key = '';

    state.updateConfigMutateAsync.mockResolvedValue(state.config);
    state.fetchModelsMutateAsync.mockResolvedValue({ source: 'local', candidates: state.modelCandidates, selected_name: 'gpt-free', policy: 'free-success-latency' });
    state.createPromptTemplateMutateAsync.mockImplementation(async (payload: { name: string; task_type: string; prompt: string; work_requirement?: string }) => ({ id: 99, name: payload.name, source: 'custom', task_type: payload.task_type, domain: 'custom', prompt: payload.prompt, work_requirement: payload.work_requirement ?? '', enabled: true }));
    state.createTaskMutateAsync.mockResolvedValue({ id: 501, status: 'pending', progress: 10, result_summary: 'queued', error_message: '', steps: [], selected_model: 'gpt-free' });
    state.activateProfileMutateAsync.mockResolvedValue(undefined);
    state.setSettingMutate.mockImplementation((_payload: unknown, options?: { onSuccess?: () => void }) => options?.onSuccess?.());
    state.resetLearningMutate.mockImplementation((_payload: unknown, options?: { onSuccess?: () => void }) => options?.onSuccess?.());
  });

  it('renders the rebuilt workbench shell and learning section markers', () => {
    render(<AIAutomation />);

    expect(screen.getByTestId('ai-automation-page')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-section')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-controls')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-preset-card')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-switch-card')).toBeInTheDocument();
    expect(screen.getByText('aiAutomation.hero.title')).toBeInTheDocument();
  });

  it('creates a task from the main chain', async () => {
    render(<AIAutomation />);

    fireEvent.change(screen.getByPlaceholderText('aiAutomation.task.placeholder'), { target: { value: '整理当前渠道和分组结构' } });
    fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.task.run' }));

    await waitFor(() => {
      expect(state.createTaskMutateAsync).toHaveBeenCalledTimes(1);
    });

    const payload = state.createTaskMutateAsync.mock.calls[0]?.[0];
    expect(payload.type).toBe('natural_language');
    expect(payload.input_text).toBe('整理当前渠道和分组结构');
    expect(payload.config_snapshot.model).toBe('gpt-free');
  });

  it('creates a custom prompt template from the template workspace', async () => {
    render(<AIAutomation />);

    fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.task.promptTemplatesTitle' }));
    fireEvent.change(screen.getByPlaceholderText('aiAutomation.task.templateNamePlaceholder'), { target: { value: 'My Template' } });
    fireEvent.change(screen.getByPlaceholderText('aiAutomation.task.templatePromptPlaceholder'), { target: { value: 'normalize channels' } });
    fireEvent.change(screen.getByPlaceholderText('aiAutomation.task.templateRequirementPlaceholder'), { target: { value: 'keep manual config' } });
    fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.task.createTemplate' }));

    await waitFor(() => {
      expect(state.createPromptTemplateMutateAsync).toHaveBeenCalledTimes(1);
    });
  });

  it('shows structured profile preview in the profiles workspace', () => {
    render(<AIAutomation />);

    fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.profiles.title' }));
    expect(screen.getByTestId('ai-profile-structured-preview')).toBeInTheDocument();
  });

  it('renders task history list and paging controls', () => {
    render(<AIAutomation />);

    fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.taskHistory.historyTitle' }));
    expect(screen.getByText(/#601/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'aiAutomation.taskHistory.historyPrev' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'aiAutomation.taskHistory.historyNext' })).toBeInTheDocument();
  });

  it('consumes queued learning focus target and scrolls to learning section', async () => {
    window.sessionStorage.setItem(AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY, 'learning');

    render(<AIAutomation />);

    await waitFor(() => {
      expect(state.scrollIntoViewMock).toHaveBeenCalled();
    });
  });

  it('renders learning summaries and state cards', () => {
    render(<AIAutomation />);

    expect(screen.getByTestId('ai-automation-learning-state-summary')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-secondary-summary')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-states')).toBeInTheDocument();
    expect(screen.getByTestId('ai-automation-learning-state-301')).toBeInTheDocument();
  });

  it('toggles learning preset and reset action', async () => {
    render(<AIAutomation />);

    fireEvent.click(screen.getByTestId('ai-automation-learning-preset-safe'));

    await waitFor(() => {
      expect(state.setSettingMutate).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByTestId('ai-automation-learning-reset'));

    await waitFor(() => {
      expect(state.resetLearningMutate).toHaveBeenCalled();
    });
  });

  it('shows empty learning state when there are no samples', () => {
    state.learning = { enabled: true, states: [] };

    render(<AIAutomation />);

    expect(screen.getByTestId('ai-automation-learning-empty')).toBeInTheDocument();
  });

	it('omits redacted api key placeholder from config save payload', async () => {
	  state.config.api_key = '[redacted]';
	  state.config.manual_config.api_key = '[redacted]';
	  state.config.effective_config.api_key = '[redacted]';

	  render(<AIAutomation />);
	  fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.config.save' }));

	  await waitFor(() => {
		expect(state.updateConfigMutateAsync).toHaveBeenCalledTimes(1);
	  });

	  expect(state.updateConfigMutateAsync.mock.calls[0]?.[0]?.api_key).toBeUndefined();
	});

	it('omits redacted api key placeholder from model fetch payload', async () => {
	  state.config.api_key = '[redacted]';
	  state.config.manual_config.api_key = '[redacted]';
	  state.config.effective_config.api_key = '[redacted]';

	  render(<AIAutomation />);
	  fireEvent.click(screen.getByRole('button', { name: 'aiAutomation.config.fetchModels' }));

	  await waitFor(() => {
		expect(state.fetchModelsMutateAsync).toHaveBeenCalledTimes(1);
	  });

	  expect(state.fetchModelsMutateAsync.mock.calls[0]?.[0]?.api_key).toBeUndefined();
	});
});
