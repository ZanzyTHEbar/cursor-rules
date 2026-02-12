package cli

import (
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
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
	var logLevelFlag string
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (optional)")
	root.PersistentFlags().StringVar(&logLevelFlag, "log-level", "", "log level (debug, info, warn, error)")

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if ctx.Viper == nil {
			ctx.Viper = viper.New()
		}
		if cfgFile != "" {
			ctx.Viper.SetConfigFile(cfgFile)
		} else {
			configDir := config.DefaultConfigDir()
			if configDir != "" {
				ctx.Viper.AddConfigPath(configDir)
			}
			ctx.Viper.SetConfigName("config")
			ctx.Viper.SetConfigType("yaml")
		}
		ctx.Viper.SetDefault("logLevel", "info")
		if err := ctx.Viper.BindEnv("logLevel", "CURSOR_RULES_LOG_LEVEL"); err != nil {
			return errors.Wrapf(err, errors.CodeInternal, "binding log level env")
		}
		if flag := cmd.PersistentFlags().Lookup("log-level"); flag != nil {
			if err := ctx.Viper.BindPFlag("logLevel", flag); err != nil {
				return errors.Wrapf(err, errors.CodeInternal, "binding log level flag")
			}
		}
		ctx.Viper.AutomaticEnv()
		if err := ctx.Viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return errors.Wrapf(err, errors.CodeInternal, "reading config")
			}
		}
		ctx.SetMessenger(NewMessenger(cmd.OutOrStdout(), cmd.ErrOrStderr(), ctx.Viper.GetString("logLevel")))
		cfgUsed := ctx.Viper.ConfigFileUsed()
		if cfgUsed == "" {
			cfgUsed = "(none)"
		}
		if ui := ctx.Messenger(); ui != nil {
			ui.Debug("Using config file: %s\n", cfgUsed)
		}

		if ctx.Viper.GetBool("enableStow") {
			if core.HasStow() {
				if err := os.Setenv("CURSOR_RULES_USE_GNUSTOW", "1"); err != nil {
					if ui := ctx.Messenger(); ui != nil {
						ui.Warn("failed to enable GNU stow: %v\n", err)
					}
				} else {
					if ui := ctx.Messenger(); ui != nil {
						ui.Debug("GNU stow enabled via config\n")
					}
				}
			} else {
				if ui := ctx.Messenger(); ui != nil {
					ui.Warn("enableStow is true but 'stow' binary not found on PATH\n")
				}
			}
		}
		if postInit != nil {
			if err := postInit(ctx.Viper); err != nil {
				if ui := ctx.Messenger(); ui != nil {
					ui.Error("postInit error: %v\n", err)
				}
				return err
			}
		}
		return nil
	}
}
