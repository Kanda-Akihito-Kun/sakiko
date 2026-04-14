import { create } from "zustand";

export type NotificationLevel = "warning" | "error";

export type NotificationSource = "backend" | "workspace" | "frontend";

export type NotificationInput = {
  level: NotificationLevel | string;
  message: string;
  source?: string;
  timestamp?: string;
  channel?: NotificationSource;
};

export type NotificationItem = {
  id: string;
  level: NotificationLevel;
  message: string;
  source: string;
  timestamp: string;
  channel: NotificationSource;
  count: number;
};

type NotificationState = {
  items: NotificationItem[];
  push: (input: NotificationInput) => void;
  dismiss: (id: string) => void;
};

const MAX_NOTIFICATION_COUNT = 4;
const DEDUPE_WINDOW_MILLIS = 3000;

export const useNotificationStore = create<NotificationState>((set) => ({
  items: [],
  push: (input) => {
    const next = normalizeNotification(input);
    if (!next) {
      return;
    }

    set((state) => {
      const duplicateIndex = state.items.findIndex((item) => (
        item.level === next.level
        && messagesEquivalent(item.message, next.message)
        && Math.abs(Date.parse(item.timestamp) - Date.parse(next.timestamp)) <= DEDUPE_WINDOW_MILLIS
      ));

      if (duplicateIndex >= 0) {
        const items = state.items.slice();
        const duplicate = items[duplicateIndex];
        items[duplicateIndex] = {
          ...duplicate,
          count: duplicate.count + 1,
          timestamp: next.timestamp,
        };
        return { items };
      }

      return {
        items: [next, ...state.items].slice(0, MAX_NOTIFICATION_COUNT),
      };
    });
  },
  dismiss: (id) => set((state) => ({
    items: state.items.filter((item) => item.id !== id),
  })),
}));

function normalizeNotification(input: NotificationInput): NotificationItem | null {
  const message = String(input.message || "").trim();
  if (!message) {
    return null;
  }

  const level = normalizeLevel(input.level);
  if (!level) {
    return null;
  }

  const timestamp = normalizeTimestamp(input.timestamp);
  return {
    id: createNotificationId(),
    level,
    message,
    source: normalizeSource(input.source, input.channel),
    timestamp,
    channel: input.channel || "backend",
    count: 1,
  };
}

function normalizeLevel(value: string): NotificationLevel | "" {
  switch (String(value || "").trim().toLowerCase()) {
    case "warn":
    case "warning":
      return "warning";
    case "error":
    case "fatal":
    case "panic":
      return "error";
    default:
      return "";
  }
}

function normalizeSource(source?: string, channel?: NotificationSource): string {
  const normalized = String(source || "").trim();
  if (normalized) {
    return normalized;
  }

  switch (channel) {
    case "workspace":
      return "workspace";
    case "frontend":
      return "frontend";
    default:
      return "backend";
  }
}

function normalizeTimestamp(value?: string): string {
  const normalized = String(value || "").trim();
  if (!normalized) {
    return new Date().toISOString();
  }

  const parsed = new Date(normalized);
  return Number.isNaN(parsed.getTime()) ? new Date().toISOString() : parsed.toISOString();
}

function createNotificationId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

function messagesEquivalent(left: string, right: string): boolean {
  const normalizedLeft = left.trim().toLowerCase();
  const normalizedRight = right.trim().toLowerCase();
  if (!normalizedLeft || !normalizedRight) {
    return false;
  }
  return normalizedLeft === normalizedRight
    || normalizedLeft.includes(normalizedRight)
    || normalizedRight.includes(normalizedLeft);
}
