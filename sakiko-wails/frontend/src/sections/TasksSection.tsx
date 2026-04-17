import { Box } from "@mui/material";
import { TaskLauncherPanel } from "../components/task/TaskLauncherPanel";
import { TaskResultsPanel } from "../components/task/TaskResultsPanel";
import type { TaskPreset, TaskPresetSelection } from "../types/dashboard";
import type { TaskConfig, TaskState, TaskStatusResponse } from "../types/sakiko";

type TasksSectionProps = {
  activeProfileId: string;
  activeTask: TaskStatusResponse | null;
  activeTaskId: string;
  submitting: boolean;
  taskConfig: TaskConfig;
  taskPreset: TaskPresetSelection;
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
  const hasActiveTask = Boolean(activeTask);

  return (
    <Box
      sx={{
        display: "grid",
        gap: 2.25,
        alignItems: "flex-start",
        minWidth: 0,
        gridTemplateColumns: hasActiveTask
          ? {
              xs: "minmax(0, 1fr)",
              xl: "minmax(360px, 420px) minmax(0, 1fr)",
            }
          : "minmax(0, 1fr)",
        "& > *": {
          minWidth: 0,
        },
      }}
    >
      <Box
        sx={{
          width: "100%",
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
