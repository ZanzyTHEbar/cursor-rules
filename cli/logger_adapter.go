package cli

import (
	"fmt"
	"log"
	"log/slog"
	"reflect"
)

// reflectLoggerAdapter calls Printf on the underlying value if available.
type reflectLoggerAdapter struct {
	v reflect.Value
}

func (r *reflectLoggerAdapter) Printf(format string, vv ...interface{}) {
	if !r.v.IsValid() {
		log.Printf(format, vv...)
		return
	}
	m := r.v.MethodByName("Printf")
	if m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(format), reflect.ValueOf(vv)})
		return
	}
	// fallback to standard logger
	log.Printf(format, vv...)
}

// NewReflectLoggerAdapter returns a Logger that will call Printf on the provided
// logger value (useful to adapt third-party loggers). If l is nil, the
// standard library logger is used.
func NewReflectLoggerAdapter(l interface{}) Logger {
	if l == nil {
		return log.Default()
	}
	return &reflectLoggerAdapter{v: reflect.ValueOf(l)}
}

// NewStdLoggerAdapter wraps the standard library logger as a Logger.
func NewStdLoggerAdapter(l *log.Logger) Logger {
	if l == nil {
		return log.Default()
	}
	return l
}

// slogAdapter adapts the global slog logger to the minimal Logger interface.
type slogAdapter struct{}

func (s *slogAdapter) Printf(format string, vv ...interface{}) {
	msg := fmt.Sprintf(format, vv...)
	slog.Info(msg)
}

// NewGoBasetoolsAdapter returns a Logger that routes to the go-basetools
// configured slog default logger. It is safe to call even before InitLogger;
// slog will use the standard library logger by default.
func NewGoBasetoolsAdapter() Logger {
	return &slogAdapter{}
}
