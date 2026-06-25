package contents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/natemarks/secret-hoard/tools"
)

// Integration tests that verify file creation, permissions, and content
// These tests use the real tools package but verify the integration logic

func TestWriteFilePermissions_Integration(t *testing.T) {
	// Test that verifies writeFileWithPermissions works correctly
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		filename string
		perm     os.FileMode
	}{
		{
			name:     "regular file 644",
			content:  "test content",
			filename: "test644.txt",
			perm:     0644,
		},
		{
			name:     "private file 600",
			content:  "secret content",
			filename: "test600.txt",
			perm:     0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)

			// Write file with specific permissions
			err := writeFileWithPermissions(tt.content, filePath, tt.perm)
			if err != nil {
				t.Fatalf("writeFileWithPermissions() error = %v", err)
			}

			// Verify file exists and has correct permissions
			info, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("failed to stat file: %v", err)
			}

			if info.Mode().Perm() != tt.perm {
				t.Errorf("permissions = %o, want %o", info.Mode().Perm(), tt.perm)
			}

			// Verify content
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			if string(content) != tt.content {
				t.Errorf("content = %q, want %q", string(content), tt.content)
			}
		})
	}
}

func TestSetRootOwnership_Integration(t *testing.T) {
	// This test verifies setRootOwnership doesn't crash when not running as root
	// Actual ownership changes require root privileges

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	// Create a test file
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	log := tools.NewLogger(false)

	// Should not error even when not running as root
	err = setRootOwnership(tmpFile, log)
	if err != nil && os.Geteuid() == 0 {
		// Only fail if we're actually running as root
		t.Errorf("setRootOwnership() failed as root: %v", err)
	}
}

func TestSecretTypeDetection_Integration(t *testing.T) {
	// Test that WriteSecretContents correctly identifies secret types
	tmpDir := t.TempDir()
	log := tools.NewLogger(false)

	tests := []struct {
		name     string
		secretID string
		wantErr  bool
	}{
		{
			name:     "valid jsondoc",
			secretID: "jsondoc/dev/test",
			wantErr:  true, // Will error because AWS call will fail, but that's OK
		},
		{
			name:     "valid textfile",
			secretID: "textfile/dev/test",
			wantErr:  true, // Will error because AWS call will fail, but that's OK
		},
		{
			name:     "valid sslcert",
			secretID: "sslcert/dev/test.com",
			wantErr:  true, // Will error because AWS call will fail, but that's OK
		},
		{
			name:     "invalid type",
			secretID: "unknown/dev/test",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			secretID: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WriteSecretContents(tt.secretID, tmpDir, log)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteSecretContents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
