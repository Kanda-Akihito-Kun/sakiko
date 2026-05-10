import DeleteOutlineRounded from "@mui/icons-material/DeleteOutlineRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import StorageRounded from "@mui/icons-material/StorageRounded";
import { Box, Card, IconButton, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import type { ProfileSummary } from "../../types/sakiko";
import { formatDataSize, formatDateTime, shouldUseEmojiFont } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileListPanelProps = {
  profiles: ProfileSummary[];
  activeProfileId: string;
  activeProfileName: string;
  submitting: boolean;
  canRefreshActiveProfile: boolean;
  onRefreshActiveProfile: () => Promise<void>;
  onDeleteActiveProfile: () => Promise<void>;
  onSelect: (profileId: string) => void;
};

export function ProfileListPanel({
  profiles,
  activeProfileId,
  activeProfileName,
  submitting,
  canRefreshActiveProfile,
  onRefreshActiveProfile,
  onDeleteActiveProfile,
  onSelect,
}: ProfileListPanelProps) {
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.profiles.list.title")}
      icon={<StorageRounded color="primary" />}
      action={activeProfileId ? (
        <Stack direction="row" spacing={0.5}>
          <IconButton
            size="small"
            color="primary"
            disabled={submitting || !canRefreshActiveProfile}
            onClick={() => void onRefreshActiveProfile()}
          >
            <RefreshRounded fontSize="small" />
          </IconButton>
          <IconButton
            size="small"
            color="error"
            disabled={submitting}
            onClick={() => {
              if (!window.confirm(t("dashboard.profiles.detail.deleteConfirm", { name: activeProfileName || t("shared.states.unnamedProfile") }))) {
                return;
              }
              void onDeleteActiveProfile();
            }}
          >
            <DeleteOutlineRounded fontSize="small" />
          </IconButton>
        </Stack>
      ) : undefined}
    >
      {profiles.length === 0 ? (
        <EmptyState
          title={t("dashboard.profiles.list.noProfilesTitle")}
          description={t("dashboard.profiles.list.noProfilesDescription")}
        />
      ) : (
        <Box
          sx={{
            display: "grid",
            gap: 1.5,
            gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
            alignItems: "stretch",
          }}
        >
          {profiles.map((profile) => {
            const selected = profile.id === activeProfileId;
            return (
              <Card
                key={profile.id}
                variant="outlined"
                role="button"
                tabIndex={0}
                onClick={() => void onSelect(profile.id)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    void onSelect(profile.id);
                  }
                }}
                sx={(theme) => ({
                  p: 1.75,
                  minWidth: 0,
                  cursor: "pointer",
                  borderRadius: 2,
                  borderColor: selected ? "transparent" : "divider",
                  bgcolor: selected ? "action.selected" : "background.paper",
                  transition: theme.transitions.create(["border-color", "background-color"], {
                    duration: theme.transitions.duration.shorter,
                  }),
                  "&:hover": {
                    borderColor: "primary.main",
                    backgroundColor: selected ? "action.selected" : "action.hover",
                  },
                  "&:focus-visible": {
                    outline: `2px solid ${theme.palette.primary.main}`,
                    outlineOffset: 2,
                  },
                })}
              >
                <Stack spacing={1.1} sx={{ minWidth: 0, height: "100%" }}>
                  <Stack sx={{ minWidth: 0 }}>
                    <Typography
                      variant="h6"
                      fontWeight={700}
                      noWrap
                      title={profile.name || t("shared.states.unnamedProfile")}
                      className={shouldUseEmojiFont("nodeName", profile.name) ? "sakiko-emoji" : undefined}
                      sx={{ minWidth: 0, flex: "1 1 auto" }}
                    >
                      {profile.name || t("shared.states.unnamedProfile")}
                    </Typography>
                  </Stack>

                  <Typography variant="body2" color="text.secondary" noWrap>
                    {t("shared.formats.nodeCount", { count: profile.nodeCount })}
                  </Typography>

                  {typeof profile.remainingBytes === "number" && profile.remainingBytes > 0 ? (
                    <Typography variant="body2" color="text.secondary" noWrap title={formatDataSize(profile.remainingBytes)}>
                      {t("dashboard.profiles.list.remainingTraffic", { defaultValue: "Remaining {{value}}", value: formatDataSize(profile.remainingBytes) })}
                    </Typography>
                  ) : null}

                  <Typography
                    variant="caption"
                    className="sakiko-mono"
                    color="text.secondary"
                    noWrap
                    title={profile.updatedAt || t("dashboard.profiles.list.noTimestamp")}
                  >
                    {profile.updatedAt ? formatDateTime(profile.updatedAt) : t("dashboard.profiles.list.noTimestamp")}
                  </Typography>
                </Stack>
              </Card>
            );
          })}
        </Box>
      )}
    </SectionCard>
  );
}
