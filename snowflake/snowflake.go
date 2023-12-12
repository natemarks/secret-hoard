package snowflake

import (
	"context"
	"encoding/json"
	"fmt"

	aws2 "github.com/natemarks/secret-hoard/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// CreateSnowflakeSecrets creates RDS secrets in AWS Secrets Manager
func CreateSnowflakeSecrets(secrets []types.SnowflakeSecret, log *zerolog.Logger) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	for _, secret := range secrets {
		// Convert SnowflakeSecret to JSON string
		secretValue, err := json.Marshal(secret.Data)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling secret data")
			continue
		}

		// Convert RDSSecretMetadata to tags
		tags := map[string]string{
			"ResourceType": secret.Metadata.ResourceType,
			"Environment":  secret.Metadata.Environment,
			"Warehouse":    secret.Metadata.Warehouse,
			"Access":       secret.Metadata.Access,
			"Source":       "secret-hoard",
		}

		// Create the secret
		createSecretInput := &secretsmanager.CreateSecretInput{
			Name:         aws.String(fmt.Sprint(secret.Metadata.SecretID())),
			SecretString: aws.String(string(secretValue)),
			Tags:         aws2.ConvertMapToTags(tags),
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
