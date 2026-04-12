export type ImportForm = {
  name: string;
  source: string;
};

export type TaskPreset = "ping" | "geo" | "speed" | "media" | "full";

export type TaskPresetSelection = TaskPreset[];
