package secretlogic

import (
	"fmt"
	"strings"
)

// BuildJSONDocPaths builds file paths for jsondoc type
func BuildJSONDocPaths(workingDir, env, access string) (metadata, contents string) {
	base := fmt.Sprintf("%s/jsondoc.%s.%s", workingDir, env, access)
	return base + ".metadata.json", base + ".contents.json"
}

// BuildTextFilePaths builds file paths for textfile type
func BuildTextFilePaths(workingDir, env, access string) (metadata, contents string) {
	base := fmt.Sprintf("%s/textfile.%s.%s", workingDir, env, access)
	return base + ".metadata.json", base + ".contents.txt"
}

// BuildSSLCertPaths builds file paths for sslcert type
func BuildSSLCertPaths(workingDir, env, commonName string) (metadata, cert, key string) {
	base := fmt.Sprintf("%s/sslcert.%s.%s", workingDir, env, commonName)
	return base + ".metadata.json", base + ".crt", base + ".key"
}

// BuildRDSPostgresPath builds file path for rdspostgres type (single file)
func BuildRDSPostgresPath(workingDir, env, instance, database, access string) string {
	return fmt.Sprintf("%s/rdspostgres.%s.%s.%s.%s.json", workingDir, env, instance, database, access)
}

// BuildSnowflakePath builds file path for snowflake type (single file)
func BuildSnowflakePath(workingDir, env, warehouse, access string) string {
	return fmt.Sprintf("%s/snowflake.%s.%s.%s.json", workingDir, env, warehouse, access)
}

// GetBaseName extracts the base filename without .metadata.json suffix
// Preserves the directory path if present
func GetBaseName(metadataFile string) string {
	return strings.TrimSuffix(metadataFile, ".metadata.json")
}
