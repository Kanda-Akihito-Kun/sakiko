import { APP_VERSION_VALUE } from "./version.generated";

const normalizedVersion = APP_VERSION_VALUE.trim();

export const APP_VERSION = normalizedVersion.startsWith("v")
  ? normalizedVersion
  : `v${normalizedVersion}`;
