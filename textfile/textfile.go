package textfile

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

	"github.com/rs/zerolog"
)

// Metadata server certificate secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // json_document
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Access       string `json:"access"`       // access type provides by the secret
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (m Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": m.ResourceType,
		"Environment":  m.Environment,
		"Access":       m.Access,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id for the secret
func (m Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v", m.ResourceType, m.Environment, m.Access)
}

// Data is the struct of the secret for s snowflake connection
type Data struct {
	Contents  string `json:"contents"`  // contents as a string
	Sha256Sum string `json:"sha256Sum"` // sha256sum of original JSON file
}

// Secret is the struct of the secret for snowflake
type Secret struct {
	Data     Data
	Metadata Metadata
}

// Exists checks if the secret exists in Secrets Manager
func (s Secret) Exists(log *zerolog.Logger) bool {

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load SDK config")
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
			log.Debug().Msgf("secret does not exist: %s", *input.SecretId)
			return false
		}
	}
	log.Debug().Msgf("secret exists: %s", *input.SecretId)
	return true
}

// Create the Secret
func (s Secret) Create(log *zerolog.Logger) {
	log.Debug().Msgf("creating text file secret: %s", s.Metadata.SecretID())
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(s.Data)
	if err != nil {
		log.Error().Err(err).Msg("error marshalling secret data")
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
		log.Error().Err(err).Msgf("error creating text file secret: %s", *createSecretInput.Name)
		return
	}
	log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
}

// Update the secret
func (s Secret) Update(overwrite bool, log *zerolog.Logger) {
	if !overwrite {
		log.Debug().Msgf("overwrite is false, skipping update for %s", s.Metadata.SecretID())
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
		log.Error().Err(err).Msg("error marshalling secret data")
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
		log.Error().Err(err).Msgf("error updating secret value: %s", *updateSecretInput.SecretId)
		return
	}

	// Update the secret tags
	tagResourceInput := &secretsmanager.TagResourceInput{
		SecretId: aws.String(fmt.Sprint(s.Metadata.SecretID())),
		Tags:     tools.ConvertMapToTags(tags),
	}
	_, err = client.TagResource(ctx, tagResourceInput)
	if err != nil {
		log.Error().Err(err).Msgf("error updating secret tags: %s", *updateSecretInput.SecretId)
		return
	}
	log.Info().Msgf("secret update successfully: %s", *updateSecretInput.SecretId)
}

// FromCSVRecord converts a CSV record to a valid Secret
func FromCSVRecord(record Record, log *zerolog.Logger) (secret Secret, err error) {

	sha256Sum, err := record.Sha256Sum()
	if err != nil {
		log.Error().Err(err).Msg("error getting sha256sum")
		return secret, err
	}
	contents, err := record.Contents()
	if err != nil {
		log.Error().Err(err).Msg("error getting JSON contents")
		return secret, err
	}

	secret = Secret{
		Data: Data{
			Contents:  contents,
			Sha256Sum: sha256Sum,
		},
		Metadata: Metadata{
			ResourceType: record.ResourceType,
			Environment:  record.Environment,
			Access:       record.Access,
		},
	}
	log.Debug().Msgf("new secret from CSV: %v", secret.Metadata.SecretID())
	return secret, err

}
