package interfaces

type BackendInfo struct {
	IP        string `json:"ip,omitempty"`
	Location  string `json:"location,omitempty"`
	Source    string `json:"source,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Error     string `json:"error,omitempty"`
}

type RemoteExecutionMode string

const (
	RemoteExecutionModeLocal        RemoteExecutionMode = "local"
	RemoteExecutionModeRemoteMaster RemoteExecutionMode = "remote-master"
	RemoteExecutionModeRemoteKnight RemoteExecutionMode = "remote-knight"
)

type RemoteExecutionContext struct {
	Mode         RemoteExecutionMode `json:"mode,omitempty"`
	RemoteTaskID string              `json:"remoteTaskId,omitempty"`
	MasterID     string              `json:"masterId,omitempty"`
	MasterName   string              `json:"masterName,omitempty"`
	KnightID     string              `json:"knightId,omitempty"`
	KnightName   string              `json:"knightName,omitempty"`
	AssignmentID string              `json:"assignmentId,omitempty"`
}

type TaskEnvironment struct {
	Identity string                  `json:"identity,omitempty"`
	Backend  *BackendInfo            `json:"backend,omitempty"`
	Remote   *RemoteExecutionContext `json:"remote,omitempty"`
}
