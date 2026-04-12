import { alpha, createTheme } from "@mui/material/styles";

export type AppThemeMode = "light" | "dark" | "system";
export type ResolvedThemeMode = Exclude<AppThemeMode, "system">;

const fontFamily = "var(--sakiko-font-sans)";

const cvrThemes = {
  light: {
    primary: "#007AFF",
    primaryDark: "#006ADC",
    secondary: "#FC9B76",
    success: "#06943D",
    error: "#FF3B30",
    warning: "#FF9500",
    textPrimary: "#000000",
    textSecondary: "#3C3C4399",
    pageBackground: "#ECECEC",
    workspaceBackground: "#F5F5F5",
    surface: "#FFFFFF",
    surfaceInset: "#FFFFFF",
    divider: "rgba(0, 0, 0, 0.06)",
    scrollbar: "#90939980",
    scrollbarTrack: "#F1F1F1",
    selectionColor: "#F5F5F5",
    sidebarBackground: "rgba(255, 255, 255, 0.92)",
    headerBackground: "rgba(245, 245, 245, 0.9)",
    logoBackground: alpha("#007AFF", 0.08),
    logoBorder: alpha("#007AFF", 0.14),
  },
  dark: {
    primary: "#0A84FF",
    primaryDark: "#0069D9",
    secondary: "#FF9F0A",
    success: "#30D158",
    error: "#FF453A",
    warning: "#FF9F0A",
    textPrimary: "#FFFFFF",
    textSecondary: "#EBEBF599",
    pageBackground: "#2E303D",
    workspaceBackground: "#1E1F27",
    surface: "#282A36",
    surfaceInset: "#242733",
    divider: "rgba(255, 255, 255, 0.06)",
    scrollbar: "#555555",
    scrollbarTrack: "#2E303D",
    selectionColor: "#3E3E3E",
    sidebarBackground: "rgba(30, 31, 39, 0.94)",
    headerBackground: "rgba(30, 31, 39, 0.88)",
    logoBackground: alpha("#0A84FF", 0.12),
    logoBorder: alpha("#0A84FF", 0.14),
  },
} as const;

export function createAppTheme(mode: ResolvedThemeMode) {
  const tokens = cvrThemes[mode];
  const isLight = mode === "light";
  const chipFill = alpha(tokens.primary, isLight ? 0.12 : 0.24);
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
      borderRadius: 4,
    },
    typography: {
      fontFamily,
      h1: {
        fontSize: "clamp(2rem, 3.1vw, 3rem)",
        lineHeight: 1.08,
        letterSpacing: "-0.04em",
        fontWeight: 700,
      },
      h4: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      h5: {
        fontWeight: 700,
      },
      h6: {
        fontWeight: 700,
        letterSpacing: "-0.01em",
      },
      subtitle1: {
        fontWeight: 700,
      },
      subtitle2: {
        letterSpacing: "0.06em",
        textTransform: "uppercase",
        fontSize: "0.72rem",
        color: tokens.textSecondary,
      },
      body2: {
        lineHeight: 1.55,
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
            "--background-color-alpha": alpha(tokens.primary, 0.1),
            "--divider-color": tokens.divider,
            "--sidebar-background": tokens.sidebarBackground,
            "--header-background": tokens.headerBackground,
            "--logo-background": tokens.logoBackground,
            "--logo-border-color": tokens.logoBorder,
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
            boxShadow: "none",
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
            minHeight: 33.375,
            borderRadius: 8,
            paddingInline: 14,
          },
          containedPrimary: {
            backgroundColor: tokens.primary,
            color: "#FFFFFF",
            "&:hover": {
              backgroundColor: tokens.primaryDark,
            },
          },
          outlined: {
            borderColor: tokens.divider,
            backgroundColor: isLight ? tokens.surface : alpha("#FFFFFF", 0.02),
            "&:hover": {
              borderColor: alpha(tokens.primary, 0.5),
              backgroundColor: hoverFill,
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
            borderRadius: 8,
            "& .MuiOutlinedInput-notchedOutline": {
              borderColor: tokens.divider,
            },
            "&:hover .MuiOutlinedInput-notchedOutline": {
              borderColor: alpha(tokens.primary, 0.36),
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
            borderRadius: 8,
            fontWeight: 600,
          },
          filledPrimary: {
            backgroundColor: chipFill,
            color: tokens.textPrimary,
          },
          outlined: {
            borderColor: tokens.divider,
            backgroundColor: isLight ? tokens.surface : alpha("#FFFFFF", 0.02),
          },
        },
      },
      MuiListItemButton: {
        styleOverrides: {
          root: {
            borderRadius: 8,
            border: "1px solid transparent",
            marginBottom: 6,
            "&:hover": {
              backgroundColor: hoverFill,
            },
            "&.Mui-selected": {
              backgroundColor: selectedFill,
              borderColor: alpha(tokens.primary, 0.2),
            },
            "&.Mui-selected:hover": {
              backgroundColor: selectedFill,
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
            backgroundColor: isLight ? tokens.surface : alpha("#FFFFFF", 0.02),
            "&:hover": {
              backgroundColor: hoverFill,
            },
            "&.Mui-selected": {
              color: tokens.textPrimary,
              backgroundColor: selectedFill,
            },
            "&.Mui-selected:hover": {
              backgroundColor: selectedFill,
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
            backgroundColor: isLight
              ? alpha("#FFFFFF", 0.9)
              : alpha(tokens.workspaceBackground, 0.72),
          },
        },
      },
      MuiLinearProgress: {
        styleOverrides: {
          root: {
            backgroundColor: isLight
              ? alpha(tokens.primary, 0.12)
              : alpha("#FFFFFF", 0.08),
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
