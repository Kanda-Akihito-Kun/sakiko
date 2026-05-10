import { SakikoService } from "../../../services/sakikoService";
import { translate } from "../../../services/i18n";
import { normalizeError } from "../../../utils/dashboard";
import type { DashboardRemoteSlice, DashboardSliceCreator } from "../types";

const remoteArchivePageSize = 10;

export const createDashboardRemoteSlice: DashboardSliceCreator<DashboardRemoteSlice> = (set, get) => ({
  remoteStatus: null,
  remoteEligibility: null,
  remotePairingCode: null,
  remoteKnights: [],
  remoteTasks: [],
  selectedRemoteKnightIDs: [],
  remoteResultArchives: [],
  remoteResultArchivesLoading: false,
  remoteResultArchiveDetails: {},
  remoteResultArchiveLoading: {},
  remoteResultArchivesVisibleCount: remoteArchivePageSize,
  remoteLoading: false,
  remoteSubmitting: false,

  refreshRemoteStatus: async () => {
    set({ remoteLoading: true, error: "" });

    try {
      const remoteStatus = await SakikoService.GetRemoteStatus();
      set({
        remoteStatus,
        remoteEligibility: remoteStatus.master?.eligibility || get().remoteEligibility,
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteLoading: false });
    }
  },

  refreshRemoteKnights: async () => {
    set({ remoteLoading: true, error: "" });

    try {
      const remoteKnights = await SakikoService.ListRemoteKnights();
      set((state) => ({
        remoteKnights,
        selectedRemoteKnightIDs: state.selectedRemoteKnightIDs.filter((knightId) => remoteKnights.some((knight) => knight.knightId === knightId)),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteLoading: false });
    }
  },

  refreshRemoteWorkspace: async (includeArchives = false) => {
    set({ remoteLoading: true, error: "" });

    try {
      const [remoteStatus, remoteKnights, remoteTasks] = await Promise.all([
        SakikoService.GetRemoteStatus(),
        SakikoService.ListRemoteKnights(),
        SakikoService.ListRemoteTasks(),
      ]);

      set((state) => ({
        remoteStatus,
        remoteEligibility: remoteStatus.master?.eligibility || state.remoteEligibility,
        remoteKnights,
        remoteTasks,
        selectedRemoteKnightIDs: state.selectedRemoteKnightIDs.filter((knightId) => remoteKnights.some((knight) => knight.knightId === knightId)),
      }));

      if (includeArchives) {
        await get().refreshRemoteResultArchives(false);
      }
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteLoading: false });
    }
  },

  refreshRemoteTasks: async () => {
    set({ remoteLoading: true, error: "" });

    try {
      const remoteTasks = await SakikoService.ListRemoteTasks();
      set({ remoteTasks });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteLoading: false });
    }
  },

  refreshRemoteResultArchives: async (resetVisibleCount = true) => {
    set((state) => ({
      error: "",
      remoteResultArchivesLoading: true,
      remoteResultArchivesVisibleCount: resetVisibleCount ? remoteArchivePageSize : state.remoteResultArchivesVisibleCount,
    }));

    try {
      const role = get().remoteStatus?.role || "standalone";
      const remoteResultArchives = role === "knight"
        ? await SakikoService.ListRemoteKnightResultArchives()
        : await SakikoService.ListRemoteMasterResultArchives();

      set((state) => ({
        remoteResultArchives,
        remoteResultArchiveDetails: Object.fromEntries(
          Object.entries(state.remoteResultArchiveDetails).filter(([taskId]) => remoteResultArchives.some((item) => item.taskId === taskId)),
        ),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteResultArchivesLoading: false });
    }
  },

  loadMoreRemoteResultArchives: () => {
    set((state) => ({
      remoteResultArchivesVisibleCount: Math.min(state.remoteResultArchives.length, state.remoteResultArchivesVisibleCount + remoteArchivePageSize),
    }));
  },

  ensureRemoteResultArchive: async (taskId) => {
    const { remoteResultArchiveDetails, remoteResultArchiveLoading, remoteStatus } = get();
    if (remoteResultArchiveDetails[taskId] || remoteResultArchiveLoading[taskId]) {
      return remoteResultArchiveDetails[taskId];
    }

    set((state) => ({
      remoteResultArchiveLoading: {
        ...state.remoteResultArchiveLoading,
        [taskId]: true,
      },
    }));

    try {
      const archive = remoteStatus?.role === "knight"
        ? await SakikoService.GetRemoteKnightResultArchive(taskId)
        : await SakikoService.GetRemoteMasterResultArchive(taskId);
      set((state) => ({
        remoteResultArchiveDetails: {
          ...state.remoteResultArchiveDetails,
          [taskId]: archive,
        },
        remoteResultArchiveLoading: {
          ...state.remoteResultArchiveLoading,
          [taskId]: false,
        },
      }));
      return archive;
    } catch (err) {
      set((state) => ({
        error: normalizeError(err),
        remoteResultArchiveLoading: {
          ...state.remoteResultArchiveLoading,
          [taskId]: false,
        },
      }));
      return undefined;
    }
  },

  handleDeleteRemoteResultArchive: async (taskId) => {
    const archive = get().remoteResultArchives.find((item) => item.taskId === taskId);
    const archiveName = archive?.taskName || taskId;

    set({
      error: "",
      message: translate("dashboard.messages.deletingArchive", `Deleting archived result ${archiveName}...`, {
        name: archiveName,
      }),
    });

    try {
      if (get().remoteStatus?.role === "knight") {
        await SakikoService.DeleteRemoteKnightResultArchive(taskId);
      } else {
        await SakikoService.DeleteRemoteMasterResultArchive(taskId);
      }

      set((state) => ({
        remoteResultArchives: state.remoteResultArchives.filter((item) => item.taskId !== taskId),
        remoteResultArchiveDetails: Object.fromEntries(
          Object.entries(state.remoteResultArchiveDetails).filter(([currentTaskId]) => currentTaskId !== taskId),
        ),
        remoteResultArchiveLoading: Object.fromEntries(
          Object.entries(state.remoteResultArchiveLoading).filter(([currentTaskId]) => currentTaskId !== taskId),
        ),
        message: translate("dashboard.messages.deletedArchive", `Deleted archived result ${archiveName}.`, {
          name: archiveName,
        }),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  toggleRemoteKnightSelection: (knightId) => {
    const normalizedKnightID = knightId.trim();
    if (!normalizedKnightID) {
      return;
    }

    set((state) => ({
      selectedRemoteKnightIDs: state.selectedRemoteKnightIDs.includes(normalizedKnightID)
        ? state.selectedRemoteKnightIDs.filter((current) => current !== normalizedKnightID)
        : [...state.selectedRemoteKnightIDs, normalizedKnightID],
    }));
  },

  handleKickRemoteKnight: async (knightId) => {
    const normalizedKnightID = knightId.trim();
    if (!normalizedKnightID) {
      return;
    }

    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.kickingKnight", "Kicking Knight..."),
    });

    try {
      const remoteStatus = await SakikoService.KickRemoteKnight(normalizedKnightID);
      set((state) => ({
        remoteStatus,
        selectedRemoteKnightIDs: state.selectedRemoteKnightIDs.filter((current) => current !== normalizedKnightID),
        message: translate("dashboard.remote.messages.knightKicked", "Knight removed."),
      }));
      void get().refreshRemoteWorkspace(false);
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleProbeRemoteMasterEligibility: async () => {
    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.probingEligibility", "Checking Master eligibility..."),
    });

    try {
      const eligibility = await SakikoService.ProbeRemoteMasterEligibility();
      set((state) => ({
        remoteEligibility: eligibility,
        remoteStatus: state.remoteStatus
          ? {
              ...state.remoteStatus,
              master: state.remoteStatus.master
                ? {
                    ...state.remoteStatus.master,
                    eligibility,
                  }
                : {
                    enabled: false,
                    eligibility,
                  },
            }
          : state.remoteStatus,
        message: eligibility.eligible
          ? translate("dashboard.remote.messages.masterEligible", "This machine can act as a Master.")
          : translate("dashboard.remote.messages.masterIneligible", "This machine is not yet eligible to act as a Master."),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleEnableRemoteMaster: async (listenHost, listenPort) => {
    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.enablingMaster", "Enabling Master mode..."),
    });

    try {
      const remoteStatus = await SakikoService.EnableRemoteMaster(listenHost, listenPort);
      set({
        remoteStatus,
        remoteEligibility: remoteStatus.master?.eligibility || get().remoteEligibility,
        message: translate("dashboard.remote.messages.masterEnabled", "Master mode enabled."),
      });
      void get().refreshRemoteWorkspace(false);
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleCreateRemotePairingCode: async (knightName, ttlSeconds) => {
    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.generatingPairingCode", "Generating pairing code..."),
    });

    try {
      const remotePairingCode = await SakikoService.CreateRemotePairingCode(knightName, ttlSeconds);
      await get().refreshRemoteStatus();
      set({
        remotePairingCode,
        message: translate("dashboard.remote.messages.pairingCodeGenerated", "Pairing code generated."),
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleEnableRemoteKnight: async (masterHost, masterPort, oneTimeCode) => {
    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.bindingKnight", "Binding Knight to Master..."),
    });

    try {
      const remoteStatus = await SakikoService.EnableRemoteKnight(masterHost, masterPort, oneTimeCode);
      set({
        remoteStatus,
        message: translate("dashboard.remote.messages.knightEnabled", "Knight mode enabled."),
      });
      void get().refreshRemoteWorkspace(false);
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleDisableRemoteMode: async () => {
    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.disablingRemote", "Leaving remote mode..."),
    });

    try {
      const remoteStatus = await SakikoService.DisableRemoteMode();
      set({
        remoteStatus,
        remoteEligibility: remoteStatus.master?.eligibility || get().remoteEligibility,
        remotePairingCode: null,
        remoteKnights: [],
        remoteTasks: [],
        selectedRemoteKnightIDs: [],
        remoteResultArchives: [],
        remoteResultArchiveDetails: {},
        remoteResultArchiveLoading: {},
        message: translate("dashboard.remote.messages.remoteDisabled", "Remote mode disabled."),
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },

  handleSubmitRemoteTask: async () => {
    const { activeProfileId, selectedRemoteKnightIDs, taskConfig, taskPreset } = get();
    if (!activeProfileId) {
      set({ error: translate("dashboard.messages.selectProfileFirst", "Select a profile first.") });
      return;
    }
    if (selectedRemoteKnightIDs.length === 0) {
      set({ error: translate("dashboard.remote.messages.selectKnightFirst", "Select at least one Knight.") });
      return;
    }

    set({
      remoteSubmitting: true,
      error: "",
      message: translate("dashboard.remote.messages.dispatchingTask", "Dispatching remote task..."),
    });

    try {
      const remoteTasks = await SakikoService.SubmitRemoteProfileTask({
        profileId: activeProfileId,
        knightIds: selectedRemoteKnightIDs,
        preset: taskPreset[0] || "ping",
        presets: taskPreset,
        config: taskConfig,
      });
      set((state) => ({
        remoteTasks: mergeRemoteTasks(remoteTasks, state.remoteTasks),
        message: translate("dashboard.remote.messages.dispatchedTasks", "Dispatched {{count}} remote task(s).", {
          count: remoteTasks.length,
        }),
      }));
      void get().refreshRemoteWorkspace(false);
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ remoteSubmitting: false });
    }
  },
});

function mergeRemoteTasks(incoming: DashboardRemoteSlice["remoteTasks"], existing: DashboardRemoteSlice["remoteTasks"]) {
  const next = new Map(existing.map((task) => [task.remoteTaskId, task]));
  for (const task of incoming) {
    next.set(task.remoteTaskId, task);
  }
  return Array.from(next.values()).sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));
}
