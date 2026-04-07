import { Box, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";

type OverviewRowProps = {
  label: string;
  value: string;
  mono?: boolean;
};

export function OverviewRow({ label, value, mono = false }: OverviewRowProps) {
  return (
    <Box
      sx={(theme) => ({
        display: "flex",
        justifyContent: "space-between",
        gap: 2,
        alignItems: "center",
        p: 1.25,
        borderRadius: 2,
        backgroundColor: alpha(theme.palette.primary.main, 0.06),
      })}
    >
      <Typography variant="body2" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="body2" className={mono ? "sakiko-mono" : undefined} noWrap sx={{ minWidth: 0, textAlign: "right" }}>
        {value}
      </Typography>
    </Box>
  );
}
