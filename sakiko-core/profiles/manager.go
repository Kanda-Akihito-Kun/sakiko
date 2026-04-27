package profiles

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/netx"
	"sakiko.local/sakiko-core/storage"

	"go.uber.org/zap"
)

type Config struct {
	StorePath      string
	FetchTimeout   time.Duration
	MaxSourceBytes int64
}

type Manager struct {
	store *storage.ProfileStore
	files *storage.ProfileContentStore

	lock     sync.RWMutex
	profiles map[string]interfaces.Profile
	order    []string

	now func() time.Time

	fetchTimeout    time.Duration
	maxProfileBytes int64
}

type fetchedProfileSource struct {
	Content    string
	Name       string
	Attributes map[string]any
}

const (
	defaultSubscriptionUserAgent = "clash-verge/v2.4.6"
	defaultMaxProfileSourceBytes = 20 * 1024 * 1024
)

func NewManager(cfg Config) (*Manager, error) {
	timeout := cfg.FetchTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	maxSourceBytes := cfg.MaxSourceBytes
	if maxSourceBytes <= 0 {
		maxSourceBytes = defaultMaxProfileSourceBytes
	}

	m := &Manager{
		store:           storage.NewProfileStore(cfg.StorePath),
		files:           storage.NewProfileContentStore(cfg.StorePath),
		profiles:        map[string]interfaces.Profile{},
		order:           []string{},
		now:             time.Now,
		fetchTimeout:    timeout,
		maxProfileBytes: maxSourceBytes,
	}

	loaded, recovered, err := m.bootstrapProfiles()
	if err != nil {
		return nil, err
	}
	for _, profile := range loaded {
		if strings.TrimSpace(profile.ID) == "" {
			continue
		}
		m.profiles[profile.ID] = profile
		m.order = append(m.order, profile.ID)
		if err := m.saveProfileFile(profile); err != nil {
			return nil, err
		}
	}
	profilesLogger().Info("profile manager initialized",
		zap.Int("loaded_profiles", len(loaded)),
		zap.Bool("recovered_profiles", recovered),
		zap.Duration("fetch_timeout", timeout),
		zap.Int64("max_source_bytes", maxSourceBytes),
		zap.String("store_path", cfg.StorePath),
	)
	if recovered {
		if err := m.persistLocked(); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) List() []interfaces.Profile {
	m.lock.RLock()
	defer m.lock.RUnlock()

	out := make([]interfaces.Profile, 0, len(m.order))
	for _, id := range m.order {
		profile, ok := m.profiles[id]
		if !ok {
			continue
		}
		out = append(out, profile)
	}
	return out
}

func (m *Manager) Get(profileID string) (interfaces.Profile, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	profile, ok := m.profiles[profileID]
	return profile, ok
}

func (m *Manager) Import(req interfaces.ProfileImportRequest) (interfaces.Profile, error) {
	if m == nil {
		profilesLogger().Warn("import rejected: profile manager not initialized")
		return interfaces.Profile{}, fmt.Errorf("profile manager not initialized")
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		return interfaces.Profile{}, fmt.Errorf("profile source is required")
	}

	profilesLogger().Info("fetching profile source for import",
		zap.String("source", sourceLabel(source)),
	)
	fetched, err := m.fetchSource(source)
	if err != nil {
		profilesLogger().Warn("fetch profile source for import failed",
			zap.String("source", sourceLabel(source)),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}
	content := fetched.Content

	nodes, err := ParseNodes(content)
	if err != nil {
		profilesLogger().Warn("parse imported profile failed",
			zap.String("source", sourceLabel(source)),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		if strings.TrimSpace(fetched.Name) != "" {
			name = fetched.Name
		} else {
			name = inferProfileName(source)
		}
	}

	profile := interfaces.Profile{
		ID:         randomID(),
		Name:       name,
		Source:     source,
		Nodes:      nodes,
		UpdatedAt:  m.now().UTC().Format(time.RFC3339),
		Attributes: mergeProfileAttributes(req.Attributes, fetched.Attributes),
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	profile.Name = m.uniqueProfileNameLocked(profile.Name, "")
	if err := m.saveProfileFile(profile); err != nil {
		return interfaces.Profile{}, err
	}

	prevOrder := append([]string{}, m.order...)
	m.profiles[profile.ID] = profile
	m.order = append(m.order, profile.ID)
	if err := m.persistLocked(); err != nil {
		delete(m.profiles, profile.ID)
		m.order = prevOrder
		if cleanupErr := m.deleteProfileFile(profile.ID); cleanupErr != nil {
			profilesLogger().Warn("rollback imported profile file failed",
				zap.String("profile_id", profile.ID),
				zap.Error(cleanupErr),
			)
			return interfaces.Profile{}, errors.Join(err, cleanupErr)
		}
		return interfaces.Profile{}, err
	}
	profilesLogger().Info("profile imported",
		zap.String("profile_id", profile.ID),
		zap.String("profile_name", profile.Name),
		zap.String("source", sourceLabel(profile.Source)),
		zap.Int("node_count", len(profile.Nodes)),
	)
	return profile, nil
}

func (m *Manager) Refresh(profileID string) (interfaces.Profile, error) {
	if m == nil {
		profilesLogger().Warn("refresh rejected: profile manager not initialized")
		return interfaces.Profile{}, fmt.Errorf("profile manager not initialized")
	}

	m.lock.RLock()
	profile, ok := m.profiles[profileID]
	m.lock.RUnlock()
	if !ok {
		profilesLogger().Debug("profile not found during refresh", zap.String("profile_id", profileID))
		return interfaces.Profile{}, fmt.Errorf("profile not found")
	}
	if strings.TrimSpace(profile.Source) == "" {
		profilesLogger().Warn("refresh rejected: profile source is empty", zap.String("profile_id", profileID))
		return interfaces.Profile{}, fmt.Errorf("profile source is empty")
	}

	profilesLogger().Info("refreshing profile",
		zap.String("profile_id", profileID),
		zap.String("source", sourceLabel(profile.Source)),
	)
	fetched, err := m.fetchSource(profile.Source)
	if err != nil {
		profilesLogger().Warn("refresh profile fetch failed",
			zap.String("profile_id", profileID),
			zap.String("source", sourceLabel(profile.Source)),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}
	nodes, err := ParseNodes(fetched.Content)
	if err != nil {
		profilesLogger().Warn("refresh profile parse failed",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return interfaces.Profile{}, err
	}

	profile.Nodes = applyNodeSelections(profile.Nodes, nodes)
	profile.UpdatedAt = m.now().UTC().Format(time.RFC3339)
	profile.Attributes = mergeProfileAttributes(profile.Attributes, fetched.Attributes)

	m.lock.Lock()
	defer m.lock.Unlock()

	previousProfile := m.profiles[profileID]
	if err := m.saveProfileFile(profile); err != nil {
		return interfaces.Profile{}, err
	}

	m.profiles[profileID] = profile
	if err := m.persistLocked(); err != nil {
		m.profiles[profileID] = previousProfile
		if restoreErr := m.saveProfileFile(previousProfile); restoreErr != nil {
			profilesLogger().Warn("restore profile file after refresh failure failed",
				zap.String("profile_id", profileID),
				zap.Error(restoreErr),
			)
			return interfaces.Profile{}, errors.Join(err, restoreErr)
		}
		return interfaces.Profile{}, err
	}
	profilesLogger().Info("profile refreshed",
		zap.String("profile_id", profile.ID),
		zap.String("profile_name", profile.Name),
		zap.Int("node_count", len(profile.Nodes)),
	)
	return profile, nil
}

func (m *Manager) SetNodeEnabled(profileID string, nodeIndex int, enabled bool) (interfaces.Profile, error) {
	if m == nil {
		profilesLogger().Warn("set node enabled rejected: profile manager not initialized")
		return interfaces.Profile{}, fmt.Errorf("profile manager not initialized")
	}

	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return interfaces.Profile{}, fmt.Errorf("profile ID is required")
	}
	if nodeIndex < 0 {
		return interfaces.Profile{}, fmt.Errorf("node index is required")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	profile, ok := m.profiles[profileID]
	if !ok {
		return interfaces.Profile{}, fmt.Errorf("profile not found")
	}
	if nodeIndex >= len(profile.Nodes) {
		return interfaces.Profile{}, fmt.Errorf("node index out of range")
	}
	if profile.Nodes[nodeIndex].Enabled == enabled {
		return profile, nil
	}

	previousProfile := profile
	profile.Nodes = cloneNodes(profile.Nodes)
	profile.Nodes[nodeIndex].Enabled = enabled
	profile.UpdatedAt = m.now().UTC().Format(time.RFC3339)
	m.profiles[profileID] = profile

	if err := m.persistLocked(); err != nil {
		m.profiles[profileID] = previousProfile
		return interfaces.Profile{}, err
	}

	profilesLogger().Info("profile node selection updated",
		zap.String("profile_id", profileID),
		zap.Int("node_index", nodeIndex),
		zap.String("node_name", profile.Nodes[nodeIndex].Name),
		zap.Bool("enabled", enabled),
	)
	return profile, nil
}

func (m *Manager) MoveNode(profileID string, nodeIndex int, targetIndex int) (interfaces.Profile, error) {
	if m == nil {
		profilesLogger().Warn("move node rejected: profile manager not initialized")
		return interfaces.Profile{}, fmt.Errorf("profile manager not initialized")
	}
	if strings.TrimSpace(profileID) == "" {
		return interfaces.Profile{}, fmt.Errorf("profile ID is required")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	profile, ok := m.profiles[profileID]
	if !ok {
		return interfaces.Profile{}, fmt.Errorf("profile not found")
	}
	if nodeIndex < 0 || nodeIndex >= len(profile.Nodes) {
		return interfaces.Profile{}, fmt.Errorf("node index out of range")
	}
	if targetIndex < 0 || targetIndex >= len(profile.Nodes) {
		return interfaces.Profile{}, fmt.Errorf("target index out of range")
	}
	if nodeIndex == targetIndex {
		return profile, nil
	}

	profile.Nodes = cloneNodes(profile.Nodes)
	moved := profile.Nodes[nodeIndex]
	nextNodes := make([]interfaces.Node, 0, len(profile.Nodes))
	for i, node := range profile.Nodes {
		if i == nodeIndex {
			continue
		}
		nextNodes = append(nextNodes, node)
	}

	if targetIndex >= len(nextNodes) {
		nextNodes = append(nextNodes, moved)
	} else {
		nextNodes = append(nextNodes[:targetIndex], append([]interfaces.Node{moved}, nextNodes[targetIndex:]...)...)
	}

	profile.Nodes = normalizeNodeOrder(nextNodes)
	previousProfile := m.profiles[profileID]
	if err := m.saveProfileFile(profile); err != nil {
		return interfaces.Profile{}, err
	}
	m.profiles[profileID] = profile
	if err := m.persistLocked(); err != nil {
		m.profiles[profileID] = previousProfile
		if restoreErr := m.saveProfileFile(previousProfile); restoreErr != nil {
			profilesLogger().Warn("restore profile file after node move failure failed",
				zap.String("profile_id", profileID),
				zap.Error(restoreErr),
			)
			return interfaces.Profile{}, errors.Join(err, restoreErr)
		}
		return interfaces.Profile{}, err
	}

	profilesLogger().Info("profile node moved",
		zap.String("profile_id", profileID),
		zap.Int("node_index", nodeIndex),
		zap.Int("target_index", targetIndex),
		zap.String("node_name", moved.Name),
	)
	return profile, nil
}

func (m *Manager) Delete(profileID string) error {
	if m == nil {
		profilesLogger().Warn("delete rejected: profile manager not initialized")
		return fmt.Errorf("profile manager not initialized")
	}

	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	profile, ok := m.profiles[profileID]
	if !ok {
		profilesLogger().Debug("profile not found during delete", zap.String("profile_id", profileID))
		return fmt.Errorf("profile not found")
	}

	prevOrder := append([]string{}, m.order...)
	delete(m.profiles, profileID)
	nextOrder := make([]string, 0, len(m.order))
	for _, id := range m.order {
		if id != profileID {
			nextOrder = append(nextOrder, id)
		}
	}
	m.order = nextOrder

	if err := m.persistLocked(); err != nil {
		m.profiles[profileID] = profile
		m.order = prevOrder
		return err
	}

	if err := m.deleteProfileFile(profileID); err != nil {
		m.profiles[profileID] = profile
		m.order = prevOrder
		if restoreErr := m.persistLocked(); restoreErr != nil {
			profilesLogger().Warn("restore profile index after delete failure failed",
				zap.String("profile_id", profileID),
				zap.Error(restoreErr),
			)
			return errors.Join(err, restoreErr)
		}
		profilesLogger().Warn("delete profile file failed",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return err
	}

	profilesLogger().Info("profile deleted",
		zap.String("profile_id", profileID),
		zap.String("profile_name", profile.Name),
	)
	return nil
}

func (m *Manager) persistLocked() error {
	profiles := make([]interfaces.Profile, 0, len(m.order))
	for _, id := range m.order {
		profile, ok := m.profiles[id]
		if !ok {
			continue
		}
		profiles = append(profiles, profile)
	}
	return m.store.Save(profiles)
}

func (m *Manager) fetchSource(source string) (fetchedProfileSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.fetchTimeout)
	defer cancel()

	profilesLogger().Debug("fetching profile source",
		zap.String("source", sourceLabel(source)),
		zap.Duration("timeout", m.fetchTimeout),
	)
	resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    source,
		Headers: map[string]string{
			"User-Agent": defaultSubscriptionUserAgent,
		},
	})
	if err != nil {
		profilesLogger().Warn("fetch profile source request failed",
			zap.String("source", sourceLabel(source)),
			zap.Error(err),
		)
		return fetchedProfileSource{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		profilesLogger().Warn("fetch profile source returned unexpected status",
			zap.String("source", sourceLabel(source)),
			zap.Int("status_code", resp.StatusCode),
		)
		return fetchedProfileSource{}, fmt.Errorf("profile source returned status %d", resp.StatusCode)
	}

	limit := m.maxProfileBytes
	if limit <= 0 {
		limit = defaultMaxProfileSourceBytes
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		profilesLogger().Warn("read profile source body failed",
			zap.String("source", sourceLabel(source)),
			zap.Error(err),
		)
		return fetchedProfileSource{}, err
	}
	if int64(len(body)) > limit {
		profilesLogger().Warn("profile source body exceeded size limit",
			zap.String("source", sourceLabel(source)),
			zap.Int64("limit_bytes", limit),
			zap.Int("received_bytes", len(body)),
		)
		return fetchedProfileSource{}, fmt.Errorf("profile source exceeds %d bytes", limit)
	}
	return fetchedProfileSource{
		Content:    string(body),
		Name:       inferProfileNameFromResponse(resp.Header, source),
		Attributes: inferProfileAttributesFromResponse(resp.Header),
	}, nil
}

func inferProfileAttributesFromResponse(header http.Header) map[string]any {
	info := parseSubscriptionUserinfo(header.Get("Subscription-Userinfo"))
	if len(info) == 0 {
		return nil
	}
	return map[string]any{
		"subscriptionUserinfo": info,
	}
}

func parseSubscriptionUserinfo(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	values := map[string]uint64{}
	for _, part := range strings.Split(raw, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil {
			continue
		}
		values[strings.ToLower(strings.TrimSpace(key))] = parsed
	}

	if len(values) == 0 {
		return nil
	}

	info := map[string]any{}
	if upload, ok := values["upload"]; ok {
		info["upload"] = upload
	}
	if download, ok := values["download"]; ok {
		info["download"] = download
	}
	if total, ok := values["total"]; ok {
		info["total"] = total
		used := values["upload"] + values["download"]
		if total > used {
			info["remaining"] = total - used
		} else {
			info["remaining"] = uint64(0)
		}
	}
	if expire, ok := values["expire"]; ok {
		info["expire"] = expire
		if expire > 0 {
			info["expiresAt"] = time.Unix(int64(expire), 0).UTC().Format(time.RFC3339)
		}
	}
	if len(info) == 0 {
		return nil
	}
	return info
}

func mergeProfileAttributes(existing any, incoming map[string]any) any {
	if len(incoming) == 0 {
		return existing
	}

	merged := coerceAttributesMap(existing)
	if merged == nil {
		merged = map[string]any{}
	}
	for key, value := range incoming {
		merged[key] = value
	}
	return merged
}

func coerceAttributesMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	if typed, ok := value.(map[string]any); ok {
		out := make(map[string]any, len(typed))
		for key, field := range typed {
			out[key] = field
		}
		return out
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func inferProfileName(source string) string {
	source = strings.TrimSpace(source)
	if name := inferProfileNameFromSource(source); name != "" {
		return name
	}
	return "Imported Profile"
}

func applyNodeSelections(reference []interfaces.Node, nodes []interfaces.Node) []interfaces.Node {
	out := normalizeNodeOrder(cloneNodes(nodes))
	for i := range out {
		out[i].Enabled = true
	}
	if len(reference) == 0 || len(out) == 0 {
		return out
	}

	type nodeState struct {
		enabled bool
		order   int
	}

	statesByName := make(map[string][]nodeState, len(reference))
	for _, node := range normalizeNodeOrder(cloneNodes(reference)) {
		key := strings.TrimSpace(node.Name)
		if key == "" {
			continue
		}
		statesByName[key] = append(statesByName[key], nodeState{
			enabled: node.Enabled,
			order:   node.Order,
		})
	}

	nextOrder := len(reference)
	for i := range out {
		key := strings.TrimSpace(out[i].Name)
		states := statesByName[key]
		if len(states) == 0 {
			out[i].Order = nextOrder
			nextOrder++
			continue
		}
		out[i].Enabled = states[0].enabled
		out[i].Order = states[0].order
		statesByName[key] = states[1:]
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Name < out[j].Name
	})

	return normalizeNodeOrder(out)
}

func cloneNodes(nodes []interfaces.Node) []interfaces.Node {
	if len(nodes) == 0 {
		return []interfaces.Node{}
	}
	out := make([]interfaces.Node, len(nodes))
	copy(out, nodes)
	return out
}

func normalizeNodeOrder(nodes []interfaces.Node) []interfaces.Node {
	if len(nodes) == 0 {
		return []interfaces.Node{}
	}

	out := cloneNodes(nodes)
	for i := range out {
		out[i].Order = i
	}
	return out
}

func randomID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (m *Manager) saveProfileFile(profile interfaces.Profile) error {
	if m == nil || m.files == nil {
		return nil
	}

	content, err := ComposeContent(profile.Nodes)
	if err != nil {
		return err
	}

	return m.files.Save(profile.ID, content)
}

func (m *Manager) deleteProfileFile(profileID string) error {
	if m == nil || m.files == nil {
		return nil
	}
	return m.files.Delete(profileID)
}

func (m *Manager) bootstrapProfiles() ([]interfaces.Profile, bool, error) {
	if m == nil || m.store == nil {
		return nil, false, fmt.Errorf("profile manager not initialized")
	}

	loaded, err := m.store.Load()
	if err == nil && len(loaded) > 0 {
		hydrated, hydrateErr := m.hydrateStoredProfiles(loaded)
		if hydrateErr == nil && len(hydrated) > 0 {
			profilesLogger().Info("loaded profiles from store", zap.Int("profile_count", len(hydrated)))
			return hydrated, false, nil
		}
		if hydrateErr != nil {
			profilesLogger().Warn("hydrate stored profiles failed", zap.Error(hydrateErr))
		}
	}

	recovered, recoverErr := m.recoverProfilesFromFiles()
	if recoverErr == nil && len(recovered) > 0 {
		profilesLogger().Warn("recovered profiles from profile content files",
			zap.Int("profile_count", len(recovered)),
		)
		return recovered, true, nil
	}

	if err != nil {
		return nil, false, err
	}
	if recoverErr != nil {
		return nil, false, recoverErr
	}
	return []interfaces.Profile{}, false, nil
}

func (m *Manager) hydrateStoredProfiles(stored []interfaces.Profile) ([]interfaces.Profile, error) {
	if m == nil || m.files == nil {
		return stored, nil
	}

	hydrated := make([]interfaces.Profile, 0, len(stored))

	for _, profile := range stored {
		profile.ID = strings.TrimSpace(profile.ID)
		if profile.ID == "" {
			continue
		}

		contentFile, err := m.files.Load(profile.ID)
		if err != nil {
			profilesLogger().Warn("profile content file missing for indexed profile",
				zap.String("profile_id", profile.ID),
				zap.Error(err),
			)
			continue
		}

		nodes, err := ParseNodes(contentFile.Content)
		if err != nil || len(nodes) == 0 {
			profilesLogger().Warn("profile content file invalid for indexed profile",
				zap.String("profile_id", profile.ID),
				zap.Error(err),
			)
			continue
		}

		profile.Nodes = applyNodeSelections(profile.Nodes, nodes)
		if strings.TrimSpace(profile.UpdatedAt) == "" {
			profile.UpdatedAt = contentFile.UpdatedAt.UTC().Format(time.RFC3339)
		}
		hydrated = append(hydrated, profile)
	}

	if len(hydrated) == 0 && len(stored) > 0 {
		return nil, fmt.Errorf("no valid profiles could be hydrated from store")
	}

	return hydrated, nil
}

func (m *Manager) recoverProfilesFromFiles() ([]interfaces.Profile, error) {
	if m == nil || m.files == nil {
		return []interfaces.Profile{}, nil
	}

	files, err := m.files.LoadAll()
	if err != nil {
		return nil, err
	}

	recovered := make([]interfaces.Profile, 0, len(files))
	for _, file := range files {
		profileID := strings.TrimSpace(file.ID)
		if profileID == "" {
			continue
		}

		nodes, err := ParseNodes(file.Content)
		if err != nil || len(nodes) == 0 {
			profilesLogger().Warn("skip invalid recovered profile file",
				zap.String("profile_id", profileID),
				zap.Error(err),
			)
			continue
		}

		recovered = append(recovered, interfaces.Profile{
			ID:        profileID,
			Name:      recoveredProfileName(profileID),
			Source:    "",
			Nodes:     nodes,
			UpdatedAt: file.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return recovered, nil
}

func recoveredProfileName(profileID string) string {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return "Recovered Profile"
	}
	if len(profileID) > 8 {
		profileID = profileID[:8]
	}
	return fmt.Sprintf("Recovered Profile %s", profileID)
}

func profilesLogger() *zap.Logger {
	return logx.Named("core.profiles")
}

func sourceLabel(source string) string {
	raw := strings.TrimSpace(source)
	if raw == "" {
		return "empty-source"
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "invalid-url"
	}
	if parsed.Host != "" {
		return parsed.Host
	}
	if parsed.Scheme != "" {
		return parsed.Scheme
	}
	return "empty-source"
}

func inferProfileNameFromResponse(header http.Header, source string) string {
	if header == nil {
		return inferProfileNameFromSource(source)
	}

	for _, key := range []string{"Profile-Title", "profile-title"} {
		if value := strings.TrimSpace(header.Get(key)); value != "" {
			return value
		}
	}

	if value := strings.TrimSpace(header.Get("Content-Disposition")); value != "" {
		if filename := parseFilenameFromContentDisposition(value); filename != "" {
			return filename
		}
	}

	return inferProfileNameFromSource(source)
}

func parseFilenameFromContentDisposition(value string) string {
	_, params, err := mime.ParseMediaType(value)
	if err == nil {
		if filename := strings.TrimSpace(params["filename"]); filename != "" {
			return filename
		}
	}

	const marker = "filename*="
	lower := strings.ToLower(value)
	index := strings.Index(lower, marker)
	if index < 0 {
		return ""
	}
	raw := strings.TrimSpace(value[index+len(marker):])
	raw = strings.Trim(raw, "\"")
	if parts := strings.SplitN(raw, "''", 2); len(parts) == 2 {
		if decoded, err := url.QueryUnescape(parts[1]); err == nil {
			return strings.TrimSpace(decoded)
		}
	}
	if decoded, err := url.QueryUnescape(raw); err == nil {
		return strings.TrimSpace(decoded)
	}
	return strings.TrimSpace(raw)
}

func inferProfileNameFromSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}

	parsed, err := url.Parse(source)
	if err != nil {
		return ""
	}

	base := strings.TrimSpace(path.Base(parsed.Path))
	switch base {
	case "", ".", "/":
		return ""
	}

	if decoded, err := url.PathUnescape(base); err == nil {
		base = decoded
	}
	return strings.TrimSpace(base)
}

func (m *Manager) uniqueProfileNameLocked(name string, excludeID string) string {
	base := strings.TrimSpace(name)
	if base == "" {
		base = "Imported Profile"
	}

	if !m.profileNameExistsLocked(base, excludeID) {
		return base
	}

	for i := 1; ; i++ {
		candidate := base + "（" + strconv.Itoa(i) + "）"
		if !m.profileNameExistsLocked(candidate, excludeID) {
			return candidate
		}
	}
}

func (m *Manager) profileNameExistsLocked(name string, excludeID string) bool {
	for id, profile := range m.profiles {
		if id == excludeID {
			continue
		}
		if strings.TrimSpace(profile.Name) == name {
			return true
		}
	}
	return false
}
