import { useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { apiClient, setAuthStoreGetter } from '../client';
import { logger } from '@/lib/logger';
import type { APIKeyAuthStatus } from './apikey';

export interface UserLoginRequest {
    username: string;
    password: string;
    expire: number;
}

export interface UserLoginResponse {
    token: string;
    expire_at: string;
    must_change_password: boolean;
}

export interface UserStatusResponse {
    ok: boolean;
    must_change_password: boolean;
}

export interface ChangePasswordRequest {
    old_password: string;
    new_password: string;
}

export interface ForceChangePasswordRequest {
    new_username?: string;
    new_password: string;
}

export interface ChangeUsernameRequest {
    current_password: string;
    new_username: string;
}

interface AuthState {
    isAuthenticated: boolean;
    isLoading: boolean;
    isAPIKeyAuth: boolean;
    token: string | null;
    expireAt: string | null;
    mustChangePassword: boolean;
    apiKeyStatus: APIKeyAuthStatus | null;

    setAuth: (token: string, expireAt: string, mustChangePassword?: boolean) => void;
    setAPIKeyAuth: (apiKey: string, status: APIKeyAuthStatus) => void;
    setMustChangePassword: (mustChangePassword: boolean) => void;
    checkAuth: () => Promise<void>;
    logout: () => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
            isAuthenticated: false,
            isLoading: true,
            isAPIKeyAuth: false,
            token: null,
            expireAt: null,
            mustChangePassword: false,
            apiKeyStatus: null,

            setAuth: (token: string, expireAt: string, mustChangePassword = false) => {
                set({
                    isAuthenticated: true,
                    isAPIKeyAuth: false,
                    token,
                    expireAt,
                    mustChangePassword,
                    apiKeyStatus: null,
                    isLoading: false,
                });
            },

            setAPIKeyAuth: (apiKey: string, status: APIKeyAuthStatus) => {
                set({
                    isAuthenticated: true,
                    isAPIKeyAuth: true,
                    token: apiKey,
                    expireAt: null,
                    mustChangePassword: false,
                    apiKeyStatus: status,
                    isLoading: false,
                });
            },

            setMustChangePassword: (mustChangePassword: boolean) => {
                set({ mustChangePassword });
            },

            checkAuth: async () => {
                const { token, expireAt, isAPIKeyAuth } = get();

                if (!token) {
                    set({
                        isAuthenticated: false,
                        isLoading: false,
                        isAPIKeyAuth: false,
                        token: null,
                        expireAt: null,
                        mustChangePassword: false,
                        apiKeyStatus: null,
                    });
                    return;
                }

                if (!isAPIKeyAuth) {
                    if (!expireAt || Date.now() >= new Date(expireAt).getTime()) {
                        get().logout();
                        return;
                    }
                }

                try {
                    const endpoint = isAPIKeyAuth ? '/api/v1/apikey/login' : '/api/v1/user/status';
                    const status = await apiClient.get<unknown>(endpoint);
                    if (isAPIKeyAuth) {
                        set({
                            isAuthenticated: true,
                            isLoading: false,
                            isAPIKeyAuth: true,
                            mustChangePassword: false,
                            apiKeyStatus: status as APIKeyAuthStatus,
                        });
                        return;
                    }

                    const mustChangePassword = !!status
                        && typeof status === 'object'
                        && 'must_change_password' in status
                        && Boolean(status.must_change_password);
                    set({
                        isAuthenticated: true,
                        isLoading: false,
                        isAPIKeyAuth: false,
                        mustChangePassword,
                        apiKeyStatus: null,
                    });
                } catch (error) {
                    logger.error('璁よ瘉楠岃瘉澶辫触:', error);
                    get().logout();
                }
            },

            logout: () => {
                set({
                    isAuthenticated: false,
                    isAPIKeyAuth: false,
                    token: null,
                    expireAt: null,
                    mustChangePassword: false,
                    apiKeyStatus: null,
                    isLoading: false,
                });
            },
        }),
        {
            name: 'auth-storage',
            partialize: (state) => ({
                token: state.token,
                expireAt: state.expireAt,
                isAPIKeyAuth: state.isAPIKeyAuth,
                mustChangePassword: state.mustChangePassword,
                apiKeyStatus: state.apiKeyStatus,
            }),
        },
    ),
);

if (typeof window !== 'undefined') {
    setAuthStoreGetter(() => {
        const state = useAuthStore.getState();
        return {
            token: state.token,
            logout: state.logout,
        };
    });
}

export function useLogin() {
    const { setAuth } = useAuthStore();

    return useMutation({
        mutationFn: async (data: UserLoginRequest) => apiClient.post<UserLoginResponse>('/api/v1/user/login', data),
        onSuccess: (data) => {
            setAuth(data.token, data.expire_at, data.must_change_password);
        },
        onError: (error) => {
            logger.error('鐧诲綍澶辫触:', error);
        },
    });
}

export function useChangePassword() {
    return useMutation({
        mutationFn: async (data: { oldPassword: string; newPassword: string }) => {
            const payload: ChangePasswordRequest = {
                old_password: data.oldPassword,
                new_password: data.newPassword,
            };
            return apiClient.post<string>('/api/v1/user/change-password', payload);
        },
        onSuccess: (message) => {
            logger.log('瀵嗙爜淇敼鎴愬姛:', message);
        },
        onError: (error) => {
            logger.error('瀵嗙爜淇敼澶辫触:', error);
        },
    });
}

export function useForceChangePassword() {
    const { setMustChangePassword } = useAuthStore();

    return useMutation({
        mutationFn: async (data: { newUsername?: string; newPassword: string }) => {
            const payload: ForceChangePasswordRequest = {
                new_username: data.newUsername?.trim() || undefined,
                new_password: data.newPassword,
            };
            return apiClient.post<UserStatusResponse>('/api/v1/user/force-change-password', payload);
        },
        onSuccess: () => {
            setMustChangePassword(false);
        },
        onError: (error) => {
            logger.error('棣栨鏀瑰瘑澶辫触:', error);
        },
    });
}

export function useChangeUsername() {
    return useMutation({
        mutationFn: async (data: { newUsername: string; currentPassword: string }) => {
            const payload: ChangeUsernameRequest = {
                current_password: data.currentPassword,
                new_username: data.newUsername,
            };
            return apiClient.post<string>('/api/v1/user/change-username', payload);
        },
        onSuccess: (message) => {
            logger.log('鐢ㄦ埛鍚嶄慨鏀规垚鍔?', message);
        },
        onError: (error) => {
            logger.error('鐢ㄦ埛鍚嶄慨鏀瑰け璐?', error);
        },
    });
}

export function useRotateJWTSecret() {
    const { logout } = useAuthStore();
    return useMutation({
        mutationFn: async () => apiClient.post<string>('/api/v1/user/rotate-secret', {}),
        onSuccess: () => {
            setTimeout(() => logout(), 1500);
        },
        onError: (error) => {
            logger.error('JWT secret rotation failed:', error);
        },
    });
}

export function useAuth() {
    const store = useAuthStore();
    const { checkAuth, isLoading } = store;

    useEffect(() => {
        if (isLoading) {
            checkAuth();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return {
        isAuthenticated: store.isAuthenticated,
        isAPIKeyAuth: store.isAPIKeyAuth,
        isLoading: store.isLoading,
        mustChangePassword: store.mustChangePassword,
        apiKeyStatus: store.apiKeyStatus,
        logout: store.logout,
    };
}
