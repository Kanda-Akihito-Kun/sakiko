package cluster

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

type stubMasterEligibilityProber struct {
	eligibility interfaces.MasterEligibility
}

func (s stubMasterEligibilityProber) ProbeMasterEligibility(ctx context.Context) interfaces.MasterEligibility {
	_ = ctx
	return s.eligibility
}

func TestIsPublicIPAddress(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "public v4", ip: "8.8.8.8", want: true},
		{name: "private v4", ip: "192.168.1.10", want: false},
		{name: "loopback", ip: "127.0.0.1", want: false},
		{name: "public v6", ip: "2606:4700:4700::1111", want: true},
		{name: "link local v6", ip: "fe80::1", want: false},
		{name: "invalid", ip: "not-an-ip", want: false},
	}

	for _, tc := range cases {
		if got := isPublicIPAddress(tc.ip); got != tc.want {
			t.Fatalf("%s: isPublicIPAddress(%q) = %v, want %v", tc.name, tc.ip, got, tc.want)
		}
	}
}

func TestEvaluateMasterEligibilityAllowsPublicServerAddress(t *testing.T) {
	originalHasDirectPublicIP := hasDirectPublicIPFunc
	originalProbeLocalUDPNAT := probeLocalUDPNATFunc
	originalProbePublicListenerReachable := probePublicListenerReachableFunc
	hasDirectPublicIPFunc = func(publicIP string) bool {
		return publicIP == "45.95.212.179"
	}
	probeLocalUDPNATFunc = func(ctx context.Context) interfaces.UDPNATInfo {
		t.Fatalf("probeLocalUDPNATFunc should not be called for direct public IP")
		return interfaces.UDPNATInfo{}
	}
	defer func() {
		hasDirectPublicIPFunc = originalHasDirectPublicIP
		probeLocalUDPNATFunc = originalProbeLocalUDPNAT
		probePublicListenerReachableFunc = originalProbePublicListenerReachable
	}()

	checkedAt := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)

	eligibility := evaluateMasterEligibility(context.Background(), interfaces.BackendInfo{
		IP: "45.95.212.179",
	}, checkedAt)

	if !eligibility.HasPublicIP {
		t.Fatalf("expected public IP to be accepted, got %+v", eligibility)
	}
	if eligibility.NATType != natTypeNAT1 {
		t.Fatalf("expected NAT type %q, got %q", natTypeNAT1, eligibility.NATType)
	}
	if !eligibility.IsNAT1 {
		t.Fatalf("expected direct public server to be treated as nat1-capable, got %+v", eligibility)
	}
	if !eligibility.Reachable {
		t.Fatalf("expected public server to be treated as reachable, got %+v", eligibility)
	}
	if !eligibility.Eligible {
		t.Fatalf("expected public server to be eligible, got %+v", eligibility)
	}
	if eligibility.Error != "" {
		t.Fatalf("expected no eligibility error, got %q", eligibility.Error)
	}
	if eligibility.CheckedAt != checkedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected checkedAt %q", eligibility.CheckedAt)
	}
}

func TestEvaluateMasterEligibilityTreatsFullConePublicServerAsNAT1(t *testing.T) {
	originalHasDirectPublicIP := hasDirectPublicIPFunc
	originalProbeLocalUDPNAT := probeLocalUDPNATFunc
	originalProbePublicListenerReachable := probePublicListenerReachableFunc
	hasDirectPublicIPFunc = func(publicIP string) bool {
		return false
	}
	probeLocalUDPNATFunc = func(ctx context.Context) interfaces.UDPNATInfo {
		return interfaces.UDPNATInfo{
			Type:         interfaces.UDPNATTypeFullCone,
			InternalIP:   "10.0.0.2",
			InternalPort: 41000,
			PublicIP:     "203.0.113.20",
			PublicPort:   41000,
		}
	}
	probePublicListenerReachableFunc = func(ctx context.Context, publicIP string) (bool, error) {
		if publicIP != "203.0.113.20" {
			t.Fatalf("unexpected public IP %q", publicIP)
		}
		return true, nil
	}
	defer func() {
		hasDirectPublicIPFunc = originalHasDirectPublicIP
		probeLocalUDPNATFunc = originalProbeLocalUDPNAT
		probePublicListenerReachableFunc = originalProbePublicListenerReachable
	}()

	eligibility := evaluateMasterEligibility(context.Background(), interfaces.BackendInfo{
		IP: "203.0.113.20",
	}, time.Now().UTC())

	if eligibility.NATType != natTypeNAT1 {
		t.Fatalf("expected full-cone public server to be upgraded to %q, got %+v", natTypeNAT1, eligibility)
	}
	if !eligibility.Eligible || !eligibility.IsNAT1 || !eligibility.Reachable {
		t.Fatalf("expected full-cone public server to be eligible, got %+v", eligibility)
	}
}

func TestEvaluateMasterEligibilityRejectsPrivateAddress(t *testing.T) {
	eligibility := evaluateMasterEligibility(context.Background(), interfaces.BackendInfo{
		IP: "192.168.1.10",
	}, time.Now().UTC())

	if eligibility.Eligible {
		t.Fatalf("expected private address to remain ineligible, got %+v", eligibility)
	}
	if eligibility.Error == "" {
		t.Fatalf("expected private address eligibility error, got %+v", eligibility)
	}
}

func TestEvaluateMasterEligibilityRejectsNAT4Network(t *testing.T) {
	originalHasDirectPublicIP := hasDirectPublicIPFunc
	originalProbeLocalUDPNAT := probeLocalUDPNATFunc
	originalProbePublicListenerReachable := probePublicListenerReachableFunc
	hasDirectPublicIPFunc = func(publicIP string) bool {
		return false
	}
	probeLocalUDPNATFunc = func(ctx context.Context) interfaces.UDPNATInfo {
		return interfaces.UDPNATInfo{
			Type:     interfaces.UDPNATTypeSymmetric,
			PublicIP: "203.0.113.20",
		}
	}
	defer func() {
		hasDirectPublicIPFunc = originalHasDirectPublicIP
		probeLocalUDPNATFunc = originalProbeLocalUDPNAT
		probePublicListenerReachableFunc = originalProbePublicListenerReachable
	}()

	eligibility := evaluateMasterEligibility(context.Background(), interfaces.BackendInfo{
		IP: "203.0.113.20",
	}, time.Now().UTC())

	if eligibility.Eligible {
		t.Fatalf("expected NAT4 network to remain ineligible, got %+v", eligibility)
	}
	if eligibility.NATType != natTypeNAT4 {
		t.Fatalf("expected NAT type %q, got %q", natTypeNAT4, eligibility.NATType)
	}
	if eligibility.IsNAT1 {
		t.Fatalf("expected NAT4 network to not be treated as nat1, got %+v", eligibility)
	}
}

func TestEnableMasterRequiresEligibleProbe(t *testing.T) {
	service := New(Config{
		MasterEligibilityProber: stubMasterEligibilityProber{
			eligibility: interfaces.MasterEligibility{
				PublicIP:    "8.8.8.8",
				HasPublicIP: true,
				NATType:     "nat1",
				IsNAT1:      true,
				Reachable:   true,
				Eligible:    true,
			},
		},
	})

	port := allocateLocalPort(t)
	status, err := service.EnableMaster(context.Background(), interfaces.ClusterEnableMasterRequest{
		ListenHost: "127.0.0.1",
		ListenPort: port,
	})
	if err != nil {
		t.Fatalf("EnableMaster() error = %v", err)
	}
	if status.Role != interfaces.ClusterRoleMaster {
		t.Fatalf("expected master role, got %q", status.Role)
	}
	if status.Master == nil || !status.Master.Enabled {
		t.Fatalf("expected master status to be enabled, got %+v", status.Master)
	}
}

func TestEnableKnightStoresBindingState(t *testing.T) {
	master, port, code := startTestMaster(t)
	defer master.DisableRemote()

	service := New(Config{})

	status, err := service.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  "127.0.0.1",
		MasterPort:  port,
		OneTimeCode: code,
	})
	if err != nil {
		t.Fatalf("EnableKnight() error = %v", err)
	}
	if status.Role != interfaces.ClusterRoleKnight {
		t.Fatalf("expected knight role, got %q", status.Role)
	}
	if status.Knight == nil || !status.Knight.Bound {
		t.Fatalf("expected knight to be bound, got %+v", status.Knight)
	}
}

func TestDisableRemoteResetsRoleToStandalone(t *testing.T) {
	master, port, code := startTestMaster(t)
	defer master.DisableRemote()

	service := New(Config{})
	if _, err := service.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  "127.0.0.1",
		MasterPort:  port,
		OneTimeCode: code,
	}); err != nil {
		t.Fatalf("EnableKnight() error = %v", err)
	}

	status := service.DisableRemote()
	if status.Role != interfaces.ClusterRoleStandalone {
		t.Fatalf("expected standalone role, got %q", status.Role)
	}
	if status.Knight != nil {
		t.Fatalf("expected knight state to be cleared, got %+v", status.Knight)
	}
}

func TestMasterBootstrapIssuesKnightCertificate(t *testing.T) {
	service := New(Config{
		MasterEligibilityProber: stubMasterEligibilityProber{
			eligibility: interfaces.MasterEligibility{
				PublicIP:    "8.8.8.8",
				HasPublicIP: true,
				NATType:     "nat1",
				IsNAT1:      true,
				Reachable:   true,
				Eligible:    true,
			},
		},
	})

	port := allocateLocalPort(t)
	if _, err := service.EnableMaster(context.Background(), interfaces.ClusterEnableMasterRequest{
		ListenHost: "127.0.0.1",
		ListenPort: port,
	}); err != nil {
		t.Fatalf("EnableMaster() error = %v", err)
	}
	defer service.DisableRemote()

	code, err := service.CreatePairingCode(interfaces.ClusterCreatePairingCodeRequest{
		KnightName: "Knight-Tokyo",
	})
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	resp, err := service.BootstrapKnight(interfaces.ClusterPairingBootstrapRequest{
		OneTimeCode: code.Code,
		KnightID:    "knight-1",
		CSRPEM:      generateCSRPEM(t, "knight-1"),
	})
	if err != nil {
		t.Fatalf("BootstrapKnight() error = %v", err)
	}
	if resp.KnightID != "knight-1" {
		t.Fatalf("expected knight ID knight-1, got %q", resp.KnightID)
	}
	if resp.ClientCertificatePEM == "" || resp.CACertificatePEM == "" {
		t.Fatalf("expected certificate bundle, got %+v", resp)
	}
}

func TestKnightRegistersWithMasterOverMTLS(t *testing.T) {
	master, port, code := startTestMaster(t)
	defer master.DisableRemote()

	knight := New(Config{})
	status, err := knight.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  "127.0.0.1",
		MasterPort:  port,
		OneTimeCode: code,
	})
	if err != nil {
		t.Fatalf("EnableKnight() error = %v", err)
	}
	defer knight.DisableRemote()

	if status.Knight == nil || !status.Knight.Connected {
		t.Fatalf("expected knight to connect immediately, got %+v", status.Knight)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		knights := master.ListKnights()
		if len(knights) == 1 && knights[0].State == interfaces.ClusterKnightStateOnline {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("expected master to observe connected knight, got %+v", master.ListKnights())
}

func TestKnightKeepsHeartbeatWhileExecutingAssignment(t *testing.T) {
	originalHeartbeatInterval := knightHeartbeatInterval
	originalHeartbeatTimeout := knightHeartbeatTimeout
	knightHeartbeatInterval = 200 * time.Millisecond
	knightHeartbeatTimeout = 1500 * time.Millisecond
	defer func() {
		knightHeartbeatInterval = originalHeartbeatInterval
		knightHeartbeatTimeout = originalHeartbeatTimeout
	}()

	master, port, code := startTestMaster(t)
	defer master.DisableRemote()

	knight := New(Config{
		KnightAssignmentRunner: func(ctx context.Context, assignment interfaces.ClusterAssignment) (interfaces.ResultArchive, error) {
			_ = ctx
			time.Sleep(knightHeartbeatTimeout + (2 * knightHeartbeatInterval))
			return interfaces.ResultArchive{
				Task: interfaces.ResultArchiveTask{
					ID:     assignment.RemoteTaskID,
					Name:   assignment.TaskName,
					Vendor: interfaces.VendorLocal,
				},
				State: interfaces.TaskState{
					TaskID: assignment.RemoteTaskID,
					Name:   assignment.TaskName,
					Status: "success",
				},
				ExitCode: "0",
			}, nil
		},
	})
	defer knight.DisableRemote()

	if _, err := knight.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  "127.0.0.1",
		MasterPort:  port,
		OneTimeCode: code,
	}); err != nil {
		t.Fatalf("EnableKnight() error = %v", err)
	}

	knightStatus := knight.Status()
	if knightStatus.Knight == nil || knightStatus.Knight.KnightID == "" {
		t.Fatalf("expected knight binding state, got %+v", knightStatus.Knight)
	}

	observedOnline := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		knights := master.ListKnights()
		if len(knights) == 1 && knights[0].State == interfaces.ClusterKnightStateOnline {
			observedOnline = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !observedOnline {
		t.Fatalf("expected master to observe connected knight before dispatch, got %+v", master.ListKnights())
	}

	tasks, err := master.DispatchTask(interfaces.ClusterDispatchTaskRequest{
		KnightIDs: []string{knightStatus.Knight.KnightID},
		Task: interfaces.Task{
			Name:   "Remote heartbeat test",
			Vendor: interfaces.VendorLocal,
			Nodes: []interfaces.Node{
				{Name: "node-1", Enabled: true},
			},
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixHTTPPing},
			},
			Config: interfaces.TaskConfig{}.Normalize(),
		},
	})
	if err != nil {
		t.Fatalf("DispatchTask() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one remote task, got %+v", tasks)
	}

	time.Sleep(knightHeartbeatTimeout + (2 * knightHeartbeatInterval))

	knights := master.ListKnights()
	if len(knights) != 1 {
		t.Fatalf("expected one connected knight, got %+v", knights)
	}
	if knights[0].State == interfaces.ClusterKnightStateOffline {
		t.Fatalf("expected busy knight to keep heartbeating, got %+v", knights[0])
	}

	deadline = time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		remoteTasks := master.ListRemoteTasks()
		if len(remoteTasks) == 1 && remoteTasks[0].State == interfaces.ClusterRemoteTaskFinished {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("expected remote task to finish after long-running execution, got %+v", master.ListRemoteTasks())
}

func TestListKnightsMarksStaleKnightOffline(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	service := New(Config{})
	service.now = func() time.Time { return now }
	service.knights["knight-stale"] = interfaces.ClusterConnectedKnight{
		KnightID:   "knight-stale",
		KnightName: "Knight-Stale",
		State:      interfaces.ClusterKnightStateOnline,
		LastSeenAt: now.Add(-knightHeartbeatTimeout - time.Second).Format(time.RFC3339),
	}

	knights := service.ListKnights()
	if len(knights) != 1 {
		t.Fatalf("expected one knight, got %+v", knights)
	}
	if knights[0].State != interfaces.ClusterKnightStateOffline {
		t.Fatalf("expected stale knight to be marked offline, got %+v", knights[0])
	}
	if knights[0].LastError == "" {
		t.Fatalf("expected stale knight to record heartbeat timeout, got %+v", knights[0])
	}
}

func TestListRemoteTasksKeepsRunningTaskWhenKnightTimesOut(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	service := New(Config{})
	service.now = func() time.Time { return now }
	service.knights["knight-timeout"] = interfaces.ClusterConnectedKnight{
		KnightID:   "knight-timeout",
		KnightName: "Knight-Timeout",
		State:      interfaces.ClusterKnightStateBusy,
		LastSeenAt: now.Add(-knightHeartbeatTimeout - time.Second).Format(time.RFC3339),
	}
	service.remoteTaskIDs = []string{"remote-task-1"}
	service.remoteTasks["remote-task-1"] = interfaces.ClusterRemoteTask{
		AssignmentID: "assignment-1",
		RemoteTaskID: "remote-task-1",
		KnightID:     "knight-timeout",
		State:        interfaces.ClusterRemoteTaskRunning,
		StartedAt:    now.Add(-2 * time.Minute).Format(time.RFC3339),
	}
	service.assignments["assignment-1"] = clusterAssignmentRecord{
		assignment: interfaces.ClusterAssignment{
			AssignmentID: "assignment-1",
			RemoteTaskID: "remote-task-1",
			KnightID:     "knight-timeout",
		},
	}

	tasks := service.ListRemoteTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected one remote task, got %+v", tasks)
	}
	if tasks[0].State != interfaces.ClusterRemoteTaskRunning {
		t.Fatalf("expected running task to remain pending when knight times out, got %+v", tasks[0])
	}
	if tasks[0].FinishedAt != "" {
		t.Fatalf("expected timed-out running task to remain unfinished, got %+v", tasks[0])
	}
	if _, ok := service.assignments["assignment-1"]; !ok {
		t.Fatalf("expected assignment lease to remain while waiting for knight recovery")
	}
}

func TestKickKnightRevokesKnightAndFailsRemoteTasks(t *testing.T) {
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	service := New(Config{})
	service.now = func() time.Time { return now }
	service.status.Role = interfaces.ClusterRoleMaster
	service.knights["knight-1"] = interfaces.ClusterConnectedKnight{
		KnightID:   "knight-1",
		KnightName: "Tokyo",
		State:      interfaces.ClusterKnightStateOnline,
	}
	service.remoteTaskIDs = []string{"remote-task-1"}
	service.remoteTasks["remote-task-1"] = interfaces.ClusterRemoteTask{
		AssignmentID: "assignment-1",
		RemoteTaskID: "remote-task-1",
		KnightID:     "knight-1",
		KnightName:   "Tokyo",
		State:        interfaces.ClusterRemoteTaskRunning,
	}
	service.assignments["assignment-1"] = clusterAssignmentRecord{
		assignment: interfaces.ClusterAssignment{
			AssignmentID: "assignment-1",
			RemoteTaskID: "remote-task-1",
			KnightID:     "knight-1",
		},
	}

	status, err := service.KickKnight(interfaces.ClusterKickKnightRequest{KnightID: "knight-1"})
	if err != nil {
		t.Fatalf("KickKnight() error = %v", err)
	}
	if status.Role != interfaces.ClusterRoleMaster {
		t.Fatalf("expected master role to remain unchanged, got %+v", status)
	}
	if len(service.ListKnights()) != 0 {
		t.Fatalf("expected kicked knight to be removed from connected list, got %+v", service.ListKnights())
	}
	if !service.isKnightRevoked("knight-1") {
		t.Fatalf("expected knight to be marked revoked")
	}
	if _, ok := service.assignments["assignment-1"]; ok {
		t.Fatalf("expected kicked knight assignment to be removed")
	}

	task := service.remoteTasks["remote-task-1"]
	if task.State != interfaces.ClusterRemoteTaskFailed {
		t.Fatalf("expected kicked knight task to fail, got %+v", task)
	}
	if task.Error != "kicked by master" {
		t.Fatalf("expected kicked-by-master error, got %+v", task)
	}
}

func TestMasterPersistenceRestoresRevokedKnightsAndRemoteTasks(t *testing.T) {
	stateDir := t.TempDir()
	service := New(Config{
		StateDir: stateDir,
		MasterEligibilityProber: stubMasterEligibilityProber{
			eligibility: interfaces.MasterEligibility{
				PublicIP:    "45.95.212.179",
				HasPublicIP: true,
				NATType:     natTypeNAT1,
				IsNAT1:      true,
				Reachable:   true,
				Eligible:    true,
			},
		},
	})

	port := allocateLocalPort(t)
	if _, err := service.EnableMaster(context.Background(), interfaces.ClusterEnableMasterRequest{
		ListenHost: "127.0.0.1",
		ListenPort: port,
	}); err != nil {
		t.Fatalf("EnableMaster() error = %v", err)
	}
	defer service.DisableRemote()

	service.lock.Lock()
	service.knights["knight-1"] = interfaces.ClusterConnectedKnight{
		KnightID:   "knight-1",
		KnightName: "Tokyo",
		State:      interfaces.ClusterKnightStateOnline,
	}
	service.lock.Unlock()

	created, err := service.DispatchTask(interfaces.ClusterDispatchTaskRequest{
		KnightIDs: []string{"knight-1"},
		Task: interfaces.Task{
			Name:   "Persisted remote task",
			Vendor: interfaces.VendorLocal,
			Nodes: []interfaces.Node{
				{Name: "node-1", Enabled: true},
			},
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixHTTPPing},
			},
			Config: interfaces.TaskConfig{}.Normalize(),
		},
	})
	if err != nil {
		t.Fatalf("DispatchTask() error = %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected one remote task, got %+v", created)
	}

	if assignment := service.leaseAssignmentForKnight("knight-1"); assignment == nil {
		t.Fatalf("expected assignment to be leased")
	}
	if _, err := service.KickKnight(interfaces.ClusterKickKnightRequest{KnightID: "knight-1"}); err != nil {
		t.Fatalf("KickKnight() error = %v", err)
	}
	if err := service.stopMasterServer(); err != nil {
		t.Fatalf("stopMasterServer() error = %v", err)
	}

	restored := New(Config{StateDir: stateDir})
	if err := restored.Restore(); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	defer restored.DisableRemote()

	if !restored.isKnightRevoked("knight-1") {
		t.Fatalf("expected revoked knight to be restored")
	}
	tasks := restored.ListRemoteTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected one restored remote task, got %+v", tasks)
	}
	if tasks[0].RemoteTaskID != created[0].RemoteTaskID {
		t.Fatalf("expected restored remote task ID %q, got %q", created[0].RemoteTaskID, tasks[0].RemoteTaskID)
	}
	if tasks[0].State != interfaces.ClusterRemoteTaskFailed {
		t.Fatalf("expected restored task to remain failed, got %+v", tasks[0])
	}
	if tasks[0].Error != "kicked by master" {
		t.Fatalf("expected restored task error, got %+v", tasks[0])
	}
}

func generateCSRPEM(t *testing.T, commonName string) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: commonName},
	}, key)
	if err != nil {
		t.Fatalf("CreateCertificateRequest() error = %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	}))
}

func allocateLocalPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func startTestMaster(t *testing.T) (*Service, int, string) {
	t.Helper()

	service := New(Config{
		MasterEligibilityProber: stubMasterEligibilityProber{
			eligibility: interfaces.MasterEligibility{
				PublicIP:    "45.95.212.179",
				HasPublicIP: true,
				NATType:     natTypeNAT1,
				IsNAT1:      true,
				Reachable:   true,
				Eligible:    true,
			},
		},
	})

	port := allocateLocalPort(t)
	if _, err := service.EnableMaster(context.Background(), interfaces.ClusterEnableMasterRequest{
		ListenHost: "127.0.0.1",
		ListenPort: port,
	}); err != nil {
		t.Fatalf("EnableMaster() error = %v", err)
	}

	code, err := service.CreatePairingCode(interfaces.ClusterCreatePairingCodeRequest{
		KnightName: "Knight-Test",
	})
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	return service, port, code.Code
}
