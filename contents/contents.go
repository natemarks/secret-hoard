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
		return nil, fmt.Errorf("invalid secret ID format: %s\n"+
			"Expected format: <type>/<env>/<identifier>\n"+
			"Examples:\n"+
			"  jsondoc/dev/app-config\n"+
			"  textfile/prod/api-key\n"+
			"  sslcert/prod/example.com", secretID)
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
		return nil, fmt.Errorf("unsupported secret type: %s\n"+
			"Supported types: jsondoc, textfile, sslcert\n"+
			"Examples:\n"+
			"  jsondoc/dev/app-config\n"+
			"  textfile/prod/api-key\n"+
			"  sslcert/prod/example.com", secretType)
	}
}

// secretWriter encapsulates the common pattern for writing secrets
type secretWriter struct {
	secretID  string
	targetDir string
	log       *tools.Logger
}

// fetchAndParse fetches a secret and returns the raw JSON value
func (w *secretWriter) fetchAndParse() (string, error) {
	w.log.Info("Fetching secret: %s", w.secretID)
	secretValue, err := tools.GetSecretValue(w.secretID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch secret from AWS Secrets Manager: %w\n"+
			"Secret ID: %s\n"+
			"Possible causes:\n"+
			"  - Secret does not exist in AWS Secrets Manager\n"+
			"  - Insufficient AWS permissions (requires secretsmanager:GetSecretValue)\n"+
			"  - AWS credentials not configured (run: aws sso login --profile claude-code)\n"+
			"  - Wrong AWS region configured", err, w.secretID)
	}
	return secretValue, nil
}

// writeFile writes content to a file with better error messages
func (w *secretWriter) writeFile(content, filePath string, perm os.FileMode) error {
	w.log.Info("Writing to: %s", filePath)
	err := os.WriteFile(filePath, []byte(content), perm)
	if err != nil {
		return fmt.Errorf("failed to write file: %w\n"+
			"Path: %s\n"+
			"Possible causes:\n"+
			"  - Target directory does not exist (create it first: mkdir -p %s)\n"+
			"  - No write permission to directory (try: sudo %s or chmod +w %s)\n"+
			"  - Disk full or quota exceeded\n"+
			"  - Path is a directory not a file",
			err, filePath, w.targetDir, os.Args[0], filepath.Dir(filePath))
	}
	return nil
}

func writeJSONDocContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic
	env, access, err := secretlogic.ParseJSONDocSecretID(secretID)
	if err != nil {
		return nil, fmt.Errorf("%w\n"+
			"Expected format: jsondoc/<env>/<access>\n"+
			"Example: jsondoc/dev/app-config", err)
	}

	writer := &secretWriter{secretID: secretID, targetDir: targetDir, log: log}

	// Fetch and parse secret data
	secretValue, err := writer.fetchAndParse()
	if err != nil {
		return nil, err
	}

	var data jsondoc.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret data as JSON: %w\n"+
			"Secret ID: %s\n"+
			"The secret value must be valid JSON in jsondoc format", err, secretID)
	}

	// Build filename and write
	filename := secretlogic.BuildContentsFilename("jsondoc", env, access, "")
	filePath := filepath.Join(targetDir, filename)

	err = writer.writeFile(data.JSONContents, filePath, 0644)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	log.Info("Successfully wrote jsondoc contents")
	return []string{absPath}, nil
}

func writeTextFileContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic (handles normalization internally)
	env, access, err := secretlogic.ParseTextFileSecretID(secretID)
	if err != nil {
		return nil, fmt.Errorf("%w\n"+
			"Expected format: textfile/<env>/<access>\n"+
			"Example: textfile/prod/api-key", err)
	}

	// Normalize for AWS API call
	normalizedID := secretlogic.NormalizeSecretID(secretID)
	writer := &secretWriter{secretID: normalizedID, targetDir: targetDir, log: log}

	// Fetch and parse secret data
	secretValue, err := writer.fetchAndParse()
	if err != nil {
		return nil, err
	}

	var data textfile.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret data as JSON: %w\n"+
			"Secret ID: %s\n"+
			"The secret value must be valid JSON in textfile format", err, secretID)
	}

	// Build filename and write
	filename := secretlogic.BuildContentsFilename("textfile", env, access, "")
	filePath := filepath.Join(targetDir, filename)

	err = writer.writeFile(data.Contents, filePath, 0644)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	log.Info("Successfully wrote textfile contents")
	return []string{absPath}, nil
}

func writeSSLCertContents(secretID, targetDir string, log *tools.Logger) ([]string, error) {
	// Parse secretID using secretlogic (handles normalization internally)
	env, commonName, err := secretlogic.ParseSSLCertSecretID(secretID)
	if err != nil {
		return nil, fmt.Errorf("%w\n"+
			"Expected format: sslcert/<env>/<commonName>\n"+
			"Example: sslcert/prod/example.com", err)
	}

	// Normalize for AWS API call
	normalizedID := secretlogic.NormalizeSecretID(secretID)
	writer := &secretWriter{secretID: normalizedID, targetDir: targetDir, log: log}

	// Fetch and parse secret data
	secretValue, err := writer.fetchAndParse()
	if err != nil {
		return nil, err
	}

	var data sslcert.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret data as JSON: %w\n"+
			"Secret ID: %s\n"+
			"The secret value must be valid JSON in sslcert format", err, secretID)
	}

	// Build filenames using secretlogic
	certFilename, keyFilename := secretlogic.BuildSSLCertFilenames(env, commonName)
	certPath := filepath.Join(targetDir, certFilename)
	keyPath := filepath.Join(targetDir, keyFilename)

	// Write certificate file with 644 permissions
	log.Info("Writing certificate to: %s", certPath)
	err = writeFileWithPermissions(data.Certificate, certPath, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write certificate: %w\n"+
			"Path: %s\n"+
			"Possible causes:\n"+
			"  - Target directory does not exist (create it first: mkdir -p %s)\n"+
			"  - No write permission to directory (try: sudo %s or chmod +w %s)",
			err, certPath, targetDir, os.Args[0], filepath.Dir(certPath))
	}

	// Write key file with 600 permissions (more restrictive for private key)
	log.Info("Writing key to: %s", keyPath)
	err = writeFileWithPermissions(data.PrivateKey, keyPath, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write private key: %w\n"+
			"Path: %s\n"+
			"Possible causes:\n"+
			"  - Target directory does not exist (create it first: mkdir -p %s)\n"+
			"  - No write permission to directory (try: sudo %s or chmod +w %s)",
			err, keyPath, targetDir, os.Args[0], filepath.Dir(keyPath))
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
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
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
