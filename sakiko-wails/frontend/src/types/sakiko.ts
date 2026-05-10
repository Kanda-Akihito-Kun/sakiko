import type { TaskState as BoundTaskState } from "../../bindings/sakiko.local/sakiko-core/interfaces/index.js";

export type ProfileSummary = {
  id: string;
  name: string;
  source: string;
  updatedAt?: string;
  nodeCount: number;
  remainingBytes?: number;
  expiresAt?: string;
};

export type ClusterRole = "standalone" | "master" | "knight";

export type MasterEligibility = {
  publicIP?: string;
  hasPublicIP: boolean;
  natType?: string;
  isNat1: boolean;
  reachable: boolean;
  eligible: boolean;
  checkedAt?: string;
  error?: string;
};

export type ClusterMasterStatus = {
  enabled: boolean;
  listenHost?: string;
  listenPort?: number;
  eligibility: MasterEligibility;
};

export type ClusterKnightStatus = {
  bound: boolean;
  knightId?: string;
  knightName?: string;
  masterHost?: string;
  masterPort?: number;
  connected?: boolean;
  lastSeenAt?: string;
  lastError?: string;
};

export type ClusterStatus = {
  role: ClusterRole;
  master?: ClusterMasterStatus;
  knight?: ClusterKnightStatus;
};

export type ClusterPairingCode = {
  code: string;
  knightName?: string;
  expiresAt?: string;
};

export type ClusterKnightState = "paired" | "online" | "busy" | "offline" | "revoked";

export type ClusterConnectedKnight = {
  knightId: string;
  knightName?: string;
  state: ClusterKnightState;
  remoteAddr?: string;
  lastSeenAt?: string;
  lastError?: string;
};

export type ClusterRemoteTaskState = "queued" | "running" | "finished" | "failed";

export type ClusterRemoteTask = {
  assignmentId: string;
  remoteTaskId: string;
  knightId: string;
  knightName?: string;
  taskName?: string;
  state: ClusterRemoteTaskState;
  createdAt?: string;
  startedAt?: string;
  finishedAt?: string;
  exitCode?: string;
  error?: string;
  archiveTaskId?: string;
  localTaskId?: string;
  runtime?: BoundTaskState;
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
