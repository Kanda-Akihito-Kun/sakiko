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
  onDeleteTask: (taskId: string) => Promise<void>;
  onOpenConfigs: () => void;
  onRunTask: () => Promise<void>;
  onStopTask: () => Promise<void>;
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
  onDeleteTask,
  onOpenConfigs,
  onRunTask,
  onStopTask,
  onTaskPresetChange,
}: TasksSectionProps) {
  return (
    <Box className="sakiko-section-grid">
      <Box className="sakiko-section-grid__sidebar">
        <TaskLauncherPanel
          activeProfileId={activeProfileId}
          activeTaskId={activeTaskId}
          submitting={submitting}
          taskConfig={taskConfig}
          taskPreset={taskPreset}
          tasks={tasks}
          onDeleteTask={onDeleteTask}
          onInspectTask={onInspectTask}
          onOpenConfigs={onOpenConfigs}
          onRunTask={onRunTask}
          onTaskPresetChange={onTaskPresetChange}
        />
      </Box>

      <Box className="sakiko-section-grid__content">
        <TaskResultsPanel activeTask={activeTask} onStopTask={onStopTask} />
      </Box>
    </Box>
  );
}
