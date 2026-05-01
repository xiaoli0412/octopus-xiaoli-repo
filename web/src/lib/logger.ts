/**
 * 统一的日志工具
 * 在生产环境中过滤敏感的日志输出
 */

const isDevelopment = process.env.NODE_ENV !== 'production';

const REDACTED = '[REDACTED]';
const sensitiveLogKeys = new Set([
    'apikey',
    'channelkey',
    'authorization',
    'token',
    'password',
    'oldpassword',
    'newpassword',
    'secret',
    'clientsecret',
    'accesstoken',
    'refreshtoken',
    'previewtoken',
    'devicecode',
    'cookie',
    'setcookie',
]);

const secretLikePattern = /sk-[A-Za-z0-9_-]{8,}|Bearer\s+[A-Za-z0-9._~+\/-]{8,}|github_pat_[A-Za-z0-9_]+|gh[opusr]_[A-Za-z0-9]+/g;

function normalizeKey(key: string): string {
    return key.toLowerCase().replace(/[^a-z0-9]/g, '');
}

function shouldRedactKey(key: string): boolean {
    const normalized = normalizeKey(key);
    return sensitiveLogKeys.has(normalized) || normalized.endsWith('password') || normalized.endsWith('secret');
}

function sanitizeString(value: string): string {
    return value.replace(secretLikePattern, REDACTED);
}

function sanitizeForLog(value: unknown, seen: WeakSet<object> = new WeakSet()): unknown {
    if (typeof value === 'string') {
        return sanitizeString(value);
    }

    if (value == null || typeof value !== 'object') {
        return value;
    }

    if (value instanceof Error) {
        const errorView: Record<string, unknown> = {
            name: value.name,
            message: sanitizeString(value.message),
        };
        if (value.stack) {
            errorView.stack = sanitizeString(value.stack);
        }
        for (const [key, nested] of Object.entries(value)) {
            errorView[key] = shouldRedactKey(key) ? REDACTED : sanitizeForLog(nested, seen);
        }
        return errorView;
    }

    if (Array.isArray(value)) {
        return value.map((item) => sanitizeForLog(item, seen));
    }

    if (value instanceof Date) {
        return value.toISOString();
    }

    if (typeof Headers !== 'undefined' && value instanceof Headers) {
        return sanitizeForLog(Object.fromEntries(value.entries()), seen);
    }

    if (typeof URLSearchParams !== 'undefined' && value instanceof URLSearchParams) {
        return sanitizeString(value.toString());
    }

    if (seen.has(value)) {
        return '[Circular]';
    }
    seen.add(value);

    const source = value as Record<string, unknown>;
    const sanitized: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(source)) {
        sanitized[key] = shouldRedactKey(key) ? REDACTED : sanitizeForLog(nested, seen);
    }
    return sanitized;
}

function sanitizeArgs(args: unknown[]): unknown[] {
    return args.map((arg) => sanitizeForLog(arg));
}

export const logger = {
    /**
     * 普通日志 - 仅在开发环境输出
     */
    log: (...args: unknown[]) => {
        if (isDevelopment) {
            console.log(...sanitizeArgs(args));
        }
    },

    /**
     * 错误日志 - 在所有环境输出
     */
    error: (...args: unknown[]) => {
        console.error(...sanitizeArgs(args));
    },

    /**
     * 警告日志 - 在所有环境输出
     */
    warn: (...args: unknown[]) => {
        console.warn(...sanitizeArgs(args));
    },

    /**
     * 调试日志 - 仅在开发环境输出
     */
    debug: (...args: unknown[]) => {
        if (isDevelopment) {
            console.debug(...sanitizeArgs(args));
        }
    },

    /**
     * 信息日志 - 仅在开发环境输出
     */
    info: (...args: unknown[]) => {
        if (isDevelopment) {
            console.info(...sanitizeArgs(args));
        }
    },
};
