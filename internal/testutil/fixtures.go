package testutil

// Common test fixtures and data

// ValidManifest returns a valid manifest YAML content
func ValidManifest() string {
	return `version: "1.0"
targets:
  - cursor
  - copilot-instr
  - copilot-prompt
`
}

// MinimalManifest returns a minimal manifest YAML content
func MinimalManifest() string {
	return `version: "1.0"
targets:
  - cursor
`
}

// ManifestWithOverrides returns a manifest with overrides
func ManifestWithOverrides() string {
	return `version: "1.0"
targets:
  - copilot-prompt
overrides:
  copilot-prompt:
    defaultMode: "agent"
    defaultTools:
      - "githubRepo"
`
}

// InvalidManifest returns an invalid manifest YAML content
func InvalidManifest() string {
	return `version: "1.0
targets: [unclosed
`
}

// ValidPresetWithFrontmatter returns a valid preset with frontmatter
func ValidPresetWithFrontmatter() string {
	return `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode and follow TypeScript best practices.
`
}

// ValidPresetMinimal returns a minimal valid preset
func ValidPresetMinimal() string {
	return `---
description: "Minimal rule"
---
Simple rule content.
`
}

// ValidPresetMultiplePatterns returns a preset with multiple apply_to patterns
func ValidPresetMultiplePatterns() string {
	return `---
description: "Multi-pattern rule"
apply_to:
  - "**/*.ts"
  - "**/*.tsx"
priority: 2
---
Apply to TypeScript and TSX files.
`
}

// PresetWithoutFrontmatter returns content without frontmatter
func PresetWithoutFrontmatter() string {
	return `This is just plain text without frontmatter.
It should fail validation.
`
}

// PresetWithInvalidFrontmatter returns content with invalid YAML frontmatter
func PresetWithInvalidFrontmatter() string {
	return `---
description: "Invalid
apply_to: [unclosed
---
Content here.
`
}

// CopilotInstructionsFormat returns expected Copilot instructions format
func CopilotInstructionsFormat() string {
	return `---
description: "Test rule"
applyTo: "**/*.ts"
---
Use strict mode and follow TypeScript best practices.
`
}

// CopilotPromptsFormat returns expected Copilot prompts format
func CopilotPromptsFormat() string {
	return `---
description: "Generate component"
mode: "chat"
---
Create a React component with TypeScript.
`
}

// LongContent returns a long content string for truncation testing
func LongContent() string {
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is a very long line of text that will be used for testing truncation. "
	}
	return content
}

// TestDirectoryStructure returns a map of file paths to content for testing
func TestDirectoryStructure() map[string]string {
	return map[string]string{
		"cursor-rules-manifest.yaml": ValidManifest(),
		"frontend.mdc":               ValidPresetWithFrontmatter(),
		"backend.mdc":                ValidPresetMinimal(),
		".cursor/rules/test.mdc":     ValidPresetWithFrontmatter(),
	}
}

// EmptyDirectoryStructure returns an empty directory structure
func EmptyDirectoryStructure() map[string]string {
	return map[string]string{}
}
