import SettingsRounded from "@mui/icons-material/SettingsRounded";
import { Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useShallow } from "zustand/react/shallow";
import { PrivacySettingsPanel } from "../../../../components/settings/PrivacySettingsPanel";
import { SettingsPanel } from "../../../../components/settings/SettingsPanel";
import { UpdatePanel } from "../../../../components/settings/UpdatePanel";
import { DetailRow } from "../../../../components/shared/DetailRow";
import { SectionCard } from "../../../../components/shared/SectionCard";
import { useDashboardStore } from "../../../../store/dashboardStore";
import type { BackendInfo } from "../../../../types/sakiko";
import { SectionLayout } from "../../components/SectionLayout";

export function SettingsDashboardSection() {
  const { t } = useTranslation();
  const view = useDashboardStore(useShallow((state) => ({
    profilesPath: state.profilesPath,
    mihomoVersion: state.mihomoVersion,
    networkEnv: state.networkEnv,
  })));

  return (
    <SectionLayout
      columns={{
        xs: "minmax(0, 1fr)",
        lg: "repeat(2, minmax(0, 1fr))",
      }}
      alignItems="stretch"
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
          <DetailRow
            label={t("settings.environment.workspace")}
            value={view.profilesPath || t("settings.environment.workspaceUnavailable")}
            mono
            multiline
          />
          <DetailRow
            label={t("settings.environment.mihomoVersion")}
            value={view.mihomoVersion || t("settings.environment.versionUnavailable")}
            mono
          />
          <DetailRow
            label={t("settings.environment.networkEnv")}
            value={formatNetworkEnv(view.networkEnv, t("settings.environment.networkUnavailable"))}
            multiline
          />
        </Stack>
      </SectionCard>
    </SectionLayout>
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
