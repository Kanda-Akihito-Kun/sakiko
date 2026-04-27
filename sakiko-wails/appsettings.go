package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"sakiko.local/sakiko-core/interfaces"
)

const (
	defaultAppLanguage = "en"
	settingsFileName   = "settings.json"
)

type AppSettings struct {
	Language                string               `json:"language"`
	DNS                     interfaces.DNSConfig `json:"dns"`
	HideProfileNameInExport bool                 `json:"hideProfileNameInExport"`
	HideCNInboundInExport   bool                 `json:"hideCNInboundInExport"`
}

type AppSettingsPatch struct {
	Language                string                `json:"language,omitempty"`
	DNS                     *interfaces.DNSConfig `json:"dns,omitempty"`
	HideProfileNameInExport *bool                 `json:"hideProfileNameInExport,omitempty"`
	HideCNInboundInExport   *bool                 `json:"hideCNInboundInExport,omitempty"`
}

func (s AppSettings) Normalize() AppSettings {
	return AppSettings{
		Language:                normalizeAppLanguage(s.Language),
		DNS:                     s.DNS.Normalize(),
		HideProfileNameInExport: s.HideProfileNameInExport,
		HideCNInboundInExport:   s.HideCNInboundInExport,
	}
}

func normalizeAppLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "zh", "zh-cn", "zh_hans", "zh-hans":
		return "zh"
	case "en", "en-us", "en_us", "en-gb", "en_gb":
		return "en"
	default:
		return defaultAppLanguage
	}
}

func defaultAppSettings() AppSettings {
	return AppSettings{
		Language:                defaultAppLanguage,
		DNS:                     interfaces.DefaultDNSConfig(),
		HideProfileNameInExport: true,
		HideCNInboundInExport:   true,
	}
}

func resolveSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sakiko", settingsFileName), nil
}

func loadAppSettings(path string) (AppSettings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultAppSettings(), nil
		}
		return AppSettings{}, err
	}

	settings := defaultAppSettings()
	if err := json.Unmarshal(raw, &settings); err != nil {
		return AppSettings{}, err
	}
	return settings.Normalize(), nil
}

func saveAppSettings(path string, settings AppSettings) error {
	settings = settings.Normalize()
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(path, raw, 0o644)
}
