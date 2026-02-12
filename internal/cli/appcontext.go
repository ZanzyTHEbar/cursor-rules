package cli

import (
	"log"

	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
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
	UI           Messenger
	transformers map[string]transform.Transformer
	app          *app.App
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
		UI:           NewMessenger(nil, nil, "info"),
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
		return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s (available: cursor, copilot-instr, copilot-prompt)", target)
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

// SetMessenger replaces the current messenger if a non-nil value is provided.
func (ctx *AppContext) SetMessenger(m Messenger) {
	if ctx == nil || m == nil {
		return
	}
	ctx.UI = m
}

// Messenger returns the currently configured messenger.
func (ctx *AppContext) Messenger() Messenger {
	if ctx == nil {
		return nil
	}
	if ctx.UI == nil {
		ctx.UI = NewMessenger(nil, nil, "info")
	}
	return ctx.UI
}

// App returns the application layer instance, creating it if needed.
func (ctx *AppContext) App() *app.App {
	if ctx == nil {
		return app.New(nil, nil)
	}
	if ctx.app == nil {
		ctx.app = app.New(ctx.Viper, ctx)
	} else {
		ctx.app.Viper = ctx.Viper
	}
	return ctx.app
}

// SetApp replaces the application layer instance.
func (ctx *AppContext) SetApp(a *app.App) {
	if ctx == nil {
		return
	}
	ctx.app = a
}
