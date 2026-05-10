package api

import (
	"fmt"
	"strings"

	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

func (s *Service) SubmitTask(req interfaces.TaskSubmitRequest, onEvent func(interfaces.Event)) (interfaces.TaskSubmitResponse, error) {
	k, err := s.requireKernel("submit task")
	if err != nil {
		return interfaces.TaskSubmitResponse{}, err
	}

	c, _ := s.requireCluster("")
	if c != nil && c.Status().Role == interfaces.ClusterRoleKnight && !req.RemoteIssued {
		apiLogger().Warn("submit task rejected: knight mode forbids standalone tasks")
		return interfaces.TaskSubmitResponse{}, fmt.Errorf("standalone task submission is disabled while running as knight")
	}

	task := req.Task
	if identity := strings.TrimSpace(task.Config.BackendIdentity); identity != "" {
		task.Environment = &interfaces.TaskEnvironment{
			Identity: identity,
		}
	}

	taskID, err := k.Submit(task, onEvent)
	if err != nil {
		apiLogger().Warn("submit task failed",
			zap.String("task_name", task.Name),
			zap.Int("node_count", len(task.Nodes)),
			zap.Error(err),
		)
		return interfaces.TaskSubmitResponse{}, err
	}

	apiLogger().Debug("task submitted",
		zap.String("task_id", taskID),
		zap.String("task_name", task.Name),
		zap.Int("node_count", len(task.Nodes)),
	)
	return interfaces.TaskSubmitResponse{TaskID: taskID}, nil
}

func (s *Service) ListTasks() interfaces.TaskListResponse {
	k, err := s.requireKernel("list tasks")
	if err != nil {
		return interfaces.TaskListResponse{Tasks: []interfaces.TaskState{}}
	}
	return interfaces.TaskListResponse{Tasks: k.ListTasks()}
}

func (s *Service) GetTask(taskID string) (interfaces.TaskStatusResponse, error) {
	k, err := s.requireKernel("get task")
	if err != nil {
		return interfaces.TaskStatusResponse{}, err
	}

	task, ok := k.GetTask(taskID)
	if !ok {
		apiLogger().Debug("task not found", zap.String("task_id", taskID))
		return interfaces.TaskStatusResponse{}, fmt.Errorf("task not found")
	}
	return task, nil
}

func (s *Service) CancelTask(taskID string) error {
	k, err := s.requireKernel("cancel task")
	if err != nil {
		return err
	}

	if err := k.CancelTask(taskID); err != nil {
		apiLogger().Warn("cancel task failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}

	apiLogger().Info("task cancel requested", zap.String("task_id", taskID))
	return nil
}

func (s *Service) DeleteTask(taskID string) error {
	k, err := s.requireKernel("delete task")
	if err != nil {
		return err
	}

	if err := k.DeleteTask(taskID); err != nil {
		apiLogger().Warn("delete task failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}

	apiLogger().Info("task deleted", zap.String("task_id", taskID))
	return nil
}

func (s *Service) RuntimeStatus() interfaces.RuntimeStatusResponse {
	k, err := s.requireKernel("runtime status")
	if err != nil {
		return interfaces.RuntimeStatusResponse{}
	}
	return interfaces.RuntimeStatusResponse{Status: k.RuntimeStatus()}
}
