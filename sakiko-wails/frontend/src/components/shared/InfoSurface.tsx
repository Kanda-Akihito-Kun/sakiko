import type { ReactNode } from "react";
import type { SxProps, Theme } from "@mui/material/styles";
import { alpha } from "@mui/material/styles";
import { Box, Stack, Typography } from "@mui/material";

type InfoSurfaceProps = {
  title: string;
  icon?: ReactNode;
  action?: ReactNode;
  children: ReactNode;
  sx?: SxProps<Theme>;
};

export function InfoSurface({ title, icon, action, children, sx }: InfoSurfaceProps) {
  return (
    <Box
      sx={[
        (theme) => ({
          border: `1px solid ${theme.palette.divider}`,
          borderRadius: 2,
          overflow: "hidden",
          minWidth: 0,
          background: "var(--surface-container-color)",
        }),
        ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
      ]}
    >
      <Box
        sx={(theme) => ({
          display: "grid",
          gridTemplateColumns: "auto minmax(0, 1fr) auto",
          alignItems: "center",
          gap: 1.25,
          px: 2,
          py: 1.5,
          backgroundColor: "var(--surface-container-high-color)",
          borderBottom: `1px solid ${theme.palette.divider}`,
        })}
      >
        {icon ? (
          <Box
            sx={(theme) => ({
              width: 40,
              height: 40,
              display: "grid",
              placeItems: "center",
              borderRadius: 999,
              backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === "light" ? 0.12 : 0.18),
              border: "none",
            })}
          >
            {icon}
          </Box>
        ) : null}

        <Typography variant="h6" fontWeight={700} noWrap>
          {title}
        </Typography>

        {action ? <Box sx={{ justifySelf: "end" }}>{action}</Box> : null}
      </Box>

      <Stack spacing={1.5} sx={{ p: 2, minWidth: 0 }}>
        {children}
      </Stack>
    </Box>
  );
}
