import { SakikoService } from "./sakikoService";
import { cacheLanguage, getCachedLanguage, resolveLanguage } from "./i18n";

export const preloadLanguage = async () => {
  const cachedLanguage = getCachedLanguage();
  if (cachedLanguage) {
    return cachedLanguage;
  }

  try {
    const settings = await SakikoService.GetAppSettings();
    if (settings?.language) {
      const resolved = resolveLanguage(settings.language);
      cacheLanguage(resolved);
      return resolved;
    }
  } catch {
    // Ignore preload errors and fall back to navigator language.
  }

  const browserLanguage = resolveLanguage(
    typeof navigator !== "undefined" ? navigator.language : undefined,
  );
  cacheLanguage(browserLanguage);
  return browserLanguage;
};
