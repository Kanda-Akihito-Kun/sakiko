import { create } from "zustand";
import { SakikoService } from "../services/sakikoService";
import {
  ProfileImportRequest,
  TaskConfig,
} from "../../bindings/sakiko.local/sakiko-core/interfaces";
import { initialImportForm, initialTaskConfig } from "../constants/dashboard";
import { translate } from "../services/i18n";
import { createImportProfilePayload, createSubmitProfileTaskPayload } from "./dashboardPayloads";
import type { ImportForm, TaskPreset, TaskPresetSelection } from "../types/dashboard";
import type { DownloadTarget, Profile, ProfileSummary, ResultArchive, ResultArchiveListItem, TaskState, TaskStatusResponse } from "../types/sakiko";
import { formatTaskPresetSelectionLabel, normalizeError, toggleTaskPresetSelection } from "../utils/dashboard";

type TaskConfigPatch = Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading" | "backendIdentity">>;

const BACKEND_IDENTITY_STORAGE_KEY = "sakiko.task-defaults.backend-identity";
let syncActiveTaskTaskId = "";
let syncActiveTaskPromise: Promise<void> | null = null;

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
  refreshResultArchives: (resetVisibleCount?: boolean) => Promise<void>;
  loadMoreResultArchives: () => void;
  ensureResultArchive: (taskId: string) => Promise<ResultArchive | undefined>;
  handleDeleteResultArchive: (taskId: string) => Promise<void>;
  syncActiveTask: () => Promise<void>;
  handleProfileSelect: (profileId: string) => Promise<void>;
  handleImport: () => Promise<void>;
  handleRefreshProfile: () => Promise<void>;
  handleDeleteProfile: () => Promise<void>;
  handleSetProfileNodeEnabled: (nodeIndex: number, enabled: boolean) => Promise<void>;
  handleMoveProfileNode: (nodeIndex: number, targetIndex: number) => Promise<void>;
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
  downloadTargetSearch: "",
  importForm: initialImportForm,
  taskPreset: ["full", "ping", "geo", "speed", "media"],
  taskConfig: new TaskConfig({
    ...initialTaskConfig,
    backendIdentity: getStoredBackendIdentity(),
  }),
  nodeFilter: "",
  loading: true,
  submitting: false,
  message: translate("shared.states.readyWithPeriod", "Ready."),
  error: "",

  setNodeFilter: (value) => set({ nodeFilter: value }),
  setDownloadTargetSearch: (value) => set({ downloadTargetSearch: value }),
  setTaskPreset: (value) => set((state) => ({ taskPreset: toggleTaskPresetSelection(state.taskPreset, value) })),
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
    if (typeof nextPatch.backendIdentity === "string") {
      nextPatch.backendIdentity = Array.from(nextPatch.backendIdentity.trim()).slice(0, 20).join("");
      persistBackendIdentity(nextPatch.backendIdentity);
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

  refreshDownloadTargets: async (search = get().downloadTargetSearch) => {
    const nextSearch = search.trim();
    set({ downloadTargetsLoading: true, downloadTargetSearch: search });

    try {
      const downloadTargets = nextSearch
        ? await SakikoService.SearchDownloadTargets(nextSearch)
        : await SakikoService.ListDownloadTargets();
      set({
        downloadTargets,
        message: nextSearch
          ? translate("dashboard.messages.downloadTargetsLoadedForSearch", `Loaded ${Math.max(downloadTargets.length - 1, 0)} Speedtest target(s) for "${nextSearch}".`, {
            count: Math.max(downloadTargets.length - 1, 0),
            search: nextSearch,
          })
          : translate("dashboard.messages.downloadTargetsLoaded", `Loaded ${Math.max(downloadTargets.length - 1, 0)} Speedtest target(s).`, {
            count: Math.max(downloadTargets.length - 1, 0),
          }),
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
          ? translate("dashboard.messages.resultArchivesLoaded", `Loaded ${resultArchives.length} archived result(s).`, {
            count: resultArchives.length,
          })
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

    set({
      error: "",
      message: translate("dashboard.messages.deletingArchive", `Deleting archived result ${archiveName}...`, {
        name: archiveName,
      }),
    });

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
        message: translate("dashboard.messages.deletedArchive", `Deleted archived result ${archiveName}.`, {
          name: archiveName,
        }),
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

    if (syncActiveTaskPromise && syncActiveTaskTaskId === activeTaskId) {
      return syncActiveTaskPromise;
    }

    syncActiveTaskTaskId = activeTaskId;
    const currentPromise = (async () => {
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
            ? translate("dashboard.messages.taskFinished", `Task ${selected.name} finished with ${detail.results?.length ?? 0} result(s).`, {
              name: selected.name,
              count: detail.results?.length ?? 0,
            })
            : get().message,
        });

        if (selected && selected.status !== "running" && !get().resultArchives.some((item) => item.taskId === activeTaskId)) {
          void get().refreshResultArchives(false);
        }
      } catch (err) {
        set({ error: normalizeError(err) });
      }
    })();

    syncActiveTaskPromise = currentPromise;
    try {
      await currentPromise;
    } finally {
      if (syncActiveTaskPromise === currentPromise) {
        syncActiveTaskPromise = null;
        syncActiveTaskTaskId = "";
      }
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
      message: translate("dashboard.messages.importingProfile", "Importing profile..."),
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
      set({
        message: translate("dashboard.messages.importedProfile", `Imported ${profile.name} (${profile.nodes.length} nodes).`, {
          name: profile.name,
          count: profile.nodes.length,
        }),
      });
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
      message: translate("dashboard.messages.refreshingProfile", "Refreshing profile..."),
    });

    try {
      const profile = await SakikoService.RefreshProfile(activeProfileId);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
      }));
      await refreshDashboard(profile.id);
      set({ message: translate("dashboard.messages.refreshedProfile", `Refreshed ${profile.name}.`, { name: profile.name }) });
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
      message: translate("dashboard.messages.deletingProfile", `Deleting ${activeProfile?.name || "profile"}...`, {
        name: activeProfile?.name || translate("shared.states.profile", "profile"),
      }),
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
      set({
        message: translate("dashboard.messages.deletedProfile", `Deleted ${activeProfile?.name || "profile"}.`, {
          name: activeProfile?.name || translate("shared.states.profile", "profile"),
        }),
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleSetProfileNodeEnabled: async (nodeIndex, enabled) => {
    const { activeProfile, activeProfileId } = get();
    if (!activeProfileId || !activeProfile || nodeIndex < 0 || nodeIndex >= activeProfile.nodes.length) {
      return;
    }

    const nodeName = activeProfile.nodes[nodeIndex]?.name || translate("shared.formats.nodeNumberLower", `node ${nodeIndex + 1}`, { index: nodeIndex + 1 });
    set({
      submitting: true,
      error: "",
      message: enabled
        ? translate("dashboard.messages.includingNode", `Including ${nodeName}...`, { name: nodeName })
        : translate("dashboard.messages.skippingNode", `Skipping ${nodeName}...`, { name: nodeName }),
    });

    try {
      const profile = await SakikoService.SetProfileNodeEnabled(activeProfileId, nodeIndex, enabled);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
        message: enabled
          ? translate("dashboard.messages.includedNode", `Included ${nodeName} for future tasks.`, { name: nodeName })
          : translate("dashboard.messages.skippedNode", `Skipped ${nodeName} for future tasks.`, { name: nodeName }),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleMoveProfileNode: async (nodeIndex, targetIndex) => {
    const { activeProfile, activeProfileId } = get();
    if (!activeProfileId || !activeProfile || nodeIndex < 0 || targetIndex < 0 || nodeIndex >= activeProfile.nodes.length || targetIndex >= activeProfile.nodes.length) {
      return;
    }

    const nodeName = activeProfile.nodes[nodeIndex]?.name || translate("shared.formats.nodeNumberLower", `node ${nodeIndex + 1}`, { index: nodeIndex + 1 });
    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.reorderingNode", `Reordering ${nodeName}...`, { name: nodeName }),
    });

    try {
      const profile = await SakikoService.MoveProfileNode(activeProfileId, nodeIndex, targetIndex);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
        message: translate("dashboard.messages.movedNode", `Moved ${nodeName}.`, { name: nodeName }),
      }));
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

    const selectedPresetCount = taskPreset.filter((preset) => preset !== "full").length;

    if (!activeProfileId) {
      set({ error: translate("dashboard.messages.selectProfileFirst", "Select a profile first.") });
      return;
    }
    if (selectedPresetCount === 0) {
      set({ error: translate("dashboard.messages.selectTestGroup", "Select at least one test group.") });
      return;
    }

    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.startingTask", `Starting ${formatTaskPresetSelectionLabel(taskPreset)} task...`, {
        preset: formatTaskPresetSelectionLabel(taskPreset),
      }),
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
        message: translate("dashboard.messages.taskAccepted", `Task ${taskId} accepted.`, { taskId }),
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

function getStoredBackendIdentity(): string {
  if (typeof window === "undefined") {
    return "";
  }

  const raw = window.localStorage.getItem(BACKEND_IDENTITY_STORAGE_KEY) || "";
  return Array.from(raw.trim()).slice(0, 20).join("");
}

function persistBackendIdentity(value: string) {
  if (typeof window === "undefined") {
    return;
  }

  const normalized = Array.from(value.trim()).slice(0, 20).join("");
  if (!normalized) {
    window.localStorage.removeItem(BACKEND_IDENTITY_STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(BACKEND_IDENTITY_STORAGE_KEY, normalized);
}

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
