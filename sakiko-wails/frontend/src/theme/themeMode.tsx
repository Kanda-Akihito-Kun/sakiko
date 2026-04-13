import { CssBaseline, ThemeProvider } from "@mui/material";
import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { AppThemeMode, createAppTheme, ExportPictureMode, ResolvedThemeMode } from "./appTheme";

const THEME_MODE_STORAGE_KEY = "sakiko.theme-mode";
const EXPORT_PICTURE_MODE_STORAGE_KEY = "sakiko.export-picture-mode";

type ThemeModeContextValue = {
  mode: AppThemeMode;
  resolvedMode: ResolvedThemeMode;
  exportPictureMode: ExportPictureMode;
  resolvedExportPictureMode: ResolvedThemeMode;
  setMode: (mode: AppThemeMode) => void;
  setExportPictureMode: (mode: ExportPictureMode) => void;
};

const ThemeModeContext = createContext<ThemeModeContextValue | null>(null);

function getSystemThemeMode(): ResolvedThemeMode {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function getStoredThemeMode(): AppThemeMode {
  const mode = window.localStorage.getItem(THEME_MODE_STORAGE_KEY);
  return mode === "light" || mode === "dark" || mode === "system" ? mode : "system";
}

function getStoredExportPictureMode(): ExportPictureMode {
  const mode = window.localStorage.getItem(EXPORT_PICTURE_MODE_STORAGE_KEY);
  return mode === "follow-theme" || mode === "light" || mode === "dark" ? mode : "follow-theme";
}

export function AppThemeProvider({ children }: PropsWithChildren) {
  const [mode, setMode] = useState<AppThemeMode>(() => getStoredThemeMode());
  const [exportPictureMode, setExportPictureMode] = useState<ExportPictureMode>(() => getStoredExportPictureMode());
  const [systemMode, setSystemMode] = useState<ResolvedThemeMode>(() => getSystemThemeMode());

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (event: MediaQueryListEvent | MediaQueryList) => {
      setSystemMode(event.matches ? "dark" : "light");
    };

    handleChange(mediaQuery);

    if (typeof mediaQuery.addEventListener === "function") {
      mediaQuery.addEventListener("change", handleChange);
      return () => mediaQuery.removeEventListener("change", handleChange);
    }

    mediaQuery.addListener(handleChange);
    return () => mediaQuery.removeListener(handleChange);
  }, []);

  useEffect(() => {
    window.localStorage.setItem(THEME_MODE_STORAGE_KEY, mode);
  }, [mode]);

  useEffect(() => {
    window.localStorage.setItem(EXPORT_PICTURE_MODE_STORAGE_KEY, exportPictureMode);
  }, [exportPictureMode]);

  const resolvedMode = mode === "system" ? systemMode : mode;
  const resolvedExportPictureMode = exportPictureMode === "follow-theme" ? resolvedMode : exportPictureMode;
  const theme = useMemo(() => createAppTheme(resolvedMode), [resolvedMode]);
  const value = useMemo(
    () => ({
      mode,
      resolvedMode,
      exportPictureMode,
      resolvedExportPictureMode,
      setMode,
      setExportPictureMode,
    }),
    [exportPictureMode, mode, resolvedExportPictureMode, resolvedMode],
  );

  return (
    <ThemeModeContext.Provider value={value}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  );
}

export function useThemeMode() {
  const context = useContext(ThemeModeContext);

  if (!context) {
    throw new Error("useThemeMode must be used within AppThemeProvider");
  }

  return context;
}
