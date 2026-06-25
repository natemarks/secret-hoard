package contents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natemarks/secret-hoard/jsondoc"
	"github.com/natemarks/secret-hoard/secretlogic"
	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/textfile"
	"github.com/natemarks/secret-hoard/tools"
)

// WriteSecretContents downloads a secret and writes its contents to files in the target directory
// Returns absolute paths to all created files
func WriteSecretContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secret ID to determine type
	parts := strings.Split(secretID, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid secret ID format: %s (expected: type/env/...)", secretID)
	}

	secretType := parts[0]
	log.Debug("Secret type: %s", secretType)
	log.Debug("Secret ID: %s", secretID)
	log.Debug("Target directory: %s", targetDir)

	switch secretType {
	case "jsondoc":
		return writeJSONDocContents(secretID, targetDir, log)
	case "textfile", "text_file":
		return writeTextFileContents(secretID, targetDir, log)
	case "sslcert", "ssl_certificate":
		return writeSSLCertContents(secretID, targetDir, log)
	default:
		return nil, fmt.Errorf("unsupported secret type: %s (supported: jsondoc, textfile, sslcert)", secretType)
	}
}

func writeJSONDocContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic
	env, access, err := secretlogic.ParseJSONDocSecretID(secretID)
	if err != nil {
		return nil, err
	}

	// Fetch secret from AWS
	log.Info("Fetching secret: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return nil, fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data jsondoc.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("error parsing secret data: %w", err)
	}

	// Build filename using secretlogic
	filename := secretlogic.BuildContentsFilename("jsondoc", env, access, "")
	filePath := filepath.Join(targetDir, filename)

	// Write contents file
	log.Info("Writing contents to: %s", filePath)
	err = tools.WriteStringToFile(data.JSONContents, filePath)
	if err != nil {
		return nil, fmt.Errorf("error writing contents file: %w", err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving absolute path: %w", err)
	}

	log.Info("Successfully wrote jsondoc contents")
	return []string{absPath}, nil
}

func writeTextFileContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic (handles normalization internally)
	env, access, err := secretlogic.ParseTextFileSecretID(secretID)
	if err != nil {
		return nil, err
	}

	// Normalize for AWS API call
	secretID = secretlogic.NormalizeSecretID(secretID)

	// Fetch secret from AWS
	log.Info("Fetching secret: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return nil, fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data textfile.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("error parsing secret data: %w", err)
	}

	// Build filename using secretlogic
	filename := secretlogic.BuildContentsFilename("textfile", env, access, "")
	filePath := filepath.Join(targetDir, filename)

	// Write contents file
	log.Info("Writing contents to: %s", filePath)
	err = tools.WriteStringToFile(data.Contents, filePath)
	if err != nil {
		return nil, fmt.Errorf("error writing contents file: %w", err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving absolute path: %w", err)
	}

	log.Info("Successfully wrote textfile contents")
	return []string{absPath}, nil
}

func writeSSLCertContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic (handles normalization internally)
	env, commonName, err := secretlogic.ParseSSLCertSecretID(secretID)
	if err != nil {
		return nil, err
	}

	// Normalize for AWS API call
	secretID = secretlogic.NormalizeSecretID(secretID)

	// Fetch secret from AWS
	log.Info("Fetching secret: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return nil, fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data sslcert.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("error parsing secret data: %w", err)
	}

	// Build filenames using secretlogic
	certFilename, keyFilename := secretlogic.BuildSSLCertFilenames(env, commonName)
	certPath := filepath.Join(targetDir, certFilename)
	keyPath := filepath.Join(targetDir, keyFilename)

	// Write certificate file with 644 permissions
	log.Info("Writing certificate to: %s", certPath)
	err = writeFileWithPermissions(data.Certificate, certPath, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing certificate file: %w", err)
	}

	// Write key file with 600 permissions (more restrictive for private key)
	log.Info("Writing key to: %s", keyPath)
	err = writeFileWithPermissions(data.PrivateKey, keyPath, 0600)
	if err != nil {
		return nil, fmt.Errorf("error writing key file: %w", err)
	}

	// Change ownership to root:root if running as root
	err = setRootOwnership(certPath, log)
	if err != nil {
		log.Info("Warning: Could not set root ownership on certificate: %v", err)
	}

	err = setRootOwnership(keyPath, log)
	if err != nil {
		log.Info("Warning: Could not set root ownership on key: %v", err)
	}

	// Return the base path without extension so scripts can easily append .crt or .key
	baseName := secretlogic.BuildContentsFilename("sslcert", env, "", commonName)
	basePath := filepath.Join(targetDir, baseName)
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving base path: %w", err)
	}

	log.Info("Successfully wrote SSL certificate and key")
	log.Info("Certificate permissions: 644, Key permissions: 600")
	log.Info("Base path: %s", absBasePath)
	return []string{absBasePath}, nil
}

// writeFileWithPermissions writes a file with specific permissions
func writeFileWithPermissions(content, path string, perm os.FileMode) error {
	err := os.WriteFile(path, []byte(content), perm)
	if err != nil {
		return err
	}
	return nil
}

// setRootOwnership attempts to set file ownership to root:root (UID=0, GID=0)
// This will only succeed if the process is running as root
func setRootOwnership(path string, log *tools.Logger) error {
	// Check if running as root
	if os.Geteuid() != 0 {
		log.Debug("Not running as root, skipping ownership change for %s", path)
		return nil
	}

	// Change ownership to root:root (UID=0, GID=0)
	err := os.Chown(path, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to chown to root: %w", err)
	}

	log.Debug("Set ownership to root:root for %s", path)
	return nil
}
