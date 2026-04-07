import { CssBaseline, ThemeProvider } from "@mui/material";
import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { AppThemeMode, createAppTheme, ResolvedThemeMode } from "./appTheme";

const STORAGE_KEY = "sakiko.theme-mode";

type ThemeModeContextValue = {
  mode: AppThemeMode;
  resolvedMode: ResolvedThemeMode;
  setMode: (mode: AppThemeMode) => void;
};

const ThemeModeContext = createContext<ThemeModeContextValue | null>(null);

function getSystemThemeMode(): ResolvedThemeMode {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function getStoredThemeMode(): AppThemeMode {
  const mode = window.localStorage.getItem(STORAGE_KEY);
  return mode === "light" || mode === "dark" || mode === "system" ? mode : "system";
}

export function AppThemeProvider({ children }: PropsWithChildren) {
  const [mode, setMode] = useState<AppThemeMode>(() => getStoredThemeMode());
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
    window.localStorage.setItem(STORAGE_KEY, mode);
  }, [mode]);

  const resolvedMode = mode === "system" ? systemMode : mode;
  const theme = useMemo(() => createAppTheme(resolvedMode), [resolvedMode]);
  const value = useMemo(
    () => ({
      mode,
      resolvedMode,
      setMode,
    }),
    [mode, resolvedMode],
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
