package core

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

const defaultHooksSubdir = "hooks"
const hooksJSONName = "hooks.json"

// HooksSubdir returns the subdir name for hooks under the package dir (default "hooks").
func HooksSubdir(configured string) string {
	if s := strings.TrimSpace(configured); s != "" {
		return s
	}
	return defaultHooksSubdir
}

// ListHookPresets lists directory names under packageDir/hooks that contain hooks.json.
func ListHookPresets(packageDir, hooksSubdir string) ([]string, error) {
	subdir := HooksSubdir(hooksSubdir)
	hooksRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid package dir or hooks subdir")
	}
	entries, err := os.ReadDir(hooksRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		jsonPath := filepath.Join(hooksRoot, name, hooksJSONName)
		if _, statErr := os.Stat(jsonPath); statErr == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// InstallHookPresetToProject installs a hook preset: copies scripts to projectRoot/.cursor/hooks/,
// rewrites command paths in hooks.json to .cursor/hooks/<script>, and writes projectRoot/.cursor/hooks.json.
func InstallHookPresetToProject(projectRoot, packageDir, presetName, hooksSubdir string) (InstallStrategy, error) {
	if err := security.ValidatePackageName(presetName); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid hook preset name")
	}
	subdir := HooksSubdir(hooksSubdir)
	presetDir, err := security.SafeJoin(packageDir, subdir, presetName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path")
	}
	jsonPath := filepath.Join(presetDir, hooksJSONName)
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeNotFound, "hook preset not found: %s", presetName)
		}
		return StrategyUnknown, err
	}
	var cfg hooksConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid hooks.json")
	}
	if cfg.Hooks == nil {
		cfg.Hooks = make(map[string][]hookDef)
	}

	destHooksDir, err := security.SafeJoin(projectRoot, ".cursor", "hooks")
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	if err := os.MkdirAll(destHooksDir, 0o755); err != nil {
		return StrategyUnknown, err
	}

	// Copy or symlink all script files from preset dir into .cursor/hooks/
	strategy := StrategyCopy
	err = filepath.WalkDir(presetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() == hooksJSONName {
			return nil
		}
		rel, relErr := filepath.Rel(presetDir, path)
		if relErr != nil {
			return relErr
		}
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path in preset")
		}
		dest := filepath.Join(destHooksDir, filepath.Base(path))
		if UseSymlink() || WantGNUStow() {
			if symErr := CreateSymlink(path, dest); symErr == nil {
				strategy = StrategySymlink
				return nil
			}
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(dest, content, 0o755)
	})
	if err != nil {
		return StrategyUnknown, err
	}

	// Rewrite command paths in cfg to .cursor/hooks/<basename>
	rewriteHookCommands(&cfg, presetDir, destHooksDir)

	out, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInternal, "marshal hooks.json")
	}
	destJSON, err := security.SafeJoin(projectRoot, ".cursor", hooksJSONName)
	if err != nil {
		return StrategyUnknown, err
	}
	if err := os.WriteFile(destJSON, out, 0o644); err != nil {
		return StrategyUnknown, err
	}
	return strategy, nil
}

type hooksConfig struct {
	Version int                  `json:"version"`
	Hooks   map[string][]hookDef `json:"hooks"`
}

type hookDef struct {
	Command   string `json:"command,omitempty"`
	Type      string `json:"type,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
	LoopLimit *int   `json:"loop_limit,omitempty"`
	Matcher   string `json:"matcher,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
	// allow unknown fields by using map or extra fields; json.Unmarshal will drop unknown
}

// rewriteHookCommands rewrites command paths in cfg from preset-relative to project-relative .cursor/hooks/<name>.
func rewriteHookCommands(cfg *hooksConfig, presetDir, destHooksDir string) {
	for event, list := range cfg.Hooks {
		for i := range list {
			cmd := strings.TrimSpace(list[i].Command)
			if cmd == "" {
				continue
			}
			// Resolve path relative to preset dir (e.g. ./scripts/format.sh -> presetDir/scripts/format.sh)
			var abs string
			if filepath.IsAbs(cmd) {
				abs = cmd
			} else {
				abs = filepath.Join(presetDir, cmd)
			}
			abs = filepath.Clean(abs)
			if _, err := os.Stat(abs); err != nil {
				continue
			}
			base := filepath.Base(abs)
			// Cursor runs project hooks from project root; use .cursor/hooks/<base>
			list[i].Command = filepath.Join(".cursor", "hooks", base)
		}
		cfg.Hooks[event] = list
	}
}

// RemoveHookPresetFromProject removes projectRoot/.cursor/hooks.json and projectRoot/.cursor/hooks/.
func RemoveHookPresetFromProject(projectRoot string) error {
	jsonPath := filepath.Join(projectRoot, ".cursor", hooksJSONName)
	hooksDir := filepath.Join(projectRoot, ".cursor", "hooks")
	if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, errors.CodeInternal, "remove hooks.json")
	}
	if err := os.RemoveAll(hooksDir); err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "remove hooks dir")
	}
	return nil
}
