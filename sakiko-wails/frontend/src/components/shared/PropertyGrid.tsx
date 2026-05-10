import type { ReactNode } from "react";
import { alpha } from "@mui/material/styles";
import { Box, Chip, Stack, Typography } from "@mui/material";

type PropertyGridProps = {
  children: ReactNode;
  columns?: {
    xs: string;
    md?: string;
    lg?: string;
  };
};

type PropertyItemProps = {
  label: string;
  value: string;
  span?: 1 | 2;
  mono?: boolean;
  multiline?: boolean;
};

type StatusSummaryTileProps = {
  title: string;
  value: string;
  color: "default" | "success" | "warning";
};

export function PropertyGrid({
  children,
  columns = {
    xs: "minmax(0, 1fr)",
    md: "repeat(2, minmax(0, 1fr))",
  },
}: PropertyGridProps) {
  return (
    <Box
      sx={{
        display: "grid",
        gap: 1.25,
        gridTemplateColumns: columns,
        minWidth: 0,
      }}
    >
      {children}
    </Box>
  );
}

export function PropertyItem({
  label,
  value,
  span = 1,
  mono = true,
  multiline = false,
}: PropertyItemProps) {
  return (
    <Box
      sx={(theme) => ({
        display: "grid",
        gap: 0.4,
        gridColumn: span === 2 ? "1 / -1" : undefined,
        px: 1.25,
        py: 1,
        borderRadius: 2,
        backgroundColor: alpha(theme.palette.background.default, 0.65),
        minWidth: 0,
      })}
    >
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography
        variant="body1"
        className={mono ? "sakiko-mono" : undefined}
        sx={{
          whiteSpace: multiline ? "pre-wrap" : "nowrap",
          overflow: "hidden",
          textOverflow: multiline ? "clip" : "ellipsis",
          overflowWrap: multiline ? "anywhere" : undefined,
          minWidth: 0,
        }}
        title={!multiline ? value : undefined}
      >
        {value}
      </Typography>
    </Box>
  );
}

export function StatusSummaryTile({ title, value, color }: StatusSummaryTileProps) {
  return (
    <Box
      sx={(theme) => ({
        p: 1.5,
        borderRadius: 2,
        minWidth: 0,
        boxShadow: `0 8px 20px ${alpha(theme.palette.common.black, 0.04)}`,
        background:
          color === "success"
            ? `linear-gradient(135deg, ${theme.palette.success.light}22, ${theme.palette.success.main}10)`
            : color === "warning"
              ? `linear-gradient(135deg, ${theme.palette.warning.light}22, ${theme.palette.warning.main}10)`
              : `linear-gradient(135deg, ${theme.palette.primary.light}16, ${theme.palette.background.paper})`,
      })}
    >
      <Stack spacing={0.75}>
        <Typography variant="caption" color="text.secondary">
          {title}
        </Typography>
        <Chip label={value} color={color} size="small" sx={{ alignSelf: "flex-start", maxWidth: "100%" }} />
      </Stack>
    </Box>
  );
}
