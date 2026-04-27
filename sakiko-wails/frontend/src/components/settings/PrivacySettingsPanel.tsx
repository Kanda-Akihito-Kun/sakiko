import SecurityRounded from "@mui/icons-material/SecurityRounded";
import { Alert, Box, CircularProgress, List, ListItem, ListItemText, Stack, Switch, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { SakikoService } from "../../services/sakikoService";
import type { AppSettings } from "../../types/appSettings";
import { SectionCard } from "../shared/SectionCard";

type PrivacySettingsState = Pick<AppSettings, "hideProfileNameInExport" | "hideCNInboundInExport">;

export function PrivacySettingsPanel() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [savingKey, setSavingKey] = useState<keyof PrivacySettingsState | "">("");
  const [error, setError] = useState("");
  const [settings, setSettings] = useState<PrivacySettingsState>({
    hideProfileNameInExport: true,
    hideCNInboundInExport: true,
  });

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError("");
      try {
        const nextSettings = await SakikoService.GetAppSettings();
        if (cancelled) {
          return;
        }
        setSettings({
          hideProfileNameInExport: nextSettings.hideProfileNameInExport,
          hideCNInboundInExport: nextSettings.hideCNInboundInExport,
        });
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : t("settings.privacy.loadFailed"));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [t]);

  const toggleSetting = async (key: keyof PrivacySettingsState) => {
    const nextValue = !settings[key];
    setSavingKey(key);
    setError("");
    try {
      const nextSettings = await SakikoService.UpdateAppSettings({
        [key]: nextValue,
      });
      setSettings({
        hideProfileNameInExport: nextSettings.hideProfileNameInExport,
        hideCNInboundInExport: nextSettings.hideCNInboundInExport,
      });
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : t("settings.privacy.saveFailed"));
    } finally {
      setSavingKey("");
    }
  };

  return (
    <SectionCard
      title={t("settings.privacy.title")}
      icon={<SecurityRounded color="primary" />}
      sx={{ height: "100%" }}
    >
      <Stack spacing={1.5}>
        {error ? <Alert severity="error">{error}</Alert> : null}

        <List disablePadding sx={{ mt: -0.5, mb: -0.5 }}>
          <ListItem
            sx={{
              px: 0,
              py: 1.5,
              gap: 2,
              borderBottom: "1px solid",
              borderColor: "divider",
            }}
          >
            <ListItemText
              sx={{ my: 0 }}
              primary={(
                <Typography variant="body1" fontWeight={600}>
                  {t("settings.privacy.hideProfileNameInExport")}
                </Typography>
              )}
            />
            <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", minWidth: 52 }}>
              {savingKey === "hideProfileNameInExport" ? (
                <CircularProgress size={18} />
              ) : (
                <Switch
                  checked={settings.hideProfileNameInExport}
                  disabled={loading || Boolean(savingKey)}
                  onChange={() => void toggleSetting("hideProfileNameInExport")}
                />
              )}
            </Box>
          </ListItem>

          <ListItem
            sx={{
              px: 0,
              py: 1.5,
              gap: 2,
            }}
          >
            <ListItemText
              sx={{ my: 0 }}
              primary={(
                <Typography variant="body1" fontWeight={600}>
                  {t("settings.privacy.hideCNInboundInExport")}
                </Typography>
              )}
            />
            <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", minWidth: 52 }}>
              {savingKey === "hideCNInboundInExport" ? (
                <CircularProgress size={18} />
              ) : (
                <Switch
                  checked={settings.hideCNInboundInExport}
                  disabled={loading || Boolean(savingKey)}
                  onChange={() => void toggleSetting("hideCNInboundInExport")}
                />
              )}
            </Box>
          </ListItem>
        </List>
      </Stack>
    </SectionCard>
  );
}
