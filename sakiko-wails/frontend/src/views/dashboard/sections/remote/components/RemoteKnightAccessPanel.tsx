import ComputerRounded from "@mui/icons-material/ComputerRounded";
import LanRounded from "@mui/icons-material/LanRounded";
import { Alert, Box, FormControlLabel, Stack, Switch, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";
import { InfoSurface } from "../../../../../components/shared/InfoSurface";
import { PropertyGrid, PropertyItem } from "../../../../../components/shared/PropertyGrid";
import { SectionCard } from "../../../../../components/shared/SectionCard";
import type { ClusterStatus } from "../../../../../types/sakiko";
import { formatDateTime } from "../../../../../utils/dashboard";

type RemoteKnightAccessPanelProps = {
  currentRole: ClusterStatus["role"];
  fields: {
    masterHost: string;
    masterPort: string;
    oneTimeCode: string;
    remoteJoinText: string;
  };
  knightBound: boolean;
  remoteStatus: ClusterStatus | null;
  remoteSubmitting: boolean;
  onToggleKnight: (enabled: boolean) => void;
  onMasterHostChange: (value: string) => void;
  onMasterPortChange: (value: string) => void;
  onOneTimeCodeChange: (value: string) => void;
  onRemoteJoinTextChange: (value: string) => void;
};

export function RemoteKnightAccessPanel({
  currentRole,
  fields,
  knightBound,
  remoteStatus,
  remoteSubmitting,
  onToggleKnight,
  onMasterHostChange,
  onMasterPortChange,
  onOneTimeCodeChange,
  onRemoteJoinTextChange,
}: RemoteKnightAccessPanelProps) {
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.remote.knightAccess")}
      icon={<ComputerRounded color="primary" />}
    >
      <Stack spacing={1.5}>
        <Alert severity="warning">
          {t("dashboard.remote.knightWarning")}
        </Alert>

        <InfoSurface
          icon={<LanRounded color="primary" />}
          title={t("dashboard.remote.knightAccess")}
        >
          <Stack spacing={1.25}>
            <TextField
              label={t("dashboard.remote.joinInfo")}
              value={fields.remoteJoinText}
              onChange={(event) => onRemoteJoinTextChange(event.target.value)}
              placeholder={t("dashboard.remote.joinInfoPlaceholder")}
              size="small"
              fullWidth
              multiline
              minRows={2}
            />
            <Box
              sx={{
                display: "grid",
                gap: 1.25,
                gridTemplateColumns: {
                  xs: "minmax(0, 1fr)",
                  sm: "minmax(0, 3fr) minmax(0, 1fr)",
                },
              }}
            >
              <TextField
                label={t("dashboard.remote.masterHost")}
                value={fields.masterHost}
                onChange={(event) => onMasterHostChange(event.target.value)}
                size="small"
                fullWidth
              />
              <TextField
                label={t("dashboard.remote.masterPort")}
                value={fields.masterPort}
                onChange={(event) => onMasterPortChange(event.target.value)}
                size="small"
                fullWidth
              />
            </Box>
            <TextField
              label={t("dashboard.remote.oneTimePairingCode")}
              value={fields.oneTimeCode}
              onChange={(event) => onOneTimeCodeChange(event.target.value)}
              size="small"
              fullWidth
            />
          </Stack>
        </InfoSurface>

        <InfoSurface
          icon={<LanRounded color="primary" />}
          title={t("dashboard.remote.knightState")}
          action={(
            <FormControlLabel
              sx={{ mr: 0 }}
              control={(
                <Switch
                  checked={currentRole === "knight"}
                  disabled={remoteSubmitting || currentRole === "master"}
                  onChange={(_event, checked) => onToggleKnight(checked)}
                />
              )}
              label={currentRole === "knight" ? t("dashboard.remote.knightOn") : t("dashboard.remote.knightOff")}
            />
          )}
        >
          <PropertyGrid>
            <PropertyItem label={t("dashboard.remote.knightState")} value={knightBound ? t("dashboard.remote.bound") : t("dashboard.remote.notBound")} />
            <PropertyItem label={t("dashboard.remote.role")} value={t(`dashboard.remote.roles.${currentRole}`)} mono={false} />
            <PropertyItem label={t("dashboard.remote.masterHost")} value={remoteStatus?.knight?.masterHost || "-"} />
            <PropertyItem label={t("dashboard.remote.masterPort")} value={remoteStatus?.knight?.masterPort ? String(remoteStatus.knight.masterPort) : "-"} />
            <PropertyItem label={t("dashboard.remote.connectionState")} value={remoteStatus?.knight?.connected ? t("dashboard.remote.connected") : t("dashboard.remote.disconnected")} />
            <PropertyItem label={t("dashboard.remote.lastSeenAt")} value={formatOptionalDateTime(remoteStatus?.knight?.lastSeenAt, "-")} />
            <PropertyItem
              label={t("dashboard.remote.localTasks")}
              value={currentRole === "knight" ? t("dashboard.remote.disabledWhileKnight") : t("dashboard.remote.allowed")}
              span={2}
              mono={false}
            />
            <PropertyItem
              label={t("dashboard.remote.error")}
              value={remoteStatus?.knight?.lastError || "-"}
              span={2}
              mono={false}
              multiline
            />
          </PropertyGrid>
        </InfoSurface>
      </Stack>
    </SectionCard>
  );
}

function formatOptionalDateTime(value: string | undefined, fallback: string): string {
  if (!value) {
    return fallback;
  }
  return formatDateTime(value);
}
