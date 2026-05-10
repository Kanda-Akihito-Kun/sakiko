package cluster

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

const (
	clusterBootstrapPath  = "/cluster/pairing/bootstrap"
	clusterHeartbeatPath  = "/cluster/knights/self/heartbeat"
	clusterPollPath       = "/cluster/knights/self/poll"
	clusterReportPath     = "/cluster/knights/self/report"
	clusterDisconnectPath = "/cluster/knights/self/disconnect"
)

type knightBootstrapBundle struct {
	KnightID         string
	KnightName       string
	MasterServerName string
	CACertificatePEM string
	ClientCertPEM    string
	ClientKeyPEM     string
}

type knightRuntime struct {
	cancel              context.CancelFunc
	done                chan struct{}
	masterHost          string
	masterPort          int
	client              *http.Client
	knightID            string
	knightName          string
	lock                sync.Mutex
	busy                bool
	currentRemoteTaskID string
	pendingReport       *interfaces.ClusterKnightReportRequest
}

func (s *Service) startMasterServer(listenHost string, listenPort int) error {
	materials, ok := s.pki.CurrentMasterMaterials()
	if !ok {
		return fmt.Errorf("master PKI materials are not ready")
	}
	if listenPort <= 0 {
		return fmt.Errorf("master listen port is required")
	}
	if err := s.stopMasterServer(); err != nil {
		return err
	}

	serverCert, err := tls.X509KeyPair([]byte(materials.ServerCertPEM), []byte(materials.ServerKeyPEM))
	if err != nil {
		return fmt.Errorf("load master TLS keypair: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM([]byte(materials.CACertPEM)) {
		return fmt.Errorf("load master CA certificate")
	}

	listenAddr := net.JoinHostPort(normalizeListenHost(listenHost), fmt.Sprintf("%d", listenPort))
	listener, err := tls.Listen("tcp", listenAddr, &tls.Config{
		MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{
			serverCert,
		},
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caPool,
	})
	if err != nil {
		return fmt.Errorf("start master listener: %w", err)
	}

	server := &http.Server{
		Handler:      s.masterMux(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	s.lock.Lock()
	s.masterServer = server
	s.lock.Unlock()

	go func() {
		err := server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			clusterLogger().Warn("master server stopped unexpectedly", zap.Error(err))
		}
	}()

	clusterLogger().Info("master server started",
		zap.String("listen_addr", listenAddr),
	)
	return nil
}

func (s *Service) stopMasterServer() error {
	if s == nil {
		return nil
	}

	s.lock.Lock()
	server := s.masterServer
	s.masterServer = nil
	s.lock.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown master server: %w", err)
	}
	return nil
}

func (s *Service) bootstrapKnight(ctx context.Context, masterHost string, masterPort int, oneTimeCode string) (knightBootstrapBundle, error) {
	keyPEM, csrPEM, err := generateKnightCSRPEM()
	if err != nil {
		return knightBootstrapBundle{}, err
	}

	reqBody := interfaces.ClusterPairingBootstrapRequest{
		OneTimeCode: oneTimeCode,
		CSRPEM:      csrPEM,
	}
	var resp interfaces.ClusterPairingBootstrapResponse
	if err := doJSONRequest(ctx, insecureBootstrapClient(), http.MethodPost, buildClusterURL(masterHost, masterPort, clusterBootstrapPath), reqBody, &resp); err != nil {
		return knightBootstrapBundle{}, err
	}

	return knightBootstrapBundle{
		KnightID:         resp.KnightID,
		KnightName:       resp.KnightName,
		MasterServerName: resp.MasterServerName,
		CACertificatePEM: resp.CACertificatePEM,
		ClientCertPEM:    resp.ClientCertificatePEM,
		ClientKeyPEM:     keyPEM,
	}, nil
}

func (s *Service) startKnightRuntime(masterHost string, masterPort int, bootstrap knightBootstrapBundle) (*knightRuntime, error) {
	client, err := newMutualTLSClient(bootstrap)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	runtime := &knightRuntime{
		cancel:     cancel,
		done:       make(chan struct{}),
		masterHost: masterHost,
		masterPort: masterPort,
		client:     client,
		knightID:   bootstrap.KnightID,
		knightName: bootstrap.KnightName,
	}

	go s.runKnightLoop(ctx, runtime)
	return runtime, nil
}

func (s *Service) stopKnightRuntime(reason string) error {
	if s == nil {
		return nil
	}

	s.lock.Lock()
	runtime := s.knightRuntime
	s.knightRuntime = nil
	s.lock.Unlock()

	if runtime == nil {
		return nil
	}

	bestEffortDisconnect(runtime, reason)
	runtime.cancel()
	<-runtime.done
	return nil
}

func (s *Service) runKnightLoop(ctx context.Context, runtime *knightRuntime) {
	defer close(runtime.done)

	beatOnce := func() {
		if s.flushPendingKnightReport(ctx, runtime) {
			return
		}

		if runtime.isBusy() {
			s.sendKnightHeartbeat(ctx, runtime, interfaces.ClusterKnightStateBusy)
			return
		}

		snapshot := s.snapshotKnightTaskState(runtime.currentRemoteTaskIDValue())
		pollReq := interfaces.ClusterKnightPollRequest{
			State: interfaces.ClusterKnightStateOnline,
			Task:  snapshot,
		}
		var pollResp interfaces.ClusterKnightPollResponse
		err := doJSONRequest(ctx, runtime.client, http.MethodPost, buildClusterURL(runtime.masterHost, runtime.masterPort, clusterPollPath), pollReq, &pollResp)
		if err != nil {
			s.updateLocalKnightConnection(false, "", err)
			if isKnightRevokedRequestError(err) {
				runtime.cancel()
			}
			return
		}
		s.updateLocalKnightConnection(true, pollResp.ServerTime, nil)
		if pollResp.Assignment != nil {
			if !runtime.trySetBusy() {
				clusterLogger().Warn("received assignment while knight runtime is already busy",
					zap.String("assignment_id", pollResp.Assignment.AssignmentID),
					zap.String("remote_task_id", pollResp.Assignment.RemoteTaskID),
					zap.String("knight_id", runtime.knightID),
				)
				return
			}
			go s.executeKnightAssignment(ctx, runtime, *pollResp.Assignment)
		}
	}

	beatOnce()

	ticker := time.NewTicker(knightHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beatOnce()
		}
	}
}

func (s *Service) executeKnightAssignment(ctx context.Context, runtime *knightRuntime, assignment interfaces.ClusterAssignment) {
	if s == nil || runtime == nil {
		return
	}
	runtime.setCurrentRemoteTaskID(assignment.RemoteTaskID)

	s.updateLocalKnightConnection(true, s.now().UTC().Format(time.RFC3339), nil)

	reportReq := interfaces.ClusterKnightReportRequest{
		AssignmentID: assignment.AssignmentID,
		RemoteTaskID: assignment.RemoteTaskID,
		FinishedAt:   s.now().UTC().Format(time.RFC3339),
	}

	if s.runKnightAssignment == nil {
		reportReq.Status = "failed"
		reportReq.Error = "knight assignment runner is not configured"
		if s.reportKnightAssignment(ctx, runtime, reportReq) {
			runtime.markIdle()
			return
		}
		runtime.setPendingReport(&reportReq)
		return
	}

	archive, err := s.runKnightAssignment(ctx, assignment)
	reportReq.FinishedAt = s.now().UTC().Format(time.RFC3339)
	if err != nil {
		reportReq.Status = "failed"
		reportReq.Error = err.Error()
	} else {
		reportReq.Status = "ok"
		reportReq.ExitCode = strings.TrimSpace(archive.ExitCode)
		reportReq.Archive = &archive
	}

	if s.reportKnightAssignment(ctx, runtime, reportReq) {
		runtime.markIdle()
		return
	}
	runtime.setPendingReport(&reportReq)
}

func (s *Service) sendKnightHeartbeat(ctx context.Context, runtime *knightRuntime, state interfaces.ClusterKnightState) {
	if runtime == nil {
		return
	}

	var resp interfaces.ClusterKnightHeartbeatResponse
	if err := doJSONRequest(ctx, runtime.client, http.MethodPost, buildClusterURL(runtime.masterHost, runtime.masterPort, clusterHeartbeatPath), interfaces.ClusterKnightHeartbeatRequest{
		State: state,
		Task:  s.snapshotKnightTaskState(runtime.currentRemoteTaskIDValue()),
	}, &resp); err != nil {
		s.updateLocalKnightConnection(false, "", err)
		if isKnightRevokedRequestError(err) {
			runtime.cancel()
		}
		return
	}
	s.updateLocalKnightConnection(true, resp.ServerTime, nil)
}

func (s *Service) snapshotKnightTaskState(remoteTaskID string) *interfaces.TaskState {
	if s == nil || s.snapshotKnightTask == nil || strings.TrimSpace(remoteTaskID) == "" {
		return nil
	}

	state, _, err := s.snapshotKnightTask(strings.TrimSpace(remoteTaskID))
	if err != nil || state == nil {
		return nil
	}
	cloned := *state
	cloned.ActiveNodes = append([]interfaces.TaskActiveNode{}, state.ActiveNodes...)
	return &cloned
}

func (s *Service) reportKnightAssignment(ctx context.Context, runtime *knightRuntime, req interfaces.ClusterKnightReportRequest) bool {
	if runtime == nil {
		return false
	}

	var resp interfaces.ClusterKnightReportResponse
	if err := doJSONRequest(ctx, runtime.client, http.MethodPost, buildClusterURL(runtime.masterHost, runtime.masterPort, clusterReportPath), req, &resp); err != nil {
		s.updateLocalKnightConnection(false, "", err)
		if isKnightRevokedRequestError(err) {
			runtime.cancel()
		}
		return false
	}
	s.updateLocalKnightConnection(true, resp.ServerTime, nil)
	return true
}

func (s *Service) flushPendingKnightReport(ctx context.Context, runtime *knightRuntime) bool {
	if runtime == nil {
		return false
	}

	req := runtime.pendingReportValue()
	if req == nil {
		return false
	}
	if !s.reportKnightAssignment(ctx, runtime, *req) {
		return false
	}

	runtime.markIdle()
	return true
}

func bestEffortDisconnect(runtime *knightRuntime, reason string) {
	if runtime == nil || runtime.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = doJSONRequest(ctx, runtime.client, http.MethodPost, buildClusterURL(runtime.masterHost, runtime.masterPort, clusterDisconnectPath), interfaces.ClusterKnightDisconnectRequest{
		Reason: strings.TrimSpace(reason),
	}, nil)
}

func (s *Service) updateLocalKnightConnection(connected bool, serverTime string, err error) {
	if s == nil {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status.Knight == nil {
		return
	}

	s.status.Knight.Connected = connected
	if strings.TrimSpace(serverTime) != "" {
		s.status.Knight.LastSeenAt = strings.TrimSpace(serverTime)
	} else if connected {
		s.status.Knight.LastSeenAt = s.now().UTC().Format(time.RFC3339)
	}
	if err != nil {
		s.status.Knight.LastError = err.Error()
	} else {
		s.status.Knight.LastError = ""
	}
}

func (r *knightRuntime) isBusy() bool {
	if r == nil {
		return false
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	return r.busy
}

func (r *knightRuntime) trySetBusy() bool {
	if r == nil {
		return false
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	if r.busy {
		return false
	}
	r.busy = true
	return true
}

func (r *knightRuntime) setBusy(busy bool) {
	if r == nil {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.busy = busy
}

func (r *knightRuntime) setCurrentRemoteTaskID(remoteTaskID string) {
	if r == nil {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.currentRemoteTaskID = strings.TrimSpace(remoteTaskID)
}

func (r *knightRuntime) currentRemoteTaskIDValue() string {
	if r == nil {
		return ""
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	return r.currentRemoteTaskID
}

func (r *knightRuntime) setPendingReport(req *interfaces.ClusterKnightReportRequest) {
	if r == nil || req == nil {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	reportCopy := *req
	r.pendingReport = &reportCopy
}

func (r *knightRuntime) pendingReportValue() *interfaces.ClusterKnightReportRequest {
	if r == nil {
		return nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	if r.pendingReport == nil {
		return nil
	}
	reportCopy := *r.pendingReport
	return &reportCopy
}

func (r *knightRuntime) markIdle() {
	if r == nil {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.busy = false
	r.currentRemoteTaskID = ""
	r.pendingReport = nil
}

func newMutualTLSClient(bundle knightBootstrapBundle) (*http.Client, error) {
	certificate, err := tls.X509KeyPair([]byte(bundle.ClientCertPEM), []byte(bundle.ClientKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("load knight client certificate: %w", err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(bundle.CACertificatePEM)) {
		return nil, fmt.Errorf("load master CA certificate")
	}

	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				RootCAs:      rootCAs,
				Certificates: []tls.Certificate{certificate},
				ServerName:   strings.TrimSpace(bundle.MasterServerName),
			},
		},
	}, nil
}

func insecureBootstrapClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			},
		},
	}
}

func generateKnightCSRPEM() (keyPEM string, csrPEM string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate knight key: %w", err)
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "sakiko-knight"},
	}, key)
	if err != nil {
		return "", "", fmt.Errorf("create knight CSR: %w", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})),
		string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: csrDER,
		})),
		nil
}

func buildClusterURL(host string, port int, path string) string {
	return fmt.Sprintf("https://%s:%d%s", strings.TrimSpace(host), port, path)
}

func normalizeListenHost(host string) string {
	value := strings.TrimSpace(host)
	if value == "" || value == "0.0.0.0" {
		return ""
	}
	return value
}

func doJSONRequest(ctx context.Context, client *http.Client, method string, url string, body any, target any) error {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		message := strings.TrimSpace(string(data))
		if message == "" {
			message = resp.Status
		}
		return fmt.Errorf("cluster request failed: %s", message)
	}

	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func isKnightRevokedRequestError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "knight revoked by master")
}
