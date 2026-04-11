import { Box, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";

type OverviewRowProps = {
  label: string;
  value: string;
  mono?: boolean;
  multiline?: boolean;
};

export function OverviewRow({ label, value, mono = false, multiline = false }: OverviewRowProps) {
  return (
    <Box
      sx={(theme) => ({
        display: "flex",
        justifyContent: "space-between",
        gap: 2,
        alignItems: multiline ? "flex-start" : "center",
        flexWrap: multiline ? "wrap" : "nowrap",
        p: 1.25,
        borderRadius: 2,
        backgroundColor: alpha(theme.palette.primary.main, 0.06),
      })}
    >
      <Typography variant="body2" color="text.secondary">
        {label}
      </Typography>
      <Typography
        variant="body2"
        className={mono ? "sakiko-mono" : undefined}
        noWrap={!multiline}
        sx={{
          minWidth: 0,
          flex: multiline ? "1 1 100%" : "0 1 auto",
          textAlign: multiline ? "left" : "right",
          whiteSpace: multiline ? "pre-wrap" : undefined,
          overflowWrap: multiline ? "anywhere" : undefined,
          wordBreak: multiline ? "break-word" : undefined,
        }}
      >
        {value}
      </Typography>
    </Box>
  );
}
