package main

import coreapp "sakiko.local/sakiko-core/app"

const (
	defaultAppLanguage = coreapp.DefaultAppLanguage
	settingsFileName   = "settings.json"
)

type AppSettings = coreapp.Settings
type AppSettingsPatch = coreapp.SettingsPatch

func normalizeAppLanguage(value string) string {
	return coreapp.NormalizeAppLanguage(value)
}

func defaultAppSettings() AppSettings {
	return coreapp.DefaultSettings()
}

func resolveSettingsPath() (string, error) {
	return coreapp.ResolveSettingsPath()
}

func loadAppSettings(path string) (AppSettings, error) {
	return coreapp.LoadSettings(path)
}

func saveAppSettings(path string, settings AppSettings) error {
	return coreapp.SaveSettings(path, settings)
}
