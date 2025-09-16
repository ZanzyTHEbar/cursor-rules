package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
			matched, _ := filepath.Match(pat, rel)
			if matched {
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

		// If symlink requested, try to create symlink
		if UseSymlink() {
			if err := CreateSymlink(path, dest); err == nil {
				return nil
			}
			// else fallthrough to copy
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
