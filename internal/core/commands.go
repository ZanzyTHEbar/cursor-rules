package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

const defaultCommandsSubdir = "commands"

// CommandsSubdir returns the reserved package subdir used for commands.
func CommandsSubdir() string {
	return defaultCommandsSubdir
}

// DefaultSharedCommandsDir returns ~/.cursor-commands by default; environment overrides allowed.
func DefaultSharedCommandsDir() string {
	// Commands live under the main cursor-rules package directory. Use that by default.
	// But if CURSOR_COMMANDS_DIR is explicitly set, use that instead.
	if v := os.Getenv("CURSOR_COMMANDS_DIR"); v != "" {
		return v
	}
	return DefaultPackageDir()
}

// InstallCommand installs a shared command into the project's .cursor/commands.
func InstallCommand(projectRoot, command string) error {
	_, err := InstallCommandToProject(projectRoot, DefaultSharedCommandsDir(), command, nil, false)
	return err
}

// InstallCommandToProject installs a command from the package dir into projectRoot/.cursor/commands.
// It prefers directory-backed commands under commands/<name>/ over flat <name>.md files.
func InstallCommandToProject(projectRoot, sourceDir, command string, excludes []string, noFlatten bool) (InstallStrategy, error) {
	commandsDir, err := security.SafeJoin(projectRoot, ".cursor", defaultCommandsSubdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	return InstallCommandToProjectWithCommandsDir(commandsDir, sourceDir, command, excludes, noFlatten)
}

// InstallCommandToProjectWithCommandsDir installs a command into the given commands directory.
func InstallCommandToProjectWithCommandsDir(commandsDir, sourceDir, command string, excludes []string, noFlatten bool) (InstallStrategy, error) {
	raw := strings.TrimSpace(command)
	normalized, err := normalizeCommandName(raw)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command name")
	}

	if strings.HasSuffix(raw, ".md") {
		if strategy, found, err := installCommandCollectionFileToCommandsDirIfExists(commandsDir, sourceDir, normalized); found || err != nil {
			return strategy, err
		}
		return installCommandFileToCommandsDir(commandsDir, sourceDir, normalized)
	}

	if strategy, found, err := installCommandSubdirToCommandsDir(commandsDir, sourceDir, normalized, excludes); found || err != nil {
		return strategy, err
	}

	if strategy, found, err := installCommandCollectionFileToCommandsDirIfExists(commandsDir, sourceDir, normalized); found || err != nil {
		return strategy, err
	}

	if strategy, found, err := installCommandFileToCommandsDirIfExists(commandsDir, sourceDir, normalized); found || err != nil {
		return strategy, err
	}

	if strategy, found, err := installLegacyCommandPackageToCommandsDir(commandsDir, sourceDir, normalized, excludes, noFlatten); found || err != nil {
		return strategy, err
	}

	return StrategyUnknown, errors.Newf(errors.CodeNotFound, "command not found: %s", command)
}

// ApplyCommandToProject installs a command into the project's .cursor/commands.
func ApplyCommandToProject(projectRoot, command, sourceDir string) error {
	_, err := InstallCommandToProject(projectRoot, sourceDir, command, nil, false)
	return err
}

// ApplyCommandWithOptionalSymlink attempts to apply a command via stow/symlink or stub.
func ApplyCommandWithOptionalSymlink(projectRoot, command, sourceDir string) error {
	normalized, err := normalizeCommandName(command)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command name")
	}
	_, err = installCommandFileToProject(projectRoot, sourceDir, normalized)
	return err
}

// ListSharedCommands returns command entries from the package dir.
// Supported source layouts are <name>.md and commands/<name>/.
func ListSharedCommands(commandsDir string) ([]string, error) {
	return ListPackageCommands(commandsDir)
}

// ListPackageCommands returns command entries from the package dir.
func ListPackageCommands(packageDir string) ([]string, error) {
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, err
	}
	names := map[string]struct{}{}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			names[e.Name()] = struct{}{}
		}
	}

	commandEntries, err := ListCommandsCollectionEntries(packageDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, name := range commandEntries {
		names[name] = struct{}{}
	}

	return sortedCommandNames(names), nil
}

// ListCommandsCollectionEntries returns command entries declared under packageDir/commands.
// Supported layouts are commands/<name>.md and commands/<name>/.
func ListCommandsCollectionEntries(packageDir string) ([]string, error) {
	commandsRoot, err := security.SafeJoin(packageDir, defaultCommandsSubdir)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid commands directory")
	}
	entries, err := listCommandEntries(commandsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return entries, nil
}

// ListInstalledCommands returns command entries from .cursor/commands or the user commands dir.
func ListInstalledCommands(commandsDir string) ([]string, error) {
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	names := map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() {
			if commandDirContainsMarkdown(filepath.Join(commandsDir, e.Name())) {
				names[e.Name()] = struct{}{}
			}
			continue
		}
		if filepath.Ext(e.Name()) == ".md" {
			names[e.Name()] = struct{}{}
		}
	}

	return sortedCommandNames(names), nil
}

// InstallCommandPackage installs an entire package directory from sharedDir into the project's
// .cursor/commands. The package is a directory under sharedDir (e.g. "tools" or "git-helpers").
// It supports excluding specific files via the excludes slice and respects a
// .cursor-commands-ignore file placed inside the package which lists patterns to skip.
// By default, packages are flattened into .cursor/commands/. Use noFlatten=true to preserve structure.
func InstallCommandPackage(projectRoot, packageName string, excludes []string, noFlatten bool) error {
	sharedDir := DefaultSharedCommandsDir()
	return InstallCommandPackageFromDir(projectRoot, sharedDir, packageName, excludes, noFlatten)
}

// InstallCommandPackageFromDir installs a command package from sourceDir into the project's .cursor/commands.
func InstallCommandPackageFromDir(projectRoot, sourceDir, packageName string, excludes []string, noFlatten bool) error {
	_, err := InstallPackageGeneric(projectRoot, sourceDir, packageName, "commands", []string{".md"}, ".cursor-commands-ignore", excludes, noFlatten)
	return err
}

// InstallCommandCollectionToProject installs every command entry found under packageDir/commands.
func InstallCommandCollectionToProject(projectRoot, packageDir string, excludes []string) (InstallStrategy, error) {
	commandsDir, err := security.SafeJoin(projectRoot, ".cursor", defaultCommandsSubdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	return InstallCommandCollectionToCommandsDir(commandsDir, packageDir, excludes)
}

// InstallCommandCollectionToCommandsDir installs every command entry into the given commands directory.
func InstallCommandCollectionToCommandsDir(commandsDir, packageDir string, excludes []string) (InstallStrategy, error) {
	names, err := ListCommandsCollectionEntries(packageDir)
	if err != nil {
		return StrategyUnknown, err
	}
	if len(names) == 0 {
		return StrategyUnknown, errors.Newf(errors.CodeNotFound, "commands collection not found: %s", filepath.Join(packageDir, defaultCommandsSubdir))
	}

	usedStrategy := StrategyCopy
	for _, name := range names {
		strategy, err := InstallCommandToProjectWithCommandsDir(commandsDir, packageDir, name, excludes, false)
		if err != nil {
			return StrategyUnknown, err
		}
		if strategy != StrategyCopy {
			usedStrategy = strategy
		}
	}
	return usedStrategy, nil
}

func normalizeCommandName(command string) (string, error) {
	normalized := strings.TrimSpace(command)
	normalized = strings.TrimSuffix(normalized, ".md")
	normalized = strings.TrimSuffix(normalized, "/")
	normalized = strings.TrimSuffix(normalized, `\`)
	normalized = strings.TrimPrefix(normalized, defaultCommandsSubdir+"/")
	if err := security.ValidatePackageName(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}

func installCommandFileToProject(projectRoot, sourceDir, command string) (InstallStrategy, error) {
	return installNamedFileResource(projectRoot, sourceDir, "commands", command, ".md")
}

func installCommandFileToCommandsDir(commandsDir, sourceDir, command string) (InstallStrategy, error) {
	return installNamedFileResourceTo(commandsDir, sourceDir, command, ".md")
}

func installCommandFileToCommandsDirIfExists(commandsDir, sourceDir, command string) (InstallStrategy, bool, error) {
	src, err := security.SafeJoin(sourceDir, command+".md")
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid shared command path")
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, false, nil
		}
		return StrategyUnknown, false, err
	}
	if info.IsDir() {
		return StrategyUnknown, false, nil
	}
	strategy, err := installCommandFileToCommandsDir(commandsDir, sourceDir, command)
	return strategy, true, err
}

func installCommandCollectionFileToCommandsDirIfExists(commandsDir, sourceDir, command string) (InstallStrategy, bool, error) {
	commandsRoot, err := security.SafeJoin(sourceDir, defaultCommandsSubdir)
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid commands directory")
	}
	src, err := security.SafeJoin(commandsRoot, command+".md")
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command file path")
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, false, nil
		}
		return StrategyUnknown, false, err
	}
	if info.IsDir() {
		return StrategyUnknown, false, nil
	}
	strategy, err := installCommandFileToCommandsDir(commandsDir, commandsRoot, command)
	return strategy, true, err
}

func installCommandSubdirToCommandsDir(commandsDir, sourceDir, command string, excludes []string) (InstallStrategy, bool, error) {
	commandsRoot, err := security.SafeJoin(sourceDir, defaultCommandsSubdir)
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid commands directory")
	}
	srcDir, err := security.SafeJoin(commandsRoot, command)
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command directory")
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, false, nil
		}
		return StrategyUnknown, false, err
	}
	if !info.IsDir() {
		return StrategyUnknown, false, nil
	}
	if !commandDirContainsMarkdown(srcDir) {
		return StrategyUnknown, false, nil
	}
	strategy, err := InstallPackageGenericToDest(commandsDir, commandsRoot, command, []string{".md"}, ".cursor-commands-ignore", excludes, true)
	return strategy, true, err
}

func installLegacyCommandPackageToCommandsDir(commandsDir, sourceDir, command string, excludes []string, noFlatten bool) (InstallStrategy, bool, error) {
	srcDir, err := security.SafeJoin(sourceDir, command)
	if err != nil {
		return StrategyUnknown, false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command package path")
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, false, nil
		}
		return StrategyUnknown, false, err
	}
	if !info.IsDir() {
		return StrategyUnknown, false, nil
	}
	if !commandDirContainsMarkdown(srcDir) {
		return StrategyUnknown, false, nil
	}
	strategy, err := InstallPackageGenericToDest(commandsDir, sourceDir, command, []string{".md"}, ".cursor-commands-ignore", excludes, noFlatten)
	return strategy, true, err
}

func listCommandEntries(commandsRoot string) ([]string, error) {
	entries, err := os.ReadDir(commandsRoot)
	if err != nil {
		return nil, err
	}

	names := map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() {
			if err := security.ValidatePackageName(e.Name()); err != nil {
				continue
			}
			if commandDirContainsMarkdown(filepath.Join(commandsRoot, e.Name())) {
				names[e.Name()] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".md") {
			names[e.Name()] = struct{}{}
		}
	}

	return sortedCommandNames(names), nil
}

func commandDirContainsMarkdown(dir string) bool {
	found := false
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found || err == fs.SkipAll
}

func sortedCommandNames(names map[string]struct{}) []string {
	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
