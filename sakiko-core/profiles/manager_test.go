package profiles

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/storage"
	"sakiko.local/sakiko-core/vendors/mihomo"
)

func TestManagerImportAndRefreshValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	content := `
proxies:
  - name: test
    type: vmess
    server: 1.1.1.1
    port: 443
`
	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:    "demo",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if profile.ID == "" {
		t.Fatalf("expected profile ID")
	}
	if len(profile.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(profile.Nodes))
	}

	profileFile := filepath.Join(dir, "profiles", profile.ID+".yaml")
	rawProfileFile, err := os.ReadFile(profileFile)
	if err != nil {
		t.Fatalf("expected generated profile file: %v", err)
	}
	if !strings.Contains(string(rawProfileFile), "proxies:") {
		t.Fatalf("expected generated yaml file to contain proxies, got %s", string(rawProfileFile))
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected generated profiles index: %v", err)
	}
	if strings.Contains(string(rawIndex), "nodes:") || strings.Contains(string(rawIndex), "payload:") {
		t.Fatalf("expected metadata-only index, got %s", string(rawIndex))
	}

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("expected list size 1, got %d", len(list))
	}

	if _, err := m.Refresh(profile.ID); err == nil {
		t.Fatalf("expected refresh error for empty source")
	}

	reloaded, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("reloaded NewManager() error = %v", err)
	}

	reloadedProfile, ok := reloaded.Get(profile.ID)
	if !ok {
		t.Fatalf("expected reloaded profile to exist")
	}
	if len(reloadedProfile.Nodes) != 1 {
		t.Fatalf("expected reloaded profile to have 1 node, got %d", len(reloadedProfile.Nodes))
	}
}

func TestManagerImportBase64Subscription(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	rawSubscription := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#demo"
	content := base64.StdEncoding.EncodeToString([]byte(rawSubscription))

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:    "base64-sub",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if len(profile.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(profile.Nodes))
	}

	if profile.Nodes[0].Name != "demo" {
		t.Fatalf("expected node name demo, got %s", profile.Nodes[0].Name)
	}

	payload := profile.Nodes[0].Payload
	for _, want := range []string{"type: ss", "server: 1.2.3.4", "port: \"8388\"", "cipher: aes-256-gcm"} {
		if !strings.Contains(payload, want) {
			t.Fatalf("expected payload to contain %q, got %s", want, payload)
		}
	}

	vendor := (&mihomo.Vendor{}).Build(profile.Nodes[0])
	if vendor.Status() != interfaces.VStatusOperational {
		t.Fatalf("expected mihomo vendor to recognize generated node payload")
	}

	profileFile := filepath.Join(dir, "profiles", profile.ID+".yaml")
	rawProfileFile, err := os.ReadFile(profileFile)
	if err != nil {
		t.Fatalf("expected generated profile file: %v", err)
	}
	if !strings.Contains(string(rawProfileFile), "proxies:") {
		t.Fatalf("expected generated yaml file to contain proxies, got %s", string(rawProfileFile))
	}
}

func TestManagerDeleteRemovesProfileFromIndexAndContentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name: "delete-me",
		Content: `
proxies:
  - name: delete-me
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
`,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	profileFile := filepath.Join(dir, "profiles", profile.ID+".yaml")
	if _, err := os.Stat(profileFile); err != nil {
		t.Fatalf("expected generated profile file: %v", err)
	}

	if err := m.Delete(profile.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if got := m.List(); len(got) != 0 {
		t.Fatalf("expected no profiles after delete, got %d", len(got))
	}
	if _, ok := m.Get(profile.ID); ok {
		t.Fatalf("expected deleted profile to be missing")
	}
	if _, err := os.Stat(profileFile); !os.IsNotExist(err) {
		t.Fatalf("expected profile file removed, stat error = %v", err)
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(rawIndex), profile.ID) {
		t.Fatalf("expected index to remove profile id, got %s", string(rawIndex))
	}
}

func TestManagerImportRollsBackWhenIndexPersistFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Point the profile index store at a directory so the write fails.
	m.store = storage.NewProfileStore(dir)

	_, err = m.Import(interfaces.ProfileImportRequest{
		Name: "rollback-import",
		Content: `
proxies:
  - name: rollback
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
`,
	})
	if err == nil {
		t.Fatalf("expected import to fail when index persistence fails")
	}

	if got := m.List(); len(got) != 0 {
		t.Fatalf("expected import rollback to leave no profiles, got %d", len(got))
	}

	entries, readErr := os.ReadDir(filepath.Join(dir, "profiles"))
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("ReadDir() error = %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected rollback to remove generated profile files, got %d entries", len(entries))
	}
}

func TestManagerRefreshRollsBackProfileFileWhenIndexPersistFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: refreshed-node
    type: ss
    server: 8.8.8.8
    port: 8443
    cipher: aes-128-gcm
    password: refreshed
`))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "refresh-rollback",
		Source: server.URL,
		Content: `
proxies:
  - name: original-node
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: original
`,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	m.store = storage.NewProfileStore(dir)

	if _, err := m.Refresh(profile.ID); err == nil {
		t.Fatalf("expected refresh to fail when index persistence fails")
	}

	current, ok := m.Get(profile.ID)
	if !ok {
		t.Fatalf("expected profile to remain present after rollback")
	}
	if len(current.Nodes) != 1 || !strings.Contains(current.Nodes[0].Payload, "1.1.1.1") {
		t.Fatalf("expected in-memory profile to keep original nodes, got %+v", current.Nodes)
	}

	profileFile := filepath.Join(dir, "profiles", profile.ID+".yaml")
	raw, err := os.ReadFile(profileFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(raw), "1.1.1.1") {
		t.Fatalf("expected profile file rollback to restore original content, got %s", string(raw))
	}
}

func TestManagerRecoversProfilesFromContentDirectoryWhenIndexMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")
	profilesDir := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const profileID = "1a36627c6a60bd0c021383b9"
	content := `
proxies:
  - name: hk-demo
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
`

	profilePath := filepath.Join(profilesDir, profileID+".yaml")
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	modTime := time.Date(2026, 3, 29, 22, 35, 48, 0, time.UTC)
	if err := os.Chtimes(profilePath, modTime, modTime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("expected recovered list size 1, got %d", len(list))
	}

	profile := list[0]
	if profile.ID != profileID {
		t.Fatalf("expected recovered profile id %s, got %s", profileID, profile.ID)
	}
	if profile.Name != "Recovered Profile 1a36627c" {
		t.Fatalf("expected recovered profile name, got %s", profile.Name)
	}
	if profile.UpdatedAt != modTime.Format(time.RFC3339) {
		t.Fatalf("expected updatedAt %s, got %s", modTime.Format(time.RFC3339), profile.UpdatedAt)
	}
	if len(profile.Nodes) != 1 {
		t.Fatalf("expected 1 recovered node, got %d", len(profile.Nodes))
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected recovered profiles.yaml to be created: %v", err)
	}
	if !strings.Contains(string(rawIndex), profileID) {
		t.Fatalf("expected recovered index to contain profile id, got %s", string(rawIndex))
	}
	if strings.Contains(string(rawIndex), "nodes:") || strings.Contains(string(rawIndex), "payload:") {
		t.Fatalf("expected recovered index to stay metadata-only, got %s", string(rawIndex))
	}
}

func TestManagerRecoversProfilesWhenIndexIsMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")
	profilesDir := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(path, []byte("profiles: ["), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	const validProfileID = "good-profile"
	if err := os.WriteFile(filepath.Join(profilesDir, validProfileID+".yaml"), []byte(`
proxies:
  - name: sg-demo
    type: ss
    server: 2.2.2.2
    port: 8443
    cipher: aes-128-gcm
    password: demo
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(profilesDir, "bad-profile.yaml"), []byte("proxies: ["), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("expected only valid recovered profile, got %d", len(list))
	}
	if list[0].ID != validProfileID {
		t.Fatalf("expected recovered profile id %s, got %s", validProfileID, list[0].ID)
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(rawIndex), "bad-profile") {
		t.Fatalf("expected malformed profile to be skipped, got %s", string(rawIndex))
	}
}

func TestInferProfileNameFromResponsePrefersContentDisposition(t *testing.T) {
	header := http.Header{}
	header.Set("Content-Disposition", "attachment;filename*=UTF-8''CTC")

	got := inferProfileNameFromResponse(header, "https://example.com/api/v1/client/subscribe?token=demo")
	if got != "CTC" {
		t.Fatalf("expected CTC, got %q", got)
	}
}

func TestUniqueProfileNameLockedAppendsChineseSuffix(t *testing.T) {
	m := &Manager{
		profiles: map[string]interfaces.Profile{
			"1": {ID: "1", Name: "CTC"},
			"2": {ID: "2", Name: "CTC（1）"},
		},
	}

	got := m.uniqueProfileNameLocked("CTC", "")
	if got != "CTC（2）" {
		t.Fatalf("expected CTC（2）, got %q", got)
	}
}
