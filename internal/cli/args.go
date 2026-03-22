package cli

import (
	"slices"

	"github.com/spf13/cobra"
)

// ReservedPositionalHelpArgs lists positional argument values that request command
// help instead of being passed to the command logic. Used by commands that accept
// an optional [name] (e.g. install, remove) so that "cursor-rules install help"
// shows usage instead of attempting to install a preset named "help".
var ReservedPositionalHelpArgs = []string{"help"}

// IsReservedHelpArg reports whether s is a reserved positional that means "show help".
func IsReservedHelpArg(s string) bool {
	return slices.Contains(ReservedPositionalHelpArgs, s)
}

// ShowHelpIfReservedArg runs the command's help when the first positional argument
// is a reserved help token and returns true so the caller can exit RunE without
// running command logic (e.g. return nil). Otherwise it returns false and the
// caller proceeds. Callers should use it at the start of RunE:
//
//	if cli.ShowHelpIfReservedArg(cmd, args) {
//		return nil
//	}
//
// This avoids returning an error so Cobra does not print "Error: ..." after the help.
func ShowHelpIfReservedArg(cmd *cobra.Command, args []string) bool {
	if len(args) == 0 || !IsReservedHelpArg(args[0]) {
		return false
	}
	if err := cmd.Help(); err != nil {
		return false
	}
	return true
}
