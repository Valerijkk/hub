/**
 * Инициализация i18next с поддержкой русского и английского языков.
 *
 * A13-016: `common` + `help` eager (login/shell shortcuts/tips); остальные
 * namespaces — dynamic import() по требованию useTranslation / t(..., { ns }).
 *
 * Язык: localStorage -> navigator, fallback ru; сохраняется в localStorage.
 */
import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import enCommon from "./locales/en/common.json";
import enHelp from "./locales/en/help.json";
import ruCommon from "./locales/ru/common.json";
import ruHelp from "./locales/ru/help.json";

const NAMESPACES = [
	"common",
	"tracker",
	"messenger",
	"tms",
	"calls",
	"dashboard",
	"settings",
	"help",
	"onboarding",
	"mail",
] as const;

type HubNamespace = (typeof NAMESPACES)[number];

const localeLoaders: Record<
	string,
	Partial<
		Record<HubNamespace, () => Promise<{ default: Record<string, unknown> }>>
	>
> = {
	ru: {
		tracker: () => import("./locales/ru/tracker.json"),
		messenger: () => import("./locales/ru/messenger.json"),
		tms: () => import("./locales/ru/tms.json"),
		calls: () => import("./locales/ru/calls.json"),
		dashboard: () => import("./locales/ru/dashboard.json"),
		settings: () => import("./locales/ru/settings.json"),
		onboarding: () => import("./locales/ru/onboarding.json"),
		mail: () => import("./locales/ru/mail.json"),
	},
	en: {
		tracker: () => import("./locales/en/tracker.json"),
		messenger: () => import("./locales/en/messenger.json"),
		tms: () => import("./locales/en/tms.json"),
		calls: () => import("./locales/en/calls.json"),
		dashboard: () => import("./locales/en/dashboard.json"),
		settings: () => import("./locales/en/settings.json"),
		onboarding: () => import("./locales/en/onboarding.json"),
		mail: () => import("./locales/en/mail.json"),
	},
};

/** Minimal backend: dynamic import per language/namespace (no http-backend dep). */
const dynamicLocaleBackend = {
	type: "backend" as const,
	init() {},
	read(
		language: string,
		namespace: string,
		callback: (
			err: Error | null,
			data: Record<string, unknown> | boolean,
		) => void,
	) {
		const lang = language.startsWith("en") ? "en" : "ru";
		if (namespace === "common") {
			callback(null, lang === "en" ? enCommon : ruCommon);
			return;
		}
		if (namespace === "help") {
			callback(null, lang === "en" ? enHelp : ruHelp);
			return;
		}
		const loader = localeLoaders[lang]?.[namespace as HubNamespace];
		if (!loader) {
			callback(new Error(`Unknown i18n ns: ${lang}/${namespace}`), false);
			return;
		}
		loader()
			.then((mod) =>
				callback(null, (mod.default ?? mod) as Record<string, unknown>),
			)
			.catch((err: Error) => callback(err, false));
	},
};

void i18n
	.use(LanguageDetector)
	.use(dynamicLocaleBackend)
	.use(initReactI18next)
	.init({
		fallbackLng: "ru",
		defaultNS: "common",
		ns: ["common", "help"],
		partialBundledLanguages: true,
		resources: {
			ru: { common: ruCommon, help: ruHelp },
			en: { common: enCommon, help: enHelp },
		},
		detection: {
			order: ["localStorage", "navigator"],
			caches: ["localStorage"],
			lookupLocalStorage: "i18nextLng",
		},
		interpolation: {
			escapeValue: false,
		},
		react: {
			useSuspense: false,
		},
	});

export default i18n;
