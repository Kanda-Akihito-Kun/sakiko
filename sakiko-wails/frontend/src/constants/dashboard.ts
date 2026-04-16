import { TaskConfig } from "../../bindings/sakiko.local/sakiko-core/interfaces";
import type { ImportForm, TaskPreset, TaskPresetSelection } from "../types/dashboard";

export const initialImportForm: ImportForm = {
  name: "",
  source: "",
};

export const initialTaskConfig = new TaskConfig({
  pingAddress: "https://cp.cloudflare.com/generate_204",
  pingAverageOver: 2,
  taskRetry: 2,
  taskTimeoutMillis: 6000,
  downloadURL: "https://speed.cloudflare.com/__down?bytes=10000000",
  downloadDuration: 10,
  downloadThreading: 8,
  backendIdentity: "",
});

export const taskPresets: TaskPreset[] = ["full", "ping", "geo", "speed", "media"];
export const taskPresetChildren: TaskPreset[] = ["ping", "geo", "speed", "media"];
export const initialTaskPresetSelection: TaskPresetSelection = ["full", "ping", "geo", "speed", "media"];
