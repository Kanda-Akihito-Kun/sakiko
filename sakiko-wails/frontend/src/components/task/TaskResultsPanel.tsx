import InsightsRounded from "@mui/icons-material/InsightsRounded";
import QueryStatsRounded from "@mui/icons-material/QueryStatsRounded";
import { alpha } from "@mui/material/styles";
import { Box, Card, Chip, Divider, Stack, Typography } from "@mui/material";
import type { TaskStatusResponse } from "../../types/sakiko";
import { formatDuration, formatMatrixPayload } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type TaskResultsPanelProps = {
  activeTask: TaskStatusResponse | null;
};

export function TaskResultsPanel({ activeTask }: TaskResultsPanelProps) {
  return (
    <SectionCard
      title="Task Results"
      subtitle={activeTask?.task?.name || "Select a task"}
      icon={<InsightsRounded color="primary" />}
    >
      {activeTask ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0, 1fr))" },
              gap: 1.5,
            }}
          >
            <ResultMetric label="Status" value={activeTask.task.status} />
            <ResultMetric label="Exit" value={`${activeTask.exitCode || "pending"}`} />
            <ResultMetric label="Progress" value={`${activeTask.task.progress}/${activeTask.task.total}`} mono />
          </Box>

          {(activeTask.results || []).length > 0 ? (
            <div className="sakiko-results-grid">
              {(activeTask.results || []).map((result, index) => (
                <Card key={`${result.proxyInfo.name}-${index}`} variant="outlined" sx={{ p: 2.25 }}>
                  <Stack spacing={1.5}>
                    <Stack direction="row" justifyContent="space-between" spacing={1}>
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="h6" noWrap>
                          {result.proxyInfo.name || "Unnamed node"}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" noWrap>
                          {result.proxyInfo.address || "No address"}
                        </Typography>
                      </Box>
                      <Chip
                        label={result.proxyInfo.type || "unknown"}
                        size="small"
                        color="primary"
                        variant="outlined"
                      />
                    </Stack>

                    <Stack direction="row" spacing={1} sx={{ flexWrap: "nowrap", overflowX: "auto", pb: 0.25 }}>
                      <Chip
                        icon={<QueryStatsRounded />}
                        label={formatDuration(result.invokeDuration)}
                        size="small"
                      />
                      <Chip
                        label={`${result.matrices.length} matrices`}
                        size="small"
                        variant="outlined"
                      />
                    </Stack>

                    {result.error && (
                      <Typography variant="body2" color="error.main">
                        {result.error}
                      </Typography>
                    )}

                    <Divider sx={{ borderColor: "divider" }} />
                    <Stack spacing={1}>
                      {result.matrices.map((matrix, matrixIndex) => (
                        <Box
                          key={`${matrix.type}-${matrixIndex}`}
                          sx={(theme) => ({
                            display: "flex",
                            justifyContent: "space-between",
                            gap: 1.5,
                            p: 1.25,
                            borderRadius: 2,
                            bgcolor: alpha(theme.palette.primary.main, 0.08),
                          })}
                        >
                          <Typography variant="body2" color="text.secondary">
                            {matrix.type}
                          </Typography>
                          <Typography variant="body2" className="sakiko-mono" noWrap sx={{ minWidth: 0, textAlign: "right" }}>
                            {formatMatrixPayload(matrix.payload, matrix.type)}
                          </Typography>
                        </Box>
                      ))}
                    </Stack>
                  </Stack>
                </Card>
              ))}
            </div>
          ) : (
            <EmptyState
              title="No results yet"
              description="Run a task or wait for the current job to finish."
            />
          )}
        </Stack>
      ) : (
        <EmptyState
          title="No task selected"
          description="Recent tasks appear here with per-node results."
        />
      )}
    </SectionCard>
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
      }}
    >
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography className={mono ? "sakiko-mono" : undefined} fontWeight={600} noWrap>
        {value}
      </Typography>
    </Card>
  );
}
