import ImageRounded from "@mui/icons-material/ImageRounded";
import LanguageRounded from "@mui/icons-material/LanguageRounded";
import PaletteRounded from "@mui/icons-material/PaletteRounded";
import {
  Box,
  Card,
  List,
  ListItem,
  MenuItem,
  ListItemText,
  ListSubheader,
  Select,
  Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useI18n } from "../../hooks/useI18n";
import { useThemeMode } from "../../theme/themeMode";
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
    <Card variant="outlined">
      <List
        disablePadding
        subheader={(
          <ListSubheader
            disableSticky
            sx={{
              px: 2.25,
              py: 1.5,
              fontSize: 16,
              fontWeight: 700,
              color: "text.primary",
              bgcolor: "transparent",
              borderBottom: "1px solid",
              borderColor: "divider",
            }}
          >
            {t("settings.title")}
          </ListSubheader>
        )}
      >
        <ListItem
          sx={{
            px: 2.25,
            py: 2,
            alignItems: "center",
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
            sx={{ minWidth: 132 }}
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
            px: 2.25,
            py: 2,
            alignItems: "center",
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
          <ThemeModeSwitch value={mode} options={themeModeOptions} onChange={setMode} />
        </ListItem>

        <ListItem
          sx={{
            px: 2.25,
            py: 2,
            alignItems: "center",
            gap: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
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
          <ThemeModeSwitch value={exportPictureMode} options={exportModeOptions} onChange={setExportPictureMode} />
        </ListItem>

      </List>
    </Card>
  );
}
