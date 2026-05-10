package app

import (
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const (
	DefaultAppLanguage = "en"
	settingsFileName   = "settings.json"
)

type Config struct {
	Mode                interfaces.Mode
	ConnConcurrency     uint
	SpeedConcurrency    uint
	SpeedInterval       time.Duration
	ProfilesPath        string
	SettingsPath        string
	ProfileFetchTimeout time.Duration
	DNS                 interfaces.DNSConfig
}

type Paths struct {
	ProfilesPath string `json:"profilesPath"`
	SettingsPath string `json:"settingsPath"`
}

type Settings struct {
	Language                string               `json:"language"`
	DNS                     interfaces.DNSConfig `json:"dns"`
	HideProfileNameInExport bool                 `json:"hideProfileNameInExport"`
	HideCNInboundInExport   bool                 `json:"hideCNInboundInExport"`
}

type SettingsPatch struct {
	Language                string                `json:"language,omitempty"`
	DNS                     *interfaces.DNSConfig `json:"dns,omitempty"`
	HideProfileNameInExport *bool                 `json:"hideProfileNameInExport,omitempty"`
	HideCNInboundInExport   *bool                 `json:"hideCNInboundInExport,omitempty"`
}

type Status struct {
	ProfilesPath     string                   `json:"profilesPath"`
	SettingsPath     string                   `json:"settingsPath,omitempty"`
	Runtime          interfaces.RuntimeStatus `json:"runtime"`
	MihomoVersion    string                   `json:"mihomoVersion,omitempty"`
	NetworkEnv       interfaces.BackendInfo   `json:"networkEnv"`
	Remote           interfaces.ClusterStatus `json:"remote,omitempty"`
	LastNetworkProbe string                   `json:"lastNetworkProbe,omitempty"`
	NetworkProbeBusy bool                     `json:"networkProbeBusy,omitempty"`
}

type ProfileSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
	NodeCount      int    `json:"nodeCount"`
	RemainingBytes uint64 `json:"remainingBytes,omitempty"`
	ExpiresAt      string `json:"expiresAt,omitempty"`
}

type ProfileTaskRequest struct {
	ProfileID string                `json:"profileId"`
	Name      string                `json:"name,omitempty"`
	Preset    string                `json:"preset"`
	Presets   []string              `json:"presets,omitempty"`
	Config    interfaces.TaskConfig `json:"config,omitempty"`
}

type RemoteProfileTaskRequest struct {
	ProfileID string                `json:"profileId"`
	KnightIDs []string              `json:"knightIds"`
	Name      string                `json:"name,omitempty"`
	Preset    string                `json:"preset"`
	Presets   []string              `json:"presets,omitempty"`
	Config    interfaces.TaskConfig `json:"config,omitempty"`
}

type ProfileTaskBuildInfo struct {
	SelectedPresets []string `json:"selectedPresets"`
	PresetLabel     string   `json:"presetLabel"`
}
