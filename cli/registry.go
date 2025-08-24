package cli

import "github.com/spf13/cobra"

// Simple global palette registry for incremental migration.
var DefaultPalette Palette

// Register registers command factories into the global palette.
func Register(factories ...CommandFactory) {
	DefaultPalette.Register(factories...)
}

// BuildRoot builds a root command using the global palette and provided context.
func BuildRoot(ctx *AppContext) *cobra.Command {
	return NewRoot(ctx, DefaultPalette)
}
