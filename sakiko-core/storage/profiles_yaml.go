package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"sakiko.local/sakiko-core/interfaces"

	"go.yaml.in/yaml/v3"
)

type profilesYAML struct {
	Profiles []profileIndexEntry `yaml:"profiles"`
}

type profileIndexEntry struct {
	ID         string      `yaml:"id"`
	Name       string      `yaml:"name"`
	Source     string      `yaml:"source"`
	UpdatedAt  string      `yaml:"updatedAt,omitempty"`
	Attributes interface{} `yaml:"attributes,omitempty"`
}

func newProfileIndexEntry(profile interfaces.Profile) profileIndexEntry {
	return profileIndexEntry{
		ID:         profile.ID,
		Name:       profile.Name,
		Source:     profile.Source,
		UpdatedAt:  profile.UpdatedAt,
		Attributes: profile.Attributes,
	}
}

func (e profileIndexEntry) toProfile() interfaces.Profile {
	return interfaces.Profile{
		ID:         e.ID,
		Name:       e.Name,
		Source:     e.Source,
		UpdatedAt:  e.UpdatedAt,
		Attributes: e.Attributes,
	}
}

type ProfileStore struct {
	path string
}

func NewProfileStore(path string) *ProfileStore {
	return &ProfileStore{path: path}
}

func (s *ProfileStore) Load() ([]interfaces.Profile, error) {
	if s == nil {
		return nil, fmt.Errorf("profile store is nil")
	}
	if s.path == "" {
		return []interfaces.Profile{}, nil
	}

	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []interfaces.Profile{}, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return []interfaces.Profile{}, nil
	}

	var file profilesYAML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, err
	}
	if file.Profiles == nil {
		return []interfaces.Profile{}, nil
	}
	profiles := make([]interfaces.Profile, 0, len(file.Profiles))
	for _, entry := range file.Profiles {
		profiles = append(profiles, entry.toProfile())
	}
	return profiles, nil
}

func (s *ProfileStore) Save(profiles []interfaces.Profile) error {
	if s == nil {
		return fmt.Errorf("profile store is nil")
	}
	if s.path == "" {
		return fmt.Errorf("profile path is required")
	}

	items := make([]profileIndexEntry, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, newProfileIndexEntry(profile))
	}

	file := profilesYAML{Profiles: items}
	raw, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(s.path, raw, 0o644)
}
