import { SakikoService } from "../../../services/sakikoService";
import { translate } from "../../../services/i18n";
import { createSubmitProfileTaskPayload } from "../../dashboardPayloads";
import { formatTaskPresetSelectionLabel, normalizeError } from "../../../utils/dashboard";
import type { DashboardSliceCreator, DashboardTasksSlice } from "../types";

let syncActiveTaskTaskId = "";
let syncActiveTaskPromise: Promise<void> | null = null;

export const createDashboardTasksSlice: DashboardSliceCreator<DashboardTasksSlice> = (set, get) => ({
  tasks: [],
  activeTaskId: "",
  activeTask: null,

  refreshTasks: async () => {
    try {
      const latestTasks = await SakikoService.ListTasks();
      set({ tasks: latestTasks });
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  handleRunTask: async () => {
    const { activeProfileId, taskConfig, taskPreset } = get();
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

  handleStopTask: async () => {
    const { activeTask, activeTaskId, tasks } = get();
    if (!activeTaskId || !activeTask) {
      return;
    }
    if (activeTask.task.status !== "running" && activeTask.task.status !== "stopping") {
      return;
    }

    set({
      error: "",
      message: translate("dashboard.messages.stoppingTask", `Stopping ${activeTask.task.name}...`, {
        name: activeTask.task.name,
      }),
      activeTask: {
        ...activeTask,
        task: {
          ...activeTask.task,
          status: "stopping",
        },
      },
      tasks: tasks.map((task) => task.taskId === activeTaskId ? { ...task, status: "stopping" } : task),
    });

    try {
      await SakikoService.CancelTask(activeTaskId);
      void get().syncActiveTask();
    } catch (err) {
      set((state) => ({
        error: normalizeError(err),
        activeTask: state.activeTask && state.activeTask.task.taskId === activeTaskId
          ? {
              ...state.activeTask,
              task: {
                ...state.activeTask.task,
                status: "running",
              },
            }
          : state.activeTask,
        tasks: state.tasks.map((task) => task.taskId === activeTaskId && task.status === "stopping"
          ? { ...task, status: "running" }
          : task),
      }));
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
          message: selected && selected.status !== "running" && selected.status !== "stopping"
            ? translate("dashboard.messages.taskFinished", `Task ${selected.name} finished with ${detail.results?.length ?? 0} result(s).`, {
              name: selected.name,
              count: detail.results?.length ?? 0,
            })
            : get().message,
        });

        if (selected && selected.status !== "running" && selected.status !== "stopping" && !get().resultArchives.some((item) => item.taskId === activeTaskId)) {
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

  handleDeleteTask: async (taskId) => {
    const task = get().tasks.find((item) => item.taskId === taskId);
    const taskName = task?.name || taskId;
    if (!taskId) {
      return;
    }

    set({
      error: "",
      message: translate("dashboard.messages.deletingTask", `Deleting task ${taskName}...`, {
        name: taskName,
      }),
    });

    try {
      await SakikoService.DeleteTask(taskId);
      set((state) => {
        const nextTasks = state.tasks.filter((item) => item.taskId !== taskId);
        const deletingActive = state.activeTaskId === taskId;

        return {
          tasks: nextTasks,
          activeTaskId: deletingActive ? "" : state.activeTaskId,
          activeTask: deletingActive ? null : state.activeTask,
          message: translate("dashboard.messages.deletedTask", `Deleted task ${taskName}.`, {
            name: taskName,
          }),
        };
      });
    } catch (err) {
      set({ error: normalizeError(err) });
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
});
