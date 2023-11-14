package types

// RDSSecretMetadata RDS secret metadata for tagging
// Environment: sandbox, dev, integration, staging, production
// Instance: the unique name of the RDS instance""
// Database: the name of the database in the instance, empty is used for master and monitoring
// Type: the type of secret (master, monitoring, app_readwrite, app_readonly)
type RDSSecretMetadata struct {
	Environment string `json:"environment"`
	Instance    string `json:"instance"`
	Database    string `json:"database"`
	Type        string `json:"type"`
}

// ToMap converts RDSSecretMetadata to a map of strings to simplify tagging
func (rm RDSSecretMetadata) ToMap() map[string]string {
	attributes := map[string]string{
		"Environment": rm.Environment,
		"Instance":    rm.Instance,
		"Database":    rm.Database,
		"Type":        rm.Database,
	}
	return attributes
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

type RDSSecret struct {
	Data     RDSSecretData
	Metadata RDSSecretMetadata
}
