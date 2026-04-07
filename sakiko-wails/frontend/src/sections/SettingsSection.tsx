import SettingsRounded from "@mui/icons-material/SettingsRounded";
import { Box, Stack } from "@mui/material";
import { TaskDefaultsPanel } from "../components/settings/TaskDefaultsPanel";
import { SettingsPanel } from "../components/settings/SettingsPanel";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";
import type { DownloadTarget, TaskConfig } from "../types/sakiko";

type SettingsSectionProps = {
  downloadTargets: DownloadTarget[];
  downloadTargetsLoading: boolean;
  mode: string;
  profilesPath: string;
  resolvedMode: string;
  taskConfig: TaskConfig;
  onPatchTaskConfig: (patch: Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading">>) => void;
  onRefreshDownloadTargets: () => Promise<void>;
};

export function SettingsSection({
  downloadTargets,
  downloadTargetsLoading,
  mode,
  profilesPath,
  resolvedMode,
  taskConfig,
  onPatchTaskConfig,
  onRefreshDownloadTargets,
}: SettingsSectionProps) {
  return (
    <Box className="sakiko-section-grid">
      <Box className="sakiko-section-grid__sidebar">
        <SettingsPanel />
      </Box>

      <Box className="sakiko-section-grid__content">
        <TaskDefaultsPanel
          downloadTargets={downloadTargets}
          downloadTargetsLoading={downloadTargetsLoading}
          taskConfig={taskConfig}
          onPatchTaskConfig={onPatchTaskConfig}
          onRefreshDownloadTargets={onRefreshDownloadTargets}
        />

        <SectionCard
          title="Environment"
          subtitle="Current appearance and workspace state"
          icon={<SettingsRounded color="primary" />}
        >
          <Stack spacing={1.25}>
            <OverviewRow label="Mode" value={mode === "system" ? `system -> ${resolvedMode}` : resolvedMode} mono />
            <OverviewRow label="Workspace" value={profilesPath || "Profiles workspace unavailable"} mono />
          </Stack>
        </SectionCard>
      </Box>
    </Box>
  );
}
