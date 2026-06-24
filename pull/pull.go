package pull

import (
	"encoding/json"
	"fmt"

	"github.com/natemarks/secret-hoard/jsondoc"
	"github.com/natemarks/secret-hoard/rdspostgres"
	"github.com/natemarks/secret-hoard/secretlogic"
	"github.com/natemarks/secret-hoard/snowflake"
	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/textfile"
	"github.com/natemarks/secret-hoard/tools"
)

// PullMetadata contains the metadata needed to pull a secret
type PullMetadata struct {
	Type       string
	Env        string
	Access     string
	Instance   string
	Database   string
	Warehouse  string
	CommonName string
}

// PullSecret downloads a secret and creates local editable files
func PullSecret(meta PullMetadata, log *tools.Logger) error {
	switch meta.Type {
	case "jsondoc":
		return pullJSONDoc(meta, log)
	case "text_file":
		return pullTextFile(meta, log)
	case "ssl_certificate":
		return pullSSLCert(meta, log)
	case "rdspostgres":
		return pullRDSPostgres(meta, log)
	case "snowflake":
		return pullSnowflake(meta, log)
	default:
		return fmt.Errorf("unknown secret type: %s", meta.Type)
	}
}

func pullJSONDoc(meta PullMetadata, log *tools.Logger) error {
	secretID := fmt.Sprintf("jsondoc/%s/%s", meta.Env, meta.Access)
	log.Info("Pulling secret: %s", secretID)

	// Get working directory
	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}
	log.Debug("Using working directory: %s", workingDir)

	// Fetch secret from AWS
	fmt.Printf("Fetching secret from AWS: %s\n", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data jsondoc.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return fmt.Errorf("error parsing secret: %w", err)
	}

	// Build file paths
	metadataFile, contentsFile := secretlogic.BuildJSONDocPaths(workingDir, meta.Env, meta.Access)

	// Create metadata file
	metadataContent := map[string]string{
		"resourceType":  "jsondoc",
		"environment":   meta.Env,
		"access":        meta.Access,
		"JSONSha256Sum": data.JSONSha256Sum,
	}

	metadataJSON, err := json.MarshalIndent(metadataContent, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(metadataJSON), metadataFile)
	if err != nil {
		return fmt.Errorf("error writing metadata file: %w", err)
	}

	// Create contents file
	err = tools.WriteStringToFile(data.JSONContents, contentsFile)
	if err != nil {
		return fmt.Errorf("error writing contents file: %w", err)
	}

	// Verify SHA256
	err = tools.CheckSha256Sum(contentsFile, data.JSONSha256Sum)
	if err != nil {
		log.Error("SHA256 verification failed")
		return fmt.Errorf("SHA256 verification failed: %w", err)
	}

	log.Info("Created files: %s, %s", metadataFile, contentsFile)
	fmt.Printf("✓ Successfully pulled secret: %s\n", secretID)
	fmt.Printf("  - %s\n", metadataFile)
	fmt.Printf("  - %s\n", contentsFile)

	return nil
}

func pullTextFile(meta PullMetadata, log *tools.Logger) error {
	secretID := fmt.Sprintf("text_file/%s/%s", meta.Env, meta.Access)
	log.Info("Pulling secret: %s", secretID)

	// Get working directory
	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}
	log.Debug("Using working directory: %s", workingDir)

	// Fetch secret from AWS
	fmt.Printf("Fetching secret from AWS: %s\n", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data textfile.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return fmt.Errorf("error parsing secret: %w", err)
	}

	// Build file paths
	metadataFile, contentsFile := secretlogic.BuildTextFilePaths(workingDir, meta.Env, meta.Access)

	// Create metadata file
	metadataContent := map[string]string{
		"resourceType": "text_file",
		"environment":  meta.Env,
		"access":       meta.Access,
		"sha256Sum":    data.Sha256Sum,
	}

	metadataJSON, err := json.MarshalIndent(metadataContent, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(metadataJSON), metadataFile)
	if err != nil {
		return fmt.Errorf("error writing metadata file: %w", err)
	}

	// Create contents file
	err = tools.WriteStringToFile(data.Contents, contentsFile)
	if err != nil {
		return fmt.Errorf("error writing contents file: %w", err)
	}

	// Verify SHA256
	err = tools.CheckSha256Sum(contentsFile, data.Sha256Sum)
	if err != nil {
		log.Error("SHA256 verification failed")
		return fmt.Errorf("SHA256 verification failed: %w", err)
	}

	log.Info("Created files: %s, %s", metadataFile, contentsFile)
	fmt.Printf("✓ Successfully pulled secret: %s\n", secretID)
	fmt.Printf("  - %s\n", metadataFile)
	fmt.Printf("  - %s\n", contentsFile)

	return nil
}

func pullSSLCert(meta PullMetadata, log *tools.Logger) error {
	secretID := fmt.Sprintf("ssl_certificate/%s/%s", meta.Env, meta.CommonName)
	log.Info("Pulling secret: %s", secretID)

	// Get working directory
	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}
	log.Debug("Using working directory: %s", workingDir)

	// Fetch secret from AWS
	fmt.Printf("Fetching secret from AWS: %s\n", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data sslcert.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return fmt.Errorf("error parsing secret: %w", err)
	}

	// Build file paths
	metadataFile, certFile, keyFile := secretlogic.BuildSSLCertPaths(workingDir, meta.Env, meta.CommonName)

	// Create metadata file with all computed fields
	metadataContent := map[string]string{
		"resourceType":      "ssl_certificate",
		"environment":       meta.Env,
		"commonName":        meta.CommonName,
		"expirationDate":    data.ExpirationDate,
		"modulus":           data.Modulus,
		"certificateSha256": data.CertificateSha256,
		"privateKeySha256":  data.PrivateKeySha256,
	}

	metadataJSON, err := json.MarshalIndent(metadataContent, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(metadataJSON), metadataFile)
	if err != nil {
		return fmt.Errorf("error writing metadata file: %w", err)
	}

	// Create certificate file
	err = tools.WriteStringToFile(data.Certificate, certFile)
	if err != nil {
		return fmt.Errorf("error writing certificate file: %w", err)
	}

	// Verify certificate SHA256
	err = tools.CheckSha256Sum(certFile, data.CertificateSha256)
	if err != nil {
		log.Error("Certificate SHA256 verification failed")
		return fmt.Errorf("certificate SHA256 verification failed: %w", err)
	}

	// Create private key file
	err = tools.WriteStringToFile(data.PrivateKey, keyFile)
	if err != nil {
		return fmt.Errorf("error writing key file: %w", err)
	}

	// Verify key SHA256
	err = tools.CheckSha256Sum(keyFile, data.PrivateKeySha256)
	if err != nil {
		log.Error("Private key SHA256 verification failed")
		return fmt.Errorf("private key SHA256 verification failed: %w", err)
	}

	log.Info("Created files: %s, %s, %s", metadataFile, certFile, keyFile)
	fmt.Printf("✓ Successfully pulled secret: %s\n", secretID)
	fmt.Printf("  - %s\n", metadataFile)
	fmt.Printf("  - %s\n", certFile)
	fmt.Printf("  - %s\n", keyFile)

	return nil
}

func pullRDSPostgres(meta PullMetadata, log *tools.Logger) error {
	secretID := fmt.Sprintf("rdspostgres/%s/%s/%s/%s", meta.Env, meta.Instance, meta.Database, meta.Access)
	log.Info("Pulling secret: %s", secretID)

	// Get working directory
	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}
	log.Debug("Using working directory: %s", workingDir)

	// Fetch secret from AWS
	fmt.Printf("Fetching secret from AWS: %s\n", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data rdspostgres.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return fmt.Errorf("error parsing secret: %w", err)
	}

	// Build file path
	filename := secretlogic.BuildRDSPostgresPath(workingDir, meta.Env, meta.Instance, meta.Database, meta.Access)

	fileContent := map[string]interface{}{
		"metadata": map[string]string{
			"resourceType": "rdspostgres",
			"environment":  meta.Env,
			"instance":     meta.Instance,
			"database":     meta.Database,
			"access":       meta.Access,
		},
		"data": data,
	}

	fileJSON, err := json.MarshalIndent(fileContent, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(fileJSON), filename)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Info("Created file: %s", filename)
	fmt.Printf("✓ Successfully pulled secret: %s\n", secretID)
	fmt.Printf("  - %s\n", filename)

	return nil
}

func pullSnowflake(meta PullMetadata, log *tools.Logger) error {
	secretID := fmt.Sprintf("snowflake/%s/%s/%s", meta.Env, meta.Warehouse, meta.Access)
	log.Info("Pulling secret: %s", secretID)

	// Get working directory
	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}
	log.Debug("Using working directory: %s", workingDir)

	// Fetch secret from AWS
	fmt.Printf("Fetching secret from AWS: %s\n", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching secret: %w", err)
	}

	// Parse secret data
	var data snowflake.Data
	err = json.Unmarshal([]byte(secretValue), &data)
	if err != nil {
		return fmt.Errorf("error parsing secret: %w", err)
	}

	// Build file path
	filename := secretlogic.BuildSnowflakePath(workingDir, meta.Env, meta.Warehouse, meta.Access)

	fileContent := map[string]interface{}{
		"metadata": map[string]string{
			"resourceType": "snowflake",
			"environment":  meta.Env,
			"warehouse":    meta.Warehouse,
			"access":       meta.Access,
		},
		"data": data,
	}

	fileJSON, err := json.MarshalIndent(fileContent, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(fileJSON), filename)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Info("Created file: %s", filename)
	fmt.Printf("✓ Successfully pulled secret: %s\n", secretID)
	fmt.Printf("  - %s\n", filename)

	return nil
}
