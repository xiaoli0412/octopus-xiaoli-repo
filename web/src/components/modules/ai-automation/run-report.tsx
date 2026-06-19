'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { ChevronDown, ChevronUp, ShieldCheck } from 'lucide-react';

import type { GovernanceSessionDetail } from '@/api/endpoints/ai-automation';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { Button } from '@/components/ui/button';
import { formatDateTimeByLocale } from '@/lib/locale';
import type { Locale as AppLocale } from '@/stores/setting';
import { cn } from '@/lib/utils';

interface RunReportProps {
	session?: GovernanceSessionDetail | null;
	locale: AppLocale;
	onApply: () => void;
	isApplying: boolean;
}

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

function localizeSessionStatus(status?: string) {
	if (!status) return '-';
	return SESSION_STATUS_LABELS[status] ?? status;
}

function statusTone(status?: string) {
	switch (status) {
		case 'ready':
		case 'succeeded':
		case 'applied':
			return 'border-emerald-500/25 bg-emerald-500/5 text-emerald-700 dark:text-emerald-300';
		case 'stale':
			return 'border-amber-500/25 bg-amber-500/5 text-amber-700 dark:text-amber-300';
		case 'failed':
			return 'border-destructive/25 bg-destructive/5 text-destructive';
		case 'applying':
		case 'running':
			return 'border-primary/25 bg-primary/5 text-primary';
		default:
			return 'border-card-border bg-background/60 text-muted-foreground';
	}
}

function localizeKnownText(value?: string) {
	if (!value) return '-';
	return value;
}

function MutationRow({ type, summary }: { type: string; summary?: string }) {
	return (
		<div className="rounded-xl border border-card-border/70 bg-muted/20 px-3 py-2.5">
			<div className="flex flex-wrap items-center gap-2">
				<div className="rounded-full border border-card-border bg-background px-2 py-0.5 text-[11px] text-muted-foreground">
					{MUTATION_TYPE_LABELS[type] ?? type}
				</div>
				<div className="break-all text-sm font-medium text-card-foreground">{localizeKnownText(summary)}</div>
			</div>
		</div>
	);
}

function StatPill({ label, value, tone = 'default' }: { label: string; value: React.ReactNode; tone?: 'default' | 'emphasis' | 'muted' }) {
	return (
		<div
			className={cn(
				'rounded-xl border px-2.5 py-2',
				tone === 'emphasis'
					? 'border-primary/20 bg-primary/5'
					: tone === 'muted'
						? 'border-card-border/60 bg-muted/20'
						: 'border-card-border/70 bg-background/55'
				)}
		>
			<div className="text-[11px] text-muted-foreground">{label}</div>
			<div className="mt-0.5 line-clamp-2 break-all text-sm font-semibold text-card-foreground">{value}</div>
		</div>
	);
}

export function RunReport({ session, locale, onApply, isApplying }: RunReportProps) {
	const t = useTranslations('aiAutomationV2');
	const [expanded, setExpanded] = useState(false);

	if (!session) {
		return (
			<section className="octo-panel p-3" data-testid="ai-run-report-empty">
				<div className="text-sm font-semibold text-card-foreground">{t('runReport.title')}</div>
				<div className="mt-1 text-xs text-muted-foreground">{t('runReport.empty')}</div>
			</section>
		);
	}

	const preview = session.preview;
	const canApply = preview?.can_apply ?? false;

	return (
		<section className="octo-panel p-3" data-testid="ai-run-report">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div className="min-w-0 flex-1">
					<div className="flex flex-wrap items-center gap-2">
						<div className="text-sm font-semibold text-card-foreground">{t('runReport.title')}</div>
						<div
							className={cn(
								'rounded-full border px-2 py-0.5 text-[11px]',
								statusTone(session.status)
							)}
						>
							{localizeSessionStatus(session.status)}
						</div>
					</div>
					<div className="mt-1.5 break-all text-sm leading-6 text-card-foreground">
						{localizeKnownText(preview?.headline || session.operator_summary || t('states.idleSummary'))}
					</div>
					<div className="mt-1 text-xs text-muted-foreground">
						{t('runReport.updatedAt')}: {session.updated_at ? formatDateTimeByLocale(session.updated_at, locale) : '-'}
					</div>
				</div>
				<div className="flex shrink-0 items-center gap-2">
					<Button
						variant="outline"
						size="sm"
						className="rounded-xl"
						onClick={() => setExpanded((value) => !value)}
						data-testid="ai-run-report-toggle"
					>
						{expanded ? t('actions.hideDetails') : t('actions.showDetails')}
						{expanded ? <ChevronUp className="ml-1 size-3.5" /> : <ChevronDown className="ml-1 size-3.5" />}
					</Button>
					<Button
						size="sm"
						className="rounded-xl border-emerald-500/30 bg-emerald-500/10 text-emerald-700 hover:bg-emerald-500/15 dark:text-emerald-300"
						onClick={onApply}
						disabled={!canApply || isApplying}
						data-testid="ai-run-report-apply"
					>
						<ShieldCheck className="mr-1 size-4" />
						{isApplying ? t('actions.applying') : t('actions.apply')}
					</Button>
				</div>
			</div>

			<div className="mt-3 grid grid-cols-2 gap-2 min-[480px]:grid-cols-4">
				<StatPill label={t('preview.groups')} value={<AnimatedNumber value={preview?.impact_counts?.groups ?? 0} />} />
				<StatPill label={t('preview.items')} value={<AnimatedNumber value={preview?.impact_counts?.items ?? 0} />} />
				<StatPill label={t('preview.overrides')} value={<AnimatedNumber value={preview?.impact_counts?.overrides ?? 0} />} />
				<StatPill
					label={t('preview.status')}
					value={canApply ? t('preview.ready') : t('preview.blocked')}
					tone={canApply ? 'emphasis' : 'muted'}
				/>
			</div>

			{expanded ? (
				<div className="mt-3 space-y-3 border-t border-card-border/70 pt-3">
					{preview?.summary_lines && preview.summary_lines.length > 0 ? (
						<div className="space-y-2">
							<div className="text-xs font-medium text-muted-foreground">{t('runReport.summaryLinesTitle')}</div>
							<div className="grid gap-2 md:grid-cols-2">
								{preview.summary_lines.map((line, index) => (
									<div key={`summary-${index}`} className="rounded-xl border border-card-border/70 bg-background/70 px-3 py-2 text-sm text-card-foreground">
										{localizeKnownText(line)}
									</div>
								))}
							</div>
						</div>
					) : null}

					{riskNotes(preview?.risk_notes).length > 0 ? (
						<div className="space-y-2">
							<div className="text-xs font-medium text-muted-foreground">{t('runReport.riskNotesTitle')}</div>
							<div className="grid gap-2 md:grid-cols-2">
								{riskNotes(preview?.risk_notes).map((line, index) => (
									<div key={`risk-${index}`} className="rounded-xl border border-amber-500/20 bg-amber-500/8 px-3 py-2 text-sm text-card-foreground">
										{localizeKnownText(line)}
									</div>
								))}
							</div>
						</div>
					) : null}

					{blockerNotes(preview?.apply_blockers).length > 0 ? (
						<div className="space-y-2">
							<div className="text-xs font-medium text-muted-foreground">{t('runReport.blockersTitle')}</div>
							<div className="grid gap-2 md:grid-cols-2">
								{blockerNotes(preview?.apply_blockers).map((line, index) => (
									<div key={`blocker-${index}`} className="rounded-xl border border-destructive/20 bg-destructive/8 px-3 py-2 text-sm text-card-foreground">
										{localizeKnownText(line)}
									</div>
								))}
							</div>
						</div>
					) : null}

					<div className="space-y-2">
						<div className="text-xs font-medium text-muted-foreground">{t('runReport.mutationsTitle')}</div>
						{preview?.mutations && preview.mutations.length > 0 ? (
							<div className="space-y-2">
								{preview.mutations.map((mutation, index) => (
									<MutationRow key={`${mutation.type}-${index}`} type={mutation.type} summary={mutation.summary} />
								))}
							</div>
						) : (
							<div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-4 text-sm text-muted-foreground">
								{t('runReport.noMutations')}
							</div>
						)}
					</div>
				</div>
			) : null}
		</section>
	);
}

function riskNotes(value?: string[]) {
	return value ?? [];
}

function blockerNotes(value?: string[]) {
	return value ?? [];
}
