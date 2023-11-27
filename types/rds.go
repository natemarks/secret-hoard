package types

import "fmt"

// RDSSecretMetadata RDS secret metadata for tagging
type RDSSecretMetadata struct {
	ResourceType string `json:"resourceType"` // rdspostgres
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Instance     string `json:"instance"`     // some_instance
	Database     string `json:"database"`     // some_database
	Access       string `json:"access"`       // master, monitoring, app_readwrite, app_readonly
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (rm RDSSecretMetadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": rm.ResourceType,
		"Environment":  rm.Environment,
		"Instance":     rm.Instance,
		"Database":     rm.Database,
		"Access":       rm.Access,
	}
	return attributes
}

// SecretID returns the secret id for the secret
func (rm RDSSecretMetadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v/%v/%v", rm.ResourceType, rm.Environment, rm.Instance, rm.Database, rm.Access)
}

// RDSSecretData is the struct of the secret generated for RDS by CDK deployment
// Password: the password for the database user
// Engine: the database engine
// Port: the port the database is listening on
// DbInstanceIdentifier: the unique name of the RDS instance
// Host: the hostname of the RDS instance
// Username: the username for the database user
type RDSSecretData struct {
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Port                 int    `json:"port"`
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"`
	Host                 string `json:"host"`
	Username             string `json:"username"`
}

// RDSSecret is the struct of the secret generated for RDS by CDK deployment
type RDSSecret struct {
	Data     RDSSecretData
	Metadata RDSSecretMetadata
}
