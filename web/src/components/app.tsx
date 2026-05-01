'use client';

import { useState, useEffect, useRef, type FormEvent } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { useAuth, useForceChangePassword } from '@/api/endpoints/user';
import { LoginForm } from '@/components/modules/login';
import { APIKeyDashboard } from '@/components/modules/apikey-dashboard';
import { ContentLoader } from '@/route/content-loader';
import { NavBar, useNavStore } from '@/components/modules/navbar';
import { DocModal } from '@/components/modules/navbar/DocModal';
import { useTranslations } from 'next-intl';
import Logo, { LOGO_DRAW_END_MS } from '@/components/modules/logo';
import { Toolbar } from '@/components/modules/toolbar';
import { ENTRANCE_VARIANTS } from '@/lib/animations/fluid-transitions';
import { useQueryClient } from '@tanstack/react-query';
import { CONTENT_MAP } from '@/route';
import { apiClient } from '@/api/client';
import { logger } from '@/lib/logger';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from '@/components/common/Toast';
import { BookOpen } from 'lucide-react';

function timeout(ms: number) {
    return new Promise<void>((resolve) => setTimeout(resolve, ms));
}

export function AppContainer() {
    const { isAuthenticated, isAPIKeyAuth, isLoading: authLoading, mustChangePassword } = useAuth();
    const { activeItem, direction } = useNavStore();
    const t = useTranslations('navbar');
    const tApp = useTranslations('app');
    const queryClient = useQueryClient();
    const forceChangePassword = useForceChangePassword();

    const [logoAnimationComplete, setLogoAnimationComplete] = useState(false);
    const [bootstrapComplete, setBootstrapComplete] = useState(false);
    const bootstrapStartedRef = useRef(false);
    const [forcedUsername, setForcedUsername] = useState('');
    const [forcedPassword, setForcedPassword] = useState('');
    const [forcedPasswordConfirm, setForcedPasswordConfirm] = useState('');
    const [docModalOpen, setDocModalOpen] = useState(false);

    useEffect(() => {
        const el = document.getElementById('initial-loader');
        if (!el) return;

        el.classList.add('octo-hide');
        const timer = setTimeout(() => el.remove(), 220);
        return () => clearTimeout(timer);
    }, []);

    useEffect(() => {
        const timer = setTimeout(() => setLogoAnimationComplete(true), LOGO_DRAW_END_MS);
        return () => clearTimeout(timer);
    }, []);

    useEffect(() => {
        if (authLoading) return;
        if (!isAuthenticated) {
            setBootstrapComplete(true);
            return;
        }
        if (mustChangePassword) {
            setBootstrapComplete(true);
            return;
        }

        if (bootstrapStartedRef.current) return;
        bootstrapStartedRef.current = true;

        let cancelled = false;

        (async () => {
            try {
                const prefetches: Array<Promise<unknown>> = [];

                if (isAPIKeyAuth) {
                    prefetches.push(
                        queryClient.prefetchQuery({
                            queryKey: ['apikey', 'dashboard', 'stats'],
                            queryFn: async () => apiClient.get('/api/v1/apikey/stats'),
                        }),
                    );
                } else {
                    const component = CONTENT_MAP[activeItem];
                    if (component?.preload) {
                        prefetches.push(component.preload());
                    }

                    switch (activeItem) {
                    case 'home':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['stats', 'total'],
                                queryFn: async () => apiClient.get('/api/v1/stats/total'),
                            }),
                        );
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['stats', 'daily'],
                                queryFn: async () => apiClient.get('/api/v1/stats/daily'),
                            }),
                        );
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['stats', 'hourly'],
                                queryFn: async () => apiClient.get('/api/v1/stats/hourly'),
                            }),
                        );
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['channels', 'list'],
                                queryFn: async () => apiClient.get('/api/v1/channel/list'),
                            }),
                        );
                        break;
                    case 'channel':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['channels', 'list'],
                                queryFn: async () => apiClient.get('/api/v1/channel/list'),
                            }),
                        );
                        break;
                    case 'group':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['groups', 'list'],
                                queryFn: async () => apiClient.get('/api/v1/group/list'),
                            }),
                        );
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['models', 'channel'],
                                queryFn: async () => apiClient.get('/api/v1/model/channel'),
                            }),
                        );
                        break;
                    case 'model':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['models', 'list'],
                                queryFn: async () => apiClient.get('/api/v1/model/list'),
                            }),
                        );
                        break;
                    case 'ai':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['ai-automation', 'config'],
                                queryFn: async () => apiClient.get('/api/v1/ai/config'),
                            }),
                        );
                        break;
                    case 'setting':
                        prefetches.push(
                            queryClient.prefetchQuery({
                                queryKey: ['apikeys', 'list'],
                                queryFn: async () => apiClient.get('/api/v1/apikey/list'),
                            }),
                        );
                        break;
                    default:
                        break;
                    }
                }

                await Promise.race([
                    Promise.allSettled(prefetches),
                    timeout(5000),
                ]);
            } catch (e) {
                logger.warn('bootstrap prefetch failed:', e);
            } finally {
                if (!cancelled) setBootstrapComplete(true);
            }
        })();

        return () => {
            cancelled = true;
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [authLoading, isAuthenticated, isAPIKeyAuth, mustChangePassword]);

    const isLoading = authLoading || !logoAnimationComplete || (isAuthenticated && !bootstrapComplete);

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background">
                <Logo size={120} animate />
            </div>
        );
    }

    if (isAPIKeyAuth) {
        return (
            <AnimatePresence mode="wait">
                <APIKeyDashboard key="apikey-dashboard" />
            </AnimatePresence>
        );
    }

    if (!isAuthenticated) {
        return (
            <AnimatePresence mode="wait">
                <LoginForm key="login" />
            </AnimatePresence>
        );
    }

    if (mustChangePassword) {
        return (
            <ForcePasswordGate
                username={forcedUsername}
                onUsernameChange={setForcedUsername}
                password={forcedPassword}
                onPasswordChange={setForcedPassword}
                confirmPassword={forcedPasswordConfirm}
                onConfirmPasswordChange={setForcedPasswordConfirm}
                pending={forceChangePassword.isPending}
                onSubmit={async () => {
                    if (!forcedPassword.trim()) {
                        toast.error(tApp('forcePassword.toast.passwordRequired'));
                        return;
                    }
                    if (forcedPassword.length < 6) {
                        toast.error(tApp('forcePassword.toast.passwordTooShort'));
                        return;
                    }
                    if (forcedPassword !== forcedPasswordConfirm) {
                        toast.error(tApp('forcePassword.toast.passwordMismatch'));
                        return;
                    }

                    try {
                        await forceChangePassword.mutateAsync({
                            newUsername: forcedUsername.trim() || undefined,
                            newPassword: forcedPassword,
                        });
                        setForcedUsername('');
                        setForcedPassword('');
                        setForcedPasswordConfirm('');
                        toast.success(tApp('forcePassword.toast.updated'));
                    } catch (error) {
                        toast.error(tApp('forcePassword.toast.updateFailed'), {
                            description: error instanceof Error ? error.message : undefined,
                        });
                    }
                }}
            />
        );
    }

    return (
        <motion.div
            key="main-app"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
            className="mx-auto flex h-dvh max-w-6xl flex-col overflow-hidden px-3 md:grid md:grid-cols-[auto_1fr] md:gap-6 md:px-6"
        >
            <NavBar />
            <main className="flex min-h-0 w-full min-w-0 flex-1 flex-col">
                <header className="my-6 flex flex-none items-center gap-x-2 px-2">
                    <Logo size={48} />
                    <div className="flex-1 overflow-hidden">
                        <AnimatePresence mode="wait" custom={direction}>
                            <motion.div
                                key={activeItem}
                                custom={direction}
                                variants={{
                                    initial: (dir: number) => ({ y: 32 * dir, opacity: 0 }),
                                    animate: { y: 0, opacity: 1 },
                                    exit: (dir: number) => ({ y: -32 * dir, opacity: 0 }),
                                }}
                                initial="initial"
                                animate="animate"
                                exit="exit"
                                transition={{ duration: 0.3 }}
                                className="flex items-center"
                            >
                                <span className="mt-1 text-3xl font-bold">{t(activeItem)}</span>
                            </motion.div>
                        </AnimatePresence>
                    </div>
                    <div className="ml-auto flex items-center gap-2">
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => setDocModalOpen(true)}
                            className="rounded-xl"
                        >
                            <BookOpen className="size-4" />
                            <span className="hidden md:inline">{t('doc')}</span>
                        </Button>
                        <Toolbar />
                    </div>
                </header>
                <AnimatePresence mode="wait" initial={false}>
                    <motion.div
                        key={activeItem}
                        variants={ENTRANCE_VARIANTS.content}
                        initial="initial"
                        animate="animate"
                        exit={{
                            opacity: 0,
                            scale: 0.98,
                        }}
                        transition={{ duration: 0.25 }}
                        className="h-full min-h-0 flex-1"
                    >
                        <ContentLoader activeRoute={activeItem} />
                    </motion.div>
                </AnimatePresence>
            </main>
            <DocModal isOpen={docModalOpen} onClose={() => setDocModalOpen(false)} />
        </motion.div>
    );
}

function ForcePasswordGate({
    username,
    onUsernameChange,
    password,
    onPasswordChange,
    confirmPassword,
    onConfirmPasswordChange,
    pending,
    onSubmit,
}: {
    username: string;
    onUsernameChange: (value: string) => void;
    password: string;
    onPasswordChange: (value: string) => void;
    confirmPassword: string;
    onConfirmPasswordChange: (value: string) => void;
    pending: boolean;
    onSubmit: () => Promise<void>;
}) {
    const t = useTranslations('app');
    const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        await onSubmit();
    };

    return (
        <motion.div
            key="force-password"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.25 }}
            className="flex min-h-screen items-center justify-center px-5 py-8"
        >
            <div className="w-full max-w-lg rounded-[2rem] border border-border/70 bg-card/95 p-6 shadow-sm sm:p-7">
                <div className="mb-6 flex items-center gap-4">
                    <Logo size={46} />
                    <div>
                        <div className="text-xl font-semibold">{t('forcePassword.title')}</div>
                        <div className="mt-1 text-sm text-muted-foreground">{t('forcePassword.descPrefix')} <span className="font-medium text-foreground">admin / admin</span>{t('forcePassword.descSuffix')}</div>
                    </div>
                </div>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">{t('forcePassword.usernameLabel')}</label>
                        <Input value={username} onChange={(event) => onUsernameChange(event.target.value)} placeholder={t('forcePassword.usernamePlaceholder')} className="h-11 rounded-xl" />
                        <div className="text-xs text-muted-foreground">{t('forcePassword.usernameHintPrefix')} <span className="font-medium text-foreground">admin</span>{t('forcePassword.usernameHintSuffix')}</div>
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">{t('forcePassword.passwordLabel')}</label>
                        <Input type="password" value={password} onChange={(event) => onPasswordChange(event.target.value)} placeholder={t('forcePassword.passwordPlaceholder')} className="h-11 rounded-xl" />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">{t('forcePassword.confirmLabel')}</label>
                        <Input type="password" value={confirmPassword} onChange={(event) => onConfirmPasswordChange(event.target.value)} placeholder={t('forcePassword.confirmPlaceholder')} className="h-11 rounded-xl" />
                    </div>
                    <div className="rounded-2xl border border-amber-500/20 bg-amber-500/8 px-4 py-3 text-xs leading-5 text-muted-foreground">
                        {t('forcePassword.lockedHint')}
                    </div>
                    <Button type="submit" disabled={pending} className="h-11 w-full rounded-xl">
                        {pending ? t('forcePassword.saving') : t('forcePassword.submit')}
                    </Button>
                </form>
            </div>
        </motion.div>
    );
}
