package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/storage"

	"go.uber.org/zap"
)

const (
	remoteKnightArchivePollInterval = 300 * time.Millisecond
	remoteKnightArchiveGracePeriod  = 5 * time.Second
)

func (s *Service) runKnightAssignment(ctx context.Context, assignment interfaces.ClusterAssignment) (interfaces.ResultArchive, error) {
	if s == nil {
		return interfaces.ResultArchive{}, errServiceNotInitialized
	}
	if assignment.Task == nil {
		return interfaces.ResultArchive{}, fmt.Errorf("cluster assignment task is required")
	}

	task := cloneRemoteAssignmentTask(*assignment.Task)
	task.Environment = ensureKnightRemoteEnvironment(task.Environment, assignment)

	resp, err := s.SubmitTask(interfaces.TaskSubmitRequest{
		Task:         task,
		RemoteIssued: true,
	}, nil)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}
	s.rememberRemoteKnightTask(assignment.RemoteTaskID, resp.TaskID)
	defer s.forgetRemoteKnightTask(assignment.RemoteTaskID)

	return s.waitRemoteKnightArchive(ctx, resp.TaskID)
}

func (s *Service) saveRemoteMasterArchive(_ context.Context, archive interfaces.ResultArchive) error {
	store, err := s.requireRemoteMasterStore("save remote master archive")
	if err != nil {
		return err
	}
	return store.SaveResultArchive(archive)
}

func (s *Service) waitRemoteKnightArchive(ctx context.Context, taskID string) (interfaces.ResultArchive, error) {
	store, err := s.requireRemoteKnightStore("wait remote knight archive")
	if err != nil {
		return interfaces.ResultArchive{}, err
	}

	ticker := time.NewTicker(remoteKnightArchivePollInterval)
	defer ticker.Stop()
	var terminalSince time.Time

	for {
		status, err := s.GetTask(taskID)
		if err == nil && isTaskTerminal(status.Task.Status) {
			archive, loadErr := store.Load(taskID)
			if loadErr == nil {
				return archive, nil
			}
			if archive, snapshotErr := s.buildRemoteKnightArchiveFromSnapshot(taskID); snapshotErr == nil {
				_ = store.SaveResultArchive(archive)
				return archive, nil
			}

			if terminalSince.IsZero() {
				terminalSince = time.Now()
			}
			if shouldRetryRemoteKnightArchiveLoad(loadErr) && time.Since(terminalSince) < remoteKnightArchiveGracePeriod {
				select {
				case <-ctx.Done():
					return interfaces.ResultArchive{}, ctx.Err()
				case <-ticker.C:
				}
				continue
			}

			if strings.EqualFold(status.Task.Status, "finished") {
				return interfaces.ResultArchive{}, fmt.Errorf("remote knight task %s finished but archive is unavailable after %s: %w", taskID, remoteKnightArchiveGracePeriod, loadErr)
			}
			return interfaces.ResultArchive{}, fmt.Errorf("remote knight task %s ended with status %s: %w", taskID, status.Task.Status, loadErr)
		}
		if err == nil {
			terminalSince = time.Time{}
		}

		select {
		case <-ctx.Done():
			return interfaces.ResultArchive{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func shouldRetryRemoteKnightArchiveLoad(err error) bool {
	return err != nil && errors.Is(err, os.ErrNotExist)
}

func (s *Service) buildRemoteKnightArchiveFromSnapshot(taskID string) (interfaces.ResultArchive, error) {
	k, err := s.requireKernel("build remote knight archive from snapshot")
	if err != nil {
		return interfaces.ResultArchive{}, err
	}

	snapshot, ok := k.GetTaskArchiveSnapshot(taskID)
	if !ok {
		return interfaces.ResultArchive{}, fmt.Errorf("task snapshot not found")
	}
	if !isTaskTerminal(snapshot.State.Status) {
		return interfaces.ResultArchive{}, fmt.Errorf("task is not terminal")
	}
	return storage.BuildResultArchive(snapshot), nil
}

func cloneRemoteAssignmentTask(task interfaces.Task) interfaces.Task {
	cloned := task
	cloned.Nodes = append([]interfaces.Node{}, task.Nodes...)
	cloned.Matrices = append([]interfaces.MatrixEntry{}, task.Matrices...)
	if task.Environment != nil {
		env := *task.Environment
		if task.Environment.Backend != nil {
			backend := *task.Environment.Backend
			env.Backend = &backend
		}
		if task.Environment.Remote != nil {
			remote := *task.Environment.Remote
			env.Remote = &remote
		}
		cloned.Environment = &env
	}
	return cloned
}

func ensureKnightRemoteEnvironment(environment *interfaces.TaskEnvironment, assignment interfaces.ClusterAssignment) *interfaces.TaskEnvironment {
	var next interfaces.TaskEnvironment
	if environment != nil {
		next = *environment
		if environment.Backend != nil {
			backend := *environment.Backend
			next.Backend = &backend
		}
	}
	next.Remote = &interfaces.RemoteExecutionContext{
		Mode:         interfaces.RemoteExecutionModeRemoteKnight,
		RemoteTaskID: strings.TrimSpace(assignment.RemoteTaskID),
		KnightID:     strings.TrimSpace(assignment.KnightID),
		KnightName:   strings.TrimSpace(assignment.KnightName),
		AssignmentID: strings.TrimSpace(assignment.AssignmentID),
	}
	return &next
}

func isTaskTerminal(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "finished", "failed", "canceled":
		return true
	default:
		return false
	}
}

func logRemoteTaskDispatch(taskID string, assignment interfaces.ClusterAssignment) {
	apiLogger().Info("remote knight task submitted",
		zap.String("task_id", taskID),
		zap.String("remote_task_id", assignment.RemoteTaskID),
		zap.String("assignment_id", assignment.AssignmentID),
		zap.String("knight_id", assignment.KnightID),
	)
}

func (s *Service) rememberRemoteKnightTask(remoteTaskID string, localTaskID string) {
	if s == nil {
		return
	}

	remoteTaskID = strings.TrimSpace(remoteTaskID)
	localTaskID = strings.TrimSpace(localTaskID)
	if remoteTaskID == "" || localTaskID == "" {
		return
	}

	s.remoteTaskLock.Lock()
	defer s.remoteTaskLock.Unlock()
	s.remoteKnightTasks[remoteTaskID] = localTaskID
}

func (s *Service) forgetRemoteKnightTask(remoteTaskID string) {
	if s == nil {
		return
	}

	remoteTaskID = strings.TrimSpace(remoteTaskID)
	if remoteTaskID == "" {
		return
	}

	s.remoteTaskLock.Lock()
	defer s.remoteTaskLock.Unlock()
	delete(s.remoteKnightTasks, remoteTaskID)
}

func (s *Service) snapshotRemoteKnightTaskState(remoteTaskID string) (*interfaces.TaskState, string, error) {
	if s == nil {
		return nil, "", errServiceNotInitialized
	}

	remoteTaskID = strings.TrimSpace(remoteTaskID)
	if remoteTaskID == "" {
		return nil, "", fmt.Errorf("remote task ID is required")
	}

	s.remoteTaskLock.RLock()
	localTaskID := strings.TrimSpace(s.remoteKnightTasks[remoteTaskID])
	s.remoteTaskLock.RUnlock()
	if localTaskID == "" {
		return nil, "", fmt.Errorf("remote knight task is not active")
	}

	status, err := s.GetTask(localTaskID)
	if err != nil {
		return nil, localTaskID, err
	}
	return &status.Task, localTaskID, nil
}
