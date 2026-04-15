import type { DownloadTarget, Profile, ResultArchiveTask, TaskActiveNode } from "../types/sakiko";
import type { TaskPreset, TaskPresetSelection } from "../types/dashboard";
import { taskPresets, taskPresetChildren } from "../constants/dashboard";
import { translate } from "../services/i18n";
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
  return translate("shared.formats.durationMillisTotal", `${duration} ms total`, { value: duration });
}

export function formatMacroLabel(value?: string): string {
  switch ((value || "").trim().toUpperCase()) {
    case "PING":
      return translate("shared.macros.ping", "Ping");
    case "GEO":
      return translate("shared.macros.geo", "Geo");
    case "SPEED":
      return translate("shared.macros.speed", "Speed");
    case "MEDIA":
      return translate("shared.macros.media", "Media Unlock");
    default:
      return value || "-";
  }
}

export function formatMatrixLabel(value?: string): string {
  switch ((value || "").trim().toUpperCase()) {
    case "TEST_PING_CONN":
      return translate("shared.matrices.testPingConn", "HTTPS Ping");
    case "TEST_PING_RTT":
      return translate("shared.matrices.testPingRtt", "RTT Ping");
    case "GEOIP_INBOUND":
      return translate("shared.matrices.geoipInbound", "Inbound GeoIP");
    case "GEOIP_OUTBOUND":
      return translate("shared.matrices.geoipOutbound", "Outbound GeoIP");
    case "SPEED_AVERAGE":
      return translate("shared.matrices.speedAverage", "Average Speed");
    case "SPEED_MAX":
      return translate("shared.matrices.speedMax", "Max Speed");
    case "SPEED_PER_SECOND":
      return translate("shared.matrices.speedPerSecond", "Per-second Speed");
    case "SPEED_TRAFFIC_USED":
      return translate("shared.matrices.speedTrafficUsed", "Traffic Used");
    case "MEDIA_UNLOCK":
      return translate("shared.matrices.mediaUnlock", "Media Unlock");
    default:
      return value || "-";
  }
}

export function formatTaskRuntimePhase(value?: string): string {
  switch ((value || "").trim().toLowerCase()) {
    case "preparing":
      return translate("shared.phases.preparing", "Preparing");
    case "macro":
      return translate("shared.phases.running", "Running");
    case "matrix":
      return translate("shared.phases.extracting", "Extracting");
    default:
      return value || "-";
  }
}

export function formatTaskStatus(value?: string): string {
  switch ((value || "").trim().toLowerCase()) {
    case "queued":
      return translate("shared.statuses.queued", "Queued");
    case "pending":
      return translate("shared.statuses.pending", "Pending");
    case "running":
      return translate("shared.statuses.running", "Running");
    case "success":
      return translate("shared.statuses.success", "Success");
    case "failed":
      return translate("shared.statuses.failed", "Failed");
    case "canceled":
      return translate("shared.statuses.canceled", "Canceled");
    default:
      return value || "-";
  }
}

export function describeTaskActiveNode(activeNode: TaskActiveNode): string {
  const matrixLabels = (activeNode.matrices || []).map((matrix) => formatMatrixLabel(matrix));
  const matrixLabel = activeNode.matrix ? formatMatrixLabel(activeNode.matrix) : "";
  const targetLabel = matrixLabel || matrixLabels.join(" / ");

  switch ((activeNode.phase || "").trim().toLowerCase()) {
    case "preparing":
      return translate("shared.runtime.preparingNode", "Preparing node runtime");
    case "matrix":
      return targetLabel
        ? translate("shared.runtime.extractingTarget", `Extracting ${targetLabel}`, { target: targetLabel })
        : translate("shared.runtime.extractingMatrix", "Extracting matrix result");
    case "macro":
      if (targetLabel) {
        return translate("shared.runtime.runningMacroTarget", `${formatMacroLabel(activeNode.macro)}: ${targetLabel}`, {
          macro: formatMacroLabel(activeNode.macro),
          target: targetLabel,
        });
      }
      return translate("shared.runtime.runningMacro", `Running ${formatMacroLabel(activeNode.macro)}`, {
        macro: formatMacroLabel(activeNode.macro),
      });
    default:
      return targetLabel || translate("shared.phases.running", "Running");
  }
}

export function summarizeActiveTaskNodes(activeNodes: TaskActiveNode[] = []): string {
  if (activeNodes.length === 0) {
    return "";
  }

  const preview = activeNodes
    .slice(0, 2)
    .map((activeNode) => translate(
      "shared.formats.testingNodeDetail",
      `${activeNode.nodeName || `Node ${activeNode.nodeIndex + 1}`}: ${describeTaskActiveNode(activeNode)}`,
      {
        name: activeNode.nodeName || translate("shared.formats.nodeNumber", `Node ${activeNode.nodeIndex + 1}`, { index: activeNode.nodeIndex + 1 }),
        detail: describeTaskActiveNode(activeNode),
      },
    ))
    .join(" | ");
  const moreCount = activeNodes.length - 2;
  return moreCount > 0
    ? translate("shared.formats.testingNodesMore", `Testing ${activeNodes.length} node(s): ${preview} | +${moreCount} more`, {
      count: activeNodes.length,
      preview,
      moreCount,
    })
    : translate("shared.formats.testingNodes", `Testing ${activeNodes.length} node(s): ${preview}`, {
      count: activeNodes.length,
      preview,
    });
}

export function toggleTaskPresetSelection(current: TaskPresetSelection, target: TaskPreset): TaskPresetSelection {
  const childSelection = new Set(current.filter((preset) => preset !== "full"));

  if (target === "full") {
    return current.includes("full")
      ? []
      : [...taskPresets];
  }

  if (childSelection.has(target)) {
    childSelection.delete(target);
  } else {
    childSelection.add(target);
  }

  return normalizeTaskPresetSelection(Array.from(childSelection));
}

export function normalizeTaskPresetSelection(value: TaskPresetSelection): TaskPresetSelection {
  const selectedChildren = new Set<TaskPreset>(
    value.filter((preset): preset is Exclude<TaskPreset, "full"> => preset !== "full" && taskPresetChildren.includes(preset)),
  );

  if (value.includes("full")) {
    for (const preset of taskPresetChildren) {
      selectedChildren.add(preset);
    }
  }

  const normalized = taskPresets.filter((preset) => (
    preset === "full"
      ? taskPresetChildren.every((item) => selectedChildren.has(item))
      : selectedChildren.has(preset)
  ));

  return normalized;
}

export function formatTaskPresetSelectionLabel(value: TaskPresetSelection): string {
  const normalized = normalizeTaskPresetSelection(value);
  if (normalized.includes("full")) {
    return formatTaskPresetLabel("full");
  }

  const selectedChildren = normalized.filter((preset) => preset !== "full");
  if (selectedChildren.length === 0) {
    return translate("shared.presets.task", "Task");
  }
  if (selectedChildren.length <= 2) {
    return selectedChildren.map((preset) => formatTaskPresetLabel(preset)).join(" + ");
  }
  return translate("shared.formats.testCount", `${selectedChildren.length} tests`, { count: selectedChildren.length });
}

export function formatTaskPresetLabel(value: TaskPreset): string {
  switch (value) {
    case "full":
      return translate("shared.presets.full", "Full");
    case "ping":
      return translate("shared.presets.ping", "Ping");
    case "geo":
      return translate("shared.presets.geo", "Geo");
    case "speed":
      return translate("shared.presets.speed", "Speed");
    case "media":
      return translate("shared.presets.media", "Media");
    default:
      return value;
  }
}

function splitPresetGroups(preset?: string): string[] {
  const normalized = (preset || "").trim().toLowerCase();
  if (!normalized) {
    return [];
  }
  if (normalized === "full") {
    return ["ping", "geo", "speed", "media"];
  }
  return normalized.split(/[+,]/).map((item) => item.trim()).filter(Boolean);
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
  const groups = splitPresetGroups(preset);
  if (groups.length === 0) {
    return translate("shared.formats.archivedMetrics", "Archived metrics");
  }

  const labels: string[] = [];
  if (groups.includes("ping")) {
    labels.push(translate("shared.metrics.ping", "TLS RTT / HTTPS Ping"));
  }
  if (groups.includes("geo")) {
    labels.push(translate("shared.metrics.geo", "Inbound / Outbound Topology"));
  }
  if (groups.includes("speed")) {
    labels.push(translate("shared.metrics.speed", "Average / Max / Per-second Speed / Traffic"));
  }
  if (groups.includes("media")) {
    labels.push(translate("shared.metrics.media", "Netflix / Hulu / Bilibili HMT"));
  }
  return labels.join(" / ") || translate("shared.formats.archivedMetrics", "Archived metrics");
}

export function formatProtocolLibraryLabel(vendor?: string): string {
  return (vendor || "").trim() || translate("shared.states.unknown", "unknown");
}

export function formatProxyTypeLabel(value?: string): string {
  switch ((value || "").trim().toLowerCase()) {
    case "shadowsocks":
      return "Shadowsocks";
    case "ssr":
      return "SSR";
    case "socks5":
      return "SOCKS5";
    case "http":
      return "HTTP";
    case "vmess":
      return "VMess";
    case "trojan":
      return "Trojan";
    case "vless":
      return "VLESS";
    case "hysteria":
      return "Hysteria";
    case "hysteria2":
      return "Hysteria2";
    case "tuic":
      return "TUIC";
    case "anytls":
      return "AnyTLS";
    case "unknown":
      return translate("shared.states.unknown", "Unknown");
    default: {
      const source = (value || "").trim();
      if (!source) {
        return translate("shared.states.unknown", "Unknown");
      }
      return source;
    }
  }
}

export function formatBackendLabel(task?: Pick<ResultArchiveTask, "environment">): string {
  const identity = (task?.environment?.identity || "").trim();
  if (identity) {
    return identity;
  }

  const backend = task?.environment?.backend;
  if (!backend) {
    return translate("shared.states.empty", "empty");
  }

  const location = (backend.location || "").trim();
  if (location) {
    return location;
  }

  const ip = (backend.ip || "").trim();
  if (ip) {
    return ip;
  }

  const error = (backend.error || "").trim();
  if (error) {
    return translate("shared.formats.backendProbeFailed", "Probe failed");
  }

  const source = (backend.source || "").trim();
  if (source) {
    return source;
  }

  return translate("shared.states.empty", "empty");
}

export function formatBackendDetail(task?: Pick<ResultArchiveTask, "environment">): string {
  const identity = (task?.environment?.identity || "").trim();
  if (identity) {
    return identity;
  }

  const backend = task?.environment?.backend;
  if (!backend) {
    return translate("shared.states.empty", "empty");
  }

  const location = (backend.location || "").trim();
  const ip = (backend.ip || "").trim();
  const source = (backend.source || "").trim();
  const error = (backend.error || "").trim();

  const parts = [location, ip ? `IP=${ip}` : "", source ? `Source=${source}` : "", error ? `Error=${error}` : ""]
    .filter(Boolean);
  return parts.join(" | ") || translate("shared.states.empty", "empty");
}

export function summarizeDownloadTarget(downloadURL?: string, downloadTargets: DownloadTarget[] = []): string {
  const trimmedURL = (downloadURL || "").trim();
  if (!trimmedURL) {
    return translate("shared.targets.unavailable", "Target unavailable");
  }

  const matched = findDownloadTarget(trimmedURL, downloadTargets);
  if (matched) {
    return formatDownloadTargetPrimaryLabel(matched);
  }

  try {
    const url = new URL(trimmedURL);
    if (url.host === "speed.cloudflare.com") {
      return translate("shared.targets.cloudflareDefault", "Cloudflare default");
    }
    return url.host || translate("shared.formats.customTarget", "Custom target");
  } catch {
    return translate("shared.formats.customTarget", "Custom target");
  }
}

export function summarizeDownloadTargetDetail(downloadURL?: string, downloadTargets: DownloadTarget[] = []): string {
  const trimmedURL = (downloadURL || "").trim();
  if (!trimmedURL) {
    return translate("shared.states.none", "N/A");
  }

  const matched = findDownloadTarget(trimmedURL, downloadTargets);
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

export function summarizeDownloadTargetFooter(downloadURL?: string, downloadTargets: DownloadTarget[] = []): string {
  const trimmedURL = (downloadURL || "").trim();
  if (!trimmedURL) {
    return translate("shared.targets.unavailable", "Target unavailable");
  }

  const matched = findDownloadTarget(trimmedURL, downloadTargets);
  if (matched) {
    return formatDownloadTargetFooterLabel(matched);
  }

  try {
    const url = new URL(trimmedURL);
    if (url.host === "speed.cloudflare.com") {
      return translate("shared.targets.cloudflareDefault", "Cloudflare default");
    }
    return url.host || translate("shared.formats.customTarget", "Custom target");
  } catch {
    return translate("shared.formats.customTarget", "Custom target");
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
    return value ? translate("shared.states.booleanTrue", "true") : translate("shared.states.booleanFalse", "false");
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
      return translate("shared.mediaStatus.yes", "Yes");
    case "no":
      return translate("shared.mediaStatus.no", "No");
    case "originals_only":
      return translate("shared.mediaStatus.originalsOnly", "Originals Only");
    case "web_only":
      return translate("shared.mediaStatus.webOnly", "Web Only");
    case "oversea_only":
      return translate("shared.mediaStatus.overseaOnly", "Oversea Only");
    case "unsupported":
      return translate("shared.mediaStatus.unsupported", "Unsupported");
    case "failed":
      return translate("shared.mediaStatus.failed", "Failed");
    default:
      return String(value || "-");
  }
}

function formatUnlockMode(value: unknown): string {
  switch (String(value || "").trim().toLowerCase()) {
    case "native":
      return translate("shared.unlockMode.native", "Native");
    case "dns":
      return translate("shared.unlockMode.dns", "DNS");
    case "unknown":
      return translate("shared.unlockMode.unknown", "Unknown");
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

function findDownloadTarget(downloadURL: string, downloadTargets: DownloadTarget[]): DownloadTarget | undefined {
  return downloadTargets.find((target) => target.downloadURL === downloadURL);
}

function formatDownloadTargetPrimaryLabel(target: DownloadTarget): string {
  const primary = (target.name || "").trim();
  if (primary) {
    return primary;
  }

  const host = (target.host || "").trim();
  if (host) {
    return host;
  }

  const endpoint = (target.endpoint || "").trim();
  if (endpoint) {
    return endpoint;
  }

  return target.source === "cloudflare"
    ? translate("shared.targets.cloudflareDefault", "Cloudflare default")
    : translate("shared.targets.speedtestTarget", "Speedtest target");
}

function formatDownloadTargetFooterLabel(target: DownloadTarget): string {
  if (target.source === "cloudflare") {
    return translate("shared.targets.cloudflareDefault", "Cloudflare default");
  }

  const location = buildDownloadTargetLocation(target);
  const provider = (target.sponsor || "").trim();
  const parts = [location, provider].filter(Boolean);
  if (parts.length > 0) {
    return parts.join(" ");
  }

  return formatDownloadTargetPrimaryLabel(target);
}

function buildDownloadTargetLocation(target: DownloadTarget): string {
  const city = (target.city || "").trim();
  const country = (target.country || "").trim();
  const countryCode = (target.countryCode || "").trim().toUpperCase();

  if (city) {
    return city;
  }
  if (country) {
    return country;
  }
  if (countryCode) {
    return countryCode;
  }
  return "";
}

function isCountryCodeField(key?: string): boolean {
  return (key || "").toLowerCase().includes("countrycode");
}

export function containsRegionalFlagEmoji(value: unknown): boolean {
  const text = String(value || "");
  return /[\u{1F1E6}-\u{1F1FF}]{2}/u.test(text);
}

export function shouldUseEmojiFont(key?: string, value?: unknown): boolean {
  return isCountryCodeField(key) || containsRegionalFlagEmoji(value);
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
