package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

type clusterAssignmentRecord struct {
	assignment interfaces.ClusterAssignment
	leasedAt   string
}

func (s *Service) DispatchTask(req interfaces.ClusterDispatchTaskRequest) ([]interfaces.ClusterRemoteTask, error) {
	if s == nil {
		return nil, fmt.Errorf("cluster service is nil")
	}
	if s.Status().Role != interfaces.ClusterRoleMaster {
		return nil, fmt.Errorf("remote task dispatch requires master mode")
	}
	if len(req.KnightIDs) == 0 {
		return nil, fmt.Errorf("at least one knight is required")
	}

	task := req.Task
	task.Name = strings.TrimSpace(task.Name)
	if task.Name == "" {
		return nil, fmt.Errorf("task name is required")
	}
	if len(task.Nodes) == 0 {
		return nil, fmt.Errorf("task nodes are required")
	}
	if len(task.Matrices) == 0 {
		return nil, fmt.Errorf("task matrices are required")
	}
	task.Config = task.Config.Normalize()

	now := s.now().UTC().Format(time.RFC3339)
	created := make([]interfaces.ClusterRemoteTask, 0, len(req.KnightIDs))

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, rawKnightID := range req.KnightIDs {
		knightID := strings.TrimSpace(rawKnightID)
		if knightID == "" {
			return nil, fmt.Errorf("knight ID is required")
		}

		knight, ok := s.knights[knightID]
		if !ok {
			return nil, fmt.Errorf("knight not found: %s", knightID)
		}

		assignmentID := randomID()
		remoteTaskID := randomID()
		taskName := buildRemoteTaskName(task.Name, knight.KnightName, len(req.KnightIDs))

		assignmentTask := cloneTask(task)
		assignmentTask.Name = taskName
		assignmentTask.Environment = withRemoteExecutionContext(task.Environment, &interfaces.RemoteExecutionContext{
			Mode:         interfaces.RemoteExecutionModeRemoteKnight,
			RemoteTaskID: remoteTaskID,
			KnightID:     knight.KnightID,
			KnightName:   knight.KnightName,
			AssignmentID: assignmentID,
		})

		assignment := interfaces.ClusterAssignment{
			AssignmentID: assignmentID,
			RemoteTaskID: remoteTaskID,
			KnightID:     knight.KnightID,
			KnightName:   knight.KnightName,
			TaskName:     taskName,
			CreatedAt:    now,
			Task:         &assignmentTask,
		}
		remoteTask := interfaces.ClusterRemoteTask{
			AssignmentID: assignmentID,
			RemoteTaskID: remoteTaskID,
			KnightID:     knight.KnightID,
			KnightName:   knight.KnightName,
			TaskName:     taskName,
			State:        interfaces.ClusterRemoteTaskQueued,
			CreatedAt:    now,
		}

		s.assignments[assignmentID] = clusterAssignmentRecord{assignment: assignment}
		s.remoteTasks[remoteTaskID] = remoteTask
		s.remoteTaskIDs = append(s.remoteTaskIDs, remoteTaskID)
		created = append(created, remoteTask)
	}

	s.persistCurrentStateLocked()
	return created, nil
}

func (s *Service) ListRemoteTasks() []interfaces.ClusterRemoteTask {
	if s == nil {
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshKnightHealthLocked(s.now().UTC())

	items := make([]interfaces.ClusterRemoteTask, 0, len(s.remoteTaskIDs))
	for _, remoteTaskID := range s.remoteTaskIDs {
		task, ok := s.remoteTasks[remoteTaskID]
		if !ok {
			continue
		}
		items = append(items, task)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items
}

func (s *Service) leaseAssignmentForKnight(knightID string) *interfaces.ClusterAssignment {
	if s == nil {
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, remoteTaskID := range s.remoteTaskIDs {
		remoteTask, ok := s.remoteTasks[remoteTaskID]
		if !ok || remoteTask.KnightID != knightID || remoteTask.State != interfaces.ClusterRemoteTaskQueued {
			continue
		}

		record, ok := s.assignments[remoteTask.AssignmentID]
		if !ok || record.assignment.Task == nil {
			continue
		}

		now := s.now().UTC().Format(time.RFC3339)
		record.leasedAt = now
		s.assignments[remoteTask.AssignmentID] = record

		remoteTask.State = interfaces.ClusterRemoteTaskRunning
		remoteTask.StartedAt = now
		remoteTask.Error = ""
		s.remoteTasks[remoteTaskID] = remoteTask
		s.persistCurrentStateLocked()

		assignment := record.assignment
		if assignment.Task != nil {
			taskCopy := cloneTask(*assignment.Task)
			assignment.Task = &taskCopy
		}
		return &assignment
	}
	return nil
}

func (s *Service) completeAssignmentReport(ctx context.Context, req interfaces.ClusterKnightReportRequest) {
	if s == nil {
		return
	}

	assignmentID := strings.TrimSpace(req.AssignmentID)
	remoteTaskID := strings.TrimSpace(req.RemoteTaskID)

	s.lock.Lock()
	remoteTask, ok := s.remoteTasks[remoteTaskID]
	if ok {
		remoteTask.FinishedAt = firstNonEmpty(strings.TrimSpace(req.FinishedAt), s.now().UTC().Format(time.RFC3339))
		remoteTask.ExitCode = strings.TrimSpace(req.ExitCode)
		remoteTask.Error = strings.TrimSpace(req.Error)
		if strings.TrimSpace(req.Error) != "" || !strings.EqualFold(strings.TrimSpace(req.Status), "ok") {
			remoteTask.State = interfaces.ClusterRemoteTaskFailed
		} else {
			remoteTask.State = interfaces.ClusterRemoteTaskFinished
			remoteTask.Error = ""
		}
		if req.Archive != nil {
			remoteTask.ArchiveTaskID = firstNonEmpty(strings.TrimSpace(req.Archive.Task.ID), remoteTask.RemoteTaskID)
		}
		if remoteTask.Runtime != nil {
			runtime := *remoteTask.Runtime
			runtime.Status = string(remoteTask.State)
			runtime.FinishedAt = remoteTask.FinishedAt
			remoteTask.Runtime = &runtime
		}
		s.remoteTasks[remoteTaskID] = remoteTask
	}
	delete(s.assignments, assignmentID)
	s.persistCurrentStateLocked()
	s.lock.Unlock()

	if req.Archive == nil || s.saveMasterArchive == nil || !ok {
		return
	}

	archive := normalizeMasterArchive(*req.Archive, remoteTask)
	if err := s.saveMasterArchive(ctx, archive); err != nil {
		clusterLogger().Warn("save remote master archive failed",
			zap.String("remote_task_id", remoteTaskID),
			zap.Error(err),
		)
	}
}

func buildRemoteTaskName(baseName string, knightName string, totalKnights int) string {
	name := strings.TrimSpace(baseName)
	knight := strings.TrimSpace(knightName)
	if knight == "" || totalKnights <= 1 {
		return name
	}
	return name + " @ " + knight
}

func cloneTask(task interfaces.Task) interfaces.Task {
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

func withRemoteExecutionContext(environment *interfaces.TaskEnvironment, remote *interfaces.RemoteExecutionContext) *interfaces.TaskEnvironment {
	var next interfaces.TaskEnvironment
	if environment != nil {
		next = *environment
		if environment.Backend != nil {
			backend := *environment.Backend
			next.Backend = &backend
		}
	}
	if remote != nil {
		remoteCopy := *remote
		next.Remote = &remoteCopy
	}
	return &next
}

func normalizeMasterArchive(archive interfaces.ResultArchive, remoteTask interfaces.ClusterRemoteTask) interfaces.ResultArchive {
	normalized := archive
	normalized.Task.ID = remoteTask.RemoteTaskID
	normalized.Task.Name = firstNonEmpty(strings.TrimSpace(remoteTask.TaskName), normalized.Task.Name)
	normalized.State.TaskID = remoteTask.RemoteTaskID
	normalized.Task.Environment = withRemoteExecutionContext(normalized.Task.Environment, &interfaces.RemoteExecutionContext{
		Mode:         interfaces.RemoteExecutionModeRemoteMaster,
		RemoteTaskID: remoteTask.RemoteTaskID,
		KnightID:     remoteTask.KnightID,
		KnightName:   remoteTask.KnightName,
		AssignmentID: remoteTask.AssignmentID,
	})
	return normalized
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
