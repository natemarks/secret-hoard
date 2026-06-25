package contents

import (
	"os"
	"path/filepath"
	"strings"
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

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "empty string",
			content: "",
			want:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:    "simple string",
			content: "hello world",
			want:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:    "json content",
			content: `{"key": "value"}`,
			want:    "9724c1e20e6e3e4d7f57ed25f9d4efb006e508590d528c90da597f6a775c13e5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateChecksum(tt.content)
			if got != tt.want {
				t.Errorf("calculateChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyFileChecksum(t *testing.T) {
	log := tools.NewLogger(false)
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		fileContent     string
		expectedContent string
		wantErr         bool
	}{
		{
			name:            "matching content",
			fileContent:     "test content",
			expectedContent: "test content",
			wantErr:         false,
		},
		{
			name:            "mismatched content",
			fileContent:     "actual content",
			expectedContent: "expected content",
			wantErr:         true,
		},
		{
			name:            "empty content matches",
			fileContent:     "",
			expectedContent: "",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with actual content
			tmpFile := filepath.Join(tmpDir, tt.name+".txt")
			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Verify against expected content
			err = verifyFileChecksum(tmpFile, tt.expectedContent, log)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyFileChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If error expected, verify it contains checksum info
			if tt.wantErr && err != nil {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "SHA256") {
					t.Errorf("error message should contain SHA256 checksums, got: %v", errMsg)
				}
				if !strings.Contains(errMsg, "Expected") || !strings.Contains(errMsg, "Actual") {
					t.Errorf("error message should show expected vs actual, got: %v", errMsg)
				}
			}
		})
	}
}

func TestVerifyFileChecksum_FileNotFound(t *testing.T) {
	log := tools.NewLogger(false)
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.txt")

	err := verifyFileChecksum(nonExistentFile, "content", log)
	if err == nil {
		t.Error("verifyFileChecksum() should error when file does not exist")
	}

	if !strings.Contains(err.Error(), "cannot read file") {
		t.Errorf("error should indicate file read failure, got: %v", err)
	}
}
