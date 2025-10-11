package security

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "valid simple path",
			path:    "file.txt",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			path:    "dir/subdir/file.txt",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "path traversal with ../",
			path:    "../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path traversal in middle",
			path:    "dir/../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path traversal at end",
			path:    "dir/..",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "null byte",
			path:    "file\x00.txt",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "multiple path traversals",
			path:    "../../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ValidatePath() error type = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestValidatePathWithinBase(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		base    string
		wantErr bool
		errType error
	}{
		{
			name:    "path within base",
			path:    filepath.Join(tmpDir, "file.txt"),
			base:    tmpDir,
			wantErr: false,
		},
		{
			name:    "nested path within base",
			path:    filepath.Join(tmpDir, "dir", "file.txt"),
			base:    tmpDir,
			wantErr: false,
		},
		{
			name:    "path equals base",
			path:    tmpDir,
			base:    tmpDir,
			wantErr: false,
		},
		{
			name:    "path outside base",
			path:    filepath.Join(tmpDir, "..", "outside.txt"),
			base:    tmpDir,
			wantErr: true,
			errType: ErrPathOutsideBase,
		},
		{
			name:    "path with traversal",
			path:    filepath.Join(tmpDir, "dir", "..", "..", "outside.txt"),
			base:    tmpDir,
			wantErr: true,
			errType: ErrPathOutsideBase, // After cleaning, this resolves outside base
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathWithinBase(tt.path, tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathWithinBase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ValidatePathWithinBase() error type = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		base    string
		elem    []string
		wantErr bool
		errType error
	}{
		{
			name:    "simple join",
			base:    tmpDir,
			elem:    []string{"file.txt"},
			wantErr: false,
		},
		{
			name:    "nested join",
			base:    tmpDir,
			elem:    []string{"dir", "subdir", "file.txt"},
			wantErr: false,
		},
		{
			name:    "empty base",
			base:    "",
			elem:    []string{"file.txt"},
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "traversal in element",
			base:    tmpDir,
			elem:    []string{"../etc/passwd"},
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "multiple traversals",
			base:    tmpDir,
			elem:    []string{"dir", "..", "..", "outside.txt"},
			wantErr: true,
			errType: ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeJoin(tt.base, tt.elem...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeJoin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("SafeJoin() error type = %v, want %v", err, tt.errType)
			}
			if !tt.wantErr {
				// Verify result is within base
				if !strings.HasPrefix(result, tt.base) {
					t.Errorf("SafeJoin() result %v not within base %v", result, tt.base)
				}
			}
		})
	}
}

func TestValidatePresetName(t *testing.T) {
	tests := []struct {
		name    string
		preset  string
		wantErr bool
		errType error
	}{
		{
			name:    "valid simple name",
			preset:  "frontend",
			wantErr: false,
		},
		{
			name:    "valid name with dash",
			preset:  "my-preset",
			wantErr: false,
		},
		{
			name:    "valid name with underscore",
			preset:  "my_preset",
			wantErr: false,
		},
		{
			name:    "empty name",
			preset:  "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "name with forward slash",
			preset:  "dir/preset",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "name with backslash",
			preset:  "dir\\preset",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "name with path traversal",
			preset:  "../preset",
			wantErr: true,
			errType: ErrInvalidPath, // Caught by path separator check first
		},
		{
			name:    "hidden file",
			preset:  ".hidden",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "null byte",
			preset:  "pre\x00set",
			wantErr: true,
			errType: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePresetName(tt.preset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePresetName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ValidatePresetName() error type = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name    string
		pkg     string
		wantErr bool
		errType error
	}{
		{
			name:    "valid simple name",
			pkg:     "frontend",
			wantErr: false,
		},
		{
			name:    "valid nested name",
			pkg:     "frontend/react",
			wantErr: false,
		},
		{
			name:    "valid deeply nested",
			pkg:     "frontend/react/hooks",
			wantErr: false,
		},
		{
			name:    "empty name",
			pkg:     "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "path traversal",
			pkg:     "../etc",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path traversal in middle",
			pkg:     "dir/../etc",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "backslash separator",
			pkg:     "dir\\subdir",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "absolute path",
			pkg:     "/etc/passwd",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "empty component",
			pkg:     "dir//subdir",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "dot component",
			pkg:     "dir/./subdir",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "dotdot component",
			pkg:     "dir/../subdir",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "hidden component",
			pkg:     "dir/.hidden",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "null byte",
			pkg:     "dir\x00/file",
			wantErr: true,
			errType: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePackageName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ValidatePackageName() error type = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "file.txt",
			want:     "file.txt",
		},
		{
			name:     "filename with path separator",
			filename: "dir/file.txt",
			want:     "dir_file.txt",
		},
		{
			name:     "filename with backslash",
			filename: "dir\\file.txt",
			want:     "dir_file.txt",
		},
		{
			name:     "filename with path traversal",
			filename: "../file.txt",
			want:     "__file.txt", // Both ../ components replaced
		},
		{
			name:     "filename with null byte",
			filename: "file\x00.txt",
			want:     "file.txt",
		},
		{
			name:     "hidden file",
			filename: ".hidden",
			want:     "hidden",
		},
		{
			name:     "empty after sanitization",
			filename: "../..",
			want:     "___", // Results in underscores, not empty
		},
		{
			name:     "multiple issues",
			filename: "../dir/.hidden\x00.txt",
			want:     "__dir_.hidden.txt", // Preserves structure after sanitization
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.filename)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidatePath(b *testing.B) {
	paths := []string{
		"file.txt",
		"dir/subdir/file.txt",
		"../etc/passwd",
		"dir/../etc/passwd",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePath(paths[i%len(paths)])
	}
}

func BenchmarkValidatePathWithinBase(b *testing.B) {
	tmpDir := os.TempDir()
	paths := []string{
		filepath.Join(tmpDir, "file.txt"),
		filepath.Join(tmpDir, "dir", "file.txt"),
		filepath.Join(tmpDir, "..", "outside.txt"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePathWithinBase(paths[i%len(paths)], tmpDir)
	}
}

func BenchmarkSafeJoin(b *testing.B) {
	tmpDir := os.TempDir()
	elems := [][]string{
		{"file.txt"},
		{"dir", "subdir", "file.txt"},
		{"../etc/passwd"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SafeJoin(tmpDir, elems[i%len(elems)]...)
	}
}

func BenchmarkValidatePresetName(b *testing.B) {
	names := []string{
		"frontend",
		"my-preset",
		"../preset",
		".hidden",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePresetName(names[i%len(names)])
	}
}

func BenchmarkSanitizeFilename(b *testing.B) {
	filenames := []string{
		"file.txt",
		"dir/file.txt",
		"../file.txt",
		".hidden",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeFilename(filenames[i%len(filenames)])
	}
}
