package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// LoadWatcherMapping looks for "watcher-mapping.yaml" inside sharedDir and
// returns a mapping of preset name -> list of project paths to auto-apply to.
// If the file does not exist, returns (nil, nil).
func LoadWatcherMapping(sharedDir string) (map[string][]string, error) {
	mappingPath := filepath.Join(sharedDir, "watcher-mapping.yaml")
	if _, err := os.Stat(mappingPath); os.IsNotExist(err) {
		return nil, nil
	}
	b, err := os.ReadFile(mappingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read watcher mapping: %w", err)
	}
	// expected format:
	// presets:
	//   frontend:
	//     - /abs/path/to/project
	//     - ../relative/project
	var raw struct {
		Presets map[string][]string `yaml:"presets"`
	}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse watcher mapping: %w", err)
	}
	if raw.Presets == nil {
		return nil, nil
	}
	// resolve relative paths (relative to sharedDir)
	resolved := make(map[string][]string)
	for preset, projects := range raw.Presets {
		for _, p := range projects {
			if !filepath.IsAbs(p) {
				p = filepath.Join(sharedDir, p)
			}
			resolved[preset] = append(resolved[preset], filepath.Clean(p))
		}
	}
	return resolved, nil
}
