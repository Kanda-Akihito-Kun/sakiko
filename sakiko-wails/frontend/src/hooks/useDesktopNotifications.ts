import { Events } from "@wailsio/runtime";
import { useEffect, useRef } from "react";
import { useDashboardStore } from "../store/dashboardStore";
import { type NotificationInput, useNotificationStore } from "../store/notificationStore";
import { normalizeError } from "../utils/dashboard";

const desktopNotificationEventName = "sakiko:desktop-notification";

export function useDesktopNotifications() {
  const pushNotification = useNotificationStore((state) => state.push);
  const dashboardError = useDashboardStore((state) => state.error);
  const lastDashboardErrorRef = useRef("");

  useEffect(() => {
    let unsubscribe = () => {};

    try {
      unsubscribe = Events.On(desktopNotificationEventName, (event) => {
        pushNotification({
          ...(event.data as NotificationInput),
          channel: "backend",
        });
      });
    } catch {
      unsubscribe = () => {};
    }

    return () => {
      unsubscribe();
    };
  }, [pushNotification]);

  useEffect(() => {
    const nextError = dashboardError.trim();
    if (!nextError) {
      lastDashboardErrorRef.current = "";
      return;
    }

    if (nextError === lastDashboardErrorRef.current) {
      return;
    }

    lastDashboardErrorRef.current = nextError;
    pushNotification({
      level: "error",
      message: nextError,
      source: "workspace",
      channel: "workspace",
    });
  }, [dashboardError, pushNotification]);

  useEffect(() => {
    const handleError = (event: ErrorEvent) => {
      pushNotification({
        level: "error",
        message: normalizeError(event.error || event.message || "Unexpected frontend error"),
        source: "frontend",
        channel: "frontend",
      });
    };
    const handleRejection = (event: PromiseRejectionEvent) => {
      pushNotification({
        level: "error",
        message: normalizeError(event.reason || "Unhandled promise rejection"),
        source: "frontend",
        channel: "frontend",
      });
    };

    window.addEventListener("error", handleError);
    window.addEventListener("unhandledrejection", handleRejection);

    return () => {
      window.removeEventListener("error", handleError);
      window.removeEventListener("unhandledrejection", handleRejection);
    };
  }, [pushNotification]);
}
