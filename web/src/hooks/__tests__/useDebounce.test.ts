import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useDebounce } from '../useDebounce';

describe('useDebounce', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('returns the initial value immediately', () => {
        const { result } = renderHook(() => useDebounce('initial', 300));
        expect(result.current).toBe('initial');
    });

    it('updates to the new value only after the delay', () => {
        const { result, rerender } = renderHook(({ value, delay }) => useDebounce(value, delay), {
            initialProps: { value: 'a', delay: 300 },
        });

        rerender({ value: 'b', delay: 300 });
        expect(result.current).toBe('a');

        act(() => {
            vi.advanceTimersByTime(299);
        });
        expect(result.current).toBe('a');

        act(() => {
            vi.advanceTimersByTime(1);
        });
        expect(result.current).toBe('b');
    });

    it('only keeps the last value when changing rapidly', () => {
        const { result, rerender } = renderHook(({ value, delay }) => useDebounce(value, delay), {
            initialProps: { value: 'first', delay: 300 },
        });

        rerender({ value: 'second', delay: 300 });
        act(() => {
            vi.advanceTimersByTime(100);
        });

        rerender({ value: 'third', delay: 300 });
        act(() => {
            vi.advanceTimersByTime(100);
        });

        rerender({ value: 'fourth', delay: 300 });
        expect(result.current).toBe('first');

        act(() => {
            vi.advanceTimersByTime(300);
        });
        expect(result.current).toBe('fourth');
    });

    it('resets the timer when the value changes before the delay elapses', () => {
        const { result, rerender } = renderHook(({ value, delay }) => useDebounce(value, delay), {
            initialProps: { value: 'a', delay: 300 },
        });

        rerender({ value: 'b', delay: 300 });
        act(() => {
            vi.advanceTimersByTime(200);
        });

        rerender({ value: 'c', delay: 300 });
        act(() => {
            vi.advanceTimersByTime(200);
        });
        expect(result.current).toBe('a');

        act(() => {
            vi.advanceTimersByTime(100);
        });
        expect(result.current).toBe('c');
    });

    it('supports generic types', () => {
        const { result, rerender } = renderHook(({ value, delay }) => useDebounce(value, delay), {
            initialProps: { value: { count: 0 }, delay: 100 },
        });

        rerender({ value: { count: 5 }, delay: 100 });
        expect(result.current).toEqual({ count: 0 });

        act(() => {
            vi.advanceTimersByTime(100);
        });
        expect(result.current).toEqual({ count: 5 });
    });
});
