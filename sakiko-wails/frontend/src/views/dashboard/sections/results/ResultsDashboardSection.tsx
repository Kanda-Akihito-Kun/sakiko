import { useShallow } from "zustand/react/shallow";
import { ResultsArchivePanel } from "../../../../components/results/ResultsArchivePanel";
import { useDashboardStore } from "../../../../store/dashboardStore";

export function ResultsDashboardSection() {
  const view = useDashboardStore(useShallow((state) => ({
    resultArchiveDetails: state.resultArchiveDetails,
    resultArchiveLoading: state.resultArchiveLoading,
    resultArchives: state.resultArchives,
    downloadTargets: state.downloadTargets,
    resultArchivesLoading: state.resultArchivesLoading,
    resultArchivesVisibleCount: state.resultArchivesVisibleCount,
    handleDeleteResultArchive: state.handleDeleteResultArchive,
    ensureResultArchive: state.ensureResultArchive,
    loadMoreResultArchives: state.loadMoreResultArchives,
    refreshResultArchives: state.refreshResultArchives,
  })));

  return (
    <ResultsArchivePanel
      archiveDetails={view.resultArchiveDetails}
      archiveLoading={view.resultArchiveLoading}
      archives={view.resultArchives}
      downloadTargets={view.downloadTargets}
      loading={view.resultArchivesLoading}
      visibleCount={view.resultArchivesVisibleCount}
      onDeleteArchive={view.handleDeleteResultArchive}
      onEnsureArchive={view.ensureResultArchive}
      onLoadMore={view.loadMoreResultArchives}
      onRefresh={view.refreshResultArchives}
    />
  );
}
