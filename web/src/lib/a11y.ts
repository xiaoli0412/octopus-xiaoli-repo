'use client';

import { useEffect, useRef, useState, type RefObject } from 'react';

/**
 * 可访问性（a11y）工具与 Hooks 集合。
 *
 * 提供：
 * - `getFocusableElements` 收集容器内可聚焦元素
 * - `useFocusTrap` 将 Tab 循环限制在容器内
 * - `useFocusReturn` 卸载后将焦点返回触发元素
 * - `useAriaLive` 返回 aria-live 区域 props
 * - `useHiddenBody` 为 body 设置 aria-hidden
 */

const FOCUSABLE_SELECTOR = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
    '[contenteditable="true"]',
    'audio[controls]',
    'video[controls]',
    'details > summary:first-of-type',
]
    .map((s) => `:not([inert]) ${s}`)
    .join(', ');

/**
 * 获取容器内所有可聚焦元素（按 DOM 顺序）。
 * 自动排除不可见元素（display:none / visibility:hidden）。
 */
export function getFocusableElements(container: HTMLElement): HTMLElement[] {
    const nodes = Array.from(
        container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)
    );
    return nodes.filter((el) => {
        if (el.hasAttribute('disabled')) return false;
        if (el.getAttribute('aria-hidden') === 'true') return false;
        const style = window.getComputedStyle(el);
        if (style.display === 'none' || style.visibility === 'hidden') return false;
        return true;
    });
}

/**
 * 焦点陷阱：当 active 为 true 时，Tab/Shift+Tab 循环限制在 ref 容器内，
 * 并自动将焦点移至容器内首个可聚焦元素。
 */
export function useFocusTrap(
    active: boolean,
    ref: RefObject<HTMLElement | null>
): void {
    useEffect(() => {
        if (!active) return;
        const container = ref.current;
        if (!container) return;

        // 将焦点移至容器内首个可聚焦元素
        const focusables = getFocusableElements(container);
        if (focusables.length > 0) {
            focusables[0].focus();
        } else {
            // 容器内无可聚焦元素时，聚焦容器本身
            container.setAttribute('tabindex', '-1');
            container.focus();
        }

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key !== 'Tab') return;
            const currentFocusables = getFocusableElements(container);
            if (currentFocusables.length === 0) {
                event.preventDefault();
                return;
            }
            const first = currentFocusables[0];
            const last = currentFocusables[currentFocusables.length - 1];

            if (event.shiftKey) {
                if (document.activeElement === first || !container.contains(document.activeElement)) {
                    event.preventDefault();
                    last.focus();
                }
            } else {
                if (document.activeElement === last || !container.contains(document.activeElement)) {
                    event.preventDefault();
                    first.focus();
                }
            }
        };

        document.addEventListener('keydown', handleKeyDown);

        return () => {
            document.removeEventListener('keydown', handleKeyDown);
            // 恢复容器 tabindex
            if (container.getAttribute('tabindex') === '-1' && !focusables.length) {
                container.removeAttribute('tabindex');
            }
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [active]);
}

/**
 * 焦点返回：组件卸载或 active 变 false 时，焦点返回触发元素。
 * triggerRef 应指向打开弹层的触发按钮。
 */
export function useFocusReturn(
    triggerRef: RefObject<HTMLElement | null>,
    active?: boolean
): void {
    const activeRef = useRef(active);
    activeRef.current = active;

    useEffect(() => {
        return () => {
            // 卸载时若此前处于激活态，归还焦点
            if (activeRef.current !== false) {
                triggerRef.current?.focus();
            }
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    useEffect(() => {
        if (active === false) {
            triggerRef.current?.focus();
        }
    }, [active, triggerRef]);
}

export type AriaLiveLevel = 'polite' | 'assertive';

export type AriaLiveResult = {
    /** 展开到承载元素上的 aria 属性 */
    props: {
        'aria-live': AriaLiveLevel;
        'aria-atomic': true;
        role: 'status' | 'alert';
    };
    /** 应在承载元素内渲染的文本（消息变化时延迟更新以触发播报） */
    message: string;
};

/**
 * aria-live 区域：消息变化时触发屏幕阅读器播报。
 * 返回 props（绑定到承载元素）与 message（在元素内渲染）。
 *
 * - polite（默认）：非紧急消息，role=status
 * - assertive：紧急消息（如错误），role=alert
 *
 * @example
 * const live = useAriaLive(errorText, 'assertive');
 * <div {...live.props}>{live.message}</div>
 */
export function useAriaLive(
    message: string,
    level: AriaLiveLevel = 'polite'
): AriaLiveResult {
    const [announced, setAnnounced] = useState('');

    useEffect(() => {
        if (message) {
            // 延迟更新，确保屏幕阅读器捕获到文本变化
            const id = window.setTimeout(() => setAnnounced(message), 50);
            return () => window.clearTimeout(id);
        }
        setAnnounced('');
    }, [message]);

    return {
        props: {
            'aria-live': level,
            'aria-atomic': true,
            role: level === 'assertive' ? 'alert' : 'status',
        },
        message: announced,
    };
}

/**
 * 隐藏 body 背景：active 为 true 时给 body 添加 aria-hidden="true"，
 * 避免弹层打开时背景内容被屏幕阅读器读取。
 */
export function useHiddenBody(active: boolean): void {
    useEffect(() => {
        if (!active) return;
        const body = document.body;
        const prev = body.getAttribute('aria-hidden');
        body.setAttribute('aria-hidden', 'true');

        return () => {
            if (prev === null) {
                body.removeAttribute('aria-hidden');
            } else {
                body.setAttribute('aria-hidden', prev);
            }
        };
    }, [active]);
}
