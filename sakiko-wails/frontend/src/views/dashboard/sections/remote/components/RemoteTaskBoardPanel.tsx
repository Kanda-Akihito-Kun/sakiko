import PendingActionsRounded from "@mui/icons-material/PendingActionsRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import { Button, Chip, List, ListItem, ListItemText } from "@mui/material";
import { useTranslation } from "react-i18next";
import { SectionCard } from "../../../../../components/shared/SectionCard";
import type { ClusterRemoteTask } from "../../../../../types/sakiko";
import { formatDateTime, summarizeActiveTaskNodes } from "../../../../../utils/dashboard";

type RemoteTaskBoardPanelProps = {
  remoteLoading: boolean;
  remoteSubmitting: boolean;
  remoteTasks: ClusterRemoteTask[];
  onRefresh: () => void;
};

export function RemoteTaskBoardPanel({
  remoteLoading,
  remoteSubmitting,
  remoteTasks,
  onRefresh,
}: RemoteTaskBoardPanelProps) {
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.remote.remoteTasks")}
      icon={<PendingActionsRounded color="primary" />}
      action={(
        <Button
          variant="outlined"
          size="small"
          startIcon={<RefreshRounded />}
          disabled={remoteLoading || remoteSubmitting}
          onClick={onRefresh}
        >
          {t("dashboard.remote.refresh")}
        </Button>
      )}
    >
      <List disablePadding>
        {remoteTasks.length === 0 ? (
          <ListItem sx={{ px: 0, py: 1.25 }}>
            <ListItemText
              primary={t("dashboard.remote.noRemoteTasks")}
              secondary={t("dashboard.remote.noRemoteTasksDetail")}
            />
          </ListItem>
        ) : remoteTasks.map((task) => (
          <ListItem
            key={task.remoteTaskId}
            sx={{
              px: 0,
              py: 1.25,
              borderBottom: "1px solid",
              borderColor: "divider",
              alignItems: "flex-start",
            }}
          >
            <ListItemText
              primary={task.taskName || task.remoteTaskId}
              secondary={[
                `${t("dashboard.remote.knightNameLabel")}: ${task.knightName || task.knightId}`,
                `${t("dashboard.remote.stateLabel")}: ${t(`dashboard.remote.taskStates.${task.state}`)}`,
                task.runtime ? `${t("dashboard.remote.progressLabel")}: ${task.runtime.progress}/${task.runtime.total}` : "",
                task.runtime?.activeNodes?.length ? `${t("dashboard.remote.runtimeLabel")}: ${summarizeActiveTaskNodes(task.runtime.activeNodes)}` : "",
                task.createdAt ? `${t("dashboard.remote.createdAt")}: ${formatDateTime(task.createdAt)}` : "",
                task.startedAt ? `${t("dashboard.remote.startedAt")}: ${formatDateTime(task.startedAt)}` : "",
                task.finishedAt ? `${t("dashboard.remote.finishedAt")}: ${formatDateTime(task.finishedAt)}` : "",
                task.exitCode ? `${t("dashboard.remote.exitCode")}: ${task.exitCode}` : "",
                task.error ? `${t("dashboard.remote.error")}: ${task.error}` : "",
              ].filter(Boolean).join("\n")}
              secondaryTypographyProps={{
                component: "div",
                sx: { whiteSpace: "pre-wrap", overflowWrap: "anywhere" },
              }}
            />
            <Chip
              label={t(`dashboard.remote.taskStates.${task.state}`)}
              color={task.state === "finished" ? "success" : task.state === "failed" ? "error" : task.state === "running" ? "warning" : "default"}
              size="small"
            />
          </ListItem>
        ))}
      </List>
    </SectionCard>
  );
}
