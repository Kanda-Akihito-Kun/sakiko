package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

const (
	appVersion          = "0.1.0"
	releasePageURL      = "https://github.com/Kanda-Akihito-Kun/sakiko/releases"
	latestReleaseAPIURL = "https://api.github.com/repos/Kanda-Akihito-Kun/sakiko/releases/latest"
	updateCheckTimeout  = 8 * time.Second
)

type ReleaseCheckResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	ReleaseName    string `json:"releaseName,omitempty"`
	ReleaseURL     string `json:"releaseURL,omitempty"`
	HasUpdate      bool   `json:"hasUpdate"`
	CheckedAt      string `json:"checkedAt,omitempty"`
}

type githubLatestReleaseResponse struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

func (s *SakikoService) CheckForUpdates() (ReleaseCheckResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPIURL, nil)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "sakiko-wails/"+appVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ReleaseCheckResult{}, fmt.Errorf("update check failed: github returned %s", resp.Status)
	}

	var payload githubLatestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ReleaseCheckResult{}, err
	}

	result := buildReleaseCheckResult(payload, time.Now())
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

func buildReleaseCheckResult(payload githubLatestReleaseResponse, now time.Time) ReleaseCheckResult {
	currentVersion := displayReleaseVersion(appVersion)
	latestVersion := displayReleaseVersion(payload.TagName)
	releaseURL := strings.TrimSpace(payload.HTMLURL)
	if releaseURL == "" {
		releaseURL = releasePageURL
	}

	return ReleaseCheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseName:    strings.TrimSpace(payload.Name),
		ReleaseURL:     releaseURL,
		HasUpdate:      compareReleaseVersions(latestVersion, currentVersion) > 0,
		CheckedAt:      now.Format(time.RFC3339),
	}
}

func displayReleaseVersion(source string) string {
	if normalized := normalizeReleaseVersion(source); normalized != "" {
		return normalized
	}
	return strings.TrimSpace(source)
}

func normalizeReleaseVersion(source string) string {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(trimmed, "V"):
		trimmed = "v" + trimmed[1:]
	case !strings.HasPrefix(trimmed, "v"):
		trimmed = "v" + trimmed
	}

	if !semver.IsValid(trimmed) {
		return ""
	}
	return trimmed
}

func compareReleaseVersions(left string, right string) int {
	normalizedLeft := normalizeReleaseVersion(left)
	normalizedRight := normalizeReleaseVersion(right)

	switch {
	case normalizedLeft != "" && normalizedRight != "":
		return semver.Compare(normalizedLeft, normalizedRight)
	case strings.EqualFold(strings.TrimPrefix(left, "v"), strings.TrimPrefix(right, "v")):
		return 0
	case left == "":
		return -1
	case right == "":
		return 1
	default:
		return strings.Compare(left, right)
	}
}
