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
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]

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

			fmt.Printf("Transforming %q to %s format:\n\n", preset, transformer.Target())

			if info.IsDir() {
				// Walk directory
				return filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
						return err
					}
					return previewTransform(path, transformer)
				})
			}

			// Single file
			return previewTransform(pkgPath, transformer)
		},
	}

	cmd.Flags().StringVar(&targetFlag, "target", "copilot-instr", "target format")

	return cmd
}

// previewTransform reads, transforms, and displays a single file transformation.
func previewTransform(path string, transformer transform.Transformer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fm, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		fmt.Printf("❌ %s: %v\n", filepath.Base(path), err)
		return nil
	}

	transformedFM, transformedBody, err := transformer.Transform(fm, body)
	if err != nil {
		fmt.Printf("❌ %s: %v\n", filepath.Base(path), err)
		return nil
	}

	if err := transformer.Validate(transformedFM); err != nil {
		fmt.Printf("⚠️  %s: validation warning: %v\n", filepath.Base(path), err)
	}

	output, err := transform.MarshalMarkdown(transformedFM, transformedBody)
	if err != nil {
		fmt.Printf("❌ %s: marshal error: %v\n", filepath.Base(path), err)
		return nil
	}

	baseName := strings.TrimSuffix(filepath.Base(path), ".mdc")
	fmt.Printf("📄 %s.mdc → %s%s\n", baseName, baseName, transformer.Extension())
	fmt.Println(string(output))
	fmt.Println("---")

	return nil
}
