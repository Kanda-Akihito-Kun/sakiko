import ErrorOutlineRounded from "@mui/icons-material/ErrorOutlineRounded";
import TerminalRounded from "@mui/icons-material/TerminalRounded";
import { Alert, Card, Chip, Stack, Typography } from "@mui/material";

type ConsoleFooterProps = {
  message: string;
  error: string;
};

export function ConsoleFooter({ message, error }: ConsoleFooterProps) {
  return (
    <Card sx={{ p: 2.25 }}>
      <Stack spacing={1.5}>
        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ minWidth: 0 }}>
          <Chip icon={<TerminalRounded />} label="Runtime Console" color="primary" variant="outlined" />
          <Typography variant="body2" color="text.secondary" noWrap sx={{ minWidth: 0 }}>
            {message}
          </Typography>
        </Stack>

        {error && (
          <Alert icon={<ErrorOutlineRounded fontSize="inherit" />} severity="error" variant="filled">
            {error}
          </Alert>
        )}
      </Stack>
    </Card>
  );
}
