package executor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"sakiko.local/sakiko-core/executor/taskpoll"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/macro"
	"sakiko.local/sakiko-core/matrix"
	"sakiko.local/sakiko-core/vendors"

	"go.uber.org/zap"
)

const defaultMacroInterval = 1500 * time.Millisecond

var (
	findMacroFunc         = macro.Find
	waitMacroIntervalFunc = waitMacroInterval
)

type Config struct {
	SpeedConcurrency uint
	ConnConcurrency  uint
	SpeedInterval    time.Duration
}

type Callbacks struct {
	OnUpdate  func(taskID string, activeNode interfaces.TaskActiveNode)
	OnProcess func(taskID string, index int, result interfaces.EntryResult, queuing int)
	OnExit    func(taskID string, results []interfaces.EntryResult, exitCode taskpoll.ExitCode)
}

type Engine struct {
	speedPoll *taskpoll.Controller
	connPoll  *taskpoll.Controller
	stop      chan struct{}
	once      sync.Once
}

func NewSerialEngine() *Engine {
	return NewEngine(Config{
		SpeedConcurrency: 1,
		ConnConcurrency:  1,
	})
}

func NewEngine(cfg Config) *Engine {
	if cfg.SpeedConcurrency == 0 {
		cfg.SpeedConcurrency = 1
	}
	if cfg.ConnConcurrency == 0 {
		cfg.ConnConcurrency = 16
	}
	e := &Engine{
		speedPoll: taskpoll.New("speed", cfg.SpeedConcurrency, cfg.SpeedInterval, 200*time.Millisecond),
		connPoll:  taskpoll.New("conn", cfg.ConnConcurrency, 0, 200*time.Millisecond),
		stop:      make(chan struct{}),
	}
	executorLogger().Info("executor initialized",
		zap.Uint("speed_concurrency", cfg.SpeedConcurrency),
		zap.Uint("conn_concurrency", cfg.ConnConcurrency),
		zap.Duration("speed_interval", cfg.SpeedInterval),
	)
	go e.speedPoll.Start(e.stop)
	go e.connPoll.Start(e.stop)
	return e
}

func (e *Engine) Stop() {
	e.once.Do(func() {
		executorLogger().Info("executor stopping")
		close(e.stop)
	})
}

func (e *Engine) Cancel(taskID string) bool {
	if e == nil {
		return false
	}
	if e.speedPoll != nil && e.speedPoll.Cancel(taskID) {
		executorLogger().Info("task cancel requested in speed queue", zap.String("task_id", taskID))
		return true
	}
	if e.connPoll != nil && e.connPoll.Cancel(taskID) {
		executorLogger().Info("task cancel requested in connection queue", zap.String("task_id", taskID))
		return true
	}
	return false
}

func (e *Engine) Submit(task interfaces.Task, cb Callbacks) (string, error) {
	if len(task.Nodes) == 0 {
		return "", fmt.Errorf("empty nodes")
	}
	if task.ID == "" {
		task.ID = randomID()
	}
	task.Config = task.Config.Normalize()

	macroSet := map[interfaces.MacroType]struct{}{}
	for _, entry := range task.Matrices {
		macroSet[matrix.Find(entry.Type).MacroJob()] = struct{}{}
	}
	macros := make([]interfaces.MacroType, 0, len(macroSet))
	for mt := range macroSet {
		if mt != interfaces.MacroInvalid {
			macros = append(macros, mt)
		}
	}

	item := (&pollItem{
		id:       task.ID,
		name:     task.Name,
		task:     task,
		matrices: task.Matrices,
		macros:   macros,
		results:  make([]interfaces.EntryResult, len(task.Nodes)),
		onUpdate: func(self *pollItem, activeNode interfaces.TaskActiveNode) {
			if cb.OnUpdate != nil {
				cb.OnUpdate(self.id, activeNode)
			}
		},
		onProcess: func(self *pollItem, idx int, result interfaces.EntryResult, c *taskpoll.Controller) {
			if cb.OnProcess != nil {
				cb.OnProcess(self.id, idx, result, c.AwaitingCount())
			}
		},
		onExit: func(self *pollItem, exitCode taskpoll.ExitCode) {
			if cb.OnExit != nil {
				cb.OnExit(self.id, append([]interfaces.EntryResult{}, self.results...), exitCode)
			}
		},
	}).Init().(*pollItem)

	isSpeed := false
	for _, mt := range macros {
		if mt == interfaces.MacroSpeed {
			isSpeed = true
			break
		}
	}
	if isSpeed {
		executorLogger().Info("task routed to speed queue",
			zap.String("task_id", task.ID),
			zap.String("task_name", task.Name),
			zap.Int("node_count", len(task.Nodes)),
			zap.Int("macro_count", len(macros)),
		)
		e.speedPoll.Push(item)
	} else {
		executorLogger().Info("task routed to connection queue",
			zap.String("task_id", task.ID),
			zap.String("task_name", task.Name),
			zap.Int("node_count", len(task.Nodes)),
			zap.Int("macro_count", len(macros)),
		)
		e.connPoll.Push(item)
	}
	return task.ID, nil
}

func runMacros(ctx context.Context, v interfaces.Vendor, task *interfaces.Task, entries []interfaces.MatrixEntry, macroTypes []interfaces.MacroType, onUpdate func(interfaces.TaskActiveNode)) (map[interfaces.MacroType]interfaces.Macro, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	out := map[interfaces.MacroType]interfaces.Macro{}
	errs := make([]error, 0, len(macroTypes))
	for index, macroType := range macroExecutionOrder(macroTypes) {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if index > 0 {
			if err := waitMacroIntervalFunc(ctx, defaultMacroInterval); err != nil {
				errs = append(errs, err)
				break
			}
		}
		notifyTaskActiveNode(onUpdate, interfaces.TaskActiveNode{
			NodeName:    v.ProxyInfo().Name,
			NodeAddress: v.ProxyInfo().Address,
			Protocol:    v.ProxyInfo().Type,
			Phase:       interfaces.TaskRuntimePhaseMacro,
			Macro:       macroType,
			Matrices:    matrixTypesForMacro(entries, macroType),
		})
		m := findMacroFunc(macroType)
		err := m.Run(ctx, v, task)
		if err != nil {
			executorLogger().Warn("macro execution failed",
				zap.String("macro_type", string(macroType)),
				zap.String("proxy_name", v.ProxyInfo().Name),
				zap.Error(err),
			)
			errs = append(errs, fmt.Errorf("%s: %w", macroType, err))
		}
		out[macroType] = m
	}
	return out, errors.Join(errs...)
}

func waitMacroInterval(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func macroExecutionOrder(macroTypes []interfaces.MacroType) []interfaces.MacroType {
	priority := map[interfaces.MacroType]int{
		interfaces.MacroPing:  0,
		interfaces.MacroGeo:   1,
		interfaces.MacroUDP:   2,
		interfaces.MacroSpeed: 3,
		interfaces.MacroMedia: 4,
	}

	seen := make(map[interfaces.MacroType]struct{}, len(macroTypes))
	ordered := make([]interfaces.MacroType, 0, len(macroTypes))

	appendIfPresent := func(target interfaces.MacroType) {
		for _, macroType := range macroTypes {
			if macroType != target {
				continue
			}
			if _, exists := seen[macroType]; exists {
				return
			}
			seen[macroType] = struct{}{}
			ordered = append(ordered, macroType)
			return
		}
	}

	appendIfPresent(interfaces.MacroPing)
	appendIfPresent(interfaces.MacroGeo)
	appendIfPresent(interfaces.MacroUDP)
	appendIfPresent(interfaces.MacroSpeed)
	appendIfPresent(interfaces.MacroMedia)

	for _, macroType := range macroTypes {
		if _, exists := seen[macroType]; exists {
			continue
		}
		if _, known := priority[macroType]; known {
			continue
		}
		seen[macroType] = struct{}{}
		ordered = append(ordered, macroType)
	}

	return ordered
}

func extractMatrices(entries []interfaces.MatrixEntry, macroMap map[interfaces.MacroType]interfaces.Macro, onUpdate func(interfaces.TaskActiveNode)) []interfaces.MatrixResult {
	out := make([]interfaces.MatrixResult, 0, len(entries))
	for _, entry := range entries {
		m := matrix.Find(entry.Type)
		mac := macroMap[m.MacroJob()]
		if mac == nil {
			mac = macro.Find(interfaces.MacroInvalid)
		}
		notifyTaskActiveNode(onUpdate, interfaces.TaskActiveNode{
			Phase:    interfaces.TaskRuntimePhaseMatrix,
			Macro:    m.MacroJob(),
			Matrix:   entry.Type,
			Matrices: []interfaces.MatrixType{entry.Type},
		})
		m.Extract(entry, mac)
		out = append(out, interfaces.MatrixResult{Type: m.Type(), Payload: m.Payload()})
	}
	return out
}

func matrixTypesForMacro(entries []interfaces.MatrixEntry, target interfaces.MacroType) []interfaces.MatrixType {
	out := make([]interfaces.MatrixType, 0, len(entries))
	for _, entry := range entries {
		if matrix.Find(entry.Type).MacroJob() != target {
			continue
		}
		out = append(out, entry.Type)
	}
	return out
}

func notifyTaskActiveNode(onUpdate func(interfaces.TaskActiveNode), activeNode interfaces.TaskActiveNode) {
	if onUpdate == nil {
		return
	}
	onUpdate(activeNode)
}

func buildVendor(task interfaces.Task, idx int) interfaces.Vendor {
	return vendors.Find(task.Vendor).Build(task.Nodes[idx])
}

func randomID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func executorLogger() *zap.Logger {
	return logx.Named("core.executor")
}
