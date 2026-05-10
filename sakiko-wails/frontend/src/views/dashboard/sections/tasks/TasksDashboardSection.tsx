import { useShallow } from "zustand/react/shallow";
import { TaskLauncherPanel } from "../../../../components/task/TaskLauncherPanel";
import { TaskResultsPanel } from "../../../../components/task/TaskResultsPanel";
import { useDashboardStore } from "../../../../store/dashboardStore";
import { SectionLayout, SectionStack } from "../../components/SectionLayout";

type TasksDashboardSectionProps = {
  onOpenConfigs: () => void;
};

export function TasksDashboardSection({ onOpenConfigs }: TasksDashboardSectionProps) {
  const view = useDashboardStore(useShallow((state) => ({
    activeProfileId: state.activeProfileId,
    activeTask: state.activeTask,
    activeTaskId: state.activeTaskId,
    submitting: state.submitting,
    taskConfig: state.taskConfig,
    taskPreset: state.taskPreset,
    tasks: state.tasks,
    handleInspectTask: state.handleInspectTask,
    handleDeleteTask: state.handleDeleteTask,
    handleRunTask: state.handleRunTask,
    handleStopTask: state.handleStopTask,
    setTaskPreset: state.setTaskPreset,
  })));

  return (
    <SectionLayout
      columns={{
        xs: "minmax(0, 1fr)",
        lg: "clamp(340px, 30vw, 420px) minmax(0, 1fr)",
      }}
    >
      <SectionStack>
        <TaskLauncherPanel
          activeProfileId={view.activeProfileId}
          activeTaskId={view.activeTaskId}
          submitting={view.submitting}
          taskConfig={view.taskConfig}
          taskPreset={view.taskPreset}
          tasks={view.tasks}
          onDeleteTask={view.handleDeleteTask}
          onInspectTask={view.handleInspectTask}
          onOpenConfigs={onOpenConfigs}
          onRunTask={view.handleRunTask}
          onTaskPresetChange={view.setTaskPreset}
        />
      </SectionStack>

      <SectionStack>
        <TaskResultsPanel activeTask={view.activeTask} onStopTask={view.handleStopTask} />
      </SectionStack>
    </SectionLayout>
  );
}
