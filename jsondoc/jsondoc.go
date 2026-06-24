package jsondoc

import (
	"fmt"

	"github.com/natemarks/secret-hoard/secrets"
	"github.com/natemarks/secret-hoard/tools"
)

// Metadata server certificate secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // json_document
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Access       string `json:"access"`       // access type provides by the secret
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (m Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": m.ResourceType,
		"Environment":  m.Environment,
		"Access":       m.Access,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id for the secret
func (m Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v", m.ResourceType, m.Environment, m.Access)
}

// Data is the struct of the secret for s snowflake connection
type Data struct {
	JSONContents  string `json:"JSONContents"`  // contents as a string
	JSONSha256Sum string `json:"JSONSha256Sum"` // sha256sum of original JSON file
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

func (sa secretAdapter) SecretID() string                               { return sa.s.Metadata.SecretID() }
func (sa secretAdapter) Metadata() map[string]string                    { return sa.s.Metadata.Map() }
func (sa secretAdapter) Data() any                                      { return sa.s.Data }
func (sa secretAdapter) Exists(log *tools.Logger) bool                  { return false } // Not used
func (sa secretAdapter) Create(log *tools.Logger) error                 { return nil }   // Not used
func (sa secretAdapter) Update(overwrite bool, log *tools.Logger) error { return nil }   // Not used

// Exists checks if the secret exists in Secrets Manager
func (s Secret) Exists(log *tools.Logger) bool {
	return secrets.GenericExists(secretAdapter{s}, log, secrets.DefaultSecretsManager())
}

// Create the Secret
func (s Secret) Create(log *tools.Logger) {
	log.Debug("creating jsondoc secret: %s", s.Metadata.SecretID())
	err := secrets.GenericCreate(secretAdapter{s}, log, secrets.DefaultSecretsManager())
	if err != nil {
		log.Error("error creating jsondoc secret: %s - %v", s.Metadata.SecretID(), err)
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
