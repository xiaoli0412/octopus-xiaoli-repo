import { defineConfig } from 'vitest/config';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	test: {
		environment: 'jsdom',
		environmentOptions: {
			jsdom: {
				url: 'http://127.0.0.1/',
			},
		},
		globals: true,
		setupFiles: ['./vitest.setup.ts'],
		css: false,
		exclude: ['e2e/**', 'node_modules/**', 'dist/**', 'out/**'],
	},
	server: {
		host: '127.0.0.1',
	},
	preview: {
		host: '127.0.0.1',
	},
	resolve: {
		alias: {
			'@': path.resolve(__dirname, './src'),
		},
	},
});
