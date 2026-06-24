package sslcert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/natemarks/secret-hoard/tools"


)

// Metadata server certificate secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // ssl_certificate
	Environment  string `json:"environment"`  // dev, integration, staging, production
	CommonName   string `json:"commonName"`   // \*.my.domain.com | server.my.domain.com
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (sfm Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": sfm.ResourceType,
		"Environment":  sfm.Environment,
		"CommonName":   sfm.CommonName,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id
func (sfm Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v", sfm.ResourceType, sfm.Environment, sfm.CommonName)
}

// Data is the struct of the secret for s snowflake connection
type Data struct {
	Certificate       string `json:"certificate"`       // Certificate file contents as string
	PrivateKey        string `json:"key"`               // PrivateKey file contents as string
	ExpirationDate    string `json:"expirationDate"`    // Expiration date in ISO 3339 format
	Modulus           string `json:"modulus"`           // Modulus of the certificate and PrivateKey
	CertificateSha256 string `json:"certificateSha256"` // SHA256 hash of the certificate file
	PrivateKeySha256  string `json:"privateKeySha256"`  // SHA256 hash of the PrivateKey file
}

// Secret is the struct of the secret for snowflake
type Secret struct {
	Data     Data
	Metadata Metadata
}

// Exists checks if the secret exists in Secrets Manager
func (s Secret) Exists(log *tools.Logger) bool {

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("unable to load SDK config")
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Input parameters for DescribeSecret API call
	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(s.Metadata.SecretID()),
	}

	// Call DescribeSecret API to check if the secret exists
	_, err = client.DescribeSecret(context.Background(), input)
	if err != nil {
		var e *types.ResourceNotFoundException
		if errors.As(err, &e) {
			log.Debug("secret does not exist: %s", *input.SecretId)
			return false
		}
	}
	log.Debug("secret exists: %s", *input.SecretId)
	return true
}

// Create the Secret
func (s Secret) Create(log *tools.Logger) {
	log.Debug("creating ssl certificate secret: %s", s.Metadata.SecretID())
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(s.Data)
	if err != nil {
		log.Error("error marshalling secret data")
		return
	}

	// Convert RDSSecretMetadata to tags
	tags := s.Metadata.Map()

	// Create the secret
	createSecretInput := &secretsmanager.CreateSecretInput{
		Name:         aws.String(fmt.Sprint(s.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
		Tags:         tools.ConvertMapToTags(tags),
	}
	_, err = client.CreateSecret(ctx, createSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		log.Error("error creating ssl certificate secret: %s", *createSecretInput.Name)
		return
	}
	log.Info("secret created successfully: %s", *createSecretInput.Name)
}

// Update the secret
func (s Secret) Update(overwrite bool, log *tools.Logger) {
	if !overwrite {
		log.Debug("overwrite is false, skipping update for %s", s.Metadata.SecretID())
		return
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(s.Data)
	if err != nil {
		log.Error("error marshalling secret data")
		return
	}

	// Convert RDSSecretMetadata to tags
	tags := s.Metadata.Map()

	// Create the secret
	// Update the secret string value
	updateSecretInput := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(fmt.Sprint(s.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
	}
	_, err = client.UpdateSecret(ctx, updateSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		log.Error("error updating secret value: %s", *updateSecretInput.SecretId)
		return
	}

	// Update the secret tags
	tagResourceInput := &secretsmanager.TagResourceInput{
		SecretId: aws.String(fmt.Sprint(s.Metadata.SecretID())),
		Tags:     tools.ConvertMapToTags(tags),
	}
	_, err = client.TagResource(ctx, tagResourceInput)
	if err != nil {
		log.Error("error updating secret tags: %s", *updateSecretInput.SecretId)
		return
	}
	log.Info("secret update successfully: %s", *updateSecretInput.SecretId)
}

