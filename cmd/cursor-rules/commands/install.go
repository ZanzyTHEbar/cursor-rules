package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/manifest"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
	"github.com/spf13/cobra"
)

// NewInstallCmd returns the install command with transformer support.
func NewInstallCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool
	var targetFlag string
	var allTargetsFlag bool

	var excludeAllFlag []string
	var noFlattenAllFlag bool
	var targetAllFlag string
	var allTargetsAllFlag bool

	cmd := &cobra.Command{
		Use:   "install <preset|package>",
		Short: "Install a preset or package into the current project",
		Long: `Install a preset or package into .cursor/rules/ (default) or 
.github/instructions/ or .github/prompts/ depending on --target flag.

Examples:
  # Install to Cursor (default)
  cursor-rules install frontend
  
  # Install to Copilot Instructions
  cursor-rules install frontend --target copilot-instr
  
  # Install to Copilot Prompts
  cursor-rules install frontend --target copilot-prompt
  
  # Install to all targets defined in manifest
  cursor-rules install frontend --all-targets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(ctx, cmd, args[0], "", excludeFlag, noFlattenFlag, targetFlag, allTargetsFlag)
		},
	}

	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package directory structure")
	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt")
	cmd.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")

	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Install all packages from the shared directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			shared := core.DefaultSharedDir()
			pkgs, err := core.ListSharedPackages(shared)
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}
			if len(pkgs) == 0 {
				if ui := ctx.Messenger(); ui != nil {
					ui.Info("No packages found in %s\n", shared)
				}
				return nil
			}
			for _, name := range pkgs {
				if err := runInstall(ctx, cmd, name, shared, excludeAllFlag, noFlattenAllFlag, targetAllFlag, allTargetsAllFlag); err != nil {
					return err
				}
			}
			return nil
		},
	}

	allCmd.Flags().StringArrayVar(&excludeAllFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	allCmd.Flags().BoolVarP(&noFlattenAllFlag, "no-flatten", "n", false, "preserve package directory structure")
	allCmd.Flags().StringVar(&targetAllFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt")
	allCmd.Flags().BoolVar(&allTargetsAllFlag, "all-targets", false, "install to all targets in manifest")

	cmd.AddCommand(allCmd)

	return cmd
}

func runInstall(ctx *cli.AppContext, cmd *cobra.Command, presetName, sharedDir string, excludeFlag []string, noFlattenFlag bool, targetFlag string, allTargetsFlag bool) error {
	wd, err := resolveWorkdir(ctx, cmd)
	if err != nil {
		return err
	}

	shared := sharedDir
	if shared == "" {
		shared = core.DefaultSharedDir()
	}
	pkgPath := filepath.Join(shared, presetName)

	info, statErr := os.Stat(pkgPath)
	isPackage := statErr == nil && info.IsDir()

	var m *manifest.Manifest
	if isPackage {
		m, err = manifest.Load(pkgPath)
		if err != nil {
			// Manifest load errors are non-fatal; proceed without it
			m = nil
		}
	}

	var targets []string
	if allTargetsFlag && m != nil && len(m.Targets) > 0 {
		targets = m.Targets
	} else {
		targets = []string{targetFlag}
	}

	effectiveExcludes := append([]string{}, excludeFlag...)
	if m != nil && len(m.Exclude) > 0 {
		effectiveExcludes = append(effectiveExcludes, m.Exclude...)
	}

	for _, tgt := range targets {
		transformer, err := ctx.Transformer(tgt)
		if err != nil {
			return err
		}

		var strategy core.InstallStrategy

		if isPackage {
			if transformer.Target() == "cursor" && (core.UseSymlink() || core.WantGNUStow()) {
				strategy, err = core.InstallPackage(wd, presetName, effectiveExcludes, noFlattenFlag)
				if err != nil {
					return fmt.Errorf("install to %s failed: %w", tgt, err)
				}
			} else {
				strategy, err = installPackageWithTransformer(wd, pkgPath, presetName, transformer, effectiveExcludes, noFlattenFlag)
				if err != nil {
					return fmt.Errorf("install to %s failed: %w", tgt, err)
				}
			}
		} else {
			strategy, err = installPresetWithTransformer(wd, pkgPath, presetName, transformer)
			if err != nil {
				return fmt.Errorf("install to %s failed: %w", tgt, err)
			}
		}

		if transformer.Target() == "cursor" {
			if ui := ctx.Messenger(); ui != nil {
				ui.Info("Install method: %s\n", strategy)
			}
		}

		if ui := ctx.Messenger(); ui != nil {
			ui.Success("✅ Installed %q to %s\n", presetName, transformer.OutputDir())
		}
	}

	return nil
}

// installPackageWithTransformer installs a package directory using the specified transformer.
func installPackageWithTransformer(
	workDir, pkgPath, presetName string,
	transformer transform.Transformer,
	excludes []string,
	noFlatten bool,
) (core.InstallStrategy, error) {

	// Create output directory
	outDir := filepath.Join(workDir, transformer.OutputDir())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return core.StrategyUnknown, fmt.Errorf("create output dir for package %q: %w", presetName, err)
	}

	// Walk package directory
	if err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk package %q: %w", presetName, err)
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
			return nil
		}

		// Check exclusions
		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		if shouldExclude(relPath, excludes) {
			return nil
		}

		if err := transformAndWriteFile(path, relPath, outDir, transformer, noFlatten); err != nil {
			return fmt.Errorf("install file from package %q: %w", presetName, err)
		}
		return nil
	}); err != nil {
		return core.StrategyUnknown, err
	}
	return core.StrategyCopy, nil
}

// installPresetWithTransformer installs a single preset file using the specified transformer.
func installPresetWithTransformer(
	workDir, presetPath, presetName string,
	transformer transform.Transformer,
) (core.InstallStrategy, error) {
	// Ensure .mdc extension
	if !strings.HasSuffix(presetPath, ".mdc") {
		presetPath += ".mdc"
	}

	// Check if file exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return core.StrategyUnknown, fmt.Errorf("preset %q not found: %s", presetName, presetPath)
	}

	// For cursor target, respect symlink mode like the old InstallPreset function
	if transformer.Target() == "cursor" {
		sharedDir := core.DefaultSharedDir()
		if core.UseSymlink() || core.WantGNUStow() {
			return core.ApplyPresetWithOptionalSymlink(workDir, presetName, sharedDir)
		}
	}

	// Create output directory
	outDir := filepath.Join(workDir, transformer.OutputDir())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return core.StrategyUnknown, fmt.Errorf("create output dir for preset %q: %w", presetName, err)
	}

	if err := transformAndWriteFile(presetPath, filepath.Base(presetPath), outDir, transformer, false); err != nil {
		return core.StrategyUnknown, fmt.Errorf("install preset %q: %w", presetName, err)
	}
	return core.StrategyCopy, nil
}

// transformAndWriteFile reads, transforms, and writes a single file.
func transformAndWriteFile(
	srcPath, relPath, outDir string,
	transformer transform.Transformer,
	noFlatten bool,
) error {
	// Read file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}

	// Split frontmatter and body
	frontmatter, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", srcPath, err)
	}

	// Transform
	transformedFM, transformedBody, err := transformer.Transform(frontmatter, body)
	if err != nil {
		return fmt.Errorf("transform %s: %w", srcPath, err)
	}

	// Validate
	if validateErr := transformer.Validate(transformedFM); validateErr != nil {
		return fmt.Errorf("validate %s: %w", srcPath, validateErr)
	}

	// Determine output path
	var outPath string
	if noFlatten {
		outPath = filepath.Join(outDir, relPath)
	} else {
		outPath = filepath.Join(outDir, filepath.Base(relPath))
	}
	outPath = strings.TrimSuffix(outPath, ".mdc") + transformer.Extension()

	// Marshal back to file
	output, err := transform.MarshalMarkdown(transformedFM, transformedBody)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", srcPath, err)
	}

	// If we are in "copy" mode but a previous install left a symlink in place,
	// remove it so we don't keep writing through the symlink (which would leave
	// the filesystem state inconsistent with the reported install method).
	if info, statErr := os.Lstat(outPath); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if rmErr := os.Remove(outPath); rmErr != nil {
				return fmt.Errorf("remove existing symlink %s: %w", outPath, rmErr)
			}
		}
	}

	// Idempotent write (hash check)
	existing, readErr := os.ReadFile(outPath)
	if readErr == nil && bytes.Equal(existing, output) {
		return nil // Skip unchanged
	}

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	// #nosec G306 - rule files are meant to be world-readable
	return os.WriteFile(outPath, output, 0o644)
}

// shouldExclude checks if a path matches any exclusion pattern.
func shouldExclude(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}
		// Also check if pattern matches any parent directory
		matched, err = filepath.Match(pattern, filepath.Dir(relPath))
		if err == nil && matched {
			return true
		}
	}
	return false
}
