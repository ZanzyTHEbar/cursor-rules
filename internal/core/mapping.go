package core

import (
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"gopkg.in/yaml.v3"
)

// LoadWatcherMapping looks for "watcher-mapping.yaml" inside packageDir and
// returns a mapping of preset name -> list of project paths to auto-apply to.
// If the file does not exist, returns (nil, nil).
func LoadWatcherMapping(packageDir string) (map[string][]string, error) {
	// Safely construct mapping file path
	mappingPath, err := security.SafeJoin(packageDir, "watcher-mapping.yaml")
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid package directory path")
	}

	if _, statErr := os.Stat(mappingPath); os.IsNotExist(statErr) {
		return nil, nil
	}

	// #nosec G304 - mappingPath is validated above and constructed from trusted packageDir
	b, readErr := os.ReadFile(mappingPath)
	if readErr != nil {
		return nil, errors.Wrapf(readErr, errors.CodeInternal, "read watcher mapping")
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
		return nil, errors.Wrapf(err, errors.CodeInternal, "parse watcher mapping")
	}
	if raw.Presets == nil {
		return nil, nil
	}
	// resolve relative paths (relative to packageDir)
	resolved := make(map[string][]string)
	for preset, projects := range raw.Presets {
		// Validate preset name
		if validErr := security.ValidatePackageName(preset); validErr != nil {
			return nil, errors.Wrapf(validErr, errors.CodeInvalidArgument, "invalid preset name %q in mapping", preset)
		}

		for _, p := range projects {
			// Validate path
			if validErr := security.ValidatePath(p); validErr != nil {
				return nil, errors.Wrapf(validErr, errors.CodeInvalidArgument, "invalid project path %q for preset %q", p, preset)
			}

			if !filepath.IsAbs(p) {
				// Safely join relative path with packageDir
				absPath, joinErr := security.SafeJoin(packageDir, p)
				if joinErr != nil {
					return nil, errors.Wrapf(joinErr, errors.CodeInvalidArgument, "invalid relative path %q for preset %q", p, preset)
				}
				p = absPath
			}
			resolved[preset] = append(resolved[preset], filepath.Clean(p))
		}
	}
	return resolved, nil
}
