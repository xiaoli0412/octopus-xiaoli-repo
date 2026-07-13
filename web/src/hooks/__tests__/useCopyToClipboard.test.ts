import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useCopyToClipboard } from '../useCopyToClipboard';

describe('useCopyToClipboard', () => {
    let writeTextMock: ReturnType<typeof vi.fn>;

    beforeEach(() => {
        vi.useFakeTimers();
        writeTextMock = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', {
            value: { writeText: writeTextMock },
            configurable: true,
            writable: true,
        });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('starts with copied = false', () => {
        const { result } = renderHook(() => useCopyToClipboard());

        expect(result.current.copied).toBe(false);
    });

    it('copies text and sets copied to true on success', async () => {
        const { result } = renderHook(() => useCopyToClipboard());

        let success: boolean | undefined;
        await act(async () => {
            success = await result.current.copy('hello world');
        });

        expect(writeTextMock).toHaveBeenCalledWith('hello world');
        expect(success).toBe(true);
        expect(result.current.copied).toBe(true);
    });

    it('returns false and keeps copied false on failure', async () => {
        writeTextMock.mockRejectedValue(new Error('clipboard denied'));

        const { result } = renderHook(() => useCopyToClipboard());

        let success: boolean | undefined;
        await act(async () => {
            success = await result.current.copy('fail');
        });

        expect(success).toBe(false);
        expect(result.current.copied).toBe(false);
    });

    it('resets copied to false after 2 seconds', async () => {
        const { result } = renderHook(() => useCopyToClipboard());

        await act(async () => {
            await result.current.copy('temp');
        });
        expect(result.current.copied).toBe(true);

        act(() => {
            vi.advanceTimersByTime(1999);
        });
        expect(result.current.copied).toBe(true);

        act(() => {
            vi.advanceTimersByTime(1);
        });
        expect(result.current.copied).toBe(false);
    });

    it('resets the timer when copying again before 2 seconds', async () => {
        const { result } = renderHook(() => useCopyToClipboard());

        await act(async () => {
            await result.current.copy('first');
        });
        expect(result.current.copied).toBe(true);

        act(() => {
            vi.advanceTimersByTime(1000);
        });

        await act(async () => {
            await result.current.copy('second');
        });
        expect(result.current.copied).toBe(true);

        act(() => {
            vi.advanceTimersByTime(1999);
        });
        expect(result.current.copied).toBe(true);

        act(() => {
            vi.advanceTimersByTime(1);
        });
        expect(result.current.copied).toBe(false);
    });
});
