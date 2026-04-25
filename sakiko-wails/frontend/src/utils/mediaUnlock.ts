const hiddenMediaPlatformKeys = new Set(["dazn", "instagram_music", "hulu_jp", "spotify", "steam"]);
const fixedMediaColumnKeys = new Set(["nodeName", "proxyType"]);

export function isVisibleMediaPlatformKey(key: string): boolean {
  const normalized = key.trim().toLowerCase();
  if (!normalized) {
    return false;
  }
  return !hiddenMediaPlatformKeys.has(normalized);
}

export function filterVisibleMediaUnlockItems(items: unknown[]): Record<string, unknown>[] {
  return items.flatMap<Record<string, unknown>>((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    const platform = String(record.platform || "").trim();
    return isVisibleMediaPlatformKey(platform) ? [record] : [];
  });
}

export function filterMediaReportSectionColumns<T extends { key: string; label?: string }>(columns: T[]): T[] {
  return columns.filter((column) => fixedMediaColumnKeys.has(column.key) || isVisibleMediaPlatformKey(column.key));
}
