package app

import (
	"fmt"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

var taskPresetOrder = []string{"ping", "geo", "udp", "speed", "media"}

func (s *Service) SubmitProfileTask(req ProfileTaskRequest) (string, error) {
	task, _, err := s.BuildProfileTask(req)
	if err != nil {
		return "", err
	}

	resp, err := s.api.SubmitTask(interfaces.TaskSubmitRequest{Task: task}, nil)
	if err != nil {
		return "", err
	}
	return resp.TaskID, nil
}

func (s *Service) SubmitRemoteProfileTask(req RemoteProfileTaskRequest) ([]interfaces.ClusterRemoteTask, error) {
	task, _, err := s.BuildProfileTask(ProfileTaskRequest{
		ProfileID: req.ProfileID,
		Name:      req.Name,
		Preset:    req.Preset,
		Presets:   req.Presets,
		Config:    req.Config,
	})
	if err != nil {
		return nil, err
	}

	resp, err := s.api.DispatchRemoteTask(interfaces.ClusterDispatchTaskRequest{
		KnightIDs: append([]string{}, req.KnightIDs...),
		Task:      task,
	})
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

func (s *Service) BuildProfileTask(req ProfileTaskRequest) (interfaces.Task, ProfileTaskBuildInfo, error) {
	if s == nil || s.api == nil {
		return interfaces.Task{}, ProfileTaskBuildInfo{}, errNilService
	}

	profileResp, err := s.api.GetProfile(req.ProfileID)
	if err != nil {
		return interfaces.Task{}, ProfileTaskBuildInfo{}, err
	}
	return BuildProfileTaskFromProfile(profileResp.Profile, req, s.now)
}

func BuildProfileTaskFromProfile(profile interfaces.Profile, req ProfileTaskRequest, now func() time.Time) (interfaces.Task, ProfileTaskBuildInfo, error) {
	if len(profile.Nodes) == 0 {
		return interfaces.Task{}, ProfileTaskBuildInfo{}, fmt.Errorf("profile has no nodes")
	}

	selectedNodes := EnabledNodes(profile.Nodes)
	if len(selectedNodes) == 0 {
		return interfaces.Task{}, ProfileTaskBuildInfo{}, fmt.Errorf("profile has no enabled nodes")
	}

	selectedPresets := NormalizeTaskPresets(req.Presets, req.Preset)
	matrices, err := PresetMatrices(selectedPresets)
	if err != nil {
		return interfaces.Task{}, ProfileTaskBuildInfo{}, err
	}

	presetLabel := FormatPresetLabel(selectedPresets)
	taskName := strings.TrimSpace(req.Name)
	if taskName == "" {
		taskName = DefaultTaskName(profile.Name, presetLabel, now)
	}

	return interfaces.Task{
			Name:   taskName,
			Vendor: interfaces.VendorMihomo,
			Context: interfaces.TaskContext{
				Preset:        presetLabel,
				ProfileID:     profile.ID,
				ProfileName:   profile.Name,
				ProfileSource: profile.Source,
			},
			Nodes:    selectedNodes,
			Matrices: matrices,
			Config:   req.Config.Normalize(),
		}, ProfileTaskBuildInfo{
			SelectedPresets: selectedPresets,
			PresetLabel:     presetLabel,
		}, nil
}

func EnabledNodes(nodes []interfaces.Node) []interfaces.Node {
	selected := make([]interfaces.Node, 0, len(nodes))
	for _, node := range nodes {
		if !node.Enabled {
			continue
		}
		selected = append(selected, node)
	}
	return selected
}

func PresetMatrices(presets []string) ([]interfaces.MatrixEntry, error) {
	selected := NormalizeTaskPresets(presets, "")
	if len(selected) == 0 {
		selected = []string{"ping"}
	}

	matrixMap := map[interfaces.MatrixType]interfaces.MatrixEntry{}
	matrixOrder := make([]interfaces.MatrixType, 0, 10)
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
		case "udp":
			appendMatrix(interfaces.MatrixEntry{Type: interfaces.MatrixUDPNATType})
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

func DefaultTaskName(profileName string, preset string, now func() time.Time) string {
	name := strings.TrimSpace(profileName)
	if name == "" {
		name = "Profile"
	}
	preset = strings.ToUpper(strings.TrimSpace(preset))
	if preset == "" {
		preset = "PING"
	}
	if now == nil {
		now = time.Now
	}
	return fmt.Sprintf("%s %s %s", name, preset, now().Format("15:04:05"))
}

func NormalizeTaskPresets(presets []string, fallback string) []string {
	input := append([]string{}, presets...)
	if len(input) == 0 && strings.TrimSpace(fallback) != "" {
		input = strings.FieldsFunc(fallback, func(r rune) bool {
			return r == '+' || r == ','
		})
	}

	selected := map[string]struct{}{}
	for _, preset := range input {
		switch strings.ToLower(strings.TrimSpace(preset)) {
		case "full":
			for _, item := range taskPresetOrder {
				selected[item] = struct{}{}
			}
		case "ping", "geo", "udp", "speed", "media":
			selected[strings.ToLower(strings.TrimSpace(preset))] = struct{}{}
		}
	}

	out := make([]string, 0, len(selected))
	for _, preset := range taskPresetOrder {
		if _, ok := selected[preset]; ok {
			out = append(out, preset)
		}
	}
	return out
}

func FormatPresetLabel(presets []string) string {
	selected := NormalizeTaskPresets(presets, "")
	if len(selected) == 0 {
		return "ping"
	}
	if len(selected) == len(taskPresetOrder) {
		return "full"
	}
	return strings.Join(selected, "+")
}
