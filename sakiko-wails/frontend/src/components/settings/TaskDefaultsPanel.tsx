import RefreshRounded from "@mui/icons-material/RefreshRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import {
  Box,
  Button,
  Chip,
  List,
  ListItemButton,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import { DownloadTargetSource } from "../../types/sakiko";
import type { DownloadTarget, TaskConfig } from "../../types/sakiko";
import { SectionCard } from "../shared/SectionCard";

type TaskConfigPatch = Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading">>;

type TaskDefaultsPanelProps = {
  downloadTargets: DownloadTarget[];
  downloadTargetsLoading: boolean;
  taskConfig: TaskConfig;
  onPatchTaskConfig: (patch: TaskConfigPatch) => void;
  onRefreshDownloadTargets: () => void;
};

export function TaskDefaultsPanel({
  downloadTargets,
  downloadTargetsLoading,
  taskConfig,
  onPatchTaskConfig,
  onRefreshDownloadTargets,
}: TaskDefaultsPanelProps) {
  const defaultTarget = downloadTargets.find((target) => target.source === DownloadTargetSource.DownloadTargetSourceCloudflare) || null;
  const speedtestTargets = downloadTargets.filter((target) => target.source === DownloadTargetSource.DownloadTargetSourceSpeedtest);

  return (
    <SectionCard
      title="Task Defaults"
      subtitle="Shared configuration applied when tasks are submitted"
      icon={<TuneRounded color="primary" />}
    >
      <Stack spacing={2}>
        <Typography variant="body2" color="text.secondary">
          Keep launcher tabs focused on task type selection. Adjust shared runtime defaults here instead.
        </Typography>

        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
            gap: 1.5,
          }}
        >
          <TextField
            fullWidth
            label="Ping URL"
            value={taskConfig.pingAddress}
            onChange={(event) => onPatchTaskConfig({ pingAddress: event.target.value })}
            helperText="Used by ping and full tasks"
          />
          <TextField
            fullWidth
            label="Timeout (ms)"
            type="number"
            value={taskConfig.taskTimeoutMillis}
            onChange={(event) => onPatchTaskConfig({ taskTimeoutMillis: Number(event.target.value) })}
          />
          <TextField
            fullWidth
            label="Download URL"
            value={taskConfig.downloadURL}
            onChange={(event) => onPatchTaskConfig({ downloadURL: event.target.value })}
            helperText="Manual override for speed and full tasks"
          />
          <TextField
            fullWidth
            label="Duration (s)"
            type="number"
            value={taskConfig.downloadDuration}
            onChange={(event) => onPatchTaskConfig({ downloadDuration: Number(event.target.value) })}
            inputProps={{ min: 5, max: 20 }}
            helperText="Used by speed and full tasks. Range: 5-20s"
          />
        </Box>

        <Stack spacing={1.25}>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="subtitle2" color="text.secondary">
                Download Target
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Choose a Speedtest target to fill the download URL, or keep the Cloudflare default.
              </Typography>
            </Box>
            <Button
              size="small"
              variant="outlined"
              startIcon={<RefreshRounded />}
              disabled={downloadTargetsLoading}
              onClick={onRefreshDownloadTargets}
            >
              Refresh
            </Button>
          </Stack>

          {defaultTarget ? (
            <Button
              variant={taskConfig.downloadURL === defaultTarget.downloadURL ? "contained" : "outlined"}
              onClick={() => onPatchTaskConfig({ downloadURL: defaultTarget.downloadURL })}
              sx={{ justifyContent: "flex-start" }}
            >
              Use {defaultTarget.name}
            </Button>
          ) : null}

          <List
            disablePadding
            sx={{
              maxHeight: 320,
              overflowY: "auto",
              borderRadius: 2,
              border: (theme) => `1px solid ${theme.palette.divider}`,
            }}
          >
            {speedtestTargets.length === 0 ? (
              <Box px={1.5} py={2}>
                <Typography variant="body2" color="text.secondary">
                  {downloadTargetsLoading ? "Loading Speedtest targets..." : "No Speedtest targets loaded yet."}
                </Typography>
              </Box>
            ) : (
              speedtestTargets.map((target) => (
                <ListItemButton
                  key={target.id}
                  selected={taskConfig.downloadURL === target.downloadURL}
                  onClick={() => onPatchTaskConfig({ downloadURL: target.downloadURL })}
                  sx={{ alignItems: "flex-start", flexDirection: "column", gap: 0.5 }}
                >
                  <Stack width="100%" direction="row" justifyContent="space-between" spacing={1}>
                    <Typography variant="body2" fontWeight={600} noWrap sx={{ minWidth: 0 }}>
                      {target.city || target.name}
                      {target.country ? `, ${target.country}` : ""}
                    </Typography>
                    {taskConfig.downloadURL === target.downloadURL ? (
                      <Chip size="small" color="primary" label="Selected" />
                    ) : null}
                  </Stack>
                  <Typography variant="caption" color="text.secondary" noWrap sx={{ width: "100%" }}>
                    {target.sponsor || "Unknown sponsor"}
                  </Typography>
                  <Typography variant="caption" className="sakiko-mono" color="text.secondary" noWrap sx={{ width: "100%" }}>
                    {target.host || target.endpoint}
                  </Typography>
                </ListItemButton>
              ))
            )}
          </List>
        </Stack>

        <Stack spacing={1}>
          <Typography variant="subtitle2" color="text.secondary">
            Speed Mode
          </Typography>
          <ToggleButtonGroup
            exclusive
            fullWidth
            value={taskConfig.downloadThreading > 1 ? "multi" : "single"}
            onChange={(_event, value: "single" | "multi" | null) => {
              if (value === "single") {
                onPatchTaskConfig({ downloadThreading: 1 });
              }
              if (value === "multi") {
                onPatchTaskConfig({ downloadThreading: 8 });
              }
            }}
          >
            <ToggleButton value="single">Single-thread</ToggleButton>
            <ToggleButton value="multi">Multi-thread</ToggleButton>
          </ToggleButtonGroup>
          <Typography variant="caption" color="text.secondary">
            {taskConfig.downloadThreading > 1
              ? `Using ${taskConfig.downloadThreading} parallel download streams`
              : "Using 1 download stream"}
          </Typography>
        </Stack>
      </Stack>
    </SectionCard>
  );
}
