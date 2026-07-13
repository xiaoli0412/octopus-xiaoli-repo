import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useMediaQuery, useIsMobile, useIsTablet } from '../useMediaQuery';

type MatchMediaListener = (e: MediaQueryListEvent) => void;

interface MockMQL {
    matches: boolean;
    media: string;
    onchange: null;
    addEventListener: ReturnType<typeof vi.fn>;
    removeEventListener: ReturnType<typeof vi.fn>;
    addListener: ReturnType<typeof vi.fn>;
    removeListener: ReturnType<typeof vi.fn>;
    dispatchEvent: ReturnType<typeof vi.fn>;
    __listeners: Set<MatchMediaListener>;
    __setMatches: (m: boolean) => void;
}

function installMatchMedia(initialMatches: boolean): MockMQL {
    const listeners = new Set<MatchMediaListener>();
    const mql: MockMQL = {
        matches: initialMatches,
        media: '',
        onchange: null,
        addEventListener: vi.fn((event: string, listener: MatchMediaListener) => {
            if (event === 'change') listeners.add(listener);
        }),
        removeEventListener: vi.fn((event: string, listener: MatchMediaListener) => {
            if (event === 'change') listeners.delete(listener);
        }),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
        __listeners: listeners,
        __setMatches: (m: boolean) => {
            mql.matches = m;
        },
    };

    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue(mql));
    return mql;
}

describe('useMediaQuery', () => {
    beforeEach(() => {
        vi.stubGlobal('matchMedia', undefined);
    });

    afterEach(() => {
        vi.unstubAllGlobals();
        vi.restoreAllMocks();
    });

    it('returns true after mount when the query matches', () => {
        installMatchMedia(true);

        const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'));

        expect(result.current).toBe(true);
    });

    it('returns false when the query does not match', () => {
        installMatchMedia(false);

        const { result } = renderHook(() => useMediaQuery('(min-width: 9999px)'));

        expect(result.current).toBe(false);
    });

    it('updates when the match state changes via the change event', () => {
        const mql = installMatchMedia(false);

        const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'));
        expect(result.current).toBe(false);

        act(() => {
            mql.__setMatches(true);
            mql.__listeners.forEach((listener) => listener({ matches: true } as MediaQueryListEvent));
        });

        expect(result.current).toBe(true);
    });

    it('removes the change listener on unmount', () => {
        const mql = installMatchMedia(true);

        const { unmount } = renderHook(() => useMediaQuery('(min-width: 768px)'));

        unmount();

        expect(mql.removeEventListener).toHaveBeenCalledWith('change', expect.any(Function));
    });

    it('does not throw when window.matchMedia is unavailable', () => {
        // matchMedia remains undefined (stubbed in beforeEach)
        const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'));

        expect(result.current).toBe(false);
    });

    it('re-subscribes when the query changes', () => {
        const mql = installMatchMedia(true);

        const { rerender } = renderHook(({ query }) => useMediaQuery(query), {
            initialProps: { query: '(min-width: 768px)' },
        });

        rerender({ query: '(min-width: 1024px)' });

        expect(mql.removeEventListener).toHaveBeenCalledWith('change', expect.any(Function));
    });
});

describe('useIsMobile', () => {
    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('returns true when max-width: 768px matches', () => {
        installMatchMedia(true);

        const { result } = renderHook(() => useIsMobile());

        expect(result.current).toBe(true);
    });

    it('returns false when max-width: 768px does not match', () => {
        installMatchMedia(false);

        const { result } = renderHook(() => useIsMobile());

        expect(result.current).toBe(false);
    });
});

describe('useIsTablet', () => {
    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('returns true when max-width: 1024px matches', () => {
        installMatchMedia(true);

        const { result } = renderHook(() => useIsTablet());

        expect(result.current).toBe(true);
    });

    it('returns false when max-width: 1024px does not match', () => {
        installMatchMedia(false);

        const { result } = renderHook(() => useIsTablet());

        expect(result.current).toBe(false);
    });
});
