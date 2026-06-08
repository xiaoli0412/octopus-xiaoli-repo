'use client';

import { useTranslations } from 'next-intl';
import { Info, Tag, Github, AlertTriangle, Download, Loader2 } from 'lucide-react';
import { APP_VERSION, GITHUB_REPO } from '@/lib/info';
import { useLatestInfo, useNowVersion, useUpdateCore, useUpdateStatus } from '@/api/endpoints/update';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/common/Toast';
import { isOctopusCacheName, isFontCacheName, SW_MESSAGE_TYPE } from '@/lib/sw';
import { formatReleaseTagDisplay, isSameReleaseVersion } from './info-logic';

const WINDOWS_UPDATE_UNSUPPORTED_TEXT = 'self-update is not supported on windows';
function resolveUpdateErrorMessage(error: unknown) {
    if (error instanceof Error) return error.message;
    if (error && typeof error === 'object' && 'message' in error) {
        const message = (error as { message?: unknown }).message;
        return typeof message === 'string' ? message : undefined;
    }
    return undefined;
}

function isUnsupportedUpdateError(error: unknown) {
    if (!error || typeof error !== 'object') return false;

    const code = 'code' in error ? (error as { code?: unknown }).code : undefined;
    if (code === 501) return true;

    const message = resolveUpdateErrorMessage(error);
    return typeof message === 'string'
        && message.toLowerCase().includes(WINDOWS_UPDATE_UNSUPPORTED_TEXT);
}

function getLocalizedUnsupportedReason(reason: string | undefined, fallback: string) {
    const normalizedReason = reason?.trim();
    if (!normalizedReason) return fallback;

    return normalizedReason.toLowerCase().includes(WINDOWS_UPDATE_UNSUPPORTED_TEXT)
        ? fallback
        : normalizedReason;
}
export function SettingInfo() {
    const t = useTranslations('setting');
    const latestInfoQuery = useLatestInfo();
    const nowVersionQuery = useNowVersion();
    const updateStatusQuery = useUpdateStatus();
    const updateCore = useUpdateCore();

    const backendNowVersion = nowVersionQuery.data || '';
    const latestVersion = latestInfoQuery.data?.tag_name || '';
    const latestVersionLabel = formatReleaseTagDisplay(latestVersion);
    const updateStatus = updateStatusQuery.data;

    // 前端版本与后端当前版本不一致 → 浏览器缓存问题
    const isCacheMismatch = !!backendNowVersion && !isSameReleaseVersion(backendNowVersion, APP_VERSION);
    // 最新版本与后端当前版本不一致 → 有新版本可更新
    const hasNewVersion = !!latestVersion && !!backendNowVersion && !isSameReleaseVersion(latestVersion, backendNowVersion);
    const hasResolvedUpdateCapability = updateStatusQuery.isSuccess;
    const selfUpdateSupported = updateStatus?.self_update_supported ?? false;
    const updateUnsupportedDescription = getLocalizedUnsupportedReason(
        updateStatus?.self_update_unsupported_reason,
        t('info.updateUnsupportedHint')
    );

    const clearCacheAndReload = async () => {
        // 通知 Service Worker 清理缓存
        if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
            navigator.serviceWorker.controller.postMessage({ type: SW_MESSAGE_TYPE.CLEAR_CACHE });
        }
        // 同时也从主线程清理（双保险），但保留字体缓存
        if ('caches' in window) {
            const names = await caches.keys();
            await Promise.all(
                names
                    .filter((name) => isOctopusCacheName(name) && !isFontCacheName(name))
                    .map((name) => caches.delete(name))
            );
        }
        // 注销当前 SW，下次加载会重新注册
        if ('serviceWorker' in navigator) {
            const registrations = await navigator.serviceWorker.getRegistrations();
            await Promise.all(registrations.map((reg) => reg.unregister()));
        }
        // 强制刷新（跳过缓存）
        window.location.reload();
    };

    const handleForceRefresh = () => {
        clearCacheAndReload();
    };

    const handleUpdate = () => {
        if (!hasResolvedUpdateCapability) {
            return;
        }
        if (!selfUpdateSupported) {
            toast.error(t('info.updateUnsupported'), { description: updateUnsupportedDescription });
            return;
        }
        updateCore.mutate(undefined, {
            onSuccess: () => {
                toast.success(t('info.updateSuccess'));
                // 更新成功后清理缓存并刷新
                setTimeout(() => {
                    clearCacheAndReload();
                }, 1500);
            },
            onError: (error) => {
                if (isUnsupportedUpdateError(error)) {
                    toast.error(t('info.updateUnsupported'), { description: updateUnsupportedDescription });
                    return;
                }
                const description = resolveUpdateErrorMessage(error);
                toast.error(t('info.updateFailed'), description ? { description } : undefined);
            }
        });
    };

    return (
        <div className="octo-setting-card">
            <h2 className="octo-setting-heading">
                <Info className="size-4" />
                {t('info.title')}
            </h2>
            {/* GitHub 仓库 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Github className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('info.github')}</span>
                </div>
                <a
                    href={GITHUB_REPO}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="min-w-0 truncate text-sm text-primary hover:underline md:text-right"
                >
                    {GITHUB_REPO.replace('https://github.com/', '')}
                </a>
            </div>
            {/* 当前版本 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Tag className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('info.currentVersion')}</span>
                </div>
                <div className="flex min-w-0 items-center gap-2 md:justify-end">
                    {nowVersionQuery.isLoading ? (
                        <Loader2 className="size-4 animate-spin text-muted-foreground" />
                    ) : (
                        <code className="text-sm font-mono text-muted-foreground">
                            {backendNowVersion || t('info.unknown')}
                        </code>
                    )}
                </div>
            </div>

            {/* 最新版本 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Download className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('info.latestVersion')}</span>
                </div>
                <div className="flex min-w-0 items-center gap-2 md:justify-end">
                            {latestInfoQuery.isLoading ? (
                        <Loader2 className="size-4 animate-spin text-muted-foreground" />
                    ) : (
                        <code className="text-sm font-mono text-muted-foreground">
                            {latestVersionLabel || t('info.unknown')}
                        </code>
                    )}
                </div>
            </div>

            {/* 浏览器缓存问题警告 */}
            {isCacheMismatch && (
                <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-xl space-y-2">
                    <div className="flex items-start gap-3">
                        <AlertTriangle className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
                        <div className="flex-1 space-y-1">
                            <p className="text-sm text-destructive font-medium">
                                {t('info.versionMismatch')}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                {t('info.versionMismatchHint', { frontend: APP_VERSION, backend: backendNowVersion })}
                            </p>
                        </div>
                    </div>
                    <div className="flex justify-end">
                        <Button
                            variant="destructive"
                            size="sm"
                            onClick={handleForceRefresh}
                            className="rounded-xl"
                        >
                            {t('info.forceRefresh')}
                        </Button>
                    </div>
                </div>
            )}

            {/* 有新版本可更新 */}
            {hasNewVersion && hasResolvedUpdateCapability && selfUpdateSupported && (
                <div className="p-3 bg-primary/10 border border-primary/20 rounded-xl space-y-2">
                    <div className="flex items-start gap-3">
                        <Download className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                        <div className="flex-1 space-y-1">
                            <p className="text-sm text-primary font-medium">
                                {t('info.newVersionAvailable')}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                {t('info.newVersionAvailableHint')}
                            </p>
                        </div>
                    </div>
                    <div className="flex justify-end">
                        <Button
                            variant="default"
                            size="sm"
                            onClick={handleUpdate}
                            disabled={updateCore.isPending || !hasResolvedUpdateCapability}
                            className="rounded-xl"
                        >
                            {updateCore.isPending ? t('info.updating') : t('info.updateNow')}
                        </Button>
                    </div>
                </div>
            )}

            {hasNewVersion && hasResolvedUpdateCapability && !selfUpdateSupported && (
                <div className="p-3 bg-muted/50 border border-border rounded-xl space-y-2">
                    <div className="flex items-start gap-3">
                        <AlertTriangle className="h-5 w-5 text-muted-foreground shrink-0 mt-0.5" />
                        <div className="flex-1 space-y-1">
                            <p className="text-sm text-card-foreground font-medium">
                                {t('info.updateUnsupported')}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                {updateUnsupportedDescription}
                            </p>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

