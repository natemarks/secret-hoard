package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func CreateRDSSecrets(secrets []types.RDSSecret, log *zerolog.Logger) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	for _, secret := range secrets {
		// Convert RDSSecretData to JSON string
		secretValue, err := json.Marshal(secret.Data)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling secret data")
			continue
		}

		// Convert RDSSecretMetadata to tags
		tags := map[string]string{
			"Environment": secret.Metadata.Environment,
			"Instance":    secret.Metadata.Instance,
			"Database":    secret.Metadata.Database,
			"Type":        secret.Metadata.Type,
		}

		// Create the secret
		createSecretInput := &secretsmanager.CreateSecretInput{
			Name: aws.String(fmt.Sprintf(
				"%v/%v/%v/%v",
				secret.Metadata.Environment,
				secret.Metadata.Instance,
				secret.Metadata.Database,
				secret.Metadata.Type,
			),
			), // dev/instance/database/type
			SecretString: aws.String(string(secretValue)),
			Tags:         convertMapToTags(tags),
		}

		_, err = client.CreateSecret(ctx, createSecretInput)
		if err != nil {
			log.Error().Err(err).Msg("error creating secret")
			continue
		}
		log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
	}

	return nil
}

// Convert a map to a list of tags
func convertMapToTags(tags map[string]string) []smtypes.Tag {
	var tagList []smtypes.Tag
	for key, value := range tags {
		tag := smtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}
		tagList = append(tagList, tag)
	}
	return tagList
}
