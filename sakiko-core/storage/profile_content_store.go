package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ProfileContentStore struct {
	dir string
}

type ProfileContentFile struct {
	ID        string
	Path      string
	Content   string
	UpdatedAt time.Time
}

func NewProfileContentStore(indexPath string) *ProfileContentStore {
	baseDir := filepath.Dir(indexPath)
	if baseDir == "." || baseDir == "" {
		baseDir = ""
	}

	contentDir := "profiles"
	if baseDir != "" {
		contentDir = filepath.Join(baseDir, "profiles")
	}

	return &ProfileContentStore{dir: contentDir}
}

func (s *ProfileContentStore) Save(profileID string, content string) error {
	if s == nil {
		return fmt.Errorf("profile content store is nil")
	}
	if profileID == "" {
		return fmt.Errorf("profile ID is required")
	}
	if s.dir == "" {
		return fmt.Errorf("profile content directory is required")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.dir, profileID+".yaml"), []byte(content), 0o644)
}

func (s *ProfileContentStore) Path(profileID string) string {
	if s == nil || s.dir == "" || profileID == "" {
		return ""
	}
	return filepath.Join(s.dir, profileID+".yaml")
}

func (s *ProfileContentStore) Load(profileID string) (ProfileContentFile, error) {
	if s == nil {
		return ProfileContentFile{}, fmt.Errorf("profile content store is nil")
	}

	path := s.Path(profileID)
	if path == "" {
		return ProfileContentFile{}, fmt.Errorf("profile ID is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return ProfileContentFile{}, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return ProfileContentFile{}, err
	}

	return ProfileContentFile{
		ID:        profileID,
		Path:      path,
		Content:   string(raw),
		UpdatedAt: info.ModTime(),
	}, nil
}

func (s *ProfileContentStore) Delete(profileID string) error {
	if s == nil {
		return fmt.Errorf("profile content store is nil")
	}

	path := s.Path(profileID)
	if path == "" {
		return fmt.Errorf("profile ID is required")
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *ProfileContentStore) LoadAll() ([]ProfileContentFile, error) {
	if s == nil {
		return nil, fmt.Errorf("profile content store is nil")
	}
	if s.dir == "" {
		return []ProfileContentFile{}, nil
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []ProfileContentFile{}, nil
		}
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	files := make([]ProfileContentFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		fullPath := filepath.Join(s.dir, name)
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, ProfileContentFile{
			ID:        strings.TrimSuffix(name, filepath.Ext(name)),
			Path:      fullPath,
			Content:   string(raw),
			UpdatedAt: info.ModTime(),
		})
	}

	return files, nil
}
