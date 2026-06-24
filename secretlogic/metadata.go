package secretlogic

import (
	"fmt"
)

// BuildMetadataMap creates the metadata structure for a secret type
// Returns a map suitable for JSON marshaling
func BuildMetadataMap(secretType string, metadata map[string]string) (map[string]string, error) {
	result := map[string]string{
		"resourceType": secretType,
	}

	env, ok := metadata["environment"]
	if !ok || env == "" {
		return nil, fmt.Errorf("missing required field: environment")
	}
	result["environment"] = env

	switch secretType {
	case "jsondoc":
		access, ok := metadata["access"]
		if !ok || access == "" {
			return nil, fmt.Errorf("jsondoc missing required field: access")
		}
		result["access"] = access
		// JSONSha256Sum will be set by caller or marked as REPLACE-ME

	case "text_file":
		access, ok := metadata["access"]
		if !ok || access == "" {
			return nil, fmt.Errorf("text_file missing required field: access")
		}
		result["access"] = access
		// sha256Sum will be set by caller or marked as REPLACE-ME

	case "ssl_certificate":
		commonName, ok := metadata["commonname"]
		if !ok || commonName == "" {
			return nil, fmt.Errorf("ssl_certificate missing required field: commonname")
		}
		result["commonName"] = commonName
		// Computed fields (expiration, modulus, hashes) will be set by caller or REPLACE-ME

	case "snowflake":
		warehouse, ok := metadata["warehouse"]
		if !ok || warehouse == "" {
			return nil, fmt.Errorf("snowflake missing required field: warehouse")
		}
		access, ok := metadata["access"]
		if !ok || access == "" {
			return nil, fmt.Errorf("snowflake missing required field: access")
		}
		result["warehouse"] = warehouse
		result["access"] = access

	case "rdspostgres":
		instance, ok := metadata["instance"]
		if !ok || instance == "" {
			return nil, fmt.Errorf("rdspostgres missing required field: instance")
		}
		database, ok := metadata["database"]
		if !ok || database == "" {
			return nil, fmt.Errorf("rdspostgres missing required field: database")
		}
		access, ok := metadata["access"]
		if !ok || access == "" {
			return nil, fmt.Errorf("rdspostgres missing required field: access")
		}
		result["instance"] = instance
		result["database"] = database
		result["access"] = access

	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}

	return result, nil
}

// RequiredMetadataFields returns the list of required metadata field names for a secret type
func RequiredMetadataFields(secretType string) ([]string, error) {
	switch secretType {
	case "jsondoc", "text_file":
		return []string{"environment", "access"}, nil
	case "ssl_certificate":
		return []string{"environment", "commonname"}, nil
	case "snowflake":
		return []string{"environment", "warehouse", "access"}, nil
	case "rdspostgres":
		return []string{"environment", "instance", "database", "access"}, nil
	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
}
