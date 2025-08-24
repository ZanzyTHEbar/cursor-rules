package cli

import (
	"bytes"

	"github.com/spf13/cobra"
)

// Execute runs the provided root command with args and returns captured stdout/stderr.
func Execute(root *cobra.Command, args ...string) (string, error) {
	_, out, err := ExecuteC(root, args...)
	return out, err
}

// ExecuteC returns the executed command plus captured combined output.
func ExecuteC(root *cobra.Command, args ...string) (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	c, err := root.ExecuteC()
	return c, buf.String(), err
}
