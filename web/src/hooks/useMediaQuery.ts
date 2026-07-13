'use client';

import { useEffect, useState } from 'react';

export function useMediaQuery(query: string): boolean {
    const [matches, setMatches] = useState<boolean>(false);

    useEffect(() => {
        if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
            return;
        }

        const mql = window.matchMedia(query);
        setMatches(mql.matches);

        const onChange = () => setMatches(mql.matches);
        mql.addEventListener('change', onChange);

        return () => mql.removeEventListener('change', onChange);
    }, [query]);

    return matches;
}

export function useIsMobile(): boolean {
    return useMediaQuery('(max-width: 768px)');
}

export function useIsTablet(): boolean {
    return useMediaQuery('(max-width: 1024px)');
}
