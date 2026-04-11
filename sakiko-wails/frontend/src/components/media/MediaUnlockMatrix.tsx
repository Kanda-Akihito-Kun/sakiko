import { Box, Chip, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@mui/material";
import { alpha, type Theme } from "@mui/material/styles";
import type { MediaMatrixColumn, MediaMatrixRow } from "../../utils/mediaMatrix";
import { mediaCellTone } from "../../utils/mediaMatrix";

type MediaUnlockMatrixProps = {
  columns: MediaMatrixColumn[];
  rows: MediaMatrixRow[];
  title?: string;
  subtitle?: string;
};

export function MediaUnlockMatrix({ columns, rows, title, subtitle }: MediaUnlockMatrixProps) {
  if (columns.length === 0 || rows.length === 0) {
    return null;
  }

  return (
    <Stack spacing={1.25}>
      {(title || subtitle) ? (
        <Box>
          {title ? <Typography variant="subtitle1">{title}</Typography> : null}
          {subtitle ? (
            <Typography variant="body2" color="text.secondary">
              {subtitle}
            </Typography>
          ) : null}
        </Box>
      ) : null}

      <Box
        className="sakiko-table"
        sx={(theme) => ({
          borderRadius: 2,
          border: `1px solid ${theme.palette.divider}`,
          overflow: "auto",
          bgcolor: "background.paper",
        })}
      >
        <Table size="small" stickyHeader sx={{ minWidth: 880 }}>
          <TableHead>
            <TableRow>
              <TableCell sx={{ minWidth: 220 }}>Node</TableCell>
              <TableCell sx={{ minWidth: 120 }}>Protocol</TableCell>
              {columns.map((column) => (
                <TableCell key={column.key} align="center" sx={{ minWidth: 136 }}>
                  {column.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row, rowIndex) => (
              <TableRow key={`${row.nodeName}-${row.proxyType}-${rowIndex}`} hover>
                <TableCell>
                  <Typography variant="body2" fontWeight={600}>
                    {row.nodeName}
                  </Typography>
                </TableCell>
                <TableCell className="sakiko-mono">
                  {row.proxyType}
                </TableCell>
                {columns.map((column) => {
                  const value = row.cells[column.key] || "-";
                  const tone = mediaCellTone(value);
                  return (
                    <TableCell key={column.key} align="center">
                      <Chip
                        size="small"
                        label={value}
                        sx={(theme) => ({
                          width: "100%",
                          maxWidth: 180,
                          borderRadius: 1.5,
                          fontWeight: 600,
                          color: resolveToneColor(theme, tone),
                          bgcolor: alpha(resolveToneBase(theme, tone), 0.16),
                          "& .MuiChip-label": {
                            display: "block",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          },
                        })}
                        title={value}
                      />
                    </TableCell>
                  );
                })}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Stack>
  );
}

function resolveToneBase(theme: Theme, tone: "success" | "warning" | "error" | "neutral"): string {
  switch (tone) {
    case "success":
      return theme.palette.success.main;
    case "warning":
      return theme.palette.warning.main;
    case "error":
      return theme.palette.error.main;
    default:
      return theme.palette.text.secondary;
  }
}

function resolveToneColor(theme: Theme, tone: "success" | "warning" | "error" | "neutral"): string {
  switch (tone) {
    case "success":
      return theme.palette.success.dark;
    case "warning":
      return theme.palette.warning.dark;
    case "error":
      return theme.palette.error.dark;
    default:
      return theme.palette.text.secondary;
  }
}
