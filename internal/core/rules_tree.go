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

// RulesTree represents the package rules directory contents in a structured form.
type RulesTree struct {
	PackageDir string
	Presets    []string
	Packages   []RulesPackage
}

// RulesPackage captures a package name and the rule files found within it.
type RulesPackage struct {
	Name  string
	Files []string
}

// BuildRulesTree walks the configured packageDir and returns a structured view of
// root-level presets and package contents. Missing directories are handled
// gracefully with empty slices.
func BuildRulesTree(packageDir string) (*RulesTree, error) {
	tree := &RulesTree{PackageDir: packageDir}

	info, err := os.Stat(packageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tree, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return tree, errors.Newf(errors.CodeFailedPrecondition, "package dir is not a directory: %s", packageDir)
	}

	presets, err := listRootRuleFiles(packageDir)
	if err != nil {
		return nil, err
	}
	sort.Strings(presets)
	tree.Presets = presets

	pkgNames, err := ListPackageDirs(packageDir)
	if err != nil {
		return nil, err
	}
	sort.Strings(pkgNames)

	for _, name := range pkgNames {
		// Validate package names to avoid traversing unexpected paths.
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		files, err := collectPackageRuleFiles(packageDir, name)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)
		tree.Packages = append(tree.Packages, RulesPackage{Name: name, Files: files})
	}

	return tree, nil
}

func listRootRuleFiles(packageDir string) ([]string, error) {
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".mdc" || ext == ".md" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func collectPackageRuleFiles(packageDir, packageName string) ([]string, error) {
	pkgDir := filepath.Join(packageDir, packageName)
	info, err := os.Stat(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return []string{}, nil
	}

	ignorePatterns, err := readPackageIgnorePatterns(pkgDir, ".cursor-rules-ignore")
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mdc" && ext != ".md" {
			return nil
		}

		rel, relErr := filepath.Rel(pkgDir, path)
		if relErr != nil {
			return relErr
		}
		if err := security.ValidatePath(rel); err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid file path in package %s", packageName)
		}

		for _, pat := range ignorePatterns {
			matched, matchErr := filepath.Match(pat, rel)
			if matchErr == nil && matched {
				return nil
			}
		}

		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readPackageIgnorePatterns(pkgDir, ignoreFileName string) ([]string, error) {
	ignorePath := filepath.Join(pkgDir, ignoreFileName)
	b, err := os.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(b), "\n")
	patterns := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}
