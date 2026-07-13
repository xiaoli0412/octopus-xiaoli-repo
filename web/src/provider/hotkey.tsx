'use client';

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { createHotkeyManager, type HotkeyBinding, type HotkeyManager, type HotkeyScope } from '@/lib/hotkey';
import { useSettingStore } from '@/stores/setting';
import { useNavStore, type NavItem } from '@/components/modules/navbar';
import { HotkeyHelp } from '@/components/common/HotkeyHelp';
import { useTranslations } from 'next-intl';

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

interface HotkeyContextValue {
    register: (binding: HotkeyBinding) => () => void;
    enable: () => void;
    disable: () => void;
    isEnabled: () => boolean;
    getAllBindings: () => HotkeyBinding[];
}

const HotkeyContext = createContext<HotkeyContextValue | null>(null);

export function useHotkeyContext(): HotkeyContextValue {
    const ctx = useContext(HotkeyContext);
    if (!ctx) {
        throw new Error('useHotkeyContext must be used within HotkeyProvider');
    }
    return ctx;
}

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export function HotkeyProvider({ children }: { children: ReactNode }) {
    const [manager] = useState<HotkeyManager>(() => createHotkeyManager());
    const hotkeyEnabled = useSettingStore((s) => s.hotkeyEnabled);

    // 同步 store 中的 hotkeyEnabled 到管理器
    useEffect(() => {
        if (hotkeyEnabled) {
            manager.enable();
        } else {
            manager.disable();
        }
    }, [hotkeyEnabled, manager]);

    // 监听全局 keydown 事件并分发
    useEffect(() => {
        const handler = (event: KeyboardEvent) => manager.handleEvent(event);
        window.addEventListener('keydown', handler);
        return () => window.removeEventListener('keydown', handler);
    }, [manager]);

    const contextValue = useMemo<HotkeyContextValue>(
        () => ({
            register: (binding: HotkeyBinding) => manager.register(binding),
            enable: () => manager.enable(),
            disable: () => manager.disable(),
            isEnabled: () => manager.isEnabled(),
            getAllBindings: () => manager.getAllBindings(),
        }),
        [manager],
    );

    return (
        <HotkeyContext.Provider value={contextValue}>
            {children}
            <GlobalHotkeys manager={manager} />
            <HotkeyHelp />
        </HotkeyContext.Provider>
    );
}

// ---------------------------------------------------------------------------
// 全局快捷键注册（导航序列 + 帮助面板触发）
// ---------------------------------------------------------------------------

const NAV_TARGETS: Array<{ keys: string; navItem: NavItem; labelKey: string }> = [
    { keys: 'g d', navItem: 'home', labelKey: 'hotkey.nav.dashboard' },
    { keys: 'g c', navItem: 'channel', labelKey: 'hotkey.nav.channel' },
    { keys: 'g g', navItem: 'group', labelKey: 'hotkey.nav.group' },
    { keys: 'g m', navItem: 'model', labelKey: 'hotkey.nav.model' },
    { keys: 'g l', navItem: 'log', labelKey: 'hotkey.nav.log' },
    { keys: 'g s', navItem: 'setting', labelKey: 'hotkey.nav.setting' },
];

function GlobalHotkeys({ manager }: { manager: HotkeyManager }) {
    const t = useTranslations();
    const setActiveItem = useNavStore((s) => s.setActiveItem);

    // 使用 ref 保持最新的翻译，避免每次渲染重新注册
    const tRef = useRef(t);
    tRef.current = t;

    useEffect(() => {
        const unregisters: Array<() => void> = [];

        for (const target of NAV_TARGETS) {
            unregisters.push(
                manager.register({
                    id: `global-nav-${target.navItem}`,
                    keys: target.keys,
                    description: tRef.current(target.labelKey),
                    scope: 'global' as HotkeyScope,
                    handler: () => setActiveItem(target.navItem),
                }),
            );
        }

        return () => unregisters.forEach((fn) => fn());
    }, [manager, setActiveItem]);

    return null;
}

// ---------------------------------------------------------------------------
// React Hook：在组件中注册快捷键
// ---------------------------------------------------------------------------

/**
 * 在组件挂载时注册快捷键，卸载时自动注销。
 *
 * 使用 ref 包装 handler，使注册只执行一次但回调始终指向最新闭包。
 * 当绑定的描述/键位/启用状态变化时（如切换语言），自动重新注册。
 * 适用于模块级和组件级快捷键。
 */
export function useHotkeys(bindings: HotkeyBinding[]): void {
    const ctx = useContext(HotkeyContext);
    const bindingsRef = useRef(bindings);
    bindingsRef.current = bindings;

    // 序列化绑定内容（排除 handler 函数）用于依赖追踪
    const depsKey = bindings
        .map((b) => `${b.id}:${b.keys}:${b.description}:${b.scope}:${b.enabled ?? true}`)
        .join('|');

    useEffect(() => {
        if (!ctx) return;
        // 注册稳定 wrapper，调用时从 ref 读取最新 handler
        const wrappers = bindingsRef.current.map((b) => ({
            ...b,
            handler: () => {
                const latest = bindingsRef.current.find((x) => x.id === b.id);
                latest?.handler();
            },
        }));
        const unregisters = wrappers.map((w) => ctx.register(w));
        return () => unregisters.forEach((fn) => fn());
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [ctx, depsKey]);
}

/**
 * 便捷 Hook：注册单个快捷键
 */
export function useHotkey(
    id: string,
    keys: string,
    description: string,
    scope: HotkeyScope,
    handler: () => void,
    enabled?: boolean,
): void {
    const binding = useMemo<HotkeyBinding>(
        () => ({ id, keys, description, scope, handler, enabled }),
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [id, keys, description, scope, enabled],
    );
    const handlerRef = useRef(handler);
    handlerRef.current = handler;

    const stableBinding = useMemo<HotkeyBinding>(
        () => ({ ...binding, handler: () => handlerRef.current() }),
        [binding],
    );

    useHotkeys([stableBinding]);
}

// ---------------------------------------------------------------------------
// 工具栏交互辅助函数（供模块快捷键使用）
// ---------------------------------------------------------------------------

/** 点击工具栏的「新建」按钮 */
export function triggerToolbarCreate(): void {
    const trigger = document.querySelector<HTMLElement>('[data-slot="toolbar-create-trigger"]');
    trigger?.click();
}

/** 聚焦工具栏的搜索输入框（如未展开则先展开） */
export function focusToolbarSearch(): void {
    const input = document.querySelector<HTMLInputElement>('[data-slot="toolbar-search-expanded"] input');
    if (input) {
        input.focus();
        return;
    }
    const trigger = document.querySelector<HTMLElement>('[data-slot="toolbar-search-trigger"]');
    if (trigger) {
        trigger.click();
        // 展开后输入框带有 autoFocus，额外 focus 确保聚焦
        setTimeout(() => {
            const newInput = document.querySelector<HTMLInputElement>('[data-slot="toolbar-search-expanded"] input');
            newInput?.focus();
        }, 60);
    }
}
