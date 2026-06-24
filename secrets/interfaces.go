package secrets

import (
	"context"

	"github.com/rs/zerolog"
)

// SecretsManager abstracts AWS Secrets Manager operations for testing
type SecretsManager interface {
	// DescribeSecret checks if a secret exists
	DescribeSecret(ctx context.Context, secretID string) (bool, error)

	// CreateSecret creates a new secret with tags
	CreateSecret(ctx context.Context, secretID string, value interface{}, tags map[string]string) error

	// UpdateSecret updates an existing secret
	UpdateSecret(ctx context.Context, secretID string, value interface{}) error

	// GetSecretValue retrieves a secret's value
	GetSecretValue(ctx context.Context, secretID string) (string, error)
}

// FileSystem abstracts filesystem operations for testing
type FileSystem interface {
	// GetWorkingDir returns the working directory path
	GetWorkingDir() (string, error)

	// ReadFile reads a file's contents
	ReadFile(path string) ([]byte, error)

	// WriteFile writes content to a file
	WriteFile(path string, content []byte) error

	// FileExists checks if a file exists
	FileExists(path string) bool

	// GetSHA256Sum computes SHA256 of a file
	GetSHA256Sum(path string) (string, error)
}

// Secret is the interface that all secret types must implement
type Secret interface {
	// SecretID returns the AWS Secrets Manager secret ID
	SecretID() string

	// Metadata returns the secret's metadata as a tag map
	Metadata() map[string]string

	// Data returns the secret's data payload
	Data() interface{}

	// Exists checks if the secret exists in AWS
	Exists(log *zerolog.Logger) bool

	// Create creates the secret in AWS with proper tags
	Create(log *zerolog.Logger) error

	// Update updates the secret in AWS
	Update(overwrite bool, log *zerolog.Logger) error
}

// SecretType represents a type of secret
type SecretType string

const (
	TypeJSONDoc        SecretType = "jsondoc"
	TypeTextField      SecretType = "text_file"
	TypeSSLCert        SecretType = "ssl_certificate"
	TypeRDSPostgres    SecretType = "rdspostgres"
	TypeSnowflake      SecretType = "snowflake"
)

// ValidSecretTypes returns all valid secret types
func ValidSecretTypes() []SecretType {
	return []SecretType{
		TypeJSONDoc,
		TypeTextField,
		TypeSSLCert,
		TypeRDSPostgres,
		TypeSnowflake,
	}
}

// ValidSecretTypeStrings returns valid secret types as strings
func ValidSecretTypeStrings() []string {
	types := ValidSecretTypes()
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}
