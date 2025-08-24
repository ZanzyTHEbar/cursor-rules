package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigureRoot attaches common flags and a PersistentPreRunE to an existing
// root command to initialize the provided AppContext.Viper instance. If
// postInit is non-nil it will be invoked after the config has been read.
func ConfigureRoot(root *cobra.Command, ctx *AppContext, postInit func(*viper.Viper) error) {
	if root == nil {
		return
	}
	if ctx == nil {
		ctx = NewAppContext(nil, nil)
	}

	var cfgFile string
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (optional)")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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
			} else {
				return fmt.Errorf("reading config: %w", err)
			}
		}
		ctx.Logger.Printf("Using config file: %s", ctx.Viper.ConfigFileUsed())
		if postInit != nil {
			if err := postInit(ctx.Viper); err != nil {
				// log and return error
				ctx.Logger.Printf("postInit error: %v", err)
				return err
			}
		}
		return nil
	}
}
