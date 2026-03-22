package core

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
	"gopkg.in/yaml.v3"
)

const (
	commandSkillMarkerKey   = "cursor-rules-kind"
	commandSkillMarkerValue = "command"
)

// ListCursorCompatibleCommands returns command names from packageDir/commands.
// It supports native markdown commands plus compatibility formats like
// *.command.mdc and directory-backed command bundles.
func ListCursorCompatibleCommands(packageDir string) ([]string, error) {
	commandsRoot, err := security.SafeJoin(packageDir, defaultCommandsSubdir)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid commands directory")
	}
	entries, err := os.ReadDir(commandsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			if err := security.ValidatePackageName(entry.Name()); err != nil {
				continue
			}
			if commandBundleContainsContent(filepath.Join(commandsRoot, entry.Name())) {
				names[entry.Name()] = struct{}{}
			}
			continue
		}

		name, ok := commandNameFromFilename(entry.Name())
		if !ok {
			continue
		}
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		names[name] = struct{}{}
	}

	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

// InstallCommandAsSkillToDir installs a source command into Cursor's skills
// layout, converting it to a SKILL.md with explicit invocation semantics.
func InstallCommandAsSkillToDir(skillsDir, packageDir, command string, excludes []string) (InstallStrategy, error) {
	name, srcPath, isDir, err := locateCommandCompatSource(packageDir, command)
	if err != nil {
		return StrategyUnknown, err
	}
	if isDir {
		return installCommandBundleAsSkillToDir(skillsDir, srcPath, name, excludes)
	}
	return installCommandFileAsSkillToDir(skillsDir, srcPath, name)
}

// InstallCommandCollectionAsSkillsToDir installs all compatible commands from
// packageDir/commands into Cursor's skills directory.
func InstallCommandCollectionAsSkillsToDir(skillsDir, packageDir string, excludes []string) (InstallStrategy, error) {
	names, err := ListCursorCompatibleCommands(packageDir)
	if err != nil {
		return StrategyUnknown, err
	}
	if len(names) == 0 {
		return StrategyUnknown, errors.Newf(errors.CodeNotFound, "commands collection not found: %s", filepath.Join(packageDir, defaultCommandsSubdir))
	}

	for _, name := range names {
		if _, err := InstallCommandAsSkillToDir(skillsDir, packageDir, name, excludes); err != nil {
			return StrategyUnknown, err
		}
	}
	return StrategyCopy, nil
}

// ListInstalledCommandSkills returns installed command names by reading skills
// that were generated from commands by this tool.
func ListInstalledCommandSkills(skillsDir string) ([]string, error) {
	skillNames, err := ListSkillDirsFrom(skillsDir)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, name := range skillNames {
		skillPath := filepath.Join(skillsDir, name)
		if isCommandSkill(skillPath) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

func isCommandSkill(skillDir string) bool {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return false
	}
	node, _, err := transform.SplitFrontmatter(data)
	if err != nil {
		return false
	}
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return false
	}
	meta, ok := fm["metadata"].(map[string]interface{})
	if !ok {
		return false
	}
	v, ok := meta[commandSkillMarkerKey].(string)
	return ok && strings.TrimSpace(v) == commandSkillMarkerValue
}

func locateCommandCompatSource(packageDir, command string) (name, path string, isDir bool, err error) {
	name, err = normalizeCommandCompatName(command)
	if err != nil {
		return "", "", false, err
	}
	commandsRoot, err := security.SafeJoin(packageDir, defaultCommandsSubdir)
	if err != nil {
		return "", "", false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid commands directory")
	}

	dirPath, err := security.SafeJoin(commandsRoot, name)
	if err != nil {
		return "", "", false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command path")
	}
	if info, statErr := os.Stat(dirPath); statErr == nil && info.IsDir() {
		if commandBundleContainsContent(dirPath) {
			return name, dirPath, true, nil
		}
	}

	for _, candidate := range []string{name + ".command.mdc", name + ".md"} {
		filePath, joinErr := security.SafeJoin(commandsRoot, candidate)
		if joinErr != nil {
			return "", "", false, errors.Wrapf(joinErr, errors.CodeInvalidArgument, "invalid command source path")
		}
		if info, statErr := os.Stat(filePath); statErr == nil && !info.IsDir() {
			return name, filePath, false, nil
		}
	}

	return "", "", false, errors.Newf(errors.CodeNotFound, "command not found: %s", command)
}

func normalizeCommandCompatName(command string) (string, error) {
	normalized := strings.TrimSpace(command)
	normalized = strings.TrimSuffix(normalized, ".command.mdc")
	normalized = strings.TrimSuffix(normalized, ".md")
	normalized = strings.TrimSuffix(normalized, "/")
	normalized = strings.TrimSuffix(normalized, `\`)
	normalized = strings.TrimPrefix(normalized, defaultCommandsSubdir+"/")
	if err := security.ValidatePackageName(normalized); err != nil {
		return "", errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command name")
	}
	return normalized, nil
}

func commandNameFromFilename(name string) (string, bool) {
	switch {
	case strings.HasSuffix(name, ".command.mdc"):
		return strings.TrimSuffix(name, ".command.mdc"), true
	case strings.HasSuffix(name, ".md"):
		return strings.TrimSuffix(name, ".md"), true
	default:
		return "", false
	}
}

func commandBundleContainsContent(dir string) bool {
	found := false
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := commandNameFromFilename(d.Name()); ok {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found || err == fs.SkipAll
}

func installCommandFileAsSkillToDir(skillsDir, srcPath, commandName string) (InstallStrategy, error) {
	body, description, err := readCommandSource(srcPath)
	if err != nil {
		return StrategyUnknown, err
	}
	skillDir, err := security.SafeJoin(skillsDir, commandName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid skill destination")
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return StrategyUnknown, err
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	content, err := marshalCommandAsSkill(commandName, description, body)
	if err != nil {
		return StrategyUnknown, err
	}
	return StrategyCopy, writeIfChanged(skillPath, content)
}

func installCommandBundleAsSkillToDir(skillsDir, srcDir, commandName string, excludes []string) (InstallStrategy, error) {
	primary, err := choosePrimaryCommandDoc(srcDir, commandName)
	if err != nil {
		return StrategyUnknown, err
	}
	body, description, err := readCommandSource(primary)
	if err != nil {
		return StrategyUnknown, err
	}

	destRoot, err := security.SafeJoin(skillsDir, commandName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid skill destination")
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return StrategyUnknown, err
	}

	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == filepath.Base(primary) || path == primary {
			return nil
		}
		if shouldExcludeCompat(rel, excludes) {
			return nil
		}
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path in command bundle")
		}
		dest := filepath.Join(destRoot, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFileWithDirs(dest, data, 0o644)
	})
	if err != nil {
		return StrategyUnknown, err
	}

	content, err := marshalCommandAsSkill(commandName, description, body)
	if err != nil {
		return StrategyUnknown, err
	}
	return StrategyCopy, writeIfChanged(filepath.Join(destRoot, "SKILL.md"), content)
}

func choosePrimaryCommandDoc(srcDir, commandName string) (string, error) {
	preferred := []string{
		commandName + ".command.mdc",
		"COMMAND.command.mdc",
		commandName + ".md",
		"README.md",
	}
	for _, rel := range preferred {
		path := filepath.Join(srcDir, rel)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	var docs []string
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := commandNameFromFilename(d.Name()); ok {
			docs = append(docs, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(docs)
	if len(docs) == 0 {
		return "", errors.Newf(errors.CodeNotFound, "command bundle missing markdown source: %s", srcDir)
	}
	return docs[0], nil
}

func readCommandSource(path string) (body, description string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	if strings.HasSuffix(path, ".mdc") {
		node, mdBody, splitErr := transform.SplitFrontmatter(data)
		if splitErr != nil {
			return "", "", splitErr
		}
		var fm map[string]interface{}
		if err := node.Decode(&fm); err != nil {
			return "", "", err
		}
		if desc, ok := fm["description"].(string); ok {
			description = strings.TrimSpace(desc)
		}
		body = mdBody
	} else {
		body = strings.TrimSpace(string(data))
	}

	if description == "" {
		description = "Explicitly invoked command migrated to a Cursor skill."
	}
	return body, description, nil
}

func marshalCommandAsSkill(name, description, body string) ([]byte, error) {
	fm := map[string]interface{}{
		"name":                     name,
		"description":              description,
		"disable-model-invocation": true,
		"metadata": map[string]string{
			commandSkillMarkerKey: commandSkillMarkerValue,
		},
	}
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "marshal command skill frontmatter")
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(body))
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func shouldExcludeCompat(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}
		matched, err = filepath.Match(pattern, filepath.Dir(relPath))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func writeIfChanged(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return nil
	}
	return writeFileWithDirs(path, data, 0o644)
}

func writeFileWithDirs(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}
