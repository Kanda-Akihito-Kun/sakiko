import { useTranslation } from "react-i18next";
import { ResultsArchivePanel } from "../../../../../components/results/ResultsArchivePanel";
import type { DownloadTarget, ResultArchive, ResultArchiveListItem } from "../../../../../types/sakiko";

type RemoteResultArchivesPanelProps = {
  archives: ResultArchiveListItem[];
  archiveDetails: Record<string, ResultArchive | undefined>;
  archiveLoading: Record<string, boolean | undefined>;
  downloadTargets: DownloadTarget[];
  loading: boolean;
  role: "standalone" | "master" | "knight";
  visibleCount: number;
  onDeleteArchive: (taskId: string) => Promise<void>;
  onEnsureArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  onLoadMore: () => void;
  onRefresh: () => Promise<void>;
};

export function RemoteResultArchivesPanel(props: RemoteResultArchivesPanelProps) {
  const { t } = useTranslation();
  const title = props.role === "knight" ? t("dashboard.remote.knightResultArchivesTitle") : t("dashboard.remote.masterResultArchivesTitle");
  const emptyTitle = props.role === "knight" ? t("dashboard.remote.noKnightResultArchives") : t("dashboard.remote.noMasterResultArchives");
  const emptyDescription = props.role === "knight"
    ? t("dashboard.remote.noKnightResultArchivesDetail")
    : t("dashboard.remote.noMasterResultArchivesDetail");

  return (
    <ResultsArchivePanel
      archiveDetails={props.archiveDetails}
      archiveLoading={props.archiveLoading}
      archives={props.archives}
      downloadTargets={props.downloadTargets}
      loading={props.loading}
      visibleCount={props.visibleCount}
      onDeleteArchive={props.onDeleteArchive}
      onEnsureArchive={props.onEnsureArchive}
      onLoadMore={props.onLoadMore}
      onRefresh={props.onRefresh}
      title={title}
      emptyTitle={emptyTitle}
      emptyDescription={emptyDescription}
      loadingArchiveLabel={t("dashboard.remote.loadingArchive")}
    />
  );
}
