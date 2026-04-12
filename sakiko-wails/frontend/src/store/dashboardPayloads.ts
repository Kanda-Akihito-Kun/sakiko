import type { ImportForm, TaskPresetSelection } from "../types/dashboard";

type ImportProfilePayload = {
  name: string;
  source: string;
};

type TaskConfigShape = {
  pingAddress: string;
  pingAverageOver: number;
  taskRetry: number;
  taskTimeoutMillis: number;
  downloadURL: string;
  downloadDuration: number;
  downloadThreading: number;
};

type SubmitProfileTaskPayload = {
  profileId: string;
  preset: string;
  presets: string[];
  config: TaskConfigShape;
};

export function createImportProfilePayload(importForm: ImportForm): ImportProfilePayload {
  return {
    name: importForm.name.trim(),
    source: importForm.source.trim(),
  };
}

export function createSubmitProfileTaskPayload(
  activeProfileId: string,
  taskPreset: TaskPresetSelection,
  taskConfig: TaskConfigShape,
): SubmitProfileTaskPayload {
  const normalizedPresets = taskPreset.filter((preset) => preset !== "full");
  const presetLabel = normalizedPresets.length === 4
    ? "full"
    : normalizedPresets.join("+");

  return {
    profileId: activeProfileId,
    preset: presetLabel,
    presets: normalizedPresets,
    config: {
      pingAddress: taskConfig.pingAddress,
      pingAverageOver: taskConfig.pingAverageOver,
      taskRetry: taskConfig.taskRetry,
      taskTimeoutMillis: taskConfig.taskTimeoutMillis,
      downloadURL: taskConfig.downloadURL,
      downloadDuration: taskConfig.downloadDuration,
      downloadThreading: taskConfig.downloadThreading,
    },
  };
}
