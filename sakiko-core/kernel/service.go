package kernel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"sakiko.local/sakiko-core/executor"
	"sakiko.local/sakiko-core/executor/taskpoll"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"

	"go.uber.org/zap"
)

type Config struct {
	Mode             interfaces.Mode
	ConnConcurrency  uint
	SpeedConcurrency uint
	SpeedInterval    time.Duration
	ArchiveWriter    interfaces.ResultArchiveWriter
}

const maxRetainedFinishedTasks = 5

type Service struct {
	engine        *executor.Engine
	now           func() time.Time
	archiveWriter interfaces.ResultArchiveWriter

	lock          sync.RWMutex
	tasks         map[string]*taskRecord
	finishedOrder []string
}

type taskRecord struct {
	task        interfaces.Task
	state       interfaces.TaskState
	results     []interfaces.EntryResult
	exitCode    string
	activeNodes map[int]interfaces.TaskActiveNode
}

func New(cfg Config) (*Service, error) {
	var engine *executor.Engine
	switch cfg.Mode {
	case "", interfaces.ModeParallel:
		engine = executor.NewEngine(executor.Config{
			ConnConcurrency:  cfg.ConnConcurrency,
			SpeedConcurrency: cfg.SpeedConcurrency,
			SpeedInterval:    cfg.SpeedInterval,
		})
	case interfaces.ModeSerial:
		engine = executor.NewSerialEngine()
	default:
		return nil, fmt.Errorf("unsupported mode: %s", cfg.Mode)
	}
	service := &Service{
		engine:        engine,
		now:           time.Now,
		archiveWriter: cfg.ArchiveWriter,
		tasks:         map[string]*taskRecord{},
		finishedOrder: []string{},
	}
	kernelLogger().Info("kernel initialized",
		zap.String("mode", string(cfg.Mode)),
		zap.Uint("conn_concurrency", cfg.ConnConcurrency),
		zap.Uint("speed_concurrency", cfg.SpeedConcurrency),
		zap.Duration("speed_interval", cfg.SpeedInterval),
	)
	return service, nil
}

func (s *Service) Stop() {
	if s != nil && s.engine != nil {
		kernelLogger().Info("kernel stopping")
		s.engine.Stop()
	}
}

func (s *Service) Submit(task interfaces.Task, onEvent func(interfaces.Event)) (string, error) {
	total := len(task.Nodes)
	if task.ID == "" {
		task.ID = randomID()
	}
	kernelLogger().Info("submitting task",
		zap.String("task_id", task.ID),
		zap.String("task_name", task.Name),
		zap.Int("node_count", total),
		zap.Int("matrix_count", len(task.Matrices)),
		zap.String("vendor", string(task.Vendor)),
	)
	s.addTask(task, total)

	taskID, err := s.engine.Submit(task, executor.Callbacks{
		OnUpdate: func(taskID string, activeNode interfaces.TaskActiveNode) {
			state := s.updateTaskActivity(taskID, activeNode)
			kernelLogger().Debug("task activity",
				zap.String("task_id", taskID),
				zap.Int("node_index", activeNode.NodeIndex),
				zap.String("node_name", activeNode.NodeName),
				zap.String("phase", string(activeNode.Phase)),
				zap.String("macro", string(activeNode.Macro)),
				zap.String("matrix", string(activeNode.Matrix)),
				zap.Int("active_nodes", len(state.ActiveNodes)),
			)
		},
		OnProcess: func(taskID string, index int, result interfaces.EntryResult, queuing int) {
			state := s.completeTaskNode(taskID, index, total, index+1, queuing)
			kernelLogger().Debug("task progress",
				zap.String("task_id", taskID),
				zap.Int("progress", index+1),
				zap.Int("total", total),
				zap.Int("queuing", queuing),
				zap.String("proxy_name", result.ProxyInfo.Name),
				zap.String("result_error", result.Error),
			)
			if onEvent != nil {
				onEvent(interfaces.Event{
					Type:    interfaces.EventProcess,
					TaskID:  taskID,
					Index:   index,
					Queuing: queuing,
					Result:  result,
					Task:    state,
				})
			}
		},
		OnExit: func(taskID string, results []interfaces.EntryResult, exitCode taskpoll.ExitCode) {
			state, snapshot := s.finishTask(taskID, results, exitCode)
			kernelLogger().Info("task finished",
				zap.String("task_id", taskID),
				zap.String("exit_code", string(exitCode)),
				zap.Int("result_count", len(results)),
			)
			if s.archiveWriter != nil && snapshot.Task.ID != "" {
				if err := s.archiveWriter.SaveTaskArchive(snapshot); err != nil {
					kernelLogger().Warn("save task archive failed",
						zap.String("task_id", taskID),
						zap.Error(err),
					)
				}
			}
			if onEvent != nil {
				onEvent(interfaces.Event{
					Type:     interfaces.EventExit,
					TaskID:   taskID,
					Results:  results,
					Task:     state,
					ExitCode: string(exitCode),
				})
			}
		},
	})
	if err != nil {
		s.deleteTask(task.ID)
		kernelLogger().Warn("submit task failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return "", err
	}
	return taskID, nil
}

func (s *Service) ListTasks() []interfaces.TaskState {
	s.lock.RLock()
	defer s.lock.RUnlock()

	out := make([]interfaces.TaskState, 0, len(s.tasks))
	for _, task := range s.tasks {
		out = append(out, task.state)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return newerTaskState(out[i], out[j])
	})
	return out
}

func (s *Service) GetTask(taskID string) (interfaces.TaskStatusResponse, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskStatusResponse{}, false
	}
	return interfaces.TaskStatusResponse{
		Task:     task.state,
		Results:  append([]interfaces.EntryResult{}, task.results...),
		ExitCode: task.exitCode,
	}, true
}

func (s *Service) RuntimeStatus() interfaces.RuntimeStatus {
	s.lock.RLock()
	defer s.lock.RUnlock()

	running := 0
	for _, task := range s.tasks {
		if task.state.Status == "running" || task.state.Status == "stopping" {
			running++
		}
	}
	return interfaces.RuntimeStatus{
		Running:     running > 0,
		RunningTask: running,
		TotalTask:   len(s.tasks),
	}
}

func (s *Service) CancelTask(taskID string) error {
	if s == nil || s.engine == nil {
		return fmt.Errorf("kernel not initialized")
	}

	s.lock.Lock()
	record, ok := s.tasks[taskID]
	if !ok {
		s.lock.Unlock()
		return fmt.Errorf("task not found")
	}
	switch record.state.Status {
	case "finished", "failed", "canceled":
		s.lock.Unlock()
		return fmt.Errorf("task already finished")
	case "stopping":
		s.lock.Unlock()
		return nil
	default:
		record.state.Status = "stopping"
		record.state.ActiveNodes = buildActiveNodes(record.activeNodes)
	}
	s.lock.Unlock()

	if !s.engine.Cancel(taskID) {
		s.lock.Lock()
		if current, ok := s.tasks[taskID]; ok && current.state.Status == "stopping" {
			current.state.Status = "running"
		}
		s.lock.Unlock()
		return fmt.Errorf("task is not cancelable")
	}

	kernelLogger().Info("task cancel requested", zap.String("task_id", taskID))
	return nil
}

func (s *Service) DeleteTask(taskID string) error {
	if s == nil {
		return fmt.Errorf("kernel not initialized")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	record, ok := s.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found")
	}
	switch record.state.Status {
	case "running", "stopping":
		return fmt.Errorf("task is still active")
	default:
		s.removeFinishedTaskLocked(taskID)
		delete(s.tasks, taskID)
		return nil
	}
}

func (s *Service) addTask(task interfaces.Task, total int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.tasks[task.ID] = &taskRecord{
		task: task,
		state: interfaces.TaskState{
			TaskID:      task.ID,
			Name:        task.Name,
			Status:      "running",
			Total:       total,
			StartedAt:   s.now().UTC().Format(time.RFC3339),
			ActiveNodes: []interfaces.TaskActiveNode{},
		},
		activeNodes: map[int]interfaces.TaskActiveNode{},
	}
}

func (s *Service) updateTaskActivity(taskID string, activeNode interfaces.TaskActiveNode) interfaces.TaskState {
	s.lock.Lock()
	defer s.lock.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskState{}
	}
	if task.activeNodes == nil {
		task.activeNodes = map[int]interfaces.TaskActiveNode{}
	}
	activeNode.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	task.activeNodes[activeNode.NodeIndex] = activeNode
	task.state.ActiveNodes = buildActiveNodes(task.activeNodes)
	return task.state
}

func (s *Service) completeTaskNode(taskID string, nodeIndex int, total int, progress int, queuing int) interfaces.TaskState {
	s.lock.Lock()
	defer s.lock.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskState{}
	}
	delete(task.activeNodes, nodeIndex)
	task.state.Total = total
	task.state.Progress = progress
	task.state.Queuing = queuing
	task.state.ActiveNodes = buildActiveNodes(task.activeNodes)
	return task.state
}

func (s *Service) finishTask(taskID string, results []interfaces.EntryResult, exitCode taskpoll.ExitCode) (interfaces.TaskState, interfaces.TaskArchiveSnapshot) {
	s.lock.Lock()
	defer s.lock.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskState{}, interfaces.TaskArchiveSnapshot{}
	}
	if exitCode == taskpoll.ExitCodeStopped {
		task.state.Status = "canceled"
	} else {
		task.state.Status = "finished"
	}
	task.state.ActiveNodes = nil
	task.state.Queuing = 0
	task.state.FinishedAt = s.now().UTC().Format(time.RFC3339)
	task.results = append([]interfaces.EntryResult{}, results...)
	task.exitCode = string(exitCode)
	task.activeNodes = map[int]interfaces.TaskActiveNode{}
	s.rememberFinishedTaskLocked(taskID)
	s.pruneFinishedTasksLocked()
	return task.state, interfaces.TaskArchiveSnapshot{
		Task:     task.task,
		State:    task.state,
		Results:  append([]interfaces.EntryResult{}, task.results...),
		ExitCode: task.exitCode,
	}
}

func (s *Service) deleteTask(taskID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.removeFinishedTaskLocked(taskID)
	delete(s.tasks, taskID)
}

func (s *Service) rememberFinishedTaskLocked(taskID string) {
	s.removeFinishedTaskLocked(taskID)
	s.finishedOrder = append(s.finishedOrder, taskID)
}

func (s *Service) pruneFinishedTasksLocked() {
	for len(s.finishedOrder) > maxRetainedFinishedTasks {
		evictedID := s.finishedOrder[0]
		s.finishedOrder = s.finishedOrder[1:]
		record, ok := s.tasks[evictedID]
		if !ok {
			continue
		}
		switch record.state.Status {
		case "finished", "failed", "canceled":
			delete(s.tasks, evictedID)
		}
	}
}

func (s *Service) removeFinishedTaskLocked(taskID string) {
	for i, existing := range s.finishedOrder {
		if existing != taskID {
			continue
		}
		s.finishedOrder = append(s.finishedOrder[:i], s.finishedOrder[i+1:]...)
		return
	}
}

func randomID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func kernelLogger() *zap.Logger {
	return logx.Named("core.kernel")
}

func newerTaskState(left interfaces.TaskState, right interfaces.TaskState) bool {
	leftStartedAt := parseTaskTimestamp(left.StartedAt)
	rightStartedAt := parseTaskTimestamp(right.StartedAt)
	if leftStartedAt != rightStartedAt {
		return leftStartedAt.After(rightStartedAt)
	}

	leftFinishedAt := parseTaskTimestamp(left.FinishedAt)
	rightFinishedAt := parseTaskTimestamp(right.FinishedAt)
	if leftFinishedAt != rightFinishedAt {
		return leftFinishedAt.After(rightFinishedAt)
	}

	return left.TaskID > right.TaskID
}

func parseTaskTimestamp(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func buildActiveNodes(activeNodes map[int]interfaces.TaskActiveNode) []interfaces.TaskActiveNode {
	if len(activeNodes) == 0 {
		return []interfaces.TaskActiveNode{}
	}

	out := make([]interfaces.TaskActiveNode, 0, len(activeNodes))
	for _, activeNode := range activeNodes {
		out = append(out, activeNode)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].NodeIndex != out[j].NodeIndex {
			return out[i].NodeIndex < out[j].NodeIndex
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out
}
