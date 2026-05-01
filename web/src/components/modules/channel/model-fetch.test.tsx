import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as React from 'react';

import { ChannelForm, type ChannelFormData } from './Form';
import { ChannelType } from '@/api/endpoints/channel';

const mocks = vi.hoisted(() => ({
	fetchModelMutateAsyncMock: vi.fn(),
	toastWarningMock: vi.fn(),
	toastErrorMock: vi.fn(),
	toastSuccessMock: vi.fn(),
}));

vi.mock('next-intl', () => ({
	useTranslations: (namespace?: string) => (key: string, values?: Record<string, unknown>) => {
		if (!values) return `${namespace}.${key}`;
		return `${namespace}.${key}:${JSON.stringify(values)}`;
	},
}));

vi.mock('@/components/common/Toast', () => ({
	toast: {
		warning: mocks.toastWarningMock,
		error: mocks.toastErrorMock,
		success: mocks.toastSuccessMock,
	},
}));

vi.mock('@/api/endpoints/channel', async () => {
	const actual = await vi.importActual<typeof import('@/api/endpoints/channel')>('@/api/endpoints/channel');
	return {
		...actual,
		useFetchModel: () => ({
			isPending: false,
			mutateAsync: mocks.fetchModelMutateAsyncMock,
		}),
		useRouteTargetOverrideList: () => ({
			data: [],
			isLoading: false,
		}),
		useUpsertRouteTargetOverride: () => ({
			isPending: false,
			mutateAsync: vi.fn(),
		}),
		useDeleteRouteTargetOverride: () => ({
			isPending: false,
			mutateAsync: vi.fn(),
		}),
		useTestChannelModelsByConfig: () => ({
			isPending: false,
			mutateAsync: vi.fn(),
		}),
		useCopilotRequestDeviceCode: () => ({ isPending: false, mutateAsync: vi.fn() }),
		useCopilotPollToken: () => ({ isPending: false, mutateAsync: vi.fn() }),
		useAntigravityOAuthStart: () => ({ isPending: false, mutateAsync: vi.fn() }),
		useAntigravityOAuthPoll: () => ({ isPending: false, mutateAsync: vi.fn() }),
	};
});

vi.mock('@/api/endpoints/providers', () => ({
	useProviders: () => ({ data: [] }),
}));

vi.mock('@/api/endpoints/model', () => ({
	BILLING_MODE_OPTIONS: [],
	PROBE_POLICY_OPTIONS: [],
}));

vi.mock('@/components/ui/select', async () => {
	const React = await import('react');
	type SelectItemLikeProps = { children?: React.ReactNode; value?: string };
	const SelectContext = React.createContext<{
		items: Array<{ label: string; value: string }>;
		onValueChange?: (value: string) => void;
		value?: string;
	}>({ items: [], onValueChange: undefined, value: undefined });

	function collectItems(children: React.ReactNode, items: Array<{ label: string; value: string }> = []) {
		React.Children.forEach(children, (child) => {
			if (!React.isValidElement(child)) return;
			const element = child as React.ReactElement<SelectItemLikeProps>;
			if ((element.type as { __testIsSelectItem?: boolean }).__testIsSelectItem) {
				items.push({
					label: typeof element.props.children === 'string' ? element.props.children : String(element.props.value),
					value: element.props.value ?? '',
				});
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

	function SelectTrigger({ className, children, id }: { className?: string; children?: React.ReactNode; id?: string }) {
		const context = React.useContext(SelectContext);
		return (
			<select id={id} className={className} role="combobox" onChange={(event) => context.onValueChange?.(event.target.value)} value={context.value}>
				{children}
				{context.items.map((item) => (
					<option key={item.value} value={item.value}>{item.label}</option>
				))}
			</select>
		);
	}

	function SelectValue() { return null; }
	function SelectContent({ children }: { children?: React.ReactNode }) { return <>{children}</>; }
	function SelectItem() { return null; }
	(SelectItem as { __testIsSelectItem?: boolean }).__testIsSelectItem = true;

	return { Select, SelectContent, SelectItem, SelectTrigger, SelectValue };
});

vi.mock('@/components/ui/switch', () => ({
	Switch: ({ checked, onCheckedChange }: { checked?: boolean; onCheckedChange?: (value: boolean) => void }) => (
		<button role="switch" aria-checked={checked ? 'true' : 'false'} onClick={() => onCheckedChange?.(!checked)} type="button">
			switch
		</button>
	),
}));

vi.mock('@/components/ui/button', () => ({
	Button: ({ children, type = 'button', onClick, disabled, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
		<button type={type} onClick={onClick} disabled={disabled} {...props}>{children}</button>
	),
}));

vi.mock('@/components/ui/input', () => ({
	Input: (props: React.InputHTMLAttributes<HTMLInputElement>) => <input {...props} />,
}));

vi.mock('@/components/ui/badge', () => ({
	Badge: ({ children, ...props }: React.HTMLAttributes<HTMLSpanElement>) => <span {...props}>{children}</span>,
}));

vi.mock('@/components/ui/dialog', () => ({
	Dialog: ({ children, open }: { children: React.ReactNode; open?: boolean }) => open ? <div>{children}</div> : null,
	DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogTitle: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('@/components/common/HelpHint', () => ({
	HelpHint: ({ children }: { children: React.ReactNode }) => <span>{children}</span>,
}));

vi.mock('@/components/ui/accordion', () => ({
	Accordion: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	AccordionItem: ({ children, ...props }: { children: React.ReactNode }) => <div {...props}>{children}</div>,
	AccordionTrigger: ({ children, addon, addonClassName, showIndicator: _showIndicator, ...props }: { children: React.ReactNode; addon?: React.ReactNode; addonClassName?: string; showIndicator?: boolean }) => (
		<div>
			<button type="button" {...props}>{children}</button>
			{addon ? <div className={addonClassName}>{addon}</div> : null}
		</div>
	),
	AccordionContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

function buildFormData(): ChannelFormData {
	return {
		name: '测试渠道',
		type: ChannelType.OpenAIChat,
		key_management_mode: 'classified',
		key_routing_policy: 'round_robin',
		base_urls: [{ url: 'http://127.0.0.1:19001/v1', delay: 0 }],
		custom_header: [],
		channel_proxy: '',
		param_override: '',
		keys: [{ enabled: true, channel_key: 'sk-test-key', source_type: 'unknown', remark: '', allowed_models: '' }],
		model: '',
		custom_model: '',
		enabled: true,
		proxy: false,
		auto_sync: false,
		auto_group: 0,
		match_regex: '',
	};
}

function renderForm(initialData?: Partial<ChannelFormData>) {
	function Wrapper() {
		const [formData, setFormData] = React.useState<ChannelFormData>({ ...buildFormData(), ...initialData });
		return (
			<ChannelForm
				formData={formData}
				onFormDataChange={setFormData}
				onSubmit={(event) => event.preventDefault()}
				isPending={false}
				submitText="submit"
				pendingText="pending"
				idPrefix="test-channel"
			/>
		);
	}

	return render(<Wrapper />);
}

describe('ChannelForm model fetch', () => {
	beforeEach(() => {
		mocks.fetchModelMutateAsyncMock.mockReset();
		mocks.toastWarningMock.mockReset();
		mocks.toastErrorMock.mockReset();
		mocks.toastSuccessMock.mockReset();
	});

	it('requests model list for the current key and opens the fetched model dialog', async () => {
		mocks.fetchModelMutateAsyncMock.mockResolvedValue(['gpt-4o-mini', 'gpt-4.1']);

		renderForm();

		fireEvent.click(screen.getByTestId('test-channel-key-trigger-0'));
		fireEvent.click(screen.getAllByRole('button', { name: 'channel.form.selectModels' })[0]);

		await waitFor(() => {
			expect(mocks.fetchModelMutateAsyncMock).toHaveBeenCalledTimes(1);
		});

		expect(mocks.fetchModelMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			type: ChannelType.OpenAIChat,
			base_url: 'http://127.0.0.1:19001/v1',
			key: 'sk-test-key',
			proxy: false,
			channel_proxy: null,
			custom_header: [],
		}));

		await waitFor(() => {
			expect(screen.getByText('channel.form.keyFetchedModelList:{"index":1}')).toBeInTheDocument();
			expect(screen.getByText('gpt-4o-mini')).toBeInTheDocument();
			expect(screen.getByText('gpt-4.1')).toBeInTheDocument();
		});
	});

	it('keeps classified mode on per-key model search and hides the global model section', () => {
		renderForm();

		expect(screen.getByTestId('test-channel-key-fetch-models-0')).toBeInTheDocument();
		expect(screen.queryByTestId('test-channel-global-fetch-models')).not.toBeInTheDocument();
	});

	it('warns instead of requesting when base url or key is missing', async () => {
		renderForm({
			base_urls: [{ url: '', delay: 0 }],
			keys: [{ enabled: true, channel_key: '', source_type: 'unknown', remark: '', allowed_models: '' }],
		});

		expect(screen.getByRole('button', { name: 'channel.form.selectModels' })).toBeDisabled();
		fireEvent.click(screen.getByTestId('test-channel-key-trigger-0'));
		fireEvent.click(screen.getByRole('button', { name: 'channel.form.selectModels' }));
		expect(mocks.fetchModelMutateAsyncMock).not.toHaveBeenCalled();
		expect(mocks.toastWarningMock).not.toHaveBeenCalled();
	});

	it('adds a manual model into the current classified key scope', async () => {
		renderForm();

		fireEvent.click(screen.getByTestId('test-channel-key-trigger-0'));
		const draftInput = await screen.findByTestId('test-channel-key-model-draft-0');
		fireEvent.change(draftInput, { target: { value: 'gpt-4.1-mini' } });
		fireEvent.click(screen.getByTestId('test-channel-key-model-add-0'));

		expect(await screen.findByText('gpt-4.1-mini')).toBeInTheDocument();
		expect((draftInput as HTMLInputElement).value).toBe('');
	});
});
