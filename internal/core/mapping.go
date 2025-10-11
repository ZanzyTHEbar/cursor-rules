package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"gopkg.in/yaml.v3"
)

// LoadWatcherMapping looks for "watcher-mapping.yaml" inside sharedDir and
// returns a mapping of preset name -> list of project paths to auto-apply to.
// If the file does not exist, returns (nil, nil).
func LoadWatcherMapping(sharedDir string) (map[string][]string, error) {
	// Safely construct mapping file path
	mappingPath, err := security.SafeJoin(sharedDir, "watcher-mapping.yaml")
	if err != nil {
		return nil, fmt.Errorf("invalid shared directory path: %w", err)
	}
	
	if _, statErr := os.Stat(mappingPath); os.IsNotExist(statErr) {
		return nil, nil
	}
	
	// #nosec G304 - mappingPath is validated above and constructed from trusted sharedDir
	b, readErr := os.ReadFile(mappingPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read watcher mapping: %w", readErr)
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
		// Validate preset name
		if validErr := security.ValidatePackageName(preset); validErr != nil {
			return nil, fmt.Errorf("invalid preset name %q in mapping: %w", preset, validErr)
		}
		
		for _, p := range projects {
			// Validate path
			if validErr := security.ValidatePath(p); validErr != nil {
				return nil, fmt.Errorf("invalid project path %q for preset %q: %w", p, preset, validErr)
			}
			
			if !filepath.IsAbs(p) {
				// Safely join relative path with sharedDir
				absPath, joinErr := security.SafeJoin(sharedDir, p)
				if joinErr != nil {
					return nil, fmt.Errorf("invalid relative path %q for preset %q: %w", p, preset, joinErr)
				}
				p = absPath
			}
			resolved[preset] = append(resolved[preset], filepath.Clean(p))
		}
	}
	return resolved, nil
}
