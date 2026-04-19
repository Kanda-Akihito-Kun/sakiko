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
	"sakiko.local/sakiko-core/internal/testkit"
	"sakiko.local/sakiko-core/storage"
	"sakiko.local/sakiko-core/vendors/mihomo"
)

func TestManagerImportAndRefreshValidation(t *testing.T) {
	dir, path := testkit.TempProfilesStore(t)

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if _, err := m.Import(interfaces.ProfileImportRequest{Name: "missing-source"}); err == nil {
		t.Fatalf("expected import error for empty source")
	}

	responseBody := `
proxies:
  - name: test
    type: vmess
    server: 1.1.1.1
    port: 443
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "demo",
		Source: server.URL,
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
	if profile.Nodes[0].Protocol != "vmess" || profile.Nodes[0].Server != "1.1.1.1" || profile.Nodes[0].Port != "443" {
		t.Fatalf("expected imported node metadata, got %+v", profile.Nodes[0])
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
	if strings.Contains(string(rawIndex), "payload:") {
		t.Fatalf("expected index to omit payload content, got %s", string(rawIndex))
	}

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("expected list size 1, got %d", len(list))
	}

	responseBody = `
proxies:
  - name: refreshed
    type: vmess
    server: 2.2.2.2
    port: 8443
`
	refreshed, err := m.Refresh(profile.ID)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if len(refreshed.Nodes) != 1 || !strings.Contains(refreshed.Nodes[0].Payload, "2.2.2.2") {
		t.Fatalf("expected refreshed profile nodes, got %+v", refreshed.Nodes)
	}
	if refreshed.Nodes[0].Server != "2.2.2.2" || refreshed.Nodes[0].Port != "8443" {
		t.Fatalf("expected refreshed node metadata, got %+v", refreshed.Nodes[0])
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
	dir, path := testkit.TempProfilesStore(t)

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	rawSubscription := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#demo"
	content := base64.StdEncoding.EncodeToString([]byte(rawSubscription))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "base64-sub",
		Source: server.URL,
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
	if profile.Nodes[0].Protocol != "ss" || profile.Nodes[0].Server != "1.2.3.4" || profile.Nodes[0].Port != "8388" {
		t.Fatalf("expected imported share-link metadata, got %+v", profile.Nodes[0])
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

func TestManagerImportAnyTLSSubscription(t *testing.T) {
	_, path := testkit.TempProfilesStore(t)

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: anytls-demo
    type: anytls
    server: 1.2.3.4
    port: 443
    password: demo
    sni: example.com
    udp: true
`))
	}))
	defer server.Close()

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "anytls-sub",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if len(profile.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(profile.Nodes))
	}
	if profile.Nodes[0].Protocol != "anytls" || profile.Nodes[0].Server != "1.2.3.4" || profile.Nodes[0].Port != "443" {
		t.Fatalf("expected imported AnyTLS metadata, got %+v", profile.Nodes[0])
	}

	vendor := (&mihomo.Vendor{}).Build(profile.Nodes[0])
	if vendor.Status() != interfaces.VStatusOperational {
		t.Fatalf("expected mihomo vendor to recognize AnyTLS payload")
	}
	if vendor.ProxyInfo().Type != interfaces.ProxyAnyTLS {
		t.Fatalf("expected AnyTLS proxy type, got %+v", vendor.ProxyInfo())
	}
}

func TestManagerDeleteRemovesProfileFromIndexAndContentDirectory(t *testing.T) {
	dir, path := testkit.TempProfilesStore(t)

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: delete-me
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
`))
	}))
	defer server.Close()

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "delete-me",
		Source: server.URL,
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
	dir, path := testkit.TempProfilesStore(t)

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Point the profile index store at a directory so the write fails.
	m.store = storage.NewProfileStore(dir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: rollback
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
`))
	}))
	defer server.Close()

	_, err = m.Import(interfaces.ProfileImportRequest{
		Name:   "rollback-import",
		Source: server.URL,
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
	dir, path := testkit.TempProfilesStore(t)

	responseBody := `
proxies:
  - name: original-node
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: original
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "refresh-rollback",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	m.store = storage.NewProfileStore(dir)
	responseBody = `
proxies:
  - name: refreshed-node
    type: ss
    server: 8.8.8.8
    port: 8443
    cipher: aes-128-gcm
    password: refreshed
`

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
	if current.Nodes[0].Server != "1.1.1.1" {
		t.Fatalf("expected in-memory metadata to keep original server, got %+v", current.Nodes[0])
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
	dir, path := testkit.TempProfilesStore(t)
	profilesDir := filepath.Join(dir, "profiles")
	testkit.MustMkdirAll(t, profilesDir)

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
	testkit.MustWriteString(t, profilePath, content)

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
	if profile.Nodes[0].Protocol != "ss" || profile.Nodes[0].Server != "1.1.1.1" || profile.Nodes[0].Port != "443" {
		t.Fatalf("expected recovered node metadata, got %+v", profile.Nodes[0])
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected recovered profiles.yaml to be created: %v", err)
	}
	if !strings.Contains(string(rawIndex), profileID) {
		t.Fatalf("expected recovered index to contain profile id, got %s", string(rawIndex))
	}
	if strings.Contains(string(rawIndex), "payload:") {
		t.Fatalf("expected recovered index to omit payload content, got %s", string(rawIndex))
	}
}

func TestManagerSetNodeEnabledPersistsSelectionAndSurvivesReload(t *testing.T) {
	_, path := testkit.TempProfilesStore(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: hk-01
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
  - name: us-02
    type: ss
    server: 2.2.2.2
    port: 8443
    cipher: aes-128-gcm
    password: demo
`))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "selection-demo",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !profile.Nodes[0].Enabled || !profile.Nodes[1].Enabled {
		t.Fatalf("expected imported nodes enabled by default, got %+v", profile.Nodes)
	}

	updated, err := m.SetNodeEnabled(profile.ID, 1, false)
	if err != nil {
		t.Fatalf("SetNodeEnabled() error = %v", err)
	}
	if !updated.Nodes[0].Enabled || updated.Nodes[1].Enabled {
		t.Fatalf("expected only second node disabled, got %+v", updated.Nodes)
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(rawIndex), "enabled: false") {
		t.Fatalf("expected index to persist disabled selection, got %s", string(rawIndex))
	}
	if strings.Contains(string(rawIndex), "password: demo") || strings.Contains(string(rawIndex), "payload:") {
		t.Fatalf("expected index to omit sensitive payloads, got %s", string(rawIndex))
	}

	reloaded, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("reloaded NewManager() error = %v", err)
	}

	reloadedProfile, ok := reloaded.Get(profile.ID)
	if !ok {
		t.Fatalf("expected reloaded profile to exist")
	}
	if !reloadedProfile.Nodes[0].Enabled || reloadedProfile.Nodes[1].Enabled {
		t.Fatalf("expected reloaded profile to preserve node selection, got %+v", reloadedProfile.Nodes)
	}
}

func TestManagerRefreshPreservesNodeSelectionByName(t *testing.T) {
	_, path := testkit.TempProfilesStore(t)

	responseBody := `
proxies:
  - name: hk-01
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: original
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "refresh-selection",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if _, err := m.SetNodeEnabled(profile.ID, 0, false); err != nil {
		t.Fatalf("SetNodeEnabled() error = %v", err)
	}

	responseBody = `
proxies:
  - name: hk-01
    type: ss
    server: 8.8.8.8
    port: 8443
    cipher: aes-128-gcm
    password: refreshed
`
	refreshed, err := m.Refresh(profile.ID)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if len(refreshed.Nodes) != 1 {
		t.Fatalf("expected 1 refreshed node, got %d", len(refreshed.Nodes))
	}
	if refreshed.Nodes[0].Enabled {
		t.Fatalf("expected disabled selection to survive refresh, got %+v", refreshed.Nodes)
	}
	if !strings.Contains(refreshed.Nodes[0].Payload, "8.8.8.8") {
		t.Fatalf("expected refreshed payload, got %s", refreshed.Nodes[0].Payload)
	}
	if refreshed.Nodes[0].Server != "8.8.8.8" || refreshed.Nodes[0].Port != "8443" {
		t.Fatalf("expected refreshed metadata, got %+v", refreshed.Nodes[0])
	}
}

func TestManagerMoveNodePersistsOrderAndSurvivesReload(t *testing.T) {
	_, path := testkit.TempProfilesStore(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
proxies:
  - name: hk-01
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
  - name: us-02
    type: ss
    server: 2.2.2.2
    port: 8443
    cipher: aes-128-gcm
    password: demo
  - name: jp-03
    type: ss
    server: 3.3.3.3
    port: 9443
    cipher: aes-128-gcm
    password: demo
`))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "move-demo",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	moved, err := m.MoveNode(profile.ID, 2, 0)
	if err != nil {
		t.Fatalf("MoveNode() error = %v", err)
	}
	if len(moved.Nodes) != 3 {
		t.Fatalf("expected 3 nodes after move, got %d", len(moved.Nodes))
	}
	if moved.Nodes[0].Name != "jp-03" || moved.Nodes[0].Order != 0 {
		t.Fatalf("expected moved node first with normalized order, got %+v", moved.Nodes[0])
	}
	if moved.Nodes[1].Name != "hk-01" || moved.Nodes[1].Order != 1 {
		t.Fatalf("expected hk-01 second, got %+v", moved.Nodes[1])
	}
	if moved.Nodes[2].Name != "us-02" || moved.Nodes[2].Order != 2 {
		t.Fatalf("expected us-02 third, got %+v", moved.Nodes[2])
	}

	rawIndex, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(rawIndex), "order: 0") || !strings.Contains(string(rawIndex), "order: 2") {
		t.Fatalf("expected index to persist node order, got %s", string(rawIndex))
	}

	reloaded, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("reloaded NewManager() error = %v", err)
	}

	reloadedProfile, ok := reloaded.Get(profile.ID)
	if !ok {
		t.Fatalf("expected reloaded profile to exist")
	}
	if reloadedProfile.Nodes[0].Name != "jp-03" || reloadedProfile.Nodes[1].Name != "hk-01" || reloadedProfile.Nodes[2].Name != "us-02" {
		t.Fatalf("expected reloaded profile to preserve node order, got %+v", reloadedProfile.Nodes)
	}
}

func TestManagerRefreshPreservesCustomNodeOrderByName(t *testing.T) {
	_, path := testkit.TempProfilesStore(t)

	responseBody := `
proxies:
  - name: hk-01
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: demo
  - name: us-02
    type: ss
    server: 2.2.2.2
    port: 8443
    cipher: aes-128-gcm
    password: demo
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	m, err := NewManager(Config{StorePath: path})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile, err := m.Import(interfaces.ProfileImportRequest{
		Name:   "refresh-order",
		Source: server.URL,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if _, err := m.MoveNode(profile.ID, 1, 0); err != nil {
		t.Fatalf("MoveNode() error = %v", err)
	}

	responseBody = `
proxies:
  - name: us-02
    type: ss
    server: 9.9.9.9
    port: 9443
    cipher: aes-128-gcm
    password: refreshed
  - name: hk-01
    type: ss
    server: 8.8.8.8
    port: 8443
    cipher: aes-128-gcm
    password: refreshed
`
	refreshed, err := m.Refresh(profile.ID)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.Nodes[0].Name != "us-02" || refreshed.Nodes[1].Name != "hk-01" {
		t.Fatalf("expected refresh to preserve custom order, got %+v", refreshed.Nodes)
	}
	if refreshed.Nodes[0].Order != 0 || refreshed.Nodes[1].Order != 1 {
		t.Fatalf("expected refresh to normalize order fields, got %+v", refreshed.Nodes)
	}
}

func TestManagerRecoversProfilesWhenIndexIsMalformed(t *testing.T) {
	dir, path := testkit.TempProfilesStore(t)
	profilesDir := filepath.Join(dir, "profiles")
	testkit.MustMkdirAll(t, profilesDir)

	testkit.MustWriteString(t, path, "profiles: [")

	const validProfileID = "good-profile"
	testkit.MustWriteString(t, filepath.Join(profilesDir, validProfileID+".yaml"), `
proxies:
  - name: sg-demo
    type: ss
    server: 2.2.2.2
    port: 8443
    cipher: aes-128-gcm
    password: demo
`)

	testkit.MustWriteString(t, filepath.Join(profilesDir, "bad-profile.yaml"), "proxies: [")

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
