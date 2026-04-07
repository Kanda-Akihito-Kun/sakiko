package interfaces

type BackendInfo struct {
	IP        string `json:"ip,omitempty"`
	Location  string `json:"location,omitempty"`
	Source    string `json:"source,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Error     string `json:"error,omitempty"`
}

type TaskEnvironment struct {
	Backend *BackendInfo `json:"backend,omitempty"`
}
