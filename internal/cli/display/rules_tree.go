package display

import (
	"fmt"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// FormatRulesTree renders the rules tree as a simple text tree with grouping.
func FormatRulesTree(tree *core.RulesTree) string {
	if tree == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "package dir: %s\n", tree.PackageDir)

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

func branchPrefix(idx, total int) string {
	if total == 0 {
		return ""
	}
	if idx == total-1 {
		return "└─"
	}
	return "├─"
}
