package secretlogic

import (
	"fmt"
	"strings"
)

// SecretIDParts represents the parsed components of a secret ID
type SecretIDParts struct {
	Type       string
	Env        string
	Access     string
	Instance   string
	Database   string
	Warehouse  string
	CommonName string
}

// NormalizeSecretID converts user-friendly secret type to AWS format
// Examples: textfile -> text_file, sslcert -> ssl_certificate
func NormalizeSecretID(secretID string) string {
	secretID = strings.Replace(secretID, "textfile/", "text_file/", 1)
	secretID = strings.Replace(secretID, "sslcert/", "ssl_certificate/", 1)
	return secretID
}

// ParseJSONDocSecretID parses a jsondoc secret ID
// Format: jsondoc/<env>/<access>
func ParseJSONDocSecretID(secretID string) (env, access string, err error) {
	parts := strings.Split(secretID, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid jsondoc secret ID: %s\n"+
			"Expected format: jsondoc/<env>/<access>\n"+
			"  <env>: Environment (e.g., dev, prod, staging)\n"+
			"  <access>: Access identifier (e.g., app-config, db-creds)\n"+
			"Example: jsondoc/dev/app-config", secretID)
	}
	return parts[1], parts[2], nil
}

// ParseTextFileSecretID parses a text_file secret ID
// Format: text_file/<env>/<access> or textfile/<env>/<access>
func ParseTextFileSecretID(secretID string) (env, access string, err error) {
	// Normalize first
	secretID = NormalizeSecretID(secretID)

	parts := strings.Split(secretID, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid textfile secret ID: %s\n"+
			"Expected format: textfile/<env>/<access>\n"+
			"  <env>: Environment (e.g., dev, prod, staging)\n"+
			"  <access>: Access identifier (e.g., api-key, token)\n"+
			"Example: textfile/prod/api-key", secretID)
	}
	return parts[1], parts[2], nil
}

// ParseSSLCertSecretID parses an ssl_certificate secret ID
// Format: ssl_certificate/<env>/<commonName> or sslcert/<env>/<commonName>
func ParseSSLCertSecretID(secretID string) (env, commonName string, err error) {
	// Normalize first
	secretID = NormalizeSecretID(secretID)

	parts := strings.Split(secretID, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid sslcert secret ID: %s\n"+
			"Expected format: sslcert/<env>/<commonName>\n"+
			"  <env>: Environment (e.g., dev, prod, staging)\n"+
			"  <commonName>: Domain name (e.g., example.com, *.example.com)\n"+
			"Example: sslcert/prod/example.com", secretID)
	}
	return parts[1], parts[2], nil
}

// BuildContentsFilename builds the filename for a secret's contents file
func BuildContentsFilename(secretType, env, access, commonName string) string {
	switch secretType {
	case "jsondoc":
		return fmt.Sprintf("jsondoc.%s.%s.contents.json", env, access)
	case "textfile", "text_file":
		return fmt.Sprintf("textfile.%s.%s.contents.txt", env, access)
	case "sslcert", "ssl_certificate":
		return fmt.Sprintf("sslcert.%s.%s", env, commonName) // Base name without extension
	default:
		return ""
	}
}

// BuildSSLCertFilenames builds both certificate and key filenames
func BuildSSLCertFilenames(env, commonName string) (certFilename, keyFilename string) {
	base := fmt.Sprintf("sslcert.%s.%s", env, commonName)
	return base + ".crt", base + ".key"
}
