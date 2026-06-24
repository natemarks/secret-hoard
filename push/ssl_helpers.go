package push

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/natemarks/secret-hoard/secretlogic"
	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/tools"
)

// sslCertFiles holds the paths to SSL certificate files
type sslCertFiles struct {
	metadataFile string
	certFile     string
	keyFile      string
}

// buildSSLCertFilePaths constructs file paths from metadata file
func buildSSLCertFilePaths(metadataFile string) sslCertFiles {
	baseName := strings.TrimSuffix(metadataFile, ".metadata.json")
	return sslCertFiles{
		metadataFile: metadataFile,
		certFile:     baseName + ".crt",
		keyFile:      baseName + ".key",
	}
}

// readSSLCertFiles reads certificate and key file contents
func readSSLCertFiles(files sslCertFiles) (certContents, keyContents string, err error) {
	certContents, err = tools.ReadFileToString(files.certFile)
	if err != nil {
		return "", "", fmt.Errorf("error reading certificate file: %w", err)
	}

	keyContents, err = tools.ReadFileToString(files.keyFile)
	if err != nil {
		return "", "", fmt.Errorf("error reading key file: %w", err)
	}

	return certContents, keyContents, nil
}

// validateSSLCertPair validates that cert and key match
func validateSSLCertPair(certFile, keyFile string) (modulus string, err error) {
	certModulus, err := extractCertificateModulus(certFile)
	if err != nil {
		return "", fmt.Errorf("error extracting certificate modulus: %w", err)
	}

	keyModulus, err := extractPrivateKeyModulus(keyFile)
	if err != nil {
		return "", fmt.Errorf("error extracting private key modulus: %w", err)
	}

	if certModulus != keyModulus {
		return "", fmt.Errorf("certificate and private key moduli do not match")
	}

	return certModulus, nil
}

// computeSSLCertHashes computes SHA256 sums for cert and key
func computeSSLCertHashes(certFile, keyFile string) (certSha256, keySha256 string, err error) {
	certSha256, err = tools.GetSHA256Sum(certFile)
	if err != nil {
		return "", "", fmt.Errorf("error computing certificate SHA256: %w", err)
	}

	keySha256, err = tools.GetSHA256Sum(keyFile)
	if err != nil {
		return "", "", fmt.Errorf("error computing key SHA256: %w", err)
	}

	return certSha256, keySha256, nil
}

// buildSSLCertSecret constructs the complete SSL certificate secret
func buildSSLCertSecret(
	environment, commonName string,
	certContents, keyContents string,
	expiration, modulus string,
	certSha256, keySha256 string,
) sslcert.Secret {
	return sslcert.Secret{
		Metadata: sslcert.Metadata{
			ResourceType: "ssl_certificate",
			Environment:  environment,
			CommonName:   commonName,
		},
		Data: sslcert.Data{
			Certificate:       certContents,
			PrivateKey:        keyContents,
			ExpirationDate:    expiration,
			Modulus:           modulus,
			CertificateSha256: certSha256,
			PrivateKeySha256:  keySha256,
		},
	}
}

// handleSSLCertCreate handles creation of a new SSL certificate secret
func handleSSLCertCreate(secret sslcert.Secret, log *tools.Logger) error {
	secretID := secret.Metadata.SecretID()

	fmt.Printf("\nSecret does not exist. Creating new secret: %s\n\n", secretID)

	// Show what will be created
	fmt.Println("New secret contents:")
	secretJSON, _ := json.MarshalIndent(secret.Data, "", "  ")
	fmt.Println(string(secretJSON))
	fmt.Println()

	// Prompt for confirmation
	confirmString := GenerateRandomString(4)
	if !PromptForConfirmation(confirmString) {
		fmt.Println("Creation cancelled.")
		return nil
	}

	fmt.Println("Creating secret in AWS Secrets Manager...")
	secret.Create(log)
	log.Info("Secret created: %s", secretID)
	fmt.Printf("✓ Secret created successfully: %s\n", secretID)

	return nil
}

// handleSSLCertUpdate handles update of an existing SSL certificate secret
func handleSSLCertUpdate(localSecret sslcert.Secret, log *tools.Logger) error {
	secretID := localSecret.Metadata.SecretID()

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
	if secretlogic.SecretsAreEqual(localSecret.Data, remoteSecret.Data) {
		fmt.Println("✓ Local and remote secrets are identical. No update needed.")
		return nil
	}

	// Generate and display diff
	fmt.Printf("\nChanges detected in secret: %s\n\n", secretID)
	diff := secretlogic.GenerateJSONDiff(localSecret.Data, remoteSecret.Data)
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
	log.Info("Secret updated: %s", secretID)
	fmt.Printf("✓ Secret updated successfully: %s\n", secretID)

	return nil
}
