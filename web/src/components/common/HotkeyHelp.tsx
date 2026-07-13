'use client';

import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X } from 'lucide-react';
import { useHotkeyContext } from '@/provider/hotkey';
import { formatHotkeyForDisplay, type HotkeyBinding, type HotkeyScope } from '@/lib/hotkey';
import { useTranslations } from 'next-intl';

const SCOPE_ORDER: HotkeyScope[] = ['global', 'module', 'local'];

export function HotkeyHelp() {
    const t = useTranslations();
    const ctx = useHotkeyContext();
    const [open, setOpen] = useState(false);
    const [bindings, setBindings] = useState<HotkeyBinding[]>([]);

    // 注册 `?` 切换帮助面板（全局）
    useEffect(() => {
        const unregister = ctx.register({
            id: 'hotkey-help-toggle',
            keys: '?',
            description: t('hotkey.help.title'),
            scope: 'global',
            handler: () => setOpen((prev) => !prev),
        });
        return () => unregister();
    }, [ctx, t]);

    // 打开时注册 Esc 关闭（local 优先级最高），关闭时自动注销
    useEffect(() => {
        if (!open) return;
        const unregister = ctx.register({
            id: 'hotkey-help-close',
            keys: 'Escape',
            description: t('hotkey.help.close'),
            scope: 'local',
            handler: () => setOpen(false),
        });
        return () => unregister();
    }, [ctx, open, t]);

    // 打开时刷新绑定列表
    useEffect(() => {
        if (open) {
            setBindings(ctx.getAllBindings());
        }
    }, [open, ctx]);

    // 按作用域分组
    const grouped = SCOPE_ORDER.map((scope) => ({
        scope,
        items: bindings.filter((b) => b.scope === scope && b.id !== 'hotkey-help-close'),
    })).filter((g) => g.items.length > 0);

    return (
        <AnimatePresence>
            {open && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="fixed inset-0 z-[100] flex items-center justify-center bg-background/80 backdrop-blur-sm"
                    onClick={() => setOpen(false)}
                    role="dialog"
                    aria-modal="true"
                    aria-label={t('hotkey.help.title')}
                >
                    <motion.div
                        initial={{ scale: 0.95, opacity: 0, y: 10 }}
                        animate={{ scale: 1, opacity: 1, y: 0 }}
                        exit={{ scale: 0.95, opacity: 0, y: 10 }}
                        transition={{ duration: 0.2 }}
                        className="flex max-h-[calc(100vh-4rem)] w-[min(640px,calc(100vw-2rem))] flex-col overflow-hidden rounded-3xl border border-border/60 bg-card p-6 shadow-2xl"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <div className="mb-4 flex shrink-0 items-center justify-between">
                            <h2 className="text-lg font-semibold text-foreground">{t('hotkey.help.title')}</h2>
                            <button
                                type="button"
                                onClick={() => setOpen(false)}
                                className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-primary/20"
                                aria-label={t('hotkey.help.close')}
                            >
                                <X className="size-4" />
                            </button>
                        </div>
                        <div className="min-h-0 flex-1 space-y-5 overflow-auto">
                            {grouped.map(({ scope, items }) => (
                                <div key={scope}>
                                    <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                                        {t(`hotkey.scope.${scope}`)}
                                    </h3>
                                    <div className="space-y-1.5">
                                        {items.map((binding) => (
                                            <div
                                                key={binding.id}
                                                className="flex items-center justify-between gap-3 rounded-lg px-2 py-1.5 hover:bg-muted/30"
                                            >
                                                <span className="text-sm text-foreground/80">
                                                    {binding.description}
                                                </span>
                                                <kbd className="inline-flex shrink-0 items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2 py-0.5 font-mono text-xs text-foreground shadow-sm">
                                                    {formatHotkeyForDisplay(binding.keys)}
                                                </kbd>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}
