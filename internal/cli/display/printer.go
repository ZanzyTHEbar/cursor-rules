package display

import (
	"fmt"
	"io"
)

// Printer wraps a messenger and writers for consistent output.
type Printer struct {
	UI  Messenger
	Out io.Writer
	Err io.Writer
}

// Messenger defines the output contract used by Printer.
type Messenger interface {
	Success(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// NewPrinter builds a Printer from the messenger and writers.
func NewPrinter(ui Messenger, out, err io.Writer) Printer {
	return Printer{UI: ui, Out: out, Err: err}
}

func (p Printer) Info(format string, args ...interface{}) {
	if p.UI != nil {
		p.UI.Info(format, args...)
		return
	}
	if p.Out != nil {
		fmt.Fprintf(p.Out, format, args...)
	}
}

func (p Printer) Success(format string, args ...interface{}) {
	if p.UI != nil {
		p.UI.Success(format, args...)
		return
	}
	if p.Out != nil {
		fmt.Fprintf(p.Out, format, args...)
	}
}

func (p Printer) Warn(format string, args ...interface{}) {
	if p.UI != nil {
		p.UI.Warn(format, args...)
		return
	}
	if p.Out != nil {
		fmt.Fprintf(p.Out, format, args...)
	}
}

func (p Printer) Error(format string, args ...interface{}) {
	if p.UI != nil {
		p.UI.Error(format, args...)
		return
	}
	if p.Err != nil {
		fmt.Fprintf(p.Err, format, args...)
	}
}
