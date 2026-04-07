import StorageRounded from "@mui/icons-material/StorageRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import { Box, Stack } from "@mui/material";
import { ConsoleFooter } from "../components/layout/ConsoleFooter";
import { EmptyState } from "../components/shared/EmptyState";
import { OverviewRow } from "../components/shared/OverviewRow";
import { SectionCard } from "../components/shared/SectionCard";
import type { ProfileSummary, TaskStatusResponse } from "../types/sakiko";

type ConsoleSectionProps = {
  activeTask: TaskStatusResponse | null;
  error: string;
  message: string;
  profiles: ProfileSummary[];
};

export function ConsoleSection({ activeTask, error, message, profiles }: ConsoleSectionProps) {
  return (
    <Stack spacing={2}>
      <ConsoleFooter message={message} error={error} />

      {(activeTask || profiles.length > 0) ? (
        <Box className="sakiko-overview-grid">
          <SectionCard
            title="Recent Profiles"
            subtitle="Current workspace inventory"
            icon={<StorageRounded color="primary" />}
          >
            <Stack spacing={1}>
              {profiles.slice(0, 4).map((profile) => (
                <OverviewRow
                  key={profile.id}
                  label={profile.name || "Unnamed profile"}
                  value={`${profile.nodeCount} nodes`}
                />
              ))}
            </Stack>
          </SectionCard>

          <SectionCard
            title="Task State"
            subtitle="Most recent task selection"
            icon={<TuneRounded color="primary" />}
          >
            {activeTask ? (
              <Stack spacing={1}>
                <OverviewRow label="Task" value={activeTask.task.name} />
                <OverviewRow label="Status" value={activeTask.task.status} />
                <OverviewRow
                  label="Progress"
                  value={`${activeTask.task.progress}/${activeTask.task.total}`}
                  mono
                />
              </Stack>
            ) : (
              <EmptyState
                title="No task selected"
                description="Run a task or inspect one from the task list to populate this view."
              />
            )}
          </SectionCard>
        </Box>
      ) : (
        <EmptyState
          title="Workspace is empty"
          description="Import a profile first, then this console view will start reflecting runtime state."
        />
      )}
    </Stack>
  );
}
