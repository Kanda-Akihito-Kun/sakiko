import { useEffect } from "react";
import { useShallow } from "zustand/react/shallow";
import { DNSSettingsPanel } from "../../../../components/settings/DNSSettingsPanel";
import { TaskDefaultsPanel } from "../../../../components/settings/TaskDefaultsPanel";
import { useDashboardStore } from "../../../../store/dashboardStore";
import { SectionLayout, SectionSpan } from "../../components/SectionLayout";

export function ConfigsDashboardSection() {
  const view = useDashboardStore(useShallow((state) => ({
    downloadTargets: state.downloadTargets,
    downloadTargetSearch: state.downloadTargetSearch,
    downloadTargetsLoading: state.downloadTargetsLoading,
    taskConfig: state.taskConfig,
    patchTaskConfig: state.patchTaskConfig,
    setDownloadTargetSearch: state.setDownloadTargetSearch,
    refreshDownloadTargets: state.refreshDownloadTargets,
  })));

  useEffect(() => {
    if (view.downloadTargets.length > 0 || view.downloadTargetsLoading) {
      return;
    }
    void view.refreshDownloadTargets();
  }, [view.downloadTargets.length, view.downloadTargetsLoading, view.refreshDownloadTargets]);

  return (
    <SectionLayout
      columns={{
        xs: "minmax(0, 1fr)",
        lg: "clamp(340px, 30vw, 420px) minmax(0, 1fr)",
      }}
    >
      <SectionSpan>
        <TaskDefaultsPanel
          downloadTargets={view.downloadTargets}
          downloadTargetSearch={view.downloadTargetSearch}
          downloadTargetsLoading={view.downloadTargetsLoading}
          taskConfig={view.taskConfig}
          onPatchTaskConfig={view.patchTaskConfig}
          onDownloadTargetSearchChange={view.setDownloadTargetSearch}
          onRefreshDownloadTargets={view.refreshDownloadTargets}
        />
      </SectionSpan>
      <SectionSpan>
        <DNSSettingsPanel />
      </SectionSpan>
    </SectionLayout>
  );
}
