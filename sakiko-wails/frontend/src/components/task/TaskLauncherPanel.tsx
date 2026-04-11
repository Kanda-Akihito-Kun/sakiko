import PlayCircleFilledWhiteRounded from "@mui/icons-material/PlayCircleFilledWhiteRounded";
import ScheduleRounded from "@mui/icons-material/ScheduleRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import { Box, Button, Chip, LinearProgress, List, ListItemButton, ListItemText, Stack, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import { taskPresets } from "../../constants/dashboard";
import type { TaskPreset } from "../../types/dashboard";
import type { TaskConfig, TaskState } from "../../types/sakiko";
import { summarizeDownloadTarget } from "../../utils/dashboard";
import { SectionCard } from "../shared/SectionCard";

type TaskLauncherPanelProps = {
  activeProfileId: string;
  activeTaskId: string;
  submitting: boolean;
  taskConfig: TaskConfig;
  taskPreset: TaskPreset;
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
  return (
    <SectionCard
      title="Task Launcher"
      subtitle="Choose a task preset"
      icon={<TuneRounded color="primary" />}
    >
      <Stack spacing={2}>
        <Box sx={{ maxWidth: "100%", overflowX: "auto", pb: 0.25 }}>
          <ToggleButtonGroup
            exclusive
            value={taskPreset}
            onChange={(_event, value: TaskPreset | null) => {
              if (value) {
                onTaskPresetChange(value);
              }
            }}
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
              <ToggleButton key={preset} value={preset}>
                {preset.toUpperCase()}
              </ToggleButton>
            ))}
          </ToggleButtonGroup>
        </Box>

        <Stack spacing={1.25}>
          <Stack direction="row" justifyContent="flex-end" alignItems="flex-start" spacing={1}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<SettingsRounded />}
              onClick={onOpenSettings}
            >
              Open Settings
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
              <Chip size="small" label={`Timeout ${taskConfig.taskTimeoutMillis} ms`} variant="outlined" />
              <Chip size="small" label={`Duration ${taskConfig.downloadDuration}s`} variant="outlined" />
              <Chip
                size="small"
                label={taskConfig.downloadThreading > 1 ? `${taskConfig.downloadThreading} threads` : "Single-thread"}
                variant="outlined"
              />
              <Chip size="small" label={summarizeDownloadTarget(taskConfig.downloadURL)} variant="outlined" />
            </Stack>
          </Box>
        </Stack>

        <Stack direction="row" spacing={1.5} alignItems="center">
          <Button
            variant="contained"
            startIcon={<PlayCircleFilledWhiteRounded />}
            disabled={submitting || !activeProfileId}
            onClick={onRunTask}
          >
            Run {taskPreset.toUpperCase()}
          </Button>
          <Chip
            icon={<ScheduleRounded />}
            label={activeProfileId ? "Target profile selected" : "Select a profile to run tasks"}
            variant="outlined"
            color={activeProfileId ? "success" : "default"}
          />
        </Stack>

        <List disablePadding>
          {tasks.map((task) => (
            <ListItemButton
              key={task.taskId}
              selected={task.taskId === activeTaskId}
              onClick={() => void onInspectTask(task.taskId)}
              sx={{ alignItems: "flex-start", flexDirection: "column" }}
            >
              <Stack width="100%" spacing={1}>
                <Stack direction="row" justifyContent="space-between" spacing={1}>
                  <ListItemText
                    primary={task.name}
                    secondary={task.status}
                    primaryTypographyProps={{ fontWeight: 600, noWrap: true }}
                    secondaryTypographyProps={{ noWrap: true }}
                    sx={{ minWidth: 0 }}
                  />
                  <Typography variant="caption" className="sakiko-mono" color="text.secondary">
                    {task.progress}/{task.total}
                  </Typography>
                </Stack>
                <LinearProgress
                  variant={task.total > 0 ? "determinate" : "indeterminate"}
                  value={task.total > 0 ? (task.progress / task.total) * 100 : 0}
                  sx={{ width: "100%", height: 6, borderRadius: 999 }}
                />
              </Stack>
            </ListItemButton>
          ))}
        </List>
      </Stack>
    </SectionCard>
  );
}
