/**
 * 键盘快捷键注册器
 *
 * 支持三种快捷键格式：
 * - 单键：`?`, `/`, `n`
 * - 组合键：`Ctrl+K`, `Cmd+Shift+P`
 * - 序列键：`g d`（按 g 后按 d）
 *
 * 支持三种作用域与优先级：
 * - global  全局生效（优先级最低）
 * - module  当前模块生效
 * - local   当前组件生效（优先级最高）
 *
 * 冲突时优先级：local > module > global
 * 同作用域内后注册的覆盖先注册的。
 */

export type HotkeyScope = 'global' | 'module' | 'local';

export interface HotkeyBinding {
    /** 全局唯一标识，同 id 注册会覆盖旧的 */
    id: string;
    /** 快捷键定义，如 `Ctrl+K`、`?`、`g d` */
    keys: string;
    /** 用户可读描述，用于帮助面板 */
    description: string;
    /** 作用域 */
    scope: HotkeyScope;
    /** 触发回调 */
    handler: () => void;
    /** 是否启用，默认 true */
    enabled?: boolean;
}

interface NormalizedStroke {
    key: string;
    ctrl: boolean;
    shift: boolean;
    alt: boolean;
    meta: boolean;
}

interface ParsedBinding {
    type: 'single' | 'combo' | 'sequence';
    strokes: NormalizedStroke[];
}

interface InternalBinding extends HotkeyBinding {
    parsed: ParsedBinding;
    seq: number;
}

export interface HotkeyManager {
    /** 注册一个快捷键绑定，返回注销函数 */
    register(binding: HotkeyBinding): () => void;
    /** 启用快捷键系统 */
    enable(): void;
    /** 禁用快捷键系统 */
    disable(): void;
    /** 当前是否启用 */
    isEnabled(): boolean;
    /** 获取所有已注册绑定（用于帮助面板） */
    getAllBindings(): HotkeyBinding[];
    /** 处理原生 keydown 事件 */
    handleEvent(event: KeyboardEvent): void;
}

const SCOPE_PRIORITY: Record<HotkeyScope, number> = {
    local: 3,
    module: 2,
    global: 1,
};

/** 序列键续按超时（毫秒） */
const SEQUENCE_TIMEOUT_MS = 500;

// ---------------------------------------------------------------------------
// 解析工具
// ---------------------------------------------------------------------------

function normalizeKey(key: string): string {
    return key.trim().toLowerCase();
}

function parseModifier(mod: string): Partial<Pick<NormalizedStroke, 'ctrl' | 'shift' | 'alt' | 'meta'>> | null {
    const m = mod.trim().toLowerCase();
    switch (m) {
        case 'ctrl':
        case 'control':
            return { ctrl: true };
        case 'cmd':
        case 'meta':
        case 'command':
            return { meta: true };
        case 'shift':
            return { shift: true };
        case 'alt':
        case 'option':
        case 'opt':
            return { alt: true };
        default:
            return null;
    }
}

function parseStroke(strokeStr: string): NormalizedStroke {
    const parts = strokeStr.split('+').map((p) => p.trim()).filter(Boolean);
    const result: NormalizedStroke = { key: '', ctrl: false, shift: false, alt: false, meta: false };

    if (parts.length === 0) return result;

    // 最后一个部分是主键，前面都是修饰键
    result.key = normalizeKey(parts[parts.length - 1]);

    for (let i = 0; i < parts.length - 1; i++) {
        const mod = parseModifier(parts[i]);
        if (mod) {
            if (mod.ctrl) result.ctrl = true;
            if (mod.shift) result.shift = true;
            if (mod.alt) result.alt = true;
            if (mod.meta) result.meta = true;
        }
    }

    return result;
}

function parseBinding(keys: string): ParsedBinding {
    const trimmed = keys.trim();
    const parts = trimmed.split(/\s+/).filter(Boolean);

    if (parts.length > 1) {
        return {
            type: 'sequence',
            strokes: parts.map(parseStroke),
        };
    }

    const stroke = parseStroke(parts[0] || trimmed);
    const isCombo = stroke.ctrl || stroke.shift || stroke.alt || stroke.meta;
    return {
        type: isCombo ? 'combo' : 'single',
        strokes: [stroke],
    };
}

function eventToStroke(event: { key: string; ctrlKey: boolean; shiftKey: boolean; altKey: boolean; metaKey: boolean }): NormalizedStroke {
    return {
        key: normalizeKey(event.key),
        ctrl: event.ctrlKey,
        shift: event.shiftKey,
        alt: event.altKey,
        meta: event.metaKey,
    };
}

/**
 * 单键匹配：只比较 key 和 ctrl/meta 修饰。
 * 不检查 shift/alt，因为 `?` 等符号本身就需要 shift 才能输入。
 */
function singleMatch(binding: NormalizedStroke, event: NormalizedStroke): boolean {
    return (
        binding.key === event.key &&
        !event.ctrl &&
        !event.meta &&
        !binding.ctrl &&
        !binding.meta
    );
}

/** 组合键匹配：严格比较所有修饰键 */
function comboMatch(binding: NormalizedStroke, event: NormalizedStroke): boolean {
    return (
        binding.key === event.key &&
        binding.ctrl === event.ctrl &&
        binding.shift === event.shift &&
        binding.alt === event.alt &&
        binding.meta === event.meta
    );
}

// ---------------------------------------------------------------------------
// 输入聚焦检测
// ---------------------------------------------------------------------------

function isInputFocused(doc: Document): boolean {
    const el = doc.activeElement;
    if (!el) return false;
    const tag = el.tagName.toLowerCase();
    if (tag === 'input' || tag === 'textarea' || tag === 'select') return true;
    if (el instanceof HTMLElement && el.isContentEditable) return true;
    return false;
}

// ---------------------------------------------------------------------------
// 管理器工厂
// ---------------------------------------------------------------------------

export function createHotkeyManager(doc: Document = document): HotkeyManager {
    const bindings: InternalBinding[] = [];
    let enabled = true;
    let seqCounter = 0;
    let sequenceBuffer: NormalizedStroke[] = [];
    let sequenceTimer: ReturnType<typeof setTimeout> | null = null;

    function clearSequenceBuffer(): void {
        sequenceBuffer = [];
        if (sequenceTimer !== null) {
            clearTimeout(sequenceTimer);
            sequenceTimer = null;
        }
    }

    function resetSequenceTimer(): void {
        if (sequenceTimer !== null) {
            clearTimeout(sequenceTimer);
        }
        sequenceTimer = setTimeout(() => {
            sequenceBuffer = [];
            sequenceTimer = null;
        }, SEQUENCE_TIMEOUT_MS);
    }

    /**
     * 判断给定的 stroke 是否是某个已注册序列的前缀（第一个键）
     */
    function isSequencePrefix(stroke: NormalizedStroke): boolean {
        return bindings.some(
            (b) =>
                b.parsed.type === 'sequence' &&
                b.parsed.strokes.length > 0 &&
                singleMatch(b.parsed.strokes[0], stroke),
        );
    }

    /**
     * 检查 buffer 是否是某个序列的有效前缀
     */
    function bufferIsPrefix(buffer: NormalizedStroke[]): boolean {
        return bindings.some((b) => {
            if (b.parsed.type !== 'sequence') return false;
            if (buffer.length >= b.parsed.strokes.length) return false;
            for (let i = 0; i < buffer.length; i++) {
                if (!singleMatch(b.parsed.strokes[i], buffer[i])) return false;
            }
            return true;
        });
    }

    /**
     * 查找与当前 buffer 完全匹配的序列绑定（按优先级排序）
     */
    function findSequenceMatch(buffer: NormalizedStroke[]): InternalBinding | null {
        const matches = bindings.filter((b) => {
            if (b.parsed.type !== 'sequence') return false;
            if (b.parsed.strokes.length !== buffer.length) return false;
            return b.parsed.strokes.every((s, i) => singleMatch(s, buffer[i]));
        });
        return pickBest(matches);
    }

    /**
     * 查找与 stroke 匹配的单键 / 组合键绑定（按优先级排序）
     */
    function findImmediateMatch(stroke: NormalizedStroke, inputActive: boolean): InternalBinding | null {
        const matches = bindings.filter((b) => {
            if (b.parsed.type === 'sequence') return false;
            if (b.parsed.type === 'combo') {
                return comboMatch(b.parsed.strokes[0], stroke);
            }
            // single
            if (inputActive) return false; // 输入聚焦时不响应单键
            return singleMatch(b.parsed.strokes[0], stroke);
        });
        return pickBest(matches);
    }

    /** 在匹配列表中按优先级 + 注册顺序选出最佳绑定 */
    function pickBest(list: InternalBinding[]): InternalBinding | null {
        const filtered = list.filter((b) => b.enabled !== false);
        if (filtered.length === 0) return null;
        let best = filtered[0];
        for (let i = 1; i < filtered.length; i++) {
            const cur = filtered[i];
            const curPri = SCOPE_PRIORITY[cur.scope];
            const bestPri = SCOPE_PRIORITY[best.scope];
            if (curPri > bestPri || (curPri === bestPri && cur.seq > best.seq)) {
                best = cur;
            }
        }
        return best;
    }

    function handleEvent(event: KeyboardEvent): void {
        if (!enabled) return;

        const stroke = eventToStroke(event);
        // 忽略纯修饰键按下（如单独按 Shift）
        if (!stroke.key || stroke.key === 'shift' || stroke.key === 'control' || stroke.key === 'alt' || stroke.key === 'meta') {
            return;
        }

        const inputActive = isInputFocused(doc);
        const isCombo = stroke.ctrl || stroke.meta;

        // 输入聚焦时，仅响应组合键（Ctrl/Cmd 开头），不响应单键与序列键
        if (inputActive && !isCombo) {
            clearSequenceBuffer();
            return;
        }

        // 序列键处理
        if (sequenceBuffer.length > 0) {
            // 已在序列中，追加当前键
            const newBuffer = [...sequenceBuffer, stroke];
            const fullMatch = findSequenceMatch(newBuffer);

            if (fullMatch) {
                clearSequenceBuffer();
                event.preventDefault();
                fullMatch.handler();
                return;
            }

            if (bufferIsPrefix(newBuffer)) {
                // 仍然是某个序列的前缀，继续等待
                sequenceBuffer = newBuffer;
                resetSequenceTimer();
                return;
            }

            // 不是前缀也不是完整匹配，清空 buffer，继续按单键/组合键处理当前 stroke
            clearSequenceBuffer();
        }

        // 检查是否是序列的起始键
        if (!inputActive && isSequencePrefix(stroke)) {
            sequenceBuffer = [stroke];
            resetSequenceTimer();
            // 注意：不在此处触发单键，等待序列完成或超时
            // 如果存在同 key 的单键绑定，序列超时后不会自动触发它（标准行为）
            return;
        }

        // 单键 / 组合键匹配
        const match = findImmediateMatch(stroke, inputActive);
        if (match) {
            event.preventDefault();
            match.handler();
        }
    }

    function register(binding: HotkeyBinding): () => void {
        const internal: InternalBinding = {
            ...binding,
            parsed: parseBinding(binding.keys),
            seq: seqCounter++,
        };

        // 同 id 覆盖
        const existingIdx = bindings.findIndex((b) => b.id === binding.id);
        if (existingIdx >= 0) {
            bindings[existingIdx] = internal;
        } else {
            bindings.push(internal);
        }

        return () => {
            const idx = bindings.findIndex((b) => b.id === binding.id && b.seq === internal.seq);
            if (idx >= 0) {
                bindings.splice(idx, 1);
            }
        };
    }

    function enable(): void {
        enabled = true;
    }

    function disable(): void {
        enabled = false;
        clearSequenceBuffer();
    }

    function isEnabled(): boolean {
        return enabled;
    }

    function getAllBindings(): HotkeyBinding[] {
        return bindings.map(({ parsed: _parsed, seq: _seq, ...rest }) => rest);
    }

    return {
        register,
        enable,
        disable,
        isEnabled,
        getAllBindings,
        handleEvent,
    };
}

// ---------------------------------------------------------------------------
// 导出解析工具（供测试和帮助面板使用）
// ---------------------------------------------------------------------------

export function parseHotkeyString(keys: string): ParsedBinding {
    return parseBinding(keys);
}

export function formatHotkeyForDisplay(keys: string): string {
    return keys
        .split(/\s+/)
        .map((part) =>
            part
                .split('+')
                .map((p) => {
                    const t = p.trim();
                    if (!t) return t;
                    const lower = t.toLowerCase();
                    if (lower === 'cmd' || lower === 'meta' || lower === 'command') return '⌘';
                    if (lower === 'ctrl' || lower === 'control') return 'Ctrl';
                    if (lower === 'shift') return 'Shift';
                    if (lower === 'alt' || lower === 'option' || lower === 'opt') return 'Alt';
                    return t;
                })
                .join('+'),
        )
        .join(' ');
}
