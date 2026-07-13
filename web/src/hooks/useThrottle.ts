'use client';

import { useCallback, useRef } from 'react';

export function useThrottle<T extends (...args: any[]) => void>(fn: T, delay: number): T {
    const lastRunRef = useRef<number | null>(null);
    const fnRef = useRef<T>(fn);
    fnRef.current = fn;

    const throttled = useCallback(
        (...args: Parameters<T>) => {
            const now = Date.now();
            if (lastRunRef.current === null || now - lastRunRef.current >= delay) {
                lastRunRef.current = now;
                fnRef.current(...args);
            }
        },
        [delay],
    ) as T;

    return throttled;
}
