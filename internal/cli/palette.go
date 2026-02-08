package cli

import "github.com/spf13/cobra"

// CommandFactory builds a cobra command using the provided AppContext.
type CommandFactory func(*AppContext) *cobra.Command

// Palette is an ordered collection of CommandFactory instances that can be
// composed into a root command.
type Palette []CommandFactory

// Register adds factories to the palette.
func (p *Palette) Register(factories ...CommandFactory) {
	*p = append(*p, factories...)
}

// Commands builds concrete *cobra.Command slice using the provided context.
func (p Palette) Commands(ctx *AppContext) []*cobra.Command {
	cmds := make([]*cobra.Command, 0, len(p))
	for _, f := range p {
		if f == nil {
			continue
		}
		cmd := f(ctx)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

// FromCommand wraps an existing *cobra.Command as a CommandFactory.
func FromCommand(c *cobra.Command) CommandFactory {
	return func(_ *AppContext) *cobra.Command { return c }
}
