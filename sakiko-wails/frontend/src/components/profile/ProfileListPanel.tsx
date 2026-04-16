import StorageRounded from "@mui/icons-material/StorageRounded";
import { Box, Chip, List, ListItemButton, ListItemText, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.profiles.list.title")}
      subtitle={t("dashboard.profiles.list.loaded", { count: profiles.length })}
      icon={<StorageRounded color="primary" />}
    >
      <List disablePadding>
        {profiles.map((profile) => (
          <ListItemButton
            key={profile.id}
            selected={profile.id === activeProfileId}
            onClick={() => void onSelect(profile.id)}
            sx={{ alignItems: "flex-start", px: 1.75, py: 1.5, minWidth: 0 }}
          >
            <ListItemText
              primary={profile.name || t("shared.states.unnamedProfile")}
              secondary={
                <Box sx={{ display: "grid", gap: 0.75, mt: 0.75 }}>
                  <Typography variant="body2" color="text.secondary" noWrap>
                    {t("shared.formats.nodeCount", { count: profile.nodeCount })}
                  </Typography>
                  <Typography
                    variant="caption"
                    className="sakiko-mono"
                    color="text.secondary"
                    noWrap
                    title={profile.updatedAt || t("dashboard.profiles.list.noTimestamp")}
                  >
                    {profile.updatedAt || t("dashboard.profiles.list.noTimestamp")}
                  </Typography>
                </Box>
              }
              primaryTypographyProps={{
                fontWeight: 600,
                noWrap: true,
                title: profile.name || t("shared.states.unnamedProfile"),
                className: shouldUseEmojiFont("nodeName", profile.name) ? "sakiko-emoji" : undefined,
              }}
              sx={{ minWidth: 0, mr: 1 }}
            />
            <Chip label={profile.id.slice(0, 8)} size="small" variant="outlined" sx={{ flexShrink: 0 }} />
          </ListItemButton>
        ))}

        {profiles.length === 0 && (
          <EmptyState
            title={t("dashboard.profiles.list.noProfilesTitle")}
            description={t("dashboard.profiles.list.noProfilesDescription")}
          />
        )}
      </List>
    </SectionCard>
  );
}
