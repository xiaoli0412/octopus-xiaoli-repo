'use client';

import { useMemo, useState } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

/**
 * 安全解析 JSON 字符串。解析失败时返回 null。
 */
function tryParseJSON(value: string): unknown {
    if (!value || !value.trim()) return null;
    try {
        return JSON.parse(value);
    } catch {
        return null;
    }
}

type DiffKind = 'added' | 'removed' | 'changed' | 'unchanged';

interface DiffNode {
    key: string;
    kind: DiffKind;
    beforeValue?: unknown;
    afterValue?: unknown;
    children?: DiffNode[];
}

/**
 * 递归对比两个值，生成 diff 节点树。
 */
function diffValues(key: string, before: unknown, after: unknown): DiffNode {
    const beforeIsObj = isPlainObject(before);
    const afterIsObj = isPlainObject(after);

    if (beforeIsObj && afterIsObj) {
        return {
            key,
            kind: 'unchanged',
            children: diffObjects(before as Record<string, unknown>, after as Record<string, unknown>),
        };
    }

    if (before === undefined && after !== undefined) {
        return { key, kind: 'added', afterValue: after };
    }
    if (before !== undefined && after === undefined) {
        return { key, kind: 'removed', beforeValue: before };
    }

    if (JSON.stringify(before) !== JSON.stringify(after)) {
        return { key, kind: 'changed', beforeValue: before, afterValue: after };
    }

    return { key, kind: 'unchanged', beforeValue: before, afterValue: after };
}

function diffObjects(before: Record<string, unknown>, after: Record<string, unknown>): DiffNode[] {
    const keys = new Set<string>([...Object.keys(before ?? {}), ...Object.keys(after ?? {})]);
    const nodes: DiffNode[] = [];
    for (const key of keys) {
        const bVal = before?.[key];
        const aVal = after?.[key];
        const bDefined = bVal !== undefined;
        const aDefined = aVal !== undefined;

        if (bDefined && aDefined) {
            nodes.push(diffValues(key, bVal, aVal));
        } else if (aDefined) {
            nodes.push({ key, kind: 'added', afterValue: aVal });
        } else {
            nodes.push({ key, kind: 'removed', beforeValue: bVal });
        }
    }
    return nodes;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
    return !!value && typeof value === 'object' && !Array.isArray(value);
}

function formatScalar(value: unknown): string {
    if (value === undefined) return '';
    if (value === null) return 'null';
    if (typeof value === 'string') return value;
    return JSON.stringify(value);
}

const KIND_STYLES: Record<DiffKind, string> = {
    added: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
    removed: 'bg-rose-500/10 text-rose-600 dark:text-rose-400',
    changed: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
    unchanged: 'text-foreground/80',
};

const KIND_LABEL: Record<DiffKind, string> = {
    added: '+',
    removed: '-',
    changed: '~',
    unchanged: ' ',
};

const MAX_RENDER_LINES = 200;

function DiffNodeRow({ node, depth }: { node: DiffNode; depth: number }) {
    const [open, setOpen] = useState(depth < 2);
    const hasChildren = !!node.children && node.children.length > 0;
    const childKinds = node.children?.some((c) => c.kind !== 'unchanged') ?? false;
    const rowKind: DiffKind = hasChildren ? (childKinds ? 'changed' : 'unchanged') : node.kind;

    return (
        <div>
            <div
                className={cn('flex items-start gap-1 rounded px-1 py-0.5 font-mono text-xs', KIND_STYLES[rowKind])}
                style={{ paddingLeft: `${depth * 12 + 4}px` }}
            >
                {hasChildren ? (
                    <button
                        type="button"
                        onClick={() => setOpen((v) => !v)}
                        className="mt-0.5 shrink-0 text-muted-foreground hover:text-foreground"
                        aria-label={open ? 'collapse' : 'expand'}
                    >
                        {open ? <ChevronDown className="size-3" /> : <ChevronRight className="size-3" />}
                    </button>
                ) : (
                    <span className="w-3 shrink-0" />
                )}
                <span className="shrink-0 select-none text-muted-foreground">{KIND_LABEL[node.kind]}</span>
                <span className="shrink-0 font-medium">{node.key}:</span>
                {hasChildren ? (
                    <span className="text-muted-foreground">
                        {`{${node.children?.length ?? 0}}`}
                    </span>
                ) : node.kind === 'removed' ? (
                    <span className="break-all">{formatScalar(node.beforeValue)}</span>
                ) : node.kind === 'added' ? (
                    <span className="break-all">{formatScalar(node.afterValue)}</span>
                ) : node.kind === 'changed' ? (
                    <span className="break-all">
                        <span className="line-through opacity-60">{formatScalar(node.beforeValue)}</span>
                        {' → '}
                        <span>{formatScalar(node.afterValue)}</span>
                    </span>
                ) : (
                    <span className="break-all text-muted-foreground">{formatScalar(node.afterValue)}</span>
                )}
            </div>
            {hasChildren && open && (
                <div>
                    {node.children!.map((child) => (
                        <DiffNodeRow key={child.key} node={child} depth={depth + 1} />
                    ))}
                </div>
            )}
        </div>
    );
}

export interface JsonDiffProps {
    before: string;
    after: string;
    className?: string;
}

/**
 * JSON diff 组件
 *
 * 解析 before / after JSON 字符串，逐字段对比并高亮差异：
 * - 新增字段：绿色
 * - 删除字段：红色
 * - 修改字段：黄色
 * - 大 JSON 超过 MAX_RENDER_LINES 时截断显示。
 *
 * @example
 * <JsonDiff before={log.before_json} after={log.after_json} />
 */
export function JsonDiff({ before, after, className }: JsonDiffProps) {
    const { nodes, fallback } = useMemo(() => {
        const beforeObj = tryParseJSON(before);
        const afterObj = tryParseJSON(after);

        if (beforeObj === null && afterObj === null) {
            return {
                nodes: [] as DiffNode[],
                fallback: {
                    before: before || '(empty)',
                    after: after || '(empty)',
                    parseError: !!(before.trim() || after.trim()),
                },
            };
        }

        const bObj = isPlainObject(beforeObj) ? beforeObj : { value: beforeObj };
        const aObj = isPlainObject(afterObj) ? afterObj : { value: afterObj };

        return { nodes: diffObjects(bObj, aObj), fallback: null };
    }, [before, after]);

    if (fallback) {
        return (
            <div className={cn('space-y-2 text-xs', className)}>
                {fallback.parseError && (
                    <div className="rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-amber-600 dark:text-amber-400">
                        JSON 解析失败，以下为原始文本。
                    </div>
                )}
                <div className="grid gap-2 md:grid-cols-2">
                    <pre className="overflow-auto rounded-md border border-border/60 bg-muted/30 p-2 text-xs text-foreground/80">
                        {fallback.before}
                    </pre>
                    <pre className="overflow-auto rounded-md border border-border/60 bg-muted/30 p-2 text-xs text-foreground/80">
                        {fallback.after}
                    </pre>
                </div>
            </div>
        );
    }

    const rendered = nodes.slice(0, MAX_RENDER_LINES);
    const truncated = nodes.length - rendered.length;

    return (
        <div className={cn('rounded-md border border-border/60 bg-muted/20 p-2', className)}>
            <div className="space-y-0.5">
                {rendered.map((node) => (
                    <DiffNodeRow key={node.key} node={node} depth={0} />
                ))}
            </div>
            {truncated > 0 && (
                <div className="mt-2 px-1 text-xs text-muted-foreground">
                    … {truncated} 个字段未展示
                </div>
            )}
        </div>
    );
}
