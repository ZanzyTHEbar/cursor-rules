package commands

import "github.com/ZanzyTHEbar/cursor-rules/internal/cli"

// RegisterAll registers all command factories into the global palette.
func RegisterAll() {
	cli.Register(
		NewInstallCmd,
		NewRemoveCmd,
		NewSyncCmd,
		NewWatchCmd,
		NewListCmd,
		NewEffectiveCmd,
		NewPolicyCmd,
		NewInitCmd,
		NewTransformCmd,
		NewConfigCmd,
		NewInfoCmd,
	)
}
