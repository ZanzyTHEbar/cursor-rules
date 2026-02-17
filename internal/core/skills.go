package core

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

const defaultSkillsSubdir = "skills"

// SkillsSubdir returns the subdir name for skills under the package dir (default "skills").
func SkillsSubdir(configured string) string {
	if s := strings.TrimSpace(configured); s != "" {
		return s
	}
	return defaultSkillsSubdir
}

// ListSkillDirs lists directory names under packageDir/skills that contain SKILL.md.
// skillsSubdir can be empty to use "skills".
func ListSkillDirs(packageDir, skillsSubdir string) ([]string, error) {
	subdir := SkillsSubdir(skillsSubdir)
	skillsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid package dir or skills subdir")
	}
	entries, err := os.ReadDir(skillsRoot)
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
		skillPath := filepath.Join(skillsRoot, name, "SKILL.md")
		if _, statErr := os.Stat(skillPath); statErr == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// ReadSkillMeta reads name and description from a skill's SKILL.md frontmatter.
func ReadSkillMeta(skillDir string) (name, description string, err error) {
	skillMD := filepath.Join(skillDir, "SKILL.md")
	data, err := os.ReadFile(skillMD)
	if err != nil {
		return "", "", errors.Wrapf(err, errors.CodeNotFound, "read SKILL.md")
	}
	node, _, err := transform.SplitFrontmatter(data)
	if err != nil {
		return "", "", errors.Wrapf(err, errors.CodeInvalidArgument, "parse SKILL.md frontmatter")
	}
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return "", "", errors.Wrapf(err, errors.CodeInvalidArgument, "decode frontmatter")
	}
	if n, ok := fm["name"]; ok {
		if s, ok := n.(string); ok {
			name = strings.TrimSpace(s)
		}
	}
	if d, ok := fm["description"]; ok {
		if s, ok := d.(string); ok {
			description = strings.TrimSpace(s)
		}
	}
	return name, description, nil
}

// InstallSkillToProject installs a skill directory from packageDir into projectRoot/.cursor/skills/<skillName>/.
func InstallSkillToProject(projectRoot, packageDir, skillName, skillsSubdir string) (InstallStrategy, error) {
	if err := security.ValidatePackageName(skillName); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid skill name")
	}
	subdir := SkillsSubdir(skillsSubdir)
	skillSrc, err := security.SafeJoin(packageDir, subdir, skillName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path")
	}
	info, err := os.Stat(skillSrc)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeNotFound, "skill not found: %s", skillName)
		}
		return StrategyUnknown, err
	}
	if !info.IsDir() {
		return StrategyUnknown, errors.Newf(errors.CodeFailedPrecondition, "skill path is not a directory: %s", skillSrc)
	}
	skillMD := filepath.Join(skillSrc, "SKILL.md")
	if _, err := os.Stat(skillMD); err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeFailedPrecondition, "skill missing SKILL.md: %s", skillName)
		}
		return StrategyUnknown, err
	}
	destRoot, err := security.SafeJoin(projectRoot, ".cursor", "skills", skillName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return StrategyUnknown, err
	}

	skillsRoot := filepath.Join(packageDir, subdir)
	if WantGNUStow() && HasStow() {
		cursorSkillsDir := filepath.Join(projectRoot, ".cursor", "skills")
		if err := os.MkdirAll(cursorSkillsDir, 0o755); err != nil {
			return StrategyUnknown, err
		}
		cmd := exec.Command("stow", "-v", "-d", skillsRoot, "-t", cursorSkillsDir, skillName)
		if out, cmdErr := cmd.CombinedOutput(); cmdErr == nil {
			_ = out
			return StrategyStow, nil
		}
	}

	var strategy InstallStrategy = StrategyCopy
	err = filepath.WalkDir(skillSrc, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(skillSrc, path)
		if relErr != nil {
			return relErr
		}
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid file path in skill %s", skillName)
		}
		dest := filepath.Join(destRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if UseSymlink() || WantGNUStow() {
			if symErr := CreateSymlink(path, dest); symErr == nil {
				strategy = StrategySymlink
				return nil
			}
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if writeErr := os.WriteFile(dest, data, 0o644); writeErr != nil {
			return writeErr
		}
		return nil
	})
	if err != nil {
		return StrategyUnknown, err
	}
	return strategy, nil
}
