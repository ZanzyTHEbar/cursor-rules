package cli

import (
	"fmt"
	"log"

	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
	"github.com/spf13/viper"
)

// Logger is a minimal logging interface used by AppContext.
type Logger interface {
	Printf(format string, v ...interface{})
}

// AppContext carries shared dependencies for command factories.
type AppContext struct {
	Viper        *viper.Viper
	Logger       Logger
	transformers map[string]transform.Transformer
}

// NewAppContext creates a default AppContext with provided logger and viper instance.
// Passing nil for the logger uses the standard library default logger.
func NewAppContext(v *viper.Viper, l Logger) *AppContext {
	if v == nil {
		v = viper.New()
	}
	if l == nil {
		l = log.Default()
	}

	ctx := &AppContext{
		Viper:        v,
		Logger:       l,
		transformers: make(map[string]transform.Transformer),
	}

	// Register default transformers
	ctx.RegisterTransformer("cursor", transform.NewCursorTransformer())
	ctx.RegisterTransformer("copilot-instr", transform.NewCopilotInstructionsTransformer())
	ctx.RegisterTransformer("copilot-prompt", transform.NewCopilotPromptsTransformer())

	return ctx
}

// RegisterTransformer adds a transformer to the context.
func (ctx *AppContext) RegisterTransformer(name string, t transform.Transformer) {
	ctx.transformers[name] = t
}

// Transformer retrieves a transformer by name.
func (ctx *AppContext) Transformer(target string) (transform.Transformer, error) {
	t, ok := ctx.transformers[target]
	if !ok {
		return nil, fmt.Errorf("unknown target: %s (available: cursor, copilot-instr, copilot-prompt)", target)
	}
	return t, nil
}

// AvailableTargets returns a list of registered transformer names.
func (ctx *AppContext) AvailableTargets() []string {
	targets := make([]string, 0, len(ctx.transformers))
	for k := range ctx.transformers {
		targets = append(targets, k)
	}
	return targets
}
