import CloseRounded from "@mui/icons-material/CloseRounded";
import ErrorOutlineRounded from "@mui/icons-material/ErrorOutlineRounded";
import WarningAmberRounded from "@mui/icons-material/WarningAmberRounded";
import { Alert, AlertTitle, IconButton, Stack, Typography } from "@mui/material";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNotificationStore, type NotificationItem } from "../../store/notificationStore";

const AUTO_HIDE_MILLIS: Record<NotificationItem["level"], number> = {
  warning: 8000,
  error: 12000,
};

export function DesktopNotificationCenter() {
  const items = useNotificationStore((state) => state.items);
  const dismiss = useNotificationStore((state) => state.dismiss);

  if (items.length === 0) {
    return null;
  }

  return (
    <Stack
      spacing={1.25}
      sx={{
        position: "fixed",
        top: 20,
        right: 20,
        width: "min(420px, calc(100vw - 32px))",
        zIndex: (theme) => theme.zIndex.snackbar,
        pointerEvents: "none",
      }}
    >
      {items.map((item) => (
        <DesktopNotificationToast key={item.id} item={item} onDismiss={dismiss} />
      ))}
    </Stack>
  );
}

type DesktopNotificationToastProps = {
  item: NotificationItem;
  onDismiss: (id: string) => void;
};

function DesktopNotificationToast({ item, onDismiss }: DesktopNotificationToastProps) {
  const { t } = useTranslation();

  useEffect(() => {
    const timer = window.setTimeout(() => {
      onDismiss(item.id);
    }, AUTO_HIDE_MILLIS[item.level]);

    return () => {
      window.clearTimeout(timer);
    };
  }, [item.id, item.level, onDismiss]);

  return (
    <Alert
      severity={item.level}
      variant="filled"
      icon={item.level === "error" ? <ErrorOutlineRounded fontSize="inherit" /> : <WarningAmberRounded fontSize="inherit" />}
      action={(
        <IconButton
          size="small"
          color="inherit"
          aria-label={t("shared.notifications.dismiss", "Dismiss notification")}
          onClick={() => onDismiss(item.id)}
        >
          <CloseRounded fontSize="small" />
        </IconButton>
      )}
      sx={{
        pointerEvents: "auto",
        alignItems: "flex-start",
        borderRadius: 2.5,
        boxShadow: (theme) => theme.shadows[12],
        "& .MuiAlert-action": {
          alignItems: "flex-start",
          pt: 0.5,
        },
      }}
    >
      <AlertTitle sx={{ mb: 0.5 }}>
        {item.level === "error"
          ? t("shared.notifications.errorTitle", "Error")
          : t("shared.notifications.warningTitle", "Warning")}
      </AlertTitle>
      <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
        {item.message}
      </Typography>
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 1 }}>
        <Typography variant="caption" sx={{ opacity: 0.9 }}>
          {formatNotificationSource(item.source, t)}
        </Typography>
        <Typography variant="caption" sx={{ opacity: 0.7 }}>
          {formatNotificationTime(item.timestamp)}
        </Typography>
        {item.count > 1 ? (
          <Typography variant="caption" sx={{ opacity: 0.9, fontWeight: 700 }}>
            {t("shared.notifications.repeatCount", "x{{count}}", { count: item.count })}
          </Typography>
        ) : null}
      </Stack>
    </Alert>
  );
}

function formatNotificationSource(source: string, t: (key: string, defaultValue: string) => string): string {
  switch (source.trim().toLowerCase()) {
    case "backend":
      return t("shared.notifications.sources.backend", "Backend");
    case "workspace":
      return t("shared.notifications.sources.workspace", "Workspace");
    case "frontend":
      return t("shared.notifications.sources.frontend", "Frontend");
    default:
      return source;
  }
}

function formatNotificationTime(timestamp: string): string {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}
