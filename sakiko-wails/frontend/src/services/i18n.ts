import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import type { AppLanguage } from "../types/appSettings";

export const supportedLanguages: AppLanguage[] = ["en", "zh"];
export const FALLBACK_LANGUAGE: AppLanguage = "en";

const LANGUAGE_STORAGE_KEY = "sakiko.language";

const normalizeLanguage = (language?: string) =>
  language?.toLowerCase().replace(/_/g, "-");

export const resolveLanguage = (language?: string): AppLanguage => {
  const normalized = normalizeLanguage(language);
  if (!normalized) {
    return FALLBACK_LANGUAGE;
  }

  if (normalized === "zh-cn" || normalized === "zh-hans") {
    return "zh";
  }

  if (normalized === "en" || normalized.startsWith("en-")) {
    return "en";
  }

  if (normalized === "zh" || normalized.startsWith("zh-")) {
    return "zh";
  }

  return FALLBACK_LANGUAGE;
};

const getLanguageStorage = () => {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    return window.localStorage;
  } catch {
    return null;
  }
};

export const cacheLanguage = (language: string) => {
  const storage = getLanguageStorage();
  if (!storage) {
    return;
  }
  storage.setItem(LANGUAGE_STORAGE_KEY, resolveLanguage(language));
};

export const getCachedLanguage = (): AppLanguage | undefined => {
  const storage = getLanguageStorage();
  if (!storage) {
    return undefined;
  }

  const cached = storage.getItem(LANGUAGE_STORAGE_KEY);
  return cached ? resolveLanguage(cached) : undefined;
};

type LocaleModule = {
  default: Record<string, unknown>;
};

const localeModules = import.meta.glob<LocaleModule>("../locales/*/index.ts");

const localeLoaders = Object.entries(localeModules).reduce<Record<string, () => Promise<LocaleModule>>>(
  (acc, [path, loader]) => {
    const match = path.match(/[/\\]locales[/\\]([^/\\]+)[/\\]index\.ts$/);
    if (match) {
      acc[match[1]] = loader;
    }
    return acc;
  },
  {},
);

export const loadLanguage = async (language: AppLanguage) => {
  const loader = localeLoaders[language];
  if (!loader) {
    throw new Error(`Locale loader not found for language "${language}"`);
  }
  const module = await loader();
  return module.default;
};

i18n.use(initReactI18next).init({
  resources: {},
  lng: FALLBACK_LANGUAGE,
  fallbackLng: FALLBACK_LANGUAGE,
  interpolation: {
    escapeValue: false,
  },
});

export const changeLanguage = async (language: string) => {
  const targetLanguage = resolveLanguage(language);
  if (!i18n.hasResourceBundle(targetLanguage, "translation")) {
    const resources = await loadLanguage(targetLanguage);
    i18n.addResourceBundle(targetLanguage, "translation", resources);
  }

  await i18n.changeLanguage(targetLanguage);
  cacheLanguage(targetLanguage);
  return targetLanguage;
};

export const initializeLanguage = async (initialLanguage: string = FALLBACK_LANGUAGE) => {
  await changeLanguage(initialLanguage);
};

export function translate(key: string, fallback: string, options?: Record<string, unknown>) {
  if (!i18n.isInitialized) {
    return fallback;
  }
  return i18n.t(key, { defaultValue: fallback, ...options });
}

export default i18n;
