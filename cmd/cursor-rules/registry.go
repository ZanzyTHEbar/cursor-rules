package main

import (
	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/cmd/cursor-rules/commands"
)

// Register commands into the global CLI palette
func init() {
	// register all command factories into the global palette
	cli.Register(
		commands.NewInstallCmd,
		commands.NewRemoveCmd,
		commands.NewSyncCmd,
		commands.NewWatchCmd,
		commands.NewListCmd,
		commands.NewEffectiveCmd,
		commands.NewPolicyCmd,
		commands.NewInitCmd,
	)
}

// NOTE: registration lives in init above
