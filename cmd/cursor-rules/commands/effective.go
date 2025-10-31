package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewEffectiveCmd returns the effective command with multi-target support.
func NewEffectiveCmd(ctx *cli.AppContext) *cobra.Command {
	var targetFlag string

	cmd := &cobra.Command{
		Use:   "effective",
		Short: "Show effective merged rules for current workspace",
		Long: `Display the merged rules that would be active in the current workspace.
For Copilot targets, simulates the non-deterministic merge order.

Examples:
  # Show Cursor rules
  cursor-rules effective
  
  # Show Copilot instructions
  cursor-rules effective --target copilot-instr`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get workdir
			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, err := cmd.Root().Flags().GetString("workdir")
				if err != nil {
					return fmt.Errorf("failed to get workdir flag: %w", err)
				}
				if w == "" {
					var err error
					w, err = filepath.Abs(".")
					if err != nil {
						return fmt.Errorf("failed to get absolute path: %w", err)
					}
				}
				wd = w
			}

			// For cursor target, use existing implementation
			if targetFlag == "cursor" {
				out, err := core.EffectiveRules(wd)
				if err != nil {
					return err
				}
				fmt.Println(out)
				return nil
			}

			// For Copilot targets, show merged files
			transformer, err := ctx.Transformer(targetFlag)
			if err != nil {
				return err
			}

			rulesDir := filepath.Join(wd, transformer.OutputDir())

			fmt.Printf("# Effective Rules (%s)\n\n", transformer.Target())
			fmt.Printf("Source: %s\n\n", rulesDir)

			// Check if directory exists
			if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
				fmt.Printf("No rules found in %s\n", rulesDir)
				return nil
			}

			// Collect all rule files
			var files []string
			if err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.HasSuffix(path, transformer.Extension()) {
					files = append(files, path)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("failed to walk rules directory: %w", err)
			}

			if len(files) == 0 {
				fmt.Printf("No %s files found in %s\n", transformer.Extension(), rulesDir)
				return nil
			}

			// For Copilot, simulate merge order (alphabetical for determinism in preview)
			sort.Strings(files)

			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					continue
				}
				fmt.Printf("## %s\n\n", filepath.Base(file))
				fmt.Println(string(data))
				fmt.Println("\n---")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "target format to show: cursor|copilot-instr|copilot-prompt")

	return cmd
}
