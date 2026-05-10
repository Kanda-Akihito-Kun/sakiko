package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	DefaultReleasePageURL      = "https://github.com/Kanda-Akihito-Kun/sakiko/releases"
	DefaultLatestReleaseAPIURL = "https://api.github.com/repos/Kanda-Akihito-Kun/sakiko/releases/latest"
	DefaultUpdateCheckTimeout  = 8 * time.Second
)

type ReleaseConfig struct {
	CurrentVersion string
	ReleasePageURL string
	LatestAPIURL   string
	UserAgent      string
	HTTPClient     *http.Client
	Timeout        time.Duration
	Now            func() time.Time
}

type ReleaseCheckResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	ReleaseName    string `json:"releaseName,omitempty"`
	ReleaseURL     string `json:"releaseURL,omitempty"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	HasUpdate      bool   `json:"hasUpdate"`
	CheckedAt      string `json:"checkedAt,omitempty"`
}

type GithubLatestReleaseResponse struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

func CheckForUpdates(ctx context.Context, cfg ReleaseConfig) (ReleaseCheckResult, error) {
	cfg = cfg.Normalize()
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.LatestAPIURL, nil)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", cfg.UserAgent)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ReleaseCheckResult{}, fmt.Errorf("update check failed: github returned %s", resp.Status)
	}

	var payload GithubLatestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ReleaseCheckResult{}, err
	}
	return BuildReleaseCheckResult(payload, cfg), nil
}

func (cfg ReleaseConfig) Normalize() ReleaseConfig {
	if strings.TrimSpace(cfg.CurrentVersion) == "" {
		cfg.CurrentVersion = DefaultAppVersion
	}
	if strings.TrimSpace(cfg.ReleasePageURL) == "" {
		cfg.ReleasePageURL = DefaultReleasePageURL
	}
	if strings.TrimSpace(cfg.LatestAPIURL) == "" {
		cfg.LatestAPIURL = DefaultLatestReleaseAPIURL
	}
	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = "sakiko/" + cfg.CurrentVersion
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultUpdateCheckTimeout
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return cfg
}

func BuildReleaseCheckResult(payload GithubLatestReleaseResponse, cfg ReleaseConfig) ReleaseCheckResult {
	cfg = cfg.Normalize()
	currentVersion := DisplayReleaseVersion(cfg.CurrentVersion)
	latestVersion := DisplayReleaseVersion(payload.TagName)
	releaseURL := strings.TrimSpace(payload.HTMLURL)
	if releaseURL == "" {
		releaseURL = cfg.ReleasePageURL
	}

	return ReleaseCheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseName:    strings.TrimSpace(payload.Name),
		ReleaseURL:     releaseURL,
		ReleaseNotes:   strings.TrimSpace(payload.Body),
		HasUpdate:      CompareReleaseVersions(latestVersion, currentVersion) > 0,
		CheckedAt:      cfg.Now().Format(time.RFC3339),
	}
}

func DisplayReleaseVersion(source string) string {
	if normalized := NormalizeReleaseVersion(source); normalized != "" {
		return normalized
	}
	return strings.TrimSpace(source)
}

func NormalizeReleaseVersion(source string) string {
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

func CompareReleaseVersions(left string, right string) int {
	normalizedLeft := NormalizeReleaseVersion(left)
	normalizedRight := NormalizeReleaseVersion(right)

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
