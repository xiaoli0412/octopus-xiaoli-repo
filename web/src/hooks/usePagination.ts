'use client';

import { useCallback, useState } from 'react';

interface UsePaginationOptions {
    initialPage?: number;
    initialPageSize?: number;
    total: number;
}

interface UsePaginationReturn {
    page: number;
    pageSize: number;
    totalPages: number;
    setPage: (p: number) => void;
    setPageSize: (s: number) => void;
    nextPage: () => void;
    prevPage: () => void;
    canNextPage: boolean;
    canPrevPage: boolean;
}

export function usePagination(options: UsePaginationOptions): UsePaginationReturn {
    const { initialPage = 1, initialPageSize = 10, total } = options;
    const [page, setPageState] = useState<number>(initialPage);
    const [pageSize, setPageSizeState] = useState<number>(initialPageSize);

    const totalPages = pageSize > 0 ? Math.max(1, Math.ceil(total / pageSize)) : 1;

    const setPage = useCallback(
        (p: number) => {
            setPageState(Math.max(1, Math.min(p, totalPages)));
        },
        [totalPages],
    );

    const setPageSize = useCallback((s: number) => {
        setPageSizeState(s);
        setPageState(1);
    }, []);

    const nextPage = useCallback(() => {
        setPageState((p) => Math.min(p + 1, totalPages));
    }, [totalPages]);

    const prevPage = useCallback(() => {
        setPageState((p) => Math.max(p - 1, 1));
    }, []);

    const canNextPage = page < totalPages;
    const canPrevPage = page > 1;

    return {
        page,
        pageSize,
        totalPages,
        setPage,
        setPageSize,
        nextPage,
        prevPage,
        canNextPage,
        canPrevPage,
    };
}
