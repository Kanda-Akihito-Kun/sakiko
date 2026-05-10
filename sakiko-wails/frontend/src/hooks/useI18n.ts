import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { SakikoService } from "../services/sakikoService";
import { changeLanguage, resolveLanguage, supportedLanguages } from "../services/i18n";

export function useI18n() {
  const { i18n, t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  const switchLanguage = useCallback(async (language: string) => {
    const targetLanguage = resolveLanguage(language);
    if (i18n.language === targetLanguage) {
      return;
    }

    setIsLoading(true);
    try {
      await changeLanguage(targetLanguage);
      await SakikoService.UpdateAppSettings({ language: targetLanguage });
    } finally {
      setIsLoading(false);
    }
  }, [i18n.language]);

  return {
    currentLanguage: resolveLanguage(i18n.language),
    supportedLanguages,
    switchLanguage,
    isLoading,
    t,
  };
}
