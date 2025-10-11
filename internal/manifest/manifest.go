package manifest

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest defines the structure of cursor-rules-manifest.yaml files.
type Manifest struct {
	Version   string              `yaml:"version"`
	Targets   []string            `yaml:"targets"`
	Overrides map[string]Override `yaml:"overrides,omitempty"`
	Exclude   []string            `yaml:"exclude,omitempty"`
}

// Override defines target-specific configuration overrides.
type Override struct {
	DefaultMode  string   `yaml:"defaultMode,omitempty"`
	DefaultTools []string `yaml:"defaultTools,omitempty"`
	IncludeRefs  []string `yaml:"includeRefs,omitempty"`
}

// Load reads and parses a cursor-rules-manifest.yaml file from the given package path.
// Returns nil if the manifest file doesn't exist (it's optional).
func Load(pkgPath string) (*Manifest, error) {
	manifestPath := filepath.Join(pkgPath, "cursor-rules-manifest.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Manifest is optional
		}
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// HasTarget checks if the manifest includes a specific target.
func (m *Manifest) HasTarget(target string) bool {
	if m == nil {
		return false
	}
	for _, t := range m.Targets {
		if t == target {
			return true
		}
	}
	return false
}

// GetOverride retrieves target-specific overrides if they exist.
func (m *Manifest) GetOverride(target string) *Override {
	if m == nil || m.Overrides == nil {
		return nil
	}
	if override, ok := m.Overrides[target]; ok {
		return &override
	}
	return nil
}
