export type AppLanguage = "en" | "zh";

export type DNSConfig = {
  bootstrapServers: string[];
  resolverServers: string[];
};

export type AppSettings = {
  language: AppLanguage;
  dns: DNSConfig;
  hideProfileNameInExport: boolean;
  hideCNInboundInExport: boolean;
};

export type AppSettingsPatch = Partial<AppSettings>;
