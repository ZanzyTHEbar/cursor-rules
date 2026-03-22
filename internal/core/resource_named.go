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
)

func listNamedFileResources(root, ext string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ext {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ext)
		if err := security.ValidatePackageName(base); err != nil {
			continue
		}
		names = append(names, base)
	}
	sort.Strings(names)
	return names, nil
}

func listNamedDirResources(root, sentinel string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		sentinelPath := filepath.Join(root, name, sentinel)
		if _, err := os.Stat(sentinelPath); err == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func installNamedFileResource(projectRoot, sourceRoot, destSubdir, name, ext string) (InstallStrategy, error) {
	destDir, err := security.SafeJoin(projectRoot, ".cursor", destSubdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid destination path")
	}
	return installNamedFileResourceTo(destDir, sourceRoot, name, ext)
}

func installNamedFileResourceTo(destDir, sourceRoot, name, ext string) (InstallStrategy, error) {
	if err := security.ValidatePackageName(name); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource name")
	}
	src, err := security.SafeJoin(sourceRoot, name+ext)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid source path")
	}
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeNotFound, "resource not found: %s", name)
		}
		return StrategyUnknown, err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return StrategyUnknown, err
	}
	dest, err := security.SafeJoin(destDir, name+ext)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource destination")
	}
	return ApplySourceToDest(sourceRoot, src, dest, name)
}

func installNamedDirectoryResourceTo(destParent, sourceRoot, name, sentinel string) (InstallStrategy, error) {
	if err := security.ValidatePackageName(name); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource name")
	}
	srcDir, err := security.SafeJoin(sourceRoot, name)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid source path")
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeNotFound, "resource not found: %s", name)
		}
		return StrategyUnknown, err
	}
	if !info.IsDir() {
		return StrategyUnknown, errors.Newf(errors.CodeFailedPrecondition, "resource path is not a directory: %s", srcDir)
	}
	sentinelPath := filepath.Join(srcDir, sentinel)
	if _, err := os.Stat(sentinelPath); err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeFailedPrecondition, "resource missing %s: %s", sentinel, name)
		}
		return StrategyUnknown, err
	}
	destRoot, err := security.SafeJoin(destParent, name)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid destination path")
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return StrategyUnknown, err
	}
	if WantGNUStow() && HasStow() {
		if err := os.MkdirAll(destParent, 0o755); err != nil {
			return StrategyUnknown, err
		}
		cmd := exec.Command("stow", "-v", "-d", sourceRoot, "-t", destParent, name)
		if out, cmdErr := cmd.CombinedOutput(); cmdErr == nil {
			_ = out
			return StrategyStow, nil
		}
	}

	strategy := StrategyCopy
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
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid file path in resource %s", name)
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

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		perm := info.Mode().Perm()
		if perm == 0 {
			perm = 0o600
		}
		return os.WriteFile(dest, data, perm)
	})
	if err != nil {
		return StrategyUnknown, err
	}
	return strategy, nil
}

func removeInstalledNamedFileResourceFrom(destDir, name, ext string) (bool, error) {
	if err := security.ValidatePackageName(name); err != nil {
		return false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource name")
	}
	target, err := security.SafeJoin(destDir, name+ext)
	if err != nil {
		return false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid installed resource file path")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, os.Remove(target)
}

func removeInstalledNamedDirResourceFrom(destDir, name string) (bool, error) {
	if err := security.ValidatePackageName(name); err != nil {
		return false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource name")
	}
	target, err := security.SafeJoin(destDir, name)
	if err != nil {
		return false, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid installed resource directory path")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, os.RemoveAll(target)
}
