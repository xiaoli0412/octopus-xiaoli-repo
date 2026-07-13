'use client';

import { type ReactNode } from 'react';
import { Inbox } from 'lucide-react';
import { motion, useReducedMotion } from 'motion/react';
import { cn } from '@/lib/utils';

export interface EmptyStateProps {
    /** 自定义图标，默认使用 lucide-react 的 Inbox。 */
    icon?: ReactNode;
    /** 主标题。未提供时回退到 children，保证向后兼容。 */
    title?: ReactNode;
    /** 描述文案，渲染在标题下方。 */
    description?: ReactNode;
    /** 操作区，可放置“创建”按钮等 CTA。 */
    action?: ReactNode;
    /** 透传到根容器的 className。 */
    className?: string;
    /** 向后兼容：未提供 title 时作为标题内容渲染。 */
    children?: ReactNode;
}

/**
 * Octopus 统一空状态组件。
 *
 * - 居中布局，垂直间距合理，复用 octo-empty 设计 token。
 * - 使用 motion 做淡入动画，并通过 useReducedMotion 尊重
 *   prefers-reduced-motion，关闭动画时直接呈现。
 * - 默认图标为 Inbox，可通过 icon 替换。
 * - 向后兼容旧用法 <EmptyState>{text}</EmptyState>。
 */
export function EmptyState({
    icon,
    title,
    description,
    action,
    className,
    children,
}: EmptyStateProps) {
    const reduceMotion = useReducedMotion();
    const heading = title ?? children;

    return (
        <motion.div
            className={cn('octo-empty', className)}
            role="status"
            aria-live="polite"
            initial={reduceMotion ? false : { opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.24, ease: 'easeOut' }}
        >
            <span className="flex size-10 items-center justify-center rounded-full bg-muted/60 text-muted-foreground" aria-hidden="true">
                {icon ?? <Inbox className="size-5" />}
            </span>
            {heading !== undefined && heading !== null ? (
                <div className="text-base font-medium text-foreground">{heading}</div>
            ) : null}
            {description ? (
                <div className="max-w-[28rem] text-sm text-muted-foreground">{description}</div>
            ) : null}
            {action ? <div className="pt-1">{action}</div> : null}
        </motion.div>
    );
}

export default EmptyState;
