import type { ImportForm, TaskPreset } from "../types/dashboard";

type ImportProfilePayload = {
  name: string;
  source: string;
  content?: string;
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
  config: TaskConfigShape;
};

export function createImportProfilePayload(importForm: ImportForm): ImportProfilePayload {
  return {
    name: importForm.name.trim(),
    source: importForm.source.trim(),
    content: importForm.content.trim(),
  };
}

export function createSubmitProfileTaskPayload(
  activeProfileId: string,
  taskPreset: TaskPreset,
  taskConfig: TaskConfigShape,
): SubmitProfileTaskPayload {
  return {
    profileId: activeProfileId,
    preset: taskPreset,
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
