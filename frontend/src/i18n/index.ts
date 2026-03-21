/**
 * Инициализация i18next с поддержкой русского и английского языков.
 *
 * Поддерживаемые namespace: common, tracker, messenger, tms, calls, dashboard, settings.
 * Язык определяется автоматически (localStorage -> navigator) с fallback на русский.
 * Сохраняется в localStorage при переключении.
 */
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

import ruCommon from "./locales/ru/common.json";
import ruTracker from "./locales/ru/tracker.json";
import ruMessenger from "./locales/ru/messenger.json";
import ruTms from "./locales/ru/tms.json";
import ruCalls from "./locales/ru/calls.json";
import ruDashboard from "./locales/ru/dashboard.json";
import ruSettings from "./locales/ru/settings.json";

import enCommon from "./locales/en/common.json";
import enTracker from "./locales/en/tracker.json";
import enMessenger from "./locales/en/messenger.json";
import enTms from "./locales/en/tms.json";
import enCalls from "./locales/en/calls.json";
import enDashboard from "./locales/en/dashboard.json";
import enSettings from "./locales/en/settings.json";

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: "ru",
    defaultNS: "common",
    ns: ["common", "tracker", "messenger", "tms", "calls", "dashboard", "settings"],
    resources: {
      ru: {
        common: ruCommon,
        tracker: ruTracker,
        messenger: ruMessenger,
        tms: ruTms,
        calls: ruCalls,
        dashboard: ruDashboard,
        settings: ruSettings,
      },
      en: {
        common: enCommon,
        tracker: enTracker,
        messenger: enMessenger,
        tms: enTms,
        calls: enCalls,
        dashboard: enDashboard,
        settings: enSettings,
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
