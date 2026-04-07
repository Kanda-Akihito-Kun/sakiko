import { useEffect, useRef } from "react";
import { useDashboardStore } from "../store/dashboardStore";

export function useDashboardLifecycle() {
  const refreshDashboard = useDashboardStore((state) => state.refreshDashboard);
  const refreshDownloadTargets = useDashboardStore((state) => state.refreshDownloadTargets);
  const refreshResultArchives = useDashboardStore((state) => state.refreshResultArchives);
  const activeTaskId = useDashboardStore((state) => state.activeTaskId);
  const activeTaskStatus = useDashboardStore((state) => state.activeTask?.task?.status);
  const syncActiveTask = useDashboardStore((state) => state.syncActiveTask);
  const initializedRef = useRef(false);

  useEffect(() => {
    if (initializedRef.current) {
      return;
    }

    initializedRef.current = true;
    void refreshDashboard();
    void refreshDownloadTargets();
    void refreshResultArchives();
  }, [refreshDashboard, refreshDownloadTargets, refreshResultArchives]);

  useEffect(() => {
    if (!activeTaskId) {
      return;
    }

    void syncActiveTask();
  }, [activeTaskId, syncActiveTask]);

  useEffect(() => {
    if (!activeTaskId || activeTaskStatus !== "running") {
      return;
    }

    const timer = window.setInterval(() => {
      void syncActiveTask();
    }, 500);

    return () => {
      window.clearInterval(timer);
    };
  }, [activeTaskId, activeTaskStatus, syncActiveTask]);
}
