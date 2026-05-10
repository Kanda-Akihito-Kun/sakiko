package interfaces

type ClusterRole string

const (
	ClusterRoleStandalone ClusterRole = "standalone"
	ClusterRoleMaster     ClusterRole = "master"
	ClusterRoleKnight     ClusterRole = "knight"
)

type MasterEligibility struct {
	PublicIP    string `json:"publicIP,omitempty"`
	HasPublicIP bool   `json:"hasPublicIP"`
	NATType     string `json:"natType,omitempty"`
	IsNAT1      bool   `json:"isNat1"`
	Reachable   bool   `json:"reachable"`
	Eligible    bool   `json:"eligible"`
	CheckedAt   string `json:"checkedAt,omitempty"`
	Error       string `json:"error,omitempty"`
}

type ClusterMasterStatus struct {
	Enabled     bool              `json:"enabled"`
	ListenHost  string            `json:"listenHost,omitempty"`
	ListenPort  int               `json:"listenPort,omitempty"`
	Eligibility MasterEligibility `json:"eligibility"`
}

type ClusterKnightStatus struct {
	Bound      bool   `json:"bound"`
	KnightID   string `json:"knightId,omitempty"`
	KnightName string `json:"knightName,omitempty"`
	MasterHost string `json:"masterHost,omitempty"`
	MasterPort int    `json:"masterPort,omitempty"`
	Connected  bool   `json:"connected"`
	LastSeenAt string `json:"lastSeenAt,omitempty"`
	LastError  string `json:"lastError,omitempty"`
}

type ClusterStatus struct {
	Role   ClusterRole          `json:"role"`
	Master *ClusterMasterStatus `json:"master,omitempty"`
	Knight *ClusterKnightStatus `json:"knight,omitempty"`
}

type ClusterStatusResponse struct {
	Status ClusterStatus `json:"status"`
}

type ClusterMasterEligibilityResponse struct {
	Eligibility MasterEligibility `json:"eligibility"`
	Status      ClusterStatus     `json:"status"`
}

type ClusterEnableMasterRequest struct {
	ListenHost string `json:"listenHost,omitempty"`
	ListenPort int    `json:"listenPort,omitempty"`
}

type ClusterEnableMasterResponse struct {
	Status ClusterStatus `json:"status"`
}

type ClusterEnableKnightRequest struct {
	MasterHost  string `json:"masterHost"`
	MasterPort  int    `json:"masterPort"`
	OneTimeCode string `json:"oneTimeCode"`
}

type ClusterEnableKnightResponse struct {
	Status ClusterStatus `json:"status"`
}

type ClusterDisableRemoteResponse struct {
	Status ClusterStatus `json:"status"`
}

type ClusterCreatePairingCodeRequest struct {
	KnightName string `json:"knightName,omitempty"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"`
}

type ClusterPairingCode struct {
	Code       string `json:"code"`
	KnightName string `json:"knightName,omitempty"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
}

type ClusterCreatePairingCodeResponse struct {
	PairingCode ClusterPairingCode `json:"pairingCode"`
	Status      ClusterStatus      `json:"status"`
}

type ClusterPairingBootstrapRequest struct {
	OneTimeCode string `json:"oneTimeCode"`
	KnightID    string `json:"knightId,omitempty"`
	KnightName  string `json:"knightName,omitempty"`
	CSRPEM      string `json:"csrPem"`
}

type ClusterPairingBootstrapResponse struct {
	KnightID             string        `json:"knightId"`
	KnightName           string        `json:"knightName,omitempty"`
	ClientCertificatePEM string        `json:"clientCertificatePem"`
	CACertificatePEM     string        `json:"caCertificatePem"`
	MasterServerName     string        `json:"masterServerName,omitempty"`
	Status               ClusterStatus `json:"status"`
}

type ClusterKnightState string

const (
	ClusterKnightStatePaired  ClusterKnightState = "paired"
	ClusterKnightStateOnline  ClusterKnightState = "online"
	ClusterKnightStateBusy    ClusterKnightState = "busy"
	ClusterKnightStateOffline ClusterKnightState = "offline"
	ClusterKnightStateRevoked ClusterKnightState = "revoked"
)

type ClusterConnectedKnight struct {
	KnightID   string             `json:"knightId"`
	KnightName string             `json:"knightName,omitempty"`
	State      ClusterKnightState `json:"state"`
	RemoteAddr string             `json:"remoteAddr,omitempty"`
	LastSeenAt string             `json:"lastSeenAt,omitempty"`
	LastError  string             `json:"lastError,omitempty"`
}

type ClusterListKnightsResponse struct {
	Knights []ClusterConnectedKnight `json:"knights"`
}

type ClusterKickKnightRequest struct {
	KnightID string `json:"knightId"`
}

type ClusterKickKnightResponse struct {
	Status ClusterStatus `json:"status"`
}

type ClusterKnightHeartbeatRequest struct {
	State ClusterKnightState `json:"state,omitempty"`
	Task  *TaskState         `json:"task,omitempty"`
}

type ClusterKnightHeartbeatResponse struct {
	Ack        bool   `json:"ack"`
	ServerTime string `json:"serverTime,omitempty"`
}

type ClusterAssignment struct {
	AssignmentID string `json:"assignmentId,omitempty"`
	RemoteTaskID string `json:"remoteTaskId,omitempty"`
	KnightID     string `json:"knightId,omitempty"`
	KnightName   string `json:"knightName,omitempty"`
	TaskName     string `json:"taskName,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	Task         *Task  `json:"task,omitempty"`
}

type ClusterRemoteTaskState string

const (
	ClusterRemoteTaskQueued   ClusterRemoteTaskState = "queued"
	ClusterRemoteTaskRunning  ClusterRemoteTaskState = "running"
	ClusterRemoteTaskFinished ClusterRemoteTaskState = "finished"
	ClusterRemoteTaskFailed   ClusterRemoteTaskState = "failed"
)

type ClusterRemoteTask struct {
	AssignmentID  string                 `json:"assignmentId"`
	RemoteTaskID  string                 `json:"remoteTaskId"`
	KnightID      string                 `json:"knightId"`
	KnightName    string                 `json:"knightName,omitempty"`
	TaskName      string                 `json:"taskName,omitempty"`
	State         ClusterRemoteTaskState `json:"state"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	StartedAt     string                 `json:"startedAt,omitempty"`
	FinishedAt    string                 `json:"finishedAt,omitempty"`
	ExitCode      string                 `json:"exitCode,omitempty"`
	Error         string                 `json:"error,omitempty"`
	ArchiveTaskID string                 `json:"archiveTaskId,omitempty"`
	LocalTaskID   string                 `json:"localTaskId,omitempty"`
	Runtime       *TaskState             `json:"runtime,omitempty"`
}

type ClusterDispatchTaskRequest struct {
	KnightIDs []string `json:"knightIds"`
	Task      Task     `json:"task"`
}

type ClusterDispatchTaskResponse struct {
	Tasks []ClusterRemoteTask `json:"tasks"`
}

type ClusterListRemoteTasksResponse struct {
	Tasks []ClusterRemoteTask `json:"tasks"`
}

type ClusterKnightPollRequest struct {
	State ClusterKnightState `json:"state,omitempty"`
	Task  *TaskState         `json:"task,omitempty"`
}

type ClusterKnightPollResponse struct {
	Ack        bool               `json:"ack"`
	ServerTime string             `json:"serverTime,omitempty"`
	Assignment *ClusterAssignment `json:"assignment,omitempty"`
}

type ClusterKnightReportRequest struct {
	AssignmentID string         `json:"assignmentId,omitempty"`
	RemoteTaskID string         `json:"remoteTaskId,omitempty"`
	Status       string         `json:"status,omitempty"`
	ExitCode     string         `json:"exitCode,omitempty"`
	Error        string         `json:"error,omitempty"`
	FinishedAt   string         `json:"finishedAt,omitempty"`
	Archive      *ResultArchive `json:"archive,omitempty"`
}

type ClusterKnightReportResponse struct {
	Ack        bool   `json:"ack"`
	ServerTime string `json:"serverTime,omitempty"`
}

type ClusterKnightDisconnectRequest struct {
	Reason string `json:"reason,omitempty"`
}

type ClusterKnightDisconnectResponse struct {
	Ack        bool   `json:"ack"`
	ServerTime string `json:"serverTime,omitempty"`
}
