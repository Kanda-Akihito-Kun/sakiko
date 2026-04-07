import { Button, ButtonGroup } from "@mui/material";
import { AppThemeMode } from "../../theme/appTheme";

type ThemeModeSwitchProps = {
  value: AppThemeMode;
  onChange: (mode: AppThemeMode) => void;
};

const themeModes: Array<{ value: AppThemeMode; label: string }> = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
];

export function ThemeModeSwitch({ value, onChange }: ThemeModeSwitchProps) {
  return (
    <ButtonGroup size="small" sx={{ my: "4px" }}>
      {themeModes.map((mode) => (
        <Button
          key={mode.value}
          variant={mode.value === value ? "contained" : "outlined"}
          onClick={() => onChange(mode.value)}
          sx={{ textTransform: "capitalize", minWidth: 72 }}
        >
          {mode.label}
        </Button>
      ))}
    </ButtonGroup>
  );
}
