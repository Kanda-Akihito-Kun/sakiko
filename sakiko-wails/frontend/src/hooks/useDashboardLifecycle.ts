import { useEffect, useRef } from "react";
import { useDashboardStore } from "../store/dashboardStore";

export function useDashboardLifecycle() {
  const refreshDashboard = useDashboardStore((state) => state.refreshDashboard);
  const refreshResultArchives = useDashboardStore((state) => state.refreshResultArchives);
  const activeTaskId = useDashboardStore((state) => state.activeTaskId);
  const activeTaskStatus = useDashboardStore((state) => state.activeTask?.task?.status);
  const tasks = useDashboardStore((state) => state.tasks);
  const refreshTasks = useDashboardStore((state) => state.refreshTasks);
  const handleInspectTask = useDashboardStore((state) => state.handleInspectTask);
  const syncActiveTask = useDashboardStore((state) => state.syncActiveTask);
  const initializedRef = useRef(false);

  useEffect(() => {
    if (initializedRef.current) {
      return;
    }

    initializedRef.current = true;
    void refreshDashboard();
    void refreshResultArchives();
  }, [refreshDashboard, refreshResultArchives]);

  useEffect(() => {
    if (!activeTaskId) {
      return;
    }

    void syncActiveTask();
  }, [activeTaskId, syncActiveTask]);

  useEffect(() => {
    const runningTask = tasks.find((task) => task.status === "running" || task.status === "stopping");
    if (!runningTask) {
      return;
    }
    if (activeTaskId === runningTask.taskId && (activeTaskStatus === "running" || activeTaskStatus === "stopping")) {
      return;
    }
    if (activeTaskId && activeTaskStatus && activeTaskStatus !== "finished" && activeTaskStatus !== "failed" && activeTaskStatus !== "canceled") {
      return;
    }

    void handleInspectTask(runningTask.taskId);
  }, [activeTaskId, activeTaskStatus, handleInspectTask, tasks]);

  useEffect(() => {
    if (!activeTaskId || (activeTaskStatus !== "running" && activeTaskStatus !== "stopping")) {
      return;
    }

    let cancelled = false;
    let timer = 0;

    const loop = async () => {
      await syncActiveTask();
      if (cancelled) {
        return;
      }
      timer = window.setTimeout(() => {
        void loop();
      }, 500);
    };

    timer = window.setTimeout(() => {
      void loop();
    }, 500);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [activeTaskId, activeTaskStatus, syncActiveTask]);

  useEffect(() => {
    if (activeTaskId) {
      return;
    }

    let cancelled = false;
    let timer = 0;

    const loop = async () => {
      await refreshTasks();
      if (cancelled) {
        return;
      }
      timer = window.setTimeout(() => {
        void loop();
      }, 1000);
    };

    timer = window.setTimeout(() => {
      void loop();
    }, 1000);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [activeTaskId, refreshTasks]);
}
