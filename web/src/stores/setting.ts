import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Locale = 'zh-Hans' | 'zh-Hant' | 'en' | 'ja';

const VALID_LOCALES: Locale[] = ['zh-Hans', 'zh-Hant', 'en', 'ja'];

interface SettingState {
    locale: Locale;
    setLocale: (locale: Locale) => void;
    hotkeyEnabled: boolean;
    setHotkeyEnabled: (enabled: boolean) => void;
}

export const useSettingStore = create<SettingState>()(
    persist(
        (set) => ({
            locale: 'zh-Hans',
            setLocale: (locale) => set({ locale }),
            hotkeyEnabled: true,
            setHotkeyEnabled: (enabled) => set({ hotkeyEnabled: enabled }),
        }),
        {
            name: 'octopus-settings',
            version: 2,
            migrate: (persistedState: unknown) => {
                const state = persistedState as Partial<SettingState>;
                return {
                    locale: VALID_LOCALES.includes(state?.locale as Locale) ? state.locale! : 'zh-Hans',
                    hotkeyEnabled: typeof state?.hotkeyEnabled === 'boolean' ? state.hotkeyEnabled : true,
                };
            },
        }
    )
);
