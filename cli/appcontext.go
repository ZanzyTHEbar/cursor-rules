package cli

import (
	"log"

	"github.com/spf13/viper"
)

// Logger is a minimal logging interface used by AppContext.
type Logger interface {
	Printf(format string, v ...interface{})
}

// AppContext carries shared dependencies for command factories.
type AppContext struct {
	Viper  *viper.Viper
	Logger Logger
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
	return &AppContext{Viper: v, Logger: l}
}
