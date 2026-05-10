package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"sakiko.local/sakiko-core/interfaces"
)

func (s Settings) Normalize() Settings {
	return Settings{
		Language:                NormalizeAppLanguage(s.Language),
		DNS:                     s.DNS.Normalize(),
		HideProfileNameInExport: s.HideProfileNameInExport,
		HideCNInboundInExport:   s.HideCNInboundInExport,
	}
}

func NormalizeAppLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "zh", "zh-cn", "zh_hans", "zh-hans":
		return "zh"
	case "en", "en-us", "en_us", "en-gb", "en_gb":
		return "en"
	default:
		return DefaultAppLanguage
	}
}

func DefaultSettings() Settings {
	return Settings{
		Language:                DefaultAppLanguage,
		DNS:                     interfaces.DefaultDNSConfig(),
		HideProfileNameInExport: true,
		HideCNInboundInExport:   true,
	}
}

func LoadSettings(path string) (Settings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return Settings{}, err
	}

	settings := DefaultSettings()
	if err := json.Unmarshal(raw, &settings); err != nil {
		return Settings{}, err
	}
	return settings.Normalize(), nil
}

func SaveSettings(path string, settings Settings) error {
	settings = settings.Normalize()
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, raw, 0o644)
}

func ApplySettingsPatch(settings Settings, patch SettingsPatch) Settings {
	if strings.TrimSpace(patch.Language) != "" {
		settings.Language = patch.Language
	}
	if patch.DNS != nil {
		settings.DNS = patch.DNS.Normalize()
	}
	if patch.HideProfileNameInExport != nil {
		settings.HideProfileNameInExport = *patch.HideProfileNameInExport
	}
	if patch.HideCNInboundInExport != nil {
		settings.HideCNInboundInExport = *patch.HideCNInboundInExport
	}
	return settings.Normalize()
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	_ = os.Remove(path)
	return os.Rename(tmpPath, path)
}
