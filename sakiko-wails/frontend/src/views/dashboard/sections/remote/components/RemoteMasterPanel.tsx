import KeyRounded from "@mui/icons-material/KeyRounded";
import PublicRounded from "@mui/icons-material/PublicRounded";
import RadarRounded from "@mui/icons-material/RadarRounded";
import SyncRounded from "@mui/icons-material/SyncRounded";
import ShieldRounded from "@mui/icons-material/ShieldRounded";
import ContentCopyRounded from "@mui/icons-material/ContentCopyRounded";
import {
  Alert,
  Box,
  Button,
  Divider,
  FormControlLabel,
  Stack,
  Switch,
  TextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { DetailRow } from "../../../../../components/shared/DetailRow";
import { InfoSurface } from "../../../../../components/shared/InfoSurface";
import { PropertyGrid, PropertyItem, StatusSummaryTile } from "../../../../../components/shared/PropertyGrid";
import { SectionCard } from "../../../../../components/shared/SectionCard";
import type { ClusterConnectedKnight, ClusterStatus, MasterEligibility } from "../../../../../types/sakiko";
import { formatDateTime } from "../../../../../utils/dashboard";
import { buildRemoteJoinText } from "../../../../../utils/remoteInvite";

type RemoteMasterPanelProps = {
  currentEligibility: MasterEligibility | null;
  currentRole: ClusterStatus["role"];
  fields: {
    knightName: string;
    listenHost: string;
    listenPort: string;
    ttlSeconds: string;
  };
  masterEnabled: boolean;
  remoteKnights: ClusterConnectedKnight[];
  remoteLoading: boolean;
  remotePairingCode: {
    code: string;
    knightName?: string;
    expiresAt?: string;
  } | null;
  remoteSubmitting: boolean;
  onCreatePairingCode: () => void;
  onToggleMaster: (enabled: boolean) => void;
  onProbeEligibility: () => void;
  onRefresh: () => void;
  onKnightNameChange: (value: string) => void;
  onListenHostChange: (value: string) => void;
  onListenPortChange: (value: string) => void;
  onTTLChange: (value: string) => void;
};

export function RemoteMasterPanel({
  currentEligibility,
  currentRole,
  fields,
  masterEnabled,
  remoteKnights,
  remoteLoading,
  remotePairingCode,
  remoteSubmitting,
  onCreatePairingCode,
  onToggleMaster,
  onProbeEligibility,
  onRefresh,
  onKnightNameChange,
  onListenHostChange,
  onListenPortChange,
  onTTLChange,
}: RemoteMasterPanelProps) {
  const { t } = useTranslation();
  const joinHost = currentEligibility?.publicIP || fields.listenHost;
  const joinText = buildRemoteJoinText({
    host: joinHost,
    port: fields.listenPort,
    code: remotePairingCode?.code || "",
  });
  const copyJoinText = () => {
    if (!joinText || typeof navigator === "undefined" || !navigator.clipboard) {
      return;
    }
    void navigator.clipboard.writeText(joinText);
  };

  return (
    <SectionCard
      title={t("dashboard.remote.masterConsole")}
      icon={<ShieldRounded color="primary" />}
      action={(
        <Button
          variant="outlined"
          size="small"
          startIcon={<SyncRounded />}
          disabled={remoteLoading || remoteSubmitting}
          onClick={onRefresh}
        >
          {t("dashboard.remote.refresh")}
        </Button>
      )}
    >
      <Stack spacing={2}>
        <Alert severity="info">
          {t("dashboard.remote.connectionRule")}
        </Alert>

        <PropertyGrid
          columns={{
            xs: "minmax(0, 1fr)",
            md: "repeat(3, minmax(0, 1fr))",
          }}
        >
          <StatusSummaryTile title={t("dashboard.remote.currentRole")} value={t(`dashboard.remote.roles.${currentRole}`)} color={roleColor(currentRole)} />
          <StatusSummaryTile
            title={t("dashboard.remote.masterState")}
            value={masterEnabled ? t("dashboard.remote.listening") : t("dashboard.remote.offline")}
            color={masterEnabled ? "success" : "default"}
          />
          <StatusSummaryTile title={t("dashboard.remote.knightCount")} value={String(remoteKnights.length)} color="default" />
        </PropertyGrid>

        <InfoSurface
          icon={<PublicRounded color="primary" />}
          title={t("dashboard.remote.probeCardTitle")}
          action={(
            <FormControlLabel
              sx={{ mr: 0 }}
              control={(
                <Switch
                  checked={masterEnabled}
                  disabled={remoteSubmitting || currentRole === "knight"}
                  onChange={(_event, checked) => onToggleMaster(checked)}
                />
              )}
              label={masterEnabled ? t("dashboard.remote.masterOn") : t("dashboard.remote.masterOff")}
            />
          )}
        >
          <PropertyGrid>
            <PropertyItem label={t("dashboard.remote.publicIP")} value={currentEligibility?.publicIP || t("dashboard.remote.unavailable")} />
            <PropertyItem label={t("dashboard.remote.natType")} value={currentEligibility?.natType || t("dashboard.remote.unknown")} />
            <PropertyItem label={t("dashboard.remote.reachable")} value={formatBool(currentEligibility?.reachable, t)} />
            <PropertyItem label={t("dashboard.remote.eligible")} value={formatBool(currentEligibility?.eligible, t)} />
            <PropertyItem label={t("dashboard.remote.checkedAt")} value={formatOptionalDateTime(currentEligibility?.checkedAt, t("dashboard.remote.waiting"))} />
            <PropertyItem
              label={t("dashboard.remote.error")}
              value={currentEligibility?.error || t("dashboard.remote.unknown")}
              span={2}
              multiline
            />
          </PropertyGrid>

          <Divider />

          <PropertyGrid>
            <TextField
              label={t("dashboard.remote.listenHost")}
              value={fields.listenHost}
              onChange={(event) => onListenHostChange(event.target.value)}
              size="small"
              fullWidth
            />
            <TextField
              label={t("dashboard.remote.listenPort")}
              value={fields.listenPort}
              onChange={(event) => onListenPortChange(event.target.value)}
              size="small"
              fullWidth
            />
          </PropertyGrid>

          <Button
            variant="outlined"
            startIcon={<RadarRounded />}
            disabled={remoteSubmitting || currentRole === "knight"}
            onClick={onProbeEligibility}
          >
            {t("dashboard.remote.probeEligibility")}
          </Button>
        </InfoSurface>

        <InfoSurface
          icon={<KeyRounded color="primary" />}
          title={t("dashboard.remote.pairingCode")}
        >
          <Box
            sx={{
              display: "grid",
              gap: 1.25,
              gridTemplateColumns: {
                xs: "minmax(0, 1fr)",
                lg: "minmax(0, 0.95fr) minmax(0, 1.05fr)",
              },
              minWidth: 0,
            }}
          >
            <Stack spacing={1.25}>
              <TextField
                label={t("dashboard.remote.knightName")}
                value={fields.knightName}
                onChange={(event) => onKnightNameChange(event.target.value)}
                size="small"
                fullWidth
              />
              <TextField
                label={t("dashboard.remote.ttlSeconds")}
                value={fields.ttlSeconds}
                onChange={(event) => onTTLChange(event.target.value)}
                size="small"
                fullWidth
              />
              <Button
                variant="contained"
                startIcon={<KeyRounded />}
                disabled={remoteSubmitting || currentRole !== "master"}
                onClick={onCreatePairingCode}
              >
                {t("dashboard.remote.generatePairingCode")}
              </Button>
            </Stack>

            <Stack spacing={1.25}>
              <TextField
                label={t("dashboard.remote.joinInfo")}
                value={joinText || t("dashboard.remote.noJoinInfo")}
                size="small"
                fullWidth
                multiline
                minRows={2}
                InputProps={{
                  readOnly: true,
                  sx: { fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" },
                  endAdornment: (
                    <Button
                      size="small"
                      startIcon={<ContentCopyRounded />}
                      disabled={!joinText}
                      onClick={copyJoinText}
                      sx={{ whiteSpace: "nowrap", ml: 1 }}
                    >
                      {t("dashboard.remote.copy")}
                    </Button>
                  ),
                }}
              />
              <DetailRow
                label={t("dashboard.remote.code")}
                value={remotePairingCode?.code || t("dashboard.remote.noPairingCode")}
                mono
                multiline
              />
              <DetailRow label={t("dashboard.remote.knightName")} value={remotePairingCode?.knightName || "-"} />
              <DetailRow label={t("dashboard.remote.expiresAt")} value={formatOptionalDateTime(remotePairingCode?.expiresAt, "-")} mono multiline />
            </Stack>
          </Box>
        </InfoSurface>
      </Stack>
    </SectionCard>
  );
}

function formatBool(value: boolean | undefined, t: ReturnType<typeof useTranslation>["t"]): string {
  if (typeof value !== "boolean") {
    return t("dashboard.remote.waiting");
  }
  return value ? t("dashboard.remote.yes") : t("dashboard.remote.no");
}

function formatOptionalDateTime(value: string | undefined, fallback: string): string {
  if (!value) {
    return fallback;
  }
  return formatDateTime(value);
}

function roleColor(role: ClusterStatus["role"]): "default" | "success" | "warning" {
  if (role === "master") {
    return "success";
  }
  if (role === "knight") {
    return "warning";
  }
  return "default";
}
