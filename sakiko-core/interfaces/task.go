package interfaces

const (
	defaultDownloadURL      = "https://speed.cloudflare.com/__down?bytes=10000000"
	defaultDownloadDuration = int64(10)
	minDownloadDuration     = int64(5)
	maxDownloadDuration     = int64(20)
)

type Node struct {
	Name     string `json:"name"`
	Order    int    `json:"order,omitempty" yaml:"-"`
	Protocol string `json:"protocol,omitempty" yaml:"-"`
	Server   string `json:"server,omitempty" yaml:"-"`
	Port     string `json:"port,omitempty" yaml:"-"`
	UDP      *bool  `json:"udp,omitempty" yaml:"-"`
	Payload  string `json:"payload,omitempty" yaml:"-"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

type TaskConfig struct {
	PingAddress       string `json:"pingAddress"`
	PingAverageOver   uint16 `json:"pingAverageOver"`
	TaskRetry         uint   `json:"taskRetry"`
	TaskTimeoutMillis uint   `json:"taskTimeoutMillis"`
	DownloadURL       string `json:"downloadURL"`
	DownloadDuration  int64  `json:"downloadDuration"`
	DownloadThreading uint   `json:"downloadThreading"`
}

func (c TaskConfig) Normalize() TaskConfig {
	if c.PingAddress == "" {
		c.PingAddress = "https://www.gstatic.com/generate_204"
	}
	if c.PingAverageOver == 0 {
		c.PingAverageOver = 1
	}
	if c.TaskRetry == 0 {
		c.TaskRetry = 1
	}
	if c.TaskTimeoutMillis == 0 {
		c.TaskTimeoutMillis = 6000
	}
	if c.DownloadURL == "" {
		c.DownloadURL = defaultDownloadURL
	}
	if c.DownloadDuration <= 0 {
		c.DownloadDuration = defaultDownloadDuration
	}
	if c.DownloadDuration < minDownloadDuration {
		c.DownloadDuration = minDownloadDuration
	}
	if c.DownloadDuration > maxDownloadDuration {
		c.DownloadDuration = maxDownloadDuration
	}
	if c.DownloadThreading == 0 {
		c.DownloadThreading = 1
	}
	return c
}

type Task struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Vendor      VendorType       `json:"vendor"`
	Context     TaskContext      `json:"context,omitempty"`
	Environment *TaskEnvironment `json:"environment,omitempty"`
	Nodes       []Node           `json:"nodes"`
	Matrices    []MatrixEntry    `json:"matrices"`
	Config      TaskConfig       `json:"config"`
}

type MacroType string

const (
	MacroPing    MacroType = "PING"
	MacroSpeed   MacroType = "SPEED"
	MacroGeo     MacroType = "GEO"
	MacroMedia   MacroType = "MEDIA"
	MacroInvalid MacroType = "INVALID"
)

type MatrixType string

const (
	MatrixHTTPPing      MatrixType = "TEST_PING_CONN"
	MatrixRTTPing       MatrixType = "TEST_PING_RTT"
	MatrixAverageSpeed  MatrixType = "SPEED_AVERAGE"
	MatrixMaxSpeed      MatrixType = "SPEED_MAX"
	MatrixPerSecSpeed   MatrixType = "SPEED_PER_SECOND"
	MatrixTrafficUsed   MatrixType = "SPEED_TRAFFIC_USED"
	MatrixInboundGeoIP  MatrixType = "GEOIP_INBOUND"
	MatrixOutboundGeoIP MatrixType = "GEOIP_OUTBOUND"
	MatrixMediaUnlock   MatrixType = "MEDIA_UNLOCK"
	MatrixInvalid       MatrixType = "INVALID"
)

type MatrixEntry struct {
	Type   MatrixType `json:"type"`
	Params string     `json:"params,omitempty"`
}

type MatrixResult struct {
	Type    MatrixType `json:"type"`
	Payload any        `json:"payload"`
}

type GeoIPInfo struct {
	Address        string `json:"address,omitempty"`
	IP             string `json:"ip,omitempty"`
	ASN            int    `json:"asn,omitempty"`
	ASOrganization string `json:"asOrganization,omitempty"`
	ISP            string `json:"isp,omitempty"`
	Country        string `json:"country,omitempty"`
	City           string `json:"city,omitempty"`
	CountryCode    string `json:"countryCode,omitempty"`
	Error          string `json:"error,omitempty"`
}

type MediaUnlockStatus string

const (
	MediaUnlockStatusYes           MediaUnlockStatus = "yes"
	MediaUnlockStatusNo            MediaUnlockStatus = "no"
	MediaUnlockStatusOriginalsOnly MediaUnlockStatus = "originals_only"
	MediaUnlockStatusWebOnly       MediaUnlockStatus = "web_only"
	MediaUnlockStatusOverseaOnly   MediaUnlockStatus = "oversea_only"
	MediaUnlockStatusUnsupported   MediaUnlockStatus = "unsupported"
	MediaUnlockStatusFailed        MediaUnlockStatus = "failed"
)

type MediaUnlockMode string

const (
	MediaUnlockModeNative  MediaUnlockMode = "native"
	MediaUnlockModeDNS     MediaUnlockMode = "dns"
	MediaUnlockModeUnknown MediaUnlockMode = "unknown"
)

type MediaUnlockPlatform string

const (
	MediaUnlockPlatformNetflix        MediaUnlockPlatform = "netflix"
	MediaUnlockPlatformHulu           MediaUnlockPlatform = "hulu"
	MediaUnlockPlatformHuluJP         MediaUnlockPlatform = "hulu_jp"
	MediaUnlockPlatformBilibiliHMT    MediaUnlockPlatform = "bilibili_hmt"
	MediaUnlockPlatformBilibiliTW     MediaUnlockPlatform = "bilibili_tw"
	MediaUnlockPlatformYouTubePremium MediaUnlockPlatform = "youtube_premium"
	MediaUnlockPlatformPrimeVideo     MediaUnlockPlatform = "prime_video"
	MediaUnlockPlatformHBOMax         MediaUnlockPlatform = "hbo_max"
	MediaUnlockPlatformAbema          MediaUnlockPlatform = "abema"
	MediaUnlockPlatformTikTok         MediaUnlockPlatform = "tiktok"
	MediaUnlockPlatformSpotify        MediaUnlockPlatform = "spotify"
	MediaUnlockPlatformSteam          MediaUnlockPlatform = "steam"
	MediaUnlockPlatformChatGPT        MediaUnlockPlatform = "chatgpt"
	MediaUnlockPlatformClaude         MediaUnlockPlatform = "claude"
	MediaUnlockPlatformGemini         MediaUnlockPlatform = "gemini"
)

type MediaUnlockPlatformResult struct {
	Platform MediaUnlockPlatform `json:"platform"`
	Name     string              `json:"name"`
	Status   MediaUnlockStatus   `json:"status"`
	Region   string              `json:"region,omitempty"`
	Mode     MediaUnlockMode     `json:"mode,omitempty"`
	Error    string              `json:"error,omitempty"`
	Display  string              `json:"display,omitempty"`
}

type MediaUnlockResult struct {
	Items []MediaUnlockPlatformResult `json:"items"`
}

type EntryResult struct {
	ProxyInfo      ProxyInfo      `json:"proxyInfo"`
	InvokeDuration int64          `json:"invokeDuration"`
	Matrices       []MatrixResult `json:"matrices"`
	Error          string         `json:"error,omitempty"`
}

type TaskRuntimePhase string

const (
	TaskRuntimePhasePreparing TaskRuntimePhase = "preparing"
	TaskRuntimePhaseMacro     TaskRuntimePhase = "macro"
	TaskRuntimePhaseMatrix    TaskRuntimePhase = "matrix"
)

type TaskActiveNode struct {
	NodeIndex   int              `json:"nodeIndex"`
	NodeName    string           `json:"nodeName"`
	NodeAddress string           `json:"nodeAddress,omitempty"`
	Protocol    ProxyType        `json:"protocol,omitempty"`
	Attempt     int              `json:"attempt,omitempty"`
	Phase       TaskRuntimePhase `json:"phase,omitempty"`
	Macro       MacroType        `json:"macro,omitempty"`
	Matrix      MatrixType       `json:"matrix,omitempty"`
	Matrices    []MatrixType     `json:"matrices,omitempty"`
	UpdatedAt   string           `json:"updatedAt,omitempty"`
}

type TaskState struct {
	TaskID      string           `json:"taskId"`
	Name        string           `json:"name"`
	Status      string           `json:"status"`
	Progress    int              `json:"progress"`
	Total       int              `json:"total"`
	Queuing     int              `json:"queuing"`
	StartedAt   string           `json:"startedAt"`
	FinishedAt  string           `json:"finishedAt,omitempty"`
	ActiveNodes []TaskActiveNode `json:"activeNodes,omitempty"`
}

type EventType string

const (
	EventProcess EventType = "process"
	EventExit    EventType = "exit"
)

type Event struct {
	Type     EventType     `json:"type"`
	TaskID   string        `json:"taskId"`
	Index    int           `json:"index,omitempty"`
	Queuing  int           `json:"queuing,omitempty"`
	Result   EntryResult   `json:"result,omitempty"`
	Results  []EntryResult `json:"results,omitempty"`
	Task     TaskState     `json:"task"`
	ExitCode string        `json:"exitCode,omitempty"`
}

type Macro interface {
	Type() MacroType
	Run(proxy Vendor, task *Task) error
}

type Matrix interface {
	Type() MatrixType
	MacroJob() MacroType
	Extract(entry MatrixEntry, macro Macro)
	Payload() any
}
