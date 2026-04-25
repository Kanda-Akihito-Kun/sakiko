import { Box } from "@mui/material";
import { DNSSettingsPanel } from "../components/settings/DNSSettingsPanel";
import { TaskDefaultsPanel } from "../components/settings/TaskDefaultsPanel";
import type { DownloadTarget, TaskConfig } from "../types/sakiko";

type OverviewSectionProps = {
  downloadTargets: DownloadTarget[];
  downloadTargetSearch: string;
  downloadTargetsLoading: boolean;
  taskConfig: TaskConfig;
  onPatchTaskConfig: (patch: Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading" | "backendIdentity">>) => void;
  onDownloadTargetSearchChange: (value: string) => void;
  onRefreshDownloadTargets: (search?: string) => Promise<void>;
};

export function OverviewSection({
  downloadTargets,
  downloadTargetSearch,
  downloadTargetsLoading,
  taskConfig,
  onPatchTaskConfig,
  onDownloadTargetSearchChange,
  onRefreshDownloadTargets,
}: OverviewSectionProps) {
  return (
    <Box className="sakiko-overview-grid">
      <Box sx={{ gridColumn: "1 / -1" }}>
        <TaskDefaultsPanel
          downloadTargets={downloadTargets}
          downloadTargetSearch={downloadTargetSearch}
          downloadTargetsLoading={downloadTargetsLoading}
          taskConfig={taskConfig}
          onPatchTaskConfig={onPatchTaskConfig}
          onDownloadTargetSearchChange={onDownloadTargetSearchChange}
          onRefreshDownloadTargets={onRefreshDownloadTargets}
        />
      </Box>
      <Box sx={{ gridColumn: "1 / -1" }}>
        <DNSSettingsPanel />
      </Box>
    </Box>
  );
}
