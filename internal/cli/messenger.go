package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Messenger centralizes user-facing output formatting.
type Messenger interface {
	Success(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// NewMessenger returns a Messenger writing to the provided writers (or STDOUT/ERR if nil).
func NewMessenger(out, err io.Writer, level string) Messenger {
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return &stdMessenger{
		out:   out,
		err:   err,
		level: parseMessengerLevel(level),
	}
}

type logSeverity int

const (
	severityDebug logSeverity = iota
	severityInfo
	severityWarn
	severityError
)

type stdMessenger struct {
	out   io.Writer
	err   io.Writer
	level logSeverity
}

func (m *stdMessenger) Success(format string, args ...interface{}) {
	m.printf(m.out, format, args...)
}

func (m *stdMessenger) Info(format string, args ...interface{}) {
	if m.shouldPrint(severityInfo) {
		m.printf(m.out, format, args...)
	}
}

func (m *stdMessenger) Warn(format string, args ...interface{}) {
	if m.shouldPrint(severityWarn) {
		m.printf(m.out, format, args...)
	}
}

func (m *stdMessenger) Error(format string, args ...interface{}) {
	if m.shouldPrint(severityError) {
		m.printf(m.err, format, args...)
	}
}

func (m *stdMessenger) Debug(format string, args ...interface{}) {
	if m.shouldPrint(severityDebug) {
		m.printf(m.out, format, args...)
	}
}

func (m *stdMessenger) shouldPrint(severity logSeverity) bool {
	return severity >= m.level
}

func (m *stdMessenger) printf(w io.Writer, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func parseMessengerLevel(level string) logSeverity {
	switch strings.ToLower(level) {
	case "debug":
		return severityDebug
	case "warn", "warning":
		return severityWarn
	case "error":
		return severityError
	default:
		return severityInfo
	}
}
