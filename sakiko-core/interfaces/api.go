package interfaces

type TaskSubmitRequest struct {
	Task         Task `json:"task"`
	RemoteIssued bool `json:"remoteIssued,omitempty"`
}

type TaskSubmitResponse struct {
	TaskID string `json:"taskId"`
}

type TaskListResponse struct {
	Tasks []TaskState `json:"tasks"`
}

type TaskStatusResponse struct {
	Task     TaskState     `json:"task"`
	Results  []EntryResult `json:"results,omitempty"`
	ExitCode string        `json:"exitCode,omitempty"`
}
