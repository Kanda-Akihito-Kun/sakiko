import DeleteOutlineRounded from "@mui/icons-material/DeleteOutlineRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SearchRounded from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  Card,
  Chip,
  InputAdornment,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import type { Profile } from "../../types/sakiko";
import type { FilteredProfileNode } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileDetailPanelProps = {
  activeProfile: Profile | null;
  filteredNodes: FilteredProfileNode[];
  nodeFilter: string;
  submitting: boolean;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => void;
  onNodeFilterChange: (value: string) => void;
  onRefreshProfile: () => void;
  onDeleteProfile: () => void;
};

export function ProfileDetailPanel({
  activeProfile,
  filteredNodes,
  nodeFilter,
  submitting,
  onNodeEnabledChange,
  onNodeFilterChange,
  onRefreshProfile,
  onDeleteProfile,
}: ProfileDetailPanelProps) {
  const canRefreshProfile = Boolean(activeProfile?.source?.trim());
  const enabledNodeCount = activeProfile?.nodes.filter((node) => node.enabled).length ?? 0;

  return (
    <SectionCard
      title="Profile Detail"
      subtitle={activeProfile?.source || "Pick a profile"}
      icon={<HubRounded color="primary" />}
      subtitleWrap
    >
      {activeProfile ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(210px, 1fr))",
              gap: 1.5,
            }}
          >
            <MetricCard label="Selected" value={activeProfile.name} />
            <MetricCard label="Nodes" value={`${activeProfile.nodes.length}`} />
            <MetricCard label="Enabled" value={`${enabledNodeCount}`} />
            <MetricCard label="Updated" value={activeProfile.updatedAt || "Unknown"} mono />
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="flex-start" useFlexGap sx={{ flexWrap: "wrap" }}>
            <TextField
              value={nodeFilter}
              onChange={(event) => onNodeFilterChange(event.target.value)}
              placeholder="Filter conditions"
              sx={{ flex: "1 1 320px", minWidth: 240 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRounded sx={{ color: "text.secondary" }} />
                  </InputAdornment>
                ),
              }}
            />
            <Button
              variant="outlined"
              startIcon={<RefreshRounded />}
              disabled={submitting || !canRefreshProfile}
              onClick={onRefreshProfile}
              sx={{ minWidth: 210, flex: "0 0 auto" }}
            >
              Refresh Subscription
            </Button>
            <Button
              variant="outlined"
              color="error"
              startIcon={<DeleteOutlineRounded />}
              disabled={submitting}
              onClick={() => {
                if (!activeProfile) {
                  return;
                }
                if (!window.confirm(`Delete profile "${activeProfile.name || "Unnamed profile"}"?`)) {
                  return;
                }
                onDeleteProfile();
              }}
              sx={{ minWidth: 132, flex: "0 0 auto" }}
            >
              Delete
            </Button>
          </Stack>

          <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: "wrap", pb: 0.25 }}>
            <Chip label={`${filteredNodes.length} matching nodes`} color="primary" />
            <Chip label={`${enabledNodeCount} enabled for tasks`} color="success" variant="outlined" />
            <Chip label={activeProfile.source || "Recovered local profile"} variant="outlined" />
          </Stack>

          <TableContainer className="sakiko-table">
            <Table size="small" stickyHeader>
              <TableHead>
                <TableRow>
                  <TableCell>Node</TableCell>
                  <TableCell>Server</TableCell>
                  <TableCell align="right">Port</TableCell>
                  <TableCell align="right">Task Scope</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredNodes.map(({ node, index }) => {
                  return (
                    <TableRow key={`${node.name}-${index}`} hover>
                      <TableCell>
                        <Stack spacing={0.25}>
                          <Typography fontWeight={600} noWrap>{node.name}</Typography>
                          <Stack direction="row" spacing={0.75} sx={{ flexWrap: "wrap", pt: 0.25 }}>
                            <Chip
                              size="small"
                              label={node.protocol || "unknown"}
                              variant="outlined"
                              sx={{ textTransform: "uppercase" }}
                            />
                            <Chip
                              size="small"
                              label={node.udp === true ? "UDP" : node.udp === false ? "No UDP" : "UDP ?"}
                              color={node.udp === true ? "success" : "default"}
                              variant={node.udp === true ? "filled" : "outlined"}
                            />
                          </Stack>
                        </Stack>
                      </TableCell>
                      <TableCell className="sakiko-mono">{node.server || "unknown server"}</TableCell>
                      <TableCell align="right" className="sakiko-mono">
                        {node.port || "-"}
                      </TableCell>
                      <TableCell align="right">
                        <Button
                          size="small"
                          variant={node.enabled ? "contained" : "outlined"}
                          color={node.enabled ? "success" : "inherit"}
                          disabled={submitting}
                          onClick={() => onNodeEnabledChange(index, !node.enabled)}
                        >
                          {node.enabled ? "Included" : "Skipped"}
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
        </Stack>
      ) : (
        <EmptyState
          title="No profile selected"
          description="Choose one on the left to inspect nodes and launch tasks."
        />
      )}
    </SectionCard>
  );
}

type MetricCardProps = {
  label: string;
  value: string;
  mono?: boolean;
};

function MetricCard({ label, value, mono = false }: MetricCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        p: 1.75,
        bgcolor: "background.default",
        borderColor: "divider",
        minWidth: 0,
      }}
    >
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography
        className={mono ? "sakiko-mono" : undefined}
        fontWeight={600}
        sx={{
          overflowWrap: "anywhere",
          wordBreak: "break-word",
          whiteSpace: "pre-wrap",
        }}
      >
        {value}
      </Typography>
    </Card>
  );
}
