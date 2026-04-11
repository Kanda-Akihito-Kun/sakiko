import ErrorOutlineRounded from "@mui/icons-material/ErrorOutlineRounded";
import InfoOutlined from "@mui/icons-material/InfoOutlined";
import { Alert, Card, Chip, Stack, Typography } from "@mui/material";

type WorkspaceStatusPanelProps = {
  message: string;
  error: string;
};

export function WorkspaceStatusPanel({ message, error }: WorkspaceStatusPanelProps) {
  return (
    <Card sx={{ p: 2.25 }}>
      <Stack spacing={1.5}>
        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ minWidth: 0 }}>
          <Chip icon={<InfoOutlined />} label="Workspace Status" color="primary" variant="outlined" />
          <Typography variant="body2" color="text.secondary" noWrap sx={{ minWidth: 0 }}>
            {message}
          </Typography>
        </Stack>

        {error ? (
          <Alert icon={<ErrorOutlineRounded fontSize="inherit" />} severity="error" variant="filled">
            {error}
          </Alert>
        ) : null}
      </Stack>
    </Card>
  );
}
