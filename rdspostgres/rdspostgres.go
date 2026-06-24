package rdspostgres

import (
	"fmt"

	"github.com/natemarks/secret-hoard/secrets"
	"github.com/natemarks/secret-hoard/tools"
)
// Metadata RDS secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // rdspostgres
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Instance     string `json:"instance"`     // some_instance
	Database     string `json:"database"`     // some_database
	Access       string `json:"access"`       // master, monitoring, app_readwrite, app_readonly
}

// Map converts Metadata to a map of strings to simplify tagging
func (rm Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": rm.ResourceType,
		"Environment":  rm.Environment,
		"Instance":     rm.Instance,
		"Database":     rm.Database,
		"Access":       rm.Access,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id for the rdsSecret
func (rm Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v/%v/%v", rm.ResourceType, rm.Environment, rm.Instance, rm.Database, rm.Access)
}

// Data is the struct of the secret generated for RDS by CDK deployment
// Password: the password for the database user
// Engine: the database engine
// Port: the port the database is listening on
// DbInstanceIdentifier: the unique name of the RDS instance
// Host: the hostname of the RDS instance
// Username: the username for the database user
type Data struct {
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Port                 int    `json:"port"`
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"`
	Host                 string `json:"host"`
	Username             string `json:"username"`
}

// Secret is the struct of the secret generated for RDS by CDK deployment
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
