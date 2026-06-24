package push

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natemarks/secret-hoard/jsondoc"
	"github.com/natemarks/secret-hoard/rdspostgres"
	"github.com/natemarks/secret-hoard/snowflake"
	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/textfile"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// PushSecret uploads a secret from local files after showing diff and requiring confirmation
func PushSecret(metadataFile string, log *zerolog.Logger) error {
	// Read metadata file
	metadataJSON, err := tools.ReadFileToString(metadataFile)
	if err != nil {
		return fmt.Errorf("error reading metadata file: %w", err)
	}

	// Parse metadata to determine type
	var metadataMap map[string]interface{}
	err = json.Unmarshal([]byte(metadataJSON), &metadataMap)
	if err != nil {
		return fmt.Errorf("error parsing metadata JSON: %w", err)
	}

	resourceType, ok := metadataMap["resourceType"].(string)
	if !ok {
		return fmt.Errorf("metadata file missing resourceType field")
	}

	log.Debug().Msgf("Processing secret type: %s", resourceType)

	// Route to type-specific handler
	switch resourceType {
	case "jsondoc":
		return pushJSONDoc(metadataFile, metadataMap, log)
	case "text_file":
		return pushTextFile(metadataFile, metadataMap, log)
	case "ssl_certificate":
		return pushSSLCert(metadataFile, metadataMap, log)
	case "rdspostgres":
		return pushRDSPostgres(metadataFile, metadataMap, log)
	case "snowflake":
		return pushSnowflake(metadataFile, metadataMap, log)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

func pushJSONDoc(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
	// Extract metadata
	environment := metadataMap["environment"].(string)
	access := metadataMap["access"].(string)

	// Build file paths
	baseName := strings.TrimSuffix(metadataFile, ".metadata.json")
	contentsFile := baseName + ".contents.json"

	// Read contents file
	contents, err := tools.ReadFileToString(contentsFile)
	if err != nil {
		return fmt.Errorf("error reading contents file: %w", err)
	}

	// Compute SHA256
	sha256Sum, err := tools.GetSHA256Sum(contentsFile)
	if err != nil {
		return fmt.Errorf("error computing SHA256: %w", err)
	}

	// Build local secret
	localSecret := jsondoc.Secret{
		Metadata: jsondoc.Metadata{
			ResourceType: "jsondoc",
			Environment:  environment,
			Access:       access,
		},
		Data: jsondoc.Data{
			JSONContents:  contents,
			JSONSha256Sum: sha256Sum,
		},
	}

	secretID := localSecret.Metadata.SecretID()
	log.Info().Msgf("Secret ID: %s", secretID)

	// Check if secret exists
	fmt.Printf("Checking if secret exists: %s\n", secretID)
	if !localSecret.Exists(log) {
		// Secret doesn't exist - create it
		fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

		// Show what will be created
		fmt.Println("New secret contents:")
		secretJSON, _ := json.MarshalIndent(localSecret.Data, "", "  ")
		fmt.Println(string(secretJSON))
		fmt.Println()

		// Prompt for confirmation
		confirmString := GenerateRandomString(4)
		if !PromptForConfirmation(confirmString) {
			fmt.Println("Creation cancelled.")
			return nil
		}

		fmt.Println("Creating secret in AWS Secrets Manager...")
		localSecret.Create(log)
		log.Info().Msgf("Secret created: %s", secretID)
		fmt.Printf("✓ Secret created successfully: %s\n", secretID)
		return nil
	}

	// Secret exists - fetch and compare
	fmt.Println("Fetching current version from AWS...")
	remoteSecretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching remote secret: %w", err)
	}

	var remoteData jsondoc.Data
	err = json.Unmarshal([]byte(remoteSecretValue), &remoteData)
	if err != nil {
		return fmt.Errorf("error parsing remote secret: %w", err)
	}

	remoteSecret := jsondoc.Secret{
		Metadata: localSecret.Metadata,
		Data:     remoteData,
	}

	// Check if secrets are equal
	fmt.Println("Comparing local and remote versions...")
	if SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
	fmt.Println(diff)

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update secret
	fmt.Println("Updating secret in AWS Secrets Manager...")
	localSecret.Update(true, log)
	log.Info().Msgf("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}

func pushTextFile(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
	// Extract metadata
	environment := metadataMap["environment"].(string)
	access := metadataMap["access"].(string)

	// Build file paths
	baseName := strings.TrimSuffix(metadataFile, ".metadata.json")
	contentsFile := baseName + ".contents.txt"

	// Read contents file
	contents, err := tools.ReadFileToString(contentsFile)
	if err != nil {
		return fmt.Errorf("error reading contents file: %w", err)
	}

	// Compute SHA256
	sha256Sum, err := tools.GetSHA256Sum(contentsFile)
	if err != nil {
		return fmt.Errorf("error computing SHA256: %w", err)
	}

	// Build local secret
	localSecret := textfile.Secret{
		Metadata: textfile.Metadata{
			ResourceType: "text_file",
			Environment:  environment,
			Access:       access,
		},
		Data: textfile.Data{
			Contents:  contents,
			Sha256Sum: sha256Sum,
		},
	}

	secretID := localSecret.Metadata.SecretID()
	log.Info().Msgf("Secret ID: %s", secretID)

	// Check if secret exists
	fmt.Printf("Checking if secret exists: %s\n", secretID)
	if !localSecret.Exists(log) {
		// Secret doesn't exist - create it
		fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

		// Show what will be created
		fmt.Println("New secret contents:")
		secretJSON, _ := json.MarshalIndent(localSecret.Data, "", "  ")
		fmt.Println(string(secretJSON))
		fmt.Println()

		// Prompt for confirmation
		confirmString := GenerateRandomString(4)
		if !PromptForConfirmation(confirmString) {
			fmt.Println("Creation cancelled.")
			return nil
		}

		fmt.Println("Creating secret in AWS Secrets Manager...")
		localSecret.Create(log)
		log.Info().Msgf("Secret created: %s", secretID)
		fmt.Printf("✓ Secret created successfully: %s\n", secretID)
		return nil
	}

	// Fetch remote secret
	remoteSecretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching remote secret: %w", err)
	}

	var remoteData textfile.Data
	err = json.Unmarshal([]byte(remoteSecretValue), &remoteData)
	if err != nil {
		return fmt.Errorf("error parsing remote secret: %w", err)
	}

	remoteSecret := textfile.Secret{
		Metadata: localSecret.Metadata,
		Data:     remoteData,
	}

	// Check if secrets are equal
	fmt.Println("Comparing local and remote versions...")
	if SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
	fmt.Println(diff)

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update secret
	fmt.Println("Updating secret in AWS Secrets Manager...")
	localSecret.Update(true, log)
	log.Info().Msgf("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}

func pushSSLCert(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
	// Extract metadata
	environment := metadataMap["environment"].(string)
	commonName := metadataMap["commonName"].(string)

	// Build file paths
	baseName := strings.TrimSuffix(metadataFile, ".metadata.json")
	certFile := baseName + ".crt"
	keyFile := baseName + ".key"

	// Read certificate and key files
	certContents, err := tools.ReadFileToString(certFile)
	if err != nil {
		return fmt.Errorf("error reading certificate file: %w", err)
	}

	keyContents, err := tools.ReadFileToString(keyFile)
	if err != nil {
		return fmt.Errorf("error reading key file: %w", err)
	}

	// Extract expiration date
	expiration, err := extractExpiration(certFile)
	if err != nil {
		return fmt.Errorf("error extracting expiration: %w", err)
	}

	// Extract modulus from cert and key
	certModulus, err := extractCertificateModulus(certFile)
	if err != nil {
		return fmt.Errorf("error extracting certificate modulus: %w", err)
	}

	keyModulus, err := extractPrivateKeyModulus(keyFile)
	if err != nil {
		return fmt.Errorf("error extracting private key modulus: %w", err)
	}

	if certModulus != keyModulus {
		return fmt.Errorf("certificate and private key moduli do not match")
	}

	// Compute SHA256 sums
	certSha256, err := tools.GetSHA256Sum(certFile)
	if err != nil {
		return fmt.Errorf("error computing certificate SHA256: %w", err)
	}

	keySha256, err := tools.GetSHA256Sum(keyFile)
	if err != nil {
		return fmt.Errorf("error computing key SHA256: %w", err)
	}

	// Build local secret
	localSecret := sslcert.Secret{
		Metadata: sslcert.Metadata{
			ResourceType: "ssl_certificate",
			Environment:  environment,
			CommonName:   commonName,
		},
		Data: sslcert.Data{
			Certificate:       certContents,
			PrivateKey:        keyContents,
			ExpirationDate:    expiration,
			Modulus:           certModulus,
			CertificateSha256: certSha256,
			PrivateKeySha256:  keySha256,
		},
	}

	secretID := localSecret.Metadata.SecretID()
	log.Info().Msgf("Secret ID: %s", secretID)

	// Check if secret exists
	fmt.Printf("Checking if secret exists: %s\n", secretID)
	if !localSecret.Exists(log) {
		// Secret doesn't exist - create it
		fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

		// Show what will be created
		fmt.Println("New secret contents:")
		secretJSON, _ := json.MarshalIndent(localSecret.Data, "", "  ")
		fmt.Println(string(secretJSON))
		fmt.Println()

		// Prompt for confirmation
		confirmString := GenerateRandomString(4)
		if !PromptForConfirmation(confirmString) {
			fmt.Println("Creation cancelled.")
			return nil
		}

		fmt.Println("Creating secret in AWS Secrets Manager...")
		localSecret.Create(log)
		log.Info().Msgf("Secret created: %s", secretID)
		fmt.Printf("✓ Secret created successfully: %s\n", secretID)
		return nil
	}

	// Fetch remote secret
	remoteSecretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching remote secret: %w", err)
	}

	var remoteData sslcert.Data
	err = json.Unmarshal([]byte(remoteSecretValue), &remoteData)
	if err != nil {
		return fmt.Errorf("error parsing remote secret: %w", err)
	}

	remoteSecret := sslcert.Secret{
		Metadata: localSecret.Metadata,
		Data:     remoteData,
	}

	// Check if secrets are equal
	fmt.Println("Comparing local and remote versions...")
	if SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
	fmt.Println(diff)

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update secret
	fmt.Println("Updating secret in AWS Secrets Manager...")
	localSecret.Update(true, log)
	log.Info().Msgf("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}

func pushRDSPostgres(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
	// For single-file types, the metadata file IS the complete file
	fileJSON, err := tools.ReadFileToString(metadataFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var fileData struct {
		Metadata rdspostgres.Metadata `json:"metadata"`
		Data     rdspostgres.Data     `json:"data"`
	}

	err = json.Unmarshal([]byte(fileJSON), &fileData)
	if err != nil {
		return fmt.Errorf("error parsing file: %w", err)
	}

	localSecret := rdspostgres.Secret{
		Metadata: fileData.Metadata,
		Data:     fileData.Data,
	}

	secretID := localSecret.Metadata.SecretID()
	log.Info().Msgf("Secret ID: %s", secretID)

	// Check if secret exists
	fmt.Printf("Checking if secret exists: %s\n", secretID)
	if !localSecret.Exists(log) {
		// Secret doesn't exist - create it
		fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

		// Show what will be created
		fmt.Println("New secret contents:")
		secretJSON, _ := json.MarshalIndent(localSecret.Data, "", "  ")
		fmt.Println(string(secretJSON))
		fmt.Println()

		// Prompt for confirmation
		confirmString := GenerateRandomString(4)
		if !PromptForConfirmation(confirmString) {
			fmt.Println("Creation cancelled.")
			return nil
		}

		fmt.Println("Creating secret in AWS Secrets Manager...")
		localSecret.Create(log)
		log.Info().Msgf("Secret created: %s", secretID)
		fmt.Printf("✓ Secret created successfully: %s\n", secretID)
		return nil
	}

	// Fetch remote secret
	remoteSecretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching remote secret: %w", err)
	}

	var remoteData rdspostgres.Data
	err = json.Unmarshal([]byte(remoteSecretValue), &remoteData)
	if err != nil {
		return fmt.Errorf("error parsing remote secret: %w", err)
	}

	remoteSecret := rdspostgres.Secret{
		Metadata: localSecret.Metadata,
		Data:     remoteData,
	}

	// Check if secrets are equal
	fmt.Println("Comparing local and remote versions...")
	if SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
	fmt.Println(diff)

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update secret
	fmt.Println("Updating secret in AWS Secrets Manager...")
	localSecret.Update(true, log)
	log.Info().Msgf("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}

func pushSnowflake(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
	// For single-file types, the metadata file IS the complete file
	fileJSON, err := tools.ReadFileToString(metadataFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var fileData struct {
		Metadata snowflake.Metadata `json:"metadata"`
		Data     snowflake.Data     `json:"data"`
	}

	err = json.Unmarshal([]byte(fileJSON), &fileData)
	if err != nil {
		return fmt.Errorf("error parsing file: %w", err)
	}

	localSecret := snowflake.Secret{
		Metadata: fileData.Metadata,
		Data:     fileData.Data,
	}

	secretID := localSecret.Metadata.SecretID()
	log.Info().Msgf("Secret ID: %s", secretID)

	// Check if secret exists
	fmt.Printf("Checking if secret exists: %s\n", secretID)
	if !localSecret.Exists(log) {
		// Secret doesn't exist - create it
		fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

		// Show what will be created
		fmt.Println("New secret contents:")
		secretJSON, _ := json.MarshalIndent(localSecret.Data, "", "  ")
		fmt.Println(string(secretJSON))
		fmt.Println()

		// Prompt for confirmation
		confirmString := GenerateRandomString(4)
		if !PromptForConfirmation(confirmString) {
			fmt.Println("Creation cancelled.")
			return nil
		}

		fmt.Println("Creating secret in AWS Secrets Manager...")
		localSecret.Create(log)
		log.Info().Msgf("Secret created: %s", secretID)
		fmt.Printf("✓ Secret created successfully: %s\n", secretID)
		return nil
	}

	// Fetch remote secret
	remoteSecretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		return fmt.Errorf("error fetching remote secret: %w", err)
	}

	var remoteData snowflake.Data
	err = json.Unmarshal([]byte(remoteSecretValue), &remoteData)
	if err != nil {
		return fmt.Errorf("error parsing remote secret: %w", err)
	}

	remoteSecret := snowflake.Secret{
		Metadata: localSecret.Metadata,
		Data:     remoteData,
	}

	// Check if secrets are equal
	fmt.Println("Comparing local and remote versions...")
	if SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
	fmt.Println(diff)

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Update secret
	fmt.Println("Updating secret in AWS Secrets Manager...")
	localSecret.Update(true, log)
	log.Info().Msgf("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}

// Helper functions for SSL certificate processing

func extractExpiration(certFile string) (string, error) {
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	return cert.NotAfter.Format(time.RFC3339), nil
}

func extractCertificateModulus(certFile string) (string, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate public key is not RSA")
	}

	return rsaPublicKey.N.String(), nil
}

func extractPrivateKeyModulus(keyFile string) (string, error) {
	privateKeyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}

	return rsaPrivateKey.N.String(), nil
}

// GetBaseName extracts the base name from a metadata file path
func GetBaseName(metadataFile string) string {
	base := filepath.Base(metadataFile)
	return strings.TrimSuffix(base, ".metadata.json")
}
