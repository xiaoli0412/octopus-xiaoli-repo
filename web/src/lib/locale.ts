import { enUS, ja, zhCN, zhTW } from 'date-fns/locale';
import type { Locale as DateFnsLocale } from 'date-fns';
import type { Locale as AppLocale } from '@/stores/setting';

const intlLocaleMap: Record<AppLocale, string> = {
	'zh-Hans': 'zh-CN',
	'zh-Hant': 'zh-TW',
	en: 'en-US',
	ja: 'ja-JP',
};

const dateFnsLocaleMap: Record<AppLocale, DateFnsLocale> = {
	'zh-Hans': zhCN,
	'zh-Hant': zhTW,
	en: enUS,
	ja,
};

const localeOrder: AppLocale[] = ['zh-Hans', 'zh-Hant', 'en', 'ja'];

export function getIntlLocale(locale: AppLocale): string {
	return intlLocaleMap[locale];
}

export function getDateFnsLocale(locale: AppLocale): DateFnsLocale {
	return dateFnsLocaleMap[locale];
}

export function getNextLocale(locale: AppLocale): AppLocale {
	const currentIndex = localeOrder.indexOf(locale);
	if (currentIndex < 0) return localeOrder[0];
	return localeOrder[(currentIndex + 1) % localeOrder.length];
}

export function formatNumberByLocale(
	value: number,
	locale: AppLocale,
	options?: Intl.NumberFormatOptions,
): string {
	return new Intl.NumberFormat(getIntlLocale(locale), options).format(value);
}

export function formatDateByLocale(
	value: Date | number | string,
	locale: AppLocale,
	options?: Intl.DateTimeFormatOptions,
): string {
	const date = value instanceof Date ? value : new Date(value);
	if (Number.isNaN(date.getTime())) {
		return String(value);
	}

	return new Intl.DateTimeFormat(getIntlLocale(locale), options).format(date);
}

export function formatDateTimeByLocale(value: Date | number | string, locale: AppLocale): string {
	return formatDateByLocale(value, locale, {
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
	});
}
