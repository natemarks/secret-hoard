package secretlogic

import (
	"fmt"
	"strings"
)

// BuildBaseFilename creates the base filename (without extension) from secret type and metadata
// Example: "jsondoc.dev.app" or "rdspostgres.prod.db1.myapp.readonly"
func BuildBaseFilename(secretType string, metadata map[string]string) (string, error) {
	parts := []string{secretType}

	env, ok := metadata["environment"]
	if !ok || env == "" {
		return "", fmt.Errorf("missing required field: environment")
	}
	parts = append(parts, env)

	// Type-specific fields in order
	switch secretType {
	case "jsondoc", "text_file":
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("%s missing required field: access", secretType)
		}
		parts = append(parts, access)

	case "snowflake":
		warehouse, ok := metadata["warehouse"]
		if !ok || warehouse == "" {
			return "", fmt.Errorf("snowflake missing required field: warehouse")
		}
		access, ok := metadata["access"]
		if !ok || access == "" {
			return "", fmt.Errorf("snowflake missing required field: access")
		}
		parts = append(parts, warehouse, access)

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
		parts = append(parts, instance, database, access)

	case "ssl_certificate":
		commonName, ok := metadata["commonname"]
		if !ok || commonName == "" {
			return "", fmt.Errorf("ssl_certificate missing required field: commonname")
		}
		parts = append(parts, commonName)

	default:
		return "", fmt.Errorf("unknown secret type: %s", secretType)
	}

	return strings.Join(parts, "."), nil
}

// FilesToGenerate returns the list of filenames that should be created for a secret type
// Returns relative filenames without directory path
func FilesToGenerate(secretType string, metadata map[string]string) ([]string, error) {
	base, err := BuildBaseFilename(secretType, metadata)
	if err != nil {
		return nil, err
	}

	switch secretType {
	case "jsondoc":
		return []string{
			base + ".metadata.json",
			base + ".contents.json",
		}, nil

	case "text_file":
		return []string{
			base + ".metadata.json",
			base + ".contents.txt",
		}, nil

	case "ssl_certificate":
		return []string{
			base + ".metadata.json",
			base + ".crt",
			base + ".key",
		}, nil

	case "rdspostgres", "snowflake":
		return []string{
			base + ".json",
		}, nil

	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
}
