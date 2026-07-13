import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createHotkeyManager, formatHotkeyForDisplay } from './hotkey';

function createKeydownEvent(
    key: string,
    options: { ctrlKey?: boolean; shiftKey?: boolean; altKey?: boolean; metaKey?: boolean } = {},
): KeyboardEvent {
    return new KeyboardEvent('keydown', {
        key,
        ctrlKey: options.ctrlKey ?? false,
        shiftKey: options.shiftKey ?? false,
        altKey: options.altKey ?? false,
        metaKey: options.metaKey ?? false,
        bubbles: true,
        cancelable: true,
    });
}

describe('HotkeyManager', () => {
    let manager: ReturnType<typeof createHotkeyManager>;
    const inputs: HTMLInputElement[] = [];

    beforeEach(() => {
        manager = createHotkeyManager(document);
        (document.activeElement as HTMLElement | null)?.blur?.();
    });

    afterEach(() => {
        for (const input of inputs) {
            input.remove();
        }
        inputs.length = 0;
    });

    function focusInput(): HTMLInputElement {
        const input = document.createElement('input');
        document.body.appendChild(input);
        input.focus();
        inputs.push(input);
        return input;
    }

    // -------------------------------------------------------------------------
    // 注册与注销
    // -------------------------------------------------------------------------

    it('registers and triggers a single key handler', () => {
        const handler = vi.fn();
        manager.register({
            id: 'test-single',
            keys: 'n',
            description: 'New',
            scope: 'global',
            handler,
        });

        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('unregisters a binding via the returned function', () => {
        const handler = vi.fn();
        const unregister = manager.register({
            id: 'test-unregister',
            keys: 'n',
            description: 'New',
            scope: 'global',
            handler,
        });

        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).toHaveBeenCalledTimes(1);

        unregister();

        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('re-registers and overwrites when using the same id', () => {
        const first = vi.fn();
        const second = vi.fn();

        manager.register({ id: 'test-overwrite', keys: 'n', description: 'First', scope: 'global', handler: first });
        manager.register({ id: 'test-overwrite', keys: 'n', description: 'Second', scope: 'global', handler: second });

        manager.handleEvent(createKeydownEvent('n'));
        expect(second).toHaveBeenCalledTimes(1);
        expect(first).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // 作用域优先级
    // -------------------------------------------------------------------------

    it('respects scope priority: local > module > global', () => {
        const globalHandler = vi.fn();
        const moduleHandler = vi.fn();
        const localHandler = vi.fn();

        manager.register({ id: 'test-prio-global', keys: 'k', description: 'G', scope: 'global', handler: globalHandler });
        manager.register({ id: 'test-prio-module', keys: 'k', description: 'M', scope: 'module', handler: moduleHandler });
        manager.register({ id: 'test-prio-local', keys: 'k', description: 'L', scope: 'local', handler: localHandler });

        manager.handleEvent(createKeydownEvent('k'));
        expect(localHandler).toHaveBeenCalledTimes(1);
        expect(moduleHandler).not.toHaveBeenCalled();
        expect(globalHandler).not.toHaveBeenCalled();
    });

    it('later registration wins within the same scope', () => {
        const first = vi.fn();
        const second = vi.fn();

        manager.register({ id: 'test-same-scope-1', keys: 'k', description: 'First', scope: 'global', handler: first });
        manager.register({ id: 'test-same-scope-2', keys: 'k', description: 'Second', scope: 'global', handler: second });

        manager.handleEvent(createKeydownEvent('k'));
        expect(second).toHaveBeenCalledTimes(1);
        expect(first).not.toHaveBeenCalled();
    });

    it('higher scope priority wins even if registered first', () => {
        const globalHandler = vi.fn();
        const localHandler = vi.fn();

        manager.register({ id: 'test-prio-order-g', keys: 'k', description: 'G', scope: 'global', handler: globalHandler });
        manager.register({ id: 'test-prio-order-l', keys: 'k', description: 'L', scope: 'local', handler: localHandler });

        manager.handleEvent(createKeydownEvent('k'));
        expect(localHandler).toHaveBeenCalledTimes(1);
        expect(globalHandler).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // 组合键解析
    // -------------------------------------------------------------------------

    it('triggers a combo key (Ctrl+K)', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-combo-ctrl', keys: 'Ctrl+K', description: 'Command', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('k', { ctrlKey: true }));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('triggers Cmd+Shift+P combo', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-combo-cmd-shift', keys: 'Cmd+Shift+P', description: 'Palette', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('p', { metaKey: true, shiftKey: true }));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('does not trigger combo with wrong modifiers', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-combo-wrong', keys: 'Ctrl+K', description: 'Command', scope: 'global', handler });

        // Cmd+K instead of Ctrl+K
        manager.handleEvent(createKeydownEvent('k', { metaKey: true }));
        expect(handler).not.toHaveBeenCalled();
    });

    it('does not trigger combo without modifiers', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-combo-no-mod', keys: 'Ctrl+K', description: 'Command', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('k'));
        expect(handler).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // 序列键解析
    // -------------------------------------------------------------------------

    it('triggers a sequence key (g d)', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-seq-gd', keys: 'g d', description: 'Dashboard', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('g'));
        manager.handleEvent(createKeydownEvent('d'));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('does not trigger sequence with wrong second key (g x)', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-seq-gx', keys: 'g d', description: 'Dashboard', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('g'));
        manager.handleEvent(createKeydownEvent('x'));
        expect(handler).not.toHaveBeenCalled();
    });

    it('triggers multiple different sequences sharing the same prefix', () => {
        const dashHandler = vi.fn();
        const channelHandler = vi.fn();

        manager.register({ id: 'test-seq-dash', keys: 'g d', description: 'Dashboard', scope: 'global', handler: dashHandler });
        manager.register({ id: 'test-seq-chan', keys: 'g c', description: 'Channel', scope: 'global', handler: channelHandler });

        manager.handleEvent(createKeydownEvent('g'));
        manager.handleEvent(createKeydownEvent('d'));
        expect(dashHandler).toHaveBeenCalledTimes(1);
        expect(channelHandler).not.toHaveBeenCalled();

        manager.handleEvent(createKeydownEvent('g'));
        manager.handleEvent(createKeydownEvent('c'));
        expect(channelHandler).toHaveBeenCalledTimes(1);
    });

    // -------------------------------------------------------------------------
    // 输入聚焦时不响应单键
    // -------------------------------------------------------------------------

    it('does not respond to single keys when input is focused', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-input-single', keys: 'n', description: 'New', scope: 'global', handler });

        focusInput();

        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).not.toHaveBeenCalled();
    });

    it('still responds to combo keys when input is focused', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-input-combo', keys: 'Ctrl+K', description: 'Command', scope: 'global', handler });

        focusInput();

        manager.handleEvent(createKeydownEvent('k', { ctrlKey: true }));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('does not respond to sequence keys when input is focused', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-input-seq', keys: 'g d', description: 'Dashboard', scope: 'global', handler });

        focusInput();

        manager.handleEvent(createKeydownEvent('g'));
        manager.handleEvent(createKeydownEvent('d'));
        expect(handler).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // enable / disable 开关
    // -------------------------------------------------------------------------

    it('does not respond when disabled', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-disabled', keys: 'n', description: 'New', scope: 'global', handler });

        manager.disable();
        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).not.toHaveBeenCalled();

        manager.enable();
        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('isEnabled reflects the current state', () => {
        expect(manager.isEnabled()).toBe(true);
        manager.disable();
        expect(manager.isEnabled()).toBe(false);
        manager.enable();
        expect(manager.isEnabled()).toBe(true);
    });

    // -------------------------------------------------------------------------
    // getAllBindings
    // -------------------------------------------------------------------------

    it('getAllBindings returns all registered bindings', () => {
        manager.register({ id: 'test-getall-1', keys: 'n', description: 'New', scope: 'global', handler: () => {} });
        manager.register({ id: 'test-getall-2', keys: 'Ctrl+K', description: 'Command', scope: 'module', handler: () => {} });

        const bindings = manager.getAllBindings();
        expect(bindings).toHaveLength(2);
        const ids = bindings.map((b) => b.id);
        expect(ids).toContain('test-getall-1');
        expect(ids).toContain('test-getall-2');
    });

    it('getAllBindings excludes unregistered bindings', () => {
        const unregister = manager.register({ id: 'test-getall-remove', keys: 'n', description: 'New', scope: 'global', handler: () => {} });
        expect(manager.getAllBindings()).toHaveLength(1);

        unregister();
        expect(manager.getAllBindings()).toHaveLength(0);
    });

    // -------------------------------------------------------------------------
    // enabled 字段
    // -------------------------------------------------------------------------

    it('respects enabled=false on binding', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-enabled-field', keys: 'n', description: 'New', scope: 'global', handler, enabled: false });

        manager.handleEvent(createKeydownEvent('n'));
        expect(handler).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // 特殊键
    // -------------------------------------------------------------------------

    it('triggers ? key handler (shift+/ on US keyboard)', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-question', keys: '?', description: 'Help', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('?', { shiftKey: true }));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('triggers / key handler', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-slash', keys: '/', description: 'Search', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('/'));
        expect(handler).toHaveBeenCalledTimes(1);
    });

    it('ignores pure modifier key presses', () => {
        const handler = vi.fn();
        manager.register({ id: 'test-modifier', keys: 'Shift', description: 'Shift', scope: 'global', handler });

        manager.handleEvent(createKeydownEvent('Shift'));
        expect(handler).not.toHaveBeenCalled();
    });

    // -------------------------------------------------------------------------
    // formatHotkeyForDisplay
    // -------------------------------------------------------------------------

    it('formatHotkeyForDisplay renders modifier symbols', () => {
        expect(formatHotkeyForDisplay('Ctrl+K')).toBe('Ctrl+K');
        expect(formatHotkeyForDisplay('Cmd+Shift+P')).toBe('⌘+Shift+P');
        expect(formatHotkeyForDisplay('g d')).toBe('g d');
        expect(formatHotkeyForDisplay('?')).toBe('?');
    });
});
