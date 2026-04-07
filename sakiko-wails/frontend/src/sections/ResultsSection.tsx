import { ResultsArchivePanel } from "../components/results/ResultsArchivePanel";
import type { DownloadTarget, ResultArchive, ResultArchiveListItem } from "../types/sakiko";

type ResultsSectionProps = {
  archiveDetails: Record<string, ResultArchive | undefined>;
  archiveLoading: Record<string, boolean | undefined>;
  archives: ResultArchiveListItem[];
  downloadTargets: DownloadTarget[];
  loading: boolean;
  visibleCount: number;
  onDeleteArchive: (taskId: string) => Promise<void>;
  onEnsureArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  onLoadMore: () => void;
  onRefresh: () => Promise<void>;
};

export function ResultsSection(props: ResultsSectionProps) {
  return <ResultsArchivePanel {...props} />;
}
