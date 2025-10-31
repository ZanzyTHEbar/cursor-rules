package transform

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCursorTransformer(t *testing.T) {
	transformer := NewCursorTransformer()

	input := `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode.`

	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	// Transform should be identity
	outFM, outBody, err := transformer.Transform(fm, body)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if outBody != body {
		t.Errorf("Body changed: expected %q, got %q", body, outBody)
	}

	// Validate
	if err := transformer.Validate(outFM); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Check metadata
	if transformer.Target() != "cursor" {
		t.Errorf("Target: expected 'cursor', got %q", transformer.Target())
	}
	if transformer.Extension() != ".mdc" {
		t.Errorf("Extension: expected '.mdc', got %q", transformer.Extension())
	}
	if transformer.OutputDir() != ".cursor/rules" {
		t.Errorf("OutputDir: expected '.cursor/rules', got %q", transformer.OutputDir())
	}
}

func TestCopilotInstructionsTransformer(t *testing.T) {
	transformer := NewCopilotInstructionsTransformer()

	tests := []struct {
		name        string
		input       string
		wantApplyTo string
		wantDesc    string
		wantErr     bool
	}{
		{
			name: "basic transformation",
			input: `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode.`,
			wantApplyTo: "**/*.ts",
			wantDesc:    "Test rule",
			wantErr:     false,
		},
		{
			name: "array apply_to",
			input: `---
description: "Multi-pattern"
apply_to:
  - "**/*.ts"
  - "**/*.tsx"
---
Body`,
			wantApplyTo: "**/*.ts,**/*.tsx",
			wantDesc:    "Multi-pattern",
			wantErr:     false,
		},
		{
			name: "missing apply_to uses default",
			input: `---
description: "No pattern"
---
Body`,
			wantApplyTo: "**",
			wantDesc:    "No pattern",
			wantErr:     false,
		},
		{
			name: "missing description gets default",
			input: `---
apply_to: "**/*.js"
---
Body`,
			wantApplyTo: "**/*.js",
			wantDesc:    "Imported from Cursor rules",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := SplitFrontmatter([]byte(tt.input))
			if err != nil {
				t.Fatalf("SplitFrontmatter failed: %v", err)
			}

			outFM, _, err := transformer.Transform(fm, body)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Transform error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Validate
			if err := transformer.Validate(outFM); err != nil {
				t.Errorf("Validate failed: %v", err)
			}

			// Check transformed frontmatter
			var result map[string]interface{}
			if err := outFM.Decode(&result); err != nil {
				t.Fatalf("Decode result failed: %v", err)
			}

			if desc := result["description"].(string); desc != tt.wantDesc {
				t.Errorf("description: expected %q, got %q", tt.wantDesc, desc)
			}

			if applyTo := result["applyTo"].(string); applyTo != tt.wantApplyTo {
				t.Errorf("applyTo: expected %q, got %q", tt.wantApplyTo, applyTo)
			}

			// Ensure Cursor-specific fields are removed
			if _, ok := result["priority"]; ok {
				t.Error("priority field should be removed")
			}
			if _, ok := result["apply_to"]; ok {
				t.Error("apply_to field should be renamed to applyTo")
			}
		})
	}
}

func TestCopilotPromptsTransformer(t *testing.T) {
	transformer := NewCopilotPromptsTransformer()

	input := `---
description: "Generate component"
apply_to: "**/*.tsx"
---
Create a React component.`

	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	outFM, _, err := transformer.Transform(fm, body)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	// Validate
	if err := transformer.Validate(outFM); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Check transformed frontmatter
	var result map[string]interface{}
	if err := outFM.Decode(&result); err != nil {
		t.Fatalf("Decode result failed: %v", err)
	}

	// Should have mode
	if mode, ok := result["mode"].(string); !ok || mode != "chat" {
		t.Errorf("mode: expected 'chat', got %v", result["mode"])
	}

	// Should NOT have applyTo (removed for prompts)
	if _, ok := result["applyTo"]; ok {
		t.Error("applyTo should be removed for prompts")
	}

	// Check metadata
	if transformer.Target() != "copilot-prompt" {
		t.Errorf("Target: expected 'copilot-prompt', got %q", transformer.Target())
	}
	if transformer.Extension() != ".prompt.md" {
		t.Errorf("Extension: expected '.prompt.md', got %q", transformer.Extension())
	}
}

func TestTransformIdempotent(t *testing.T) {
	transformer := NewCopilotInstructionsTransformer()

	input := `---
description: "Test"
apply_to: "**/*.ts"
---
Body`

	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	// Transform twice
	out1, body1, err := transformer.Transform(fm, body)
	if err != nil {
		t.Fatalf("First transform failed: %v", err)
	}

	out2, body2, err := transformer.Transform(out1, body1)
	if err != nil {
		t.Fatalf("Second transform failed: %v", err)
	}

	// Marshal both
	data1, _ := MarshalMarkdown(out1, body1)
	data2, _ := MarshalMarkdown(out2, body2)

	// Should be identical
	if string(data1) != string(data2) {
		t.Error("Transform is not idempotent")
		t.Logf("First:\n%s", data1)
		t.Logf("Second:\n%s", data2)
	}
}

func TestTruncateBody(t *testing.T) {
	transformer := NewCopilotInstructionsTransformer()
	transformer.MaxTokens = 10 // Very small for testing

	longBody := strings.Repeat("a", 1000)

	fm := &yaml.Node{}
	fm.Encode(map[string]interface{}{
		"description": "Test",
		"applyTo":     "**",
	})

	_, outBody, err := transformer.Transform(fm, longBody)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	// Should be truncated
	if len(outBody) >= len(longBody) {
		t.Error("Body was not truncated")
	}

	if !strings.Contains(outBody, "truncated") {
		t.Error("Truncation message not found")
	}
}

func TestValidateGlobPattern(t *testing.T) {
	transformer := NewCopilotInstructionsTransformer()

	tests := []struct {
		pattern string
		wantErr bool
	}{
		{"**/*.ts", false},
		{"**/*.{ts,tsx}", false},
		{"src/**/*.js", false},
		{"**", false},
		{"[invalid", true}, // Invalid glob
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			err := transformer.validateGlobPattern(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGlobPattern(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid frontmatter",
			input: `---
key: value
---
Body`,
			wantErr: false,
		},
		{
			name:    "missing delimiters",
			input:   "No frontmatter here",
			wantErr: true,
		},
		{
			name: "empty body",
			input: `---
key: value
---`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := SplitFrontmatter([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalMarkdown(t *testing.T) {
	fm := &yaml.Node{}
	fm.Encode(map[string]interface{}{
		"description": "Test",
		"applyTo":     "**/*.ts",
	})

	body := "Test body content"

	output, err := MarshalMarkdown(fm, body)
	if err != nil {
		t.Fatalf("MarshalMarkdown failed: %v", err)
	}

	// Should have delimiters
	if !strings.HasPrefix(string(output), "---\n") {
		t.Error("Output should start with ---")
	}

	// Should contain body
	if !strings.Contains(string(output), body) {
		t.Error("Output should contain body")
	}

	// Should be parseable
	_, _, err = SplitFrontmatter(output)
	if err != nil {
		t.Errorf("Marshaled output is not parseable: %v", err)
	}
}

// Additional edge case tests

func TestSplitFrontmatterEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "extra dashes in body",
			input: `---
description: "Test"
---
Body with --- in it
More content`,
			wantErr: false,
		},
		{
			name: "empty body",
			input: `---
description: "Test"
---`,
			wantErr: false,
		},
		{
			name: "whitespace before frontmatter",
			input: `
---
description: "Test"
---
Body`,
			wantErr: false,
		},
		{
			name:    "no delimiters",
			input:   "Just text",
			wantErr: true,
		},
		{
			name: "only one delimiter",
			input: `---
description: "Test"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := SplitFrontmatter([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCopilotInstructionsTransformerEdgeCases(t *testing.T) {
	transformer := NewCopilotInstructionsTransformer()

	tests := []struct {
		name        string
		input       string
		wantApplyTo string
		wantDesc    string
	}{
		{
			name: "null apply_to",
			input: `---
description: "Test"
apply_to: null
---
Body`,
			wantApplyTo: "**",
			wantDesc:    "Test",
		},
		{
			name: "numeric apply_to",
			input: `---
description: "Test"
apply_to: 123
---
Body`,
			wantApplyTo: "**",
			wantDesc:    "Test",
		},
		{
			name: "empty string apply_to",
			input: `---
description: "Test"
apply_to: ""
---
Body`,
			wantApplyTo: "**",
			wantDesc:    "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := SplitFrontmatter([]byte(tt.input))
			if err != nil {
				t.Fatalf("SplitFrontmatter failed: %v", err)
			}

			outFM, _, err := transformer.Transform(fm, body)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			var result map[string]interface{}
			if err := outFM.Decode(&result); err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if result["applyTo"] != tt.wantApplyTo {
				t.Errorf("applyTo = %v, want %v", result["applyTo"], tt.wantApplyTo)
			}

			if result["description"] != tt.wantDesc {
				t.Errorf("description = %v, want %v", result["description"], tt.wantDesc)
			}
		})
	}
}

func TestTransformerMetadata(t *testing.T) {
	tests := []struct {
		name    string
		trans   Transformer
		wantTgt string
		wantExt string
		wantDir string
	}{
		{
			name:    "cursor",
			trans:   NewCursorTransformer(),
			wantTgt: "cursor",
			wantExt: ".mdc",
			wantDir: ".cursor/rules",
		},
		{
			name:    "copilot-instr",
			trans:   NewCopilotInstructionsTransformer(),
			wantTgt: "copilot-instr",
			wantExt: ".instructions.md",
			wantDir: ".github/instructions",
		},
		{
			name:    "copilot-prompt",
			trans:   NewCopilotPromptsTransformer(),
			wantTgt: "copilot-prompt",
			wantExt: ".prompt.md",
			wantDir: ".github/prompts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.trans.Target(); got != tt.wantTgt {
				t.Errorf("Target() = %v, want %v", got, tt.wantTgt)
			}
			if got := tt.trans.Extension(); got != tt.wantExt {
				t.Errorf("Extension() = %v, want %v", got, tt.wantExt)
			}
			if got := tt.trans.OutputDir(); got != tt.wantDir {
				t.Errorf("OutputDir() = %v, want %v", got, tt.wantDir)
			}
		})
	}
}
