package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func TempProfilesStore(t *testing.T) (root string, storePath string) {
	t.Helper()

	root = t.TempDir()
	return root, filepath.Join(root, "profiles.yaml")
}

func MustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func MustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func MustWriteString(t *testing.T, path string, data string) {
	t.Helper()
	MustWriteFile(t, path, []byte(data))
}
