'use client';

import { useTheme } from 'next-themes';
import { useTranslations } from 'next-intl';
import { Sun, Moon, Monitor, Languages } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useSettingStore, type Locale } from '@/stores/setting';

export function SettingAppearance() {
    const t = useTranslations('setting');
    const { theme, setTheme } = useTheme();
    const { locale, setLocale } = useSettingStore();

    return (
        <div className="octo-setting-card">
            <h2 className="octo-setting-heading">
                <Sun className="size-4" />
                {t('appearance')}
            </h2>

            {/* 主题 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    {theme === 'dark' ? <Moon className="size-4 text-muted-foreground" /> : <Sun className="size-4 text-muted-foreground" />}
                    <span className="text-sm font-medium">{t('theme.label')}</span>
                </div>
                <Select value={theme} onValueChange={setTheme}>
                    <SelectTrigger className="h-9 w-full rounded-xl md:w-40">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                        <SelectItem value="light" className="rounded-xl">
                            <Sun className="size-4" />
                            {t('theme.light')}
                        </SelectItem>
                        <SelectItem value="dark" className="rounded-xl">
                            <Moon className="size-4" />
                            {t('theme.dark')}
                        </SelectItem>
                        <SelectItem value="system" className="rounded-xl">
                            <Monitor className="size-4" />
                            {t('theme.system')}
                        </SelectItem>
                    </SelectContent>
                </Select>
            </div>

            {/* 语言 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Languages className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('language.label')}</span>
                </div>
                <Select value={locale} onValueChange={(v) => setLocale(v as Locale)}>
                    <SelectTrigger className="h-9 w-full rounded-xl md:w-40">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                        <SelectItem value="zh-Hans" className="rounded-xl">{t('language.zh-Hans')}</SelectItem>
                        <SelectItem value="zh-Hant" className="rounded-xl">{t('language.zh-Hant')}</SelectItem>
                        <SelectItem value="en" className="rounded-xl">{t('language.en')}</SelectItem>
                        <SelectItem value="ja" className="rounded-xl">{t('language.ja')}</SelectItem>
                    </SelectContent>
                </Select>
            </div>
        </div>
    );
}

