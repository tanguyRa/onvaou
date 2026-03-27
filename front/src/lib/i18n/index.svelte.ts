import i18next from '../../../node_modules/i18next';
import en from './locales/en.json';
import fr from './locales/fr.json';
import es from './locales/es.json';
import cn from './locales/cn.json';
import { browser } from '$app/environment';

let currentLanguage = $state<string>('en');

const languages = [
	{ code: 'en', label: 'EN', translation: en },
	{ code: 'fr', label: 'FR', translation: fr },
	{ code: 'es', label: 'ES', translation: es },
	{ code: 'cn', label: '简体', translation: cn }
];

i18next.init({
	lng: browser ? (localStorage.getItem('preferredLanguage') || 'en') : 'en',
	fallbackLng: 'en',
	resources: languages.reduce(
		(acc, lang) => {
			acc[lang.code] = { translation: lang.translation };
			return acc;
		},
		{} as Record<string, { translation: Record<string, any> }>
	),
	interpolation: { escapeValue: false }
});

currentLanguage = i18next.language;

i18next.on('languageChanged', (lng: string) => {
	currentLanguage = lng;
});

export function changeLanguage(lng: string) {
	i18next.changeLanguage(lng);
	if (browser) {
		localStorage.setItem('preferredLanguage', lng);
	}
}

export function t(key: string): string {
	// Access currentLanguage to create reactive dependency
	void currentLanguage;
	return i18next.t(key);
}

export function tError(
	rawMessage: string | undefined | null,
	fallbackKey: string
): string {
	const message = (rawMessage || '').toLowerCase();
	void currentLanguage;

	if (!message) return i18next.t(fallbackKey);

	if (
		message.includes('failed to proxy request to go api') ||
		message.includes('bad gateway') ||
		message.includes('connectionrefused') ||
		message.includes('failedtoopensocket') ||
		message.includes('unable to connect')
	) {
		return i18next.t('common.errors.apiUnavailable');
	}

	if (message.includes('network') || message.includes('fetch')) {
		return i18next.t('common.errors.network');
	}

	return i18next.t(fallbackKey);
}

export function getLanguage(): string {
	return currentLanguage;
}

interface Language {
	code: string;
	label: string;
}
export function getLanguages(): Language[] {
	return languages.map((lang) => ({ code: lang.code, label: lang.label }));
}
