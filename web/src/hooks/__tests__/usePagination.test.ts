import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { usePagination } from '../usePagination';

describe('usePagination', () => {
    it('uses default initial page and page size', () => {
        const { result } = renderHook(() => usePagination({ total: 100 }));

        expect(result.current.page).toBe(1);
        expect(result.current.pageSize).toBe(10);
        expect(result.current.totalPages).toBe(10);
    });

    it('respects custom initial page and page size', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 100, initialPage: 3, initialPageSize: 20 }),
        );

        expect(result.current.page).toBe(3);
        expect(result.current.pageSize).toBe(20);
        expect(result.current.totalPages).toBe(5);
    });

    it('calculates totalPages correctly with remainders', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 105, initialPageSize: 20 }),
        );

        expect(result.current.totalPages).toBe(6);
    });

    it('clamps setPage to a minimum of 1', () => {
        const { result } = renderHook(() => usePagination({ total: 100 }));

        act(() => {
            result.current.setPage(0);
        });

        expect(result.current.page).toBe(1);
    });

    it('clamps setPage to totalPages', () => {
        const { result } = renderHook(() => usePagination({ total: 50, initialPageSize: 10 }));

        act(() => {
            result.current.setPage(999);
        });

        expect(result.current.page).toBe(5);
    });

    it('navigates to the next page', () => {
        const { result } = renderHook(() => usePagination({ total: 100 }));

        act(() => {
            result.current.nextPage();
        });

        expect(result.current.page).toBe(2);
    });

    it('navigates to the previous page', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 100, initialPage: 3 }),
        );

        act(() => {
            result.current.prevPage();
        });

        expect(result.current.page).toBe(2);
    });

    it('does not go beyond the last page', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 20, initialPage: 2, initialPageSize: 10 }),
        );

        act(() => {
            result.current.nextPage();
        });

        expect(result.current.page).toBe(2);
    });

    it('does not go below the first page', () => {
        const { result } = renderHook(() => usePagination({ total: 100 }));

        act(() => {
            result.current.prevPage();
        });

        expect(result.current.page).toBe(1);
    });

    it('resets to page 1 when setPageSize is called', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 100, initialPage: 5, initialPageSize: 10 }),
        );

        act(() => {
            result.current.setPageSize(20);
        });

        expect(result.current.pageSize).toBe(20);
        expect(result.current.page).toBe(1);
        expect(result.current.totalPages).toBe(5);
    });

    it('reports canNextPage and canPrevPage correctly', () => {
        const { result } = renderHook(() =>
            usePagination({ total: 30, initialPage: 2, initialPageSize: 10 }),
        );

        expect(result.current.canPrevPage).toBe(true);
        expect(result.current.canNextPage).toBe(true);

        act(() => {
            result.current.setPage(1);
        });
        expect(result.current.canPrevPage).toBe(false);
        expect(result.current.canNextPage).toBe(true);

        act(() => {
            result.current.setPage(3);
        });
        expect(result.current.canPrevPage).toBe(true);
        expect(result.current.canNextPage).toBe(false);
    });

    it('handles zero total items', () => {
        const { result } = renderHook(() => usePagination({ total: 0 }));

        expect(result.current.totalPages).toBe(1);
        expect(result.current.canNextPage).toBe(false);
        expect(result.current.canPrevPage).toBe(false);
    });
});
