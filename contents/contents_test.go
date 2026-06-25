package contents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/natemarks/secret-hoard/tools"
)

func TestWriteFileWithPermissions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "write with 644 permissions",
			content: "test content",
			perm:    0644,
			wantErr: false,
		},
		{
			name:    "write with 600 permissions",
			content: "secret content",
			perm:    0600,
			wantErr: false,
		},
		{
			name:    "write with 755 permissions",
			content: "executable content",
			perm:    0755,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.txt")

			// Write file
			err := writeFileWithPermissions(tt.content, tmpFile, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeFileWithPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file exists
			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Errorf("file was not created: %s", tmpFile)
				return
			}

			// Verify content
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Errorf("failed to read file: %v", err)
				return
			}
			if string(content) != tt.content {
				t.Errorf("content = %q, want %q", string(content), tt.content)
			}

			// Verify permissions
			info, err := os.Stat(tmpFile)
			if err != nil {
				t.Errorf("failed to stat file: %v", err)
				return
			}
			if info.Mode().Perm() != tt.perm {
				t.Errorf("permissions = %o, want %o", info.Mode().Perm(), tt.perm)
			}
		})
	}
}

func TestSetRootOwnership(t *testing.T) {
	// This test only verifies the function doesn't crash when not running as root
	// Actual ownership change requires root privileges and is tested in integration tests

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	// Create a test file
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a properly initialized logger
	log := tools.NewLogger(false)

	// This should not error even if not running as root
	err = setRootOwnership(tmpFile, log)
	if err != nil {
		// Only fail if we're running as root - otherwise it's expected to be skipped
		if os.Geteuid() == 0 {
			t.Errorf("setRootOwnership() failed as root: %v", err)
		}
	}
}

func TestWriteSecretContents_InvalidSecretID(t *testing.T) {
	tests := []struct {
		name     string
		secretID string
		wantErr  bool
	}{
		{
			name:     "empty secret ID",
			secretID: "",
			wantErr:  true,
		},
		{
			name:     "invalid format - no slash",
			secretID: "invalid",
			wantErr:  true,
		},
		{
			name:     "unsupported type",
			secretID: "unknown/dev/test",
			wantErr:  true,
		},
	}

	tmpDir := t.TempDir()
	log := tools.NewLogger(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WriteSecretContents(tt.secretID, tmpDir, log)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteSecretContents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
