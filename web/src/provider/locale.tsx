'use client';

import { type ReactNode } from 'react';
import { NextIntlClientProvider } from 'next-intl';
import { useSettingStore, type Locale } from '@/stores/setting';

import zh_HansMessages from '../../public/locale/zh-Hans.json';
import zh_HantMessages from '../../public/locale/zh-Hant.json';
import enMessages from '../../public/locale/en.json';
import jaMessages from '../../public/locale/ja.json';

type IntlMessages = Record<string, unknown>;

function isPlainObject(value: unknown): value is IntlMessages {
	return !!value && typeof value === 'object' && !Array.isArray(value);
}

function mergeMessages(base: IntlMessages, override: IntlMessages): IntlMessages {
	const result: IntlMessages = { ...base };
	for (const [key, value] of Object.entries(override)) {
		const current = result[key];
		if (isPlainObject(current) && isPlainObject(value)) {
			result[key] = mergeMessages(current, value);
			continue;
		}
		result[key] = value;
	}
	return result;
}

const zhHansBase = zh_HansMessages as IntlMessages;
const enBase = mergeMessages(zhHansBase, enMessages as IntlMessages);

const messages: Record<Locale, IntlMessages> = {
	'zh-Hans': zhHansBase,
	'zh-Hant': mergeMessages(zhHansBase, zh_HantMessages as IntlMessages),
	en: enBase,
	ja: mergeMessages(enBase, jaMessages as IntlMessages),
};

export function LocaleProvider({ children }: { children: ReactNode }) {
	const { locale } = useSettingStore();

	return (
		<NextIntlClientProvider
			locale={locale}
			messages={messages[locale]}
			timeZone="Asia/Shanghai"
		>
			{children}
		</NextIntlClientProvider>
	);
}
