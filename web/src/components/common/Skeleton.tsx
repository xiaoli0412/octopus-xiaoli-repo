'use client';

import { type ReactNode, useMemo } from 'react';
import { cn } from '@/lib/utils';

/**
 * Octopus 统一骨架屏组件
 *
 * 设计原则：
 * - 复用 globals.css 的设计 tokens（bg-muted / rounded-md / octo-section）
 * - 使用 Tailwind 内置 animate-pulse，并通过 motion-reduce: 变体尊重
 *   prefers-reduced-motion，自动停止闪烁。
 * - 仅做占位，不引入新颜色或硬编码色值。
 */

const PULSE = 'animate-pulse motion-reduce:animate-none';

/** 解析宽度：数字按 rem，字符串原样使用（支持 "100%"、"12rem" 等）。 */
function resolveWidth(width?: string | number): string | undefined {
    if (width === undefined) return undefined;
    return typeof width === 'number' ? `${width}rem` : width;
}

/** 单行骨架条。 */
export function SkeletonLine({
    width,
    className,
}: {
    width?: string | number;
    className?: string;
}) {
    const style = useMemo(() => {
        const w = resolveWidth(width);
        return w ? { width: w } : undefined;
    }, [width]);
    return (
        <div
            style={style}
            className={cn('h-3.5 w-full rounded-md bg-muted', PULSE, className)}
            aria-hidden="true"
        />
    );
}

/** 卡片骨架：标题行 + 内容行 + 操作行，匹配 octo-section 布局。 */
export function SkeletonCard({
    className,
    rows = 3,
    showActions = true,
}: {
    className?: string;
    rows?: number;
    showActions?: boolean;
}) {
    return (
        <div
            className={cn(
                'min-w-0 rounded-[1.05rem] border border-border/80 bg-card p-3 shadow-sm',
                className,
            )}
            aria-hidden="true"
        >
            <div className="flex items-center justify-between gap-2">
                <div className={cn('h-4 w-1/3 rounded-md bg-muted', PULSE)} />
                <div className={cn('h-5 w-5 rounded-md bg-muted', PULSE)} />
            </div>
            <div className="mt-3 space-y-2">
                {Array.from({ length: rows }).map((_, i) => (
                    <div
                        key={i}
                        className={cn(
                            'h-3.5 rounded-md bg-muted',
                            PULSE,
                            i === rows - 1 ? 'w-2/3' : 'w-full',
                        )}
                    />
                ))}
            </div>
            {showActions ? (
                <div className="mt-3 flex items-center gap-2">
                    <div className={cn('h-7 w-16 rounded-lg bg-muted', PULSE)} />
                    <div className={cn('h-7 w-12 rounded-lg bg-muted', PULSE)} />
                </div>
            ) : null}
        </div>
    );
}

/** 表格骨架：表头 + N 行数据，每行若干单元格。 */
export function SkeletonTable({
    rows = 5,
    columns = 4,
    className,
}: {
    rows?: number;
    columns?: number;
    className?: string;
}) {
    return (
        <div
            className={cn(
                'min-w-0 overflow-hidden rounded-[1.05rem] border border-border/80 bg-card shadow-sm',
                className,
            )}
            aria-hidden="true"
        >
            <div className="flex items-center gap-3 border-b border-border/70 px-3 py-2.5">
                {Array.from({ length: columns }).map((_, i) => (
                    <div
                        key={i}
                        className={cn(
                            'h-3.5 rounded-md bg-muted',
                            PULSE,
                            i === 0 ? 'flex-1' : 'w-20',
                        )}
                    />
                ))}
            </div>
            <div className="divide-y divide-border/60">
                {Array.from({ length: rows }).map((_, r) => (
                    <div key={r} className="flex items-center gap-3 px-3 py-2.5">
                        {Array.from({ length: columns }).map((_, c) => (
                            <div
                                key={c}
                                className={cn(
                                    'h-3.5 rounded-md bg-muted',
                                    PULSE,
                                    c === 0 ? 'flex-1' : 'w-20',
                                )}
                            />
                        ))}
                    </div>
                ))}
            </div>
        </div>
    );
}

/** 列表骨架：图标 + 标题 + 副标题，匹配渠道/分组/模型列表项。 */
export function SkeletonList({
    count = 4,
    className,
}: {
    count?: number;
    className?: string;
}) {
    return (
        <div className={cn('space-y-2', className)} aria-hidden="true">
            {Array.from({ length: count }).map((_, i) => (
                <div
                    key={i}
                    className="flex items-center gap-3 rounded-[1.05rem] border border-border/80 bg-card p-3 shadow-sm"
                >
                    <div className={cn('size-9 shrink-0 rounded-xl bg-muted', PULSE)} />
                    <div className="min-w-0 flex-1 space-y-2">
                        <div className={cn('h-4 w-2/5 rounded-md bg-muted', PULSE)} />
                        <div className={cn('h-3 w-3/5 rounded-md bg-muted', PULSE)} />
                    </div>
                    <div className={cn('h-7 w-12 rounded-lg bg-muted', PULSE)} />
                </div>
            ))}
        </div>
    );
}

/** 详情骨架：封面/头部 + 标题 + 段落。 */
export function SkeletonDetail({
    className,
    paragraphs = 3,
}: {
    className?: string;
    paragraphs?: number;
}) {
    return (
        <div
            className={cn(
                'min-w-0 space-y-3 rounded-[1.05rem] border border-border/80 bg-card p-4 shadow-sm',
                className,
            )}
            aria-hidden="true"
        >
            <div className={cn('h-24 w-full rounded-xl bg-muted', PULSE)} />
            <div className={cn('h-5 w-1/2 rounded-md bg-muted', PULSE)} />
            <div className="space-y-2 pt-1">
                {Array.from({ length: paragraphs }).map((_, i) => (
                    <div
                        key={i}
                        className={cn(
                            'h-3.5 rounded-md bg-muted',
                            PULSE,
                            i === paragraphs - 1 ? 'w-2/3' : 'w-full',
                        )}
                    />
                ))}
            </div>
        </div>
    );
}

/**
 * 骨架网格：按响应式列数渲染 N 个 SkeletonCard。
 * 供 VirtualizedGrid 等列表在加载态时复用真实布局。
 */
export function SkeletonGrid({
    count = 6,
    columns = { default: 1, md: 2, lg: 3 },
    gap = 12,
    className,
    cardProps,
}: {
    count?: number;
    columns?: Record<string, number>;
    gap?: number;
    className?: string;
    cardProps?: React.ComponentProps<typeof SkeletonCard>;
}) {
    const colClass = useMemo(() => {
        const parts: string[] = [];
        const map: Record<string, string> = {
            default: 'grid-cols-',
            sm: 'sm:grid-cols-',
            md: 'md:grid-cols-',
            lg: 'lg:grid-cols-',
            xl: 'xl:grid-cols-',
            '2xl': '2xl:grid-cols-',
        };
        for (const key of Object.keys(columns)) {
            const prefix = map[key];
            if (prefix) parts.push(`${prefix}${columns[key]}`);
        }
        return parts.length ? parts.join(' ') : 'grid-cols-1';
    }, [columns]);

    return (
        <div
            className={cn('grid', colClass, className)}
            style={{ gap: `${gap}px` }}
            aria-hidden="true"
        >
            {Array.from({ length: count }).map((_, i) => (
                <SkeletonCard key={i} {...cardProps} />
            ))}
        </div>
    );
}

export type { ReactNode };
