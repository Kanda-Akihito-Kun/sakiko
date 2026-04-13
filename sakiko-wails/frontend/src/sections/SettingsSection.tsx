import SettingsRounded from "@mui/icons-material/SettingsRounded";
import { Box, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { TaskDefaultsPanel } from "../components/settings/TaskDefaultsPanel";
import { SettingsPanel } from "../components/settings/SettingsPanel";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";
import type { DownloadTarget, TaskConfig } from "../types/sakiko";

type SettingsSectionProps = {
  downloadTargets: DownloadTarget[];
  downloadTargetSearch: string;
  downloadTargetsLoading: boolean;
  mode: string;
  profilesPath: string;
  resolvedMode: string;
  taskConfig: TaskConfig;
  onPatchTaskConfig: (patch: Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading" | "backendIdentity">>) => void;
  onDownloadTargetSearchChange: (value: string) => void;
  onRefreshDownloadTargets: (search?: string) => Promise<void>;
};

export function SettingsSection({
  downloadTargets,
  downloadTargetSearch,
  downloadTargetsLoading,
  mode,
  profilesPath,
  resolvedMode,
  taskConfig,
  onPatchTaskConfig,
  onDownloadTargetSearchChange,
  onRefreshDownloadTargets,
}: SettingsSectionProps) {
  const { t } = useTranslation();

  return (
    <Box className="sakiko-section-grid">
      <Box className="sakiko-section-grid__sidebar">
        <SettingsPanel />
      </Box>

      <Box className="sakiko-section-grid__content">
        <TaskDefaultsPanel
          downloadTargets={downloadTargets}
          downloadTargetSearch={downloadTargetSearch}
          downloadTargetsLoading={downloadTargetsLoading}
          taskConfig={taskConfig}
          onPatchTaskConfig={onPatchTaskConfig}
          onDownloadTargetSearchChange={onDownloadTargetSearchChange}
          onRefreshDownloadTargets={onRefreshDownloadTargets}
        />

        <SectionCard
          title={t("settings.environment.title")}
          subtitle={t("settings.environment.subtitle")}
          icon={<SettingsRounded color="primary" />}
        >
          <Stack spacing={1.25}>
            <OverviewRow
              label={t("settings.environment.mode")}
              value={mode === "system" ? `${t("settings.themeOptions.system")} -> ${t(`settings.themeOptions.${resolvedMode}`)}` : t(`settings.themeOptions.${resolvedMode}`)}
              mono
            />
            <OverviewRow
              label={t("settings.environment.workspace")}
              value={profilesPath || t("settings.environment.workspaceUnavailable")}
              mono
            />
          </Stack>
        </SectionCard>
      </Box>
    </Box>
  );
}
