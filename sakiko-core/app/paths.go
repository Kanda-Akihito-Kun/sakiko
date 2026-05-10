package app

import (
	"os"
	"path/filepath"
)

func ResolvePaths(cfg Config) (Paths, error) {
	profilesPath := cfg.ProfilesPath
	if profilesPath == "" {
		resolved, err := ResolveProfilesPath()
		if err != nil {
			return Paths{}, err
		}
		profilesPath = resolved
	}

	settingsPath := cfg.SettingsPath
	if settingsPath == "" {
		resolved, err := ResolveSettingsPath()
		if err != nil {
			return Paths{}, err
		}
		settingsPath = resolved
	}

	return Paths{
		ProfilesPath: profilesPath,
		SettingsPath: settingsPath,
	}, nil
}

func ResolveProfilesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sakiko", "profiles.yaml"), nil
}

func ResolveSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sakiko", settingsFileName), nil
}
