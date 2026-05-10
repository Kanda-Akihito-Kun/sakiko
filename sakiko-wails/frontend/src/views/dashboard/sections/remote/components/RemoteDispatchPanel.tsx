import LaunchRounded from "@mui/icons-material/LaunchRounded";
import PlaylistPlayRounded from "@mui/icons-material/PlaylistPlayRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import { Alert, Box, Button, Chip, Stack, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { taskPresets } from "../../../../../constants/dashboard";
import type { TaskPreset, TaskPresetSelection } from "../../../../../types/dashboard";
import type { ClusterRemoteTask, ClusterStatus, TaskConfig } from "../../../../../types/sakiko";
import { formatTaskPresetLabel, formatTaskPresetSelectionLabel, summarizeDownloadTarget } from "../../../../../utils/dashboard";
import { SectionCard } from "../../../../../components/shared/SectionCard";

type RemoteDispatchPanelProps = {
  activeProfileName?: string;
  currentRole: ClusterStatus["role"];
  remoteSubmitting: boolean;
  remoteTasks: ClusterRemoteTask[];
  selectedKnightCount: number;
  taskConfig: TaskConfig;
  taskPreset: TaskPresetSelection;
  onDispatchTask: () => void;
  onOpenProfiles: () => void;
  onOpenConfigs: () => void;
  onTaskPresetChange: (preset: TaskPreset) => void;
};

export function RemoteDispatchPanel({
  activeProfileName,
  currentRole,
  remoteSubmitting,
  remoteTasks,
  selectedKnightCount,
  taskConfig,
  taskPreset,
  onDispatchTask,
  onOpenProfiles,
  onOpenConfigs,
  onTaskPresetChange,
}: RemoteDispatchPanelProps) {
  const { t } = useTranslation();
  const activeTaskCount = remoteTasks.filter((task) => task.state === "queued" || task.state === "running").length;
  const hasActiveProfile = Boolean(activeProfileName);
  const canDispatch = currentRole === "master" && selectedKnightCount > 0 && hasActiveProfile && !remoteSubmitting;
  const selectedPresetLabel = formatTaskPresetSelectionLabel(taskPreset);
  const dispatchLabel = resolveDispatchLabel({
    activeTaskCount,
    canDispatch,
    currentRole,
    hasActiveProfile,
    remoteSubmitting,
    selectedKnightCount,
    t,
  });

  return (
    <SectionCard
      title={t("dashboard.remote.remoteTaskWorkspace")}
      icon={<TuneRounded color="primary" />}
      action={(
        <Stack direction="row" spacing={1}>
          <Button size="small" variant="outlined" startIcon={<PlaylistPlayRounded />} onClick={onOpenProfiles}>
            {t("shared.actions.openProfiles")}
          </Button>
          <Button size="small" variant="outlined" startIcon={<SettingsRounded />} onClick={onOpenConfigs}>
            {t("shared.actions.openConfigs")}
          </Button>
        </Stack>
      )}
    >
      <Stack spacing={2}>
        <Alert severity="info">
          {t("dashboard.remote.remoteTaskWorkspaceHint")}
        </Alert>

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

        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          <Chip size="small" label={t("dashboard.remote.dispatchProfile", { name: activeProfileName || t("dashboard.workspace.noActiveProfile") })} />
          <Chip size="small" label={t("dashboard.remote.dispatchPreset", { preset: selectedPresetLabel })} />
          <Chip size="small" label={t("dashboard.remote.selectedKnights", { count: selectedKnightCount })} />
          <Chip size="small" label={t("shared.formats.timeoutMillis", { value: taskConfig.taskTimeoutMillis })} variant="outlined" />
          <Chip size="small" label={t("shared.formats.durationSeconds", { value: taskConfig.downloadDuration })} variant="outlined" />
          <Chip
            size="small"
            label={taskConfig.downloadThreading > 1 ? t("shared.formats.threadsCount", { count: taskConfig.downloadThreading }) : t("dashboard.tasks.launcher.singleThread")}
            variant="outlined"
          />
          <Chip size="small" label={summarizeDownloadTarget(taskConfig.downloadURL)} variant="outlined" />
          <Chip
            size="small"
            color={activeTaskCount > 0 ? "warning" : "default"}
            label={t("dashboard.remote.activeRemoteTasks", { count: activeTaskCount })}
            variant={activeTaskCount > 0 ? "filled" : "outlined"}
          />
        </Stack>

        <Typography variant="caption" color="text.secondary">
          {t("dashboard.remote.remoteTaskWorkspaceDetail")}
        </Typography>

        <Button
          variant={activeTaskCount > 0 ? "outlined" : "contained"}
          startIcon={<LaunchRounded />}
          disabled={!canDispatch}
          onClick={onDispatchTask}
        >
          {dispatchLabel}
        </Button>
      </Stack>
    </SectionCard>
  );
}

function resolveDispatchLabel(args: {
  activeTaskCount: number;
  canDispatch: boolean;
  currentRole: ClusterStatus["role"];
  hasActiveProfile: boolean;
  remoteSubmitting: boolean;
  selectedKnightCount: number;
  t: ReturnType<typeof useTranslation>["t"];
}) {
  const {
    activeTaskCount,
    currentRole,
    hasActiveProfile,
    remoteSubmitting,
    selectedKnightCount,
    t,
  } = args;

  if (remoteSubmitting) {
    return t("dashboard.remote.dispatchingTask");
  }
  if (currentRole !== "master") {
    return t("dashboard.remote.masterOnlyDispatch");
  }
  if (!hasActiveProfile) {
    return t("dashboard.remote.selectProfileFirst");
  }
  if (selectedKnightCount <= 0) {
    return t("dashboard.remote.selectKnightFirst");
  }
  if (activeTaskCount > 0) {
    return t("dashboard.remote.dispatchAnotherTask");
  }
  return t("dashboard.remote.dispatchTask");
}
