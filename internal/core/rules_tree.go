package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

// RulesTree represents the shared rules directory contents in a structured form.
type RulesTree struct {
	SharedDir string
	Presets   []string
	Packages  []RulesPackage
}

// RulesPackage captures a package name and the rule files found within it.
type RulesPackage struct {
	Name  string
	Files []string
}

// BuildRulesTree walks the configured sharedDir and returns a structured view of
// root-level presets and package contents. Missing directories are handled
// gracefully with empty slices.
func BuildRulesTree(sharedDir string) (*RulesTree, error) {
	tree := &RulesTree{SharedDir: sharedDir}

	info, err := os.Stat(sharedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tree, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return tree, fmt.Errorf("shared dir is not a directory: %s", sharedDir)
	}

	presets, err := listRootRuleFiles(sharedDir)
	if err != nil {
		return nil, err
	}
	sort.Strings(presets)
	tree.Presets = presets

	pkgNames, err := ListSharedPackages(sharedDir)
	if err != nil {
		return nil, err
	}
	sort.Strings(pkgNames)

	for _, name := range pkgNames {
		// Validate package names to avoid traversing unexpected paths.
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		files, err := collectPackageRuleFiles(sharedDir, name)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)
		tree.Packages = append(tree.Packages, RulesPackage{Name: name, Files: files})
	}

	return tree, nil
}

// FormatRulesTree renders the rules tree as a simple text tree with grouping.
func FormatRulesTree(tree *RulesTree) string {
	if tree == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "shared dir: %s\n", tree.SharedDir)

	// Presets section
	b.WriteString("presets:\n")
	if len(tree.Presets) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for i, p := range tree.Presets {
			fmt.Fprintf(&b, "  %s %s\n", branchPrefix(i, len(tree.Presets)), p)
		}
	}

	// Packages section
	b.WriteString("packages:\n")
	if len(tree.Packages) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for pkgIdx, pkg := range tree.Packages {
			fmt.Fprintf(&b, "  %s %s/\n", branchPrefix(pkgIdx, len(tree.Packages)), pkg.Name)

			parentIndent := "  │"
			if pkgIdx == len(tree.Packages)-1 {
				parentIndent = "    "
			}

			if len(pkg.Files) == 0 {
				fmt.Fprintf(&b, "%s  (no rule files)\n", parentIndent)
				continue
			}
			for fileIdx, f := range pkg.Files {
				fmt.Fprintf(&b, "%s  %s %s\n", parentIndent, branchPrefix(fileIdx, len(pkg.Files)), f)
			}
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func listRootRuleFiles(sharedDir string) ([]string, error) {
	entries, err := os.ReadDir(sharedDir)
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

func collectPackageRuleFiles(sharedDir, packageName string) ([]string, error) {
	pkgDir := filepath.Join(sharedDir, packageName)
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
			return fmt.Errorf("invalid file path in package %s: %w", packageName, err)
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

func branchPrefix(idx, total int) string {
	if total == 0 {
		return ""
	}
	if idx == total-1 {
		return "└─"
	}
	return "├─"
}
