'use client';

import { forwardRef } from 'react';
import { useTranslations } from 'next-intl';
import { BrainCircuit, ClipboardList, History, Layers3, Undo2 } from 'lucide-react';

import type {
	AIGovernanceLearningSummary,
	ExpertPresetView,
	GovernanceApplyRunView,
	GovernanceRollbackPointView,
	GovernanceSessionDetail,
	GovernanceSessionSummary,
	StrategyProfileSummary,
} from '@/api/endpoints/ai-automation';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { formatDateTimeByLocale } from '@/lib/locale';
import type { Locale as AppLocale } from '@/stores/setting';
import { cn } from '@/lib/utils';

export type WorkspaceTab = 'preview' | 'profiles' | 'rollback' | 'history' | 'expert';

interface SessionDetailProps {
	tab: WorkspaceTab;
	onTabChange: (tab: WorkspaceTab) => void;
	session?: GovernanceSessionDetail | null;
	sessions: GovernanceSessionSummary[];
	rollbackPoints: GovernanceRollbackPointView[];
	strategyProfiles: StrategyProfileSummary[];
	presets: ExpertPresetView[];
	learningSummary?: AIGovernanceLearningSummary | null;
	selectedSessionID?: number;
	onSelectSession: (id: number) => void;
	newProfileName: string;
	onNewProfileNameChange: (value: string) => void;
	onCreateProfile: () => void;
	onActivateProfile: (id: number) => void;
	onRollback: (point: GovernanceRollbackPointView) => void;
	isCreatingProfile: boolean;
	isApplying: boolean;
	locale: AppLocale;
}

const WORKSPACE_TABS: Array<{ key: WorkspaceTab; icon: typeof ClipboardList; labelKey: string }> = [
	{ key: 'preview', icon: ClipboardList, labelKey: 'workspace.preview' },
	{ key: 'profiles', icon: Layers3, labelKey: 'workspace.profiles' },
	{ key: 'rollback', icon: Undo2, labelKey: 'workspace.rollback' },
	{ key: 'history', icon: History, labelKey: 'workspace.history' },
	{ key: 'expert', icon: BrainCircuit, labelKey: 'workspace.expert' },
];

const SESSION_STATUS_LABELS: Record<string, string> = {
	draft: '草稿',
	planning: '规划中',
	ready: '可应用',
	stale: '需重算',
	applying: '应用中',
	applied: '已应用',
	failed: '失败',
	succeeded: '成功',
	pending: '待处理',
	running: '执行中',
	validating: '校验中',
	rolled_back: '已回滚',
	completed: '已完成',
};

const PROFILE_STATUS_LABELS: Record<string, string> = {
	draft: '草稿',
	ready: '可启用',
	active: '已启用',
	archived: '已归档',
	invalid: '无效',
};

const EXPERT_DEPTH_LABELS: Record<string, string> = {
	standard: '标准审阅',
	deep: '深度审阅',
	light: '轻量审阅',
};

const PRESET_NAME_LABELS: Record<string, string> = {
	balanced: '均衡总控',
	conservative: '保守审阅',
	deep_review: '深度审阅',
};

function localizeSessionStatus(status?: string) {
	if (!status) return '-';
	return SESSION_STATUS_LABELS[status] ?? status;
}

function localizeProfileStatus(status?: string) {
	if (!status) return '-';
	return PROFILE_STATUS_LABELS[status] ?? localizeSessionStatus(status);
}

function localizePresetName(id: string, fallback?: string) {
	return PRESET_NAME_LABELS[id] ?? fallback ?? id;
}

function localizePresetDepth(value?: string) {
	if (!value) return '-';
	return EXPERT_DEPTH_LABELS[value] ?? value;
}

function localizeKnownText(value?: string) {
	if (!value) return '-';
	return value;
}

function statusTone(status?: string) {
	switch (status) {
		case 'ready':
			return 'border-primary/25 bg-primary/5 text-primary';
		case 'applied':
		case 'succeeded':
			return 'border-emerald-500/25 bg-emerald-500/5 text-emerald-700 dark:text-emerald-300';
		case 'stale':
			return 'border-amber-500/25 bg-amber-500/5 text-amber-700 dark:text-amber-300';
		case 'failed':
			return 'border-destructive/25 bg-destructive/5 text-destructive';
		case 'applying':
		case 'running':
			return 'border-accent/25 bg-accent/5 text-accent-foreground';
		default:
			return 'border-card-border bg-background/60 text-muted-foreground';
	}
}

function StatMiniCard({ title, value, tone = 'default' }: { title: string; value: React.ReactNode; tone?: 'default' | 'emphasis' }) {
	return (
		<div
			className={cn(
				'rounded-xl border px-2.5 py-2',
				tone === 'emphasis' ? 'border-primary/20 bg-primary/5' : 'border-card-border/70 bg-background/55'
				)}
		>
			<div className="text-[11px] text-muted-foreground">{title}</div>
			<div className="mt-0.5 line-clamp-2 break-all text-sm font-semibold text-card-foreground">{value}</div>
		</div>
	);
}

function MutationRow({ type, summary }: { type: string; summary?: string }) {
	const MUTATION_TYPE_LABELS: Record<string, string> = {
		group_upsert: '分组更新',
		group_item_attach: '挂入分组',
		group_item_detach: '移出分组',
		group_item_reorder: '调整顺序',
		route_target_override_upsert: '写入路由目标',
		route_target_override_delete: '删除路由目标',
		llm_price_upsert: '补全价格',
		dynamic_routing_setting_set: '更新动态路由',
		runtime_policy_set: '更新治理策略',
		strategy_profile_activate: '切换策略',
	};

	return (
		<div className="rounded-xl border border-card-border/70 bg-muted/20 px-3 py-2.5">
			<div className="flex flex-wrap items-center gap-2">
				<div className="rounded-full border border-card-border bg-background px-2 py-0.5 text-[11px] text-muted-foreground">{MUTATION_TYPE_LABELS[type] ?? type}</div>
				<div className="break-all text-sm font-medium text-card-foreground">{localizeKnownText(summary)}</div>
			</div>
		</div>
	);
}

function ScrollPanel({ children, testId }: { children: React.ReactNode; testId?: string }) {
	return (
		<div data-testid={testId} className="h-[400px] overflow-y-auto overscroll-contain px-1 py-1">
			{children}
		</div>
	);
}

export const SessionDetail = forwardRef<HTMLDivElement, SessionDetailProps>(function SessionDetail(props, learningRef) {
	const t = useTranslations('aiAutomationV2');
	const {
		tab,
		onTabChange,
		session,
		sessions,
		rollbackPoints,
		strategyProfiles,
		presets,
		learningSummary,
		selectedSessionID,
		onSelectSession,
		newProfileName,
		onNewProfileNameChange,
		onCreateProfile,
		onActivateProfile,
		onRollback,
		isCreatingProfile,
		locale,
	} = props;

	const selectedPreset = presets.find((item) => item.id === session?.expert_preset_id) ?? presets[0];

	return (
		<section className="octo-panel flex min-h-0 flex-col" data-testid="ai-session-detail">
			<div className="flex flex-wrap gap-2 border-b border-card-border px-3 py-2.5">
				{WORKSPACE_TABS.map((item) => {
					const Icon = item.icon;
					const active = tab === item.key;
					return (
						<button
							key={item.key}
							type="button"
							data-testid={`ai-governance-tab-${item.key}`}
							className={cn(
								'inline-flex h-9 shrink-0 items-center gap-2 rounded-full border px-3 text-xs transition',
								active
									? 'border-primary/30 bg-primary/10 text-primary'
									: 'border-card-border bg-card text-muted-foreground hover:border-primary/20 hover:text-card-foreground'
								)}
							onClick={() => onTabChange(item.key)}
						>
							<Icon className="size-4" />
							{t(item.labelKey)}
						</button>
					);
				})}
			</div>

			<div className="min-h-0 flex-1 p-3">
				{tab === 'preview' ? (
					<ScrollPanel testId="ai-governance-workspace-preview">
						<div className="space-y-3">
							{session ? (
								<>
									<div className="grid grid-cols-2 gap-2 min-[480px]:grid-cols-4">
										<StatMiniCard title={t('preview.groups')} value={<AnimatedNumber value={session.preview?.impact_counts?.groups ?? 0} />} />
										<StatMiniCard title={t('preview.items')} value={<AnimatedNumber value={session.preview?.impact_counts?.items ?? 0} />} />
										<StatMiniCard title={t('preview.overrides')} value={<AnimatedNumber value={session.preview?.impact_counts?.overrides ?? 0} />} />
										<StatMiniCard
											title={t('preview.status')}
											value={session.preview?.can_apply ? t('preview.ready') : t('preview.blocked')}
											tone={session.preview?.can_apply ? 'emphasis' : 'default'}
										/>
									</div>
									{session.preview?.summary_lines && session.preview.summary_lines.length > 0 ? (
										<div className="grid gap-2 md:grid-cols-2">
											{session.preview.summary_lines.map((line, index) => (
												<div key={`line-${index}`} className="rounded-xl border border-card-border/70 bg-background/70 px-3 py-2 text-sm text-card-foreground">
													{localizeKnownText(line)}
												</div>
											))}
										</div>
									) : null}
									<div className="space-y-2">
										<div className="text-xs font-medium text-muted-foreground">{t('workspace.detailsTitle')}</div>
										<div className="grid gap-2">
											<div className="rounded-xl border border-card-border/70 bg-muted/20 p-3">
												<div className="text-[11px] text-muted-foreground">{t('sessionDetail.sessionInfo')}</div>
												<div className="mt-1 text-sm font-medium text-card-foreground">#{session.id}</div>
												<div className="mt-1 break-all text-xs leading-5 text-muted-foreground">{localizeKnownText(session.operator_summary || t('states.idleSummary'))}</div>
											</div>
											<div className="rounded-xl border border-card-border/70 bg-muted/20 p-3">
												<div className="text-[11px] text-muted-foreground">{t('sessionDetail.snapshotChecksum')}</div>
												<div className="mt-1 break-all font-mono text-xs leading-6 text-card-foreground">{session.snapshot_checksum || '-'}</div>
											</div>
										</div>
									</div>
									<div className="space-y-2">
										<div className="text-xs font-medium text-muted-foreground">{t('runReport.mutationsTitle')}</div>
										{session.preview?.mutations && session.preview.mutations.length > 0 ? (
											<div className="space-y-2">
												{session.preview.mutations.map((mutation, index) => (
													<MutationRow key={`${mutation.type}-${index}`} type={mutation.type} summary={mutation.summary} />
												))}
											</div>
										) : (
											<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-4 text-sm text-muted-foreground">{t('states.noPreview')}</div>
										)}
									</div>
								</>
							) : (
								<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{t('states.noPreview')}</div>
							)}
						</div>
					</ScrollPanel>
				) : null}

				{tab === 'profiles' ? (
					<ScrollPanel testId="ai-governance-workspace-profiles">
						<div className="space-y-4">
							<div className="flex flex-col gap-3 md:flex-row">
								<Input
									className="h-11 rounded-xl"
									placeholder={t('profiles.namePlaceholder')}
									value={newProfileName}
									onChange={(event) => onNewProfileNameChange(event.target.value)}
								/>
								<Button
									variant="outline"
									className="rounded-xl"
									onClick={onCreateProfile}
									disabled={isCreatingProfile || !session}
								>
									{t('profiles.create')}
								</Button>
							</div>
							<div className="space-y-3">
								{strategyProfiles.length > 0 ? (
									strategyProfiles.map((profile) => (
										<div key={profile.id} className="rounded-2xl border border-card-border/70 bg-muted/20 p-4">
											<div className="flex flex-wrap items-center justify-between gap-3">
												<div className="min-w-0 flex-1">
													<div className="break-all text-sm font-medium text-card-foreground">{localizeKnownText(profile.name)}</div>
													<div className="mt-1 break-all text-xs text-muted-foreground">{localizeKnownText(profile.summary || t('profiles.noSummary'))}</div>
												</div>
												<div className="flex items-center gap-2">
													<div
														className={cn(
															'rounded-full border px-3 py-1 text-[11px]',
															profile.is_active
																? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
																: 'border-card-border bg-background text-muted-foreground'
															)}
													>
														{profile.is_active ? t('profiles.active') : localizeProfileStatus(profile.status)}
													</div>
													<Button variant="outline" className="rounded-xl" onClick={() => onActivateProfile(profile.id)} disabled={profile.is_active}>
														{t('profiles.activate')}
													</Button>
												</div>
											</div>
										</div>
									))
								) : (
									<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{t('profiles.empty')}</div>
								)}
							</div>
						</div>
					</ScrollPanel>
				) : null}

				{tab === 'rollback' ? (
					<ScrollPanel testId="ai-governance-workspace-rollback">
						<div className="space-y-3">
							{rollbackPoints.length > 0 ? (
								rollbackPoints.map((point) => (
									<div key={point.id} className="rounded-2xl border border-card-border/70 bg-muted/20 p-4">
										<div className="flex flex-wrap items-center justify-between gap-3">
											<div className="min-w-0 flex-1">
												<div className="text-sm font-medium text-card-foreground">#{point.id}</div>
												<div className="mt-1 break-all text-xs text-muted-foreground">{localizeKnownText(point.summary)}</div>
												<div className="mt-1 text-[11px] text-muted-foreground">{formatDateTimeByLocale(point.created_at, locale)}</div>
											</div>
											<Button variant="outline" className="rounded-xl" onClick={() => onRollback(point)}>
												{t('actions.rollback')}
											</Button>
										</div>
									</div>
								))
							) : (
								<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{t('states.noRollbackPoints')}</div>
							)}
						</div>
					</ScrollPanel>
				) : null}

				{tab === 'history' ? (
					<ScrollPanel testId="ai-governance-workspace-history">
						<div className="space-y-5">
							<div className="space-y-3">
								{sessions.length > 0 ? (
									sessions.map((item) => (
										<button
											key={item.id}
											type="button"
											className={cn(
												'w-full rounded-2xl border p-4 text-left transition',
												selectedSessionID === item.id
													? 'border-primary/30 bg-primary/10 text-primary'
													: 'border-card-border bg-muted/20 text-card-foreground hover:border-primary/20'
											)}
											onClick={() => onSelectSession(item.id)}
										>
											<div className="flex flex-wrap items-center justify-between gap-2">
												<div className="break-all text-sm font-medium">#{item.id} · {localizeKnownText(item.goal)}</div>
												<div className={cn('rounded-full border px-2.5 py-1 text-[11px]', statusTone(item.status))}>{localizeSessionStatus(item.status)}</div>
											</div>
											<div className="mt-2 break-all text-xs leading-5 text-muted-foreground">{localizeKnownText(item.operator_summary)}</div>
										</button>
									))
								) : (
									<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{t('states.noHistory')}</div>
								)}
							</div>

							<div ref={learningRef} data-ai-focus-target="learning" className="rounded-2xl border border-card-border/70 bg-muted/20 p-4">
								<div className="flex items-center gap-2 text-sm font-semibold text-card-foreground">
									<BrainCircuit className="size-4 text-primary" />
									{t('history.learningTitle')}
								</div>
								<div className="mt-2 text-xs leading-5 text-muted-foreground">{t('history.learningDesc')}</div>
								<div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
									<StatMiniCard title={t('history.learningEnabled')} value={learningSummary?.enabled ? t('sidebar.learningOn') : t('sidebar.learningOff')} />
									<StatMiniCard title={t('history.learningSamples')} value={<><AnimatedNumber value={learningSummary?.sample_count ?? 0} /> 条</>} />
									<StatMiniCard title={t('history.learningTopTarget')} value={learningSummary?.top_target ? localizeKnownText(learningSummary.top_target) : '-'} />
									<StatMiniCard
										title={t('history.learningUpdated')}
										value={learningSummary?.last_sample_at ? formatDateTimeByLocale(String(learningSummary.last_sample_at), locale) : '-'}
									/>
								</div>
							</div>
						</div>
					</ScrollPanel>
				) : null}

				{tab === 'expert' ? (
					<ScrollPanel testId="ai-governance-workspace-expert">
						<div className="space-y-4">
							<div className="grid gap-3 sm:grid-cols-3">
								{presets.map((preset) => (
									<div
										key={preset.id}
										className={cn(
											'rounded-2xl border p-4',
											session?.expert_preset_id === preset.id
												? 'border-primary/30 bg-primary/10 text-primary'
												: 'border-card-border bg-muted/20 text-card-foreground'
										)}
									>
										<div className="text-sm font-medium">{localizePresetName(preset.id, preset.name)}</div>
										<div className="mt-1 text-xs leading-5 text-muted-foreground">{localizeKnownText(preset.description)}</div>
									</div>
								))}
							</div>
							{selectedPreset ? (
								<div className="rounded-2xl border border-card-border/70 bg-muted/20 p-4">
									<div className="text-sm font-semibold text-card-foreground">{localizePresetName(selectedPreset.id, selectedPreset.name)}</div>
									<div className="mt-2 break-all text-sm leading-6 text-muted-foreground">{localizeKnownText(selectedPreset.description)}</div>
									<div className="mt-3 grid gap-3 sm:grid-cols-3">
										<StatMiniCard title={t('sessionDetail.presetReviewDepth')} value={localizePresetDepth(selectedPreset.review_depth)} />
										<StatMiniCard title={t('sessionDetail.presetCleanup')} value={selectedPreset.cleanup_stale ? t('expert.enabled') : t('expert.disabled')} />
										<StatMiniCard title={t('sessionDetail.presetSyncBindings')} value={selectedPreset.sync_bindings ? t('expert.enabled') : t('expert.disabled')} />
									</div>
								</div>
							) : null}
						</div>
					</ScrollPanel>
				) : null}
			</div>
		</section>
	);
});
