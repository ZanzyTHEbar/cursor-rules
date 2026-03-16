package core

import (
	"os"
	"path/filepath"
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
	return ListSkillDirsFrom(skillsRoot)
}

// ListSkillDirsFrom lists skill directory names under the given skills root (e.g. .cursor/skills).
func ListSkillDirsFrom(skillsDir string) ([]string, error) {
	return listNamedDirResources(skillsDir, "SKILL.md")
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
	subdir := SkillsSubdir(skillsSubdir)
	skillsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path")
	}
	destParent, err := security.SafeJoin(projectRoot, ".cursor", "skills")
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid destination path")
	}
	return installNamedDirectoryResourceTo(destParent, skillsRoot, skillName, "SKILL.md")
}

// InstallSkillToDir installs a skill directory into the given skills directory.
func InstallSkillToDir(skillsDir, packageDir, skillName, skillsSubdir string) (InstallStrategy, error) {
	subdir := SkillsSubdir(skillsSubdir)
	skillsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path")
	}
	return installNamedDirectoryResourceTo(skillsDir, skillsRoot, skillName, "SKILL.md")
}
