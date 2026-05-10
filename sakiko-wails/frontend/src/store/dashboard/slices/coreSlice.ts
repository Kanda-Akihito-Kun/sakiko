import { initialImportForm } from "../../../constants/dashboard";
import { SakikoService } from "../../../services/sakikoService";
import { translate } from "../../../services/i18n";
import { normalizeError, toggleTaskPresetSelection } from "../../../utils/dashboard";
import { createInitialTaskConfig, normalizeTaskConfigPatch, resolveActiveProfileId } from "../helpers";
import type { DashboardCoreSlice, DashboardSliceCreator } from "../types";

export const createDashboardCoreSlice: DashboardSliceCreator<DashboardCoreSlice> = (set, get) => ({
  profilesPath: "",
  mihomoVersion: "",
  networkEnv: null,
  downloadTargets: [],
  downloadTargetsLoading: false,
  downloadTargetSearch: "",
  importForm: initialImportForm,
  taskPreset: ["full", "ping", "geo", "udp", "speed", "media"],
  taskConfig: createInitialTaskConfig(),
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
    const nextPatch = normalizeTaskConfigPatch(patch);
    set((state) => ({
      taskConfig: {
        ...state.taskConfig,
        ...nextPatch,
      },
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

      const current = get();
      const status = statusResult.status === "fulfilled" ? statusResult.value : null;
      const nextTasks = tasksResult.status === "fulfilled" && Array.isArray(tasksResult.value)
        ? tasksResult.value
        : current.tasks;
      const nextProfiles = profilesResult.status === "fulfilled" && Array.isArray(profilesResult.value)
        ? profilesResult.value
        : current.profiles;
      const loadErrors = [
        statusResult.status === "rejected" ? `DesktopStatus: ${normalizeError(statusResult.reason)}` : "",
        tasksResult.status === "rejected" ? `ListTasks: ${normalizeError(tasksResult.reason)}` : "",
        profilesResult.status === "rejected" ? `ListProfileSummaries: ${normalizeError(profilesResult.reason)}` : "",
      ].filter(Boolean);
      const targetId = resolveActiveProfileId(nextProfiles, preferredProfileId || current.activeProfileId);
      const fallbackProfile = current.activeProfile?.id === targetId ? current.activeProfile : null;

      set({
        profilesPath: status?.profilesPath || current.profilesPath,
        mihomoVersion: status?.mihomoVersion || current.mihomoVersion,
        networkEnv: status?.networkEnv || current.networkEnv,
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
});
