package storage

import (
	"errors"
	"fmt"
	"os"

	"sakiko.local/sakiko-core/interfaces"

	"go.yaml.in/yaml/v3"
)

type profilesYAML struct {
	Profiles []profileIndexEntry `yaml:"profiles"`
}

type profileNodeIndexEntry struct {
	Name    string `yaml:"name"`
	Order   int    `yaml:"order"`
	Enabled bool   `yaml:"enabled"`
}

type profileIndexEntry struct {
	ID         string                  `yaml:"id"`
	Name       string                  `yaml:"name"`
	Source     string                  `yaml:"source"`
	Nodes      []profileNodeIndexEntry `yaml:"nodes,omitempty"`
	UpdatedAt  string                  `yaml:"updatedAt,omitempty"`
	Attributes interface{}             `yaml:"attributes,omitempty"`
}

func newProfileIndexEntry(profile interfaces.Profile) profileIndexEntry {
	nodes := make([]profileNodeIndexEntry, 0, len(profile.Nodes))
	for _, node := range profile.Nodes {
		nodes = append(nodes, profileNodeIndexEntry{
			Name:    node.Name,
			Order:   node.Order,
			Enabled: node.Enabled,
		})
	}

	return profileIndexEntry{
		ID:         profile.ID,
		Name:       profile.Name,
		Source:     profile.Source,
		Nodes:      nodes,
		UpdatedAt:  profile.UpdatedAt,
		Attributes: profile.Attributes,
	}
}

func (e profileIndexEntry) toProfile() interfaces.Profile {
	nodes := make([]interfaces.Node, 0, len(e.Nodes))
	for _, node := range e.Nodes {
		nodes = append(nodes, interfaces.Node{
			Name:    node.Name,
			Order:   node.Order,
			Enabled: node.Enabled,
		})
	}

	return interfaces.Profile{
		ID:         e.ID,
		Name:       e.Name,
		Source:     e.Source,
		Nodes:      nodes,
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

	return writeFileAtomic(s.path, raw, 0o644)
}
