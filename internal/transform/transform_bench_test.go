package transform

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// BenchmarkCursorTransformer benchmarks the Cursor transformer
func BenchmarkCursorTransformer(b *testing.B) {
	transformer := NewCursorTransformer()
	input := `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode and follow TypeScript best practices.
Always use explicit types.
Avoid using 'any' type.
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkCopilotInstructionsTransformer benchmarks the Copilot Instructions transformer
func BenchmarkCopilotInstructionsTransformer(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()
	input := `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode and follow TypeScript best practices.
Always use explicit types.
Avoid using 'any' type.
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkCopilotPromptsTransformer benchmarks the Copilot Prompts transformer
func BenchmarkCopilotPromptsTransformer(b *testing.B) {
	transformer := NewCopilotPromptsTransformer()
	input := `---
description: "Generate component"
apply_to: "**/*.tsx"
---
Create a React component with TypeScript.
Include proper types and documentation.
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkSplitFrontmatter benchmarks frontmatter parsing
func BenchmarkSplitFrontmatter(b *testing.B) {
	input := []byte(`---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode and follow TypeScript best practices.
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = SplitFrontmatter(input)
	}
}

// BenchmarkMarshalMarkdown benchmarks markdown marshaling
func BenchmarkMarshalMarkdown(b *testing.B) {
	fm := &yaml.Node{}
	fm.Encode(map[string]interface{}{
		"description": "Test rule",
		"applyTo":     "**/*.ts",
	})
	body := "Use strict mode and follow TypeScript best practices."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalMarkdown(fm, body)
	}
}

// BenchmarkTransformWithLargeBody benchmarks transformation with large content
func BenchmarkTransformWithLargeBody(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()

	// Create large body content
	var bodyBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		bodyBuilder.WriteString("This is a line of content that will be repeated many times. ")
	}
	largeBody := bodyBuilder.String()

	input := `---
description: "Large content test"
apply_to: "**/*.ts"
---
` + largeBody

	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkTransformWithMultiplePatterns benchmarks transformation with array patterns
func BenchmarkTransformWithMultiplePatterns(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()
	input := `---
description: "Multi-pattern rule"
apply_to:
  - "**/*.ts"
  - "**/*.tsx"
  - "**/*.js"
  - "**/*.jsx"
priority: 2
---
Apply to all JavaScript and TypeScript files.
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkValidateGlobPattern benchmarks glob pattern validation
func BenchmarkValidateGlobPattern(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()
	pattern := "**/*.{ts,tsx,js,jsx}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transformer.validateGlobPattern(pattern)
	}
}

// BenchmarkTransformIdempotent benchmarks idempotent transformation
func BenchmarkTransformIdempotent(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()
	input := `---
description: "Test"
apply_to: "**/*.ts"
---
Body content
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	// First transformation
	fm, body, err = transformer.Transform(fm, body)
	if err != nil {
		b.Fatalf("First transform failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = transformer.Transform(fm, body)
	}
}

// BenchmarkParallelTransform benchmarks parallel transformation
func BenchmarkParallelTransform(b *testing.B) {
	transformer := NewCopilotInstructionsTransformer()
	input := `---
description: "Test rule"
apply_to: "**/*.ts"
---
Use strict mode.
`
	fm, body, err := SplitFrontmatter([]byte(input))
	if err != nil {
		b.Fatalf("SplitFrontmatter failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = transformer.Transform(fm, body)
		}
	})
}

// BenchmarkCompleteWorkflow benchmarks a complete transformation workflow
func BenchmarkCompleteWorkflow(b *testing.B) {
	input := []byte(`---
description: "Complete workflow test"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode and follow best practices.
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Parse
		fm, body, err := SplitFrontmatter(input)
		if err != nil {
			b.Fatal(err)
		}

		// Transform
		transformer := NewCopilotInstructionsTransformer()
		outFM, outBody, err := transformer.Transform(fm, body)
		if err != nil {
			b.Fatal(err)
		}

		// Marshal
		_, err = MarshalMarkdown(outFM, outBody)
		if err != nil {
			b.Fatal(err)
		}
	}
}
