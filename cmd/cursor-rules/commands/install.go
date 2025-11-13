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
			presetName := args[0]

			// Get workdir
			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, err := cmd.Root().Flags().GetString("workdir")
				if err != nil {
					return err
				}
				if w == "" {
					var absErr error
					w, absErr = filepath.Abs(".")
					if absErr != nil {
						return fmt.Errorf("failed to get absolute path: %w", absErr)
					}
				}
				wd = w
			}

			// Get shared directory
			shared := core.DefaultSharedDir()
			pkgPath := filepath.Join(shared, presetName)

			// Check if it's a directory (package) or single file
			info, err := os.Stat(pkgPath)
			isPackage := err == nil && info.IsDir()

			// Load manifest if exists
			var m *manifest.Manifest
			if isPackage {
				m, err = manifest.Load(pkgPath)
				if err != nil {
					// Manifest load errors are non-fatal; proceed without it
					m = nil
				}
			}

			// Determine targets
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

			// Install to each target
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
					// Single file install
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
		},
	}

	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package directory structure")
	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt")
	cmd.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")

	return cmd
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
