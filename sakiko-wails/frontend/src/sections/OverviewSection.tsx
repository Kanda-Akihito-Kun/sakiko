import HomeRounded from "@mui/icons-material/HomeRounded";
import { Box, Stack } from "@mui/material";
import { DashboardHero } from "../components/layout/DashboardHero";
import { WorkspaceStatusPanel } from "../components/layout/WorkspaceStatusPanel";
import { ProfileListPanel } from "../components/profile/ProfileListPanel";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";
import type { ProfileSummary, TaskState, TaskStatusResponse } from "../types/sakiko";

type OverviewSectionProps = {
  activeProfileId: string;
  activeTask: TaskStatusResponse | null;
  error: string;
  loading: boolean;
  message: string;
  profiles: ProfileSummary[];
  profilesPath: string;
  tasks: TaskState[];
  onSelectProfile: (profileId: string) => Promise<void>;
};

export function OverviewSection({
  activeProfileId,
  activeTask,
  error,
  loading,
  message,
  profiles,
  profilesPath,
  tasks,
  onSelectProfile,
}: OverviewSectionProps) {
  return (
    <Stack spacing={2}>
      <DashboardHero
        activeTaskName={activeTask?.task?.name}
        loading={loading}
        profileCount={profiles.length}
        profilesPath={profilesPath}
        taskCount={tasks.length}
      />

      <Box className="sakiko-overview-grid">
        <ProfileListPanel
          profiles={profiles}
          activeProfileId={activeProfileId}
          onSelect={onSelectProfile}
        />

        <SectionCard
          title="Workspace Summary"
          subtitle="Current state snapshot"
          icon={<HomeRounded color="primary" />}
        >
          <Stack spacing={1.25}>
            <OverviewRow
              label="Profiles Path"
              value={profilesPath || "Loading..."}
              mono
              multiline
            />
            <OverviewRow
              label="Active Task"
              value={activeTask?.task?.name || "No active task selected"}
            />
            <OverviewRow
              label="Runtime Message"
              value={message}
            />
          </Stack>
        </SectionCard>
      </Box>

      <WorkspaceStatusPanel message={message} error={error} />
    </Stack>
  );
}
