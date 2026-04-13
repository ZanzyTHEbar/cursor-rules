package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

// InstallOpenCodeCommandToDir installs a command into OpenCode's native commands layout.
func InstallOpenCodeCommandToDir(commandsDir, packageDir, command string, excludes []string) (InstallStrategy, error) {
	name, srcPath, isDir, err := locateCommandCompatSource(packageDir, command)
	if err != nil {
		return StrategyUnknown, err
	}
	if isDir {
		return installOpenCodeCommandBundleToDir(commandsDir, srcPath, name, excludes)
	}
	return installOpenCodeCommandFileToDir(commandsDir, srcPath, name)
}

// InstallOpenCodeCommandCollectionToDir installs all compatible commands into OpenCode's native commands directory.
func InstallOpenCodeCommandCollectionToDir(commandsDir, packageDir string, excludes []string) (InstallStrategy, error) {
	names, err := ListCursorCompatibleCommands(packageDir)
	if err != nil {
		return StrategyUnknown, err
	}
	if len(names) == 0 {
		return StrategyUnknown, errors.Newf(errors.CodeNotFound, "commands collection not found: %s", filepath.Join(packageDir, defaultCommandsSubdir))
	}

	for _, name := range names {
		if _, err := InstallOpenCodeCommandToDir(commandsDir, packageDir, name, excludes); err != nil {
			return StrategyUnknown, err
		}
	}
	return StrategyCopy, nil
}

func installOpenCodeCommandFileToDir(commandsDir, srcPath, commandName string) (InstallStrategy, error) {
	dest, err := security.SafeJoin(commandsDir, commandName+".md")
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command destination")
	}
	content, err := readOpenCodeCommandSource(srcPath)
	if err != nil {
		return StrategyUnknown, err
	}
	return StrategyCopy, writeIfChanged(dest, content)
}

func installOpenCodeCommandBundleToDir(commandsDir, srcDir, commandName string, excludes []string) (InstallStrategy, error) {
	destRoot, err := security.SafeJoin(commandsDir, commandName)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command destination")
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return StrategyUnknown, err
	}

	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
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
		if shouldExcludeCompat(rel, excludes) {
			return nil
		}
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path in command bundle")
		}

		destRel := rel
		var content []byte
		switch {
		case strings.HasSuffix(rel, ".command.mdc"):
			destRel = strings.TrimSuffix(rel, ".command.mdc") + ".md"
			content, err = readOpenCodeCommandSource(path)
		case strings.HasSuffix(rel, ".md"):
			content, err = os.ReadFile(path)
		default:
			content, err = os.ReadFile(path)
		}
		if err != nil {
			return err
		}

		if err := security.ValidatePath(destRel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid destination path in command bundle")
		}
		return writeFileWithDirs(filepath.Join(destRoot, destRel), content, 0o644)
	})
	if err != nil {
		return StrategyUnknown, err
	}
	return StrategyCopy, nil
}

func readOpenCodeCommandSource(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(path, ".mdc") {
		return data, nil
	}
	_, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return []byte{}, nil
	}
	return []byte(body + "\n"), nil
}
