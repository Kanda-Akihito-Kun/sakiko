package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type replacement struct {
	old string
	new string
}

func main() {
	targets := map[string][]replacement{
		filepath.Join("windows", "info.json"): {
			{`This is a comment`, ``},
		},
		filepath.Join("darwin", "Info.plist"): {
			{`This is a comment`, ``},
		},
		filepath.Join("darwin", "Info.dev.plist"): {
			{`This is a comment`, ``},
		},
		filepath.Join("ios", "Info.plist"): {
			{`This is a comment`, ``},
		},
		filepath.Join("ios", "Info.dev.plist"): {
			{`This is a comment`, ``},
		},
	}

	for path, replacements := range targets {
		if err := sanitizeFile(path, replacements); err != nil {
			fmt.Fprintf(os.Stderr, "sanitize %s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

func sanitizeFile(path string, replacements []replacement) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated := string(content)
	for _, replacement := range replacements {
		updated = strings.ReplaceAll(updated, replacement.old, replacement.new)
	}

	if updated == string(content) {
		return nil
	}

	return os.WriteFile(path, []byte(updated), 0o644)
}
