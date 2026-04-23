package interfaces

type TaskContext struct {
	Preset        string `json:"preset,omitempty"`
	ProfileID     string `json:"profileId,omitempty"`
	ProfileName   string `json:"profileName,omitempty"`
	ProfileSource string `json:"profileSource,omitempty"`
}

type TaskArchiveSnapshot struct {
	Task     Task          `json:"task"`
	State    TaskState     `json:"state"`
	Results  []EntryResult `json:"results,omitempty"`
	ExitCode string        `json:"exitCode,omitempty"`
}

type ResultArchiveWriter interface {
	SaveTaskArchive(snapshot TaskArchiveSnapshot) error
}

type ResultReportColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type ResultReportSection struct {
	Kind    string               `json:"kind"`
	Title   string               `json:"title"`
	Columns []ResultReportColumn `json:"columns,omitempty"`
	Rows    []map[string]any     `json:"rows,omitempty"`
	Summary map[string]any       `json:"summary,omitempty"`
}

type ResultReport struct {
	GeneratedAt string                `json:"generatedAt"`
	Sections    []ResultReportSection `json:"sections,omitempty"`
}

type ResultArchiveNode struct {
	Name  string `json:"name"`
	Order int    `json:"order,omitempty"`
}

type ResultArchiveTask struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Vendor      VendorType          `json:"vendor"`
	Context     TaskContext         `json:"context,omitempty"`
	Environment *TaskEnvironment    `json:"environment,omitempty"`
	Nodes       []ResultArchiveNode `json:"nodes"`
	Matrices    []MatrixEntry       `json:"matrices"`
	Config      TaskConfig          `json:"config"`
}

type ResultArchive struct {
	Version  int               `json:"version"`
	Task     ResultArchiveTask `json:"task"`
	State    TaskState         `json:"state"`
	Results  []EntryResult     `json:"results,omitempty"`
	ExitCode string            `json:"exitCode,omitempty"`
	Report   ResultReport      `json:"report"`
}

type ResultArchiveListItem struct {
	TaskID       string `json:"taskId"`
	TaskName     string `json:"taskName"`
	Preset       string `json:"preset,omitempty"`
	ProfileID    string `json:"profileId,omitempty"`
	ProfileName  string `json:"profileName,omitempty"`
	StartedAt    string `json:"startedAt,omitempty"`
	FinishedAt   string `json:"finishedAt,omitempty"`
	ExitCode     string `json:"exitCode,omitempty"`
	NodeCount    int    `json:"nodeCount"`
	SectionCount int    `json:"sectionCount"`
}

type ResultArchiveListResponse struct {
	Archives []ResultArchiveListItem `json:"archives"`
}

type ResultArchiveGetResponse struct {
	Archive ResultArchive `json:"archive"`
}

type ResultArchiveDeleteRequest struct {
	TaskID string `json:"taskId"`
}

type ResultArchiveDeleteResponse struct {
	TaskID string `json:"taskId"`
}
