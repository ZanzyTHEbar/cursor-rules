package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewInstallCmd returns the install command. Accepts AppContext for future use.
func NewInstallCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var flattenFlag bool
	cmd := &cobra.Command{
		Use:   "install <preset|package>",
		Short: "Install a preset or package into the current project (.cursor/rules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			// prefer workdir from AppContext.Viper, fallback to flag
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
					w, _ = filepath.Abs(".")
				}
				wd = w
			}

			// Check if target is a package directory in shared dir
			shared := core.DefaultSharedDir()
			pkgPath := filepath.Join(shared, target)
			if info, err := os.Stat(pkgPath); err == nil && info.IsDir() {
				// treat as package install
				if err := core.InstallPackage(wd, target, excludeFlag, flattenFlag); err != nil {
					return fmt.Errorf("package install failed: %w", err)
				}
				fmt.Printf("Installed package %q into %s/.cursor/rules/\n", target, wd)
				return nil
			}

			// fallback to single preset install
			if err := core.InstallPreset(wd, target); err != nil {
				return fmt.Errorf("install failed: %w", err)
			}
			fmt.Printf("Installed preset %q into %s/.cursor/rules/\n", target, wd)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	cmd.Flags().BoolVar(&flattenFlag, "flatten", false, "flatten package files into .cursor/rules/ instead of preserving package subpaths")
	return cmd
}
