'use client';

import { useCallback, useEffect, useState } from 'react';

export function useLocalStorage<T>(
    key: string,
    initialValue: T,
): [T, (value: T | ((prev: T) => T)) => void] {
    const [storedValue, setStoredValue] = useState<T>(() => {
        if (typeof window === 'undefined') {
            return initialValue;
        }
        try {
            const item = window.localStorage.getItem(key);
            return item ? (JSON.parse(item) as T) : initialValue;
        } catch {
            return initialValue;
        }
    });

    const setValue = useCallback(
        (value: T | ((prev: T) => T)) => {
            setStoredValue((prev) => {
                const nextValue = value instanceof Function ? (value as (prev: T) => T)(prev) : value;
                try {
                    window.localStorage.setItem(key, JSON.stringify(nextValue));
                } catch {
                    // ignore write errors
                }
                return nextValue;
            });
        },
        [key],
    );

    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        const handleStorageChange = (e: StorageEvent) => {
            if (e.key !== key) {
                return;
            }
            try {
                setStoredValue(e.newValue ? (JSON.parse(e.newValue) as T) : initialValue);
            } catch {
                // ignore parse errors
            }
        };

        window.addEventListener('storage', handleStorageChange);
        return () => window.removeEventListener('storage', handleStorageChange);
    }, [key, initialValue]);

    return [storedValue, setValue];
}
