import { Box } from "@mui/material";
import { TaskLauncherPanel } from "../components/task/TaskLauncherPanel";
import { TaskResultsPanel } from "../components/task/TaskResultsPanel";
import type { TaskPreset } from "../types/dashboard";
import type { TaskConfig, TaskState, TaskStatusResponse } from "../types/sakiko";

type TasksSectionProps = {
  activeProfileId: string;
  activeTask: TaskStatusResponse | null;
  activeTaskId: string;
  submitting: boolean;
  taskConfig: TaskConfig;
  taskPreset: TaskPreset;
  tasks: TaskState[];
  onInspectTask: (taskId: string) => Promise<void>;
  onOpenSettings: () => void;
  onRunTask: () => Promise<void>;
  onTaskPresetChange: (value: TaskPreset) => void;
};

export function TasksSection({
  activeProfileId,
  activeTask,
  activeTaskId,
  submitting,
  taskConfig,
  taskPreset,
  tasks,
  onInspectTask,
  onOpenSettings,
  onRunTask,
  onTaskPresetChange,
}: TasksSectionProps) {
  return (
    <Box
      sx={{
        display: "flex",
        gap: 2.25,
        alignItems: "flex-start",
        minWidth: 0,
        "@media (max-width: 1240px)": {
          flexDirection: "column",
        },
      }}
    >
      <Box
        sx={{
          flex: "0 0 auto",
          width: "fit-content",
          maxWidth: "100%",
          minWidth: 0,
        }}
      >
        <TaskLauncherPanel
          activeProfileId={activeProfileId}
          activeTaskId={activeTaskId}
          submitting={submitting}
          taskConfig={taskConfig}
          taskPreset={taskPreset}
          tasks={tasks}
          onInspectTask={onInspectTask}
          onOpenSettings={onOpenSettings}
          onRunTask={onRunTask}
          onTaskPresetChange={onTaskPresetChange}
        />
      </Box>

      <Box
        sx={{
          flex: "1 1 0",
          minWidth: 0,
          width: "100%",
        }}
      >
        <TaskResultsPanel activeTask={activeTask} />
      </Box>
    </Box>
  );
}
