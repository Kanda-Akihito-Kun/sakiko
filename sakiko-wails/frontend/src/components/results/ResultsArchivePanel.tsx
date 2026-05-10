import DeleteOutlineRounded from "@mui/icons-material/DeleteOutlineRounded";
import ExpandMoreRounded from "@mui/icons-material/ExpandMoreRounded";
import InsightsRounded from "@mui/icons-material/InsightsRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SaveAltRounded from "@mui/icons-material/SaveAltRounded";
import {
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Collapse,
  Divider,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { alpha, type Theme } from "@mui/material/styles";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { SakikoService } from "../../services/sakikoService";
import { useThemeMode } from "../../theme/themeMode";
import type { DownloadTarget, ResultArchive, ResultArchiveListItem } from "../../types/sakiko";
import {
  formatBackendDetail,
  formatBackendLabel,
  formatDateTime,
  formatProtocolLibraryLabel,
  formatReportValue,
  shouldUseEmojiFont,
  summarizeDownloadTarget,
  summarizeResultMetrics,
} from "../../utils/dashboard";
import {
  buildExportSections,
  exportResultArchiveImage,
  type ExportColumn,
  type ExportSection,
} from "../../utils/resultExport";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ResultsArchivePanelProps = {
  archives: ResultArchiveListItem[];
  archiveDetails: Record<string, ResultArchive | undefined>;
  archiveLoading: Record<string, boolean | undefined>;
  loading: boolean;
  downloadTargets: DownloadTarget[];
  visibleCount: number;
  onDeleteArchive: (taskId: string) => Promise<void>;
  onEnsureArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  onLoadMore: () => void;
  onRefresh: () => Promise<void>;
  title?: string;
  emptyTitle?: string;
  emptyDescription?: string;
  loadingArchiveLabel?: string;
};

export function ResultsArchivePanel({
  archives,
  archiveDetails,
  archiveLoading,
  loading,
  downloadTargets,
  visibleCount,
  onDeleteArchive,
  onEnsureArchive,
  onLoadMore,
  onRefresh,
  title,
  emptyTitle,
  emptyDescription,
  loadingArchiveLabel,
}: ResultsArchivePanelProps) {
  const { resolvedExportPictureMode } = useThemeMode();
  const { t } = useTranslation();
  const [expandedTaskId, setExpandedTaskId] = useState("");
  const [exportingTaskId, setExportingTaskId] = useState("");
  const [deletingTaskId, setDeletingTaskId] = useState("");
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const visibleArchives = archives.slice(0, visibleCount);
  const canLoadMore = visibleCount < archives.length;

  useEffect(() => {
    if (!canLoadMore || !sentinelRef.current) {
      return;
    }

    const root = document.querySelector(".sakiko-content__body");
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          onLoadMore();
        }
      },
      {
        root,
        rootMargin: "160px 0px",
        threshold: 0,
      },
    );

    observer.observe(sentinelRef.current);
    return () => {
      observer.disconnect();
    };
  }, [canLoadMore, onLoadMore, visibleCount]);

  const toggleArchive = (taskId: string, detail: ResultArchive | undefined, detailLoading: boolean) => {
    setExpandedTaskId((current) => current === taskId ? "" : taskId);
    if (expandedTaskId !== taskId && !detail && !detailLoading) {
      void onEnsureArchive(taskId);
    }
  };

  return (
    <SectionCard
      title={title || t("dashboard.results.title")}
      icon={<InsightsRounded color="primary" />}
      action={(
        <Chip
          icon={<RefreshRounded />}
          label={loading ? t("dashboard.results.syncing") : `${visibleArchives.length}/${archives.length}`}
          variant="outlined"
          onClick={() => void onRefresh()}
          clickable
        />
      )}
    >
      {loading && archives.length === 0 ? (
        <Stack spacing={2} alignItems="center" py={4}>
          <CircularProgress size={28} />
          <Typography variant="body2" color="text.secondary">
            {t("dashboard.results.loading")}
          </Typography>
        </Stack>
      ) : archives.length === 0 ? (
        <EmptyState
          title={emptyTitle || t("dashboard.results.noArchivesTitle")}
          description={emptyDescription || t("dashboard.results.noArchivesDescription")}
        />
      ) : (
        <Stack spacing={1.5}>
          {visibleArchives.map((archive) => {
            const detail = archiveDetails[archive.taskId];
            const detailLoading = Boolean(archiveLoading[archive.taskId]);
            const expanded = expandedTaskId === archive.taskId;

            return (
              <Card
                key={archive.taskId}
                variant="outlined"
                sx={{
                  borderRadius: 2.5,
                  overflow: "hidden",
                  bgcolor: "background.paper",
                }}
              >
                <Box
                  role="button"
                  tabIndex={0}
                  onClick={() => toggleArchive(archive.taskId, detail, detailLoading)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      toggleArchive(archive.taskId, detail, detailLoading);
                    }
                  }}
                  sx={(theme) => ({
                    px: 2,
                    py: 1.5,
                    cursor: "pointer",
                    outline: "none",
                    transition: theme.transitions.create(["background-color"]),
                    "&:hover": {
                      bgcolor: alpha(theme.palette.primary.main, 0.035),
                    },
                    "&:focus-visible": {
                      boxShadow: `inset 0 0 0 2px ${theme.palette.primary.main}`,
                    },
                  })}
                >
                  <Stack spacing={1.25}>
                    <Stack
                      direction="row"
                      justifyContent="space-between"
                      spacing={1}
                      sx={{ minWidth: 0, alignItems: "flex-start", flexWrap: { xs: "wrap", lg: "nowrap" } }}
                    >
                      <Stack direction="row" spacing={1.25} sx={{ minWidth: 0, flex: "1 1 320px" }} alignItems="center">
                        <ExpandMoreRounded
                          sx={(theme) => ({
                            color: "text.secondary",
                            transform: expanded ? "rotate(180deg)" : "rotate(0deg)",
                            transition: theme.transitions.create("transform", { duration: theme.transitions.duration.shorter }),
                            flexShrink: 0,
                          })}
                        />
                        <Box sx={{ minWidth: 0 }}>
                          <Typography variant="subtitle1" noWrap title={archive.taskName || archive.taskId}>
                            {archive.taskName || archive.taskId}
                          </Typography>
                          <Typography variant="body2" color="text.secondary" noWrap title={archive.taskId}>
                            {archive.taskId}
                          </Typography>
                        </Box>
                      </Stack>

                      <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap" sx={{ flexShrink: 0 }}>
                        <Button
                          size="small"
                          color="error"
                          variant="outlined"
                          startIcon={deletingTaskId === archive.taskId ? <CircularProgress size={16} /> : <DeleteOutlineRounded fontSize="small" />}
                          disabled={deletingTaskId === archive.taskId || exportingTaskId === archive.taskId}
                          onClick={async (event) => {
                            event.preventDefault();
                            event.stopPropagation();
                            setDeletingTaskId(archive.taskId);
                            try {
                              await onDeleteArchive(archive.taskId);
                              setExpandedTaskId((current) => current === archive.taskId ? "" : current);
                            } finally {
                              setDeletingTaskId("");
                            }
                          }}
                        >
                          {t("shared.actions.delete")}
                        </Button>
                        <Button
                          size="small"
                          variant="outlined"
                          startIcon={exportingTaskId === archive.taskId ? <CircularProgress size={16} /> : <SaveAltRounded fontSize="small" />}
                          disabled={exportingTaskId === archive.taskId || deletingTaskId === archive.taskId}
                          onClick={async (event) => {
                            event.preventDefault();
                            event.stopPropagation();
                            setExportingTaskId(archive.taskId);
                            try {
                              const nextDetail = detail || await onEnsureArchive(archive.taskId);
                              if (!nextDetail) {
                                throw new Error(t("dashboard.results.archiveUnavailable"));
                              }
                              const settings = await SakikoService.GetAppSettings();
                              await exportResultArchiveImage(nextDetail, downloadTargets, resolvedExportPictureMode, {
                                hideProfileNameInExport: settings.hideProfileNameInExport,
                                hideCNInboundInExport: settings.hideCNInboundInExport,
                              });
                            } finally {
                              setExportingTaskId("");
                            }
                          }}
                        >
                          {t("shared.actions.export")}
                        </Button>
                        <Chip
                          label={archive.exitCode || t("shared.states.pending")}
                          color={archive.exitCode === "success" ? "success" : "default"}
                          size="small"
                          variant="outlined"
                        />
                      </Stack>
                    </Stack>

                    <Box
                      sx={{
                        display: "grid",
                        gridTemplateColumns: { xs: "1fr", md: "repeat(6, minmax(0, 1fr))" },
                        gap: 1,
                      }}
                    >
                      <SummaryMetric label={t("dashboard.results.summary.testTime")} value={formatDateTime(archive.finishedAt || archive.startedAt)} />
                      <SummaryMetric label={t("dashboard.results.summary.profile")} value={archive.profileName || t("shared.states.unknownProfile")} />
                      {archive.knightName ? <SummaryMetric label={t("dashboard.results.summary.knight")} value={archive.knightName} /> : null}
                      <SummaryMetric label={t("dashboard.results.summary.metrics")} value={summarizeResultMetrics(archive.preset)} />
                      <SummaryMetric label={t("dashboard.results.summary.nodes")} value={`${archive.nodeCount}`} mono />
                      <SummaryMetric
                        label={t("dashboard.results.summary.protocolLibrary")}
                        value={detail ? formatProtocolLibraryLabel(detail.task.vendor) : t("shared.formats.loadDetailToInspect")}
                      />
                      <SummaryMetric
                        label={t("dashboard.results.summary.backend")}
                        value={detail ? formatBackendLabel(detail.task) : t("shared.formats.loadDetailToInspect")}
                        title={detail ? formatBackendDetail(detail.task) : undefined}
                      />
                    </Box>
                  </Stack>
                </Box>

                <Collapse in={expanded} timeout="auto" unmountOnExit>
                  <Divider />
                  <Box sx={{ px: { xs: 1, md: 2 }, py: 2 }}>
                    {detailLoading && !detail ? (
                      <Stack direction="row" spacing={1.25} alignItems="center" py={1}>
                        <CircularProgress size={20} />
                        <Typography variant="body2" color="text.secondary">
                          {loadingArchiveLabel || t("dashboard.results.loadingArchive")}
                        </Typography>
                      </Stack>
                    ) : detail ? (
                      <ArchiveReportPreview detail={detail} downloadTargets={downloadTargets} />
                    ) : (
                      <EmptyState
                        title={t("dashboard.results.archiveUnavailableTitle")}
                        description={t("dashboard.results.archiveUnavailableDescription")}
                      />
                    )}
                  </Box>
                </Collapse>
              </Card>
            );
          })}

          {canLoadMore ? (
            <Stack ref={sentinelRef} alignItems="center" py={1.5} spacing={1}>
              <CircularProgress size={20} />
              <Typography variant="caption" color="text.secondary">
                {t("dashboard.results.scrollToLoadMore")}
              </Typography>
            </Stack>
          ) : null}
        </Stack>
      )}
    </SectionCard>
  );
}

type ArchiveReportPreviewProps = {
  detail: ResultArchive;
  downloadTargets: DownloadTarget[];
};

function ArchiveReportPreview({ detail, downloadTargets }: ArchiveReportPreviewProps) {
  const sections = buildExportSections(detail);
  const title = summarizeResultMetrics(detail.task.context?.preset || "");
  const profileName = detail.task.context?.profileName || "Unknown Profile";
  const knightName = resolveArchiveKnightName(detail);
  const heading = knightName ? `Sakiko - ${title} | ${profileName} | ${knightName}` : `Sakiko - ${title} | ${profileName}`;

  return (
    <Card
      variant="outlined"
      sx={(theme) => ({
        position: "relative",
        overflow: "hidden",
        borderRadius: 2.5,
        bgcolor: theme.palette.mode === "dark" ? alpha(theme.palette.common.white, 0.025) : "#fff",
        "&::before": {
          content: "\"sakiko\"",
          position: "absolute",
          inset: "42px 0 auto 0",
          textAlign: "center",
          fontSize: { xs: 56, md: 92 },
          fontWeight: 800,
          letterSpacing: 8,
          color: alpha(theme.palette.primary.main, theme.palette.mode === "dark" ? 0.08 : 0.045),
          pointerEvents: "none",
          userSelect: "none",
        },
      })}
    >
      <Stack spacing={2} sx={{ position: "relative", p: { xs: 1.5, md: 2 } }}>
        <Box textAlign="center">
          <Typography variant="h6" fontWeight={800}>
            {heading}
          </Typography>
          <Typography variant="caption" color="text.secondary" className="sakiko-mono">
            Task: {detail.task.id}
          </Typography>
        </Box>

        <Divider />

        <Stack spacing={2}>
          {sections.map((section) => (
            <ReportSectionTable key={`${detail.task.id}-${section.kind}`} section={section} archive={detail} />
          ))}
        </Stack>

        <ReportFooter detail={detail} downloadTargets={downloadTargets} />
      </Stack>
    </Card>
  );
}

type ReportSectionTableProps = {
  section: ExportSection;
  archive: ResultArchive;
};

function ReportSectionTable({ section, archive }: ReportSectionTableProps) {
  const minWidth = Math.max(960, section.columns.reduce((sum, column) => sum + column.width, 0));

  return (
    <Stack spacing={0.75}>
      <Typography variant="subtitle1" textAlign="center" fontWeight={800}>
        {section.title}
      </Typography>
      <TableContainer
        className="sakiko-table"
        sx={(theme) => ({
          borderRadius: 0.75,
          borderColor: theme.palette.divider,
          bgcolor: "background.paper",
        })}
      >
        <Table size="small" stickyHeader sx={{ minWidth }}>
          <TableHead>
            <TableRow>
              {section.columns.map((column) => (
                <TableCell
                  key={column.key}
                  align={tableCellAlign(column)}
                  sx={(theme) => ({
                    width: column.width,
                    minWidth: Math.min(column.width, 260),
                    whiteSpace: "normal",
                    bgcolor: alpha(theme.palette.success.main, theme.palette.mode === "dark" ? 0.16 : 0.09),
                    borderColor: theme.palette.divider,
                    color: "text.primary",
                    fontWeight: 800,
                    fontSize: 12,
                    lineHeight: 1.2,
                  })}
                >
                  {column.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {section.rows.map((row, rowIndex) => (
              <TableRow key={`${section.kind}-${rowIndex}`} hover>
                {section.columns.map((column) => {
                  const value = row[column.key];
                  return (
                    <TableCell
                      key={column.key}
                      align={tableCellAlign(column)}
                      className={buildCellClassName(column.key, value)}
                      sx={(theme) => ({
                        width: column.width,
                        minWidth: Math.min(column.width, 260),
                        borderColor: theme.palette.divider,
                        bgcolor: resolveReportCellFill(theme),
                        color: "text.primary",
                        whiteSpace: shouldWrap(column.key) ? "normal" : "nowrap",
                        fontSize: 12.5,
                        py: 0.85,
                      })}
                      title={renderReportCellValue(section.kind, column.key, value, archive)}
                    >
                      <CellContent columnKey={column.key}>
                        {column.key === "perSecondBytesPerSecond" && Array.isArray(value) ? (
                          <SpeedSparkline values={value} />
                        ) : (
                          renderReportCellValue(section.kind, column.key, value, archive)
                        )}
                      </CellContent>
                    </TableCell>
                  );
                })}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Stack>
  );
}

function ReportFooter({ detail, downloadTargets }: ArchiveReportPreviewProps) {
  const config = detail.task.config;
  const runtimeSeconds = calculateRuntimeSeconds(detail);
  const target = summarizeDownloadTarget(config.downloadURL, downloadTargets);
  const title = summarizeResultMetrics(detail.task.context?.preset || "");
  const knightName = resolveArchiveKnightName(detail);

  return (
    <Stack spacing={0.7} sx={{ pt: 0.5 }}>
      <Divider />
      <Typography variant="caption">
        Protocol Library={formatProtocolLibraryLabel(detail.task.vendor)}&nbsp;&nbsp;
        Backend={formatBackendLabel(detail.task)}&nbsp;&nbsp;
        SpeedTest Target={target}&nbsp;&nbsp;
        Profile={detail.task.context?.profileName || "Unknown"}&nbsp;&nbsp;
        {knightName ? <>Knight={knightName}&nbsp;&nbsp;</> : null}
        Preset={title}
      </Typography>
      <Typography variant="caption">
        Nodes={detail.task.nodes.length}&nbsp;&nbsp;
        Threads={config.downloadThreading || 0}&nbsp;&nbsp;
        Duration={runtimeSeconds}s&nbsp;&nbsp;
        Ping Samples={config.pingAverageOver || 0}
      </Typography>
      <Typography variant="caption" color="text.secondary">
        Tested At: {formatDateTime(detail.state.finishedAt || detail.state.startedAt)}&nbsp;&nbsp;
        Results are for reference only.
      </Typography>
    </Stack>
  );
}

function resolveArchiveKnightName(archive: ResultArchive): string {
  return archive.task.environment?.remote?.knightName?.trim() || "";
}

type SummaryMetricProps = {
  label: string;
  value: string;
  mono?: boolean;
  title?: string;
};

function SummaryMetric({ label, value, mono = false, title }: SummaryMetricProps) {
  return (
    <Box
      sx={(theme) => ({
        px: 1.25,
        py: 1,
        borderRadius: 2,
        bgcolor: alpha(theme.palette.primary.main, 0.06),
        minWidth: 0,
      })}
    >
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography
        variant="body2"
        className={buildClassName(mono ? "sakiko-mono" : "", shouldUseEmojiFont(label, value) ? "sakiko-emoji" : "")}
        noWrap
        title={title || value}
      >
        {value}
      </Typography>
    </Box>
  );
}

function SpeedSparkline({ values }: { values: unknown[] }) {
  const numericValues = values
    .map((value) => typeof value === "number" ? value : Number.parseFloat(String(value)))
    .filter((value) => Number.isFinite(value) && value >= 0);
  const max = Math.max(1, ...numericValues);

  return (
    <Stack direction="row" spacing={0.25} alignItems="flex-end" justifyContent="center" sx={{ height: 24, minWidth: 96 }}>
      {numericValues.slice(0, 24).map((value, index) => (
        <Box
          key={`${value}-${index}`}
          sx={(theme) => ({
            width: 4,
            height: Math.max(3, Math.round((value / max) * 22)),
            borderRadius: 0.5,
            bgcolor: alpha(theme.palette.success.main, 0.35 + (value / max) * 0.5),
          })}
        />
      ))}
    </Stack>
  );
}

function CellContent({ columnKey, children }: { columnKey: string; children: ReactNode }) {
  if (columnKey === "perSecondBytesPerSecond") {
    return <>{children}</>;
  }

  return (
    <Box
      sx={{
        display: "-webkit-box",
        maxHeight: 40,
        overflow: "hidden",
        textOverflow: "ellipsis",
        WebkitBoxOrient: "vertical",
        WebkitLineClamp: shouldWrap(columnKey) ? 2 : 1,
        wordBreak: shouldWrap(columnKey) ? "break-word" : "normal",
        lineHeight: 1.35,
      }}
    >
      {children}
    </Box>
  );
}

function renderReportCellValue(kind: string, key: string, value: unknown, archive: ResultArchive): string {
  if (key === "error" && (value === null || value === undefined || String(value).trim() === "")) {
    return "OK";
  }
  if (key === "nodeName" && typeof value === "string") {
    return truncateText(value, archive.task.context?.preset === "full" ? 36 : 42);
  }
  if (kind === "media_unlock_table" && key !== "nodeName" && key !== "proxyType") {
    const text = formatReportValue(value, key);
    return text.toLowerCase().includes("failed") ? "Test Failed" : text;
  }
  if (key === "outboundIP" || key === "inboundIP") {
    return maskIPAddress(value);
  }
  return formatReportValue(value, key);
}

function resolveReportCellFill(theme: Theme): string {
  return theme.palette.background.paper;
}

function tableCellAlign(column: ExportColumn): "left" | "center" | "right" | "justify" | "inherit" {
  if (column.align === "center" || column.align === "right" || column.align === "left") {
    return column.align;
  }
  return "left";
}

function shouldUseMono(key: string): boolean {
  return key.toLowerCase().includes("ip")
    || key.toLowerCase().includes("asn")
    || key.toLowerCase().includes("endpoint")
    || key.toLowerCase().includes("trafficusedbytes")
    || key.toLowerCase().includes("bytespersecond")
    || key.toLowerCase().includes("millis");
}

function shouldWrap(key: string): boolean {
  return key.toLowerCase().includes("organization")
    || key.toLowerCase().includes("info")
    || key.toLowerCase().includes("endpoint")
    || key.toLowerCase().includes("metrics")
    || key.toLowerCase().includes("error")
    || key.toLowerCase().includes("persecond");
}

function buildCellClassName(key: string, value: unknown): string {
  return buildClassName(
    shouldUseMono(key) ? "sakiko-mono" : "",
    shouldUseEmojiFont(key, formatReportValue(value, key)) ? "sakiko-emoji" : "",
  ) || "";
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

function truncateText(value: string, limit: number): string {
  const text = value.trim();
  if (text.length <= limit) {
    return text;
  }
  return `${text.slice(0, Math.max(0, limit - 3))}...`;
}

function buildClassName(...parts: string[]): string | undefined {
  const value = parts.filter(Boolean).join(" ").trim();
  return value || undefined;
}
