package secretlogic

import (
	"fmt"
)

// BuildSecretID constructs the AWS Secrets Manager secret ID from type and metadata
// This is a pure function with no I/O dependencies
func BuildSecretID(secretType string, metadata map[string]string) (string, error) {
	env, ok := metadata["environment"]
	if !ok || env == "" {
		return "", fmt.Errorf("missing required field: environment")
	}

	switch secretType {
	case "jsondoc":
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("jsondoc missing required field: access")
		}
		return fmt.Sprintf("jsondoc/%s/%s", env, access), nil

	case "text_file":
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("text_file missing required field: access")
		}
		return fmt.Sprintf("text_file/%s/%s", env, access), nil

	case "ssl_certificate":
		commonName, ok := metadata["commonname"]
		if !ok || commonName == "" {
			return "", fmt.Errorf("ssl_certificate missing required field: commonname")
		}
		return fmt.Sprintf("ssl_certificate/%s/%s", env, commonName), nil

	case "snowflake":
		warehouse, ok := metadata["warehouse"]
		if !ok || warehouse == "" {
			return "", fmt.Errorf("snowflake missing required field: warehouse")
		}
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("snowflake missing required field: access")
		}
		return fmt.Sprintf("snowflake/%s/%s/%s", env, warehouse, access), nil

	case "rdspostgres":
		instance, ok := metadata["instance"]
		if !ok || instance == "" {
			return "", fmt.Errorf("rdspostgres missing required field: instance")
		}
		database, ok := metadata["database"]
		if !ok || database == "" {
			return "", fmt.Errorf("rdspostgres missing required field: database")
		}
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("rdspostgres missing required field: access")
		}
		return fmt.Sprintf("rdspostgres/%s/%s/%s/%s", env, instance, database, access), nil

	default:
		return "", fmt.Errorf("unknown secret type: %s", secretType)
	}
}
