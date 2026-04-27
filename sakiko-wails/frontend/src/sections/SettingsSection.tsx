import SettingsRounded from "@mui/icons-material/SettingsRounded";
import { Box, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { PrivacySettingsPanel } from "../components/settings/PrivacySettingsPanel";
import { UpdatePanel } from "../components/settings/UpdatePanel";
import { SettingsPanel } from "../components/settings/SettingsPanel";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";
import type { BackendInfo } from "../types/sakiko";

type SettingsSectionProps = {
  profilesPath: string;
  mihomoVersion: string;
  networkEnv: BackendInfo | null;
};

export function SettingsSection({
  profilesPath,
  mihomoVersion,
  networkEnv,
}: SettingsSectionProps) {
  const { t } = useTranslation();

  return (
    <Box
      sx={{
        display: "grid",
        gap: 2.25,
        gridTemplateColumns: {
          xs: "minmax(0, 1fr)",
          lg: "repeat(2, minmax(0, 1fr))",
        },
        alignItems: "stretch",
        "& > *": {
          minWidth: 0,
        },
      }}
    >
      <PrivacySettingsPanel />
      <SettingsPanel />
      <UpdatePanel />

      <SectionCard
        title={t("settings.environment.title")}
        icon={<SettingsRounded color="primary" />}
        sx={{ height: "100%" }}
      >
        <Stack spacing={1.25}>
          <OverviewRow
            label={t("settings.environment.workspace")}
            value={profilesPath || t("settings.environment.workspaceUnavailable")}
            mono
            multiline
          />
          <OverviewRow
            label={t("settings.environment.mihomoVersion")}
            value={mihomoVersion || t("settings.environment.versionUnavailable")}
            mono
          />
          <OverviewRow
            label={t("settings.environment.networkEnv")}
            value={formatNetworkEnv(networkEnv, t("settings.environment.networkUnavailable"))}
            multiline
          />
        </Stack>
      </SectionCard>
    </Box>
  );
}

function formatNetworkEnv(networkEnv: BackendInfo | null, fallback: string): string {
  if (!networkEnv) {
    return fallback;
  }

  const parts = [
    networkEnv.location || "",
    networkEnv.ip ? `IP=${networkEnv.ip}` : "",
    networkEnv.source ? `Source=${networkEnv.source}` : "",
    networkEnv.error ? `Error=${networkEnv.error}` : "",
  ].filter(Boolean);

  return parts.join("\n") || fallback;
}
