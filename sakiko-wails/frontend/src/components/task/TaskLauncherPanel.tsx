import PlayCircleFilledWhiteRounded from "@mui/icons-material/PlayCircleFilledWhiteRounded";
import ScheduleRounded from "@mui/icons-material/ScheduleRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import { Box, Button, Chip, LinearProgress, List, ListItemButton, ListItemText, Stack, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { taskPresets } from "../../constants/dashboard";
import type { TaskPreset, TaskPresetSelection } from "../../types/dashboard";
import type { TaskConfig, TaskState } from "../../types/sakiko";
import { formatTaskPresetLabel, formatTaskPresetSelectionLabel, formatTaskStatus, shouldUseEmojiFont, summarizeActiveTaskNodes, summarizeDownloadTarget } from "../../utils/dashboard";
import { SectionCard } from "../shared/SectionCard";

type TaskLauncherPanelProps = {
  activeProfileId: string;
  activeTaskId: string;
  submitting: boolean;
  taskConfig: TaskConfig;
  taskPreset: TaskPresetSelection;
  tasks: TaskState[];
  onInspectTask: (taskId: string) => void;
  onOpenSettings: () => void;
  onRunTask: () => void;
  onTaskPresetChange: (preset: TaskPreset) => void;
};

function getTaskProgressValue(progress: number, total: number) {
  if (total <= 0) {
    return 0;
  }

  return Math.min(100, Math.max(0, (progress / total) * 100));
}

export function TaskLauncherPanel({
  activeProfileId,
  activeTaskId,
  submitting,
  taskConfig,
  taskPreset,
  tasks,
  onInspectTask,
  onOpenSettings,
  onRunTask,
  onTaskPresetChange,
}: TaskLauncherPanelProps) {
  const { t } = useTranslation();
  const hasActiveProfile = activeProfileId.trim().length > 0;
  const hasRunnablePreset = taskPreset.some((preset) => preset !== "full");
  const canRunTask = !submitting && hasActiveProfile && hasRunnablePreset;
  const selectedPresetLabel = formatTaskPresetSelectionLabel(taskPreset);
  const profileStatusLabel = hasActiveProfile
    ? t("dashboard.tasks.launcher.targetProfileSelected")
    : t("dashboard.tasks.launcher.selectProfileToRun");

  return (
    <SectionCard
      title={t("dashboard.tasks.launcher.title")}
      subtitle={t("dashboard.tasks.launcher.subtitle")}
      icon={<TuneRounded color="primary" />}
      action={(
        <Button
          size="small"
          variant="outlined"
          startIcon={<SettingsRounded />}
          onClick={onOpenSettings}
        >
          {t("shared.actions.openSettings")}
        </Button>
      )}
    >
      <Stack spacing={2}>
        <Box sx={{ maxWidth: "100%", overflowX: "auto", pb: 0.25 }}>
          <ToggleButtonGroup
            value={taskPreset}
            sx={{
              display: "flex",
              flexWrap: "wrap",
              gap: 0.75,
              width: "100%",
              "& .MuiToggleButton-root": {
                flex: "1 1 104px",
                whiteSpace: "nowrap",
                border: "1px solid",
                borderColor: "divider",
                borderRadius: 2,
                ml: "0 !important",
              },
            }}
          >
            {taskPresets.map((preset) => (
              <ToggleButton key={preset} value={preset} onClick={() => onTaskPresetChange(preset)}>
                {formatTaskPresetLabel(preset)}
              </ToggleButton>
            ))}
          </ToggleButtonGroup>
        </Box>

        <Box
          sx={{
            display: "grid",
            gap: 1.25,
            gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 280px), 1fr))",
            alignItems: "stretch",
            minWidth: 0,
            "& > *": {
              minWidth: 0,
            },
          }}
        >
          <Box
            sx={{
              borderRadius: 2,
              border: (theme) => `1px solid ${theme.palette.divider}`,
              bgcolor: "background.default",
              px: 1.5,
              py: 1.25,
            }}
          >
            <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
              <Chip size="small" label={t("shared.formats.timeoutMillis", { value: taskConfig.taskTimeoutMillis })} variant="outlined" />
              <Chip size="small" label={t("shared.formats.durationSeconds", { value: taskConfig.downloadDuration })} variant="outlined" />
              <Chip
                size="small"
                label={taskConfig.downloadThreading > 1 ? t("shared.formats.threadsCount", { count: taskConfig.downloadThreading }) : t("dashboard.tasks.launcher.singleThread")}
                variant="outlined"
              />
              <Chip size="small" label={summarizeDownloadTarget(taskConfig.downloadURL)} variant="outlined" />
            </Stack>
          </Box>

          <Box
            sx={{
              display: "grid",
              gap: 1,
              gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 200px), 1fr))",
              alignItems: "center",
            }}
          >
            <Button
              variant="contained"
              startIcon={<PlayCircleFilledWhiteRounded />}
              disabled={!canRunTask}
              onClick={onRunTask}
              sx={{ minWidth: 0 }}
            >
              {t("dashboard.tasks.launcher.runPreset", { preset: selectedPresetLabel })}
            </Button>
            <Chip
              icon={<ScheduleRounded />}
              label={profileStatusLabel}
              variant="outlined"
              color={hasActiveProfile ? "success" : "default"}
              sx={{
                width: "100%",
                maxWidth: "100%",
                "& .MuiChip-label": {
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                },
              }}
            />
          </Box>
        </Box>

        <List disablePadding>
          {tasks.map((task) => {
            const activeSummary = task.status === "running" ? summarizeActiveTaskNodes(task.activeNodes ?? []) : "";
            const hasKnownProgress = task.total > 0;
            const progressValue = getTaskProgressValue(task.progress, task.total);

            return (
              <ListItemButton
                key={task.taskId}
                selected={task.taskId === activeTaskId}
                onClick={() => void onInspectTask(task.taskId)}
                sx={{ alignItems: "flex-start", flexDirection: "column", minWidth: 0 }}
              >
                <Stack width="100%" spacing={1}>
                  <Stack direction="row" justifyContent="space-between" spacing={1} sx={{ minWidth: 0, alignItems: "flex-start" }}>
                    <ListItemText
                      primary={task.name}
                      secondary={formatTaskStatus(task.status)}
                      primaryTypographyProps={{
                        fontWeight: 600,
                        noWrap: true,
                        title: task.name,
                        className: shouldUseEmojiFont("nodeName", task.name) ? "sakiko-emoji" : undefined,
                      }}
                      secondaryTypographyProps={{ noWrap: true }}
                      sx={{ minWidth: 0 }}
                    />
                    <Typography variant="caption" className="sakiko-mono" color="text.secondary" sx={{ flexShrink: 0 }}>
                      {task.progress}/{task.total}
                    </Typography>
                  </Stack>
                  <LinearProgress
                    variant={hasKnownProgress ? "determinate" : "indeterminate"}
                    value={progressValue}
                    sx={{ width: "100%", height: 6, borderRadius: 999 }}
                  />
                  {activeSummary ? (
                    <Typography variant="caption" color="text.secondary" sx={{ overflowWrap: "anywhere", wordBreak: "break-word" }}>
                      {activeSummary}
                    </Typography>
                  ) : null}
                </Stack>
              </ListItemButton>
            );
          })}
        </List>
      </Stack>
    </SectionCard>
  );
}
