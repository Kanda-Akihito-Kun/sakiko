import type { DownloadTarget, Profile, ResultArchiveTask } from "../types/sakiko";
import { filterVisibleMediaUnlockItems } from "./mediaUnlock";

export type FilteredProfileNode = {
  index: number;
  node: Profile["nodes"][number];
};

export function normalizeError(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }

  return String(err);
}

export function getFilteredNodes(activeProfile: Profile | null, nodeFilter: string) {
  const keyword = nodeFilter.trim().toLowerCase();
  if (!activeProfile) {
    return [];
  }

  if (!keyword) {
    return activeProfile.nodes.map((node, index) => ({ node, index }));
  }

  return activeProfile.nodes.flatMap((node, index) => (
    [
      node.name,
      node.protocol || "",
      node.server || "",
      node.port || "",
      node.udp === true ? "udp" : node.udp === false ? "no udp" : "",
    ].some((value) => value.toLowerCase().includes(keyword))
      ? [{ node, index }]
      : []
  ));
}

export function formatMatrixPayload(payload: unknown, type?: string): string {
  if (typeof payload === "number") {
    if (isTrafficMatrix(type)) {
      return formatMegabytes(payload);
    }
    if (isSpeedMatrix(type)) {
      return formatMbps(payload);
    }
    return `${payload}`;
  }

  if (typeof payload === "string") {
    return payload;
  }

  if (Array.isArray(payload)) {
    return payload.join(", ");
  }

  if (payload && typeof payload === "object") {
    const record = payload as Record<string, unknown>;

    if (typeof record.ip === "string") {
      const address = typeof record.address === "string" ? record.address : "";
      const asn = typeof record.asn === "number" && record.asn > 0 ? `AS${record.asn}` : "";
      const org = typeof record.asOrganization === "string" ? record.asOrganization : "";
      const country = typeof record.country === "string" ? record.country : "";
      const city = typeof record.city === "string" ? record.city : "";
      const countryCode = typeof record.countryCode === "string" ? record.countryCode : "";
      const error = typeof record.error === "string" ? record.error : "";
      const location = [formatCountryCodeWithFlag(countryCode), country, city].filter(Boolean).join(" ");
      const endpoint = address && address !== record.ip ? `${address} -> ${record.ip}` : `${record.ip}`;
      const parts = [endpoint, asn, org, location].filter(Boolean);
      return error ? `${parts.join(" | ")} | ${error}` : parts.join(" | ");
    }

    if (typeof record.value === "number" || typeof record.value === "string") {
      if (typeof record.value === "number" && isTrafficMatrix(type)) {
        return formatMegabytes(record.value);
      }
      if (typeof record.value === "number" && isSpeedMatrix(type)) {
        return formatMbps(record.value);
      }
      return `${record.value}`;
    }

    if (Array.isArray(record.values)) {
      if (isPerSecondSpeedMatrix(type)) {
        return record.values
          .map((value) => (typeof value === "number" ? formatMbps(value) : String(value)))
          .join(", ");
      }
      return record.values.join(", ");
    }

    if (Array.isArray(record.items)) {
      return (
        filterVisibleMediaUnlockItems(record.items)
          .map((item) => formatMediaUnlockItem(item))
          .filter(Boolean)
          .join(" | ")
      ) || "-";
    }
  }

  return JSON.stringify(payload);
}

export function formatDuration(duration: number): string {
  return `${duration} ms total`;
}

export function formatDateTime(value?: string): string {
  if (!value) {
    return "N/A";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

export function summarizeResultMetrics(preset?: string): string {
  switch ((preset || "").trim().toLowerCase()) {
    case "ping":
      return "TLS RTT / HTTPS Ping";
    case "geo":
      return "Inbound / Outbound Topology";
    case "speed":
      return "Average / Max / Per-second Speed / Traffic";
    case "media":
      return "Netflix / Hulu / Bilibili HMT";
    case "full":
      return "Ping / Topology / Speed / Traffic / Media";
    default:
      return "Archived metrics";
  }
}

export function formatProtocolLibraryLabel(vendor?: string): string {
  return (vendor || "").trim() || "unknown";
}

export function formatBackendLabel(task?: Pick<ResultArchiveTask, "environment">): string {
  const backend = task?.environment?.backend;
  if (!backend) {
    return "Unknown backend";
  }

  const location = (backend.location || "").trim();

  if (location) {
    return location;
  }

  return "Unknown backend";
}

export function summarizeDownloadTarget(downloadURL?: string, downloadTargets: DownloadTarget[] = []): string {
  const trimmedURL = (downloadURL || "").trim();
  if (!trimmedURL) {
    return "Target unavailable";
  }

  const matched = downloadTargets.find((target) => target.downloadURL === trimmedURL);
  if (matched) {
    return formatDownloadTargetLabel(matched);
  }

  try {
    const url = new URL(trimmedURL);
    if (url.host === "speed.cloudflare.com") {
      return "Cloudflare Default";
    }
    return url.host || "Custom target";
  } catch {
    return "Custom target";
  }
}

export function summarizeDownloadTargetDetail(downloadURL?: string, downloadTargets: DownloadTarget[] = []): string {
  const trimmedURL = (downloadURL || "").trim();
  if (!trimmedURL) {
    return "N/A";
  }

  const matched = downloadTargets.find((target) => target.downloadURL === trimmedURL);
  if (matched) {
    return matched.host || matched.endpoint || matched.downloadURL;
  }

  try {
    const url = new URL(trimmedURL);
    return url.host || trimmedURL;
  } catch {
    return trimmedURL;
  }
}

export function formatReportValue(value: unknown, key?: string): string {
  if (value === null || value === undefined || value === "") {
    return "-";
  }

  if ((key || "") === "status") {
    return formatMediaStatus(value);
  }

  if ((key || "") === "unlockMode") {
    return formatUnlockMode(value);
  }

  if (isCountryCodeField(key)) {
    return formatCountryCodeWithFlag(value);
  }

  if (typeof value === "number") {
    if ((key || "").toLowerCase().includes("trafficusedbytes")) {
      return formatMegabytes(value);
    }
    if ((key || "").toLowerCase().includes("bytespersecond")) {
      return formatMbps(value);
    }
    if ((key || "").toLowerCase().includes("millis")) {
      return `${value} ms`;
    }
    return `${value}`;
  }

  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }

  if (Array.isArray(value)) {
    if ((key || "").toLowerCase().includes("bytespersecond")) {
      return value.map((item) => (typeof item === "number" ? formatMbps(item) : String(item))).join(", ");
    }
    return value.map((item) => String(item)).join(", ");
  }

  if (typeof value === "object") {
    return JSON.stringify(value);
  }

  return String(value);
}

function isSpeedMatrix(type?: string): boolean {
  return type === "SPEED_AVERAGE"
    || type === "SPEED_MAX"
    || type === "SPEED_PER_SECOND";
}

function isPerSecondSpeedMatrix(type?: string): boolean {
  return type === "SPEED_PER_SECOND";
}

function isTrafficMatrix(type?: string): boolean {
  return type === "SPEED_TRAFFIC_USED";
}

function formatMediaUnlockItem(value: unknown): string {
  if (!value || typeof value !== "object") {
    return String(value || "");
  }

  const item = value as Record<string, unknown>;
  const display = String(item.display || "").trim();
  if (display) {
    return display;
  }
  const name = String(item.name || item.platform || "Unknown").trim();
  const status = String(item.status || "unknown").trim();
  const region = String(item.region || "").trim();
  const mode = String(item.mode || "").trim();
  const error = String(item.error || "").trim();

  const parts = [name, formatMediaStatus(status)];
  if (region) {
    parts.push(region);
  }
  if (mode) {
    parts.push(formatUnlockMode(mode));
  }
  if (error) {
    parts.push(error);
  }
  return parts.filter(Boolean).join(": ");
}

function formatMediaStatus(value: unknown): string {
  switch (String(value || "").trim().toLowerCase()) {
    case "yes":
      return "Yes";
    case "no":
      return "No";
    case "originals_only":
      return "Originals Only";
    case "web_only":
      return "Web Only";
    case "oversea_only":
      return "Oversea Only";
    case "unsupported":
      return "Unsupported";
    case "failed":
      return "Failed";
    default:
      return String(value || "-");
  }
}

function formatUnlockMode(value: unknown): string {
  switch (String(value || "").trim().toLowerCase()) {
    case "native":
      return "Native";
    case "dns":
      return "DNS";
    case "unknown":
      return "Unknown";
    default:
      return String(value || "-");
  }
}

function formatMbps(bytesPerSecond: number): string {
  const mbps = (bytesPerSecond * 8) / 1_000_000;
  return `${mbps.toFixed(2)} Mbps`;
}

function formatMegabytes(bytes: number): string {
  const megabytes = bytes / 1_000_000;
  return `${megabytes.toFixed(2)} MB`;
}

function formatDownloadTargetLabel(target: DownloadTarget): string {
  const location = [target.city, target.country].filter(Boolean).join(", ");
  const primary = (target.name || "").trim() || (target.host || "").trim() || "Speed Target";
  return location ? `${primary} 路 ${location}` : primary;
}

function isCountryCodeField(key?: string): boolean {
  return (key || "").toLowerCase().includes("countrycode");
}

function formatCountryCodeWithFlag(value: unknown): string {
  const code = String(value || "").trim().toUpperCase();
  if (!code) {
    return "-";
  }
  if (!/^[A-Z]{2}$/.test(code)) {
    return code;
  }

  const flag = Array.from(code)
    .map((char) => String.fromCodePoint(127397 + char.charCodeAt(0)))
    .join("");
  return `${flag} ${code}`;
}
