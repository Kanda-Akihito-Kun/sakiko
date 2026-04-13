import { Button, ButtonGroup } from "@mui/material";
import type { AppThemeMode, ExportPictureMode } from "../../theme/appTheme";

type ModeOption<T extends string> = {
  value: T;
  label: string;
};

type ThemeModeSwitchProps<T extends string> = {
  value: T;
  options: Array<ModeOption<T>>;
  onChange: (mode: T) => void;
};

const themeModes: Array<ModeOption<AppThemeMode>> = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
];

export const exportPictureModes: Array<ModeOption<ExportPictureMode>> = [
  { value: "follow-theme", label: "Follow Theme" },
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
];

export function ThemeModeSwitch<T extends string>({ value, options, onChange }: ThemeModeSwitchProps<T>) {
  return (
    <ButtonGroup size="small" sx={{ my: "4px" }}>
      {options.map((mode) => (
        <Button
          key={mode.value}
          variant={mode.value === value ? "contained" : "outlined"}
          onClick={() => onChange(mode.value)}
          sx={{ textTransform: "none", minWidth: 72 }}
        >
          {mode.label}
        </Button>
      ))}
    </ButtonGroup>
  );
}

export const appThemeModes = themeModes;
