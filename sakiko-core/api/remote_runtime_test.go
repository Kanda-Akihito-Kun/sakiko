package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestShouldRetryRemoteKnightArchiveLoad(t *testing.T) {
	notExistErr := &os.PathError{
		Op:   "open",
		Path: filepath.Join("results", "remote-knight", "task.json"),
		Err:  os.ErrNotExist,
	}
	if !shouldRetryRemoteKnightArchiveLoad(notExistErr) {
		t.Fatalf("expected not-exist archive load error to be retryable")
	}

	if shouldRetryRemoteKnightArchiveLoad(errors.New("permission denied")) {
		t.Fatalf("expected non-not-exist archive load error to be non-retryable")
	}
}
