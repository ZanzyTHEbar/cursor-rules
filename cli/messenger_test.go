package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestMessengerSuppressesDebugAtInfoLevel(t *testing.T) {
	var out bytes.Buffer
	m := NewMessenger(&out, &out, "info")

	m.Debug("hidden\n")
	if out.Len() != 0 {
		t.Fatalf("expected no debug output at info level, got %q", out.String())
	}

	m.Success("visible\n")
	if !strings.Contains(out.String(), "visible") {
		t.Fatalf("expected success output, got %q", out.String())
	}
}

func TestMessengerShowsDebugAtDebugLevel(t *testing.T) {
	var out bytes.Buffer
	m := NewMessenger(&out, &out, "debug")

	m.Debug("details\n")
	if !strings.Contains(out.String(), "details") {
		t.Fatalf("expected debug output when log-level=debug, got %q", out.String())
	}
}
