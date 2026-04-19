package executor

import (
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sakiko.local/sakiko-core/executor/taskpoll"
	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

type pollItem struct {
	id       string
	name     string
	task     interfaces.Task
	matrices []interfaces.MatrixEntry
	macros   []interfaces.MacroType
	results  []interfaces.EntryResult

	onUpdate  func(self *pollItem, activeNode interfaces.TaskActiveNode)
	onProcess func(self *pollItem, idx int, result interfaces.EntryResult, c *taskpoll.Controller)
	onExit    func(self *pollItem, exitCode taskpoll.ExitCode)

	lock     sync.Mutex
	exitOnce sync.Once
	canceled atomic.Bool
}

var executeNodeAttemptFunc = executeNodeAttempt

func (p *pollItem) ID() string {
	return p.id
}

func (p *pollItem) TaskName() string {
	if p.name != "" {
		return p.name
	}
	return p.id
}

func (p *pollItem) Count() int {
	return len(p.task.Nodes)
}

func (p *pollItem) Yield(idx int, c *taskpoll.Controller) {
	result := interfaces.EntryResult{}
	if p.IsCanceled() {
		return
	}
	defer func() {
		p.lock.Lock()
		p.results[idx] = result
		p.lock.Unlock()
		if p.onProcess != nil {
			p.onProcess(p, idx, result, c)
		}
	}()

	attempts := nodeRetryAttempts(p.task, p.macros)
	if attempts < 1 {
		attempts = 1
	}

	bestScore := -1
	for attempt := 1; attempt <= attempts; attempt++ {
		if p.IsCanceled() {
			break
		}
		candidate := executeNodeAttemptFunc(p.task, idx, p.matrices, p.macros, func(activeNode interfaces.TaskActiveNode) {
			p.emitNodeUpdate(idx, attempt, activeNode)
		})
		score := countMatrixFailures(p.matrices, candidate)
		if bestScore < 0 || score < bestScore || (score == bestScore && strings.TrimSpace(result.Error) != "" && strings.TrimSpace(candidate.Error) == "") {
			result = candidate
			bestScore = score
		}
		if score == 0 {
			break
		}
		if attempt < attempts {
			if p.IsCanceled() {
				break
			}
			executorLogger().Info("retrying node execution",
				zap.String("task_id", p.id),
				zap.Int("index", idx),
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", attempts),
				zap.String("proxy_name", candidate.ProxyInfo.Name),
				zap.Int("failure_score", score),
			)
			time.Sleep(200 * time.Millisecond)
		}
	}

	executorLogger().Debug("node execution completed",
		zap.String("task_id", p.id),
		zap.Int("index", idx),
		zap.String("proxy_name", result.ProxyInfo.Name),
		zap.Int64("invoke_duration_ms", result.InvokeDuration),
		zap.String("result_error", result.Error),
	)
}

func (p *pollItem) Cancel() {
	p.canceled.Store(true)
}

func (p *pollItem) IsCanceled() bool {
	return p.canceled.Load()
}

func (p *pollItem) emitNodeUpdate(idx int, attempt int, activeNode interfaces.TaskActiveNode) {
	if p.onUpdate == nil {
		return
	}

	node := p.task.Nodes[idx]
	activeNode.NodeIndex = idx
	if strings.TrimSpace(activeNode.NodeName) == "" {
		activeNode.NodeName = node.Name
	}
	if strings.TrimSpace(activeNode.NodeAddress) == "" {
		activeNode.NodeAddress = node.Server
	}
	if activeNode.Attempt == 0 {
		activeNode.Attempt = attempt
	}
	p.onUpdate(p, activeNode)
}

func nodeRetryAttempts(task interfaces.Task, macros []interfaces.MacroType) int {
	attempts := int(task.Config.TaskRetry)
	if attempts < 1 {
		return 1
	}

	for _, macroType := range macros {
		switch macroType {
		case interfaces.MacroSpeed, interfaces.MacroMedia:
			// Expensive macros already have internal retry / retry-like behavior.
			// Re-running the whole node makes full tasks disproportionately slow.
			return 1
		}
	}

	return attempts
}

func executeNodeAttempt(task interfaces.Task, idx int, matrices []interfaces.MatrixEntry, macros []interfaces.MacroType, onUpdate func(interfaces.TaskActiveNode)) interfaces.EntryResult {
	start := time.Now().UnixMilli()
	result := interfaces.EntryResult{}

	vendor := buildVendor(task, idx)
	result.ProxyInfo = vendor.ProxyInfo()
	notifyTaskActiveNode(onUpdate, interfaces.TaskActiveNode{
		NodeName:    result.ProxyInfo.Name,
		NodeAddress: result.ProxyInfo.Address,
		Protocol:    result.ProxyInfo.Type,
		Phase:       interfaces.TaskRuntimePhasePreparing,
	})
	if vendor.Status() != interfaces.VStatusOperational {
		result.Error = "vendor not ready"
		result.InvokeDuration = time.Now().UnixMilli() - start
		executorLogger().Warn("vendor not ready",
			zap.String("task_id", task.ID),
			zap.Int("index", idx),
			zap.String("proxy_name", result.ProxyInfo.Name),
			zap.String("vendor", string(vendor.Type())),
		)
		return result
	}
	macroMap, err := runMacros(vendor, &task, matrices, macros, onUpdate)
	result.InvokeDuration = time.Now().UnixMilli() - start
	if err != nil {
		result.Error = err.Error()
		executorLogger().Warn("node execution failed",
			zap.String("task_id", task.ID),
			zap.Int("index", idx),
			zap.String("proxy_name", result.ProxyInfo.Name),
			zap.Error(err),
		)
	}
	result.Matrices = extractMatrices(matrices, macroMap, onUpdate)
	return result
}

func countMatrixFailures(entries []interfaces.MatrixEntry, result interfaces.EntryResult) int {
	failures := 0
	if strings.TrimSpace(result.Error) != "" {
		failures++
	}
	for _, entry := range entries {
		if isMatrixFailed(entry.Type, result.Matrices) {
			failures++
		}
	}
	return failures
}

func isMatrixFailed(target interfaces.MatrixType, matrices []interfaces.MatrixResult) bool {
	matrix, ok := findMatrixResult(target, matrices)
	if !ok {
		return true
	}

	switch target {
	case interfaces.MatrixRTTPing, interfaces.MatrixHTTPPing, interfaces.MatrixAverageSpeed, interfaces.MatrixMaxSpeed, interfaces.MatrixTrafficUsed:
		payload, ok := decodePayload[struct {
			Value uint64 `json:"value"`
		}](matrix.Payload)
		return !ok || payload.Value == 0
	case interfaces.MatrixPerSecSpeed:
		payload, ok := decodePayload[struct {
			Values []uint64 `json:"values"`
		}](matrix.Payload)
		if !ok || len(payload.Values) == 0 {
			return true
		}
		for _, value := range payload.Values {
			if value > 0 {
				return false
			}
		}
		return true
	case interfaces.MatrixInboundGeoIP, interfaces.MatrixOutboundGeoIP:
		payload, ok := decodePayload[interfaces.GeoIPInfo](matrix.Payload)
		if !ok {
			return true
		}
		if strings.TrimSpace(payload.Error) != "" {
			return true
		}
		return payload.IP == "" && payload.Address == "" && payload.CountryCode == "" && payload.Country == "" && payload.City == "" && payload.ASN == 0 && payload.ASOrganization == "" && payload.ISP == ""
	case interfaces.MatrixMediaUnlock:
		return false
	default:
		return false
	}
}

func findMatrixResult(target interfaces.MatrixType, matrices []interfaces.MatrixResult) (interfaces.MatrixResult, bool) {
	for _, matrix := range matrices {
		if matrix.Type == target {
			return matrix, true
		}
	}
	return interfaces.MatrixResult{}, false
}

func decodePayload[T any](payload any) (T, bool) {
	var value T
	raw, err := json.Marshal(payload)
	if err != nil {
		return value, false
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, false
	}
	return value, true
}

func (p *pollItem) OnExit(exitCode taskpoll.ExitCode) {
	p.exitOnce.Do(func() {
		if p.onExit != nil {
			p.onExit(p, exitCode)
		}
	})
}

func (p *pollItem) Init() taskpoll.Item {
	return p
}
