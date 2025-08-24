package cli

import (
	"log"

	"github.com/spf13/viper"
)

// AppContext carries shared dependencies for command factories.
type AppContext struct {
	Viper  *viper.Viper
	Logger *log.Logger
}

// NewAppContext creates a default AppContext with provided logger and viper instance.
func NewAppContext(v *viper.Viper, l *log.Logger) *AppContext {
	if v == nil {
		v = viper.New()
	}
	if l == nil {
		l = log.Default()
	}
	return &AppContext{Viper: v, Logger: l}
}
