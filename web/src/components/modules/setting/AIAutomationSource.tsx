'use client';

import { Bot, ExternalLink } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useAIGovernanceOverview, useStrategyProfiles } from '@/api/endpoints/ai-automation';
import { Button } from '@/components/ui/button';
import { useNavStore } from '@/components/modules/navbar';

export function SettingAIAutomationSource() {
	const t = useTranslations('setting');
	const overviewQuery = useAIGovernanceOverview();
	const profilesQuery = useStrategyProfiles();
	const setActiveItem = useNavStore((state) => state.setActiveItem);

	const overview = overviewQuery.data;
	const strategyProfiles = profilesQuery.data ?? [];
	const openAICenter = () => setActiveItem('ai');

	const modeLabel = overview?.execution_source.mode === 'manual' ? '手动配置' : overview?.execution_source.mode === 'ai_profile' ? 'AI 策略方案' : '-';
	const sessionStatusLabelMap: Record<string, string> = {
		draft: '草稿',
		planning: '规划中',
		ready: '可应用',
		stale: '需重算',
		applying: '应用中',
		applied: '已应用',
		failed: '失败',
	};
	const recentStatusLabel = overview?.recent_session?.status ? (sessionStatusLabelMap[overview.recent_session.status] ?? overview.recent_session.status) : '-';

	return (
		<div className="octo-setting-card" data-testid="setting-ai-governance-source">
			<div className="octo-toolbar">
				<div className="min-w-0 flex-1">
					<h3 className="octo-setting-heading">
						<Bot className="size-4" />
						{t('aiAutomationSource.title')}
					</h3>
				</div>
				<Button type="button" variant="outline" className="h-10 w-full rounded-xl sm:w-auto sm:min-w-40" onClick={openAICenter}>
					<ExternalLink className="size-4" />
					{t('aiAutomationSource.openCenter')}
				</Button>
			</div>

			<div className="rounded-2xl border border-border/70 bg-muted/20 p-3">
				<div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
					<div className="octo-stat-card text-xs">
						<div className="text-muted-foreground">{t('aiAutomationSource.mode')}</div>
						<div className="mt-1 font-medium text-card-foreground break-all">{modeLabel}</div>
					</div>
					<div className="octo-stat-card text-xs">
						<div className="text-muted-foreground">{t('aiAutomationSource.profileLabel')}</div>
						<div className="mt-1 font-medium text-card-foreground">{overview?.active_strategy_profile?.name ?? t('aiAutomationSource.noActiveProfile')}</div>
					</div>
					<div className="octo-stat-card text-xs">
						<div className="text-muted-foreground">{t('aiAutomationSource.profileConfidence')}</div>
						<div className="mt-1 font-medium text-card-foreground">{overview?.learning ? `样本 ${overview.learning.sample_count}` : '样本 0'}</div>
					</div>
					<div className="octo-stat-card text-xs">
						<div className="text-muted-foreground">{t('aiAutomationSource.status')}</div>
						<div className="mt-1 font-medium text-card-foreground">{recentStatusLabel}</div>
					</div>
				</div>
				<div className="mt-2 line-clamp-2 rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs leading-5 text-muted-foreground">
					<span>{overview?.recent_session?.operator_summary || t('aiAutomationSource.profileSummaryFallback')}</span>
					<span> · {t('aiAutomationSource.profileEmpty')} {strategyProfiles.length > 0 ? `${strategyProfiles.length}` : '0'}</span>
					<span className="sr-only">{t('aiAutomationSource.manualSafety')}</span>
				</div>
			</div>
		</div>
	);
}
