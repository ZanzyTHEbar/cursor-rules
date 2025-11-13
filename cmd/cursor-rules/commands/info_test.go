package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestInfoCommandPrintsSections(t *testing.T) {
	tmp := t.TempDir()
	v := viper.New()
	v.Set("workdir", tmp)
	ctx := cli.NewAppContext(v, nil)

	root := &cobra.Command{Use: "cursor-rules"}
	root.PersistentFlags().StringP("workdir", "w", "", "workspace root")
	root.AddCommand(NewInfoCmd(ctx))

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"info", "--workdir", tmp})

	if err := root.Execute(); err != nil {
		t.Fatalf("info command failed: %v", err)
	}

	output := buf.String()
	for _, section := range []string{"Configuration", "Workspace"} {
		if !strings.Contains(output, section) {
			t.Fatalf("expected output to contain section %q; output=%s", section, output)
		}
	}
}
