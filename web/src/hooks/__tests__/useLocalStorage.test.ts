import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useLocalStorage } from '../useLocalStorage';

const STORAGE_KEY = 'test-key';

describe('useLocalStorage', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    afterEach(() => {
        window.localStorage.clear();
        vi.restoreAllMocks();
    });

    it('returns the initial value when localStorage is empty', () => {
        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'default'));

        expect(result.current[0]).toBe('default');
    });

    it('reads the value from localStorage on mount', () => {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify('stored-value'));

        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'default'));

        expect(result.current[0]).toBe('stored-value');
    });

    it('writes the value to localStorage when set', () => {
        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'default'));

        act(() => {
            result.current[1]('new-value');
        });

        expect(result.current[0]).toBe('new-value');
        expect(window.localStorage.getItem(STORAGE_KEY)).toBe(JSON.stringify('new-value'));
    });

    it('supports functional updates', () => {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(10));

        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 0));

        act(() => {
            result.current[1]((prev) => prev + 5);
        });

        expect(result.current[0]).toBe(15);
        expect(window.localStorage.getItem(STORAGE_KEY)).toBe(JSON.stringify(15));
    });

    it('serializes objects to JSON', () => {
        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, { name: '', age: 0 }));

        act(() => {
            result.current[1]({ name: 'octopus', age: 3 });
        });

        expect(result.current[0]).toEqual({ name: 'octopus', age: 3 });
        expect(window.localStorage.getItem(STORAGE_KEY)).toBe(JSON.stringify({ name: 'octopus', age: 3 }));
    });

    it('falls back to initial value when stored JSON is invalid', () => {
        window.localStorage.setItem(STORAGE_KEY, 'not-valid-json');

        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'fallback'));

        expect(result.current[0]).toBe('fallback');
    });

    it('returns fallback when getItem throws', () => {
        vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('read error');
        });

        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'fallback'));

        expect(result.current[0]).toBe('fallback');
    });

    it('handles setItem errors gracefully', () => {
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('write error');
        });

        const { result } = renderHook(() => useLocalStorage(STORAGE_KEY, 'initial'));

        act(() => {
            result.current[1]('updated');
        });

        expect(result.current[0]).toBe('updated');
    });
});
