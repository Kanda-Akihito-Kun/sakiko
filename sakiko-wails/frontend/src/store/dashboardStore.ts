import { create } from "zustand";
import { SakikoService } from "../services/sakikoService";
import {
  ProfileImportRequest,
  TaskConfig,
} from "../../bindings/sakiko.local/sakiko-core/interfaces";
import { initialImportForm, initialTaskConfig } from "../constants/dashboard";
import { createImportProfilePayload, createSubmitProfileTaskPayload } from "./dashboardPayloads";
import type { ImportForm, TaskPreset } from "../types/dashboard";
import type { DownloadTarget, Profile, ProfileSummary, ResultArchive, ResultArchiveListItem, TaskState, TaskStatusResponse } from "../types/sakiko";
import { normalizeError } from "../utils/dashboard";

type TaskConfigPatch = Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading">>;

type DashboardState = {
  profiles: ProfileSummary[];
  activeProfileId: string;
  activeProfile: Profile | null;
  tasks: TaskState[];
  activeTaskId: string;
  activeTask: TaskStatusResponse | null;
  resultArchives: ResultArchiveListItem[];
  resultArchivesLoading: boolean;
  resultArchiveDetails: Record<string, ResultArchive | undefined>;
  resultArchiveLoading: Record<string, boolean | undefined>;
  resultArchivesVisibleCount: number;
  profilesPath: string;
  downloadTargets: DownloadTarget[];
  downloadTargetsLoading: boolean;
  importForm: ImportForm;
  taskPreset: TaskPreset;
  taskConfig: TaskConfig;
  nodeFilter: string;
  loading: boolean;
  submitting: boolean;
  message: string;
  error: string;
  setNodeFilter: (value: string) => void;
  setTaskPreset: (value: TaskPreset) => void;
  updateImportForm: (field: keyof ImportForm, value: string) => void;
  patchTaskConfig: (patch: TaskConfigPatch) => void;
  refreshDashboard: (preferredProfileId?: string) => Promise<void>;
  refreshDownloadTargets: () => Promise<void>;
  refreshResultArchives: (resetVisibleCount?: boolean) => Promise<void>;
  loadMoreResultArchives: () => void;
  ensureResultArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  handleDeleteResultArchive: (taskId: string) => Promise<void>;
  syncActiveTask: () => Promise<void>;
  handleProfileSelect: (profileId: string) => Promise<void>;
  handleImport: () => Promise<void>;
  handleRefreshProfile: () => Promise<void>;
  handleDeleteProfile: () => Promise<void>;
  handleRunTask: () => Promise<void>;
  handleInspectTask: (taskId: string) => Promise<void>;
  clearError: () => void;
};

export const useDashboardStore = create<DashboardState>((set, get) => ({
  profiles: [],
  activeProfileId: "",
  activeProfile: null,
  tasks: [],
  activeTaskId: "",
  activeTask: null,
  resultArchives: [],
  resultArchivesLoading: false,
  resultArchiveDetails: {},
  resultArchiveLoading: {},
  resultArchivesVisibleCount: 10,
  profilesPath: "",
  downloadTargets: [],
  downloadTargetsLoading: false,
  importForm: initialImportForm,
  taskPreset: "ping",
  taskConfig: initialTaskConfig,
  nodeFilter: "",
  loading: true,
  submitting: false,
  message: "Ready.",
  error: "",

  setNodeFilter: (value) => set({ nodeFilter: value }),
  setTaskPreset: (value) => set({ taskPreset: value }),
  updateImportForm: (field, value) => {
    set((state) => ({
      importForm: {
        ...state.importForm,
        [field]: value,
      },
    }));
  },
  patchTaskConfig: (patch) => {
    const nextPatch = { ...patch };
    if (typeof nextPatch.taskTimeoutMillis === "number") {
      if (!Number.isFinite(nextPatch.taskTimeoutMillis)) {
        delete nextPatch.taskTimeoutMillis;
      } else {
        nextPatch.taskTimeoutMillis = Math.max(1, Math.floor(nextPatch.taskTimeoutMillis));
      }
    }
    if (typeof nextPatch.downloadDuration === "number" && Number.isFinite(nextPatch.downloadDuration)) {
      nextPatch.downloadDuration = Math.min(20, Math.max(5, nextPatch.downloadDuration));
    } else if (typeof nextPatch.downloadDuration === "number") {
      delete nextPatch.downloadDuration;
    }
    if (typeof nextPatch.downloadThreading === "number") {
      if (!Number.isFinite(nextPatch.downloadThreading)) {
        delete nextPatch.downloadThreading;
      } else {
        nextPatch.downloadThreading = Math.max(1, Math.floor(nextPatch.downloadThreading));
      }
    }

    set((state) => ({
      taskConfig: new TaskConfig({
        ...state.taskConfig,
        ...nextPatch,
      }),
    }));
  },
  clearError: () => set({ error: "" }),

  refreshDashboard: async (preferredProfileId) => {
    set({ loading: true, error: "" });

    try {
      const [statusResult, tasksResult, profilesResult] = await Promise.allSettled([
        SakikoService.DesktopStatus(),
        SakikoService.ListTasks(),
        SakikoService.ListProfileSummaries(),
      ]);

      const currentActiveProfileId = get().activeProfileId;
      const currentActiveProfile = get().activeProfile;
      const currentProfilesPath = get().profilesPath;
      const currentTasks = get().tasks;
      const currentProfiles = get().profiles;
      const status = statusResult.status === "fulfilled" ? statusResult.value : null;
      const nextTasks = tasksResult.status === "fulfilled" && Array.isArray(tasksResult.value)
        ? tasksResult.value
        : currentTasks;
      const nextProfiles = profilesResult.status === "fulfilled" && Array.isArray(profilesResult.value)
        ? profilesResult.value
        : currentProfiles;
      const loadErrors = [
        statusResult.status === "rejected" ? `DesktopStatus: ${normalizeError(statusResult.reason)}` : "",
        tasksResult.status === "rejected" ? `ListTasks: ${normalizeError(tasksResult.reason)}` : "",
        profilesResult.status === "rejected" ? `ListProfileSummaries: ${normalizeError(profilesResult.reason)}` : "",
      ].filter(Boolean);
      const targetId = resolveActiveProfileId(nextProfiles, preferredProfileId || currentActiveProfileId);
      const fallbackProfile = currentActiveProfile?.id === targetId ? currentActiveProfile : null;

      set({
        profilesPath: status?.profilesPath || currentProfilesPath,
        tasks: nextTasks,
        profiles: nextProfiles,
        activeProfileId: targetId,
        activeProfile: fallbackProfile,
        error: loadErrors.join(" | "),
      });

      if (!targetId) {
        return;
      }

      try {
        const activeProfile = await SakikoService.GetProfile(targetId);
        if (get().activeProfileId !== targetId) {
          return;
        }

        set({ activeProfile });
      } catch (err) {
        set((state) => ({
          error: normalizeError(err),
          activeProfile: state.activeProfile?.id === targetId ? state.activeProfile : fallbackProfile,
        }));
      }
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ loading: false });
    }
  },

  refreshDownloadTargets: async () => {
    set({ downloadTargetsLoading: true });

    try {
      const downloadTargets = await SakikoService.ListDownloadTargets();
      set({
        downloadTargets,
        message: `Loaded ${Math.max(downloadTargets.length - 1, 0)} Speedtest target(s).`,
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ downloadTargetsLoading: false });
    }
  },

  refreshResultArchives: async (resetVisibleCount = true) => {
    set((state) => ({
      error: "",
      resultArchivesLoading: true,
      resultArchivesVisibleCount: resetVisibleCount ? 10 : state.resultArchivesVisibleCount,
    }));

    try {
      const resultArchives = await SakikoService.ListResultArchives();
      set((state) => ({
        resultArchives,
        resultArchiveDetails: Object.fromEntries(
          Object.entries(state.resultArchiveDetails).filter(([taskId]) => resultArchives.some((item) => item.taskId === taskId)),
        ),
        message: resultArchives.length > 0
          ? `Loaded ${resultArchives.length} archived result(s).`
          : state.message,
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ resultArchivesLoading: false });
    }
  },

  loadMoreResultArchives: () => {
    set((state) => ({
      resultArchivesVisibleCount: Math.min(state.resultArchives.length, state.resultArchivesVisibleCount + 10),
    }));
  },

  ensureResultArchive: async (taskId) => {
    const { resultArchiveDetails, resultArchiveLoading } = get();
    if (resultArchiveDetails[taskId] || resultArchiveLoading[taskId]) {
      return resultArchiveDetails[taskId];
    }

    set((state) => ({
      resultArchiveLoading: {
        ...state.resultArchiveLoading,
        [taskId]: true,
      },
    }));

    try {
      const archive = await SakikoService.GetResultArchive(taskId);
      set((state) => ({
        resultArchiveDetails: {
          ...state.resultArchiveDetails,
          [taskId]: archive,
        },
        resultArchiveLoading: {
          ...state.resultArchiveLoading,
          [taskId]: false,
        },
      }));
      return archive;
    } catch (err) {
      set((state) => ({
        error: normalizeError(err),
        resultArchiveLoading: {
          ...state.resultArchiveLoading,
          [taskId]: false,
        },
      }));
      return undefined;
    }
  },

  handleDeleteResultArchive: async (taskId) => {
    const archive = get().resultArchives.find((item) => item.taskId === taskId);
    const archiveName = archive?.taskName || taskId;

    set({ error: "", message: `Deleting archived result ${archiveName}...` });

    try {
      await SakikoService.DeleteResultArchive(taskId);
      set((state) => ({
        resultArchives: state.resultArchives.filter((item) => item.taskId !== taskId),
        resultArchiveDetails: Object.fromEntries(
          Object.entries(state.resultArchiveDetails).filter(([currentTaskId]) => currentTaskId !== taskId),
        ),
        resultArchiveLoading: Object.fromEntries(
          Object.entries(state.resultArchiveLoading).filter(([currentTaskId]) => currentTaskId !== taskId),
        ),
        message: `Deleted archived result ${archiveName}.`,
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  syncActiveTask: async () => {
    const activeTaskId = get().activeTaskId;
    if (!activeTaskId) {
      return;
    }

    try {
      const detail = await SakikoService.GetTask(activeTaskId);
      if (get().activeTaskId !== activeTaskId) {
        return;
      }

      set({ activeTask: detail });

      const latestTasks = await SakikoService.ListTasks();
      if (get().activeTaskId !== activeTaskId) {
        return;
      }

      const selected = latestTasks.find((item) => item.taskId === activeTaskId);
      set({
        tasks: latestTasks,
        message: selected && selected.status !== "running"
          ? `Task ${selected.name} finished with ${detail.results?.length ?? 0} result(s).`
          : get().message,
      });

      if (selected && selected.status !== "running" && !get().resultArchives.some((item) => item.taskId === activeTaskId)) {
        void get().refreshResultArchives(false);
      }
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  handleProfileSelect: async (profileId) => {
    set({
      activeProfileId: profileId,
      activeProfile: get().activeProfile?.id === profileId ? get().activeProfile : null,
      error: "",
    });

    try {
      const profile = await SakikoService.GetProfile(profileId);
      set({ activeProfile: profile });
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  handleImport: async () => {
    const { importForm, refreshDashboard } = get();

    set({
      submitting: true,
      error: "",
      message: "Importing profile...",
    });

    try {
      const request: ProfileImportRequest = createImportProfilePayload(importForm);
      const profile = await SakikoService.ImportProfile(request);

      set((state) => ({
        importForm: initialImportForm,
        activeProfileId: profile.id,
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
      }));
      await refreshDashboard(profile.id);
      set({ message: `Imported ${profile.name} (${profile.nodes.length} nodes).` });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleRefreshProfile: async () => {
    const { activeProfile, activeProfileId, refreshDashboard } = get();
    if (!activeProfileId || !activeProfile?.source?.trim()) {
      return;
    }

    set({
      submitting: true,
      error: "",
      message: "Refreshing profile...",
    });

    try {
      const profile = await SakikoService.RefreshProfile(activeProfileId);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
      }));
      await refreshDashboard(profile.id);
      set({ message: `Refreshed ${profile.name}.` });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleDeleteProfile: async () => {
    const { activeProfile, activeProfileId, profiles, refreshDashboard } = get();
    if (!activeProfileId) {
      return;
    }

    set({
      submitting: true,
      error: "",
      message: `Deleting ${activeProfile?.name || "profile"}...`,
    });

    try {
      await SakikoService.DeleteProfile(activeProfileId);
      const nextProfiles = removeProfileSummary(profiles, activeProfileId);
      const nextActiveProfileId = nextProfiles[0]?.id || "";

      set({
        profiles: nextProfiles,
        activeProfileId: nextActiveProfileId,
        activeProfile: null,
        nodeFilter: "",
      });

      await refreshDashboard(nextActiveProfileId);
      set({ message: `Deleted ${activeProfile?.name || "profile"}.` });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleRunTask: async () => {
    const {
      activeProfileId,
      taskConfig,
      taskPreset,
    } = get();

    if (!activeProfileId) {
      set({ error: "Select a profile first." });
      return;
    }

    set({
      submitting: true,
      error: "",
      message: `Starting ${taskPreset} task...`,
    });

    try {
      const taskId = await SakikoService.SubmitProfileTask(
        createSubmitProfileTaskPayload(activeProfileId, taskPreset, taskConfig),
      );

      const [detail, latestTasks] = await Promise.all([
        SakikoService.GetTask(taskId),
        SakikoService.ListTasks(),
      ]);

      set({
        activeTaskId: taskId,
        activeTask: detail,
        tasks: latestTasks,
        message: `Task ${taskId} accepted.`,
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleInspectTask: async (taskId) => {
    set({ activeTaskId: taskId, error: "" });

    try {
      const detail = await SakikoService.GetTask(taskId);
      if (get().activeTaskId !== taskId) {
        return;
      }

      set({ activeTask: detail });
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },
}));

function upsertProfileSummary(profiles: ProfileSummary[], nextProfile: Profile): ProfileSummary[] {
  const nextSummary = toProfileSummary(nextProfile);
  const nextProfiles = profiles.slice();
  const targetIndex = nextProfiles.findIndex((profile) => profile.id === nextSummary.id);
  if (targetIndex >= 0) {
    nextProfiles[targetIndex] = nextSummary;
    return nextProfiles;
  }
  return [nextSummary, ...nextProfiles];
}

function removeProfileSummary(profiles: ProfileSummary[], profileID: string): ProfileSummary[] {
  return profiles.filter((profile) => profile.id !== profileID);
}

function resolveActiveProfileId(profiles: ProfileSummary[], preferredProfileId?: string): string {
  if (preferredProfileId && profiles.some((profile) => profile.id === preferredProfileId)) {
    return preferredProfileId;
  }
  return profiles[0]?.id || "";
}

function toProfileSummary(profile: Profile): ProfileSummary {
  return {
    id: profile.id,
    name: profile.name,
    source: profile.source,
    updatedAt: profile.updatedAt,
    nodeCount: profile.nodes.length,
  };
}
