package cli

import (
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

// GetBoolFlag returns true if the named persistent or inherited flag is set and equals "true".
func GetBoolFlag(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	var f *pflag.Flag
	if flag := cmd.Flags().Lookup(name); flag != nil {
		f = flag
	} else if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		f = flag
	} else if root := cmd.Root(); root != nil {
		f = root.PersistentFlags().Lookup(name)
	}
	if f == nil {
		return false
	}
	return strings.EqualFold(f.Value.String(), "true")
}

// ResolveDestination returns the effective project root and whether the destination is user/global.
// Precedence: --dir user|global → user; --dir <path> → path; --global → user; else --workdir / Viper workdir → path.
func ResolveDestination(a *app.App, cmd *cobra.Command) (workdir string, isUser bool, err error) {
	if a == nil || cmd == nil {
		return "", false, nil
	}
	dir := strings.TrimSpace(GetOptionalFlag(cmd, "dir"))
	if dir != "" {
		if strings.EqualFold(dir, "user") || strings.EqualFold(dir, "global") {
			return config.GlobalProjectRoot(), true, nil
		}
		return dir, false, nil
	}
	if GetBoolFlag(cmd, "global") {
		return config.GlobalProjectRoot(), true, nil
	}
	wd := GetOptionalFlag(cmd, "workdir")
	if a.Viper != nil && wd == "" {
		wd = strings.TrimSpace(a.Viper.GetString("workdir"))
	}
	resolved, err := a.ResolveWorkdir(wd, true)
	if err != nil {
		return "", false, err
	}
	return resolved, false, nil
}
