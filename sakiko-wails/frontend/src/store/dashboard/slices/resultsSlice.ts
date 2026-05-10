import { SakikoService } from "../../../services/sakikoService";
import { translate } from "../../../services/i18n";
import { normalizeError } from "../../../utils/dashboard";
import type { DashboardResultsSlice, DashboardSliceCreator } from "../types";

export const createDashboardResultsSlice: DashboardSliceCreator<DashboardResultsSlice> = (set, get) => ({
  resultArchives: [],
  resultArchivesLoading: false,
  resultArchiveDetails: {},
  resultArchiveLoading: {},
  resultArchivesVisibleCount: 10,

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
});
