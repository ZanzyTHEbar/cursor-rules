package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTestFile creates a test file with the given content
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("Failed to create parent directories for %s: %v", path, err)
	}

	// #nosec G306 - test files don't need strict permissions
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
	return path
}

// CreateTestManifest creates a test manifest file
func CreateTestManifest(t *testing.T, dir, content string) string {
	t.Helper()
	return CreateTestFile(t, dir, "cursor-rules-manifest.yaml", content)
}

// CreateTestPreset creates a test preset file
func CreateTestPreset(t *testing.T, dir, name, content string) string {
	t.Helper()
	return CreateTestFile(t, dir, name+".mdc", content)
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// AssertFileContent checks file content matches expected
func AssertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("File content mismatch for %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

// AssertFileContains checks if file contains substring
func AssertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if !contains(string(content), substr) {
		t.Errorf("File %s does not contain %q\nContent:\n%s", path, substr, content)
	}
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// AssertErrorContains checks if error message contains substring
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", substr)
	}
	if !contains(err.Error(), substr) {
		t.Errorf("Error %q does not contain %q", err.Error(), substr)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, got, notWant interface{}) {
	t.Helper()
	if got == notWant {
		t.Errorf("got %v, expected different value", got)
	}
}

// AssertStringContains checks if string contains substring
func AssertStringContains(t *testing.T, str, substr string) {
	t.Helper()
	if !contains(str, substr) {
		t.Errorf("String %q does not contain %q", str, substr)
	}
}

// AssertStringNotContains checks if string does not contain substring
func AssertStringNotContains(t *testing.T, str, substr string) {
	t.Helper()
	if contains(str, substr) {
		t.Errorf("String %q should not contain %q", str, substr)
	}
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s != "" && substr != "" && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CreateTestDir creates a temporary test directory structure
func CreateTestDir(t *testing.T, structure map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	for path, content := range structure {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		// #nosec G306 - test files don't need strict permissions
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	return tmpDir
}

// MustReadFile reads a file and fails the test if it cannot be read
func MustReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// MustWriteFile writes a file and fails the test if it cannot be written
func MustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("Failed to create parent directories for %s: %v", path, err)
	}
	// #nosec G306 - test files don't need strict permissions
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}
