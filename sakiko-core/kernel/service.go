package kernel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

type Service struct {
	engine        *executor.Engine
	now           func() time.Time
	archiveWriter interfaces.ResultArchiveWriter

	lock  sync.RWMutex
	tasks map[string]*taskRecord
}

type taskRecord struct {
	task     interfaces.Task
	state    interfaces.TaskState
	results  []interfaces.EntryResult
	exitCode string
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
		OnProcess: func(taskID string, index int, result interfaces.EntryResult, queuing int) {
			state := s.updateTaskProgress(taskID, total, index+1, queuing)
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
		if task.state.Status == "running" {
			running++
		}
	}
	return interfaces.RuntimeStatus{
		Running:     running > 0,
		RunningTask: running,
		TotalTask:   len(s.tasks),
	}
}

func (s *Service) addTask(task interfaces.Task, total int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.tasks[task.ID] = &taskRecord{
		task: task,
		state: interfaces.TaskState{
			TaskID:    task.ID,
			Name:      task.Name,
			Status:    "running",
			Total:     total,
			StartedAt: s.now().UTC().Format(time.RFC3339),
		},
	}
}

func (s *Service) updateTaskProgress(taskID string, total int, progress int, queuing int) interfaces.TaskState {
	s.lock.Lock()
	defer s.lock.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskState{}
	}
	task.state.Total = total
	task.state.Progress = progress
	task.state.Queuing = queuing
	return task.state
}

func (s *Service) finishTask(taskID string, results []interfaces.EntryResult, exitCode taskpoll.ExitCode) (interfaces.TaskState, interfaces.TaskArchiveSnapshot) {
	s.lock.Lock()
	defer s.lock.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return interfaces.TaskState{}, interfaces.TaskArchiveSnapshot{}
	}
	task.state.Status = "finished"
	task.state.FinishedAt = s.now().UTC().Format(time.RFC3339)
	task.results = append([]interfaces.EntryResult{}, results...)
	task.exitCode = string(exitCode)
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

	delete(s.tasks, taskID)
}

func randomID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func kernelLogger() *zap.Logger {
	return logx.Named("core.kernel")
}
