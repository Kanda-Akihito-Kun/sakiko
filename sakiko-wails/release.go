package main

import (
	"context"
	"strings"
	"time"

	coreapp "sakiko.local/sakiko-core/app"

	"github.com/pkg/browser"
	"go.uber.org/zap"
)

const (
	appVersion          = coreapp.DefaultAppVersion
	releasePageURL      = coreapp.DefaultReleasePageURL
	latestReleaseAPIURL = coreapp.DefaultLatestReleaseAPIURL
	updateCheckTimeout  = coreapp.DefaultUpdateCheckTimeout
)

type ReleaseCheckResult = coreapp.ReleaseCheckResult
type githubLatestReleaseResponse = coreapp.GithubLatestReleaseResponse

func (s *SakikoService) CheckForUpdates() (ReleaseCheckResult, error) {
	result, err := coreapp.CheckForUpdates(context.Background(), releaseConfig())
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	wailsServiceLogger().Info("update check completed",
		zap.String("current_version", result.CurrentVersion),
		zap.String("latest_version", result.LatestVersion),
		zap.Bool("has_update", result.HasUpdate),
	)
	return result, nil
}

func (s *SakikoService) OpenReleasePage(url string) error {
	targetURL := strings.TrimSpace(url)
	if targetURL == "" {
		targetURL = releasePageURL
	}
	if err := browser.OpenURL(targetURL); err != nil {
		return err
	}
	wailsServiceLogger().Info("release page opened", zap.String("url", targetURL))
	return nil
}

func releaseConfig() coreapp.ReleaseConfig {
	return coreapp.ReleaseConfig{
		CurrentVersion: appVersion,
		ReleasePageURL: releasePageURL,
		LatestAPIURL:   latestReleaseAPIURL,
		UserAgent:      "sakiko-wails/" + appVersion,
		Timeout:        updateCheckTimeout,
	}
}

func buildReleaseCheckResult(payload githubLatestReleaseResponse, now time.Time) ReleaseCheckResult {
	cfg := releaseConfig()
	cfg.Now = func() time.Time {
		return now
	}
	return coreapp.BuildReleaseCheckResult(payload, cfg)
}

func displayReleaseVersion(source string) string {
	return coreapp.DisplayReleaseVersion(source)
}

func normalizeReleaseVersion(source string) string {
	return coreapp.NormalizeReleaseVersion(source)
}

func compareReleaseVersions(left string, right string) int {
	return coreapp.CompareReleaseVersions(left, right)
}
