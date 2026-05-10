package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const knightCertificatePrefix = "sakiko-knight-"

func (s *Service) masterMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(clusterBootstrapPath, s.handleBootstrapKnight)
	mux.HandleFunc(clusterHeartbeatPath, s.handleKnightHeartbeat)
	mux.HandleFunc(clusterPollPath, s.handleKnightPoll)
	mux.HandleFunc(clusterReportPath, s.handleKnightReport)
	mux.HandleFunc(clusterDisconnectPath, s.handleKnightDisconnect)
	return mux
}

func (s *Service) handleBootstrapKnight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeClusterError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req interfaces.ClusterPairingBootstrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeClusterError(w, http.StatusBadRequest, fmt.Sprintf("invalid bootstrap request: %v", err))
		return
	}

	resp, err := s.BootstrapKnight(req)
	if err != nil {
		writeClusterError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeClusterJSON(w, http.StatusOK, resp)
}

func (s *Service) handleKnightHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeClusterError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	knightID, knightName, remoteAddr, err := requireKnightIdentity(r)
	if err != nil {
		writeClusterError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if s.isKnightRevoked(knightID) {
		writeClusterError(w, http.StatusForbidden, "knight revoked by master")
		return
	}

	var req interfaces.ClusterKnightHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeClusterError(w, http.StatusBadRequest, fmt.Sprintf("invalid heartbeat request: %v", err))
		return
	}

	state := req.State
	if state == "" {
		state = interfaces.ClusterKnightStateOnline
	}
	s.upsertKnight(interfaces.ClusterConnectedKnight{
		KnightID:   knightID,
		KnightName: knightName,
		State:      state,
		RemoteAddr: remoteAddr,
		LastSeenAt: s.now().UTC().Format(time.RFC3339),
	})
	s.updateKnightRuntimeSnapshot(knightID, req.Task)

	writeClusterJSON(w, http.StatusOK, interfaces.ClusterKnightHeartbeatResponse{
		Ack:        true,
		ServerTime: s.now().UTC().Format(time.RFC3339),
	})
}

func (s *Service) handleKnightPoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeClusterError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	knightID, knightName, remoteAddr, err := requireKnightIdentity(r)
	if err != nil {
		writeClusterError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if s.isKnightRevoked(knightID) {
		writeClusterError(w, http.StatusForbidden, "knight revoked by master")
		return
	}

	var req interfaces.ClusterKnightPollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeClusterError(w, http.StatusBadRequest, fmt.Sprintf("invalid poll request: %v", err))
		return
	}

	state := req.State
	if state == "" {
		state = interfaces.ClusterKnightStateOnline
	}
	var assignment *interfaces.ClusterAssignment
	if state != interfaces.ClusterKnightStateBusy {
		assignment = s.leaseAssignmentForKnight(knightID)
	}
	if assignment != nil {
		state = interfaces.ClusterKnightStateBusy
	}
	s.upsertKnight(interfaces.ClusterConnectedKnight{
		KnightID:   knightID,
		KnightName: knightName,
		State:      state,
		RemoteAddr: remoteAddr,
		LastSeenAt: s.now().UTC().Format(time.RFC3339),
	})
	s.updateKnightRuntimeSnapshot(knightID, req.Task)

	writeClusterJSON(w, http.StatusOK, interfaces.ClusterKnightPollResponse{
		Ack:        true,
		ServerTime: s.now().UTC().Format(time.RFC3339),
		Assignment: assignment,
	})
}

func (s *Service) handleKnightReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeClusterError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	knightID, knightName, remoteAddr, err := requireKnightIdentity(r)
	if err != nil {
		writeClusterError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if s.isKnightRevoked(knightID) {
		writeClusterError(w, http.StatusForbidden, "knight revoked by master")
		return
	}

	var req interfaces.ClusterKnightReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeClusterError(w, http.StatusBadRequest, fmt.Sprintf("invalid report request: %v", err))
		return
	}

	state := interfaces.ClusterKnightStateOnline
	if strings.EqualFold(strings.TrimSpace(req.Status), "running") {
		state = interfaces.ClusterKnightStateBusy
	}
	s.upsertKnight(interfaces.ClusterConnectedKnight{
		KnightID:   knightID,
		KnightName: knightName,
		State:      state,
		RemoteAddr: remoteAddr,
		LastSeenAt: s.now().UTC().Format(time.RFC3339),
		LastError:  strings.TrimSpace(req.Error),
	})
	s.completeAssignmentReport(r.Context(), req)

	writeClusterJSON(w, http.StatusOK, interfaces.ClusterKnightReportResponse{
		Ack:        true,
		ServerTime: s.now().UTC().Format(time.RFC3339),
	})
}

func (s *Service) handleKnightDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeClusterError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	knightID, knightName, remoteAddr, err := requireKnightIdentity(r)
	if err != nil {
		writeClusterError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if s.isKnightRevoked(knightID) {
		writeClusterError(w, http.StatusForbidden, "knight revoked by master")
		return
	}

	var req interfaces.ClusterKnightDisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeClusterError(w, http.StatusBadRequest, fmt.Sprintf("invalid disconnect request: %v", err))
		return
	}

	s.upsertKnight(interfaces.ClusterConnectedKnight{
		KnightID:   knightID,
		KnightName: knightName,
		State:      interfaces.ClusterKnightStateOffline,
		RemoteAddr: remoteAddr,
		LastSeenAt: s.now().UTC().Format(time.RFC3339),
		LastError:  strings.TrimSpace(req.Reason),
	})

	writeClusterJSON(w, http.StatusOK, interfaces.ClusterKnightDisconnectResponse{
		Ack:        true,
		ServerTime: s.now().UTC().Format(time.RFC3339),
	})
}

func requireKnightIdentity(r *http.Request) (knightID string, knightName string, remoteAddr string, err error) {
	if r == nil || r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return "", "", "", fmt.Errorf("knight client certificate is required")
	}

	cert := r.TLS.PeerCertificates[0]
	commonName := strings.TrimSpace(cert.Subject.CommonName)
	if !strings.HasPrefix(commonName, knightCertificatePrefix) {
		return "", "", "", fmt.Errorf("invalid knight certificate subject")
	}
	knightID = strings.TrimPrefix(commonName, knightCertificatePrefix)
	if len(cert.Subject.Organization) > 0 {
		knightName = strings.TrimSpace(cert.Subject.Organization[0])
	}
	return knightID, knightName, strings.TrimSpace(r.RemoteAddr), nil
}

func (s *Service) upsertKnight(knight interfaces.ClusterConnectedKnight) {
	if s == nil || strings.TrimSpace(knight.KnightID) == "" {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	existing, ok := s.knights[knight.KnightID]
	if ok {
		if strings.TrimSpace(knight.KnightName) == "" {
			knight.KnightName = existing.KnightName
		}
		if strings.TrimSpace(knight.RemoteAddr) == "" {
			knight.RemoteAddr = existing.RemoteAddr
		}
		if strings.TrimSpace(knight.LastSeenAt) == "" {
			knight.LastSeenAt = existing.LastSeenAt
		}
		if strings.TrimSpace(knight.LastError) == "" {
			knight.LastError = existing.LastError
		}
		if knight.State == "" {
			knight.State = existing.State
		}
	}
	if knight.State == "" {
		knight.State = interfaces.ClusterKnightStatePaired
	}
	s.knights[knight.KnightID] = knight
}

func (s *Service) isKnightRevoked(knightID string) bool {
	if s == nil || strings.TrimSpace(knightID) == "" {
		return false
	}

	s.lock.RLock()
	defer s.lock.RUnlock()
	_, revoked := s.revokedKnights[strings.TrimSpace(knightID)]
	return revoked
}

func (s *Service) updateKnightRuntimeSnapshot(knightID string, snapshot *interfaces.TaskState) {
	if s == nil || snapshot == nil {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for remoteTaskID, remoteTask := range s.remoteTasks {
		if remoteTask.KnightID != knightID || remoteTask.State != interfaces.ClusterRemoteTaskRunning {
			continue
		}

		runtime := *snapshot
		remoteTask.LocalTaskID = strings.TrimSpace(snapshot.TaskID)
		remoteTask.Runtime = &runtime
		remoteTask.Error = ""
		s.remoteTasks[remoteTaskID] = remoteTask
	}
}

func writeClusterJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeClusterError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, strings.TrimSpace(message), statusCode)
}
