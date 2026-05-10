package api

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"sakiko.local/sakiko-core/cluster"
	"sakiko.local/sakiko-core/interfaces"
)

func TestSubmitTaskRejectsStandaloneRequestInKnightMode(t *testing.T) {
	service, err := New(Config{
		Mode:         interfaces.ModeParallel,
		ProfilesPath: filepath.Join(t.TempDir(), "profiles.yaml"),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	master := clusterTestMasterForAPI(t)
	defer master.service.DisableRemote()

	if _, err := service.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  "127.0.0.1",
		MasterPort:  master.port,
		OneTimeCode: master.code,
	}); err != nil {
		t.Fatalf("EnableKnight() error = %v", err)
	}

	_, err = service.SubmitTask(interfaces.TaskSubmitRequest{}, nil)
	if err == nil {
		t.Fatalf("expected knight-mode submit rejection, got nil")
	}
	if !strings.Contains(err.Error(), "disabled while running as knight") {
		t.Fatalf("expected knight-mode rejection error, got %v", err)
	}
}

type apiTestMaster struct {
	service *Service
	port    int
	code    string
}

func clusterTestMasterForAPI(t *testing.T) apiTestMaster {
	t.Helper()

	service, err := New(Config{
		Mode:         interfaces.ModeParallel,
		ProfilesPath: filepath.Join(t.TempDir(), "master-profiles.yaml"),
	})
	if err != nil {
		t.Fatalf("New() master error = %v", err)
	}

	port := allocateAPITestPort(t)
	service.cluster = clusterTestEligibleService()
	if _, err := service.EnableMaster(interfaces.ClusterEnableMasterRequest{
		ListenHost: "127.0.0.1",
		ListenPort: port,
	}); err != nil {
		t.Fatalf("EnableMaster() error = %v", err)
	}

	code, err := service.CreatePairingCode(interfaces.ClusterCreatePairingCodeRequest{
		KnightName: "Knight-API",
	})
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	return apiTestMaster{
		service: service,
		port:    port,
		code:    code.PairingCode.Code,
	}
}

func allocateAPITestPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func clusterTestEligibleService() *cluster.Service {
	return cluster.New(cluster.Config{
		MasterEligibilityProber: clusterStubMasterEligibilityProber{
			eligibility: interfaces.MasterEligibility{
				PublicIP:    "45.95.212.179",
				HasPublicIP: true,
				NATType:     "public",
				Reachable:   true,
				Eligible:    true,
			},
		},
	})
}

type clusterStubMasterEligibilityProber struct {
	eligibility interfaces.MasterEligibility
}

func (s clusterStubMasterEligibilityProber) ProbeMasterEligibility(ctx context.Context) interfaces.MasterEligibility {
	_ = ctx
	return s.eligibility
}
