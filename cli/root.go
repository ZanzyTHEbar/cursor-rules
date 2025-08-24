package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRoot constructs the root command and composes the provided palette using
// the given AppContext. Config initialization is performed in
// PersistentPreRunE so construction is side-effect free and testable.
func NewRoot(ctx *AppContext, p Palette) *cobra.Command {
	if ctx == nil {
		ctx = NewAppContext(nil, nil)
	}
	if p == nil {
		p = Palette{}
	}

	var cfgFile string

	root := &cobra.Command{
		Use:   "alpha-engine",
		Short: "Alpha Engine CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx.Viper == nil {
				ctx.Viper = viper.New()
			}
			if cfgFile != "" {
				ctx.Viper.SetConfigFile(cfgFile)
			}
			ctx.Viper.AutomaticEnv()
			if err := ctx.Viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					// no config is fine
					return nil
				}
				return fmt.Errorf("reading config: %w", err)
			}
			ctx.Logger.Printf("Using config file: %s", ctx.Viper.ConfigFileUsed())
			return nil
		},
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file path (optional)")

	// compose commands from palette
	root.AddCommand(p.Commands(ctx)...)
	return root
}
