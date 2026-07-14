'use client';

import { useEffect } from 'react';

export default function GlobalError({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    useEffect(() => {
        console.error('Global error caught:', error);
    }, [error]);

    const handleClearAndReload = () => {
        try {
            localStorage.clear();
            sessionStorage.clear();
        } catch {
            // ignore
        }
        window.location.href = '/';
    };

    return (
        <html>
            <body>
                <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minHeight: '100vh',
                    fontFamily: 'system-ui, -apple-system, sans-serif',
                    textAlign: 'center',
                    padding: '2rem',
                    background: '#fafafa',
                    color: '#1a1a1a',
                }}>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 600, marginBottom: '0.5rem' }}>
                        Octopus
                    </h1>
                    <p style={{ fontSize: '0.875rem', color: '#666', marginBottom: '1.5rem' }}>
                        页面加载异常，可能是浏览器缓存了旧版本数据。
                    </p>
                    <p style={{ fontSize: '0.75rem', color: '#999', marginBottom: '1.5rem', maxWidth: '400px', wordBreak: 'break-all' }}>
                        {error.message || 'Unknown error'}
                    </p>
                    <button
                        onClick={handleClearAndReload}
                        style={{
                            padding: '0.625rem 1.5rem',
                            fontSize: '0.875rem',
                            fontWeight: 500,
                            borderRadius: '0.5rem',
                            border: '1px solid #ddd',
                            background: '#fff',
                            cursor: 'pointer',
                            transition: 'background 0.2s',
                        }}
                    >
                        清除缓存并重新加载
                    </button>
                    <button
                        onClick={() => reset()}
                        style={{
                            marginTop: '0.5rem',
                            padding: '0.5rem 1.25rem',
                            fontSize: '0.8rem',
                            fontWeight: 400,
                            borderRadius: '0.5rem',
                            border: 'none',
                            background: 'transparent',
                            color: '#666',
                            cursor: 'pointer',
                        }}
                    >
                        重试（不清除缓存）
                    </button>
                </div>
            </body>
        </html>
    );
}
