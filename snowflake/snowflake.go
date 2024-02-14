package snowflake

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

// Metadata RDS secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // snowflake
	Environment  string `json:"environment"`  // dev, integration, staging, production
	Warehouse    string `json:"warehouse"`    // some_warehouse
	Access       string `json:"access"`       // readwrite, admin
}

// Map converts Metadata to a map of strings to simplify tagging
func (rm Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": rm.ResourceType,
		"Environment":  rm.Environment,
		"Warehouse":    rm.Warehouse,
		"Access":       rm.Access,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id for the rdsSecret
func (rm Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v/%v", rm.ResourceType, rm.Environment, rm.Warehouse, rm.Access)
}

// Data is the struct of the secret for a snowflake connection
// Password: the password for the database user
// AccountName: the database engine
// Warehouse: the port the database is listening on
// Username: the username for the database user
type Data struct {
	Password    string `json:"password"`
	AccountName string `json:"accountName"`
	Warehouse   string `json:"warehouse"`
	Username    string `json:"username"`
}

// Secret is the struct of the secret generated for RDS by CDK deployment
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

// Create the secret in secretsmanager
func (s Secret) Create(log *zerolog.Logger) {
	log.Debug().Msgf("creating RDS rdsSecret: %s", s.Metadata.SecretID())
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

	// Create the rdsSecret
	createSecretInput := &secretsmanager.CreateSecretInput{
		Name:         aws.String(fmt.Sprint(s.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
		Tags:         tools.ConvertMapToTags(tags),
	}
	_, err = client.CreateSecret(ctx, createSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		log.Error().Err(err).Msgf("error creating rdsSecret: %s", *createSecretInput.Name)
		return
	}
	log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
}

// Update the RDS rdsSecret
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

	// Create the rdsSecret
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

	secret = Secret{
		Data: Data{
			Password:    record.Password,
			AccountName: record.AccountName,
			Warehouse:   record.Warehouse,
			Username:    record.Username,
		},
		Metadata: Metadata{
			ResourceType: record.ResourceType,
			Environment:  record.Environment,
			Warehouse:    record.Warehouse,
			Access:       record.Access,
		},
	}
	log.Debug().Msgf("new secret from CSV: %v", secret.Metadata.SecretID())
	return secret, err

}
