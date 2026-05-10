import { alpha, createTheme } from "@mui/material/styles";
import type { Shadows } from "@mui/material/styles";

export type AppThemeMode = "light" | "dark" | "system";
export type ExportPictureMode = "follow-theme" | "light" | "dark";
export type ResolvedThemeMode = Exclude<AppThemeMode, "system">;

const fontFamily = "var(--sakiko-font-sans)";

const cvrThemes = {
  light: {
    primary: "#0061A4",
    primaryDark: "#00497D",
    primaryContainer: "#D1E4FF",
    primaryContainerHover: "#BDD8F6",
    onPrimaryContainer: "#001D36",
    secondary: "#535F70",
    success: "#1B7D3C",
    error: "#BA1A1A",
    warning: "#825500",
    textPrimary: "#191C20",
    textSecondary: "#43474E",
    pageBackground: "#F3F7FC",
    workspaceBackground: "#FDFCFF",
    surface: "#FDFCFF",
    surfaceInset: "#EEF3FA",
    surfaceContainer: "#F0F4FA",
    surfaceContainerHigh: "#E8EEF5",
    divider: "#C3C7CF",
    scrollbar: "rgba(67, 71, 78, 0.5)",
    scrollbarTrack: "#E8EEF5",
    selectionColor: "#FFFFFF",
    sidebarBackground: "rgba(253, 252, 255, 0.96)",
    headerBackground: "rgba(253, 252, 255, 0.94)",
    logoBackground: alpha("#0061A4", 0.1),
    logoBorder: alpha("#0061A4", 0.18),
  },
  dark: {
    primary: "#9ECAFF",
    primaryDark: "#77B5F5",
    primaryContainer: "#00497D",
    primaryContainerHover: "#0A5489",
    onPrimaryContainer: "#D1E4FF",
    secondary: "#BBC7DB",
    success: "#7DDA92",
    error: "#FFB4AB",
    warning: "#FFDDB0",
    textPrimary: "#E3E7EE",
    textSecondary: "#C3C7CF",
    pageBackground: "#101418",
    workspaceBackground: "#111820",
    surface: "#1A1C20",
    surfaceInset: "#20262D",
    surfaceContainer: "#1F252C",
    surfaceContainerHigh: "#2A3038",
    divider: "#44474E",
    scrollbar: "rgba(195, 199, 207, 0.42)",
    scrollbarTrack: "#151A20",
    selectionColor: "#003258",
    sidebarBackground: "rgba(26, 28, 32, 0.96)",
    headerBackground: "rgba(26, 28, 32, 0.94)",
    logoBackground: alpha("#9ECAFF", 0.1),
    logoBorder: alpha("#9ECAFF", 0.18),
  },
} as const;

export function createAppTheme(mode: ResolvedThemeMode) {
  const tokens = cvrThemes[mode];
  const isLight = mode === "light";
  const hoverFill = alpha(tokens.primary, isLight ? 0.08 : 0.12);
  const selectedFill = alpha(tokens.primary, isLight ? 0.14 : 0.28);

  return createTheme({
    palette: {
      mode,
      primary: {
        main: tokens.primary,
        dark: tokens.primaryDark,
        contrastText: isLight ? "#FFFFFF" : "#FFFFFF",
      },
      secondary: {
        main: tokens.secondary,
      },
      success: {
        main: tokens.success,
      },
      error: {
        main: tokens.error,
      },
      warning: {
        main: tokens.warning,
      },
      info: {
        main: tokens.primary,
      },
      background: {
        default: tokens.workspaceBackground,
        paper: tokens.surface,
      },
      text: {
        primary: tokens.textPrimary,
        secondary: tokens.textSecondary,
      },
      divider: tokens.divider,
      action: {
        hover: hoverFill,
        selected: selectedFill,
        focus: alpha(tokens.primary, isLight ? 0.16 : 0.2),
      },
    },
    shape: {
      borderRadius: 8,
    },
    shadows: Array(25).fill("none") as Shadows,
    typography: {
      fontFamily,
      h1: {
        fontSize: "clamp(2rem, 3.1vw, 3rem)",
        lineHeight: 1.08,
        letterSpacing: 0,
        fontWeight: 700,
      },
      h4: {
        fontWeight: 700,
        letterSpacing: 0,
      },
      h5: {
        fontWeight: 700,
        letterSpacing: 0,
      },
      h6: {
        fontWeight: 700,
        letterSpacing: 0,
      },
      subtitle1: {
        fontWeight: 700,
      },
      subtitle2: {
        letterSpacing: 0,
        textTransform: "uppercase",
        fontSize: "0.72rem",
        color: tokens.textSecondary,
      },
      body2: {
        lineHeight: 1.6,
      },
      button: {
        fontWeight: 600,
        textTransform: "none",
        letterSpacing: 0,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          ":root": {
            colorScheme: mode,
            "--primary-main": tokens.primary,
            "--selection-color": tokens.selectionColor,
            "--scroller-color": tokens.scrollbar,
            "--scrollbar-bg": tokens.scrollbarTrack,
            "--background-color": tokens.workspaceBackground,
            "--page-background-color": tokens.pageBackground,
            "--surface-color": tokens.surface,
            "--surface-inset-color": tokens.surfaceInset,
            "--surface-container-color": tokens.surfaceContainer,
            "--surface-container-high-color": tokens.surfaceContainerHigh,
            "--outline-variant-color": tokens.divider,
            "--background-color-alpha": alpha(tokens.primary, 0.1),
            "--divider-color": tokens.divider,
            "--sidebar-background": tokens.sidebarBackground,
            "--header-background": tokens.headerBackground,
            "--logo-background": tokens.logoBackground,
            "--logo-border-color": tokens.logoBorder,
            "--shadow-soft": mode === "light"
              ? "0 1px 2px rgba(23, 29, 28, 0.08)"
              : "0 1px 0 rgba(255, 255, 255, 0.04)",
            "--shadow-card": mode === "light"
              ? "0 1px 2px rgba(23, 29, 28, 0.08)"
              : "0 1px 0 rgba(255, 255, 255, 0.04)",
          },
          "::selection": {
            color: "var(--selection-color)",
            backgroundColor: "var(--primary-main)",
          },
          "*::-webkit-scrollbar": {
            width: "8px",
            height: "8px",
            backgroundColor: "var(--scrollbar-bg)",
          },
          "*::-webkit-scrollbar-thumb": {
            borderRadius: "4px",
            backgroundColor: "var(--scroller-color)",
          },
          "*::-webkit-scrollbar-corner": {
            backgroundColor: "transparent",
          },
          body: {
            minWidth: 320,
            height: "100vh",
            overflow: "hidden",
            backgroundColor: tokens.pageBackground,
            color: tokens.textPrimary,
          },
          "#root": {
            height: "100vh",
            minHeight: "100vh",
            overflow: "hidden",
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            boxShadow: "none",
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundColor: tokens.surface,
            backgroundImage: "none",
            border: `1px solid ${tokens.divider}`,
            borderRadius: 8,
            boxShadow: mode === "light" ? "var(--shadow-card)" : "none",
            color: tokens.textPrimary,
          },
        },
      },
      MuiButtonGroup: {
        styleOverrides: {
          grouped: {
            borderColor: tokens.divider,
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            minHeight: 36,
            borderRadius: 999,
            paddingInline: 18,
          },
          containedPrimary: {
            backgroundColor: tokens.primary,
            color: isLight ? "#FFFFFF" : "#003258",
            "&:hover": {
              backgroundColor: tokens.primaryDark,
            },
          },
          outlined: {
            borderColor: tokens.divider,
            backgroundColor: "transparent",
            "&:hover": {
              borderColor: tokens.primary,
              backgroundColor: alpha(tokens.primary, isLight ? 0.08 : 0.12),
            },
          },
          text: {
            "&:hover": {
              backgroundColor: alpha(tokens.primary, isLight ? 0.08 : 0.12),
            },
          },
        },
      },
      MuiTextField: {
        defaultProps: {
          size: "small",
          variant: "outlined",
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: tokens.surfaceInset,
            borderRadius: 12,
            "& .MuiOutlinedInput-notchedOutline": {
              borderColor: "transparent",
            },
            "&:hover .MuiOutlinedInput-notchedOutline": {
              borderColor: alpha(tokens.primary, 0.5),
            },
            "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
              borderColor: tokens.primary,
            },
          },
          input: {
            paddingBlock: 10,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderRadius: 999,
            fontWeight: 600,
          },
          filledPrimary: {
            backgroundColor: tokens.primaryContainer,
            color: tokens.onPrimaryContainer,
          },
          outlined: {
            borderColor: tokens.divider,
            backgroundColor: "transparent",
          },
        },
      },
      MuiListItemButton: {
        styleOverrides: {
          root: {
            borderRadius: 8,
            border: "1px solid transparent",
            marginBottom: 4,
            transition: "background-color 160ms ease, border-color 160ms ease",
            "&:hover": {
              backgroundColor: hoverFill,
            },
            "&.Mui-selected": {
              backgroundColor: tokens.primaryContainer,
              borderColor: "transparent",
            },
            "&.Mui-selected:hover": {
              backgroundColor: tokens.primaryContainerHover,
            },
          },
        },
      },
      MuiToggleButtonGroup: {
        styleOverrides: {
          grouped: {
            borderRadius: 8,
            borderColor: tokens.divider,
          },
        },
      },
      MuiToggleButton: {
        styleOverrides: {
          root: {
            paddingInline: 16,
            color: tokens.textSecondary,
            borderColor: tokens.divider,
            backgroundColor: tokens.surfaceInset,
            "&:hover": {
              backgroundColor: hoverFill,
            },
            "&.Mui-selected": {
              color: tokens.onPrimaryContainer,
              backgroundColor: tokens.primaryContainer,
            },
            "&.Mui-selected:hover": {
              backgroundColor: tokens.primaryContainerHover,
            },
          },
        },
      },
      MuiTableCell: {
        styleOverrides: {
          root: {
            borderBottom: `1px solid ${tokens.divider}`,
          },
          head: {
            color: tokens.textSecondary,
            backgroundColor: tokens.surfaceContainerHigh,
          },
        },
      },
      MuiTableRow: {
        styleOverrides: {
          root: {
            "&.MuiTableRow-hover:hover": {
              backgroundColor: alpha(tokens.primary, isLight ? 0.06 : 0.1),
            },
          },
        },
      },
      MuiLinearProgress: {
        styleOverrides: {
          root: {
            borderRadius: 999,
            backgroundColor: isLight
              ? alpha(tokens.primary, 0.12)
              : alpha("#FFFFFF", 0.08),
          },
        },
      },
      MuiSwitch: {
        styleOverrides: {
          root: {
            width: 46,
            height: 30,
            padding: 0,
          },
          switchBase: {
            padding: 4,
            "&.Mui-checked": {
              transform: "translateX(16px)",
              color: isLight ? "#FFFFFF" : "#003258",
              "& + .MuiSwitch-track": {
                backgroundColor: tokens.primary,
                opacity: 1,
              },
            },
          },
          thumb: {
            width: 22,
            height: 22,
            boxShadow: "none",
          },
          track: {
            borderRadius: 999,
            backgroundColor: tokens.surfaceContainerHigh,
            border: `1px solid ${tokens.divider}`,
            opacity: 1,
          },
        },
      },
      MuiTabs: {
        styleOverrides: {
          indicator: {
            height: 3,
            borderRadius: 999,
          },
        },
      },
      MuiTab: {
        styleOverrides: {
          root: {
            minHeight: 44,
            fontWeight: 700,
            letterSpacing: 0,
            textTransform: "none",
          },
        },
      },
      MuiAlert: {
        styleOverrides: {
          root: {
            borderRadius: 8,
          },
          filledError: {
            backgroundColor: isLight ? "#FFEFEE" : "#4B252A",
            color: isLight ? "#7A1F16" : "#FFFFFF",
          },
        },
      },
    },
  });
}
