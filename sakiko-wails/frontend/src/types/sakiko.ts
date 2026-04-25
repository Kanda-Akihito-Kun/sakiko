export type ProfileSummary = {
  id: string;
  name: string;
  source: string;
  updatedAt?: string;
  nodeCount: number;
  remainingBytes?: number;
  expiresAt?: string;
};

export type {
  BackendInfo,
  DownloadTarget,
  EntryResult,
  TaskActiveNode,
  MatrixResult,
  Profile,
  ResultArchive,
  ResultArchiveListItem,
  ResultArchiveTask,
  ResultReport,
  ResultReportColumn,
  ResultReportSection,
  TaskConfig,
  TaskContext,
  TaskState,
  TaskStatusResponse,
} from "../../bindings/sakiko.local/sakiko-core/interfaces/index.js";

export {
  DownloadTargetSource,
} from "../../bindings/sakiko.local/sakiko-core/interfaces/index.js";
