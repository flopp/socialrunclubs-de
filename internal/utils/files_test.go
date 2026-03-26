package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", tmpfile.Name(), true},
		{"non-existent file", "/path/to/nonexistent/file.txt", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			if result != tt.expected {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMakeDir(t *testing.T) {
	tmpBase, err := os.MkdirTemp("", "test-makedir-*")
	if err != nil {
		t.Fatalf("Failed to create temp base dir: %v", err)
	}
	defer os.RemoveAll(tmpBase)

	tests := []struct {
		name      string
		dir       string
		shouldErr bool
	}{
		{"create single directory", filepath.Join(tmpBase, "newdir"), false},
		{"create nested directories", filepath.Join(tmpBase, "a", "b", "c"), false},
		{"existing directory", tmpBase, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(tt.dir)
			err := MakeDir(tt.dir)
			if (err != nil) != tt.shouldErr {
				t.Errorf("MakeDir(%q) error = %v, want error = %v", tt.dir, err, tt.shouldErr)
			}
			if !tt.shouldErr {
				if !FileExists(tt.dir) {
					t.Errorf("MakeDir(%q) did not create directory", tt.dir)
				}
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpBase, err := os.MkdirTemp("", "test-copyfile-*")
	if err != nil {
		t.Fatalf("Failed to create temp base dir: %v", err)
	}
	defer os.RemoveAll(tmpBase)

	srcPath := filepath.Join(tmpBase, "source.txt")
	dstPath := filepath.Join(tmpBase, "dest", "destination.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(srcPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	tests := []struct {
		name      string
		src       string
		dst       string
		shouldErr bool
	}{
		{"copy to new directory", srcPath, dstPath, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(filepath.Dir(tt.dst))
			err := CopyFile(tt.src, tt.dst)
			if (err != nil) != tt.shouldErr {
				t.Errorf("CopyFile(%q, %q) error = %v, want error = %v", tt.src, tt.dst, err, tt.shouldErr)
			}
			if !tt.shouldErr {
				if !FileExists(tt.dst) {
					t.Errorf("CopyFile did not create destination file")
				}
				content, err := os.ReadFile(tt.dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
				}
				if string(content) != testContent {
					t.Errorf("CopyFile content mismatch: got %q, want %q", string(content), testContent)
				}
				srcInfo, _ := os.Stat(tt.src)
				dstInfo, _ := os.Stat(tt.dst)
				if srcInfo.Mode() != dstInfo.Mode() {
					t.Errorf("CopyFile permissions not copied: got %o, want %o", dstInfo.Mode(), srcInfo.Mode())
				}
			}
		})
	}

	t.Run("copy non-existent file", func(t *testing.T) {
		err := CopyFile("/nonexistent/file.txt", filepath.Join(tmpBase, "dest.txt"))
		if err == nil {
			t.Errorf("CopyFile expected error for non-existent source, got none")
		}
	})
}
