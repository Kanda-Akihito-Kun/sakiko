import DeleteOutlineRounded from "@mui/icons-material/DeleteOutlineRounded";
import ExpandMoreRounded from "@mui/icons-material/ExpandMoreRounded";
import InsightsRounded from "@mui/icons-material/InsightsRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SaveAltRounded from "@mui/icons-material/SaveAltRounded";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
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
  summarizeDownloadTargetDetail,
  summarizeResultMetrics,
} from "../../utils/dashboard";
import { buildMediaMatrixFromSection } from "../../utils/mediaMatrix";
import { exportResultArchiveImage } from "../../utils/resultExport";
import { MediaUnlockMatrix } from "../media/MediaUnlockMatrix";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";
import { ResultEntryCard } from "../task/ResultEntryCard";

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
}: ResultsArchivePanelProps) {
  const { resolvedExportPictureMode } = useThemeMode();
  const { t } = useTranslation();
  const [expandedTaskIds, setExpandedTaskIds] = useState<string[]>([]);
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

  return (
    <SectionCard
      title={t("dashboard.results.title")}
      subtitle={t("dashboard.results.archivedCount", { count: archives.length })}
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
          title={t("dashboard.results.noArchivesTitle")}
          description={t("dashboard.results.noArchivesDescription")}
        />
      ) : (
        <Stack spacing={1.5}>
          {visibleArchives.map((archive) => {
            const detail = archiveDetails[archive.taskId];
            const detailLoading = Boolean(archiveLoading[archive.taskId]);
            const expanded = expandedTaskIds.includes(archive.taskId);

            return (
              <Accordion
                key={archive.taskId}
                expanded={expanded}
                onChange={(_event, nextExpanded) => {
                  setExpandedTaskIds((current) => {
                    if (nextExpanded) {
                      return current.includes(archive.taskId) ? current : [...current, archive.taskId];
                    }
                    return current.filter((taskId) => taskId !== archive.taskId);
                  });
                  if (nextExpanded && !detail && !detailLoading) {
                    void onEnsureArchive(archive.taskId);
                  }
                }}
                disableGutters
                elevation={0}
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 2,
                  overflow: "hidden",
                  "&::before": {
                    display: "none",
                  },
                }}
              >
                <AccordionSummary
                  expandIcon={<ExpandMoreRounded />}
                  sx={{
                    px: 2,
                    py: 1.25,
                    bgcolor: "background.paper",
                    "& .MuiAccordionSummary-content": {
                      margin: 0,
                      minWidth: 0,
                    },
                  }}
                >
                  <Stack spacing={1.25} width="100%">
                    <Stack
                      direction="row"
                      justifyContent="space-between"
                      spacing={1}
                      sx={{ minWidth: 0, alignItems: "flex-start", flexWrap: { xs: "wrap", lg: "nowrap" } }}
                    >
                      <Box sx={{ minWidth: 0, flex: "1 1 320px" }}>
                        <Typography variant="subtitle1" noWrap title={archive.taskName || archive.taskId}>
                          {archive.taskName || archive.taskId}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" noWrap title={archive.taskId}>
                          {archive.taskId}
                        </Typography>
                      </Box>
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
                            const archiveName = archive.taskName || archive.taskId;
                            if (!window.confirm(t("dashboard.results.deleteConfirm", { name: archiveName }))) {
                              return;
                            }
                            setDeletingTaskId(archive.taskId);
                            try {
                              await onDeleteArchive(archive.taskId);
                              setExpandedTaskIds((current) => current.filter((taskId) => taskId !== archive.taskId));
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
                              await exportResultArchiveImage(nextDetail, downloadTargets, resolvedExportPictureMode);
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
                </AccordionSummary>

                <AccordionDetails sx={{ px: 2, py: 2 }}>
                  {detailLoading && !detail ? (
                    <Stack direction="row" spacing={1.25} alignItems="center" py={1}>
                      <CircularProgress size={20} />
                      <Typography variant="body2" color="text.secondary">
                        {t("dashboard.results.loadingArchive")}
                      </Typography>
                    </Stack>
                  ) : detail ? (
                    <ArchiveDetail detail={detail} downloadTargets={downloadTargets} />
                  ) : (
                    <EmptyState
                      title={t("dashboard.results.archiveUnavailableTitle")}
                      description={t("dashboard.results.archiveUnavailableDescription")}
                    />
                  )}
                </AccordionDetails>
              </Accordion>
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

type ArchiveDetailProps = {
  detail: ResultArchive;
  downloadTargets: DownloadTarget[];
};

function ArchiveDetail({ detail, downloadTargets }: ArchiveDetailProps) {
  const { t } = useTranslation();
  const targetSummary = summarizeDownloadTarget(detail.task.config.downloadURL, downloadTargets);
  const targetDetail = summarizeDownloadTargetDetail(detail.task.config.downloadURL, downloadTargets);

  return (
    <Stack spacing={2}>
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "repeat(4, minmax(0, 1fr))" },
          gap: 1.5,
        }}
      >
        <DetailMetric label={t("dashboard.results.detail.preset")} value={detail.task.context?.preset || t("shared.states.unknown")} />
        <DetailMetric label={t("dashboard.results.detail.protocolLibrary")} value={formatProtocolLibraryLabel(detail.task.vendor)} />
        <DetailMetric label={t("dashboard.results.detail.backend")} value={formatBackendLabel(detail.task)} title={formatBackendDetail(detail.task)} />
        <DetailMetric label={t("dashboard.results.detail.speedTarget")} value={targetSummary} title={targetDetail} />
        <DetailMetric label={t("dashboard.results.detail.exitCode")} value={detail.exitCode || t("shared.states.pending")} />
        <DetailMetric label={t("dashboard.results.detail.started")} value={formatDateTime(detail.state.startedAt)} />
        <DetailMetric label={t("dashboard.results.detail.finished")} value={formatDateTime(detail.state.finishedAt)} />
        <DetailMetric label={t("dashboard.results.detail.profileSource")} value={detail.task.context?.profileSource || t("shared.states.none")} mono />
      </Box>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
          gap: 1.5,
        }}
      >
        <DetailMetric label={t("dashboard.results.detail.pingUrl")} value={detail.task.config.pingAddress} mono />
        <DetailMetric label={t("dashboard.results.detail.timeout")} value={t("shared.formats.timeoutMillis", { value: detail.task.config.taskTimeoutMillis })} mono />
        <DetailMetric label={t("dashboard.results.detail.downloadUrl")} value={detail.task.config.downloadURL} mono title={detail.task.config.downloadURL} />
        <DetailMetric label={t("dashboard.results.detail.speedSettings")} value={`${t("shared.formats.durationSeconds", { value: detail.task.config.downloadDuration })} / ${t("shared.formats.threadsCount", { count: detail.task.config.downloadThreading })}`} />
      </Box>

      <Stack spacing={1}>
        <Typography variant="subtitle2" color="text.secondary">
          {t("dashboard.results.detail.testedNodes")}
        </Typography>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          {detail.task.nodes.map((node) => (
            <Chip
              key={node.name}
              size="small"
              label={node.name}
              variant="outlined"
              sx={buildEmojiChipSx("nodeName", node.name)}
            />
          ))}
        </Stack>
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2" color="text.secondary">
          {t("dashboard.results.detail.reportSections")}
        </Typography>
        {detail.report.sections?.length ? (
          detail.report.sections.map((section) => {
            const mediaMatrix = section.kind === "media_unlock_table"
              ? buildMediaMatrixFromSection(section)
              : null;
            if (section.kind === "media_unlock_table" && mediaMatrix && mediaMatrix.columns.length === 0) {
              return null;
            }

            return (
              <Card key={`${detail.task.id}-${section.kind}`} variant="outlined" sx={{ p: 1.5 }}>
                <Stack spacing={1.25}>
                  <Stack direction="row" justifyContent="space-between" spacing={1}>
                    <Box>
                      <Typography variant="subtitle1">
                        {section.kind === "media_unlock_table" ? t("dashboard.results.detail.mediaUnlockTitle") : section.title}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {section.kind}
                      </Typography>
                    </Box>
                    <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" justifyContent="flex-end">
                      {Object.entries(section.summary || {}).map(([key, value]) => (
                        <Chip
                          key={key}
                          size="small"
                          variant="outlined"
                          label={`${key}: ${formatReportValue(value, key)}`}
                          sx={buildEmojiChipSx(key, value)}
                        />
                      ))}
                    </Stack>
                  </Stack>

                  {mediaMatrix ? (
                    <MediaUnlockMatrix
                      columns={mediaMatrix.columns}
                      rows={mediaMatrix.rows}
                    />
                  ) : (
                    <Box className="sakiko-table">
                      <Table size="small" stickyHeader>
                        <TableHead>
                          <TableRow>
                            {section.columns?.map((column) => (
                              <TableCell key={column.key}>{column.label}</TableCell>
                            ))}
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {(section.rows || []).map((row, rowIndex) => (
                            <TableRow key={`${section.kind}-${rowIndex}`} hover>
                              {section.columns?.map((column) => (
                                <TableCell
                                  key={column.key}
                                  className={buildCellClassName(column.key, row[column.key])}
                                  sx={{ whiteSpace: shouldWrap(column.key) ? "normal" : "nowrap" }}
                                >
                                  {formatReportValue(row[column.key], column.key)}
                                </TableCell>
                              ))}
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </Box>
                  )}
                </Stack>
              </Card>
            );
          })
        ) : (
          <EmptyState
            title={t("dashboard.results.detail.noReportSectionsTitle")}
            description={t("dashboard.results.detail.noReportSectionsDescription")}
          />
        )}
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2" color="text.secondary">
          {t("dashboard.results.detail.rawResults")}
        </Typography>
        {(detail.results || []).length ? (
          <div className="sakiko-results-grid">
            {(detail.results || []).map((result, index) => (
              <ResultEntryCard key={`${result.proxyInfo.name}-${index}`} result={result} />
            ))}
          </div>
        ) : (
          <EmptyState
            title={t("dashboard.results.detail.noRawResultsTitle")}
            description={t("dashboard.results.detail.noRawResultsDescription")}
          />
        )}
      </Stack>
    </Stack>
  );
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

type DetailMetricProps = {
  label: string;
  value: string;
  mono?: boolean;
  title?: string;
};

function DetailMetric({ label, value, mono = false, title }: DetailMetricProps) {
  return (
    <Card variant="outlined" sx={{ p: 1.5, minWidth: 0 }}>
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography
        variant="body2"
        className={buildClassName(mono ? "sakiko-mono" : "", shouldUseEmojiFont(label, value) ? "sakiko-emoji" : "")}
        sx={{ wordBreak: "break-all" }}
        title={title || value}
      >
        {value}
      </Typography>
    </Card>
  );
}

function shouldUseMono(key: string): boolean {
  return key.toLowerCase().includes("ip")
    || key.toLowerCase().includes("asn")
    || key.toLowerCase().includes("trafficusedbytes")
    || key.toLowerCase().includes("bytespersecond")
    || key.toLowerCase().includes("millis");
}

function shouldWrap(key: string): boolean {
  return key.toLowerCase().includes("organization")
    || key.toLowerCase().includes("error")
    || key.toLowerCase().includes("persecond");
}

function buildCellClassName(key: string, value: unknown): string {
  return buildClassName(
    shouldUseMono(key) ? "sakiko-mono" : "",
    shouldUseEmojiFont(key, formatReportValue(value, key)) ? "sakiko-emoji" : "",
  ) || "";
}

function buildEmojiChipSx(key: string, value: unknown) {
  return shouldUseEmojiFont(key, value)
    ? {
        "& .MuiChip-label": {
          fontFamily: "var(--sakiko-font-sans)",
        },
      }
    : undefined;
}

function buildClassName(...parts: string[]): string | undefined {
  const value = parts.filter(Boolean).join(" ").trim();
  return value || undefined;
}
