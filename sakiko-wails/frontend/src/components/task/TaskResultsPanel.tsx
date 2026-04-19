import InsightsRounded from "@mui/icons-material/InsightsRounded";
import StopCircleRounded from "@mui/icons-material/StopCircleRounded";
import { Box, Button, Card, Chip, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import type { TaskActiveNode, TaskStatusResponse } from "../../types/sakiko";
import { describeTaskActiveNode, formatMacroLabel, formatMatrixLabel, formatTaskRuntimePhase, formatTaskStatus, shouldUseEmojiFont, summarizeActiveTaskNodes } from "../../utils/dashboard";
import { buildMediaMatrixFromResults } from "../../utils/mediaMatrix";
import { MediaUnlockMatrix } from "../media/MediaUnlockMatrix";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";
import { ResultEntryCard } from "./ResultEntryCard";

type TaskResultsPanelProps = {
  activeTask: TaskStatusResponse | null;
  onStopTask: () => Promise<void>;
};

export function TaskResultsPanel({ activeTask, onStopTask }: TaskResultsPanelProps) {
  const { t } = useTranslation();
  const mediaMatrix = buildMediaMatrixFromResults(activeTask?.results || []);
  const activeNodes = activeTask?.task?.activeNodes || [];
  const activeSummary = summarizeActiveTaskNodes(activeNodes);
  const isCancelable = activeTask?.task?.status === "running" || activeTask?.task?.status === "stopping";
  const isStopping = activeTask?.task?.status === "stopping";

  return (
    <SectionCard
      title={t("dashboard.tasks.results.title")}
      subtitle={activeTask?.task?.name || t("shared.states.selectTask")}
      icon={<InsightsRounded color="primary" />}
      action={isCancelable ? (
        <Button
          size="small"
          variant="outlined"
          color="error"
          startIcon={<StopCircleRounded />}
          onClick={() => void onStopTask()}
          disabled={isStopping}
        >
          {t("shared.actions.stop", "Stop")}
        </Button>
      ) : undefined}
    >
      {activeTask ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "repeat(4, minmax(0, 1fr))" },
              gap: 1.5,
            }}
          >
            <ResultMetric label={t("dashboard.tasks.results.labels.status")} value={formatTaskStatus(activeTask.task.status)} />
            <ResultMetric label={t("dashboard.tasks.results.labels.exit")} value={`${activeTask.exitCode || t("shared.states.pending")}`} />
            <ResultMetric label={t("dashboard.tasks.results.labels.progress")} value={`${activeTask.task.progress}/${activeTask.task.total}`} mono />
            <ResultMetric
              label={t("dashboard.tasks.results.labels.liveWorkload")}
              value={activeTask.task.status === "running" ? t("shared.formats.nodesCount", { count: activeNodes.length }) : t("shared.states.idle")}
              mono={activeTask.task.status === "running"}
            />
          </Box>

          {activeTask.task.status === "running" ? (
            <Stack spacing={1.25}>
              <Typography variant="subtitle2" color="text.secondary">
                {t("dashboard.tasks.results.currentlyTesting")}
              </Typography>
              <Typography variant="body2" color="text.secondary" className={shouldUseEmojiFont("nodeName", activeSummary) ? "sakiko-emoji" : undefined}>
                {activeSummary || t("dashboard.tasks.results.waitingForExecution")}
              </Typography>
              {activeNodes.length > 0 ? (
                <div className="sakiko-results-grid">
                  {activeNodes.map((activeNode) => (
                    <ActiveNodeCard key={`${activeNode.nodeIndex}-${activeNode.updatedAt || activeNode.phase}`} activeNode={activeNode} />
                  ))}
                </div>
              ) : null}
            </Stack>
          ) : null}

          {(activeTask.results || []).length > 0 ? (
            <Stack spacing={2}>
              {mediaMatrix.columns.length > 0 ? (
                <MediaUnlockMatrix
                  columns={mediaMatrix.columns}
                  rows={mediaMatrix.rows}
                  title={t("dashboard.tasks.results.mediaUnlockTitle")}
                  subtitle={t("dashboard.tasks.results.mediaUnlockSubtitle")}
                />
              ) : null}

              <div className="sakiko-results-grid">
                {(activeTask.results || []).map((result, index) => (
                  <ResultEntryCard key={`${result.proxyInfo.name}-${index}`} result={result} />
                ))}
              </div>
            </Stack>
          ) : (
            <EmptyState
              title={t("dashboard.tasks.results.noResultsTitle")}
              description={t("dashboard.tasks.results.noResultsDescription")}
            />
          )}
        </Stack>
      ) : (
        <EmptyState
          title={t("dashboard.tasks.results.noTaskTitle")}
          description={t("dashboard.tasks.results.noTaskDescription")}
        />
      )}
    </SectionCard>
  );
}

type ActiveNodeCardProps = {
  activeNode: TaskActiveNode;
};

function ActiveNodeCard({ activeNode }: ActiveNodeCardProps) {
  const { t } = useTranslation();
  const phaseLabel = formatTaskRuntimePhase(activeNode.phase);
  const matrixLabels = activeNode.matrix
    ? [formatMatrixLabel(activeNode.matrix)]
    : (activeNode.matrices || []).map((matrix) => formatMatrixLabel(matrix));

  return (
    <Card variant="outlined" sx={{ p: 2, height: "100%", minWidth: 0 }}>
      <Stack spacing={1.25} sx={{ minWidth: 0, height: "100%" }}>
        <Stack direction="row" justifyContent="space-between" spacing={1} sx={{ minWidth: 0, alignItems: "flex-start" }}>
          <Box sx={{ minWidth: 0, flex: "1 1 auto" }}>
            <Typography
              variant="subtitle1"
              noWrap
              title={activeNode.nodeName || t("shared.formats.nodeNumber", { index: activeNode.nodeIndex + 1 })}
              className={shouldUseEmojiFont("nodeName", activeNode.nodeName) ? "sakiko-emoji" : undefined}
            >
              {activeNode.nodeName || t("shared.formats.nodeNumber", { index: activeNode.nodeIndex + 1 })}
            </Typography>
            <Typography
              variant="body2"
              color="text.secondary"
              noWrap
              title={activeNode.nodeAddress || t("shared.states.addressPending")}
            >
              {activeNode.nodeAddress || t("shared.states.addressPending")}
            </Typography>
          </Box>
          <Chip
            label={(activeNode.attempt || 0) > 1 ? t("shared.formats.attempt", { count: activeNode.attempt }) : phaseLabel}
            size="small"
            color="primary"
            variant="outlined"
          />
        </Stack>

        <Typography variant="body2" sx={{ overflowWrap: "anywhere", wordBreak: "break-word" }}>
          {describeTaskActiveNode(activeNode)}
        </Typography>

        <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }} useFlexGap>
          <Chip label={phaseLabel} size="small" variant="outlined" />
          {activeNode.macro ? (
            <Chip label={formatMacroLabel(activeNode.macro)} size="small" variant="outlined" />
          ) : null}
          {matrixLabels.map((label) => (
            <Chip key={label} label={label} size="small" variant="outlined" />
          ))}
        </Stack>
      </Stack>
    </Card>
  );
}

type ResultMetricProps = {
  label: string;
  value: string;
  mono?: boolean;
};

function ResultMetric({ label, value, mono = false }: ResultMetricProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        p: 1.75,
        bgcolor: "background.default",
        borderColor: "divider",
        minWidth: 0,
      }}
    >
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography
        className={mono ? "sakiko-mono" : undefined}
        fontWeight={600}
        noWrap
        title={value}
      >
        {value}
      </Typography>
    </Card>
  );
}
