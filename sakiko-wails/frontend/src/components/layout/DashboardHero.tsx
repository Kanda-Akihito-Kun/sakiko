import AutoAwesomeRounded from "@mui/icons-material/AutoAwesomeRounded";
import DnsRounded from "@mui/icons-material/DnsRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import PlayCircleOutlineRounded from "@mui/icons-material/PlayCircleOutlineRounded";
import { Box, Card, Chip, Divider, Stack, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";

type DashboardHeroProps = {
  activeTaskName?: string;
  loading: boolean;
  profileCount: number;
  profilesPath: string;
  taskCount: number;
};

export function DashboardHero({
  activeTaskName,
  loading,
  profileCount,
  profilesPath,
  taskCount,
}: DashboardHeroProps) {
  return (
    <Card variant="outlined" sx={{ p: { xs: 2.25, md: 2.75 } }}>
      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1.5fr) minmax(300px, 0.95fr)" },
          alignItems: "stretch",
        }}
      >
        <Stack spacing={2}>
          <Stack spacing={1.25}>
            <Chip
              icon={<AutoAwesomeRounded />}
              label="Workspace"
              color="primary"
              variant="outlined"
              sx={{ alignSelf: "flex-start" }}
            />
            <Typography variant="h4">sakiko</Typography>
            <Typography variant="body1" color="text.secondary" sx={{ maxWidth: 720 }}>
              Local profile management and task execution backed by{" "}
              <Box component="span" className="sakiko-mono">sakiko-core</Box>.
              Import subscriptions, inspect nodes, and launch test jobs from one workspace.
            </Typography>
          </Stack>

          <Stack
            direction="row"
            spacing={1.5}
            sx={{ flexWrap: "nowrap", overflowX: "auto", pb: 0.25 }}
          >
            <Chip
              icon={<DnsRounded />}
              label={`${profileCount} profile${profileCount === 1 ? "" : "s"}`}
              color="primary"
            />
            <Chip
              icon={<HubRounded />}
              label={`${taskCount} task${taskCount === 1 ? "" : "s"}`}
              color="secondary"
              variant="outlined"
            />
            <Chip
              icon={<PlayCircleOutlineRounded />}
              label={activeTaskName ? `Focused: ${activeTaskName}` : "No active task selected"}
              variant="outlined"
            />
          </Stack>
        </Stack>

        <Stack spacing={1.5}>
          <Card
            variant="outlined"
            sx={{
              p: 0,
              overflow: "hidden",
            }}
          >
            <Stack divider={<Divider flexItem sx={{ borderColor: "divider" }} />}>
              <WorkspaceRow
                label="Profile Store"
                value={profilesPath || "Loading..."}
                mono
              />
              <WorkspaceRow
                label="Runtime State"
                value={loading ? "Syncing" : "Ready"}
                accent={loading ? "warning" : "success"}
              />
              <WorkspaceRow
                label="Focused Task"
                value={activeTaskName || "No active task selected"}
              />
            </Stack>
          </Card>
        </Stack>
      </Box>
    </Card>
  );
}

type WorkspaceRowProps = {
  label: string;
  value: string;
  mono?: boolean;
  accent?: "warning" | "success";
};

function WorkspaceRow({ label, value, mono = false, accent }: WorkspaceRowProps) {
  return (
    <Box sx={{ px: 2, py: 1.5 }}>
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography
        variant="body2"
        className={mono ? "sakiko-mono" : undefined}
        noWrap
        sx={(theme) => ({
          color:
            accent === "warning"
              ? theme.palette.warning.main
              : accent === "success"
                ? theme.palette.success.main
                : theme.palette.text.primary,
          bgcolor: accent ? alpha(theme.palette[accent].main, 0.08) : "transparent",
          borderRadius: accent ? 1.5 : 0,
          display: "inline-flex",
          alignItems: "center",
          minHeight: accent ? 28 : "auto",
          px: accent ? 1 : 0,
          maxWidth: "100%",
        })}
      >
        {value}
      </Typography>
    </Box>
  );
}
