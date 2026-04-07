package protocol

import "encoding/json"

const Version = "v1"

const (
	EventAuthChallenge = "auth.challenge"
	EventAuthVerify    = "auth.verify"
	EventAuthOK        = "auth.ok"
	EventError         = "system.error"

	EventTaskSubmit   = "task.submit"
	EventTaskAccepted = "task.accepted"
	EventTaskProgress = "task.progress"
	EventTaskExit     = "task.exit"
	EventTaskList     = "task.list"
	EventTaskStatus   = "task.status"
	EventRuntimeStats = "runtime.status"

	EventProfileImport   = "profile.import"
	EventProfileImported = "profile.imported"
	EventProfileList     = "profile.list"
	EventProfileGet      = "profile.get"
	EventProfileRefresh  = "profile.refresh"
	EventProfileUpdated  = "profile.updated"
)

type Envelope struct {
	Version   string          `json:"version,omitempty"`
	RequestID string          `json:"requestId,omitempty"`
	Event     string          `json:"event"`
	Timestamp int64           `json:"ts"`
	Nonce     string          `json:"nonce"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Signature string          `json:"signature,omitempty"`
}

type ChallengePayload struct {
	Challenge string `json:"challenge"`
	ServerTS  int64  `json:"serverTs"`
}

type VerifyPayload struct {
	ClientID  string `json:"clientId"`
	Challenge string `json:"challenge"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
