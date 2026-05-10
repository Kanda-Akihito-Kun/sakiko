import HubRounded from "@mui/icons-material/HubRounded";
import {
  Alert,
  Button,
  Checkbox,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Stack,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { SectionCard } from "../../../../../components/shared/SectionCard";
import type { ClusterConnectedKnight, ClusterStatus } from "../../../../../types/sakiko";
import { formatDateTime } from "../../../../../utils/dashboard";

type RemoteKnightsListPanelProps = {
  currentRole: ClusterStatus["role"];
  remoteKnights: ClusterConnectedKnight[];
  selectedKnightIDs: string[];
  remoteSubmitting: boolean;
  onKickKnight: (knightId: string) => void;
  onToggleKnight: (knightId: string) => void;
};

export function RemoteKnightsListPanel({
  currentRole,
  remoteKnights,
  remoteSubmitting,
  selectedKnightIDs,
  onKickKnight,
  onToggleKnight,
}: RemoteKnightsListPanelProps) {
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.remote.connectedKnights")}
      icon={<HubRounded color="primary" />}
    >
      <Stack spacing={1.25}>
        <Alert severity="info">
          {t("dashboard.remote.connectedKnightsHint")}
        </Alert>
        <List disablePadding>
          {remoteKnights.length === 0 ? (
            <ListItem
              sx={{
                px: 0,
                py: 1,
                borderBottom: "1px solid",
                borderColor: "divider",
              }}
            >
              <ListItemText
                primary={t("dashboard.remote.noConnectedKnights")}
                secondary={t("dashboard.remote.noConnectedKnightsDetail")}
              />
            </ListItem>
          ) : remoteKnights.map((knight) => (
            <ListItem
              key={knight.knightId}
              disablePadding
              sx={{
                px: 0,
                borderBottom: "1px solid",
                borderColor: "divider",
              }}
            >
              <ListItemButton
                disabled={currentRole !== "master"}
                onClick={() => onToggleKnight(knight.knightId)}
                sx={{ px: 0, py: 1, alignItems: "center" }}
              >
                {currentRole === "master" ? (
                  <Checkbox
                    edge="start"
                    checked={selectedKnightIDs.includes(knight.knightId)}
                    tabIndex={-1}
                    disableRipple
                    sx={{
                      alignSelf: "center",
                      mr: 0.5,
                      "& .MuiSvgIcon-root": {
                        fontSize: 28,
                        borderRadius: 1.75,
                      },
                    }}
                  />
                ) : null}
                <Stack direction="row" spacing={1.25} sx={{ width: "100%", alignItems: "flex-start" }}>
                  <ListItemText
                    primary={knight.knightName || knight.knightId}
                    secondary={[
                      `${t("dashboard.remote.knightID")}: ${knight.knightId}`,
                      `${t("dashboard.remote.stateLabel")}: ${t(`dashboard.remote.knightStates.${knight.state}`)}`,
                      knight.remoteAddr ? `${t("dashboard.remote.remoteAddress")}: ${knight.remoteAddr}` : "",
                      knight.lastSeenAt ? `${t("dashboard.remote.lastSeenAt")}: ${formatDateTime(knight.lastSeenAt)}` : "",
                      knight.lastError ? `${t("dashboard.remote.error")}: ${knight.lastError}` : "",
                    ].filter(Boolean).join("\n")}
                    secondaryTypographyProps={{
                      component: "div",
                      sx: { whiteSpace: "pre-wrap", overflowWrap: "anywhere" },
                    }}
                  />
                  {currentRole === "master" ? (
                    <Button
                      color="error"
                      size="small"
                      variant="outlined"
                      disabled={remoteSubmitting}
                      onClick={(event) => {
                        event.preventDefault();
                        event.stopPropagation();
                        onKickKnight(knight.knightId);
                      }}
                    >
                      {t("dashboard.remote.kickKnight")}
                    </Button>
                  ) : null}
                </Stack>
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Stack>
    </SectionCard>
  );
}
