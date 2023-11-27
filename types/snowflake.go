package types

import "fmt"

// SnowflakeSecretMetadata Snowflake secret metadata for tagging
type SnowflakeSecretMetadata struct {
	ResourceType string `json:"resourceType"` // snowflake
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Warehouse    string `json:"warehouse"`    // some_warehouse
	Access       string `json:"access"`       // readwrite, admin
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (sfm SnowflakeSecretMetadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": sfm.ResourceType,
		"Environment":  sfm.Environment,
		"Warehouse":    sfm.Warehouse,
		"Access":       sfm.Access,
	}
	return attributes
}

// SecretID returns the secret id for the secret
func (sfm SnowflakeSecretMetadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v/%v", sfm.ResourceType, sfm.Environment, sfm.Warehouse, sfm.Access)
}

// SnowflakeSecretData is the struct of the secret for s snowflake connection
// Password: the password for the database user
// AccountName: the database engine
// Warehouse: the port the database is listening on
// Username: the username for the database user
type SnowflakeSecretData struct {
	Password    string `json:"password"`
	AccountName string `json:"accountName"`
	Warehouse   string `json:"warehouse"`
	Username    string `json:"username"`
}

// SnowflakeSecret is the struct of the secret for snowflake
type SnowflakeSecret struct {
	Data     SnowflakeSecretData
	Metadata SnowflakeSecretMetadata
}
