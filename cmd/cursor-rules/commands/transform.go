package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
	"github.com/spf13/cobra"
)

// NewTransformCmd returns a command for previewing transformations.
func NewTransformCmd(ctx *cli.AppContext) *cobra.Command {
	var targetFlag string

	cmd := &cobra.Command{
		Use:   "transform <preset>",
		Short: "Preview frontmatter transformation for a preset",
		Long: `Dry-run transformation to see how Cursor rules will be converted
to Copilot format without writing files.

Example:
  cursor-rules transform frontend --target copilot-instr`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			preset := args[0]
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}

			transformer, err := ctx.Transformer(targetFlag)
			if err != nil {
				return err
			}

			shared := core.DefaultSharedDir()
			pkgPath := filepath.Join(shared, preset)

			// Check if it's a directory or single file
			info, err := os.Stat(pkgPath)
			if err != nil {
				// Try with .mdc extension
				pkgPath += ".mdc"
				info, err = os.Stat(pkgPath)
				if err != nil {
					return fmt.Errorf("preset not found: %s", preset)
				}
			}

			if ui != nil {
				ui.Info("Transforming %q to %s format:\n\n", preset, transformer.Target())
			} else {
				fmt.Printf("Transforming %q to %s format:\n\n", preset, transformer.Target())
			}

			if info.IsDir() {
				// Walk directory
				return filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
						return err
					}
					return previewTransform(path, transformer, ui)
				})
			}

			// Single file
			return previewTransform(pkgPath, transformer, ui)
		},
	}

	cmd.Flags().StringVar(&targetFlag, "target", "copilot-instr", "target format")

	return cmd
}

// previewTransform reads, transforms, and displays a single file transformation.
func previewTransform(path string, transformer transform.Transformer, ui cli.Messenger) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fm, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		printError(ui, "❌ %s: %v\n", filepath.Base(path), err)
		return nil
	}

	transformedFM, transformedBody, err := transformer.Transform(fm, body)
	if err != nil {
		printError(ui, "❌ %s: %v\n", filepath.Base(path), err)
		return nil
	}

	if validateErr := transformer.Validate(transformedFM); validateErr != nil {
		printWarn(ui, "⚠️  %s: validation warning: %v\n", filepath.Base(path), validateErr)
	}

	output, err := transform.MarshalMarkdown(transformedFM, transformedBody)
	if err != nil {
		printError(ui, "❌ %s: marshal error: %v\n", filepath.Base(path), err)
		return nil
	}

	baseName := strings.TrimSuffix(filepath.Base(path), ".mdc")
	printInfo(ui, "📄 %s.mdc → %s%s\n", baseName, baseName, transformer.Extension())
	printInfo(ui, "%s\n", string(output))
	printInfo(ui, "---\n")

	return nil
}

func printInfo(ui cli.Messenger, format string, args ...interface{}) {
	if ui != nil {
		ui.Info(format, args...)
		return
	}
	fmt.Printf(format, args...)
}

func printWarn(ui cli.Messenger, format string, args ...interface{}) {
	if ui != nil {
		ui.Warn(format, args...)
		return
	}
	fmt.Printf(format, args...)
}

func printError(ui cli.Messenger, format string, args ...interface{}) {
	if ui != nil {
		ui.Error(format, args...)
		return
	}
	fmt.Printf(format, args...)
}
