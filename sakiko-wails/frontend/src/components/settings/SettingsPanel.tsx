import MonitorRounded from "@mui/icons-material/MonitorRounded";
import PaletteRounded from "@mui/icons-material/PaletteRounded";
import {
  Box,
  Card,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListSubheader,
  Stack,
  Typography,
} from "@mui/material";
import { useThemeMode } from "../../theme/themeMode";
import { ThemeModeSwitch } from "./ThemeModeSwitch";

export function SettingsPanel() {
  const { mode, resolvedMode, setMode } = useThemeMode();

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
            Basic
          </ListSubheader>
        )}
      >
        <ListItem
          sx={{
            px: 2.25,
            py: 2,
            alignItems: "flex-start",
            gap: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <ListItemText
            primary={(
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <PaletteRounded fontSize="small" color="primary" />
                <Typography variant="body1" fontWeight={600}>
                  Theme Mode
                </Typography>
              </Box>
            )}
            secondary="Choose the appearance mode."
          />
          <ThemeModeSwitch value={mode} onChange={setMode} />
        </ListItem>

        <ListItem
          sx={{
            px: 2.25,
            py: 1.75,
            alignItems: "center",
            gap: 2,
          }}
        >
          <ListItemText
            primary={(
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <MonitorRounded fontSize="small" color="primary" />
                <Typography variant="body1" fontWeight={600}>
                  Applied Theme
                </Typography>
              </Box>
            )}
            secondary={
              mode === "system"
                ? "The interface follows the current operating system appearance."
                : "The interface uses the selected fixed appearance."
            }
          />
          <Stack direction="row" spacing={1} alignItems="center">
            <Chip
              size="small"
              color="primary"
              label={resolvedMode === "light" ? "Light Active" : "Dark Active"}
            />
          </Stack>
        </ListItem>
      </List>
    </Card>
  );
}
