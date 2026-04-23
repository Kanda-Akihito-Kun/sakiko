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
        display: "grid",
        gridTemplateColumns: multiline ? "minmax(0, 1fr)" : "minmax(104px, auto) minmax(0, 1fr)",
        gap: 2,
        alignItems: multiline ? "flex-start" : "center",
        p: 1.25,
        borderRadius: 2,
        backgroundColor: alpha(theme.palette.primary.main, 0.06),
        minWidth: 0,
      })}
    >
      <Typography variant="body2" color="text.secondary" sx={{ minWidth: 0 }}>
        {label}
      </Typography>
      <Typography
        variant="body2"
        className={mono ? "sakiko-mono" : undefined}
        noWrap={!multiline}
        title={!multiline ? value : undefined}
        sx={{
          minWidth: 0,
          textAlign: multiline ? "left" : "right",
          justifySelf: multiline ? "stretch" : "end",
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
