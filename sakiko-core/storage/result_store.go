package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const resultArchiveVersion = 1
const resultArchiveSummarySuffix = ".meta.json"

type ResultStore struct {
	dir string
}

func NewResultStore(indexPath string) *ResultStore {
	baseDir := filepath.Dir(indexPath)
	if baseDir == "." || baseDir == "" {
		baseDir = ""
	}

	resultDir := "results"
	if baseDir != "" {
		resultDir = filepath.Join(baseDir, "results")
	}

	return &ResultStore{dir: resultDir}
}

func (s *ResultStore) SaveTaskArchive(snapshot interfaces.TaskArchiveSnapshot) error {
	if s == nil {
		return fmt.Errorf("result store is nil")
	}
	if snapshot.Task.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if s.dir == "" {
		return fmt.Errorf("result directory is required")
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	archive := buildResultArchive(snapshot)
	raw, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return err
	}

	if err := writeFileAtomic(s.Path(snapshot.Task.ID), raw, 0o644); err != nil {
		return err
	}

	return s.saveSummary(buildResultArchiveListItem(archive))
}

func (s *ResultStore) Path(taskID string) string {
	if s == nil || s.dir == "" || taskID == "" {
		return ""
	}
	return filepath.Join(s.dir, taskID+".json")
}

func (s *ResultStore) summaryPath(taskID string) string {
	if s == nil || s.dir == "" || taskID == "" {
		return ""
	}
	return filepath.Join(s.dir, taskID+resultArchiveSummarySuffix)
}

func (s *ResultStore) Load(taskID string) (interfaces.ResultArchive, error) {
	if s == nil {
		return interfaces.ResultArchive{}, fmt.Errorf("result store is nil")
	}

	path := s.Path(taskID)
	if path == "" {
		return interfaces.ResultArchive{}, fmt.Errorf("task ID is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}

	var archive interfaces.ResultArchive
	if err := json.Unmarshal(raw, &archive); err != nil {
		return interfaces.ResultArchive{}, err
	}
	return archive, nil
}

func (s *ResultStore) Delete(taskID string) error {
	if s == nil {
		return fmt.Errorf("result store is nil")
	}
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("task ID is required")
	}

	summaryPath := s.summaryPath(taskID)
	archivePath := s.Path(taskID)
	deleted := false
	errs := make([]error, 0, 2)

	if summaryPath != "" {
		if err := os.Remove(summaryPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, err)
			}
		} else {
			deleted = true
		}
	}

	if archivePath != "" {
		if err := os.Remove(archivePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, err)
			}
		} else {
			deleted = true
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	if deleted {
		return nil
	}
	return os.ErrNotExist
}

func (s *ResultStore) List() ([]interfaces.ResultArchiveListItem, error) {
	if s == nil {
		return nil, fmt.Errorf("result store is nil")
	}
	if s.dir == "" {
		return []interfaces.ResultArchiveListItem{}, nil
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []interfaces.ResultArchiveListItem{}, nil
		}
		return nil, err
	}

	itemsByTask := make(map[string]interfaces.ResultArchiveListItem, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		lowerName := strings.ToLower(name)
		if !strings.HasSuffix(lowerName, resultArchiveSummarySuffix) {
			continue
		}

		taskID := strings.TrimSuffix(name, resultArchiveSummarySuffix)
		item, err := s.loadSummary(taskID)
		if err != nil {
			return nil, err
		}
		itemsByTask[taskID] = item
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		lowerName := strings.ToLower(name)
		if strings.HasSuffix(lowerName, resultArchiveSummarySuffix) || strings.ToLower(filepath.Ext(name)) != ".json" {
			continue
		}

		taskID := strings.TrimSuffix(name, filepath.Ext(name))
		if _, ok := itemsByTask[taskID]; ok {
			continue
		}

		archive, err := s.Load(taskID)
		if err != nil {
			return nil, err
		}

		item := buildResultArchiveListItem(archive)
		itemsByTask[taskID] = item
		if err := s.saveSummary(item); err != nil {
			return nil, err
		}
	}

	items := make([]interfaces.ResultArchiveListItem, 0, len(itemsByTask))
	for _, item := range itemsByTask {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return newerTimestamp(items[i].FinishedAt, items[j].FinishedAt)
	})
	return items, nil
}

func (s *ResultStore) loadSummary(taskID string) (interfaces.ResultArchiveListItem, error) {
	path := s.summaryPath(taskID)
	if path == "" {
		return interfaces.ResultArchiveListItem{}, fmt.Errorf("task ID is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return interfaces.ResultArchiveListItem{}, err
	}

	var item interfaces.ResultArchiveListItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return interfaces.ResultArchiveListItem{}, err
	}
	return item, nil
}

func (s *ResultStore) saveSummary(item interfaces.ResultArchiveListItem) error {
	if s == nil {
		return fmt.Errorf("result store is nil")
	}
	if item.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}

	path := s.summaryPath(item.TaskID)
	if path == "" {
		return fmt.Errorf("result summary path is required")
	}

	raw, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, raw, 0o644)
}

func buildResultArchive(snapshot interfaces.TaskArchiveSnapshot) interfaces.ResultArchive {
	return interfaces.ResultArchive{
		Version:  resultArchiveVersion,
		Task:     buildResultArchiveTask(snapshot.Task),
		State:    snapshot.State,
		Results:  append([]interfaces.EntryResult{}, snapshot.Results...),
		ExitCode: snapshot.ExitCode,
		Report:   buildResultReport(snapshot),
	}
}

func buildResultArchiveListItem(archive interfaces.ResultArchive) interfaces.ResultArchiveListItem {
	return interfaces.ResultArchiveListItem{
		TaskID:       archive.Task.ID,
		TaskName:     archive.Task.Name,
		Preset:       archive.Task.Context.Preset,
		ProfileID:    archive.Task.Context.ProfileID,
		ProfileName:  archive.Task.Context.ProfileName,
		StartedAt:    archive.State.StartedAt,
		FinishedAt:   archive.State.FinishedAt,
		ExitCode:     archive.ExitCode,
		NodeCount:    len(archive.Task.Nodes),
		SectionCount: len(archive.Report.Sections),
	}
}

func buildResultArchiveTask(task interfaces.Task) interfaces.ResultArchiveTask {
	nodes := make([]interfaces.ResultArchiveNode, 0, len(task.Nodes))
	for _, node := range task.Nodes {
		nodes = append(nodes, interfaces.ResultArchiveNode{
			Name:  node.Name,
			Order: node.Order,
		})
	}

	var environment *interfaces.TaskEnvironment
	if task.Environment != nil {
		copied := &interfaces.TaskEnvironment{
			Identity: task.Environment.Identity,
		}
		if task.Environment.Backend != nil {
			backend := *task.Environment.Backend
			copied.Backend = &backend
		}
		environment = copied
	}

	return interfaces.ResultArchiveTask{
		ID:          task.ID,
		Name:        task.Name,
		Vendor:      task.Vendor,
		Context:     task.Context,
		Environment: environment,
		Nodes:       nodes,
		Matrices:    append([]interfaces.MatrixEntry{}, task.Matrices...),
		Config:      task.Config,
	}
}

func buildResultReport(snapshot interfaces.TaskArchiveSnapshot) interfaces.ResultReport {
	generatedAt := snapshot.State.FinishedAt
	if generatedAt == "" {
		generatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	sections := make([]interfaces.ResultReportSection, 0, 2)
	if hasAnyMatrix(snapshot.Task.Matrices, interfaces.MatrixAverageSpeed, interfaces.MatrixMaxSpeed, interfaces.MatrixPerSecSpeed, interfaces.MatrixTrafficUsed) {
		sections = append(sections, buildSpeedSection(snapshot))
	}
	if hasAnyMatrix(snapshot.Task.Matrices, interfaces.MatrixInboundGeoIP, interfaces.MatrixOutboundGeoIP) {
		sections = append(sections, buildTopologySection(snapshot))
	}
	if hasAnyMatrix(snapshot.Task.Matrices, interfaces.MatrixMediaUnlock) {
		sections = append(sections, buildMediaUnlockSection(snapshot))
	}

	return interfaces.ResultReport{
		GeneratedAt: generatedAt,
		Sections:    sections,
	}
}

func buildSpeedSection(snapshot interfaces.TaskArchiveSnapshot) interfaces.ResultReportSection {
	rows := make([]map[string]any, 0, len(snapshot.Results))
	for _, result := range snapshot.Results {
		rtt, _ := extractUint64Matrix(result.Matrices, interfaces.MatrixRTTPing)
		httpPing, _ := extractUint64Matrix(result.Matrices, interfaces.MatrixHTTPPing)
		avgSpeed, _ := extractUint64Matrix(result.Matrices, interfaces.MatrixAverageSpeed)
		maxSpeed, _ := extractUint64Matrix(result.Matrices, interfaces.MatrixMaxSpeed)
		perSecond, _ := extractUint64SliceMatrix(result.Matrices, interfaces.MatrixPerSecSpeed)
		trafficUsed, trafficMeasured := extractUint64Matrix(result.Matrices, interfaces.MatrixTrafficUsed)
		rowError := speedSectionStatus(result.Error, trafficUsed, trafficMeasured)

		rows = append(rows, map[string]any{
			"nodeName":                result.ProxyInfo.Name,
			"proxyType":               result.ProxyInfo.Type,
			"address":                 result.ProxyInfo.Address,
			"rttMillis":               rtt,
			"httpPingMillis":          httpPing,
			"averageBytesPerSecond":   avgSpeed,
			"maxBytesPerSecond":       maxSpeed,
			"perSecondBytesPerSecond": perSecond,
			"trafficUsedBytes":        trafficUsed,
			"error":                   rowError,
		})
	}

	successCount := 0
	for index, row := range rows {
		row["rank"] = index + 1
		if strings.TrimSpace(fmt.Sprint(row["error"])) == "" {
			successCount++
		}
	}

	return interfaces.ResultReportSection{
		Kind:  "speed_table",
		Title: "Speed Test",
		Columns: []interfaces.ResultReportColumn{
			{Key: "rank", Label: "Rank"},
			{Key: "nodeName", Label: "Node"},
			{Key: "proxyType", Label: "Protocol"},
			{Key: "rttMillis", Label: "TLS RTT"},
			{Key: "httpPingMillis", Label: "HTTPS Ping"},
			{Key: "averageBytesPerSecond", Label: "Average Speed"},
			{Key: "maxBytesPerSecond", Label: "Max Speed"},
			{Key: "perSecondBytesPerSecond", Label: "Per-second Speed"},
			{Key: "trafficUsedBytes", Label: "Traffic Used"},
			{Key: "error", Label: "Error"},
		},
		Rows: rows,
		Summary: map[string]any{
			"nodeCount":    len(snapshot.Results),
			"successCount": successCount,
			"preset":       snapshot.Task.Context.Preset,
		},
	}
}

func speedSectionStatus(resultError string, trafficUsed uint64, trafficMeasured bool) string {
	if strings.TrimSpace(resultError) != "" {
		return resultError
	}
	if trafficMeasured && trafficUsed == 0 {
		return "Failed"
	}
	return ""
}

func buildTopologySection(snapshot interfaces.TaskArchiveSnapshot) interfaces.ResultReportSection {
	rows := make([]map[string]any, 0, len(snapshot.Results))
	successCount := 0

	for _, result := range snapshot.Results {
		inbound, _ := extractGeoMatrix(result.Matrices, interfaces.MatrixInboundGeoIP)
		outbound, _ := extractGeoMatrix(result.Matrices, interfaces.MatrixOutboundGeoIP)
		rowError := joinErrors(result.Error, inbound.Error, outbound.Error)
		if rowError == "" {
			successCount++
		}

		rows = append(rows, map[string]any{
			"nodeName":             result.ProxyInfo.Name,
			"proxyType":            result.ProxyInfo.Type,
			"address":              result.ProxyInfo.Address,
			"inboundIP":            inbound.IP,
			"inboundCountryCode":   inbound.CountryCode,
			"inboundCountry":       inbound.Country,
			"inboundCity":          inbound.City,
			"inboundASN":           inbound.ASN,
			"inboundOrganization":  preferredGeoOrganization(inbound),
			"outboundIP":           outbound.IP,
			"outboundCountryCode":  outbound.CountryCode,
			"outboundCountry":      outbound.Country,
			"outboundCity":         outbound.City,
			"outboundASN":          outbound.ASN,
			"outboundOrganization": preferredGeoOrganization(outbound),
			"error":                rowError,
		})
	}

	return interfaces.ResultReportSection{
		Kind:  "topology_table",
		Title: "Topology Analysis",
		Columns: []interfaces.ResultReportColumn{
			{Key: "nodeName", Label: "Node"},
			{Key: "proxyType", Label: "Protocol"},
			{Key: "inboundCountryCode", Label: "Inbound Region"},
			{Key: "inboundASN", Label: "Inbound ASN"},
			{Key: "inboundOrganization", Label: "Inbound Org"},
			{Key: "outboundCountryCode", Label: "Outbound Region"},
			{Key: "outboundASN", Label: "Outbound ASN"},
			{Key: "outboundOrganization", Label: "Outbound Org"},
			{Key: "outboundIP", Label: "Outbound IP"},
			{Key: "error", Label: "Error"},
		},
		Rows: rows,
		Summary: map[string]any{
			"nodeCount":    len(snapshot.Results),
			"successCount": successCount,
			"preset":       snapshot.Task.Context.Preset,
		},
	}
}

func buildMediaUnlockSection(snapshot interfaces.TaskArchiveSnapshot) interfaces.ResultReportSection {
	platforms := collectMediaPlatforms(snapshot.Results)
	columns := []interfaces.ResultReportColumn{
		{Key: "nodeName", Label: "Node"},
		{Key: "proxyType", Label: "Protocol"},
	}
	for _, platform := range platforms {
		columns = append(columns, interfaces.ResultReportColumn{
			Key:   string(platform),
			Label: mediaPlatformLabel(platform),
		})
	}

	rows := make([]map[string]any, 0, len(snapshot.Results))
	successCount := 0
	for _, result := range snapshot.Results {
		row := map[string]any{
			"nodeName":  result.ProxyInfo.Name,
			"proxyType": result.ProxyInfo.Type,
		}

		payload, ok := extractMediaUnlockMatrix(result.Matrices, interfaces.MatrixMediaUnlock)
		if !ok || len(payload.Items) == 0 {
			for _, platform := range platforms {
				row[string(platform)] = "Failed (Payload Missing)"
			}
			rows = append(rows, row)
			continue
		}

		itemsByPlatform := make(map[interfaces.MediaUnlockPlatform]interfaces.MediaUnlockPlatformResult, len(payload.Items))
		for _, item := range payload.Items {
			if !isVisibleMediaPlatform(item.Platform) {
				continue
			}
			itemsByPlatform[item.Platform] = item
			if item.Status != interfaces.MediaUnlockStatusFailed {
				successCount++
			}
		}

		for _, platform := range platforms {
			if item, ok := itemsByPlatform[platform]; ok {
				row[string(platform)] = preferredMediaDisplay(item)
				continue
			}
			row[string(platform)] = "-"
		}
		rows = append(rows, row)
	}

	return interfaces.ResultReportSection{
		Kind:    "media_unlock_table",
		Title:   "Media Unlock Matrix",
		Columns: columns,
		Rows:    rows,
		Summary: map[string]any{
			"nodeCount":     len(snapshot.Results),
			"successCount":  successCount,
			"platformCount": len(platforms),
			"preset":        snapshot.Task.Context.Preset,
		},
	}
}

func collectMediaPlatforms(results []interfaces.EntryResult) []interfaces.MediaUnlockPlatform {
	preferred := []interfaces.MediaUnlockPlatform{
		interfaces.MediaUnlockPlatformChatGPT,
		interfaces.MediaUnlockPlatformClaude,
		interfaces.MediaUnlockPlatformGemini,
		interfaces.MediaUnlockPlatformYouTubePremium,
		interfaces.MediaUnlockPlatformNetflix,
		interfaces.MediaUnlockPlatformHulu,
		interfaces.MediaUnlockPlatformPrimeVideo,
		interfaces.MediaUnlockPlatformHBOMax,
		interfaces.MediaUnlockPlatformBilibiliHMT,
		interfaces.MediaUnlockPlatformBilibiliTW,
		interfaces.MediaUnlockPlatformAbema,
		interfaces.MediaUnlockPlatformTikTok,
	}

	seen := map[interfaces.MediaUnlockPlatform]struct{}{}
	for _, result := range results {
		payload, ok := extractMediaUnlockMatrix(result.Matrices, interfaces.MatrixMediaUnlock)
		if !ok {
			continue
		}
		for _, item := range payload.Items {
			if !isVisibleMediaPlatform(item.Platform) {
				continue
			}
			seen[item.Platform] = struct{}{}
		}
	}

	platforms := make([]interfaces.MediaUnlockPlatform, 0, len(seen))
	for _, platform := range preferred {
		if _, ok := seen[platform]; ok {
			platforms = append(platforms, platform)
			delete(seen, platform)
		}
	}
	leftovers := make([]interfaces.MediaUnlockPlatform, 0, len(seen))
	for platform := range seen {
		leftovers = append(leftovers, platform)
	}
	sort.Slice(leftovers, func(i, j int) bool {
		return string(leftovers[i]) < string(leftovers[j])
	})
	platforms = append(platforms, leftovers...)
	if len(platforms) == 0 {
		return preferred[:0]
	}
	return platforms
}

func mediaPlatformLabel(platform interfaces.MediaUnlockPlatform) string {
	switch platform {
	case interfaces.MediaUnlockPlatformChatGPT:
		return "ChatGPT"
	case interfaces.MediaUnlockPlatformClaude:
		return "Claude"
	case interfaces.MediaUnlockPlatformGemini:
		return "Gemini"
	case interfaces.MediaUnlockPlatformYouTubePremium:
		return "YouTube"
	case interfaces.MediaUnlockPlatformNetflix:
		return "Netflix"
	case interfaces.MediaUnlockPlatformHulu:
		return "Hulu"
	case interfaces.MediaUnlockPlatformHuluJP:
		return "Hulu JP"
	case interfaces.MediaUnlockPlatformPrimeVideo:
		return "Prime Video"
	case interfaces.MediaUnlockPlatformHBOMax:
		return "HBO Max"
	case interfaces.MediaUnlockPlatformBilibiliHMT:
		return "Bilibili HMT"
	case interfaces.MediaUnlockPlatformBilibiliTW:
		return "Bilibili TW"
	case interfaces.MediaUnlockPlatformAbema:
		return "Abema"
	case interfaces.MediaUnlockPlatformTikTok:
		return "TikTok"
	case interfaces.MediaUnlockPlatformSpotify:
		return "Spotify"
	case interfaces.MediaUnlockPlatformSteam:
		return "Steam"
	default:
		return string(platform)
	}
}

func isVisibleMediaPlatform(platform interfaces.MediaUnlockPlatform) bool {
	switch platform {
	case
		"dazn",
		"instagram_music",
		interfaces.MediaUnlockPlatformHuluJP,
		interfaces.MediaUnlockPlatformSpotify,
		interfaces.MediaUnlockPlatformSteam:
		return false
	default:
		return true
	}
}

func preferredMediaDisplay(item interfaces.MediaUnlockPlatformResult) string {
	if strings.TrimSpace(item.Display) != "" {
		return item.Display
	}
	if strings.TrimSpace(item.Region) != "" {
		return string(item.Status) + " (" + item.Region + ")"
	}
	if strings.TrimSpace(item.Error) != "" {
		return string(item.Status) + " (" + item.Error + ")"
	}
	return string(item.Status)
}

func joinErrors(values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		duplicate := false
		for _, existing := range parts {
			if existing == trimmed {
				duplicate = true
				break
			}
		}
		if !duplicate {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " | ")
}

func preferredGeoOrganization(info interfaces.GeoIPInfo) string {
	org := strings.TrimSpace(info.ASOrganization)
	isp := strings.TrimSpace(info.ISP)
	if org == "" {
		return isp
	}
	if isp == "" {
		return org
	}
	if isGenericGeoOrganization(org) && !strings.EqualFold(org, isp) {
		return isp
	}
	return org
}

func isGenericGeoOrganization(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "private customer" ||
		lower == "private customers" ||
		lower == "customer" ||
		strings.Contains(lower, "private customer")
}

func hasAnyMatrix(entries []interfaces.MatrixEntry, targets ...interfaces.MatrixType) bool {
	for _, entry := range entries {
		for _, target := range targets {
			if entry.Type == target {
				return true
			}
		}
	}
	return false
}

func extractUint64Matrix(matrices []interfaces.MatrixResult, target interfaces.MatrixType) (uint64, bool) {
	for _, matrix := range matrices {
		if matrix.Type != target {
			continue
		}

		payload, ok := decodePayload[struct {
			Value uint64 `json:"value"`
		}](matrix.Payload)
		if ok {
			return payload.Value, true
		}
	}
	return 0, false
}

func extractUint64SliceMatrix(matrices []interfaces.MatrixResult, target interfaces.MatrixType) ([]uint64, bool) {
	for _, matrix := range matrices {
		if matrix.Type != target {
			continue
		}

		payload, ok := decodePayload[struct {
			Values []uint64 `json:"values"`
		}](matrix.Payload)
		if ok {
			return append([]uint64{}, payload.Values...), true
		}
	}
	return nil, false
}

func extractGeoMatrix(matrices []interfaces.MatrixResult, target interfaces.MatrixType) (interfaces.GeoIPInfo, bool) {
	for _, matrix := range matrices {
		if matrix.Type != target {
			continue
		}

		payload, ok := decodePayload[interfaces.GeoIPInfo](matrix.Payload)
		if ok {
			return payload, true
		}
	}
	return interfaces.GeoIPInfo{}, false
}

func extractMediaUnlockMatrix(matrices []interfaces.MatrixResult, target interfaces.MatrixType) (interfaces.MediaUnlockResult, bool) {
	for _, matrix := range matrices {
		if matrix.Type != target {
			continue
		}

		payload, ok := decodePayload[interfaces.MediaUnlockResult](matrix.Payload)
		if ok {
			return payload, true
		}
	}
	return interfaces.MediaUnlockResult{}, false
}

func decodePayload[T any](payload any) (T, bool) {
	var value T
	raw, err := json.Marshal(payload)
	if err != nil {
		return value, false
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, false
	}
	return value, true
}

func newerTimestamp(left string, right string) bool {
	leftTime, leftOK := parseTimestamp(left)
	rightTime, rightOK := parseTimestamp(right)
	if leftOK && rightOK {
		if leftTime.Equal(rightTime) {
			return left > right
		}
		return leftTime.After(rightTime)
	}
	if leftOK {
		return true
	}
	if rightOK {
		return false
	}
	return left > right
}

func parseTimestamp(raw string) (time.Time, bool) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}
