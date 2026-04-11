import type {
  DownloadTarget,
  MatrixResult,
  ResultArchive,
  ResultReportSection,
} from "../types/sakiko";
import {
  formatBackendLabel,
  formatProtocolLibraryLabel,
  summarizeDownloadTarget,
} from "./dashboard";
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
const BORDER_COLOR = "#d6e8dd";
const HEADER_FILL = "#f3fbf6";
const TEXT_COLOR = "#151515";
const MUTED_TEXT = "#56635d";
const LATENCY_BASE = "#dcf6e7";
const LATENCY_ACCENT = "#1aa36e";
const SPEED_BASE = "#dff8ea";
const SPEED_ACCENT = "#2cc381";
const WATERMARK = "sakiko";
const TOPOLOGY_INBOUND_COLUMNS = ["inboundCountryCode", "inboundASN", "inboundOrganization"] as const;
const TOPOLOGY_OUTBOUND_COLUMNS = ["outboundCountryCode", "outboundASN", "outboundOrganization", "outboundIP"] as const;

export async function exportResultArchiveImage(archive: ResultArchive, downloadTargets: DownloadTarget[] = []): Promise<void> {
  const sections = buildExportSections(archive);
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

  drawBackground(ctx, pageWidth, pageHeight);
  drawHeader(ctx, archive, pageWidth);

  let cursorY = PAGE_PADDING_Y + HEADER_HEIGHT;
  sections.forEach((section, index) => {
    cursorY = drawSection(ctx, section, archive, cursorY, pageWidth, true) + (index < sections.length - 1 ? SECTION_GAP : 0);
  });

  drawFooter(ctx, archive, downloadTargets, pageWidth, pageHeight);

  const blob = await canvasToBlob(canvas);
  downloadBlob(blob, buildFileName(archive));
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

  const sortedRows = rows
    .sort((left, right) => {
      const leftValue = numericValue(left.httpPingMillis) || numericValue(left.rttMillis);
      const rightValue = numericValue(right.httpPingMillis) || numericValue(right.rttMillis);
      return leftValue - rightValue;
    })
    .map((row, index) => ({ rank: index + 1, ...row }));

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
        title: "SpeedTest",
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
      const columns = filterMediaReportSectionColumns(section.columns || []);
      if (!columns.some((column) => column.key !== "nodeName" && column.key !== "proxyType")) {
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
          if (column.key === "proxyType") {
            return { key: column.key, label: column.label, width: 110, align: "center" as const };
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

function drawBackground(ctx: CanvasRenderingContext2D, width: number, height: number) {
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, width, height);

  ctx.fillStyle = "rgba(44, 195, 129, 0.04)";
  ctx.font = `600 58px ${FONT_FAMILY}`;
  ctx.textAlign = "center";
  for (let y = 110; y < height - 80; y += 220) {
    ctx.fillText(WATERMARK, width / 2, y);
  }
}

function drawHeader(ctx: CanvasRenderingContext2D, archive: ResultArchive, width: number) {
  const mainTitle = buildMainTitle(archive);
  const profileName = archive.task.context?.profileName || "Unknown Profile";

  ctx.fillStyle = TEXT_COLOR;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.font = `700 24px ${FONT_FAMILY}`;
  ctx.fillText(`sakiko - ${mainTitle} | ${profileName}`, width / 2, PAGE_PADDING_Y + 20);

  ctx.fillStyle = "#6a7770";
  ctx.font = `500 12px ${FONT_FAMILY}`;
  ctx.fillText(`Task: ${archive.task.id}`, width / 2, PAGE_PADDING_Y + 42);

  ctx.strokeStyle = "#e3efe7";
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
): number {
  const tableWidth = sectionTotalWidth(section);
  const tableX = Math.max(PAGE_PADDING_X, Math.floor((pageWidth - tableWidth) / 2));
  const mergePlan = buildTableMergePlan(section);
  let cursorY = y;

  if (separated) {
    ctx.fillStyle = TEXT_COLOR;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.font = `700 18px ${FONT_FAMILY}`;
    ctx.fillText(section.title, pageWidth / 2, cursorY + 12);
    cursorY += SECTION_TITLE_HEIGHT;
  }

  drawTableHeader(ctx, section.columns, tableX, cursorY);
  section.rows.forEach((row, rowIndex) => {
    drawTableRow(ctx, section, archive, row, rowIndex, tableX, cursorY + TABLE_HEADER_HEIGHT + rowIndex * ROW_HEIGHT, mergePlan);
  });

  return cursorY + TABLE_HEADER_HEIGHT + section.rows.length * ROW_HEIGHT;
}

function drawTableHeader(
  ctx: CanvasRenderingContext2D,
  columns: ExportColumn[],
  x: number,
  y: number,
) {
  let cursorX = x;
  columns.forEach((column) => {
    ctx.fillStyle = HEADER_FILL;
    ctx.fillRect(cursorX, y, column.width, TABLE_HEADER_HEIGHT);
    strokeCell(ctx, cursorX, y, column.width, TABLE_HEADER_HEIGHT);
    drawCellText(ctx, column.label, cursorX, y, column.width, TABLE_HEADER_HEIGHT, {
      align: column.align || "left",
      color: TEXT_COLOR,
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
    const fill = resolveCellFill(section, column.key, value, rowIndex);
    ctx.fillStyle = fill;
    ctx.fillRect(cursorX, y, column.width, cellHeight);
    strokeCell(ctx, cursorX, y, column.width, cellHeight);

    if (column.key === "perSecondBytesPerSecond" && Array.isArray(value)) {
      drawSparkBars(ctx, cursorX + 10, y + 8, column.width - 20, cellHeight - 16, value);
    } else {
      drawCellText(ctx, renderCellValue(section.kind, column.key, value, archive), cursorX, y, column.width, cellHeight, {
        align: column.align || "left",
        color: value && column.key === "error" ? "#9a3412" : TEXT_COLOR,
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
) {
  const numeric = values.map(numericValue).filter((value) => value > 0);
  if (numeric.length === 0) {
    drawCellText(ctx, "-", x, y, width, height, {
      align: "center",
      color: MUTED_TEXT,
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
    ctx.fillStyle = mixColor("#e4f8ec", SPEED_ACCENT, 0.35 + (value / max) * 0.65);
    ctx.fillRect(left, barY, barWidth, barHeight);
  });
}

function drawFooter(
  ctx: CanvasRenderingContext2D,
  archive: ResultArchive,
  downloadTargets: DownloadTarget[],
  width: number,
  height: number,
) {
  const footerTop = height - FOOTER_HEIGHT;
  const preset = buildMainTitle(archive);
  const runtimeSeconds = calculateRuntimeSeconds(archive);
  const config = archive.task.config;
  const protocolLibrary = formatProtocolLibraryLabel(archive.task.vendor);
  const backend = formatBackendLabel(archive.task);
  const target = summarizeDownloadTarget(config.downloadURL, downloadTargets);

  ctx.strokeStyle = "#e3efe7";
  ctx.beginPath();
  ctx.moveTo(PAGE_PADDING_X, footerTop);
  ctx.lineTo(width - PAGE_PADDING_X, footerTop);
  ctx.stroke();

  ctx.textAlign = "left";
  ctx.textBaseline = "middle";
  ctx.fillStyle = TEXT_COLOR;
  ctx.font = `500 12px ${FONT_FAMILY}`;
  ctx.fillText(
    `\u534f\u8bae\u5e93=${protocolLibrary}  \u540e\u7aef=${backend}  \u6d4b\u901f\u76ee\u6807=${target}  \u8ba2\u9605=${archive.task.context?.profileName || "Unknown"}  \u6307\u6807=${preset}`,
    PAGE_PADDING_X,
    footerTop + 18,
  );
  ctx.fillText(
    `\u8282\u70b9=${archive.task.nodes.length}  \u7ebf\u7a0b=${config.downloadThreading || 0}  \u8017\u65f6=${runtimeSeconds}s  Ping=${config.pingAverageOver || 0}\u6b21`,
    PAGE_PADDING_X,
    footerTop + 40,
  );

  ctx.fillStyle = MUTED_TEXT;
  ctx.fillText(
    `\u6d4b\u8bd5\u65f6\u95f4: ${formatUTC8DateTime(archive.state.finishedAt || archive.state.startedAt)}  \u6d4b\u8bd5\u7ed3\u679c\u4ec5\u4f9b\u53c2\u8003\uff0c\u8bf7\u4ee5\u5b9e\u9645\u60c5\u51b5\u4e3a\u51c6`,
    PAGE_PADDING_X,
    footerTop + 66,
  );
}

function resolveCellFill(
  section: ExportSection,
  key: string,
  value: unknown,
  rowIndex: number,
): string {
  if (key === "rttMillis" || key === "httpPingMillis") {
    return heatColor(numericValue(value), collectNumericValues(section.rows, key), LATENCY_BASE, LATENCY_ACCENT);
  }
  if (key === "averageBytesPerSecond" || key === "maxBytesPerSecond") {
    return heatColor(numericValue(value), collectNumericValues(section.rows, key), SPEED_BASE, SPEED_ACCENT);
  }
  if (key === "perSecondBytesPerSecond") {
    return "#ffffff";
  }
  if (key === "error" && value) {
    return "#fff1eb";
  }
  if (section.kind === "media_unlock_table" && key !== "nodeName" && key !== "proxyType") {
    switch (mediaCellTone(String(value || ""))) {
      case "success":
        return "#e5f7eb";
      case "warning":
        return "#fff6dc";
      case "error":
        return "#ffe9e7";
      default:
        return rowIndex % 2 === 0 ? "#ffffff" : "#fcfcfd";
    }
  }
  return rowIndex % 2 === 0 ? "#ffffff" : "#fcfcfd";
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
  sections.forEach((section, index) => {
    if (index > 0) {
      height += SECTION_TITLE_HEIGHT;
    }
    height += TABLE_HEADER_HEIGHT + section.rows.length * ROW_HEIGHT;
  });
  height += Math.max(0, sections.length - 1) * SECTION_GAP;
  return height;
}

function sectionTotalWidth(section: ExportSection): number {
  return section.columns.reduce((sum, column) => sum + column.width, 0);
}

function strokeCell(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number) {
  ctx.strokeStyle = BORDER_COLOR;
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

function heatColor(value: number, allValues: number[], from: string, to: string): string {
  if (value <= 0 || allValues.length === 0) {
    return "#ffffff";
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
  if (!value) {
    return "Unknown";
  }
  const source = String(value);
  return source.charAt(0).toUpperCase() + source.slice(1);
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

function buildFileName(archive: ResultArchive): string {
  const profile = sanitizeFileName(archive.task.context?.profileName || "result");
  const preset = sanitizeFileName(archive.task.context?.preset || "report");
  const timestamp = buildUTC8FileTimestamp(archive.state.finishedAt || archive.state.startedAt || new Date().toISOString());
  return `${profile}_${preset}_${timestamp}.png`;
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


