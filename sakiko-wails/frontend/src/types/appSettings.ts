export type AppLanguage = "en" | "zh";

export type AppSettings = {
  language: AppLanguage;
};

export type AppSettingsPatch = Partial<AppSettings>;
