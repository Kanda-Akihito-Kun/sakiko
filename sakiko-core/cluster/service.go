package cluster

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/cluster/pairing"
	"sakiko.local/sakiko-core/cluster/pki"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"

	"go.uber.org/zap"
)

var (
	knightHeartbeatInterval = 2 * time.Second
	knightHeartbeatTimeout  = 15 * time.Second
)

type MasterEligibilityProber interface {
	ProbeMasterEligibility(ctx context.Context) interfaces.MasterEligibility
}

type KnightAssignmentRunner func(ctx context.Context, assignment interfaces.ClusterAssignment) (interfaces.ResultArchive, error)

type MasterArchiveWriter func(ctx context.Context, archive interfaces.ResultArchive) error

type KnightProgressSnapshotter func(remoteTaskID string) (*interfaces.TaskState, string, error)

type Config struct {
	MasterEligibilityProber   MasterEligibilityProber
	KnightAssignmentRunner    KnightAssignmentRunner
	MasterArchiveWriter       MasterArchiveWriter
	KnightProgressSnapshotter KnightProgressSnapshotter
	StateDir                  string
}

type Service struct {
	prober              MasterEligibilityProber
	pki                 *pki.Service
	pairing             *pairing.Service
	runKnightAssignment KnightAssignmentRunner
	saveMasterArchive   MasterArchiveWriter
	snapshotKnightTask  KnightProgressSnapshotter
	now                 func() time.Time
	stateDir            string

	lock           sync.RWMutex
	status         interfaces.ClusterStatus
	knights        map[string]interfaces.ClusterConnectedKnight
	revokedKnights map[string]string
	assignments    map[string]clusterAssignmentRecord
	remoteTasks    map[string]interfaces.ClusterRemoteTask
	remoteTaskIDs  []string
	masterServer   *http.Server
	knightRuntime  *knightRuntime
}

func New(cfg Config) *Service {
	prober := cfg.MasterEligibilityProber
	if prober == nil {
		prober = defaultMasterEligibilityProber{}
	}

	service := &Service{
		prober:              prober,
		pki:                 pki.New(pki.Config{}),
		pairing:             pairing.New(),
		runKnightAssignment: cfg.KnightAssignmentRunner,
		saveMasterArchive:   cfg.MasterArchiveWriter,
		snapshotKnightTask:  cfg.KnightProgressSnapshotter,
		now:                 time.Now,
		stateDir:            strings.TrimSpace(cfg.StateDir),
		knights:             map[string]interfaces.ClusterConnectedKnight{},
		revokedKnights:      map[string]string{},
		assignments:         map[string]clusterAssignmentRecord{},
		remoteTasks:         map[string]interfaces.ClusterRemoteTask{},
		status: interfaces.ClusterStatus{
			Role: interfaces.ClusterRoleStandalone,
		},
	}

	clusterLogger().Info("cluster service initialized",
		zap.String("role", string(service.status.Role)),
	)
	return service
}

func (s *Service) Status() interfaces.ClusterStatus {
	if s == nil {
		return interfaces.ClusterStatus{Role: interfaces.ClusterRoleStandalone}
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshKnightHealthLocked(s.now().UTC())
	return cloneClusterStatus(s.status)
}

func (s *Service) ProbeMasterEligibility(ctx context.Context) interfaces.MasterEligibility {
	if s == nil {
		return interfaces.MasterEligibility{
			Error: "cluster service is nil",
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	eligibility := s.prober.ProbeMasterEligibility(ctx)
	if strings.TrimSpace(eligibility.CheckedAt) == "" {
		eligibility.CheckedAt = s.now().UTC().Format(time.RFC3339)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	master := cloneMasterStatus(s.status.Master)
	master.Eligibility = eligibility
	s.status.Master = master
	return eligibility
}

func (s *Service) EnableMaster(ctx context.Context, req interfaces.ClusterEnableMasterRequest) (interfaces.ClusterStatus, error) {
	if s == nil {
		return interfaces.ClusterStatus{}, fmt.Errorf("cluster service is nil")
	}

	eligibility := s.ProbeMasterEligibility(ctx)
	if !eligibility.Eligible {
		return s.Status(), fmt.Errorf("master mode is unavailable: %s", strings.TrimSpace(eligibility.Error))
	}
	if req.ListenPort <= 0 {
		return s.Status(), fmt.Errorf("master listen port is required")
	}

	serverName := strings.TrimSpace(req.ListenHost)
	if serverName == "" {
		serverName = strings.TrimSpace(eligibility.PublicIP)
	}
	if _, err := s.pki.EnsureMasterMaterials(serverName); err != nil {
		return s.Status(), err
	}
	if err := s.stopKnightRuntime("role switched to master"); err != nil {
		return s.Status(), err
	}
	if err := s.startMasterServer(strings.TrimSpace(req.ListenHost), req.ListenPort); err != nil {
		return s.Status(), err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.status.Role = interfaces.ClusterRoleMaster
	master := cloneMasterStatus(s.status.Master)
	master.Enabled = true
	master.ListenHost = strings.TrimSpace(req.ListenHost)
	master.ListenPort = req.ListenPort
	master.Eligibility = eligibility
	s.status.Master = master
	s.status.Knight = nil

	clusterLogger().Info("cluster role switched to master",
		zap.String("listen_host", master.ListenHost),
		zap.Int("listen_port", master.ListenPort),
		zap.String("public_ip", eligibility.PublicIP),
		zap.String("server_name", serverName),
	)
	s.persistMasterState(eligibility, master.ListenHost, master.ListenPort)
	return cloneClusterStatus(s.status), nil
}

func (s *Service) EnableKnight(req interfaces.ClusterEnableKnightRequest) (interfaces.ClusterStatus, error) {
	if s == nil {
		return interfaces.ClusterStatus{}, fmt.Errorf("cluster service is nil")
	}

	masterHost := strings.TrimSpace(req.MasterHost)
	oneTimeCode := strings.TrimSpace(req.OneTimeCode)
	if masterHost == "" {
		return s.Status(), fmt.Errorf("master host is required")
	}
	if req.MasterPort <= 0 {
		return s.Status(), fmt.Errorf("master port is required")
	}
	if oneTimeCode == "" {
		return s.Status(), fmt.Errorf("one-time code is required")
	}
	if err := s.stopMasterServer(); err != nil {
		return s.Status(), err
	}
	if err := s.stopKnightRuntime("rebind requested"); err != nil {
		return s.Status(), err
	}

	bootstrap, err := s.bootstrapKnight(context.Background(), masterHost, req.MasterPort, oneTimeCode)
	if err != nil {
		return s.Status(), err
	}
	runtime, err := s.startKnightRuntime(masterHost, req.MasterPort, bootstrap)
	if err != nil {
		return s.Status(), err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.status.Role = interfaces.ClusterRoleKnight
	s.status.Master = nil
	s.status.Knight = &interfaces.ClusterKnightStatus{
		Bound:      true,
		KnightID:   bootstrap.KnightID,
		KnightName: bootstrap.KnightName,
		MasterHost: masterHost,
		MasterPort: req.MasterPort,
		Connected:  true,
		LastSeenAt: s.now().UTC().Format(time.RFC3339),
	}
	s.knightRuntime = runtime

	clusterLogger().Info("cluster role switched to knight",
		zap.String("knight_id", bootstrap.KnightID),
		zap.String("master_host", masterHost),
		zap.Int("master_port", req.MasterPort),
	)
	s.persistKnightState(masterHost, req.MasterPort, bootstrap)
	return cloneClusterStatus(s.status), nil
}

func (s *Service) DisableRemote() interfaces.ClusterStatus {
	if s == nil {
		return interfaces.ClusterStatus{Role: interfaces.ClusterRoleStandalone}
	}

	_ = s.stopKnightRuntime("remote mode disabled")
	_ = s.stopMasterServer()

	s.lock.Lock()
	defer s.lock.Unlock()

	var master *interfaces.ClusterMasterStatus
	if s.status.Master != nil {
		master = cloneMasterStatus(s.status.Master)
		master.Enabled = false
		master.ListenHost = ""
		master.ListenPort = 0
	}

	s.status.Role = interfaces.ClusterRoleStandalone
	s.status.Master = master
	s.status.Knight = nil

	clusterLogger().Info("cluster role switched to standalone")
	s.clearPersistedState()
	return cloneClusterStatus(s.status)
}

func (s *Service) Stop() {
	if s == nil {
		return
	}

	_ = s.stopKnightRuntime("cluster service stopped")
	_ = s.stopMasterServer()
}

func (s *Service) CreatePairingCode(req interfaces.ClusterCreatePairingCodeRequest) (interfaces.ClusterPairingCode, error) {
	if s == nil || s.pairing == nil {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("cluster pairing service is nil")
	}
	if s.Status().Role != interfaces.ClusterRoleMaster {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("pairing code generation requires master mode")
	}
	return s.pairing.CreateCode(req)
}

func (s *Service) BootstrapKnight(req interfaces.ClusterPairingBootstrapRequest) (interfaces.ClusterPairingBootstrapResponse, error) {
	if s == nil || s.pairing == nil || s.pki == nil {
		return interfaces.ClusterPairingBootstrapResponse{}, fmt.Errorf("cluster bootstrap service is not ready")
	}
	if s.Status().Role != interfaces.ClusterRoleMaster {
		return interfaces.ClusterPairingBootstrapResponse{}, fmt.Errorf("knight bootstrap requires master mode")
	}

	code, err := s.pairing.ConsumeCode(req.OneTimeCode)
	if err != nil {
		return interfaces.ClusterPairingBootstrapResponse{}, err
	}

	knightID := strings.TrimSpace(req.KnightID)
	if knightID == "" {
		knightID = randomID()
	}
	knightName := strings.TrimSpace(req.KnightName)
	if knightName == "" {
		knightName = strings.TrimSpace(code.KnightName)
	}

	issued, err := s.pki.IssueKnightCertificate(req.CSRPEM, knightID, knightName)
	if err != nil {
		return interfaces.ClusterPairingBootstrapResponse{}, err
	}

	s.upsertKnight(interfaces.ClusterConnectedKnight{
		KnightID:   issued.KnightID,
		KnightName: issued.KnightName,
		State:      interfaces.ClusterKnightStatePaired,
	})

	return interfaces.ClusterPairingBootstrapResponse{
		KnightID:             issued.KnightID,
		KnightName:           issued.KnightName,
		ClientCertificatePEM: issued.ClientCertificatePEM,
		CACertificatePEM:     issued.CACertificatePEM,
		MasterServerName:     issued.MasterServerName,
		Status:               s.Status(),
	}, nil
}

func (s *Service) ListKnights() []interfaces.ClusterConnectedKnight {
	if s == nil {
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshKnightHealthLocked(s.now().UTC())

	knights := make([]interfaces.ClusterConnectedKnight, 0, len(s.knights))
	for _, knight := range s.knights {
		knights = append(knights, knight)
	}
	return knights
}

func (s *Service) KickKnight(req interfaces.ClusterKickKnightRequest) (interfaces.ClusterStatus, error) {
	if s == nil {
		return interfaces.ClusterStatus{}, fmt.Errorf("cluster service is nil")
	}
	if s.Status().Role != interfaces.ClusterRoleMaster {
		return s.Status(), fmt.Errorf("kick knight requires master mode")
	}

	knightID := strings.TrimSpace(req.KnightID)
	if knightID == "" {
		return s.Status(), fmt.Errorf("knight ID is required")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	knight, ok := s.knights[knightID]
	if !ok {
		return cloneClusterStatus(s.status), fmt.Errorf("knight not found: %s", knightID)
	}

	revokedAt := s.now().UTC().Format(time.RFC3339)
	reason := "kicked by master"
	knight.State = interfaces.ClusterKnightStateRevoked
	knight.LastSeenAt = revokedAt
	knight.LastError = reason
	s.revokedKnights[knightID] = reason
	delete(s.knights, knightID)

	for assignmentID, record := range s.assignments {
		if record.assignment.KnightID != knightID {
			continue
		}
		delete(s.assignments, assignmentID)
	}

	for remoteTaskID, remoteTask := range s.remoteTasks {
		if remoteTask.KnightID != knightID {
			continue
		}
		switch remoteTask.State {
		case interfaces.ClusterRemoteTaskFinished, interfaces.ClusterRemoteTaskFailed:
			continue
		default:
			remoteTask.State = interfaces.ClusterRemoteTaskFailed
			remoteTask.Error = reason
			remoteTask.FinishedAt = revokedAt
			s.remoteTasks[remoteTaskID] = remoteTask
		}
	}

	clusterLogger().Info("knight kicked by master",
		zap.String("knight_id", knightID),
		zap.String("knight_name", knight.KnightName),
	)
	s.persistCurrentStateLocked()
	return cloneClusterStatus(s.status), nil
}

func (s *Service) refreshKnightHealthLocked(now time.Time) {
	if s == nil {
		return
	}

	staleBefore := now.Add(-knightHeartbeatTimeout)

	for knightID, knight := range s.knights {
		if !isKnightHeartbeatStale(knight, staleBefore) {
			continue
		}

		knight.State = interfaces.ClusterKnightStateOffline
		if strings.TrimSpace(knight.LastError) == "" {
			knight.LastError = "knight heartbeat timed out"
		}
		s.knights[knightID] = knight
	}
}

func isKnightHeartbeatStale(knight interfaces.ClusterConnectedKnight, staleBefore time.Time) bool {
	switch knight.State {
	case interfaces.ClusterKnightStateOffline, interfaces.ClusterKnightStateRevoked, interfaces.ClusterKnightStatePaired:
		return false
	}

	lastSeenAt := strings.TrimSpace(knight.LastSeenAt)
	if lastSeenAt == "" {
		return false
	}

	parsed, err := time.Parse(time.RFC3339, lastSeenAt)
	if err != nil {
		return false
	}
	return parsed.Before(staleBefore)
}

func cloneClusterStatus(status interfaces.ClusterStatus) interfaces.ClusterStatus {
	cloned := status
	if status.Master != nil {
		cloned.Master = cloneMasterStatus(status.Master)
	}
	if status.Knight != nil {
		knight := *status.Knight
		cloned.Knight = &knight
	}
	return cloned
}

func cloneMasterStatus(master *interfaces.ClusterMasterStatus) *interfaces.ClusterMasterStatus {
	if master == nil {
		return &interfaces.ClusterMasterStatus{}
	}
	cloned := *master
	return &cloned
}

func clusterLogger() *zap.Logger {
	return logx.Named("core.cluster")
}

func randomID() string {
	var raw [12]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}
