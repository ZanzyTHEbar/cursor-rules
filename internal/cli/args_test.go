package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsReservedHelpArg(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{"help", true},
		{"Help", false},
		{"", false},
		{"frontend", false},
		{"all", false},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := IsReservedHelpArg(tt.arg)
			assert.Equal(t, tt.want, got, "IsReservedHelpArg(%q)", tt.arg)
		})
	}
}

func TestShowHelpIfReservedArg(t *testing.T) {
	t.Run("empty args returns false", func(t *testing.T) {
		c := &cobra.Command{Use: "test"}
		assert.False(t, ShowHelpIfReservedArg(c, nil))
		assert.False(t, ShowHelpIfReservedArg(c, []string{}))
	})

	t.Run("non-reserved first arg returns false", func(t *testing.T) {
		c := &cobra.Command{Use: "test"}
		assert.False(t, ShowHelpIfReservedArg(c, []string{"frontend"}))
		assert.False(t, ShowHelpIfReservedArg(c, []string{"all"}))
	})

	t.Run("reserved first arg runs help and returns true", func(t *testing.T) {
		c := &cobra.Command{
			Use:  "install [name]",
			Long: "Install something.",
			RunE: func(*cobra.Command, []string) error { return nil },
		}
		c.SetOut(bytes.NewBuffer(nil))
		c.SetErr(bytes.NewBuffer(nil))
		shown := ShowHelpIfReservedArg(c, []string{"help"})
		require.True(t, shown)
	})
}
