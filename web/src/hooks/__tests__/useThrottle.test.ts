import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useThrottle } from '../useThrottle';

describe('useThrottle', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date(1000));
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('executes on the first call', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useThrottle(fn, 300));

        act(() => {
            result.current('a');
        });

        expect(fn).toHaveBeenCalledTimes(1);
        expect(fn).toHaveBeenCalledWith('a');
    });

    it('does not execute within the throttle interval', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useThrottle(fn, 300));

        act(() => {
            result.current();
        });

        act(() => {
            vi.advanceTimersByTime(299);
        });

        act(() => {
            result.current();
        });

        expect(fn).toHaveBeenCalledTimes(1);
    });

    it('executes again after the interval has elapsed', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useThrottle(fn, 300));

        act(() => {
            result.current('first');
        });

        act(() => {
            vi.advanceTimersByTime(300);
        });

        act(() => {
            result.current('second');
        });

        expect(fn).toHaveBeenCalledTimes(2);
        expect(fn).toHaveBeenNthCalledWith(1, 'first');
        expect(fn).toHaveBeenNthCalledWith(2, 'second');
    });

    it('passes arguments through to the callback', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useThrottle(fn, 100));

        act(() => {
            result.current(1, 'two', { three: true });
        });

        expect(fn).toHaveBeenCalledWith(1, 'two', { three: true });
    });

    it('updates the throttled function when delay changes', () => {
        const fn = vi.fn();
        const { result, rerender } = renderHook(({ delay }) => useThrottle(fn, delay), {
            initialProps: { delay: 300 },
        });

        act(() => {
            result.current();
        });
        expect(fn).toHaveBeenCalledTimes(1);

        rerender({ delay: 100 });

        act(() => {
            vi.advanceTimersByTime(100);
        });

        act(() => {
            result.current();
        });
        expect(fn).toHaveBeenCalledTimes(2);
    });
});
