export type ImportForm = {
  name: string;
  source: string;
  content: string;
};

export type TaskPreset = "ping" | "geo" | "speed" | "media" | "full";
