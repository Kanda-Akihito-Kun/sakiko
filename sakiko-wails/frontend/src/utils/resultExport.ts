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

type ExportColumn = {
  key: string;
  label: string;
  width: number;
  align?: CanvasTextAlign;
};

type ExportSection = {
  kind: string;
  title: string;
  columns: ExportColumn[];
  rows: Record<string, unknown>[];
};

type TableMergePlan = {
  spans: Map<string, number>;
  skipped: Set<string>;
};

type FlagEmojiAssetMap = Map<string, HTMLImageElement | null>;

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
const HEADER_HEIGHT = 56;
const FOOTER_HEIGHT = 92;
const SECTION_GAP = 18;
const SECTION_TITLE_HEIGHT = 28;
const TABLE_HEADER_HEIGHT = 52;
const ROW_HEIGHT = 42;
const EXPORT_SCALE = 2;
const FONT_FAMILY = "'Microsoft YaHei', 'Segoe UI', 'Segoe UI Emoji', 'Apple Color Emoji', 'Noto Color Emoji', sans-serif";
const WATERMARK = "sakiko";
const TWEMOJI_BASE_URL = "https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/72x72";
const FLAG_ICON_GAP = 6;
const TOPOLOGY_INBOUND_COLUMNS = ["inboundCountryCode", "inboundASN", "inboundOrganization"] as const;
const TOPOLOGY_OUTBOUND_COLUMNS = ["outboundCountryCode", "outboundASN", "outboundOrganization", "outboundIP"] as const;

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
  const flagEmojiAssets = await loadFlagEmojiAssets(sections);
  const pageWidth = Math.max(1420, ...sections.map(sectionTotalWidth)) + PAGE_PADDING_X * 2;
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
    cursorY = drawSection(ctx, section, archive, cursorY, pageWidth, true, flagEmojiAssets, palette) + (index < sections.length - 1 ? SECTION_GAP : 0);
  });

  drawFooter(ctx, archive, downloadTargets, pageWidth, pageHeight, palette);

  const blob = await canvasToBlob(canvas);
  downloadBlob(blob, buildFileName(archive, mode));
}

function buildExportSections(archive: ResultArchive): ExportSection[] {
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

  return sections;
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
      { key: "rank", label: "Rank", width: 72, align: "center" },
      { key: "nodeName", label: "Node", width: 470 },
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
    case "speed_table":
      return {
        kind: section.kind,
        title: "Speed Test",
        columns: [
          { key: "rank", label: "Rank", width: 72, align: "center" },
          { key: "nodeName", label: "Node", width: 470 },
          { key: "proxyType", label: "Protocol", width: 170, align: "center" },
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
        })),
      };
    case "topology_table":
      return {
        kind: section.kind,
        title: "Topology Analysis",
        columns: [
          { key: "nodeName", label: "Node", width: 350 },
          { key: "inboundCountryCode", label: "Inbound Region", width: 110, align: "center" },
          { key: "inboundASN", label: "Inbound ASN", width: 126, align: "center" },
          { key: "inboundOrganization", label: "Inbound Org", width: 300 },
          { key: "outboundCountryCode", label: "Outbound Region", width: 110, align: "center" },
          { key: "outboundASN", label: "Outbound ASN", width: 126, align: "center" },
          { key: "outboundOrganization", label: "Outbound Org", width: 300 },
          { key: "outboundIP", label: "Outbound IP", width: 170, align: "center" },
          { key: "error", label: "Status", width: 150, align: "center" },
        ],
        rows: sortTopologyRows(section.rows || []),
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
            return { key: column.key, label: column.label, width: 220 };
          }
          return { key: column.key, label: column.label, width: 138, align: "center" as const };
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

function buildFallbackSection(archive: ResultArchive): ExportSection {
  return {
    kind: "raw_table",
    title: "Test Result",
    columns: [
      { key: "rank", label: "Rank", width: 72, align: "center" },
      { key: "nodeName", label: "Node", width: 460 },
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
  separated: boolean,
  flagEmojiAssets: FlagEmojiAssetMap,
  palette: ExportPalette,
): number {
  const tableWidth = sectionTotalWidth(section);
  const tableX = Math.max(PAGE_PADDING_X, Math.floor((pageWidth - tableWidth) / 2));
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
      flagEmojiAssets,
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
  flagEmojiAssets: FlagEmojiAssetMap,
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
      }, flagEmojiAssets, palette);
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
    `Tested At: ${formatUTC8DateTime(archive.state.finishedAt || archive.state.startedAt)}  Results are for reference only.`,
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

function buildFallbackMetrics(matrices: MatrixResult[]): string {
  const parts: string[] = [];
  const rtt = extractMatrixValue(matrices, "TEST_PING_RTT");
  const httpPing = extractMatrixValue(matrices, "TEST_PING_CONN");
  const avg = extractMatrixValue(matrices, "SPEED_AVERAGE");
  const max = extractMatrixValue(matrices, "SPEED_MAX");
  const traffic = extractMatrixValue(matrices, "SPEED_TRAFFIC_USED");
  if (numericValue(rtt) > 0) {
    parts.push(`TLS ${Math.round(numericValue(rtt))}ms`);
  }
  if (numericValue(httpPing) > 0) {
    parts.push(`HTTPS ${Math.round(numericValue(httpPing))}ms`);
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
  if (sections.some((section) => section.kind === "media_unlock_table")) {
    return "Media Unlock";
  }
  if ((archive.task.context?.preset || "").toLowerCase() === "ping") {
    return "Latency Test";
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
  flagEmojiAssets: FlagEmojiAssetMap,
  palette: ExportPalette,
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

  const remainder = content.slice(flag.length).trimStart() || flagCountryCode(flag);
  const iconSize = Math.max(14, Math.min(18, height - 14));
  const iconX = x + 10;
  const iconY = y + (height - iconSize) / 2;
  const textX = iconX + iconSize + FLAG_ICON_GAP;
  const maxTextWidth = Math.max(0, width - 20 - iconSize - FLAG_ICON_GAP);

  ctx.save();
  ctx.fillStyle = options.color;
  ctx.font = options.font;
  ctx.textAlign = "left";
  ctx.textBaseline = "middle";

  const image = flagEmojiAssets.get(flag) || null;
  if (image) {
    ctx.drawImage(image, iconX, iconY, iconSize, iconSize);
  } else {
    drawFlagBadge(ctx, flagCountryCode(flag), iconX, iconY, iconSize, palette);
  }

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

async function loadFlagEmojiAssets(sections: ExportSection[]): Promise<FlagEmojiAssetMap> {
  const flags = new Set<string>();
  for (const section of sections) {
    for (const row of section.rows) {
      const flag = extractLeadingFlagEmoji(String(row.nodeName || ""));
      if (flag) {
        flags.add(flag);
      }
    }
  }

  const entries = await Promise.all(Array.from(flags).map(async (flag) => [flag, await loadFlagEmojiAsset(flag)] as const));
  return new Map(entries);
}

function loadFlagEmojiAsset(flag: string): Promise<HTMLImageElement | null> {
  const codepoints = Array.from(flag)
    .map((char) => char.codePointAt(0)?.toString(16))
    .filter(Boolean)
    .join("-");
  if (!codepoints) {
    return Promise.resolve(null);
  }

  return new Promise((resolve) => {
    const image = new Image();
    image.crossOrigin = "anonymous";
    image.onload = () => resolve(image);
    image.onerror = () => resolve(null);
    image.src = `${TWEMOJI_BASE_URL}/${codepoints}.png`;
  });
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

function flagCountryCode(flag: string): string {
  return Array.from(flag)
    .map((char) => {
      const codepoint = char.codePointAt(0) || 0;
      return codepoint >= 0x1F1E6 && codepoint <= 0x1F1FF
        ? String.fromCharCode(codepoint - 127397)
        : "";
    })
    .join("");
}

function drawFlagBadge(
  ctx: CanvasRenderingContext2D,
  code: string,
  x: number,
  y: number,
  size: number,
  palette: ExportPalette,
) {
  ctx.save();
  ctx.fillStyle = palette.flagBadgeFill;
  ctx.strokeStyle = palette.flagBadgeBorder;
  ctx.lineWidth = 1;
  roundRect(ctx, x, y, size, size, 4);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = palette.text;
  ctx.font = `600 ${Math.max(7, Math.floor(size * 0.34))}px ${FONT_FAMILY}`;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(code || "--", x + size / 2, y + size / 2);
  ctx.restore();
}

function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
  const nextRadius = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + nextRadius, y);
  ctx.arcTo(x + width, y, x + width, y + height, nextRadius);
  ctx.arcTo(x + width, y + height, x, y + height, nextRadius);
  ctx.arcTo(x, y + height, x, y, nextRadius);
  ctx.arcTo(x, y, x + width, y, nextRadius);
  ctx.closePath();
}

function formatBytesAsSpeed(value: unknown): string {
  const numeric = numericValue(value);
  if (numeric <= 0) {
    return "0B";
  }
  if (numeric >= 1_000_000) {
    return `${(numeric / 1_000_000).toFixed(2)}MB`;
  }
  if (numeric >= 1_000) {
    return `${(numeric / 1_000).toFixed(2)}KB`;
  }
  return `${Math.round(numeric)}B`;
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
  const timestamp = buildUTC8FileTimestamp(archive.state.finishedAt || archive.state.startedAt || new Date().toISOString());
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

function formatUTC8DateTime(value?: string): string {
  if (!value) {
    return "N/A";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  const utc8 = new Date(date.getTime() + 8 * 60 * 60 * 1000);
  const year = utc8.getUTCFullYear();
  const month = pad2(utc8.getUTCMonth() + 1);
  const day = pad2(utc8.getUTCDate());
  const hours = pad2(utc8.getUTCHours());
  const minutes = pad2(utc8.getUTCMinutes());
  const seconds = pad2(utc8.getUTCSeconds());
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds} UTC+8`;
}

function buildUTC8FileTimestamp(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return sanitizeFileName(value);
  }

  const utc8 = new Date(date.getTime() + 8 * 60 * 60 * 1000);
  const year = utc8.getUTCFullYear();
  const month = pad2(utc8.getUTCMonth() + 1);
  const day = pad2(utc8.getUTCDate());
  const hours = pad2(utc8.getUTCHours());
  const minutes = pad2(utc8.getUTCMinutes());
  const seconds = pad2(utc8.getUTCSeconds());
  return `${year}-${month}-${day}_${hours}-${minutes}-${seconds}_UTC+8`;
}

function pad2(value: number): string {
  return String(value).padStart(2, "0");
}


