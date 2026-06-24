package sslcert

import (
	"fmt"

	"github.com/natemarks/secret-hoard/secrets"
	"github.com/natemarks/secret-hoard/tools"
)
// Metadata server certificate secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // ssl_certificate
	Environment  string `json:"environment"`  // dev, integration, staging, production
	CommonName   string `json:"commonName"`   // \*.my.domain.com | server.my.domain.com
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (sfm Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": sfm.ResourceType,
		"Environment":  sfm.Environment,
		"CommonName":   sfm.CommonName,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id
func (sfm Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v", sfm.ResourceType, sfm.Environment, sfm.CommonName)
}

// Data is the struct of the secret for s snowflake connection
type Data struct {
	Certificate       string `json:"certificate"`       // Certificate file contents as string
	PrivateKey        string `json:"key"`               // PrivateKey file contents as string
	ExpirationDate    string `json:"expirationDate"`    // Expiration date in ISO 3339 format
	Modulus           string `json:"modulus"`           // Modulus of the certificate and PrivateKey
	CertificateSha256 string `json:"certificateSha256"` // SHA256 hash of the certificate file
	PrivateKeySha256  string `json:"privateKeySha256"`  // SHA256 hash of the PrivateKey file
}

// Secret is the struct of the secret for snowflake
type Secret struct {
	Data     Data
	Metadata Metadata
}


// secretAdapter adapts Secret to implement secrets.Secret interface
type secretAdapter struct {
	s Secret
}

func (sa secretAdapter) SecretID() string              { return sa.s.Metadata.SecretID() }
func (sa secretAdapter) Metadata() map[string]string   { return sa.s.Metadata.Map() }
func (sa secretAdapter) Data() any                     { return sa.s.Data }
func (sa secretAdapter) Exists(log *tools.Logger) bool { return false }
func (sa secretAdapter) Create(log *tools.Logger) error { return nil }
func (sa secretAdapter) Update(overwrite bool, log *tools.Logger) error { return nil }

// Exists checks if the secret exists in Secrets Manager
func (s Secret) Exists(log *tools.Logger) bool {
	return secrets.GenericExists(secretAdapter{s}, log, secrets.DefaultSecretsManager())
}

// Create the Secret
func (s Secret) Create(log *tools.Logger) {
	log.Debug("creating secret: %s", s.Metadata.SecretID())
	err := secrets.GenericCreate(secretAdapter{s}, log, secrets.DefaultSecretsManager())
	if err != nil {
		log.Error("error creating secret: %s - %v", s.Metadata.SecretID(), err)
		return
	}
	log.Info("secret created successfully: %s", s.Metadata.SecretID())
}

// Update the secret
func (s Secret) Update(overwrite bool, log *tools.Logger) {
	if !overwrite {
		log.Debug("overwrite is false, skipping update for %s", s.Metadata.SecretID())
		return
	}
	err := secrets.GenericUpdate(secretAdapter{s}, log, secrets.DefaultSecretsManager(), overwrite)
	if err != nil {
		log.Error("error updating secret: %s - %v", s.Metadata.SecretID(), err)
		return
	}
	log.Info("secret updated successfully: %s", s.Metadata.SecretID())
}
