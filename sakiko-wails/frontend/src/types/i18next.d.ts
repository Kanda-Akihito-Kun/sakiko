import "i18next";
import resources from "../locales/en";

declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "translation";
    resources: typeof resources;
  }
}
