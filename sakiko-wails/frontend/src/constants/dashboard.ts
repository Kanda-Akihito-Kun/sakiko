import type { ImportForm, TaskPreset, TaskPresetSelection } from "../types/dashboard";
import type { TaskConfig } from "../types/sakiko";

export const DEFAULT_DOWNLOAD_THREADING = 4;
export const DEFAULT_BACKEND_IDENTITY = "Sakiko SpeedTest";
export const MAX_BACKEND_IDENTITY_RUNES = 30;

export function sanitizeBackendIdentity(value?: string): string {
  return Array.from((value || "").trim()).slice(0, MAX_BACKEND_IDENTITY_RUNES).join("");
}

export function resolveBackendIdentity(value?: string): string {
  return sanitizeBackendIdentity(value) || DEFAULT_BACKEND_IDENTITY;
}

export const initialImportForm: ImportForm = {
  name: "",
  source: "",
};

export const initialTaskConfig: TaskConfig = {
  pingAddress: "https://cp.cloudflare.com/generate_204",
  pingAverageOver: 2,
  taskRetry: 2,
  taskTimeoutMillis: 6000,
  downloadURL: "https://speed.cloudflare.com/__down?bytes=10000000",
  downloadDuration: 10,
  downloadThreading: DEFAULT_DOWNLOAD_THREADING,
  backendIdentity: DEFAULT_BACKEND_IDENTITY,
};

export const taskPresets: TaskPreset[] = ["full", "ping", "geo", "udp", "speed", "media"];
export const taskPresetChildren: TaskPreset[] = ["ping", "geo", "udp", "speed", "media"];
export const initialTaskPresetSelection: TaskPresetSelection = ["full", "ping", "geo", "udp", "speed", "media"];
