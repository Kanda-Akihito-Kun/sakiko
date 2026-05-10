import type { StateCreator } from "zustand";
import type { ImportForm, TaskPreset, TaskPresetSelection } from "../../types/dashboard";
import type {
  BackendInfo,
  ClusterConnectedKnight,
  ClusterPairingCode,
  ClusterRemoteTask,
  ClusterStatus,
  DownloadTarget,
  MasterEligibility,
  Profile,
  ProfileSummary,
  ResultArchive,
  ResultArchiveListItem,
  TaskConfig,
  TaskState,
  TaskStatusResponse,
} from "../../types/sakiko";

export type TaskConfigPatch = Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading" | "backendIdentity">>;

export type DashboardCoreSlice = {
  profilesPath: string;
  mihomoVersion: string;
  networkEnv: BackendInfo | null;
  downloadTargets: DownloadTarget[];
  downloadTargetsLoading: boolean;
  downloadTargetSearch: string;
  importForm: ImportForm;
  taskPreset: TaskPresetSelection;
  taskConfig: TaskConfig;
  nodeFilter: string;
  loading: boolean;
  submitting: boolean;
  message: string;
  error: string;
  setNodeFilter: (value: string) => void;
  setDownloadTargetSearch: (value: string) => void;
  setTaskPreset: (value: TaskPreset) => void;
  updateImportForm: (field: keyof ImportForm, value: string) => void;
  patchTaskConfig: (patch: TaskConfigPatch) => void;
  refreshDashboard: (preferredProfileId?: string) => Promise<void>;
  refreshDownloadTargets: (search?: string) => Promise<void>;
  clearError: () => void;
};

export type DashboardProfilesSlice = {
  profiles: ProfileSummary[];
  activeProfileId: string;
  activeProfile: Profile | null;
  handleProfileSelect: (profileId: string) => Promise<void>;
  handleImport: () => Promise<void>;
  handleRefreshProfile: () => Promise<void>;
  handleDeleteProfile: () => Promise<void>;
  handleSetProfileNodeEnabled: (nodeIndex: number, enabled: boolean) => Promise<void>;
  handleMoveProfileNode: (nodeIndex: number, targetIndex: number) => Promise<void>;
};

export type DashboardTasksSlice = {
  tasks: TaskState[];
  activeTaskId: string;
  activeTask: TaskStatusResponse | null;
  refreshTasks: () => Promise<void>;
  handleRunTask: () => Promise<void>;
  handleStopTask: () => Promise<void>;
  syncActiveTask: () => Promise<void>;
  handleDeleteTask: (taskId: string) => Promise<void>;
  handleInspectTask: (taskId: string) => Promise<void>;
};

export type DashboardResultsSlice = {
  resultArchives: ResultArchiveListItem[];
  resultArchivesLoading: boolean;
  resultArchiveDetails: Record<string, ResultArchive | undefined>;
  resultArchiveLoading: Record<string, boolean | undefined>;
  resultArchivesVisibleCount: number;
  refreshResultArchives: (resetVisibleCount?: boolean) => Promise<void>;
  loadMoreResultArchives: () => void;
  ensureResultArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  handleDeleteResultArchive: (taskId: string) => Promise<void>;
};

export type DashboardRemoteSlice = {
  remoteStatus: ClusterStatus | null;
  remoteEligibility: MasterEligibility | null;
  remotePairingCode: ClusterPairingCode | null;
  remoteKnights: ClusterConnectedKnight[];
  remoteTasks: ClusterRemoteTask[];
  selectedRemoteKnightIDs: string[];
  remoteResultArchives: ResultArchiveListItem[];
  remoteResultArchivesLoading: boolean;
  remoteResultArchiveDetails: Record<string, ResultArchive | undefined>;
  remoteResultArchiveLoading: Record<string, boolean | undefined>;
  remoteResultArchivesVisibleCount: number;
  remoteLoading: boolean;
  remoteSubmitting: boolean;
  refreshRemoteStatus: () => Promise<void>;
  refreshRemoteKnights: () => Promise<void>;
  refreshRemoteWorkspace: (includeArchives?: boolean) => Promise<void>;
  refreshRemoteTasks: () => Promise<void>;
  refreshRemoteResultArchives: (resetVisibleCount?: boolean) => Promise<void>;
  loadMoreRemoteResultArchives: () => void;
  ensureRemoteResultArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  handleDeleteRemoteResultArchive: (taskId: string) => Promise<void>;
  toggleRemoteKnightSelection: (knightId: string) => void;
  handleKickRemoteKnight: (knightId: string) => Promise<void>;
  handleProbeRemoteMasterEligibility: () => Promise<void>;
  handleEnableRemoteMaster: (listenHost: string, listenPort: number) => Promise<void>;
  handleCreateRemotePairingCode: (knightName: string, ttlSeconds: number) => Promise<void>;
  handleEnableRemoteKnight: (masterHost: string, masterPort: number, oneTimeCode: string) => Promise<void>;
  handleDisableRemoteMode: () => Promise<void>;
  handleSubmitRemoteTask: () => Promise<void>;
};

export type DashboardState =
  & DashboardCoreSlice
  & DashboardProfilesSlice
  & DashboardTasksSlice
  & DashboardResultsSlice
  & DashboardRemoteSlice;

export type DashboardSliceCreator<T> = StateCreator<DashboardState, [], [], T>;
