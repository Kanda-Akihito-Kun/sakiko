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

  return (
    <SectionCard
      title={t("dashboard.tasks.launcher.title")}
      subtitle={t("dashboard.tasks.launcher.subtitle")}
      icon={<TuneRounded color="primary" />}
    >
      <Stack spacing={2}>
        <Box sx={{ maxWidth: "100%", overflowX: "auto", pb: 0.25 }}>
          <ToggleButtonGroup
            value={taskPreset}
            sx={{
              display: "inline-flex",
              flexWrap: "nowrap",
              width: "max-content",
              minWidth: "100%",
              "& .MuiToggleButton-root": {
                flex: "0 0 auto",
                whiteSpace: "nowrap",
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

        <Stack spacing={1.25}>
          <Stack direction="row" justifyContent="flex-end" alignItems="flex-start" spacing={1} sx={{ minWidth: 0 }}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<SettingsRounded />}
              onClick={onOpenSettings}
            >
              {t("shared.actions.openSettings")}
            </Button>
          </Stack>

          <Box
            sx={{
              borderRadius: 2,
              border: (theme) => `1px solid ${theme.palette.divider}`,
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
        </Stack>

        <Stack direction="row" spacing={1.5} alignItems="center" useFlexGap flexWrap="wrap" sx={{ minWidth: 0 }}>
          <Button
            variant="contained"
            startIcon={<PlayCircleFilledWhiteRounded />}
            disabled={submitting || !activeProfileId || taskPreset.filter((preset) => preset !== "full").length === 0}
            onClick={onRunTask}
          >
            {t("dashboard.tasks.launcher.runPreset", { preset: formatTaskPresetSelectionLabel(taskPreset) })}
          </Button>
          <Chip
            icon={<ScheduleRounded />}
            label={activeProfileId ? t("dashboard.tasks.launcher.targetProfileSelected") : t("dashboard.tasks.launcher.selectProfileToRun")}
            variant="outlined"
            color={activeProfileId ? "success" : "default"}
            sx={{
              maxWidth: "100%",
              "& .MuiChip-label": {
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              },
            }}
          />
        </Stack>

        <List disablePadding>
          {tasks.map((task) => {
            const activeSummary = task.status === "running" ? summarizeActiveTaskNodes(task.activeNodes || []) : "";

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
                    variant={task.total > 0 ? "determinate" : "indeterminate"}
                    value={task.total > 0 ? (task.progress / task.total) * 100 : 0}
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
