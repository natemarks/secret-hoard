package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// CreateOrUpdateRDSSecret creates or updates a secret in AWS Secrets Manager
func CreateOrUpdateRDSSecret(secret types.RDSSecret, overwrite bool, log *zerolog.Logger) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(secret.Data)
	if err != nil {
		log.Error().Err(err).Msg("error marshalling secret data")
		return
	}
	// Convert RDSSecretMetadata to tags
	tags := map[string]string{
		"ResourceType": secret.Metadata.ResourceType,
		"Environment":  secret.Metadata.Environment,
		"Instance":     secret.Metadata.Instance,
		"Database":     secret.Metadata.Database,
		"Access":       secret.Metadata.Access,
		"Source":       "secret-hoard",
	}

	// Create the secret
	createSecretInput := &secretsmanager.CreateSecretInput{
		Name:         aws.String(fmt.Sprint(secret.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
		Tags:         ConvertMapToTags(tags),
	}
	_, err = client.CreateSecret(ctx, createSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		var e *smtypes.ResourceExistsException
		if !errors.As(err, &e) {
			log.Error().Err(err).Msgf("error creating secret: %s", *createSecretInput.Name)
			return
		}
		if !overwrite {
			log.Error().Err(err).Msgf("secret already exists: %s", *createSecretInput.Name)
			return
		}
		// Update the secret string value
		updateSecretInput := &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(fmt.Sprint(secret.Metadata.SecretID())),
			SecretString: aws.String(string(secretValue)),
		}
		_, err := client.UpdateSecret(ctx, updateSecretInput)
		if err != nil {
			log.Error().Err(err).Msgf("error updating secret string: %s", *updateSecretInput.SecretId)
			return
		}
		// Update the secret tags
		tagResourceInput := &secretsmanager.TagResourceInput{
			SecretId: aws.String(fmt.Sprint(secret.Metadata.SecretID())),
			Tags:     ConvertMapToTags(tags),
		}
		_, err = client.TagResource(ctx, tagResourceInput)
		if err != nil {
			log.Error().Err(err).Msgf("error updating secret tags: %s", *updateSecretInput.SecretId)
			return
		}
		log.Info().Msgf("secret updated successfully: %s", *updateSecretInput.SecretId)
		return
	}
	log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
}
