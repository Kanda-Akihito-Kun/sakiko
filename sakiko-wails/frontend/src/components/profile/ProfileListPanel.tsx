import StorageRounded from "@mui/icons-material/StorageRounded";
import { Box, Chip, List, ListItemButton, ListItemText, Typography } from "@mui/material";
import type { ProfileSummary } from "../../types/sakiko";
import { shouldUseEmojiFont } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileListPanelProps = {
  profiles: ProfileSummary[];
  activeProfileId: string;
  onSelect: (profileId: string) => void;
};

export function ProfileListPanel({ profiles, activeProfileId, onSelect }: ProfileListPanelProps) {
  return (
    <SectionCard
      title="Profiles"
      subtitle={`${profiles.length} loaded`}
      icon={<StorageRounded color="primary" />}
    >
      <List disablePadding>
        {profiles.map((profile) => (
          <ListItemButton
            key={profile.id}
            selected={profile.id === activeProfileId}
            onClick={() => void onSelect(profile.id)}
            sx={{ alignItems: "flex-start", px: 1.75, py: 1.5 }}
          >
            <ListItemText
              primary={profile.name || "Unnamed profile"}
              secondary={
                <Box sx={{ display: "grid", gap: 0.75, mt: 0.75 }}>
                  <Typography variant="body2" color="text.secondary" noWrap>
                    {profile.nodeCount} nodes
                  </Typography>
                  <Typography variant="caption" className="sakiko-mono" color="text.secondary" noWrap>
                    {profile.updatedAt || "No timestamp"}
                  </Typography>
                </Box>
              }
              primaryTypographyProps={{ fontWeight: 600, noWrap: true, className: shouldUseEmojiFont("nodeName", profile.name) ? "sakiko-emoji" : undefined }}
              sx={{ minWidth: 0, mr: 1 }}
            />
            <Chip label={profile.id.slice(0, 8)} size="small" variant="outlined" />
          </ListItemButton>
        ))}

        {profiles.length === 0 && (
          <EmptyState
            title="No profiles yet"
            description="Import a subscription to start the desktop flow."
          />
        )}
      </List>
    </SectionCard>
  );
}
