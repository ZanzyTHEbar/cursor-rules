package cli

import "github.com/spf13/cobra"

// GetOptionalFlag retrieves a flag value from command, inherited flags, or root persistent flags.
func GetOptionalFlag(cmd *cobra.Command, name string) string {
	if cmd == nil {
		return ""
	}
	if flag := cmd.Flags().Lookup(name); flag != nil && flag.Value.String() != "" {
		return flag.Value.String()
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil && flag.Value.String() != "" {
		return flag.Value.String()
	}
	if root := cmd.Root(); root != nil {
		if flag := root.PersistentFlags().Lookup(name); flag != nil && flag.Value.String() != "" {
			return flag.Value.String()
		}
	}
	return ""
}
