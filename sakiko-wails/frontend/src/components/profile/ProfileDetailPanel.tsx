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
import { extractNodeMeta } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileDetailPanelProps = {
  activeProfile: Profile | null;
  filteredNodes: Profile["nodes"];
  nodeFilter: string;
  submitting: boolean;
  onNodeFilterChange: (value: string) => void;
  onRefreshProfile: () => void;
  onDeleteProfile: () => void;
};

export function ProfileDetailPanel({
  activeProfile,
  filteredNodes,
  nodeFilter,
  submitting,
  onNodeFilterChange,
  onRefreshProfile,
  onDeleteProfile,
}: ProfileDetailPanelProps) {
  const canRefreshProfile = Boolean(activeProfile?.source?.trim());

  return (
    <SectionCard
      title="Profile Detail"
      subtitle={activeProfile?.source || "Pick a profile"}
      icon={<HubRounded color="primary" />}
    >
      {activeProfile ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0, 1fr))" },
              gap: 1.5,
            }}
          >
            <MetricCard label="Selected" value={activeProfile.name} />
            <MetricCard label="Nodes" value={`${activeProfile.nodes.length}`} />
            <MetricCard label="Updated" value={activeProfile.updatedAt || "Unknown"} mono />
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="flex-start">
            <TextField
              fullWidth
              value={nodeFilter}
              onChange={(event) => onNodeFilterChange(event.target.value)}
              placeholder="Filter node name / payload"
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

          <Stack direction="row" spacing={1} sx={{ flexWrap: "nowrap", overflowX: "auto", pb: 0.25 }}>
            <Chip label={`${filteredNodes.length} matching nodes`} color="primary" />
            <Chip label={activeProfile.source || "Recovered local profile"} variant="outlined" />
          </Stack>

          <TableContainer className="sakiko-table">
            <Table size="small" stickyHeader>
              <TableHead>
                <TableRow>
                  <TableCell>Node</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Server</TableCell>
                  <TableCell align="right">Port</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredNodes.map((node, index) => {
                  const meta = extractNodeMeta(node.payload);
                  return (
                    <TableRow key={`${node.name}-${index}`} hover>
                      <TableCell>
                        <Stack spacing={0.25}>
                          <Typography fontWeight={600} noWrap>{node.name}</Typography>
                          <Typography variant="caption" color="text.secondary" className="sakiko-mono" noWrap>
                            {node.payload.slice(0, 88)}
                            {node.payload.length > 88 ? "..." : ""}
                          </Typography>
                        </Stack>
                      </TableCell>
                      <TableCell>{meta.type}</TableCell>
                      <TableCell>{meta.server}</TableCell>
                      <TableCell align="right" className="sakiko-mono">
                        {meta.port}
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
      }}
    >
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography className={mono ? "sakiko-mono" : undefined} fontWeight={600} noWrap>
        {value}
      </Typography>
    </Card>
  );
}
