package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sakiko.local/sakiko-core/cluster/pki"
	"sakiko.local/sakiko-core/interfaces"
)

const clusterStateFileName = "cluster-state.json"

type persistedClusterState struct {
	Role   interfaces.ClusterRole `json:"role"`
	Master *persistedMasterState  `json:"master,omitempty"`
	Knight *persistedKnightState  `json:"knight,omitempty"`
}

type persistedMasterState struct {
	ListenHost     string                                  `json:"listenHost,omitempty"`
	ListenPort     int                                     `json:"listenPort,omitempty"`
	Eligibility    interfaces.MasterEligibility            `json:"eligibility,omitempty"`
	Materials      pki.MasterMaterials                     `json:"materials"`
	RevokedKnights map[string]string                       `json:"revokedKnights,omitempty"`
	Assignments    map[string]persistedAssignmentRecord    `json:"assignments,omitempty"`
	RemoteTasks    map[string]interfaces.ClusterRemoteTask `json:"remoteTasks,omitempty"`
	RemoteTaskIDs  []string                                `json:"remoteTaskIds,omitempty"`
}

type persistedKnightState struct {
	MasterHost string                `json:"masterHost"`
	MasterPort int                   `json:"masterPort"`
	Bootstrap  knightBootstrapBundle `json:"bootstrap"`
}

type persistedAssignmentRecord struct {
	Assignment interfaces.ClusterAssignment `json:"assignment"`
	LeasedAt   string                       `json:"leasedAt,omitempty"`
}

func (s *Service) Restore() error {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return nil
	}

	state, err := s.loadPersistedState()
	if err != nil || state == nil {
		return err
	}

	switch state.Role {
	case interfaces.ClusterRoleMaster:
		if state.Master == nil {
			return fmt.Errorf("persisted master state is missing")
		}
		if err := s.pki.LoadMasterMaterials(state.Master.Materials); err != nil {
			return err
		}
		if err := s.startMasterServer(state.Master.ListenHost, state.Master.ListenPort); err != nil {
			return err
		}

		s.lock.Lock()
		s.status.Role = interfaces.ClusterRoleMaster
		s.status.Knight = nil
		s.status.Master = &interfaces.ClusterMasterStatus{
			Enabled:     true,
			ListenHost:  state.Master.ListenHost,
			ListenPort:  state.Master.ListenPort,
			Eligibility: state.Master.Eligibility,
		}
		s.revokedKnights = cloneStringMap(state.Master.RevokedKnights)
		s.assignments = restoreAssignments(state.Master.Assignments)
		s.remoteTasks = cloneRemoteTasks(state.Master.RemoteTasks)
		s.remoteTaskIDs = restoreRemoteTaskIDs(state.Master.RemoteTaskIDs, s.remoteTasks)
		s.lock.Unlock()
		return nil

	case interfaces.ClusterRoleKnight:
		if state.Knight == nil {
			return fmt.Errorf("persisted knight state is missing")
		}
		runtime, err := s.startKnightRuntime(state.Knight.MasterHost, state.Knight.MasterPort, state.Knight.Bootstrap)
		if err != nil {
			return err
		}

		s.lock.Lock()
		s.status.Role = interfaces.ClusterRoleKnight
		s.status.Master = nil
		s.status.Knight = &interfaces.ClusterKnightStatus{
			Bound:      true,
			KnightID:   state.Knight.Bootstrap.KnightID,
			KnightName: state.Knight.Bootstrap.KnightName,
			MasterHost: state.Knight.MasterHost,
			MasterPort: state.Knight.MasterPort,
			Connected:  true,
			LastSeenAt: s.now().UTC().Format(time.RFC3339),
		}
		s.knightRuntime = runtime
		s.lock.Unlock()
		return nil
	}

	return nil
}

func (s *Service) persistMasterState(eligibility interfaces.MasterEligibility, listenHost string, listenPort int) {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return
	}

	materials, ok := s.pki.CurrentMasterMaterials()
	if !ok {
		return
	}

	_ = s.savePersistedState(persistedClusterState{
		Role: interfaces.ClusterRoleMaster,
		Master: &persistedMasterState{
			ListenHost:     strings.TrimSpace(listenHost),
			ListenPort:     listenPort,
			Eligibility:    eligibility,
			Materials:      materials,
			RevokedKnights: cloneStringMap(s.revokedKnights),
			Assignments:    persistAssignments(s.assignments),
			RemoteTasks:    cloneRemoteTasks(s.remoteTasks),
			RemoteTaskIDs:  append([]string{}, s.remoteTaskIDs...),
		},
	})
}

func (s *Service) persistCurrentStateLocked() {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return
	}

	switch s.status.Role {
	case interfaces.ClusterRoleMaster:
		if s.status.Master == nil {
			return
		}
		s.persistMasterState(s.status.Master.Eligibility, s.status.Master.ListenHost, s.status.Master.ListenPort)
	}
}

func (s *Service) persistKnightState(masterHost string, masterPort int, bootstrap knightBootstrapBundle) {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return
	}

	_ = s.savePersistedState(persistedClusterState{
		Role: interfaces.ClusterRoleKnight,
		Knight: &persistedKnightState{
			MasterHost: strings.TrimSpace(masterHost),
			MasterPort: masterPort,
			Bootstrap:  bootstrap,
		},
	})
}

func (s *Service) clearPersistedState() {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return
	}

	_ = os.Remove(s.statePath())
}

func (s *Service) loadPersistedState() (*persistedClusterState, error) {
	raw, err := os.ReadFile(s.statePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var state persistedClusterState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Service) savePersistedState(state persistedClusterState) error {
	if err := os.MkdirAll(s.stateDir, 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.statePath(), raw, 0o644)
}

func (s *Service) statePath() string {
	if s == nil || strings.TrimSpace(s.stateDir) == "" {
		return ""
	}
	return filepath.Join(s.stateDir, clusterStateFileName)
}

func persistAssignments(records map[string]clusterAssignmentRecord) map[string]persistedAssignmentRecord {
	if len(records) == 0 {
		return nil
	}

	out := make(map[string]persistedAssignmentRecord, len(records))
	for id, record := range records {
		assignment := record.assignment
		if assignment.Task != nil {
			taskCopy := cloneTask(*assignment.Task)
			assignment.Task = &taskCopy
		}
		out[id] = persistedAssignmentRecord{
			Assignment: assignment,
			LeasedAt:   record.leasedAt,
		}
	}
	return out
}

func restoreAssignments(records map[string]persistedAssignmentRecord) map[string]clusterAssignmentRecord {
	out := make(map[string]clusterAssignmentRecord, len(records))
	for id, record := range records {
		assignment := record.Assignment
		if assignment.Task != nil {
			taskCopy := cloneTask(*assignment.Task)
			assignment.Task = &taskCopy
		}
		out[id] = clusterAssignmentRecord{
			assignment: assignment,
			leasedAt:   record.LeasedAt,
		}
	}
	return out
}

func cloneRemoteTasks(tasks map[string]interfaces.ClusterRemoteTask) map[string]interfaces.ClusterRemoteTask {
	out := make(map[string]interfaces.ClusterRemoteTask, len(tasks))
	for id, task := range tasks {
		taskCopy := task
		if task.Runtime != nil {
			runtime := *task.Runtime
			runtime.ActiveNodes = append([]interfaces.TaskActiveNode{}, task.Runtime.ActiveNodes...)
			taskCopy.Runtime = &runtime
		}
		out[id] = taskCopy
	}
	return out
}

func restoreRemoteTaskIDs(ids []string, tasks map[string]interfaces.ClusterRemoteTask) []string {
	if len(ids) > 0 {
		return append([]string{}, ids...)
	}
	out := make([]string, 0, len(tasks))
	for id := range tasks {
		out = append(out, id)
	}
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
