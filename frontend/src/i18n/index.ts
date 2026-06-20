/**
 * Инициализация i18next с поддержкой русского и английского языков.
 *
 * Поддерживаемые namespace: common, tracker, messenger, tms, calls, dashboard, settings.
 * Язык определяется автоматически (localStorage -> navigator) с fallback на русский.
 * Сохраняется в localStorage при переключении.
 */
import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import enCalls from "./locales/en/calls.json";
import enCommon from "./locales/en/common.json";
import enDashboard from "./locales/en/dashboard.json";
import enMessenger from "./locales/en/messenger.json";
import enHelp from "./locales/en/help.json";
import enOnboarding from "./locales/en/onboarding.json";
import enSettings from "./locales/en/settings.json";
import enTms from "./locales/en/tms.json";
import enTracker from "./locales/en/tracker.json";
import ruCalls from "./locales/ru/calls.json";
import ruCommon from "./locales/ru/common.json";
import ruDashboard from "./locales/ru/dashboard.json";
import ruHelp from "./locales/ru/help.json";
import ruMessenger from "./locales/ru/messenger.json";
import ruOnboarding from "./locales/ru/onboarding.json";
import ruSettings from "./locales/ru/settings.json";
import ruTms from "./locales/ru/tms.json";
import ruTracker from "./locales/ru/tracker.json";

void i18n
	.use(LanguageDetector)
	.use(initReactI18next)
	.init({
		fallbackLng: "ru",
		defaultNS: "common",
		ns: [
			"common",
			"tracker",
			"messenger",
			"tms",
			"calls",
			"dashboard",
			"settings",
			"help",
			"onboarding",
		],
		resources: {
			ru: {
				common: ruCommon,
				tracker: ruTracker,
				messenger: ruMessenger,
				tms: ruTms,
				calls: ruCalls,
				dashboard: ruDashboard,
				settings: ruSettings,
				help: ruHelp,
				onboarding: ruOnboarding,
			},
			en: {
				common: enCommon,
				tracker: enTracker,
				messenger: enMessenger,
				tms: enTms,
				calls: enCalls,
				dashboard: enDashboard,
				settings: enSettings,
				help: enHelp,
				onboarding: enOnboarding,
			},
		},
		detection: {
			order: ["localStorage", "navigator"],
			caches: ["localStorage"],
			lookupLocalStorage: "i18nextLng",
		},
		interpolation: {
			escapeValue: false,
		},
	});

export default i18n;
