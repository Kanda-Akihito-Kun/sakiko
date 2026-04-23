import type { EntryResult, ResultReportSection } from "../types/sakiko";
import { filterMediaReportSectionColumns, filterVisibleMediaUnlockItems, isVisibleMediaPlatformKey } from "./mediaUnlock";

export type MediaMatrixColumn = {
  key: string;
  label: string;
};

export type MediaMatrixRow = {
  nodeName: string;
  proxyType: string;
  cells: Record<string, string>;
};

const fixedColumnKeys = new Set(["nodeName", "proxyType"]);

export function buildMediaMatrixFromResults(results: EntryResult[] = []): {
  columns: MediaMatrixColumn[];
  rows: MediaMatrixRow[];
} {
  const columnMap = new Map<string, MediaMatrixColumn>();
  const rows = results.map<MediaMatrixRow>((result) => {
    const payload = result.matrices.find((matrix) => matrix.type === "MEDIA_UNLOCK")?.payload;
    const items = Array.isArray((payload as { items?: unknown[] } | undefined)?.items)
      ? filterVisibleMediaUnlockItems((payload as { items?: unknown[] }).items || [])
      : [];
    const cells: Record<string, string> = {};

    items.forEach((item) => {
      if (!item || typeof item !== "object") {
        return;
      }
      const record = item as Record<string, unknown>;
      const key = String(record.platform || "").trim();
      const label = String(record.name || key || "Unknown").trim();
      if (!isVisibleMediaPlatformKey(key)) {
        return;
      }
      if (!columnMap.has(key)) {
        columnMap.set(key, { key, label });
      }
      cells[key] = formatMediaCellValue(record);
    });

    return {
      nodeName: result.proxyInfo.name || "Unnamed node",
      proxyType: String(result.proxyInfo.type || "unknown"),
      cells,
    };
  });

  const columns = sortMediaMatrixColumns([...columnMap.values()]);
  rows.forEach((row) => {
    columns.forEach((column) => {
      if (!row.cells[column.key]) {
        row.cells[column.key] = "-";
      }
    });
  });
  return { columns, rows };
}

export function buildMediaMatrixFromSection(section?: ResultReportSection): {
  columns: MediaMatrixColumn[];
  rows: MediaMatrixRow[];
} {
  const columns = sortMediaMatrixColumns(
    filterMediaReportSectionColumns(section?.columns || [])
      .filter((column) => !fixedColumnKeys.has(column.key))
      .map((column) => ({ key: column.key, label: column.label })),
  );

  const rows = (section?.rows || []).map<MediaMatrixRow>((row) => {
    const record = row as Record<string, unknown>;
    const cells: Record<string, string> = {};
    columns.forEach((column) => {
      cells[column.key] = String(record[column.key] || "-");
    });
    return {
      nodeName: String(record.nodeName || "Unnamed node"),
      proxyType: String(record.proxyType || "unknown"),
      cells,
    };
  });

  return { columns, rows };
}

export function mediaCellTone(value: string): "success" | "warning" | "error" | "neutral" {
  const normalized = value.trim().toLowerCase();
  if (!normalized || normalized === "-") {
    return "neutral";
  }
  if (normalized.includes("failed") || normalized.includes("blocked") || normalized === "no" || normalized.startsWith("no ")) {
    return "error";
  }
  if (normalized.includes("originals only") || normalized.includes("web only") || normalized.includes("oversea only")) {
    return "warning";
  }
  if (normalized.includes("unsupported")) {
    return "neutral";
  }
  if (normalized.includes("douyin")) {
    return "warning";
  }
  if (normalized.includes("unlock") || normalized.includes("region") || normalized.includes("currency")) {
    return "success";
  }
  return "neutral";
}

function formatMediaCellValue(item: Record<string, unknown>): string {
  const display = String(item.display || "").trim();
  if (display) {
    return display;
  }

  const status = String(item.status || "").trim().toLowerCase();
  const region = String(item.region || "").trim();
  const error = String(item.error || "").trim();

  switch (status) {
    case "yes":
      return region ? `Yes (Region: ${region})` : "Yes";
    case "no":
      if (error && region) {
        return `No (${error};Region: ${region})`;
      }
      return error ? `No (${error})` : (region ? `No (Region: ${region})` : "No");
    case "originals_only":
      return region ? `Originals Only (Region: ${region})` : "Originals Only";
    case "web_only":
      return region ? `Web Only (Region: ${region})` : "Web Only";
    case "oversea_only":
      return region ? `Oversea Only (Region: ${region})` : "Oversea Only";
    case "unsupported":
      return region ? `Unsupported (Region: ${region})` : "Unsupported";
    case "failed":
      return error ? `Failed (${error})` : "Failed";
    default:
      return display || status || "-";
  }
}

function sortMediaMatrixColumns(columns: MediaMatrixColumn[]): MediaMatrixColumn[] {
  const preferredOrder = [
    "chatgpt",
    "claude",
    "gemini",
    "youtube_premium",
    "netflix",
    "hulu",
    "prime_video",
    "hbo_max",
    "bilibili_hmt",
    "bilibili_tw",
    "abema",
    "tiktok",
  ];
  const rank = new Map(preferredOrder.map((key, index) => [key, index]));
  return [...columns].sort((left, right) => {
    const leftRank = rank.get(left.key) ?? Number.MAX_SAFE_INTEGER;
    const rightRank = rank.get(right.key) ?? Number.MAX_SAFE_INTEGER;
    if (leftRank !== rightRank) {
      return leftRank - rightRank;
    }
    return left.label.localeCompare(right.label, "en");
  });
}
