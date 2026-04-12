package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	coreapi "sakiko.local/sakiko-core/api"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"

	"go.uber.org/zap"
)

type SakikoService struct {
	api          *coreapi.Service
	profilesPath string
	once         sync.Once
	initErr      error
}

type ProfileTaskSubmitRequest struct {
	ProfileID string                `json:"profileId"`
	Name      string                `json:"name,omitempty"`
	Preset    string                `json:"preset"`
	Presets   []string              `json:"presets,omitempty"`
	Config    interfaces.TaskConfig `json:"config,omitempty"`
}

type DesktopStatus struct {
	ProfilesPath string                   `json:"profilesPath"`
	Runtime      interfaces.RuntimeStatus `json:"runtime"`
}

type ProfileSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	NodeCount int    `json:"nodeCount"`
}

func (s *SakikoService) DesktopStatus() (DesktopStatus, error) {
	if err := s.ensureReady(); err != nil {
		return DesktopStatus{}, err
	}

	return DesktopStatus{
		ProfilesPath: s.profilesPath,
		Runtime:      s.api.RuntimeStatus().Status,
	}, nil
}

func (s *SakikoService) ListProfileSummaries() ([]ProfileSummary, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	profiles := s.api.ListProfiles().Profiles
	summaries := make([]ProfileSummary, 0, len(profiles))
	for _, profile := range profiles {
		summaries = append(summaries, ProfileSummary{
			ID:        profile.ID,
			Name:      profile.Name,
			Source:    profile.Source,
			UpdatedAt: profile.UpdatedAt,
			NodeCount: len(profile.Nodes),
		})
	}
	return summaries, nil
}

func (s *SakikoService) ListProfiles() ([]interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.api.ListProfiles().Profiles, nil
}

func (s *SakikoService) GetProfile(profileID string) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	resp, err := s.api.GetProfile(profileID)
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *SakikoService) SetProfileNodeEnabled(profileID string, nodeIndex int, enabled bool) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}

	resp, err := s.api.UpdateProfileNodeSelection(interfaces.ProfileNodeSelectionUpdateRequest{
		ProfileID: profileID,
		NodeIndex: nodeIndex,
		Enabled:   enabled,
	})
	if err != nil {
		wailsServiceLogger().Warn("update profile node selection failed",
			zap.String("profile_id", profileID),
			zap.Int("node_index", nodeIndex),
			zap.Bool("enabled", enabled),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}

	return resp.Profile, nil
}

func (s *SakikoService) ListDownloadTargets() ([]interfaces.DownloadTarget, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	resp, err := s.api.ListDownloadTargets()
	if err != nil {
		return nil, err
	}
	return resp.Targets, nil
}

func (s *SakikoService) ImportProfile(req interfaces.ProfileImportRequest) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	resp, err := s.api.ImportProfile(req)
	if err != nil {
		wailsServiceLogger().Warn("import profile failed", zap.Error(err))
		return interfaces.Profile{}, err
	}
	wailsServiceLogger().Info("profile imported",
		zap.String("profile_id", resp.Profile.ID),
		zap.Int("node_count", len(resp.Profile.Nodes)),
	)
	return resp.Profile, nil
}

func (s *SakikoService) RefreshProfile(profileID string) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	resp, err := s.api.RefreshProfile(interfaces.ProfileRefreshRequest{ProfileID: profileID})
	if err != nil {
		wailsServiceLogger().Warn("refresh profile failed",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}
	wailsServiceLogger().Info("profile refreshed",
		zap.String("profile_id", resp.Profile.ID),
		zap.Int("node_count", len(resp.Profile.Nodes)),
	)
	return resp.Profile, nil
}

func (s *SakikoService) DeleteProfile(profileID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	_, err := s.api.DeleteProfile(interfaces.ProfileDeleteRequest{ProfileID: profileID})
	if err != nil {
		wailsServiceLogger().Warn("delete profile failed",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return err
	}
	wailsServiceLogger().Info("profile deleted", zap.String("profile_id", profileID))
	return nil
}

func (s *SakikoService) ListTasks() ([]interfaces.TaskState, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.api.ListTasks().Tasks, nil
}

func (s *SakikoService) GetTask(taskID string) (interfaces.TaskStatusResponse, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.TaskStatusResponse{}, err
	}
	return s.api.GetTask(taskID)
}

func (s *SakikoService) ListResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	resp, err := s.api.ListResultArchives()
	if err != nil {
		return nil, err
	}
	return resp.Archives, nil
}

func (s *SakikoService) GetResultArchive(taskID string) (interfaces.ResultArchive, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ResultArchive{}, err
	}
	resp, err := s.api.GetResultArchive(taskID)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}
	return resp.Archive, nil
}

func (s *SakikoService) DeleteResultArchive(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	_, err := s.api.DeleteResultArchive(interfaces.ResultArchiveDeleteRequest{TaskID: taskID})
	if err != nil {
		wailsServiceLogger().Warn("delete result archive failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}
	wailsServiceLogger().Info("result archive deleted", zap.String("task_id", taskID))
	return nil
}

func (s *SakikoService) SubmitProfileTask(req ProfileTaskSubmitRequest) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	wailsServiceLogger().Info("submit profile task requested",
		zap.String("profile_id", req.ProfileID),
		zap.String("preset", req.Preset),
		zap.Strings("presets", req.Presets),
		zap.String("task_name", req.Name),
	)

	profileResp, err := s.api.GetProfile(req.ProfileID)
	if err != nil {
		wailsServiceLogger().Warn("load profile for task failed",
			zap.String("profile_id", req.ProfileID),
			zap.Error(err),
		)
		return "", err
	}
	if len(profileResp.Profile.Nodes) == 0 {
		wailsServiceLogger().Warn("submit profile task rejected: profile has no nodes",
			zap.String("profile_id", req.ProfileID),
		)
		return "", fmt.Errorf("profile has no nodes")
	}
	selectedNodes := enabledNodes(profileResp.Profile.Nodes)
	if len(selectedNodes) == 0 {
		wailsServiceLogger().Warn("submit profile task rejected: profile has no enabled nodes",
			zap.String("profile_id", req.ProfileID),
		)
		return "", fmt.Errorf("profile has no enabled nodes")
	}

	selectedPresets := normalizeTaskPresets(req.Presets, req.Preset)
	matrices, err := presetMatrices(selectedPresets)
	if err != nil {
		wailsServiceLogger().Warn("resolve task preset failed",
			zap.String("profile_id", req.ProfileID),
			zap.String("preset", req.Preset),
			zap.Strings("presets", selectedPresets),
			zap.Error(err),
		)
		return "", err
	}
	presetLabel := formatPresetLabel(selectedPresets)

	taskName := strings.TrimSpace(req.Name)
	if taskName == "" {
		taskName = defaultTaskName(profileResp.Profile.Name, presetLabel)
	}

	resp, err := s.api.SubmitTask(interfaces.TaskSubmitRequest{
		Task: interfaces.Task{
			Name:   taskName,
			Vendor: interfaces.VendorMihomo,
			Context: interfaces.TaskContext{
				Preset:        presetLabel,
				ProfileID:     profileResp.Profile.ID,
				ProfileName:   profileResp.Profile.Name,
				ProfileSource: profileResp.Profile.Source,
			},
			Nodes:    selectedNodes,
			Matrices: matrices,
			Config:   req.Config.Normalize(),
		},
	}, nil)
	if err != nil {
		wailsServiceLogger().Warn("submit profile task failed",
			zap.String("profile_id", req.ProfileID),
			zap.String("preset", presetLabel),
			zap.Strings("presets", selectedPresets),
			zap.Error(err),
		)
		return "", err
	}
	wailsServiceLogger().Info("profile task submitted",
		zap.String("profile_id", req.ProfileID),
		zap.String("task_id", resp.TaskID),
		zap.String("preset", presetLabel),
		zap.Strings("presets", selectedPresets),
		zap.Int("node_count", len(selectedNodes)),
	)
	return resp.TaskID, nil
}

func enabledNodes(nodes []interfaces.Node) []interfaces.Node {
	selected := make([]interfaces.Node, 0, len(nodes))
	for _, node := range nodes {
		if !node.Enabled {
			continue
		}
		selected = append(selected, node)
	}
	return selected
}

func resolveProfilesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sakiko", "profiles.yaml"), nil
}

func presetMatrices(presets []string) ([]interfaces.MatrixEntry, error) {
	selected := normalizeTaskPresets(presets, "")
	if len(selected) == 0 {
		selected = []string{"ping"}
	}

	matrixMap := map[interfaces.MatrixType]interfaces.MatrixEntry{}
	matrixOrder := make([]interfaces.MatrixType, 0, 9)
	appendMatrix := func(entry interfaces.MatrixEntry) {
		if _, exists := matrixMap[entry.Type]; exists {
			return
		}
		matrixMap[entry.Type] = entry
		matrixOrder = append(matrixOrder, entry.Type)
	}

	for _, preset := range selected {
		switch preset {
		case "ping":
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixHTTPPing})
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixRTTPing})
		case "geo":
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixInboundGeoIP})
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixOutboundGeoIP})
		case "speed":
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixAverageSpeed})
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixMaxSpeed})
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixPerSecSpeed})
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixTrafficUsed})
		case "media":
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixMediaUnlock})
		default:
			return nil, fmt.Errorf("unsupported task preset: %s", preset)
		}
	}

	out := make([]interfaces.MatrixEntry, 0, len(matrixOrder))
	for _, matrixType := range matrixOrder {
		out = append(out, matrixMap[matrixType])
	}
	return out, nil
}

func defaultTaskName(profileName string, preset string) string {
	name := strings.TrimSpace(profileName)
	if name == "" {
		name = "Profile"
	}
	preset = strings.ToUpper(strings.TrimSpace(preset))
	if preset == "" {
		preset = "PING"
	}
	return fmt.Sprintf("%s %s %s", name, preset, time.Now().Format("15:04:05"))
}

func normalizeTaskPresets(presets []string, fallback string) []string {
	input := append([]string{}, presets...)
	if len(input) == 0 && strings.TrimSpace(fallback) != "" {
		input = strings.FieldsFunc(fallback, func(r rune) bool {
			return r == '+' || r == ','
		})
	}

	selected := map[string]struct{}{}
	order := []string{"ping", "geo", "speed", "media"}

	for _, preset := range input {
		switch strings.ToLower(strings.TrimSpace(preset)) {
		case "full":
			for _, item := range order {
				selected[item] = struct{}{}
			}
		case "ping", "geo", "speed", "media":
			selected[strings.ToLower(strings.TrimSpace(preset))] = struct{}{}
		}
	}

	out := make([]string, 0, len(selected))
	for _, preset := range order {
		if _, ok := selected[preset]; ok {
			out = append(out, preset)
		}
	}
	return out
}

func formatPresetLabel(presets []string) string {
	selected := normalizeTaskPresets(presets, "")
	if len(selected) == 0 {
		return "ping"
	}
	if len(selected) == 4 {
		return "full"
	}
	return strings.Join(selected, "+")
}

func (s *SakikoService) ensureReady() error {
	if s == nil {
		return fmt.Errorf("sakiko service is nil")
	}

	s.once.Do(func() {
		wailsServiceLogger().Info("initializing sakiko service")
		profilesPath, err := resolveProfilesPath()
		if err != nil {
			wailsServiceLogger().Error("resolve profiles path failed", zap.Error(err))
			s.initErr = err
			return
		}
		wailsServiceLogger().Info("resolved profiles path", zap.String("profiles_path", profilesPath))

		apiService, err := coreapi.New(coreapi.Config{
			Mode:                interfaces.ModeParallel,
			ConnConcurrency:     24,
			SpeedConcurrency:    1,
			SpeedInterval:       300 * time.Millisecond,
			ProfilesPath:        profilesPath,
			ProfileFetchTimeout: 20 * time.Second,
		})
		if err != nil {
			wailsServiceLogger().Error("initialize core api failed", zap.Error(err))
			s.initErr = err
			return
		}

		s.api = apiService
		s.profilesPath = profilesPath
		wailsServiceLogger().Info("sakiko service ready", zap.String("profiles_path", profilesPath))
	})

	return s.initErr
}

func wailsServiceLogger() *zap.Logger {
	return logx.Named("service")
}
