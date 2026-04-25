import type {
  DownloadTarget,
  MatrixResult,
  ResultArchive,
  ResultReportSection,
} from "../types/sakiko";
import {
  formatBackendLabel,
  formatProtocolLibraryLabel,
  formatProxyTypeLabel,
  summarizeDownloadTargetFooter,
} from "./dashboard";
import type { ResolvedThemeMode } from "../theme/appTheme";
import { filterMediaReportSectionColumns } from "./mediaUnlock";
import { mediaCellTone } from "./mediaMatrix";
import { formatDateTimeForDisplay, formatDateTimeForFileName } from "./dateTime";

export type ExportColumn = {
  key: string;
  label: string;
  width: number;
  align?: CanvasTextAlign;
};

export type ExportSection = {
  kind: string;
  title: string;
  columns: ExportColumn[];
  rows: Record<string, unknown>[];
};

type TableMergePlan = {
  spans: Map<string, number>;
  skipped: Set<string>;
};

type ExportPalette = {
  background: string;
  watermark: string;
  border: string;
  divider: string;
  headerFill: string;
  text: string;
  mutedText: string;
  latencyBase: string;
  latencyAccent: string;
  speedBase: string;
  speedAccent: string;
  errorText: string;
  errorFill: string;
  mediaSuccessFill: string;
  mediaWarningFill: string;
  mediaErrorFill: string;
  rowEvenFill: string;
  rowOddFill: string;
  flagBadgeFill: string;
  flagBadgeBorder: string;
};

const PAGE_PADDING_X = 18;
const PAGE_PADDING_Y = 16;
const MIN_CONTENT_WIDTH = 1420;
const HEADER_HEIGHT = 56;
const FOOTER_HEIGHT = 92;
const SECTION_GAP = 18;
const SECTION_TITLE_HEIGHT = 28;
const TABLE_HEADER_HEIGHT = 52;
const ROW_HEIGHT = 42;
const EXPORT_SCALE = 2;
const FONT_FAMILY = "'Microsoft YaHei', 'twemoji mozilla', 'Segoe UI', 'Segoe UI Emoji', 'Apple Color Emoji', 'Noto Color Emoji', sans-serif";
const EMOJI_FONT_FAMILY = "'twemoji mozilla', 'Segoe UI Emoji', 'Apple Color Emoji', 'Noto Color Emoji', sans-serif";
const WATERMARK = "sakiko";
const FLAG_ICON_GAP = 6;
const TOPOLOGY_INBOUND_COLUMNS = ["inboundASN", "inboundIP", "inboundInfo"] as const;
const TOPOLOGY_OUTBOUND_COLUMNS = ["outboundASN", "outboundIP", "outboundInfo"] as const;

const exportPalettes: Record<ResolvedThemeMode, ExportPalette> = {
  light: {
    background: "#ffffff",
    watermark: "rgba(44, 195, 129, 0.04)",
    border: "#d6e8dd",
    divider: "#e3efe7",
    headerFill: "#f3fbf6",
    text: "#151515",
    mutedText: "#56635d",
    latencyBase: "#dcf6e7",
    latencyAccent: "#1aa36e",
    speedBase: "#dff8ea",
    speedAccent: "#2cc381",
    errorText: "#9a3412",
    errorFill: "#fff1eb",
    mediaSuccessFill: "#e5f7eb",
    mediaWarningFill: "#fff6dc",
    mediaErrorFill: "#ffe9e7",
    rowEvenFill: "#ffffff",
    rowOddFill: "#fcfcfd",
    flagBadgeFill: "#edf6f1",
    flagBadgeBorder: "#c9ddd0",
  },
  dark: {
    background: "#17191f",
    watermark: "rgba(10, 132, 255, 0.08)",
    border: "#39424d",
    divider: "#2b3440",
    headerFill: "#222833",
    text: "#f5f7fb",
    mutedText: "#a8b3c2",
    latencyBase: "#1d3b32",
    latencyAccent: "#30d158",
    speedBase: "#1b3b33",
    speedAccent: "#34c77b",
    errorText: "#ffb39f",
    errorFill: "#4a2d2a",
    mediaSuccessFill: "#203b30",
    mediaWarningFill: "#43371f",
    mediaErrorFill: "#492e2b",
    rowEvenFill: "#1b2028",
    rowOddFill: "#212732",
    flagBadgeFill: "#24303a",
    flagBadgeBorder: "#415262",
  },
};

export async function exportResultArchiveImage(
  archive: ResultArchive,
  downloadTargets: DownloadTarget[] = [],
  mode: ResolvedThemeMode = "light",
): Promise<void> {
  const palette = exportPalettes[mode];
  const sections = buildExportSections(archive);
  await ensureExportFontsReady();
  const contentWidth = resolveContentWidth(sections);
  const pageWidth = contentWidth + PAGE_PADDING_X * 2;
  const pageHeight = calculatePageHeight(sections);

  const canvas = document.createElement("canvas");
  canvas.width = Math.ceil(pageWidth * EXPORT_SCALE);
  canvas.height = Math.ceil(pageHeight * EXPORT_SCALE);

  const ctx = canvas.getContext("2d");
  if (!ctx) {
    throw new Error("Canvas context unavailable.");
  }

  ctx.scale(EXPORT_SCALE, EXPORT_SCALE);

  drawBackground(ctx, pageWidth, pageHeight, palette);
  drawHeader(ctx, archive, pageWidth, palette);

  let cursorY = PAGE_PADDING_Y + HEADER_HEIGHT;
  sections.forEach((section, index) => {
    cursorY = drawSection(ctx, section, archive, cursorY, pageWidth, contentWidth, true, palette) + (index < sections.length - 1 ? SECTION_GAP : 0);
  });

  drawFooter(ctx, archive, downloadTargets, pageWidth, pageHeight, palette);

  const blob = await canvasToBlob(canvas);
  downloadBlob(blob, buildFileName(archive, mode));
}

export function buildExportSections(archive: ResultArchive): ExportSection[] {
  const sections: ExportSection[] = [];
  const hasSpeedSection = (archive.report.sections || []).some((section) => section.kind === "speed_table");
  const latencySection = hasSpeedSection ? null : buildLatencySection(archive);
  if (latencySection) {
    sections.push(latencySection);
  }

  for (const section of archive.report.sections || []) {
    const normalized = normalizeReportSection(section);
    if (normalized.columns.length === 0) {
      continue;
    }
    sections.push(normalized);
  }

  if (sections.length === 0) {
    sections.push(buildFallbackSection(archive));
  }

  return normalizeSectionLayouts(sections);
}

function buildLatencySection(archive: ResultArchive): ExportSection | null {
  const rows = (archive.results || [])
    .map((result) => ({
      nodeName: result.proxyInfo.name || "Unnamed node",
      proxyType: formatProxyType(result.proxyInfo.type || "unknown"),
      rttMillis: extractMatrixValue(result.matrices, "TEST_PING_RTT"),
      httpPingMillis: extractMatrixValue(result.matrices, "TEST_PING_CONN"),
      error: result.error || "",
    }))
    .filter((row) => numericValue(row.rttMillis) > 0 || numericValue(row.httpPingMillis) > 0 || row.error);

  if (rows.length === 0) {
    return null;
  }

  const sortedRows = rows.map((row, index) => ({ rank: index + 1, ...row }));

  return {
    kind: "latency_table",
    title: "Latency Test",
    columns: [
      { key: "nodeName", label: "Node", width: 230 },
      { key: "proxyType", label: "Protocol", width: 170, align: "center" },
      { key: "rttMillis", label: "TLS RTT", width: 160, align: "center" },
      { key: "httpPingMillis", label: "HTTPS Ping", width: 190, align: "center" },
      { key: "error", label: "Status", width: 220, align: "center" },
    ],
    rows: sortedRows,
  };
}

function normalizeReportSection(section: ResultReportSection): ExportSection {
  switch (section.kind) {
    case "speed_table": {
      const hasUDPNATType = sectionHasColumn(section, "udpNatType");
      return {
        kind: section.kind,
        title: "Speed Test",
        columns: [
          { key: "nodeName", label: "Node", width: 230 },
          { key: "proxyType", label: "Protocol", width: 170, align: "center" },
          ...(hasUDPNATType ? [{ key: "udpNatType", label: "UDP NAT Type", width: 180, align: "center" as const }] : []),
          { key: "rttMillis", label: "TLS RTT", width: 160, align: "center" },
          { key: "httpPingMillis", label: "HTTPS Ping", width: 190, align: "center" },
          { key: "averageBytesPerSecond", label: "Average Speed", width: 170, align: "center" },
          { key: "maxBytesPerSecond", label: "Max Speed", width: 170, align: "center" },
          { key: "perSecondBytesPerSecond", label: "Per-second Speed", width: 160, align: "center" },
          { key: "trafficUsedBytes", label: "Traffic Used", width: 150, align: "center" },
          { key: "error", label: "Status", width: 160, align: "center" },
        ],
        rows: (section.rows || []).map((row, index) => ({
          rank: row.rank ?? index + 1,
          ...row,
          proxyType: formatProxyType(row.proxyType),
          error: normalizeSpeedSectionStatus(row.error, row.trafficUsedBytes, Object.prototype.hasOwnProperty.call(row, "trafficUsedBytes")),
        })),
      };
    }
    case "topology_table":
      return {
        kind: section.kind,
        title: "Topology Analysis",
        columns: [
          { key: "nodeName", label: "Node", width: 230 },
          { key: "proxyType", label: "Protocol", width: 150, align: "center" },
          { key: "inboundASN", label: "Inbound ASN", width: 126, align: "center" },
          { key: "inboundIP", label: "Inbound IP", width: 150, align: "center" },
          { key: "inboundInfo", label: "Inbound Info", width: 280 },
          { key: "outboundASN", label: "Outbound ASN", width: 126, align: "center" },
          { key: "outboundIP", label: "Outbound IP", width: 150, align: "center" },
          { key: "outboundInfo", label: "Outbound Info", width: 280 },
          { key: "error", label: "Status", width: 150, align: "center" },
        ],
        rows: sortTopologyRows(normalizeTopologyRows(section.rows || [])),
      };
    case "udp_nat_table":
      return {
        kind: section.kind,
        title: "UDP NAT Test",
        columns: [
          { key: "nodeName", label: "Node", width: 230 },
          { key: "proxyType", label: "Protocol", width: 160, align: "center" },
          { key: "natType", label: "UDP NAT Type", width: 180, align: "center" },
          { key: "internalEndpoint", label: "Internal Endpoint", width: 220, align: "center" },
          { key: "publicEndpoint", label: "Public Endpoint", width: 220, align: "center" },
          { key: "error", label: "Status", width: 200, align: "center" },
        ],
        rows: section.rows || [],
      };
    case "media_unlock_table":
      const columns = filterMediaReportSectionColumns(section.columns || []).filter((column) => column.key !== "proxyType");
      if (!columns.some((column) => column.key !== "nodeName")) {
        return {
          kind: section.kind,
          title: "Media Unlock Test",
          columns: [],
          rows: [],
        };
      }
      return {
        kind: section.kind,
        title: "Media Unlock Test",
        columns: columns.map((column) => {
          if (column.key === "nodeName") {
            return { key: column.key, label: column.label || column.key, width: 230 };
          }
          return { key: column.key, label: column.label || column.key, width: 138, align: "center" as const };
        }),
        rows: section.rows || [],
      };
    default:
      return {
        kind: section.kind,
        title: section.title || section.kind,
        columns: (section.columns || []).map((column) => ({
          key: column.key,
          label: column.label,
          width: 180,
        })),
        rows: section.rows || [],
      };
  }
}

function sectionHasColumn(section: ResultReportSection, key: string): boolean {
  return Boolean(section.columns?.some((column) => column.key === key)
    || section.rows?.some((row) => Object.prototype.hasOwnProperty.call(row, key)));
}

function buildFallbackSection(archive: ResultArchive): ExportSection {
  return {
    kind: "raw_table",
    title: "Test Result",
    columns: [
      { key: "nodeName", label: "Node", width: 230 },
      { key: "proxyType", label: "Protocol", width: 180, align: "center" },
      { key: "status", label: "Status", width: 240, align: "center" },
      { key: "metrics", label: "Metrics", width: 500 },
    ],
    rows: (archive.results || []).map((result, index) => ({
      rank: index + 1,
      nodeName: result.proxyInfo.name || "Unnamed node",
      proxyType: formatProxyType(result.proxyInfo.type || "unknown"),
      status: result.error || "ok",
      metrics: buildFallbackMetrics(result.matrices),
    })),
  };
}

function drawBackground(ctx: CanvasRenderingContext2D, width: number, height: number, palette: ExportPalette) {
  ctx.fillStyle = palette.background;
  ctx.fillRect(0, 0, width, height);

  ctx.fillStyle = palette.watermark;
  ctx.font = `600 58px ${FONT_FAMILY}`;
  ctx.textAlign = "center";
  for (let y = 110; y < height - 80; y += 220) {
    ctx.fillText(WATERMARK, width / 2, y);
  }
}

function drawHeader(ctx: CanvasRenderingContext2D, archive: ResultArchive, width: number, palette: ExportPalette) {
  const mainTitle = buildMainTitle(archive);
  const profileName = archive.task.context?.profileName || "Unknown Profile";

  ctx.fillStyle = palette.text;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.font = `700 24px ${FONT_FAMILY}`;
  ctx.fillText(`Sakiko - ${mainTitle} | ${profileName}`, width / 2, PAGE_PADDING_Y + 20);

  ctx.fillStyle = palette.mutedText;
  ctx.font = `500 12px ${FONT_FAMILY}`;
  ctx.fillText(`Task: ${archive.task.id}`, width / 2, PAGE_PADDING_Y + 42);

  ctx.strokeStyle = palette.divider;
  ctx.beginPath();
  ctx.moveTo(PAGE_PADDING_X, PAGE_PADDING_Y + HEADER_HEIGHT - 2);
  ctx.lineTo(width - PAGE_PADDING_X, PAGE_PADDING_Y + HEADER_HEIGHT - 2);
  ctx.stroke();
}

function drawSection(
  ctx: CanvasRenderingContext2D,
  section: ExportSection,
  archive: ResultArchive,
  y: number,
  pageWidth: number,
  contentWidth: number,
  separated: boolean,
  palette: ExportPalette,
): number {
  const tableX = Math.max(PAGE_PADDING_X, Math.floor((pageWidth - contentWidth) / 2));
  const mergePlan = buildTableMergePlan(section);
  let cursorY = y;

  if (separated) {
    ctx.fillStyle = palette.text;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.font = `700 18px ${FONT_FAMILY}`;
    ctx.fillText(section.title, pageWidth / 2, cursorY + 12);
    cursorY += SECTION_TITLE_HEIGHT;
  }

  drawTableHeader(ctx, section.columns, tableX, cursorY, palette);
  section.rows.forEach((row, rowIndex) => {
    drawTableRow(
      ctx,
      section,
      archive,
      row,
      rowIndex,
      tableX,
      cursorY + TABLE_HEADER_HEIGHT + rowIndex * ROW_HEIGHT,
      mergePlan,
      palette,
    );
  });

  return cursorY + TABLE_HEADER_HEIGHT + section.rows.length * ROW_HEIGHT;
}

function drawTableHeader(
  ctx: CanvasRenderingContext2D,
  columns: ExportColumn[],
  x: number,
  y: number,
  palette: ExportPalette,
) {
  let cursorX = x;
  columns.forEach((column) => {
    ctx.fillStyle = palette.headerFill;
    ctx.fillRect(cursorX, y, column.width, TABLE_HEADER_HEIGHT);
    strokeCell(ctx, cursorX, y, column.width, TABLE_HEADER_HEIGHT, palette);
    drawCellText(ctx, column.label, cursorX, y, column.width, TABLE_HEADER_HEIGHT, {
      align: column.align || "left",
      color: palette.text,
      font: `600 12px ${FONT_FAMILY}`,
      wrap: true,
    });
    cursorX += column.width;
  });
}

function drawTableRow(
  ctx: CanvasRenderingContext2D,
  section: ExportSection,
  archive: ResultArchive,
  row: Record<string, unknown>,
  rowIndex: number,
  x: number,
  y: number,
  mergePlan: TableMergePlan,
  palette: ExportPalette,
) {
  let cursorX = x;
  section.columns.forEach((column) => {
    const mergeKey = createMergeCellKey(column.key, rowIndex);
    if (mergePlan.skipped.has(mergeKey)) {
      cursorX += column.width;
      return;
    }

    const value = row[column.key];
    const cellHeight = ROW_HEIGHT * (mergePlan.spans.get(mergeKey) || 1);
    const fill = resolveCellFill(section, column.key, value, rowIndex, palette);
    ctx.fillStyle = fill;
    ctx.fillRect(cursorX, y, column.width, cellHeight);
    strokeCell(ctx, cursorX, y, column.width, cellHeight, palette);

    if (column.key === "perSecondBytesPerSecond" && Array.isArray(value)) {
      drawSparkBars(ctx, cursorX + 10, y + 8, column.width - 20, cellHeight - 16, value, palette);
    } else if (column.key === "nodeName") {
      drawNodeNameCell(ctx, formatNodeNameForExport(String(value || ""), archive), cursorX, y, column.width, cellHeight, {
        color: palette.text,
        font: `500 13px ${FONT_FAMILY}`,
      });
    } else {
      drawCellText(ctx, renderCellValue(section.kind, column.key, value, archive), cursorX, y, column.width, cellHeight, {
        align: column.align || "left",
        color: value && column.key === "error" ? palette.errorText : palette.text,
        font: `500 13px ${FONT_FAMILY}`,
        wrap: shouldWrap(column.key),
      });
    }
    cursorX += column.width;
  });
}

function drawCellText(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  width: number,
  height: number,
  options: {
    align: CanvasTextAlign;
    color: string;
    font: string;
    wrap: boolean;
  },
) {
  const content = text || "-";
  const maxWidth = width - 20;

  ctx.save();
  ctx.fillStyle = options.color;
  ctx.font = options.font;
  ctx.textAlign = options.align;
  ctx.textBaseline = "middle";

  const lines = options.wrap ? wrapText(ctx, content, maxWidth, 2) : [truncateText(ctx, content, maxWidth)];
  const lineHeight = 15;
  const originY = y + height / 2 - ((lines.length - 1) * lineHeight) / 2;
  const originX = options.align === "center"
    ? x + width / 2
    : options.align === "right"
      ? x + width - 10
      : x + 10;

  lines.forEach((line, index) => {
    ctx.fillText(line, originX, originY + index * lineHeight);
  });
  ctx.restore();
}

function drawSparkBars(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  values: unknown[],
  palette: ExportPalette,
) {
  const numeric = values.map(numericValue).filter((value) => value > 0);
  if (numeric.length === 0) {
    drawCellText(ctx, "-", x, y, width, height, {
      align: "center",
      color: palette.mutedText,
      font: `500 13px ${FONT_FAMILY}`,
      wrap: false,
    });
    return;
  }

  const max = Math.max(...numeric, 1);
  const step = width / Math.max(numeric.length, 1);
  numeric.forEach((value, index) => {
    const left = x + index * step;
    const right = x + (index + 1) * step;
    const barWidth = Math.max(1, right - left);
    const barHeight = Math.max(3, Math.round((value / max) * height));
    const barY = y + height - barHeight;
    ctx.fillStyle = mixColor(palette.speedBase, palette.speedAccent, 0.35 + (value / max) * 0.65);
    ctx.fillRect(left, barY, barWidth, barHeight);
  });
}

function drawFooter(
  ctx: CanvasRenderingContext2D,
  archive: ResultArchive,
  downloadTargets: DownloadTarget[],
  width: number,
  height: number,
  palette: ExportPalette,
) {
  const footerTop = height - FOOTER_HEIGHT;
  const preset = buildMainTitle(archive);
  const runtimeSeconds = calculateRuntimeSeconds(archive);
  const config = archive.task.config;
  const protocolLibrary = formatProtocolLibraryLabel(archive.task.vendor);
  const backend = formatBackendLabel(archive.task);
  const target = summarizeDownloadTargetFooter(config.downloadURL, downloadTargets);

  ctx.strokeStyle = palette.divider;
  ctx.beginPath();
  ctx.moveTo(PAGE_PADDING_X, footerTop);
  ctx.lineTo(width - PAGE_PADDING_X, footerTop);
  ctx.stroke();

  ctx.textAlign = "left";
  ctx.textBaseline = "middle";
  ctx.fillStyle = palette.text;
  ctx.font = `500 12px ${FONT_FAMILY}`;
  ctx.fillText(
    `Protocol Library=${protocolLibrary}  Backend=${backend}  SpeedTest Target=${target}  Profile=${archive.task.context?.profileName || "Unknown"}  Preset=${preset}`,
    PAGE_PADDING_X,
    footerTop + 18,
  );
  ctx.fillText(
    `Nodes=${archive.task.nodes.length}  Threads=${config.downloadThreading || 0}  Duration=${runtimeSeconds}s  Ping Samples=${config.pingAverageOver || 0}`,
    PAGE_PADDING_X,
    footerTop + 40,
  );

  ctx.fillStyle = palette.mutedText;
  ctx.fillText(
    `Tested At: ${formatDateTimeForDisplay(archive.state.finishedAt || archive.state.startedAt)}  Results are for reference only.`,
    PAGE_PADDING_X,
    footerTop + 66,
  );
}

function resolveCellFill(
  section: ExportSection,
  key: string,
  value: unknown,
  rowIndex: number,
  palette: ExportPalette,
): string {
  if (key === "rttMillis" || key === "httpPingMillis") {
    return heatColor(numericValue(value), collectNumericValues(section.rows, key), palette.latencyBase, palette.latencyAccent, palette.rowEvenFill);
  }
  if (key === "averageBytesPerSecond" || key === "maxBytesPerSecond") {
    return heatColor(numericValue(value), collectNumericValues(section.rows, key), palette.speedBase, palette.speedAccent, palette.rowEvenFill);
  }
  if (key === "perSecondBytesPerSecond") {
    return rowIndex % 2 === 0 ? palette.rowEvenFill : palette.rowOddFill;
  }
  if (key === "error" && value) {
    return palette.errorFill;
  }
  if (section.kind === "media_unlock_table" && key !== "nodeName" && key !== "proxyType") {
    switch (mediaCellTone(String(value || ""))) {
      case "success":
        return palette.mediaSuccessFill;
      case "warning":
        return palette.mediaWarningFill;
      case "error":
        return palette.mediaErrorFill;
      default:
        return rowIndex % 2 === 0 ? palette.rowEvenFill : palette.rowOddFill;
    }
  }
  return rowIndex % 2 === 0 ? palette.rowEvenFill : palette.rowOddFill;
}

function renderCellValue(kind: string, key: string, value: unknown, archive: ResultArchive): string {
  if (value === null || value === undefined || value === "") {
    return key === "error" ? "OK" : "-";
  }

  if (key === "proxyType") {
    return formatProxyType(value);
  }
  if (key === "rttMillis" || key === "httpPingMillis") {
    const numeric = numericValue(value);
    return numeric > 0 ? `${Math.round(numeric)}ms` : "-";
  }
  if (key === "averageBytesPerSecond" || key === "maxBytesPerSecond") {
    return formatBytesAsSpeed(value);
  }
  if (key === "trafficUsedBytes") {
    return formatBytesAsMegabytes(value);
  }
  if (key === "error") {
    return String(value || "OK");
  }
  if (key === "status") {
    return formatMediaStatus(value);
  }
  if (key === "unlockMode") {
    return formatUnlockMode(value);
  }
  if (kind === "media_unlock_table" && key !== "nodeName" && key !== "proxyType") {
    const text = String(value || "-");
    return text.toLowerCase().includes("failed") ? "Test Failed" : text;
  }
  if (key === "outboundIP" || key === "inboundIP") {
    return maskIPAddress(value);
  }
  if (key.toLowerCase().includes("asn")) {
    const numeric = numericValue(value);
    return numeric > 0 ? `AS${numeric}` : "-";
  }
  if (Array.isArray(value)) {
    if (key === "perSecondBytesPerSecond") {
      return "";
    }
    return value.map(String).join(", ");
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  if (key === "metrics") {
    return truncateTextWithLimit(String(value), 110);
  }
  if (kind === "raw_table" && key === "status" && String(value) === "ok") {
    return "OK";
  }
  if (key === "nodeName") {
    return truncateTextWithLimit(String(value), archive.task.context?.preset === "full" ? 36 : 40);
  }
  return String(value);
}

function normalizeSpeedSectionStatus(error: unknown, trafficUsedBytes: unknown, trafficMeasured: boolean): string {
  const status = String(error || "").trim();
  if (status) {
    return status;
  }
  if (trafficMeasured && numericValue(trafficUsedBytes) <= 0) {
    return "Failed";
  }
  return "";
}

function buildFallbackMetrics(matrices: MatrixResult[]): string {
  const parts: string[] = [];
  const rtt = extractMatrixValue(matrices, "TEST_PING_RTT");
  const httpPing = extractMatrixValue(matrices, "TEST_PING_CONN");
  const udpNat = extractMatrixValue(matrices, "UDP_NAT_TYPE");
  const avg = extractMatrixValue(matrices, "SPEED_AVERAGE");
  const max = extractMatrixValue(matrices, "SPEED_MAX");
  const traffic = extractMatrixValue(matrices, "SPEED_TRAFFIC_USED");
  if (numericValue(rtt) > 0) {
    parts.push(`TLS ${Math.round(numericValue(rtt))}ms`);
  }
  if (numericValue(httpPing) > 0) {
    parts.push(`HTTPS ${Math.round(numericValue(httpPing))}ms`);
  }
  if (udpNat && typeof udpNat === "object" && "type" in (udpNat as Record<string, unknown>)) {
    parts.push(`UDP ${String((udpNat as Record<string, unknown>).type || "-")}`);
  }
  if (numericValue(avg) > 0) {
    parts.push(`Average ${formatBytesAsSpeed(avg)}`);
  }
  if (numericValue(max) > 0) {
    parts.push(`Peak ${formatBytesAsSpeed(max)}`);
  }
  if (numericValue(traffic) > 0) {
    parts.push(`Traffic ${formatBytesAsMegabytes(traffic)}`);
  }
  return parts.join(" | ") || "-";
}

function buildMainTitle(archive: ResultArchive): string {
  const preset = (archive.task.context?.preset || "").toLowerCase();
  switch (preset) {
    case "full":
      return "Full Test";
    case "ping":
      return "Latency Test";
    case "geo":
      return "Topology Analysis";
    case "udp":
      return "UDP NAT Test";
    case "speed":
      return "Speed Test";
    case "media":
      return "Media Unlock";
    default:
      break;
  }

  const sections = archive.report.sections || [];
  if (sections.some((section) => section.kind === "speed_table") && sections.some((section) => section.kind === "topology_table")) {
    return "Full Test";
  }
  if (sections.some((section) => section.kind === "speed_table")) {
    return "Speed Test";
  }
  if (sections.some((section) => section.kind === "topology_table")) {
    return "Topology Analysis";
  }
  if (sections.some((section) => section.kind === "udp_nat_table")) {
    return "UDP NAT Test";
  }
  if (sections.some((section) => section.kind === "media_unlock_table")) {
    return "Media Unlock";
  }
  return "Test Result";
}

function calculatePageHeight(sections: ExportSection[]): number {
  let height = PAGE_PADDING_Y * 2 + HEADER_HEIGHT + FOOTER_HEIGHT;
  sections.forEach((section) => {
    height += SECTION_TITLE_HEIGHT;
    height += TABLE_HEADER_HEIGHT + section.rows.length * ROW_HEIGHT;
  });
  height += Math.max(0, sections.length - 1) * SECTION_GAP;
  return height;
}

function sectionTotalWidth(section: ExportSection): number {
  return section.columns.reduce((sum, column) => sum + column.width, 0);
}

function resolveContentWidth(sections: ExportSection[]): number {
  return Math.max(MIN_CONTENT_WIDTH, ...sections.map(sectionTotalWidth));
}

function normalizeSectionLayouts(sections: ExportSection[]): ExportSection[] {
  const targetWidth = resolveContentWidth(sections);
  return sections.map((section) => normalizeSectionWidth(section, targetWidth));
}

function normalizeSectionWidth(section: ExportSection, targetWidth: number): ExportSection {
  const currentWidth = sectionTotalWidth(section);
  if (currentWidth >= targetWidth || section.columns.length === 0) {
    return {
      ...section,
      columns: section.columns.map((column) => ({ ...column })),
    };
  }

  const columns = section.columns.map((column) => ({ ...column }));
  const stretchWeights = columns.map((column) => columnStretchWeight(section.kind, column.key));
  const totalWeight = stretchWeights.reduce((sum, weight) => sum + weight, 0);
  if (totalWeight <= 0) {
    columns[columns.length - 1].width += targetWidth - currentWidth;
    return { ...section, columns };
  }

  let remaining = targetWidth - currentWidth;
  columns.forEach((column, index) => {
    const weight = stretchWeights[index];
    if (weight <= 0 || remaining <= 0) {
      return;
    }
    const grow = Math.floor((targetWidth - currentWidth) * (weight / totalWeight));
    if (grow <= 0) {
      return;
    }
    column.width += grow;
    remaining -= grow;
  });

  if (remaining > 0) {
    const preferredOrder = preferredStretchColumnIndexes(columns, stretchWeights);
    for (let offset = 0; offset < remaining; offset += 1) {
      const targetIndex = preferredOrder[offset % preferredOrder.length] ?? columns.length - 1;
      columns[targetIndex].width += 1;
    }
  }

  return {
    ...section,
    columns,
  };
}

function preferredStretchColumnIndexes(columns: ExportColumn[], weights: number[]): number[] {
  const indexes = columns
    .map((_, index) => index)
    .filter((index) => weights[index] > 0)
    .sort((left, right) => weights[right] - weights[left]);
  return indexes.length > 0 ? indexes : [columns.length - 1];
}

function columnStretchWeight(sectionKind: string, key: string): number {
  const lowerKey = key.toLowerCase();
  if (lowerKey === "nodename") {
    return sectionKind === "media_unlock_table" ? 5 : 6;
  }
  if (lowerKey.includes("organization")) {
    return 5;
  }
  if (lowerKey.includes("info")) {
    return 5;
  }
  if (lowerKey.includes("country")) {
    return 2;
  }
  if (lowerKey.includes("city")) {
    return 2;
  }
  if (lowerKey.includes("endpoint")) {
    return 4;
  }
  if (lowerKey === "metrics") {
    return 6;
  }
  if (lowerKey === "error" || lowerKey === "status") {
    return 3;
  }
  if (sectionKind === "media_unlock_table") {
    return 1;
  }
  return 0;
}

function strokeCell(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  palette: ExportPalette,
) {
  ctx.strokeStyle = palette.border;
  ctx.lineWidth = 1;
  ctx.strokeRect(x, y, width, height);
}

function collectNumericValues(rows: Record<string, unknown>[], key: string): number[] {
  return rows
    .map((row) => numericValue(row[key]))
    .filter((value) => value > 0);
}

function buildTableMergePlan(section: ExportSection): TableMergePlan {
  const plan: TableMergePlan = {
    spans: new Map<string, number>(),
    skipped: new Set<string>(),
  };
  if (section.kind !== "topology_table") {
    return plan;
  }

  applyMergeGroup(plan, section.rows, "inboundIP", TOPOLOGY_INBOUND_COLUMNS);
  applyMergeGroup(plan, section.rows, "outboundIP", TOPOLOGY_OUTBOUND_COLUMNS);
  return plan;
}

function applyMergeGroup(
  plan: TableMergePlan,
  rows: Record<string, unknown>[],
  groupKey: string,
  columns: readonly string[],
) {
  for (let start = 0; start < rows.length; start += 1) {
    const mergeValue = normalizeMergeValue(rows[start]?.[groupKey]);
    if (!mergeValue) {
      continue;
    }

    let end = start + 1;
    while (end < rows.length && normalizeMergeValue(rows[end]?.[groupKey]) === mergeValue) {
      end += 1;
    }

    const span = end - start;
    if (span > 1) {
      columns.forEach((columnKey) => {
        plan.spans.set(createMergeCellKey(columnKey, start), span);
        for (let rowIndex = start + 1; rowIndex < end; rowIndex += 1) {
          plan.skipped.add(createMergeCellKey(columnKey, rowIndex));
        }
      });
    }

    start = end - 1;
  }
}

function drawNodeNameCell(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  width: number,
  height: number,
  options: {
    color: string;
    font: string;
  },
) {
  const content = text || "-";
  const flag = extractLeadingFlagEmoji(content);
  if (!flag) {
    drawCellText(ctx, content, x, y, width, height, {
      align: "left",
      color: options.color,
      font: options.font,
      wrap: false,
    });
    return;
  }

  const remainder = content.slice(flag.length).trimStart() || content;
  const iconSize = Math.max(14, Math.min(18, height - 14));
  const iconX = x + 10;
  const textX = iconX + iconSize + FLAG_ICON_GAP;
  const maxTextWidth = Math.max(0, width - 20 - iconSize - FLAG_ICON_GAP);

  ctx.save();
  ctx.fillStyle = options.color;
  ctx.font = options.font;
  ctx.textAlign = "left";
  ctx.textBaseline = "middle";

  ctx.font = `400 ${iconSize}px ${EMOJI_FONT_FAMILY}`;
  ctx.fillText(flag, iconX, y + height / 2);
  ctx.font = options.font;
  ctx.fillText(truncateText(ctx, remainder, maxTextWidth), textX, y + height / 2);
  ctx.restore();
}

function createMergeCellKey(columnKey: string, rowIndex: number): string {
  return `${columnKey}:${rowIndex}`;
}

function normalizeMergeValue(value: unknown): string {
  const raw = String(value ?? "").trim();
  if (!raw || raw === "-") {
    return "";
  }
  return raw;
}

function normalizeTopologyRows(rows: Record<string, unknown>[]): Record<string, unknown>[] {
  return rows.map((row) => ({
    ...row,
    proxyType: formatProxyType(row.proxyType),
    inboundInfo: normalizeGeoInfo(row.inboundInfo, row.inboundOrganization, row.inboundCity),
    outboundInfo: normalizeGeoInfo(row.outboundInfo, row.outboundOrganization, row.outboundCity),
  }));
}

function normalizeGeoInfo(info: unknown, organization: unknown, city: unknown): string {
  const joined = joinGeoInfo(city, organization);
  if (joined) {
    return joined;
  }

  const raw = String(info ?? "").trim();
  if (!raw) {
    return "";
  }
  if (raw.includes(" / ")) {
    const [left, ...rest] = raw.split(" / ");
    const right = rest.join(" / ").trim();
    if (left.trim() && right) {
      return `${right} | ${left.trim()}`;
    }
  }
  return raw.replace(/\s+\/\s+/g, " | ");
}

function joinGeoInfo(city: unknown, organization: unknown): string {
  const parts = [city, organization]
    .map((value) => String(value ?? "").trim())
    .filter(Boolean);
  return parts.join(" | ");
}

function sortTopologyRows(rows: Record<string, unknown>[]): Record<string, unknown>[] {
  return [...rows].sort((left, right) => {
    const inboundCompare = compareMergeGroup(left.inboundIP, right.inboundIP);
    if (inboundCompare !== 0) {
      return inboundCompare;
    }

    const outboundCompare = compareMergeGroup(left.outboundIP, right.outboundIP);
    if (outboundCompare !== 0) {
      return outboundCompare;
    }

    const leftError = normalizeMergeValue(left.error);
    const rightError = normalizeMergeValue(right.error);
    if (!!leftError !== !!rightError) {
      return leftError ? 1 : -1;
    }

    return String(left.nodeName || "").localeCompare(String(right.nodeName || ""), "zh-Hans-CN");
  });
}

function compareMergeGroup(left: unknown, right: unknown): number {
  const leftValue = normalizeMergeValue(left);
  const rightValue = normalizeMergeValue(right);
  if (!leftValue && !rightValue) {
    return 0;
  }
  if (!leftValue) {
    return 1;
  }
  if (!rightValue) {
    return -1;
  }
  return leftValue.localeCompare(rightValue, "en");
}

function heatColor(value: number, allValues: number[], from: string, to: string, fallback: string): string {
  if (value <= 0 || allValues.length === 0) {
    return fallback;
  }
  const max = Math.max(...allValues, 1);
  const ratio = Math.min(1, value / max);
  return mixColor(from, to, 0.2 + ratio * 0.8);
}

function mixColor(from: string, to: string, ratio: number): string {
  const start = hexToRgb(from);
  const end = hexToRgb(to);
  const r = Math.round(start.r + (end.r - start.r) * ratio);
  const g = Math.round(start.g + (end.g - start.g) * ratio);
  const b = Math.round(start.b + (end.b - start.b) * ratio);
  return `rgb(${r}, ${g}, ${b})`;
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
  const raw = hex.replace("#", "");
  return {
    r: Number.parseInt(raw.slice(0, 2), 16),
    g: Number.parseInt(raw.slice(2, 4), 16),
    b: Number.parseInt(raw.slice(4, 6), 16),
  };
}

function wrapText(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  maxLines: number,
): string[] {
  const words = text.split(/\s+/);
  if (words.length <= 1) {
    return [truncateText(ctx, text, maxWidth)];
  }

  const lines: string[] = [];
  let current = "";
  words.forEach((word) => {
    const next = current ? `${current} ${word}` : word;
    if (ctx.measureText(next).width <= maxWidth) {
      current = next;
      return;
    }

    if (current) {
      lines.push(current);
      current = word;
    } else {
      lines.push(truncateText(ctx, word, maxWidth));
    }
  });

  if (current) {
    lines.push(current);
  }

  if (lines.length > maxLines) {
    const kept = lines.slice(0, maxLines);
    kept[maxLines - 1] = truncateText(ctx, kept[maxLines - 1], maxWidth);
    return kept;
  }
  return lines;
}

function truncateText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string {
  if (ctx.measureText(text).width <= maxWidth) {
    return text;
  }

  const glyphs = Array.from(text);
  let candidate = glyphs.join("");
  while (glyphs.length > 0 && ctx.measureText(`${candidate}...`).width > maxWidth) {
    glyphs.pop();
    candidate = glyphs.join("");
  }
  return `${candidate}...`;
}

function truncateTextWithLimit(text: string, limit: number): string {
  const glyphs = Array.from(text);
  return glyphs.length <= limit ? text : `${glyphs.slice(0, limit).join("")}...`;
}

function formatNodeNameForExport(text: string, archive: ResultArchive): string {
  return truncateTextWithLimit(text, archive.task.context?.preset === "full" ? 36 : 40);
}

async function ensureExportFontsReady(): Promise<void> {
  if (typeof document === "undefined" || !("fonts" in document)) {
    return;
  }

  try {
    await Promise.allSettled([
      document.fonts.load(`400 13px ${FONT_FAMILY}`),
      document.fonts.load(`400 18px ${EMOJI_FONT_FAMILY}`),
      document.fonts.ready,
    ]);
  } catch {
    // Fall back to whatever fonts are currently available.
  }
}

function extractLeadingFlagEmoji(text: string): string {
  const glyphs = Array.from(text.trimStart());
  if (glyphs.length < 2) {
    return "";
  }
  return isRegionalIndicator(glyphs[0]) && isRegionalIndicator(glyphs[1])
    ? glyphs.slice(0, 2).join("")
    : "";
}

function isRegionalIndicator(value: string): boolean {
  const codepoint = value.codePointAt(0) || 0;
  return codepoint >= 0x1F1E6 && codepoint <= 0x1F1FF;
}

function formatBytesAsSpeed(value: unknown): string {
  const numeric = numericValue(value);
  if (numeric <= 0) {
    return "0 Mbps";
  }
  return `${((numeric * 8) / 1_000_000).toFixed(2)} Mbps`;
}

function formatBytesAsMegabytes(value: unknown): string {
  const numeric = numericValue(value);
  if (numeric <= 0) {
    return "0.00 MB";
  }
  return `${(numeric / 1_000_000).toFixed(2)} MB`;
}

function formatProxyType(value: unknown): string {
  return formatProxyTypeLabel(String(value || ""));
}

function extractMatrixValue(matrices: MatrixResult[], type: string): unknown {
  const matrix = matrices.find((item) => item.type === type);
  if (!matrix) {
    return null;
  }

  if (matrix.payload && typeof matrix.payload === "object") {
    const payload = matrix.payload as Record<string, unknown>;
    if ("value" in payload) {
      return payload.value;
    }
    if ("values" in payload) {
      return payload.values;
    }
  }
  return matrix.payload;
}

function numericValue(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const numeric = Number.parseFloat(value);
    return Number.isFinite(numeric) ? numeric : 0;
  }
  return 0;
}

function shouldWrap(key: string): boolean {
  return key.toLowerCase().includes("organization")
    || key.toLowerCase().includes("info")
    || key.toLowerCase().includes("endpoint")
    || key.toLowerCase().includes("metrics")
    || key.toLowerCase().includes("error");
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

function maskIPAddress(value: unknown): string {
  if (!value) {
    return "-";
  }
  const raw = String(value).trim();
  if (raw.includes(".")) {
    const parts = raw.split(".");
    if (parts.length === 4) {
      return `${parts[0]}.${parts[1]}.***.***`;
    }
  }
  if (raw.includes(":")) {
    const parts = raw.split(":").filter(Boolean);
    if (parts.length >= 2) {
      return `${parts[0]}:${parts[1]}:****:****`;
    }
  }
  return raw;
}

function calculateRuntimeSeconds(archive: ResultArchive): number {
  const start = Date.parse(archive.state.startedAt || "");
  const end = Date.parse(archive.state.finishedAt || "");
  if (Number.isNaN(start) || Number.isNaN(end) || end <= start) {
    return 0;
  }
  return Math.max(0, Math.round((end - start) / 1000));
}

function buildFileName(archive: ResultArchive, mode: ResolvedThemeMode): string {
  const profile = sanitizeFileName(archive.task.context?.profileName || "result");
  const preset = sanitizeFileName(archive.task.context?.preset || "report");
  const timestamp = formatDateTimeForFileName(archive.state.finishedAt || archive.state.startedAt || new Date().toISOString());
  const theme = mode === "dark" ? "Dark" : "Light";
  return `${profile}_${preset}_${timestamp}_${theme}.png`;
}

function sanitizeFileName(input: string): string {
  return input.trim().replace(/[<>:"/\\|?*\u0000-\u001F]/g, "_") || "result";
}

function canvasToBlob(canvas: HTMLCanvasElement): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error("Failed to generate image blob."));
        return;
      }
      resolve(blob);
    }, "image/png");
  });
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}



