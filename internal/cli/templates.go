package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// CommandGroup describes a named group of commands for templates.
type CommandGroup struct {
	Name     string
	Commands []*cobra.Command
}

// CategorizeCommands groups commands into core and utility sections.
func CategorizeCommands(cmds []*cobra.Command) []CommandGroup {
	coreNames := map[string]struct{}{
		"install":   {},
		"remove":    {},
		"sync":      {},
		"list":      {},
		"effective": {},
	}

	var coreCmds []*cobra.Command
	var utilityCmds []*cobra.Command
	for _, cmd := range cmds {
		if !cmd.IsAvailableCommand() || cmd.Name() == "help" {
			continue
		}
		if _, ok := coreNames[cmd.Name()]; ok {
			coreCmds = append(coreCmds, cmd)
		} else {
			utilityCmds = append(utilityCmds, cmd)
		}
	}

	sort.Slice(coreCmds, func(i, j int) bool {
		return coreCmds[i].Name() < coreCmds[j].Name()
	})
	sort.Slice(utilityCmds, func(i, j int) bool {
		return utilityCmds[i].Name() < utilityCmds[j].Name()
	})

	var groups []CommandGroup
	if len(coreCmds) > 0 {
		groups = append(groups, CommandGroup{Name: "Core commands", Commands: coreCmds})
	}
	if len(utilityCmds) > 0 {
		groups = append(groups, CommandGroup{Name: "Utilities", Commands: utilityCmds})
	}
	return groups
}

// Rpad right pads a string to the given length.
func Rpad(s string, padding int) string {
	format := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(format, s)
}

// TrimTrailingWhitespaces removes trailing whitespace from a string.
func TrimTrailingWhitespaces(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}
