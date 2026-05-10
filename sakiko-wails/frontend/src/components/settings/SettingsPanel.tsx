import ImageRounded from "@mui/icons-material/ImageRounded";
import LanguageRounded from "@mui/icons-material/LanguageRounded";
import PaletteRounded from "@mui/icons-material/PaletteRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import {
  Box,
  List,
  ListItem,
  MenuItem,
  ListItemText,
  Select,
  Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useI18n } from "../../hooks/useI18n";
import { useThemeMode } from "../../theme/themeMode";
import { SectionCard } from "../shared/SectionCard";
import { appThemeModes, exportPictureModes, ThemeModeSwitch } from "./ThemeModeSwitch";

export function SettingsPanel() {
  const { t } = useTranslation();
  const { currentLanguage, isLoading, supportedLanguages, switchLanguage } = useI18n();
  const { mode, exportPictureMode, setMode, setExportPictureMode } = useThemeMode();
  const themeModeOptions = appThemeModes.map((option) => ({
    ...option,
    label: t(`settings.themeOptions.${option.value}`),
  }));
  const exportModeOptions = exportPictureModes.map((option) => ({
    ...option,
    label: option.value === "follow-theme"
      ? t("settings.themeOptions.followTheme")
      : t(`settings.themeOptions.${option.value}`),
  }));

  return (
    <SectionCard
      title={t("settings.title")}
      icon={<SettingsRounded color="primary" />}
      sx={{ height: "100%" }}
    >
      <List
        disablePadding
        sx={{ mt: -0.5, mb: -0.5 }}
      >
        <ListItem
          sx={{
            px: 0,
            py: 1.5,
            alignItems: { xs: "stretch", sm: "center" },
            flexDirection: { xs: "column", sm: "row" },
            gap: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <ListItemText
            sx={{ my: 0 }}
            primary={(
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <LanguageRounded fontSize="small" color="primary" />
                <Typography variant="body1" fontWeight={600}>
                  {t("settings.language.label")}
                </Typography>
              </Box>
            )}
          />
          <Select
            size="small"
            value={currentLanguage}
            disabled={isLoading}
            onChange={(event) => {
              void switchLanguage(String(event.target.value));
            }}
            sx={{ minWidth: { xs: 0, sm: 132 }, width: { xs: "100%", sm: "auto" } }}
          >
            {supportedLanguages.map((language) => (
              <MenuItem key={language} value={language}>
                {t(`settings.languageOptions.${language}`)}
              </MenuItem>
            ))}
          </Select>
        </ListItem>

        <ListItem
          sx={{
            px: 0,
            py: 1.5,
            alignItems: { xs: "stretch", sm: "center" },
            flexDirection: { xs: "column", sm: "row" },
            gap: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <ListItemText
            sx={{ my: 0 }}
            primary={(
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <PaletteRounded fontSize="small" color="primary" />
                <Typography variant="body1" fontWeight={600}>
                  {t("settings.themeMode.label")}
                </Typography>
              </Box>
            )}
          />
          <Box sx={{ width: { xs: "100%", sm: "auto" }, overflowX: "auto" }}>
            <ThemeModeSwitch value={mode} options={themeModeOptions} onChange={setMode} />
          </Box>
        </ListItem>

        <ListItem
          sx={{
            px: 0,
            py: 1.5,
            alignItems: { xs: "stretch", sm: "center" },
            flexDirection: { xs: "column", sm: "row" },
            gap: 2,
          }}
        >
          <ListItemText
            sx={{ my: 0 }}
            primary={(
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <ImageRounded fontSize="small" color="primary" />
                <Typography variant="body1" fontWeight={600}>
                  {t("settings.exportPictureMode.label")}
                </Typography>
              </Box>
            )}
          />
          <Box sx={{ width: { xs: "100%", sm: "auto" }, overflowX: "auto" }}>
            <ThemeModeSwitch value={exportPictureMode} options={exportModeOptions} onChange={setExportPictureMode} />
          </Box>
        </ListItem>

      </List>
    </SectionCard>
  );
}
