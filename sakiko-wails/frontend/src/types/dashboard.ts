export type ImportForm = {
  name: string;
  source: string;
};

export type TaskPreset = "ping" | "geo" | "udp" | "speed" | "media" | "full";

export type TaskPresetSelection = TaskPreset[];
