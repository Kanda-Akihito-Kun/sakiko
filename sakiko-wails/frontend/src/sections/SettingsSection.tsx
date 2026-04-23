import SettingsRounded from "@mui/icons-material/SettingsRounded";
import { Box, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { UpdatePanel } from "../components/settings/UpdatePanel";
import { SettingsPanel } from "../components/settings/SettingsPanel";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";

type SettingsSectionProps = {
  mode: string;
  profilesPath: string;
  resolvedMode: string;
};

export function SettingsSection({
  mode,
  profilesPath,
  resolvedMode,
}: SettingsSectionProps) {
  const { t } = useTranslation();

  return (
    <Box className="sakiko-section-grid">
      <Box className="sakiko-section-grid__sidebar">
        <SettingsPanel />
      </Box>

      <Box className="sakiko-section-grid__content">
        <Box
          sx={{
            display: "grid",
            gap: 2.25,
            gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
            alignItems: "start",
            "& > *": {
              minWidth: 0,
            },
          }}
        >
          <UpdatePanel />

          <SectionCard
            title={t("settings.environment.title")}
            icon={<SettingsRounded color="primary" />}
          >
            <Stack spacing={1.25}>
              <OverviewRow
                label={t("settings.environment.mode")}
                value={mode === "system" ? `${t("settings.themeOptions.system")} -> ${t(`settings.themeOptions.${resolvedMode}`)}` : t(`settings.themeOptions.${resolvedMode}`)}
                mono
              />
              <OverviewRow
                label={t("settings.environment.workspace")}
                value={profilesPath || t("settings.environment.workspaceUnavailable")}
                mono
                multiline
              />
            </Stack>
          </SectionCard>
        </Box>
      </Box>
    </Box>
  );
}
