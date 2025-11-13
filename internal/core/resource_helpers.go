package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// DefaultSharedDirWithEnv returns a shared dir path based on an env var or a default name under the user's home.
func DefaultSharedDirWithEnv(envVar, defaultName string) string {
	// We prefer using the main DefaultSharedDir for cursor-rules.
	if envVar == "CURSOR_RULES_DIR" || envVar == "CURSOR_COMMANDS_DIR" {
		// If explicitly set, respect it; otherwise fall back to DefaultSharedDir
		if v := os.Getenv(envVar); v != "" {
			return v
		}
		return DefaultSharedDir()
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if env := os.Getenv("HOME"); env != "" {
			return filepath.Join(env, defaultName)
		}
		if cwd, cwdErr := os.Getwd(); cwdErr == nil && cwd != "" {
			return filepath.Join(cwd, defaultName)
		}
		return defaultName
	}
	return filepath.Join(home, defaultName)
}

// InstallPackageGeneric installs an entire package directory from sharedDir into the project's
// destRoot (.cursor/<subdir>). It supports excludes and a package-level ignore file name.
// exts is a list of allowed file extensions (e.g. []string{".mdc",".md"}).
func InstallPackageGeneric(projectRoot, sharedDir, packageName, destSubdir string, exts []string, ignoreFileName string, excludes []string, noFlatten bool) error {
	pkgDir := filepath.Join(sharedDir, packageName)
	info, err := os.Stat(pkgDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("package not found: %s", pkgDir)
	}

	// Read ignore file if present in package dir
	ignorePath := filepath.Join(pkgDir, ignoreFileName)
	var ignorePatterns []string
	if b, err := os.ReadFile(ignorePath); err == nil {
		lines := strings.Split(string(b), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			ignorePatterns = append(ignorePatterns, l)
		}
	}

	// Merge excludes param into ignorePatterns
	for _, ex := range excludes {
		ex = strings.TrimSpace(ex)
		if ex != "" {
			ignorePatterns = append(ignorePatterns, ex)
		}
	}

	// Walk package dir and install allowed files unless excluded
	destRoot := filepath.Join(projectRoot, ".cursor", destSubdir)
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return err
	}

	err = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		okExt := false
		for _, ex := range exts {
			if filepath.Ext(path) == ex {
				okExt = true
				break
			}
		}
		if !okExt {
			return nil
		}

		rel, err := filepath.Rel(pkgDir, path)
		if err != nil {
			return err
		}

		// Check ignore patterns (simple glob match)
		for _, pat := range ignorePatterns {
			matched, matchErr := filepath.Match(pat, rel)
			if matchErr == nil && matched {
				return nil
			}
		}

		// Destination path
		var dest string
		if !noFlatten || strings.Contains(packageName, "/") {
			dest = filepath.Join(destRoot, filepath.Base(rel))
		} else {
			destName := filepath.Join(packageName, rel)
			dest = filepath.Join(destRoot, destName)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		// Delegate applying source to dest (stow/symlink or stub)
		_, applyErr := ApplySourceToDest(sharedDir, path, dest, packageName)
		return applyErr
	})
	if err != nil {
		return err
	}
	return nil
}

// AtomicWriteString writes content to dest atomically using a temp file in tmpDir.
func AtomicWriteString(tmpDir, dest, content string, perm os.FileMode) error {
	tmp, err := os.CreateTemp(tmpDir, ".stub-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// ensure cleanup on error
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close() // #nosec G104 - error path, already returning err
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close() // #nosec G104 - error path, already returning err
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return err
	}
	return os.Chmod(dest, perm)
}

// ApplySourceToDest attempts to apply a source file to dest using stow (packageName),
// then symlink, and finally falls back to writing a stub that references the source.
func ApplySourceToDest(sharedDir, src, dest, packageName string) (InstallStrategy, error) {
	destDir := filepath.Dir(dest)
	// Try GNU stow if requested
	if WantGNUStow() && HasStow() {
		// #nosec G204 - sharedDir, destDir, and packageName are validated before this call
		cmd := exec.Command("stow", "-v", "-d", sharedDir, "-t", destDir, packageName)
		if out, err := cmd.CombinedOutput(); err == nil {
			_ = out
			return StrategyStow, nil
		}
		// else fallthrough
	}
	// Try symlink if requested
	if UseSymlink() || WantGNUStow() {
		if err := CreateSymlink(src, dest); err == nil {
			return StrategySymlink, nil
		}
		// else fallthrough to stub write
	}
	// Default: write stub referencing source
	content := "---\n@file " + src + "\n"
	if err := AtomicWriteString(destDir, dest, content, 0o644); err != nil {
		return StrategyUnknown, err
	}
	return StrategyCopy, nil
}

// AtomicWriteTemplate renders tmpl with data into a temp file in tmpDir and
// atomically renames it to dest with the given permission.
func AtomicWriteTemplate(tmpDir, dest string, tmpl *template.Template, data interface{}, perm os.FileMode) error {
	tmp, err := os.CreateTemp(tmpDir, ".stub-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// ensure cleanup on error
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if err := tmpl.Execute(tmp, data); err != nil {
		_ = tmp.Close() // #nosec G104 - error path, already returning err
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close() // #nosec G104 - error path, already returning err
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return err
	}
	return os.Chmod(dest, perm)
}
