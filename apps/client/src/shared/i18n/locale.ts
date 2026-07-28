import { getLocales } from "expo-localization";

import en from "./locales/en.json";
import es from "./locales/es.json";
import fr from "./locales/fr.json";
import it from "./locales/it.json";

const supportedLanguages = ["es", "en", "it", "fr"] as const;

export type SupportedLanguage = (typeof supportedLanguages)[number];
export type TranslationKey = keyof typeof en;

type TranslationCatalog = Record<TranslationKey, string>;

const catalogs: Record<SupportedLanguage, TranslationCatalog> = { en, es, fr, it };

export function getCurrentLanguage(): SupportedLanguage {
  const languageCode = getLocales()[0]?.languageCode;
  return isSupportedLanguage(languageCode) ? languageCode : "en";
}

export function getTranslator() {
  const catalog = catalogs[getCurrentLanguage()];
  return (key: TranslationKey) => catalog[key];
}

function isSupportedLanguage(
  languageCode: string | null | undefined,
): languageCode is SupportedLanguage {
  return supportedLanguages.some((language) => language === languageCode);
}
