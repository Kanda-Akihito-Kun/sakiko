package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/api"
	"sakiko.local/sakiko-core/auth"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/protocol"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Config struct {
	Addr        string
	Path        string
	Secret      string
	CheckOrigin func(r *http.Request) bool

	Mode                interfaces.Mode
	ConnConcurrency     uint
	SpeedConcurrency    uint
	SpeedInterval       time.Duration
	ProfilesPath        string
	ProfileFetchTimeout time.Duration
}

type Server struct {
	api      *api.Service
	verifier *auth.Verifier
	upgrader websocket.Upgrader
	server   *http.Server
	now      func() time.Time
}

type session struct {
	id            string
	conn          *websocket.Conn
	challenge     string
	authenticated bool
	clientID      string
	remoteAddr    string
	writeLock     sync.Mutex
}

func New(cfg Config) (*Server, error) {
	if cfg.Path == "" {
		cfg.Path = "/ws"
	}
	apiService, err := api.New(api.Config{
		Mode:                cfg.Mode,
		ConnConcurrency:     cfg.ConnConcurrency,
		SpeedConcurrency:    cfg.SpeedConcurrency,
		SpeedInterval:       cfg.SpeedInterval,
		ProfilesPath:        cfg.ProfilesPath,
		ProfileFetchTimeout: cfg.ProfileFetchTimeout,
	})
	if err != nil {
		return nil, err
	}
	verifier, err := auth.NewVerifier(auth.Config{Secret: cfg.Secret})
	if err != nil {
		return nil, err
	}

	s := &Server{
		api:      apiService,
		verifier: verifier,
		now:      time.Now,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     cfg.CheckOrigin,
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Path, s.handleWS)
	s.server = &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}
	wsLogger().Info("websocket server initialized",
		zap.String("addr", cfg.Addr),
		zap.String("path", cfg.Path),
	)
	return s, nil
}

func (s *Server) Start() error {
	wsLogger().Info("websocket server starting", zap.String("addr", s.server.Addr))
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		wsLogger().Error("websocket server stopped with error", zap.Error(err))
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	wsLogger().Info("websocket server shutting down")
	if s.api != nil {
		s.api.Stop()
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsLogger().Warn("websocket upgrade failed",
			zap.String("remote_addr", r.RemoteAddr),
			zap.Error(err),
		)
		return
	}
	defer conn.Close()

	challenge, err := s.verifier.NewChallenge()
	if err != nil {
		wsLogger().Error("create websocket challenge failed",
			zap.String("remote_addr", r.RemoteAddr),
			zap.Error(err),
		)
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "challenge error"))
		return
	}

	sess := &session{
		id:         challenge,
		conn:       conn,
		challenge:  challenge,
		remoteAddr: r.RemoteAddr,
	}
	wsLogger().Info("websocket session opened",
		zap.String("session_id", sess.id),
		zap.String("remote_addr", sess.remoteAddr),
	)
	if err := s.writeEnvelope(sess, protocol.EventAuthChallenge, protocol.ChallengePayload{
		Challenge: challenge,
		ServerTS:  s.now().UnixMilli(),
	}, ""); err != nil {
		wsLogger().Warn("write auth challenge failed",
			zap.String("session_id", sess.id),
			zap.Error(err),
		)
		return
	}

	for {
		var env protocol.Envelope
		if err := conn.ReadJSON(&env); err != nil {
			wsLogger().Debug("websocket session closed",
				zap.String("session_id", sess.id),
				zap.String("client_id", sess.clientID),
				zap.String("remote_addr", sess.remoteAddr),
				zap.Error(err),
			)
			return
		}
		if env.Version == "" {
			env.Version = protocol.Version
		}
		if !sess.authenticated {
			if err := s.handleVerify(sess, env); err != nil {
				wsLogger().Warn("websocket authentication failed",
					zap.String("session_id", sess.id),
					zap.String("remote_addr", sess.remoteAddr),
					zap.Error(err),
				)
				_ = s.writeError(sess, env.RequestID, "auth_failed", err.Error())
				return
			}
			continue
		}

		if err := s.verifier.Verify(env, s.now()); err != nil {
			wsLogger().Warn("websocket signature verification failed",
				zap.String("session_id", sess.id),
				zap.String("request_id", env.RequestID),
				zap.String("event", env.Event),
				zap.Error(err),
			)
			_ = s.writeError(sess, env.RequestID, "invalid_signature", err.Error())
			return
		}

		wsLogger().Debug("dispatch websocket event",
			zap.String("session_id", sess.id),
			zap.String("client_id", sess.clientID),
			zap.String("request_id", env.RequestID),
			zap.String("event", env.Event),
		)
		if err := s.dispatch(sess, env); err != nil {
			wsLogger().Warn("dispatch websocket event failed",
				zap.String("session_id", sess.id),
				zap.String("request_id", env.RequestID),
				zap.String("event", env.Event),
				zap.Error(err),
			)
			_ = s.writeError(sess, env.RequestID, "dispatch_failed", err.Error())
		}
	}
}

func (s *Server) handleVerify(sess *session, env protocol.Envelope) error {
	if env.Event != protocol.EventAuthVerify {
		return fmt.Errorf("expected %s", protocol.EventAuthVerify)
	}
	if err := s.verifier.Verify(env, s.now()); err != nil {
		return err
	}
	var payload protocol.VerifyPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return err
	}
	if strings.TrimSpace(payload.Challenge) == "" || payload.Challenge != sess.challenge {
		return fmt.Errorf("challenge mismatch")
	}

	sess.authenticated = true
	sess.clientID = payload.ClientID
	wsLogger().Info("websocket client authenticated",
		zap.String("session_id", sess.id),
		zap.String("client_id", sess.clientID),
		zap.String("remote_addr", sess.remoteAddr),
	)
	return s.writeEnvelope(sess, protocol.EventAuthOK, map[string]any{
		"clientId": payload.ClientID,
		"session":  sess.id,
	}, env.RequestID)
}

func (s *Server) dispatch(sess *session, env protocol.Envelope) error {
	switch env.Event {
	case protocol.EventTaskSubmit:
		var req interfaces.TaskSubmitRequest
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return err
		}
		resp, err := s.api.SubmitTask(req, func(event interfaces.Event) {
			switch event.Type {
			case interfaces.EventProcess:
				_ = s.writeEnvelope(sess, protocol.EventTaskProgress, event, env.RequestID)
			case interfaces.EventExit:
				_ = s.writeEnvelope(sess, protocol.EventTaskExit, event, env.RequestID)
			}
		})
		if err != nil {
			return err
		}
		return s.writeEnvelope(sess, protocol.EventTaskAccepted, resp, env.RequestID)
	case protocol.EventTaskList:
		return s.writeEnvelope(sess, protocol.EventTaskList, s.api.ListTasks(), env.RequestID)
	case protocol.EventTaskStatus:
		var payload struct {
			TaskID string `json:"taskId"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		resp, err := s.api.GetTask(payload.TaskID)
		if err != nil {
			return err
		}
		return s.writeEnvelope(sess, protocol.EventTaskStatus, resp, env.RequestID)
	case protocol.EventRuntimeStats:
		return s.writeEnvelope(sess, protocol.EventRuntimeStats, s.api.RuntimeStatus(), env.RequestID)
	case protocol.EventProfileImport:
		var req interfaces.ProfileImportRequest
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return err
		}
		resp, err := s.api.ImportProfile(req)
		if err != nil {
			return err
		}
		return s.writeEnvelope(sess, protocol.EventProfileImported, resp, env.RequestID)
	case protocol.EventProfileList:
		return s.writeEnvelope(sess, protocol.EventProfileList, s.api.ListProfiles(), env.RequestID)
	case protocol.EventProfileGet:
		var payload struct {
			ProfileID string `json:"profileId"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		resp, err := s.api.GetProfile(payload.ProfileID)
		if err != nil {
			return err
		}
		return s.writeEnvelope(sess, protocol.EventProfileGet, resp, env.RequestID)
	case protocol.EventProfileRefresh:
		var req interfaces.ProfileRefreshRequest
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return err
		}
		resp, err := s.api.RefreshProfile(req)
		if err != nil {
			return err
		}
		return s.writeEnvelope(sess, protocol.EventProfileUpdated, resp, env.RequestID)
	default:
		return fmt.Errorf("unsupported event: %s", env.Event)
	}
}

func wsLogger() *zap.Logger {
	return logx.Named("core.ws")
}

func (s *Server) writeError(sess *session, requestID string, code string, message string) error {
	return s.writeEnvelope(sess, protocol.EventError, protocol.ErrorPayload{
		Code:    code,
		Message: message,
	}, requestID)
}

func (s *Server) writeEnvelope(sess *session, event string, payload any, requestID string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := protocol.Envelope{
		Version:   protocol.Version,
		RequestID: requestID,
		Event:     event,
		Timestamp: s.now().UnixMilli(),
		Nonce:     fmt.Sprintf("%d", s.now().UnixNano()),
		Payload:   body,
	}
	signature, err := s.verifier.Sign(env)
	if err != nil {
		return err
	}
	env.Signature = signature

	sess.writeLock.Lock()
	defer sess.writeLock.Unlock()
	return sess.conn.WriteJSON(env)
}
