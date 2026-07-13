import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Locale = 'zh-Hans' | 'zh-Hant' | 'en' | 'ja';

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
        }
    )
);

