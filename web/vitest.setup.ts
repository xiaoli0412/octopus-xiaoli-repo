import '@testing-library/jest-dom/vitest';

if (typeof globalThis.requestAnimationFrame !== 'function') {
	globalThis.requestAnimationFrame = (callback: FrameRequestCallback) => setTimeout(() => callback(Date.now()), 0) as unknown as number;
}

if (typeof globalThis.cancelAnimationFrame !== 'function') {
	globalThis.cancelAnimationFrame = (handle: number) => clearTimeout(handle);
}
